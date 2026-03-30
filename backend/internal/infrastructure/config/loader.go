// Package config provides configuration loading for the UNM Platform.
// It uses koanf to load configuration from YAML files and environment variables
// with a deterministic precedence order.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
)

// LoadConfig loads configuration with the following precedence (highest wins):
//  1. Environment variables with prefix UNM_
//  2. config/{environment}.yaml
//  3. config/base.yaml
//  4. Code defaults via entity.DefaultConfig()
//
// If environment is empty, it defaults to "local".
//
// Environment variable naming convention (double-underscore as level separator):
//
//	UNM_<LEVEL1>__<LEVEL2>__<KEY_NAME>
//
// Examples:
//
//	UNM_SERVER__PORT=9090          → server.port
//	UNM_AI__ENABLED=false          → ai.enabled
//	UNM_AI__ALLOWED_IPS=1.2.3.4   → ai.allowed_ips  (single _ preserved in key name)
//	UNM_ANALYSIS__BOTTLENECK__FAN_IN_WARNING=3 → analysis.bottleneck.fan_in_warning
func LoadConfig(environment string) (*entity.Config, error) {
	if environment == "" {
		environment = "local"
	}

	k := koanf.New(".")

	configDir := findConfigDir()
	if configDir == "" {
		return nil, fmt.Errorf("config directory not found: could not locate base.yaml")
	}

	// Load base.yaml (required)
	basePath := filepath.Join(configDir, "base.yaml")
	if err := k.Load(file.Provider(basePath), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("load base.yaml: %w", err)
	}

	// Load {env}.yaml (optional)
	envFile := filepath.Join(configDir, environment+".yaml")
	if _, err := os.Stat(envFile); err == nil {
		if err := k.Load(file.Provider(envFile), yaml.Parser()); err != nil {
			return nil, fmt.Errorf("load %s.yaml: %w", environment, err)
		}
	}

	// Load environment variables using double-underscore as the level separator.
	// Single underscores within a segment are preserved as literal underscores in key names.
	// e.g. UNM_AI__ALLOWED_IPS → ai.allowed_ips
	if err := k.Load(env.Provider("UNM_", ".", envKeyTransform), nil); err != nil {
		return nil, fmt.Errorf("load env vars: %w", err)
	}

	// Start from defaults, then unmarshal loaded config on top
	cfg := entity.DefaultConfig()
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	// Resolve API key secret from environment
	if cfg.AI.APIKeyEnv != "" {
		cfg.AI.APIKey = os.Getenv(cfg.AI.APIKeyEnv)
	}

	// koanf reads env var slice values as a single string; split on comma if needed.
	// e.g. UNM_AI__ALLOWED_IPS=1.2.3.4,10.0.0.0/8 → ["1.2.3.4", "10.0.0.0/8"]
	if len(cfg.AI.AllowedIPs) == 1 && strings.Contains(cfg.AI.AllowedIPs[0], ",") {
		parts := strings.Split(cfg.AI.AllowedIPs[0], ",")
		cfg.AI.AllowedIPs = make([]string, 0, len(parts))
		for _, p := range parts {
			if p = strings.TrimSpace(p); p != "" {
				cfg.AI.AllowedIPs = append(cfg.AI.AllowedIPs, p)
			}
		}
	}

	// Validate final config
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// findConfigDir searches for the config directory containing base.yaml.
// Search order:
//  1. UNM_CONFIG_DIR environment variable (absolute path)
//  2. Relative to working directory: config/, ../config/, ../../config/
func findConfigDir() string {
	// Check env var override first
	if dir := os.Getenv("UNM_CONFIG_DIR"); dir != "" {
		if _, err := os.Stat(filepath.Join(dir, "base.yaml")); err == nil {
			return dir
		}
		return ""
	}

	// Try relative to working directory
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}

	candidates := []string{
		filepath.Join(wd, "config"),
		filepath.Join(wd, "..", "config"),
		filepath.Join(wd, "..", "..", "config"),
	}

	for _, dir := range candidates {
		if _, err := os.Stat(filepath.Join(dir, "base.yaml")); err == nil {
			return dir
		}
	}

	return ""
}

// envKeyTransform converts a UNM_ prefixed env var name to a koanf key path.
// Double-underscore (__) is the level separator; single underscore is preserved.
//
//	UNM_SERVER__PORT          → server.port
//	UNM_AI__ALLOWED_IPS       → ai.allowed_ips
//	UNM_AI__ENABLED           → ai.enabled
func envKeyTransform(s string) string {
	s = strings.ToLower(strings.TrimPrefix(s, "UNM_"))
	parts := strings.Split(s, "__")
	return strings.Join(parts, ".")
}
