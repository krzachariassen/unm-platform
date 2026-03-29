package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
)

const defaultBaseURL = "https://api.openai.com/v1"
const defaultModel = "gpt-5.4-nano"

// CallOptions configures per-request overrides for Complete/CompleteJSON.
type CallOptions struct {
	Model           string // override the client's default model
	ReasoningEffort string // "low", "medium", "high", "xhigh", or "" to omit
}

// CallOption is a functional option for Complete/CompleteJSON.
type CallOption func(*CallOptions)

// WithModel overrides the model for a single call.
func WithModel(model string) CallOption {
	return func(o *CallOptions) { o.Model = model }
}

// WithReasoning sets the reasoning effort for a single call.
func WithReasoning(effort string) CallOption {
	return func(o *CallOptions) { o.ReasoningEffort = effort }
}

// ChatMessage represents a message in the OpenAI chat format.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse represents a parsed response from the OpenAI API.
type ChatResponse struct {
	Content      string
	FinishReason string
	Usage        struct {
		PromptTokens     int
		CompletionTokens int
		TotalTokens      int
	}
}

// OpenAIClient is an HTTP client for the OpenAI Chat Completions API.
type OpenAIClient struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

// defaultHTTPTimeout is the safety-net timeout applied to all OpenAI HTTP clients (6.10.23).
// Handlers set a shorter per-request context timeout; this backstop prevents hung connections
// if the context is not properly propagated.
const defaultHTTPTimeout = 120 * time.Second

// NewOpenAIClient creates a client using the UNM_OPENAI_API_KEY environment variable.
func NewOpenAIClient() (*OpenAIClient, error) {
	key := os.Getenv("UNM_OPENAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("UNM_OPENAI_API_KEY environment variable not set")
	}
	return &OpenAIClient{
		apiKey:     key,
		model:      defaultModel,
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{Timeout: defaultHTTPTimeout},
	}, nil
}

// NewOpenAIClientWithKey creates a client with an explicit API key and model.
func NewOpenAIClientWithKey(apiKey, model string) *OpenAIClient {
	return &OpenAIClient{
		apiKey:     apiKey,
		model:      model,
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{Timeout: defaultHTTPTimeout},
	}
}

// NewOpenAIClientFromConfig creates a client from an AIConfig.
func NewOpenAIClientFromConfig(cfg entity.AIConfig) (*OpenAIClient, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("AI APIKey not set (check %s env var)", cfg.APIKeyEnv)
	}
	return &OpenAIClient{
		baseURL:    cfg.BaseURL,
		model:      cfg.Model,
		apiKey:     cfg.APIKey,
		httpClient: &http.Client{Timeout: defaultHTTPTimeout},
	}, nil
}

// IsConfigured returns true if the client has a non-empty API key.
func (c *OpenAIClient) IsConfigured() bool {
	return c.apiKey != ""
}

// ResponseFormatOption specifies the response format for the OpenAI API.
type ResponseFormatOption struct {
	Type string `json:"type"`
}

type chatCompletionRequest struct {
	Model           string               `json:"model"`
	Messages        []ChatMessage        `json:"messages"`
	ReasoningEffort string               `json:"reasoning_effort,omitempty"`
	ResponseFormat  *ResponseFormatOption `json:"response_format,omitempty"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message      ChatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// supportsReasoningEffort returns true for models that accept the reasoning_effort field.
func supportsReasoningEffort(model string) bool {
	return strings.HasPrefix(model, "o1") ||
		strings.HasPrefix(model, "o3") ||
		strings.HasPrefix(model, "o4") ||
		strings.HasPrefix(model, "gpt-5")
}

// Complete sends a single-turn prompt and returns the response.
// Accepts CallOption to override model and reasoning effort per call.
func (c *OpenAIClient) Complete(ctx context.Context, systemPrompt, userMessage string, opts ...CallOption) (ChatResponse, error) {
	o := CallOptions{Model: c.model}
	for _, opt := range opts {
		opt(&o)
	}

	reqBody := chatCompletionRequest{
		Model: o.Model,
		Messages: []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMessage},
		},
	}
	if o.ReasoningEffort != "" && o.ReasoningEffort != "none" && supportsReasoningEffort(o.Model) {
		reqBody.ReasoningEffort = o.ReasoningEffort
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return ChatResponse{}, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return ChatResponse{}, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var apiResp chatCompletionResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return ChatResponse{}, fmt.Errorf("unmarshaling response: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return ChatResponse{}, fmt.Errorf("OpenAI API returned no choices")
	}

	result := ChatResponse{
		Content:      apiResp.Choices[0].Message.Content,
		FinishReason: apiResp.Choices[0].FinishReason,
	}
	result.Usage.PromptTokens = apiResp.Usage.PromptTokens
	result.Usage.CompletionTokens = apiResp.Usage.CompletionTokens
	result.Usage.TotalTokens = apiResp.Usage.TotalTokens

	log.Printf("[AI-COST] Complete model=%s prompt_tokens=%d completion_tokens=%d total_tokens=%d",
		o.Model, result.Usage.PromptTokens, result.Usage.CompletionTokens, result.Usage.TotalTokens)

	return result, nil
}

// CompleteJSON sends a single-turn prompt and returns the response with JSON response format.
// Accepts CallOption to override the model per call.
func (c *OpenAIClient) CompleteJSON(ctx context.Context, systemPrompt, userMessage string, opts ...CallOption) (string, error) {
	o := CallOptions{Model: c.model}
	for _, opt := range opts {
		opt(&o)
	}

	reqBody := chatCompletionRequest{
		Model: o.Model,
		Messages: []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMessage},
		},
		ResponseFormat: &ResponseFormatOption{Type: "json_object"},
	}
	if o.ReasoningEffort != "" && o.ReasoningEffort != "none" && supportsReasoningEffort(o.Model) {
		reqBody.ReasoningEffort = o.ReasoningEffort
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var apiResp chatCompletionResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return "", fmt.Errorf("unmarshaling response: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return "", fmt.Errorf("OpenAI API returned no choices")
	}

	log.Printf("[AI-COST] CompleteJSON model=%s prompt_tokens=%d completion_tokens=%d total_tokens=%d",
		o.Model, apiResp.Usage.PromptTokens, apiResp.Usage.CompletionTokens, apiResp.Usage.TotalTokens)

	return apiResp.Choices[0].Message.Content, nil
}
