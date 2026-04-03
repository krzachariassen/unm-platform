import type { RiskLevel, LoadLevel } from './common'

export interface AntiPattern {
  code: string
  message: string
  severity: string
}

export interface ViewNode {
  id: string
  type: string
  label: string
  data: Record<string, unknown>
}

export interface ViewEdge {
  id: string
  source: string
  target: string
  label: string
  description?: string
}

export interface ViewResponse {
  view_type: string
  nodes: ViewNode[]
  edges: ViewEdge[]
}

// Need View
export interface NeedViewResponse {
  view_type: string
  total_needs: number
  unmapped_count: number
  groups: Array<{
    actor: { id: string; label: string }
    needs: Array<{
      need: { id: string; label: string; data: { is_mapped: boolean; outcome?: string; anti_patterns?: AntiPattern[]; team_span?: number; teams?: string[]; at_risk?: boolean; unbacked?: boolean } }
      capabilities: Array<{ id: string; label: string; data: { visibility: string } }>
    }>
  }>
}

// Capability View
export interface CapabilityViewResponse {
  view_type: string
  leaf_capability_count: number
  high_span_services: Array<{ name: string; capability_count: number }>
  fragmented_capabilities: Array<{ id: string; label: string; team_count: number }>
  parent_groups: Array<{ id: string; label: string; children: string[] }>
  capabilities: Array<{
    id: string
    label: string
    description: string
    visibility: string
    is_leaf: boolean
    is_fragmented: boolean
    depended_on_by_count: number
    services: Array<{ id: string; label: string; cap_count: number }>
    teams: Array<{ id: string; label: string; type: string }>
    depends_on: Array<{ id: string; label: string }>
    children: Array<{ id: string; label: string }>
    anti_patterns?: AntiPattern[]
    external_deps?: Array<{ name: string; description?: string }>
  }>
}

// Ownership View
export interface OwnershipViewResponse {
  view_type: string
  lanes: Array<{
    team: { id: string; label: string; data: { type: string; is_overloaded: boolean; anti_patterns?: AntiPattern[]; description?: string } }
    caps: Array<{
      cap: { id: string; label: string; data: { visibility: string; is_leaf: boolean; anti_patterns?: AntiPattern[]; description?: string } }
      services: Array<{ id: string; label: string; team_id: string; team_label: string; cap_count: number }>
      cross_team: boolean
    }>
    external_deps?: Array<{ id: string; label: string; description: string; service_count: number }>
  }>
  unowned_capabilities: Array<{ id: string; label: string; data: { visibility: string } }>
  service_rows: Array<{
    service: { id: string; label: string }
    team: { id: string; label: string; data: { type: string } } | null
    capabilities: Array<{ id: string; label: string; data: { visibility: string } }>
  }>
  cross_team_capabilities: Array<{ cap_id: string; cap_label: string; team_labels: string[] }>
  high_span_services: Array<{ name: string; capability_count: number }>
  overloaded_teams: Array<{ id: string; label: string }>
  no_cap_count: number
  multi_cap_count: number
  external_dependency_count?: number
}

// Team Topology View
export interface TeamTopologyViewResponse {
  view_type: string
  teams: Array<{
    id: string
    label: string
    description: string
    type: string
    is_overloaded: boolean
    capability_count: number
    service_count: number
    interactions: Array<{ source_id: string; target_id: string; mode: string; via: string; description: string }>
    anti_patterns?: AntiPattern[]
    services: string[]
    capabilities: string[]
  }>
  interactions: Array<{ source_id: string; target_id: string; mode: string; via: string; description: string }>
}

// Cognitive Load View
export interface LoadDimension {
  value: number
  level: LoadLevel
}

export interface TeamLoad {
  team: { name: string; type: string }
  capability_count: number
  service_count: number
  dependency_count: number
  interaction_count: number
  interaction_score: number
  team_size: number
  size_is_explicit: boolean
  domain_spread: LoadDimension
  service_load: LoadDimension
  interaction_load: LoadDimension
  dependency_load: LoadDimension
  overall_level: LoadLevel
  services: string[]
  capabilities: string[]
}

export interface CognitiveLoadViewResponse {
  view_type: string
  team_loads: TeamLoad[]
}

// Signals View
export type SignalSourceType = 'model_fact' | 'analyzer_finding' | 'ai_interpretation'

export interface SignalsNeedRisk {
  need_name: string
  actor_names: string[]
  team_span: number
  teams: string[]
  source?: SignalSourceType
  explanation?: string
}

export interface SignalsCapItem {
  capability_name: string
  visibility?: string
  team_count?: number
  teams?: string[]
  source?: SignalSourceType
  explanation?: string
}

export interface SignalsTeamItem {
  team_name: string
  team_type?: string
  overall_level?: string
  capability_count?: number
  service_count?: number
  coherence_score?: number
  source?: SignalSourceType
  explanation?: string
}

export interface SignalsServiceItem {
  service_name: string
  fan_in: number
  source?: SignalSourceType
  explanation?: string
}

export interface SignalsExtDepItem {
  dep_name: string
  service_count: number
  services: string[]
  is_critical: boolean
  is_warning: boolean
  source?: SignalSourceType
  explanation?: string
}

export interface SignalsViewResponse {
  view_type: string
  health: {
    ux_risk: RiskLevel
    architecture_risk: RiskLevel
    org_risk: RiskLevel
  }
  user_experience_layer: {
    needs_requiring_3plus_teams: SignalsNeedRisk[]
    needs_with_no_capability_backing: SignalsNeedRisk[]
    needs_at_risk: SignalsNeedRisk[]
  }
  architecture_layer: {
    user_facing_caps_with_cross_team_services: SignalsCapItem[]
    capabilities_not_connected_to_any_need: SignalsCapItem[]
    capabilities_fragmented_across_teams: SignalsCapItem[]
  }
  organization_layer: {
    top_teams_by_structural_load: SignalsTeamItem[]
    critical_bottleneck_services: SignalsServiceItem[]
    low_coherence_teams: SignalsTeamItem[]
    critical_external_deps?: SignalsExtDepItem[]
  }
}

// Realization View
export interface RealizationViewResponse {
  view_type: string
  no_cap_count: number
  multi_cap_count: number
  service_rows: Array<{
    service: { id: string; label: string; description?: string }
    team: { id: string; label: string; data: { type: string } } | null
    capabilities: Array<{ id: string; label: string; data: { visibility: string } }>
    external_deps?: string[]
  }>
}

// UNM Map View
export interface UNMMapExtDep {
  id: string
  name: string
  description?: string
  service_count: number
  services: string[]
  is_critical: boolean
  is_warning: boolean
}

export interface UNMMapViewResponse {
  view_type: string
  nodes: ViewNode[]
  edges: ViewEdge[]
  external_deps?: UNMMapExtDep[]
}

// Gaps View (FC.4)
export interface GapsView {
  model_id: string
  unmapped_needs: string[]
  unrealized_capabilities: string[]
  unowned_services: string[]
  unneeded_capabilities: string[]
  orphan_services: string[]
}

// Dependencies View (FC.5)
export interface DependenciesView {
  model_id: string
  service_cycles: Array<{ path: string[] }>
  capability_cycles: Array<{ path: string[] }>
  max_service_depth: number
  max_capability_depth: number
  critical_service_path: string[]
}

// Interactions View (FC.6)
export interface InteractionsView {
  model_id: string
  mode_distribution: Record<string, number>
  isolated_teams: string[]
  over_reliant_teams: Array<{ team_name: string; mode: string; count: number }>
  all_modes_same: boolean
}
