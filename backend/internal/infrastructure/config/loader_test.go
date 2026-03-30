package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// configDir walks up from the current working directory to find the config/ directory
// containing base.yaml.
func configDir(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)
	for i := 0; i < 5; i++ {
		candidate := filepath.Join(dir, "config")
		if _, err := os.Stat(filepath.Join(candidate, "base.yaml")); err == nil {
			return candidate
		}
		dir = filepath.Dir(dir)
	}
	t.Fatal("could not find config/base.yaml")
	return ""
}

func setConfigDir(t *testing.T) {
	t.Helper()
	t.Setenv("UNM_CONFIG_DIR", configDir(t))
}

func TestLoadConfig_DefaultsFromBase(t *testing.T) {
	setConfigDir(t)

	cfg, err := LoadConfig("local")
	require.NoError(t, err)

	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "gpt-5.4-nano", cfg.AI.Model)
	assert.Equal(t, 5, cfg.Analysis.DefaultTeamSize)
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "text", cfg.Logging.Format)
	assert.True(t, cfg.Features.DebugRoutes)
}

func TestLoadConfig_ProductionOverrides(t *testing.T) {
	setConfigDir(t)

	cfg, err := LoadConfig("production")
	require.NoError(t, err)

	assert.False(t, cfg.Features.DebugRoutes)
	assert.Equal(t, "json", cfg.Logging.Format)
	assert.Equal(t, "warn", cfg.Logging.Level)
	assert.Equal(t, []string{"https://unm.internal.company.com"}, cfg.Server.CORSOrigins)
}

func TestLoadConfig_TestEnvDisablesAI(t *testing.T) {
	setConfigDir(t)

	cfg, err := LoadConfig("test")
	require.NoError(t, err)

	assert.False(t, cfg.AI.Enabled)
}

func TestLoadConfig_EnvVarOverridesFile(t *testing.T) {
	setConfigDir(t)
	t.Setenv("UNM_SERVER__PORT", "9999") // double-underscore = level separator

	cfg, err := LoadConfig("local")
	require.NoError(t, err)

	assert.Equal(t, 9999, cfg.Server.Port)
}

// ---------------------------------------------------------------------------
// envKeyTransform unit tests
// ---------------------------------------------------------------------------

func TestEnvKeyTransform_SimpleKey(t *testing.T) {
	assert.Equal(t, "server.port", envKeyTransform("UNM_SERVER__PORT"))
}

func TestEnvKeyTransform_UnderscoreInKeyName(t *testing.T) {
	assert.Equal(t, "ai.allowed_ips", envKeyTransform("UNM_AI__ALLOWED_IPS"))
}

func TestEnvKeyTransform_NestedWithUnderscore(t *testing.T) {
	assert.Equal(t, "analysis.bottleneck.fan_in_warning", envKeyTransform("UNM_ANALYSIS__BOTTLENECK__FAN_IN_WARNING"))
}

func TestEnvKeyTransform_SingleSegment(t *testing.T) {
	assert.Equal(t, "ai.enabled", envKeyTransform("UNM_AI__ENABLED"))
}

// ---------------------------------------------------------------------------
// Integration: env var sets AllowedIPs via double-underscore convention
// ---------------------------------------------------------------------------

func TestLoadConfig_EnvVar_AllowedIPs(t *testing.T) {
	setConfigDir(t)
	t.Setenv("UNM_AI__ALLOWED_IPS", "1.2.3.4,10.0.0.0/8")

	cfg, err := LoadConfig("local")
	require.NoError(t, err)

	assert.Equal(t, []string{"1.2.3.4", "10.0.0.0/8"}, cfg.AI.AllowedIPs)
}


func TestLoadConfig_MissingBaseFails(t *testing.T) {
	t.Setenv("UNM_CONFIG_DIR", "/nonexistent/path")

	_, err := LoadConfig("local")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "base.yaml")
}

func TestLoadConfig_EmptyEnvDefaultsToLocal(t *testing.T) {
	setConfigDir(t)

	cfg, err := LoadConfig("")
	require.NoError(t, err)

	// local.yaml has debug_routes: true
	assert.True(t, cfg.Features.DebugRoutes)
}

func TestLoadConfig_ResolvesAPIKey(t *testing.T) {
	setConfigDir(t)
	t.Setenv("UNM_OPENAI_API_KEY", "test-key-12345")

	cfg, err := LoadConfig("local")
	require.NoError(t, err)

	assert.Equal(t, "test-key-12345", cfg.AI.APIKey)
}

func TestLoadConfig_MissingEnvFileDoesNotFail(t *testing.T) {
	setConfigDir(t)

	// "nonexistent" environment has no YAML file — should still work with base defaults
	cfg, err := LoadConfig("nonexistent")
	require.NoError(t, err)

	assert.Equal(t, 8080, cfg.Server.Port)
}
