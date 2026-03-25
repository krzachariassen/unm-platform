// Starts the UNM server, loads inca.unm.yaml via the API, then runs all
// 30 AI questions through POST /api/models/{id}/ask and prints each response.
//
// Usage: source ../ai.env && go run ./cmd/runquestions/
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const serverAddr = "http://localhost:8765"

type aiQuestion struct {
	ID             int      `yaml:"id"`
	Category       string   `yaml:"category"`
	Template       string   `yaml:"template"`
	Question       string   `yaml:"question"`
	MustMention    []string `yaml:"must_mention"`
	MustNotMention []string `yaml:"must_not_mention"`
	Notes          string   `yaml:"notes"`
}

func main() {
	if os.Getenv("UNM_OPENAI_API_KEY") == "" {
		fmt.Fprintln(os.Stderr, "UNM_OPENAI_API_KEY not set")
		os.Exit(1)
	}

	// Start server
	fmt.Println("Starting UNM server on :8765 ...")
	cmd := exec.Command("go", "run", "./cmd/server/")
	cmd.Env = append(os.Environ(), "PORT=8765")
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to start server: %v\n", err)
		os.Exit(1)
	}
	defer cmd.Process.Kill()

	// Wait for server to be ready
	for i := 0; i < 40; i++ {
		resp, err := http.Get(serverAddr + "/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			fmt.Println("Server ready.")
			break
		}
		if i == 39 {
			fmt.Fprintln(os.Stderr, "server did not start in time")
			os.Exit(1)
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Load INCA model
	incaPath := filepath.Join("..", "examples", "inca.unm.yaml")
	yamlBytes, err := os.ReadFile(incaPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read inca: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverAddr+"/api/models/parse", "application/yaml", bytes.NewReader(yamlBytes))
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse model: %v\n", err)
		os.Exit(1)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Fprintf(os.Stderr, "parse failed %d: %s\n", resp.StatusCode, body)
		os.Exit(1)
	}
	var parseResp struct {
		ID         string `json:"id"`
		SystemName string `json:"system_name"`
	}
	if err := json.Unmarshal(body, &parseResp); err != nil {
		fmt.Fprintf(os.Stderr, "parse response: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Loaded model: %s (id=%s)\n", parseResp.SystemName, parseResp.ID)

	// Load questions
	questionsPath := filepath.Join("testdata", "ai_questions.yaml")
	qData, err := os.ReadFile(questionsPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read questions: %v\n", err)
		os.Exit(1)
	}
	var questions []aiQuestion
	if err := yaml.Unmarshal(qData, &questions); err != nil {
		fmt.Fprintf(os.Stderr, "yaml: %v\n", err)
		os.Exit(1)
	}

	sep := strings.Repeat("─", 80)
	passed, failed := 0, 0
	askURL := fmt.Sprintf("%s/api/models/%s/ask", serverAddr, parseResp.ID)

	for _, q := range questions {
		fmt.Printf("\n%s\n", sep)
		fmt.Printf("Q%d [%s]\n%s\n%s\n", q.ID, q.Category, q.Question, sep)

		payload, _ := json.Marshal(map[string]string{
			"question": q.Question,
			"category": q.Category,
		})

		r, err := http.Post(askURL, "application/json", bytes.NewReader(payload))
		if err != nil {
			fmt.Printf("HTTP ERROR: %v\n", err)
			failed++
			continue
		}
		respBody, _ := io.ReadAll(r.Body)
		r.Body.Close()

		if r.StatusCode != 200 {
			fmt.Printf("API ERROR %d: %s\n", r.StatusCode, respBody)
			failed++
			continue
		}

		var askResp struct {
			Answer       string `json:"answer"`
			FinishReason string `json:"finish_reason"`
			AIConfigured bool   `json:"ai_configured"`
		}
		if err := json.Unmarshal(respBody, &askResp); err != nil {
			fmt.Printf("DECODE ERROR: %v\n", err)
			failed++
			continue
		}

		fmt.Println(askResp.Answer)

		// Check assertions
		answer := strings.ToLower(askResp.Answer)
		ok := true
		for _, phrase := range q.MustMention {
			if !strings.Contains(answer, strings.ToLower(phrase)) {
				fmt.Printf("\n⚠  MUST MENTION (missing): %q\n", phrase)
				ok = false
			}
		}
		for _, phrase := range q.MustNotMention {
			if strings.Contains(answer, strings.ToLower(phrase)) {
				fmt.Printf("\n⚠  MUST NOT MENTION (present): %q\n", phrase)
				ok = false
			}
		}
		if ok {
			passed++
			fmt.Printf("\n✓ assertions pass  [finish_reason=%s]\n", askResp.FinishReason)
		} else {
			failed++
		}

		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("\n%s\n", sep)
	fmt.Printf("Results: %d passed, %d failed, %d total\n", passed, failed, len(questions))
}
