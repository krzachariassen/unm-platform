export interface RuntimeConfig {
  ai: {
    enabled: boolean;
  };
  features: {
    debug_routes: boolean;
  };
  analysis: {
    default_team_size: number;
    overloaded_capability_threshold: number;
    bottleneck: { fan_in_warning: number; fan_in_critical: number };
    signals: {
      need_team_span_warning: number;
      need_team_span_critical: number;
      high_span_service_threshold: number;
      interaction_over_reliance: number;
      depth_chain_threshold: number;
    };
    value_chain: { at_risk_team_span: number };
  };
}

let cached: RuntimeConfig | null = null;

export async function getRuntimeConfig(): Promise<RuntimeConfig> {
  if (cached) return cached;
  const res = await fetch('/api/config');
  cached = await res.json();
  return cached!;
}
