# Configuration Reference

This document describes the UNM Platform configuration system: how config is loaded, what keys are available, how to override them, and how to manage secrets.

## Overview

The UNM Platform uses a layered configuration system built on [koanf](https://github.com/knadh/koanf). Configuration is defined in YAML files under `config/`, with environment-specific overrides and runtime overrides via environment variables. The system is designed so that `config/base.yaml` contains every key with its default value, and environment files only override what differs.

## Loading Precedence

Configuration is merged in the following order (later sources win):

| Priority | Source | Description |
|----------|--------|-------------|
| 1 (lowest) | Code defaults | `entity.DefaultConfig()` in Go — hardcoded fallback if no config files exist |
| 2 | `config/base.yaml` | Shared defaults, always loaded. Defines every config key. |
| 3 | `config/{env}.yaml` | Environment-specific overrides. Loaded when `UNM_ENV` is set. |
| 4 (highest) | `UNM_*` env vars | Runtime overrides from environment variables. |

After merging, the loader resolves secrets (see [Secret Management](#secret-management)) and validates the final config.

## Setting the Environment

Set the `UNM_ENV` environment variable to select an environment config file:

```bash
# Use production config
export UNM_ENV=production

# Use test config
export UNM_ENV=test

# Default (when unset): local
```

The CLI also accepts `--env=<environment>`.

If `UNM_ENV` is empty or unset, it defaults to `local`, loading `config/local.yaml`.

### Available Environments

| Environment | File | Purpose |
|-------------|------|---------|
| `local` | `config/local.yaml` | Local development (default). Debug routes enabled, standard AI reasoning levels. |
| `production` | `config/production.yaml` | Production deployment. Debug routes disabled, higher timeouts, JSON logging, elevated AI reasoning. |
| `test` | `config/test.yaml` | Test runs. AI disabled, debug routes off, error-level logging only. |

## Config File Location

The loader searches for a `config/` directory containing `base.yaml` in the following order:

1. **`UNM_CONFIG_DIR` environment variable** — if set, uses this as the absolute path to the config directory
2. **Relative to working directory** — tries `config/`, `../config/`, `../../config/`

This means the server works whether run from the repo root, `backend/`, or `backend/cmd/server/`.

## Full Config Schema Reference

### Server (`server`)

| Key | Type | Default | Env Var Override | Description |
|-----|------|---------|-----------------|-------------|
| `server.port` | int | `8080` | `UNM_SERVER_PORT` | HTTP server listen port (1-65535) |
| `server.host` | string | `"0.0.0.0"` | `UNM_SERVER_HOST` | HTTP server listen address |
| `server.cors_origins` | []string | `["http://localhost:5173"]` | — | Allowed CORS origins. Must not be empty. |
| `server.read_timeout` | duration | `30s` | `UNM_SERVER_READ.TIMEOUT` | HTTP read timeout |
| `server.write_timeout` | duration | `30s` | `UNM_SERVER_WRITE.TIMEOUT` | HTTP write timeout |
| `server.shutdown_timeout` | duration | `10s` | `UNM_SERVER_SHUTDOWN.TIMEOUT` | Graceful shutdown timeout |

### Frontend (`frontend`)

| Key | Type | Default | Env Var Override | Description |
|-----|------|---------|-----------------|-------------|
| `frontend.api_base_url` | string | `"/api"` | `UNM_FRONTEND_API.BASE.URL` | Base URL path for API requests |
| `frontend.dev_proxy_target` | string | `"http://localhost:8080"` | `UNM_FRONTEND_DEV.PROXY.TARGET` | Vite dev proxy target for API calls |
| `frontend.dev_server_port` | int | `5173` | `UNM_FRONTEND_DEV.SERVER.PORT` | Vite dev server port |

### AI (`ai`)

| Key | Type | Default | Env Var Override | Description |
|-----|------|---------|-----------------|-------------|
| `ai.enabled` | bool | `true` | `UNM_AI_ENABLED` | Enable/disable AI features globally |
| `ai.provider` | string | `"openai"` | `UNM_AI_PROVIDER` | AI provider name |
| `ai.api_key_env` | string | `"UNM_OPENAI_API_KEY"` | `UNM_AI_API.KEY.ENV` | Name of the env var holding the API key (see [Secret Management](#secret-management)) |
| `ai.base_url` | string | `"https://api.openai.com/v1"` | `UNM_AI_BASE.URL` | AI provider base URL |
| `ai.model` | string | `"gpt-4o"` | `UNM_AI_MODEL` | AI model name |
| `ai.request_timeout` | duration | `60s` | `UNM_AI_REQUEST.TIMEOUT` | Timeout for AI API requests. Must be positive. |

#### AI Reasoning Effort (`ai.reasoning`)

Reasoning effort controls how much "thinking" the AI model does per request category. Higher effort = better quality but higher cost and latency.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `ai.reasoning.default` | string | `"low"` | Fallback reasoning effort for uncategorized prompts |
| `ai.reasoning.advisor` | string | `"medium"` | Standard AI Advisor questions |
| `ai.reasoning.advisor_deep` | string | `"high"` | Deep-dive AI Advisor analysis |
| `ai.reasoning.whatif` | string | `"high"` | What-if scenario analysis |
| `ai.reasoning.query` | string | `"none"` | Simple query/lookup prompts |
| `ai.reasoning.summary` | string | `"low"` | Summarization prompts |

See [Reasoning Effort Levels](#reasoning-effort-levels) for valid values.

### Analysis (`analysis`)

| Key | Type | Default | Env Var Override | Description |
|-----|------|---------|-----------------|-------------|
| `analysis.default_team_size` | int | `5` | `UNM_ANALYSIS_DEFAULT.TEAM.SIZE` | Default team size when not specified. Must be >= 1. |
| `analysis.overloaded_capability_threshold` | int | `6` | `UNM_ANALYSIS_OVERLOADED.CAPABILITY.THRESHOLD` | Number of services before a capability is considered overloaded |

#### Cognitive Load Thresholds (`analysis.cognitive_load`)

Each threshold is a pair `[warning, critical]`.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `analysis.cognitive_load.domain_spread_thresholds` | [2]int | `[4, 6]` | Number of distinct capability domains a team touches |
| `analysis.cognitive_load.service_load_thresholds` | [2]float64 | `[2.0, 3.0]` | Weighted service load score per team |
| `analysis.cognitive_load.interaction_load_thresholds` | [2]int | `[4, 7]` | Number of team interactions |
| `analysis.cognitive_load.dependency_load_thresholds` | [2]int | `[5, 9]` | Number of service dependencies |

#### Interaction Weights (`analysis.interaction_weights`)

Weights used when calculating cognitive load from team interactions.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `analysis.interaction_weights.collaboration` | int | `3` | Weight for collaboration mode (highest cognitive load) |
| `analysis.interaction_weights.facilitating` | int | `2` | Weight for facilitating mode |
| `analysis.interaction_weights.x-as-a-service` | int | `1` | Weight for x-as-a-service mode (lowest cognitive load) |

#### Bottleneck Detection (`analysis.bottleneck`)

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `analysis.bottleneck.fan_in_warning` | int | `5` | Fan-in count to trigger a warning signal |
| `analysis.bottleneck.fan_in_critical` | int | `10` | Fan-in count to trigger a critical signal. Must be greater than `fan_in_warning`. |

#### Signal Detection (`analysis.signals`)

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `analysis.signals.need_team_span_warning` | int | `2` | Number of teams spanning a need before warning |
| `analysis.signals.need_team_span_critical` | int | `3` | Number of teams spanning a need before critical |
| `analysis.signals.high_span_service_threshold` | int | `3` | Services spanning this many capabilities flag as high-span |
| `analysis.signals.interaction_over_reliance` | int | `4` | Interaction count before flagging over-reliance |
| `analysis.signals.depth_chain_threshold` | int | `4` | Dependency chain depth before flagging |

#### Value Chain (`analysis.value_chain`)

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `analysis.value_chain.at_risk_team_span` | int | `3` | Team span threshold for at-risk value chains |

### Features (`features`)

| Key | Type | Default | Env Var Override | Description |
|-----|------|---------|-----------------|-------------|
| `features.debug_routes` | bool | `true` | `UNM_FEATURES_DEBUG.ROUTES` | Enable debug HTTP routes (auto-load example models) |
| `features.debug_example_paths` | []string | (see below) | — | Paths to search for example models when debug routes are enabled |

Default `debug_example_paths`:
```yaml
- "../examples/inca.unm.yaml"
- "../../examples/inca.unm.yaml"
- "examples/inca.unm.extended.yaml"
```

### Logging (`logging`)

| Key | Type | Default | Env Var Override | Description |
|-----|------|---------|-----------------|-------------|
| `logging.level` | string | `"info"` | `UNM_LOGGING_LEVEL` | Log level: `debug`, `info`, `warn`, `error` |
| `logging.format` | string | `"text"` | `UNM_LOGGING_FORMAT` | Log format: `text` (human-readable) or `json` (structured) |

## Secret Management

The configuration system uses an **indirection pattern** for secrets. Instead of putting API keys directly in config files, the YAML holds the **name** of the environment variable that contains the secret:

```yaml
# config/base.yaml
ai:
  api_key_env: "UNM_OPENAI_API_KEY"   # This is the ENV VAR NAME, not the key itself
```

At load time, the loader resolves the actual secret:

```go
cfg.AI.APIKey = os.Getenv(cfg.AI.APIKeyEnv)  // reads UNM_OPENAI_API_KEY from environment
```

This means:
- Config files never contain secrets and are safe to commit
- Secrets live in `ai.env` (gitignored) or your deployment's secret manager
- To load credentials locally: `source ai.env` before running the server or tests

See `ai.env.example` in the repo root for the template.

## Reasoning Effort Levels

AI reasoning effort controls the depth of model thinking. Valid values for any `ai.reasoning.*` key:

| Value | Description | Cost/Latency |
|-------|-------------|-------------|
| `none` | No extended reasoning. Fast, cheapest. Good for simple lookups. | Lowest |
| `low` | Minimal reasoning. Suitable for summarization and straightforward prompts. | Low |
| `medium` | Moderate reasoning. Good balance for advisory questions. | Medium |
| `high` | Deep reasoning. Used for complex analysis and what-if scenarios. | High |
| `xhigh` | Maximum reasoning depth. Reserved for production deep-dive analysis. | Highest |

Invalid values are rejected at config validation time.

## Adding a New Environment

To add a new environment (e.g., `staging`):

1. Create `config/staging.yaml` with only the keys that differ from `base.yaml`:

```yaml
# config/staging.yaml
server:
  cors_origins:
    - "https://unm-staging.internal.company.com"

ai:
  reasoning:
    default: "medium"
    advisor: "high"

logging:
  level: "info"
  format: "json"
```

2. Set the environment variable before starting the server:

```bash
export UNM_ENV=staging
```

No code changes are needed. The loader automatically picks up `config/staging.yaml`.

## Adding a New Prompt Category

To add a new reasoning effort category (e.g., for a new AI feature called `codegen`):

1. Add the key to `ai.reasoning` in `config/base.yaml`:

```yaml
ai:
  reasoning:
    # ... existing keys ...
    codegen: "high"
```

2. Override per environment if needed (e.g., in `config/production.yaml`):

```yaml
ai:
  reasoning:
    codegen: "xhigh"
```

3. Access it in Go code via `cfg.AI.Reasoning["codegen"]`.

## Environment Variable Override Convention

Any config key can be overridden at runtime via environment variables using this mapping:

- Prefix: `UNM_`
- Separator: `_` replaces `.` (YAML nesting)
- Case: uppercase

**Examples:**

| Config Key | Environment Variable |
|------------|---------------------|
| `server.port` | `UNM_SERVER_PORT` |
| `ai.enabled` | `UNM_AI_ENABLED` |
| `ai.model` | `UNM_AI_MODEL` |
| `logging.level` | `UNM_LOGGING_LEVEL` |
| `logging.format` | `UNM_LOGGING_FORMAT` |

**Note:** For deeply nested keys with underscores in the YAML key name (e.g., `ai.api_key_env`), the env var transform converts all underscores to dots, so `UNM_AI_API_KEY_ENV` maps to `ai.api.key.env` (not `ai.api_key_env`). For these keys, use the YAML config files for overrides rather than env vars.
