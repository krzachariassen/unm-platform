package ai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/krzachariassen/unm-platform/internal/domain/entity"
)

func TestNewOpenAIClientWithKey(t *testing.T) {
	client := NewOpenAIClientWithKey("test-key", "gpt-4o")
	assert.NotNil(t, client)
	assert.True(t, client.IsConfigured())
}

func TestOpenAIClient_IsConfigured_FalseWhenEmpty(t *testing.T) {
	client := NewOpenAIClientWithKey("", "gpt-4o")
	assert.False(t, client.IsConfigured())
}

// TestOpenAIClient_Complete_RealAPI tests against the real OpenAI API.
// Requires UNM_AI_TESTS=true to run (not just the API key).
func TestOpenAIClient_Complete_RealAPI(t *testing.T) {
	if os.Getenv("UNM_AI_TESTS") != "true" {
		t.Skip("UNM_AI_TESTS not enabled — skipping real AI test (set UNM_AI_TESTS=true to run)")
	}

	client, err := NewOpenAIClient()
	require.NoError(t, err)
	require.True(t, client.IsConfigured())

	resp, err := client.Complete(
		context.Background(),
		"You are a helpful assistant. Respond with exactly one word.",
		"Say the word 'hello'.",
	)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Content)
	assert.NotEmpty(t, resp.FinishReason)
	assert.Greater(t, resp.Usage.TotalTokens, 0)
}

// TestOpenAIClient_Complete_Non200Error tests HTTP error handling using a mock server.
// This tests the HTTP client's error-handling behavior, not AI responses.
func TestOpenAIClient_Complete_Non200Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error": {"message": "rate limited"}}`))
	}))
	defer server.Close()

	client := NewOpenAIClientWithKey("test-key", "gpt-4o")
	client.baseURL = server.URL

	_, err := client.Complete(context.Background(), "system", "user")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "429")
}

// TestOpenAIClient_Complete_EmptyChoices tests empty choices handling using a mock server.
// This tests the HTTP client's edge-case handling, not AI responses.
func TestOpenAIClient_Complete_EmptyChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"choices":[],"usage":{"prompt_tokens":0,"completion_tokens":0,"total_tokens":0}}`))
	}))
	defer server.Close()

	client := NewOpenAIClientWithKey("test-key", "gpt-4o")
	client.baseURL = server.URL

	_, err := client.Complete(context.Background(), "system", "user")
	assert.Error(t, err)
}

func TestNewOpenAIClientFromConfig_MissingKey_ReturnsError(t *testing.T) {
	cfg := entity.AIConfig{
		APIKey:    "",
		APIKeyEnv: "UNM_OPENAI_API_KEY",
		BaseURL:   "https://api.openai.com/v1",
		Model:     "gpt-4o",
	}
	_, err := NewOpenAIClientFromConfig(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "UNM_OPENAI_API_KEY")
}

func TestNewOpenAIClientFromConfig_ValidConfig(t *testing.T) {
	cfg := entity.AIConfig{
		APIKey:    "test-key-123",
		APIKeyEnv: "UNM_OPENAI_API_KEY",
		BaseURL:   "https://custom.openai.com/v1",
		Model:     "gpt-4o-mini",
	}
	client, err := NewOpenAIClientFromConfig(cfg)
	require.NoError(t, err)
	assert.True(t, client.IsConfigured())
	assert.Equal(t, "https://custom.openai.com/v1", client.baseURL)
	assert.Equal(t, "gpt-4o-mini", client.model)
}

func TestComplete_ReasoningEffortInRequest(t *testing.T) {
	var capturedBody chatCompletionRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &capturedBody)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`))
	}))
	defer server.Close()

	client := NewOpenAIClientWithKey("test-key", "o3-mini")
	client.baseURL = server.URL

	_, err := client.Complete(context.Background(), "system", "user", WithReasoning("medium"))
	require.NoError(t, err)
	assert.Equal(t, "medium", capturedBody.ReasoningEffort)
}

func TestComplete_NoneReasoningEffortOmitted(t *testing.T) {
	var rawBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &rawBody)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`))
	}))
	defer server.Close()

	client := NewOpenAIClientWithKey("test-key", "gpt-4o")
	client.baseURL = server.URL

	_, err := client.Complete(context.Background(), "system", "user", WithReasoning("none"))
	require.NoError(t, err)
	_, hasReasoningEffort := rawBody["reasoning_effort"]
	assert.False(t, hasReasoningEffort, "reasoning_effort should be omitted when value is 'none'")
}

func TestComplete_WithModelOverride(t *testing.T) {
	var capturedBody chatCompletionRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &capturedBody)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`))
	}))
	defer server.Close()

	client := NewOpenAIClientWithKey("test-key", "gpt-5.4-nano")
	client.baseURL = server.URL

	_, err := client.Complete(context.Background(), "system", "user", WithModel("gpt-5-nano"))
	require.NoError(t, err)
	assert.Equal(t, "gpt-5-nano", capturedBody.Model, "model override should be used in request")
}

func TestCompleteJSON_WithModelOverride(t *testing.T) {
	var capturedBody chatCompletionRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &capturedBody)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"{}"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`))
	}))
	defer server.Close()

	client := NewOpenAIClientWithKey("test-key", "gpt-5.4-nano")
	client.baseURL = server.URL

	_, err := client.CompleteJSON(context.Background(), "system", "user", WithModel("gpt-5-nano"))
	require.NoError(t, err)
	assert.Equal(t, "gpt-5-nano", capturedBody.Model, "model override should be used in JSON request")
}

func TestCompleteJSON_SetsResponseFormat(t *testing.T) {
	var capturedBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &capturedBody)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"{\"key\":\"value\"}"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`))
	}))
	defer server.Close()

	client := NewOpenAIClientWithKey("test-key", "gpt-4o")
	client.baseURL = server.URL

	result, err := client.CompleteJSON(context.Background(), "system", "user")
	require.NoError(t, err)
	assert.Equal(t, `{"key":"value"}`, result)

	// Assert response_format is set to json_object
	rf, ok := capturedBody["response_format"].(map[string]interface{})
	require.True(t, ok, "response_format should be present")
	assert.Equal(t, "json_object", rf["type"])

	// Assert reasoning_effort is NOT set
	_, hasReasoningEffort := capturedBody["reasoning_effort"]
	assert.False(t, hasReasoningEffort, "reasoning_effort should not be set in CompleteJSON")
}

func TestComplete_NoReasoningEffortBackwardCompat(t *testing.T) {
	var rawBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &rawBody)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`))
	}))
	defer server.Close()

	client := NewOpenAIClientWithKey("test-key", "gpt-4o")
	client.baseURL = server.URL

	_, err := client.Complete(context.Background(), "system", "user")
	require.NoError(t, err)
	_, hasReasoningEffort := rawBody["reasoning_effort"]
	assert.False(t, hasReasoningEffort, "reasoning_effort should be omitted when not provided")
}
