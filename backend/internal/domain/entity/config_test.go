package entity

import (
	"strings"
	"testing"
	"time"
)

func validConfig() Config {
	return DefaultConfig()
}

func TestConfig_Validate_ValidConfig(t *testing.T) {
	cfg := validConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config to pass validation, got: %v", err)
	}
}

func TestConfig_Validate_PortZero(t *testing.T) {
	cfg := validConfig()
	cfg.Server.Port = 0
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for port 0")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "port") {
		t.Errorf("error should mention 'port', got: %v", err)
	}
}

func TestConfig_Validate_PortTooHigh(t *testing.T) {
	cfg := validConfig()
	cfg.Server.Port = 70000
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for port 70000")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "port") {
		t.Errorf("error should mention 'port', got: %v", err)
	}
}

func TestConfig_Validate_UnknownReasoningEffort(t *testing.T) {
	cfg := validConfig()
	cfg.AI.Reasoning["default"] = "ultra"
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for unknown reasoning effort 'ultra'")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "reasoning") {
		t.Errorf("error should mention 'reasoning', got: %v", err)
	}
}

func TestConfig_Validate_FanInWarningNotLessThanCritical(t *testing.T) {
	cfg := validConfig()
	cfg.Analysis.Bottleneck.FanInWarning = 10
	cfg.Analysis.Bottleneck.FanInCritical = 10
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error when fan_in_warning >= fan_in_critical")
	}

	cfg.Analysis.Bottleneck.FanInWarning = 15
	cfg.Analysis.Bottleneck.FanInCritical = 10
	err = cfg.Validate()
	if err == nil {
		t.Fatal("expected error when fan_in_warning > fan_in_critical")
	}
}

func TestConfig_Validate_EmptyCORSOrigins(t *testing.T) {
	cfg := validConfig()
	cfg.Server.CORSOrigins = []string{}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for empty CORS origins")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "cors") {
		t.Errorf("error should mention 'cors', got: %v", err)
	}
}

func TestConfig_Validate_NilCORSOrigins(t *testing.T) {
	cfg := validConfig()
	cfg.Server.CORSOrigins = nil
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for nil CORS origins")
	}
}

func TestConfig_Validate_RequestTimeoutZero(t *testing.T) {
	cfg := validConfig()
	cfg.AI.RequestTimeout = 0
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for zero request timeout")
	}
}

func TestConfig_Validate_NegativeRequestTimeout(t *testing.T) {
	cfg := validConfig()
	cfg.AI.RequestTimeout = -1 * time.Second
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for negative request timeout")
	}
}

