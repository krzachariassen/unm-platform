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

	// Load environment variables: UNM_SERVER_PORT → server.port
	if err := k.Load(env.Provider("UNM_", ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(s, "UNM_")), "_", ".")
	}), nil); err != nil {
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
