package entity

import (
	"fmt"
	"strings"
	"time"
)

// Config is the root configuration for the UNM Platform.
type Config struct {
	Server   ServerConfig   `koanf:"server"`
	Frontend FrontendConfig `koanf:"frontend"`
	AI       AIConfig       `koanf:"ai"`
	Analysis AnalysisConfig `koanf:"analysis"`
	Features FeaturesConfig `koanf:"features"`
	Logging  LoggingConfig  `koanf:"logging"`
	Storage  StorageConfig  `koanf:"storage"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port            int           `koanf:"port"`
	Host            string        `koanf:"host"`
	CORSOrigins     []string      `koanf:"cors_origins"`
	ReadTimeout     time.Duration `koanf:"read_timeout"`
	WriteTimeout    time.Duration `koanf:"write_timeout"`
	ShutdownTimeout time.Duration `koanf:"shutdown_timeout"`
	SessionTTL      time.Duration `koanf:"session_ttl"`
	StaticDir       string        `koanf:"static_dir"`
}

// FrontendConfig holds frontend-related settings.
type FrontendConfig struct {
	APIBaseURL     string `koanf:"api_base_url"`
	DevProxyTarget string `koanf:"dev_proxy_target"`
	DevServerPort  int    `koanf:"dev_server_port"`
}

// AIConfig holds AI provider settings.
type AIConfig struct {
	Enabled        bool                       `koanf:"enabled"`
	Provider       string                     `koanf:"provider"`
	APIKeyEnv      string                     `koanf:"api_key_env"`
	APIKey         string                     `koanf:"-"` // resolved at load time from APIKeyEnv, never in YAML
	BaseURL        string                     `koanf:"base_url"`
	Model          string                     `koanf:"model"`
	Models         map[string]string          `koanf:"models"`
	RequestTimeout time.Duration              `koanf:"request_timeout"`
	Timeouts       map[string]string          `koanf:"timeouts"`
	Reasoning      map[string]string          `koanf:"reasoning"`
	AllowedIPs     []string                   `koanf:"allowed_ips"` // empty = no restriction; set to restrict AI by client IP
}

// ModelForCategory returns the model to use for a given template category.
// Lookup order: full path with "/" replaced by "_" (e.g. "advisor/recommendations" → "advisor_recommendations")
// → prefix only (e.g. "advisor") → "default" key → top-level Model field.
func (c AIConfig) ModelForCategory(templateName string) string {
	flat := strings.ReplaceAll(templateName, "/", "_")
	if m, ok := c.Models[flat]; ok {
		return m
	}
	key := templateName
	if idx := strings.Index(templateName, "/"); idx >= 0 {
		key = templateName[:idx]
	}
	if m, ok := c.Models[key]; ok {
		return m
	}
	if m, ok := c.Models["default"]; ok {
		return m
	}
	return c.Model
}

// TimeoutForCategory returns the request timeout for a given template category.
// Same lookup order as ModelForCategory. Falls back to the top-level RequestTimeout.
func (c AIConfig) TimeoutForCategory(templateName string) time.Duration {
	flat := strings.ReplaceAll(templateName, "/", "_")
	if s, ok := c.Timeouts[flat]; ok {
		if d, err := time.ParseDuration(s); err == nil {
			return d
		}
	}
	key := templateName
	if idx := strings.Index(templateName, "/"); idx >= 0 {
		key = templateName[:idx]
	}
	if s, ok := c.Timeouts[key]; ok {
		if d, err := time.ParseDuration(s); err == nil {
			return d
		}
	}
	if s, ok := c.Timeouts["default"]; ok {
		if d, err := time.ParseDuration(s); err == nil {
			return d
		}
	}
	return c.RequestTimeout
}

// AnalysisConfig holds analysis engine settings.
type AnalysisConfig struct {
	DefaultTeamSize               int                     `koanf:"default_team_size"`
	OverloadedCapabilityThreshold int                     `koanf:"overloaded_capability_threshold"`
	CognitiveLoad                 CognitiveLoadConfig     `koanf:"cognitive_load"`
	InteractionWeights            InteractionWeightConfig `koanf:"interaction_weights"`
	Bottleneck                    BottleneckConfig        `koanf:"bottleneck"`
	Signals                       SignalsConfig           `koanf:"signals"`
	ValueChain                    ValueChainConfig        `koanf:"value_chain"`
}

// CognitiveLoadConfig holds cognitive load analysis thresholds.
type CognitiveLoadConfig struct {
	DomainSpreadThresholds    [2]int     `koanf:"domain_spread_thresholds"`
	ServiceLoadThresholds     [2]float64 `koanf:"service_load_thresholds"`
	InteractionLoadThresholds [2]int     `koanf:"interaction_load_thresholds"`
	DependencyLoadThresholds  [2]int     `koanf:"dependency_load_thresholds"`
}

// InteractionWeightConfig holds weights for interaction modes.
type InteractionWeightConfig struct {
	Collaboration int `koanf:"collaboration"`
	Facilitating  int `koanf:"facilitating"`
	XAsAService   int `koanf:"x-as-a-service"`
}

// BottleneckConfig holds bottleneck detection thresholds.
type BottleneckConfig struct {
	FanInWarning  int `koanf:"fan_in_warning"`
	FanInCritical int `koanf:"fan_in_critical"`
}

// SignalsConfig holds signal detection thresholds.
type SignalsConfig struct {
	NeedTeamSpanWarning      int `koanf:"need_team_span_warning"`
	NeedTeamSpanCritical     int `koanf:"need_team_span_critical"`
	HighSpanServiceThreshold int `koanf:"high_span_service_threshold"`
	InteractionOverReliance  int `koanf:"interaction_over_reliance"`
	DepthChainThreshold      int `koanf:"depth_chain_threshold"`
}

// ValueChainConfig holds value chain analysis thresholds.
type ValueChainConfig struct {
	AtRiskTeamSpan int `koanf:"at_risk_team_span"`
}

// FeaturesConfig holds feature flag settings.
type FeaturesConfig struct {
	DebugRoutes       bool     `koanf:"debug_routes"`
	DebugExamplePaths []string `koanf:"debug_example_paths"`
}

// LoggingConfig holds logging settings.
type LoggingConfig struct {
	Level  string `koanf:"level"`
	Format string `koanf:"format"`
}

// StorageConfig holds persistence backend settings.
type StorageConfig struct {
	// Driver selects the storage backend: "memory" or "postgres".
	Driver string `koanf:"driver"`
	// DatabaseURL is the PostgreSQL connection string (e.g. postgres://user:pass@host/db).
	// Set via UNM_STORAGE__DATABASE_URL environment variable.
	DatabaseURL string `koanf:"database_url"`
	// MaxConnections is the maximum number of PG pool connections.
	MaxConnections int `koanf:"max_connections"`
	// MigrateOnStartup runs golang-migrate on startup when true.
	MigrateOnStartup bool `koanf:"migrate_on_startup"`
}

// validReasoningEfforts is the set of allowed reasoning effort values.
var validReasoningEfforts = map[string]bool{
	"none":   true,
	"low":    true,
	"medium": true,
	"high":   true,
	"xhigh":  true,
}

// Validate checks the Config for invalid values and returns an error describing all problems found.
func (c *Config) Validate() error {
	var errs []string

	// Server validation
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		errs = append(errs, fmt.Sprintf("server.port must be between 1 and 65535, got %d", c.Server.Port))
	}
	if len(c.Server.CORSOrigins) == 0 {
		errs = append(errs, "server.cors_origins must not be empty")
	}

	// AI validation
	if c.AI.RequestTimeout <= 0 {
		errs = append(errs, fmt.Sprintf("ai.request_timeout must be positive, got %v", c.AI.RequestTimeout))
	}
	for key, effort := range c.AI.Reasoning {
		if !validReasoningEfforts[effort] {
			errs = append(errs, fmt.Sprintf("ai.reasoning[%s] has invalid value %q; valid values: none, low, medium, high, xhigh", key, effort))
		}
	}

	// Analysis validation
	if c.Analysis.DefaultTeamSize < 1 {
		errs = append(errs, fmt.Sprintf("analysis.default_team_size must be >= 1, got %d", c.Analysis.DefaultTeamSize))
	}
	if c.Analysis.Bottleneck.FanInWarning >= c.Analysis.Bottleneck.FanInCritical {
		errs = append(errs, fmt.Sprintf("analysis.bottleneck.fan_in_warning (%d) must be less than fan_in_critical (%d)", c.Analysis.Bottleneck.FanInWarning, c.Analysis.Bottleneck.FanInCritical))
	}

	if len(errs) > 0 {
		return fmt.Errorf("config validation failed:\n  %s", strings.Join(errs, "\n  "))
	}
	return nil
}

// DefaultConfig returns a Config populated with sensible defaults matching base.yaml.
func DefaultConfig() Config {
	return Config{
		Server: ServerConfig{
			Port:            8080,
			Host:            "0.0.0.0",
			CORSOrigins:     []string{"http://localhost:5173"},
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    30 * time.Second,
			ShutdownTimeout: 10 * time.Second,
			SessionTTL:      2 * time.Hour,
		},
		AI: AIConfig{
			Enabled:        true,
			Provider:       "openai",
			APIKeyEnv:      "UNM_OPENAI_API_KEY",
			BaseURL:        "https://api.openai.com/v1",
			Model:          "gpt-5.4-nano",
		Models: map[string]string{
			"default":                  "gpt-5.4-nano",
			"insights":                 "gpt-5-nano",
			"query":                    "gpt-5-nano",
			"advisor":                  "gpt-5.4-nano",
			"advisor_simple":           "gpt-5.4-nano",
			"advisor_complex":          "gpt-5.4-mini",
			"advisor_recommendations":  "gpt-5.4-mini",
			"advisor_whatif-scenario":   "gpt-5.4-mini",
			"advisor_extract-actions":  "gpt-5.4-mini",
			"whatif":                    "gpt-5.4-mini",
		},
		RequestTimeout: 60 * time.Second,
		Timeouts: map[string]string{
			"default":                  "60s",
			"insights":                 "30s",
			"query":                    "30s",
			"advisor_simple":           "30s",
			"advisor_complex":          "180s",
			"advisor_recommendations":  "300s",
			"advisor_whatif-scenario":   "300s",
			"advisor_extract-actions":  "120s",
			"whatif":                    "120s",
		},
		Reasoning: map[string]string{
			"default":                  "low",
			"insights":                 "none",
			"query":                    "none",
			"advisor":                  "low",
			"advisor_simple":           "none",
			"advisor_complex":          "medium",
			"advisor_recommendations":  "low",
			"advisor_whatif-scenario":   "low",
			"advisor_extract-actions":  "low",
			"whatif":                    "low",
		},
		},
		Analysis: AnalysisConfig{
			DefaultTeamSize:               5,
			OverloadedCapabilityThreshold: 6,
			CognitiveLoad: CognitiveLoadConfig{
				DomainSpreadThresholds:    [2]int{4, 6},
				ServiceLoadThresholds:     [2]float64{2.0, 3.0},
				InteractionLoadThresholds: [2]int{4, 7},
				DependencyLoadThresholds:  [2]int{5, 9},
			},
			InteractionWeights: InteractionWeightConfig{
				Collaboration: 3,
				Facilitating:  2,
				XAsAService:   1,
			},
			Bottleneck: BottleneckConfig{
				FanInWarning:  5,
				FanInCritical: 10,
			},
			Signals: SignalsConfig{
				NeedTeamSpanWarning:      2,
				NeedTeamSpanCritical:     3,
				HighSpanServiceThreshold: 3,
				InteractionOverReliance:  4,
				DepthChainThreshold:      4,
			},
			ValueChain: ValueChainConfig{
				AtRiskTeamSpan: 3,
			},
		},
		Features: FeaturesConfig{
			DebugRoutes: false,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
		},
		Storage: StorageConfig{
			Driver:           "memory",
			MaxConnections:   20,
			MigrateOnStartup: true,
		},
	}
}