func TestConfig_Validate_DefaultTeamSizeZero(t *testing.T) {
	cfg := validConfig()
	cfg.Analysis.DefaultTeamSize = 0
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for default_team_size 0")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Server defaults
	if cfg.Server.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("expected host 0.0.0.0, got %s", cfg.Server.Host)
	}
	if len(cfg.Server.CORSOrigins) != 1 || cfg.Server.CORSOrigins[0] != "http://localhost:5173" {
		t.Errorf("unexpected CORS origins: %v", cfg.Server.CORSOrigins)
	}
	if cfg.Server.ReadTimeout != 30*time.Second {
		t.Errorf("expected read timeout 30s, got %v", cfg.Server.ReadTimeout)
	}
	if cfg.Server.WriteTimeout != 30*time.Second {
		t.Errorf("expected write timeout 30s, got %v", cfg.Server.WriteTimeout)
	}
	if cfg.Server.ShutdownTimeout != 10*time.Second {
		t.Errorf("expected shutdown timeout 10s, got %v", cfg.Server.ShutdownTimeout)
	}

	// AI defaults
	if !cfg.AI.Enabled {
		t.Error("expected AI enabled by default")
	}
	if cfg.AI.Model != "gpt-5.4-nano" {
		t.Errorf("expected model gpt-5.4-nano, got %s", cfg.AI.Model)
	}
	if cfg.AI.BaseURL != "https://api.openai.com/v1" {
		t.Errorf("expected base URL https://api.openai.com/v1, got %s", cfg.AI.BaseURL)
	}
	if cfg.AI.APIKeyEnv != "UNM_OPENAI_API_KEY" {
		t.Errorf("expected api_key_env UNM_OPENAI_API_KEY, got %s", cfg.AI.APIKeyEnv)
	}
	if cfg.AI.RequestTimeout != 60*time.Second {
		t.Errorf("expected request timeout 60s, got %v", cfg.AI.RequestTimeout)
	}

	// AI reasoning
	expectedReasoning := map[string]string{
		"default":                 "low",
		"insights":                "none",
		"query":                   "none",
		"advisor":                 "low",
		"advisor_recommendations": "low",
		"advisor_whatif-scenario":  "low",
		"whatif":                   "low",
	}
	for k, v := range expectedReasoning {
		if cfg.AI.Reasoning[k] != v {
			t.Errorf("expected reasoning[%s]=%s, got %s", k, v, cfg.AI.Reasoning[k])
		}
	}

	// Analysis defaults
	if cfg.Analysis.DefaultTeamSize != 5 {
		t.Errorf("expected default_team_size 5, got %d", cfg.Analysis.DefaultTeamSize)
	}
	if cfg.Analysis.OverloadedCapabilityThreshold != 6 {
		t.Errorf("expected overloaded_capability_threshold 6, got %d", cfg.Analysis.OverloadedCapabilityThreshold)
	}

	// Cognitive load thresholds
	if cfg.Analysis.CognitiveLoad.DomainSpreadThresholds != [2]int{4, 6} {
		t.Errorf("unexpected domain_spread_thresholds: %v", cfg.Analysis.CognitiveLoad.DomainSpreadThresholds)
	}
	if cfg.Analysis.CognitiveLoad.ServiceLoadThresholds != [2]float64{2.0, 3.0} {
		t.Errorf("unexpected service_load_thresholds: %v", cfg.Analysis.CognitiveLoad.ServiceLoadThresholds)
	}
	if cfg.Analysis.CognitiveLoad.InteractionLoadThresholds != [2]int{4, 7} {
		t.Errorf("unexpected interaction_load_thresholds: %v", cfg.Analysis.CognitiveLoad.InteractionLoadThresholds)
	}
	if cfg.Analysis.CognitiveLoad.DependencyLoadThresholds != [2]int{5, 9} {
		t.Errorf("unexpected dependency_load_thresholds: %v", cfg.Analysis.CognitiveLoad.DependencyLoadThresholds)
	}

	// Interaction weights
	if cfg.Analysis.InteractionWeights.Collaboration != 3 {
		t.Errorf("expected collaboration weight 3, got %d", cfg.Analysis.InteractionWeights.Collaboration)
	}
	if cfg.Analysis.InteractionWeights.Facilitating != 2 {
		t.Errorf("expected facilitating weight 2, got %d", cfg.Analysis.InteractionWeights.Facilitating)
	}
	if cfg.Analysis.InteractionWeights.XAsAService != 1 {
		t.Errorf("expected x-as-a-service weight 1, got %d", cfg.Analysis.InteractionWeights.XAsAService)
	}

	// Bottleneck
	if cfg.Analysis.Bottleneck.FanInWarning != 5 {
		t.Errorf("expected fan_in_warning 5, got %d", cfg.Analysis.Bottleneck.FanInWarning)
	}
	if cfg.Analysis.Bottleneck.FanInCritical != 10 {
		t.Errorf("expected fan_in_critical 10, got %d", cfg.Analysis.Bottleneck.FanInCritical)
	}

	// Signals
	if cfg.Analysis.Signals.NeedTeamSpanWarning != 2 {
		t.Errorf("expected need_team_span_warning 2, got %d", cfg.Analysis.Signals.NeedTeamSpanWarning)
	}
	if cfg.Analysis.Signals.NeedTeamSpanCritical != 3 {
		t.Errorf("expected need_team_span_critical 3, got %d", cfg.Analysis.Signals.NeedTeamSpanCritical)
	}
	if cfg.Analysis.Signals.HighSpanServiceThreshold != 3 {
		t.Errorf("expected high_span_service_threshold 3, got %d", cfg.Analysis.Signals.HighSpanServiceThreshold)
	}
	if cfg.Analysis.Signals.InteractionOverReliance != 4 {
		t.Errorf("expected interaction_over_reliance 4, got %d", cfg.Analysis.Signals.InteractionOverReliance)
	}
	if cfg.Analysis.Signals.DepthChainThreshold != 4 {
		t.Errorf("expected depth_chain_threshold 4, got %d", cfg.Analysis.Signals.DepthChainThreshold)
	}

	// Value chain
	if cfg.Analysis.ValueChain.AtRiskTeamSpan != 3 {
		t.Errorf("expected at_risk_team_span 3, got %d", cfg.Analysis.ValueChain.AtRiskTeamSpan)
	}

	// Features
	if cfg.Features.DebugRoutes {
		t.Error("expected debug_routes false by default")
	}

	// Logging
	if cfg.Logging.Level != "info" {
		t.Errorf("expected logging level info, got %s", cfg.Logging.Level)
	}
	if cfg.Logging.Format != "text" {
		t.Errorf("expected logging format text, got %s", cfg.Logging.Format)
	}

	// AI model routing
	expectedModels := map[string]string{
		"default":                 "gpt-5.4-nano",
		"insights":                "gpt-5-nano",
		"query":                   "gpt-5-nano",
		"advisor":                 "gpt-5.4-nano",
		"advisor_recommendations": "gpt-5.4-mini",
		"advisor_whatif-scenario":  "gpt-5.4-mini",
		"whatif":                   "gpt-5.4-mini",
	}
	for k, v := range expectedModels {
		if cfg.AI.Models[k] != v {
			t.Errorf("expected models[%s]=%s, got %s", k, v, cfg.AI.Models[k])
		}
	}

	// DefaultConfig should pass validation
	if err := cfg.Validate(); err != nil {
		t.Fatalf("DefaultConfig should pass validation, got: %v", err)
	}
}

func TestModelForCategory_LookupOrder(t *testing.T) {
	ai := AIConfig{
		Model: "gpt-5.4-nano",
		Models: map[string]string{
			"default":                 "gpt-5.4-nano",
			"insights":                "gpt-5-nano",
			"advisor":                 "gpt-5.4-nano",
			"advisor_recommendations": "gpt-5.4-mini",
		},
	}

	tests := []struct {
		template string
		want     string
	}{
		{"insights/signals", "gpt-5-nano"},
		{"insights/dashboard", "gpt-5-nano"},
		{"advisor/general", "gpt-5.4-nano"},
		{"advisor/recommendations", "gpt-5.4-mini"},
		{"advisor/bottleneck", "gpt-5.4-nano"},
		{"query/model-summary", "gpt-5.4-nano"},
		{"unknown/thing", "gpt-5.4-nano"},
	}

	for _, tt := range tests {
		got := ai.ModelForCategory(tt.template)
		if got != tt.want {
			t.Errorf("ModelForCategory(%q) = %q, want %q", tt.template, got, tt.want)
		}
	}
}

func TestModelForCategory_FallsBackToModel(t *testing.T) {
	ai := AIConfig{
		Model:  "gpt-5.4-nano",
		Models: map[string]string{},
	}
	got := ai.ModelForCategory("advisor/general")
	if got != "gpt-5.4-nano" {
		t.Errorf("expected fallback to Model field, got %q", got)
	}
}
