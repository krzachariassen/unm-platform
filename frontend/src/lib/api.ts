import { config } from './config'

const API_BASE = config.apiBaseUrl

// Safely extract an error message from a non-OK response.
// Falls back to HTTP status text if the body is not JSON.
async function extractError(res: Response, fallback: string): Promise<string> {
  try {
    const body = await res.json()
    return body?.error ?? fallback
  } catch {
    return `${fallback} (${res.status} ${res.statusText})`
  }
}

export interface ParseResponse {
  id: string
  system_name: string
  system_description: string
  summary: {
    actors: number
    needs: number
    capabilities: number
    services: number
    teams: number
  }
  validation: {
    is_valid: boolean
    errors: ValidationItem[]
    warnings: ValidationItem[]
  }
  warnings?: string[]  // reference validation warnings from parser
}

export interface ValidationItem {
  code: string
  message: string
  entity: string
}

export interface Capability {
  name: string
  description: string
  visibility: string
  is_leaf: boolean
}

export interface Team {
  name: string
  type: string
  capability_count: number
  is_overloaded: boolean
}

export interface Need {
  name: string
  actor_name: string
  is_mapped: boolean
}

export interface Service {
  name: string
  description: string
  owner_team_name: string
}

export interface Actor {
  name: string
  description: string
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

export interface LoadDimension {
  value: number
  level: 'low' | 'medium' | 'high'
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
  overall_level: 'low' | 'medium' | 'high'
  services: string[]
  capabilities: string[]
}

export interface AntiPattern {
  code: string
  message: string
  severity: string
}

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

export interface CapabilityViewResponse {
  view_type: string
  leaf_capability_count: number
  high_span_services: Array<{ name: string; capability_count: number }>
  fragmented_capabilities: Array<{ id: string; label: string; team_count: number }>
  parent_groups: Array<{
    id: string
    label: string
    children: string[]
  }>
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

export interface CognitiveLoadViewResponse {
  view_type: string
  team_loads: TeamLoad[]
}

export interface SignalsNeedRisk {
  need_name: string
  actor_name: string
  team_span: number
  teams: string[]
}

export interface SignalsCapItem {
  capability_name: string
  visibility?: string
  team_count?: number
  teams?: string[]
}

export interface SignalsTeamItem {
  team_name: string
  team_type?: string
  overall_level?: string
  capability_count?: number
  service_count?: number
  coherence_score?: number
}

export interface SignalsServiceItem {
  service_name: string
  fan_in: number
}

export interface SignalsExtDepItem {
  dep_name: string
  service_count: number
  services: string[]
  is_critical: boolean
  is_warning: boolean
}

export interface SignalsViewResponse {
  view_type: string
  health: {
    ux_risk: 'red' | 'amber' | 'green'
    architecture_risk: 'red' | 'amber' | 'green'
    org_risk: 'red' | 'amber' | 'green'
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

export interface ChangeAction {
  type:
    | 'move_service'
    | 'split_team'
    | 'merge_teams'
    | 'add_capability'
    | 'remove_capability'
    | 'reassign_capability'
    | 'add_interaction'
    | 'remove_interaction'
    | 'update_team_size'
    | 'add_service'
    | 'remove_service'
    | 'rename_service'
    | 'add_team'
    | 'remove_team'
    | 'update_team_type'
    | 'add_need'
    | 'remove_need'
    | 'add_actor'
    | 'remove_actor'
    | 'add_service_dependency'
    | 'remove_service_dependency'
    | 'link_need_capability'
    | 'unlink_need_capability'
    | 'link_capability_service'
    | 'unlink_capability_service'
    | 'update_capability_visibility'
    | 'update_description'
  // Flat fields — set whichever apply to the action type
  service_name?: string
  from_team_name?: string
  to_team_name?: string
  original_team_name?: string
  new_team_a_name?: string
  new_team_b_name?: string
  service_assignment?: Record<string, string> // service → "a" | "b"
  team_a_name?: string
  team_b_name?: string
  new_team_name?: string
  capability_name?: string
  owner_team_name?: string
  source_team_name?: string
  target_team_name?: string
  interaction_mode?: string
  team_name?: string
  new_size?: number
  new_service_name?: string
  description?: string
  team_type?: string
  need_name?: string
  actor_name?: string
  outcome?: string
  supported_by?: string[]
  depends_on_service?: string
  role?: string
  visibility?: string
  entity_type?: string
  entity_name?: string
}

export interface ChangesetSummary {
  id: string
  model_id: string
  description: string
  action_count: number
}

export interface Changeset {
  id: string
  model_id: string
  description: string
  actions: ChangeAction[]
  created_at: string
}

export interface ImpactDelta {
  dimension: string
  before: string
  after: string
  change: 'improved' | 'regressed' | 'unchanged'
  detail: string
}

export interface ImpactResponse {
  changeset_id: string
  deltas: ImpactDelta[]
}

export interface CommitResponse {
  model_id: string
  system_name: string
  summary: Record<string, number>
  validation: {
    valid: boolean
    errors?: string[]
    warnings?: string[]
  }
}

export interface InsightItem {
  explanation: string
  suggestion: string
}

export interface InsightsStatusResponse {
  domains: Record<string, string> // domain → "pending" | "computing" | "ready" | "failed"
  all_ready: boolean
}

export interface InsightsResponse {
  domain: string
  status?: string // "computing" | "ready" | "failed"
  insights: Record<string, InsightItem>
  ai_configured: boolean
}

export interface AdvisorResponse {
  model_id: string
  category: string
  question: string
  answer: string
  finish_reason?: string
  ai_configured: boolean
}

export const api = {
  async parseModel(yaml: string, previousModelId?: string): Promise<ParseResponse> {
    const headers: Record<string, string> = { 'Content-Type': 'application/yaml' }
    if (previousModelId) headers['X-Replace-Model'] = previousModelId
    const res = await fetch(`${API_BASE}/models/parse`, {
      method: 'POST',
      headers,
      body: yaml,
    })
    if (!res.ok) throw new Error(await extractError(res, 'Parse failed'))
    return res.json()
  },

  async getCapabilities(id: string): Promise<{ capabilities: Capability[] }> {
    const res = await fetch(`${API_BASE}/models/${id}/capabilities`)
    if (!res.ok) throw new Error(await extractError(res, 'Failed to fetch capabilities'))
    return res.json()
  },

  async getTeams(id: string): Promise<{ teams: Team[] }> {
    const res = await fetch(`${API_BASE}/models/${id}/teams`)
    if (!res.ok) throw new Error(await extractError(res, 'Failed to fetch teams'))
    return res.json()
  },

  async getNeeds(id: string): Promise<{ needs: Need[] }> {
    const res = await fetch(`${API_BASE}/models/${id}/needs`)
    if (!res.ok) throw new Error(await extractError(res, 'Failed to fetch needs'))
    return res.json()
  },

  async getServices(id: string): Promise<{ services: Service[] }> {
    const res = await fetch(`${API_BASE}/models/${id}/services`)
    if (!res.ok) throw new Error(await extractError(res, 'Failed to fetch services'))
    return res.json()
  },

  async getActors(id: string): Promise<{ actors: Actor[] }> {
    const res = await fetch(`${API_BASE}/models/${id}/actors`)
    if (!res.ok) throw new Error(await extractError(res, 'Failed to fetch actors'))
    return res.json()
  },

  async getView(id: string, viewType: string): Promise<ViewResponse> {
    const res = await fetch(`${API_BASE}/models/${id}/views/${viewType}`)
    if (!res.ok) throw new Error(await extractError(res, `Failed to fetch ${viewType} view`))
    return res.json()
  },

  async loadExample(previousModelId?: string): Promise<ParseResponse> {
    const headers: Record<string, string> = {}
    if (previousModelId) headers['X-Replace-Model'] = previousModelId
    const res = await fetch(`${API_BASE}/debug/load-example`, { method: 'POST', headers })
    if (!res.ok) throw new Error(await extractError(res, 'Failed to load example'))
    return res.json()
  },

  async getSignals(id: string): Promise<SignalsViewResponse> {
    const res = await fetch(`${API_BASE}/models/${id}/views/signals`)
    if (!res.ok) throw new Error(await extractError(res, 'Failed to fetch signals'))
    return res.json()
  },

  async createChangeset(modelId: string, body: { id: string; description: string; actions: ChangeAction[] }): Promise<ChangesetSummary> {
    const res = await fetch(`${API_BASE}/models/${modelId}/changesets`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    })
    if (!res.ok) throw new Error(await extractError(res, 'Failed to create changeset'))
    return res.json()
  },

  async getChangeset(modelId: string, changesetId: string): Promise<Changeset> {
    const res = await fetch(`${API_BASE}/models/${modelId}/changesets/${changesetId}`)
    if (!res.ok) throw new Error(await extractError(res, 'Failed to fetch changeset'))
    return res.json()
  },

  async getChangesetImpact(modelId: string, changesetId: string): Promise<ImpactResponse> {
    const res = await fetch(`${API_BASE}/models/${modelId}/changesets/${changesetId}/impact`)
    if (!res.ok) throw new Error(await extractError(res, 'Failed to fetch impact'))
    return res.json()
  },

  async askAdvisor(modelId: string, question: string, category?: string): Promise<AdvisorResponse> {
    const res = await fetch(`${API_BASE}/models/${modelId}/ask`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ question, category: category ?? 'general' }),
    })
    if (!res.ok) throw new Error(await extractError(res, 'Advisor request failed'))
    return res.json()
  },

  async applyChangeset(modelId: string, changesetId: string): Promise<Record<string, unknown>> {
    const res = await fetch(`${API_BASE}/models/${modelId}/changesets/${changesetId}/apply`, {
      method: 'POST',
    })
    if (!res.ok) throw new Error(await extractError(res, 'Failed to apply changeset'))
    return res.json()
  },

  async commitChangeset(modelId: string, changesetId: string): Promise<CommitResponse> {
    const res = await fetch(`${API_BASE}/models/${modelId}/changesets/${changesetId}/commit`, {
      method: 'POST',
    })
    if (!res.ok) throw new Error(await extractError(res, 'Failed to commit changeset'))
    return res.json()
  },

  async exportModel(modelId: string): Promise<string> {
    const res = await fetch(`${API_BASE}/models/${modelId}/export`)
    if (!res.ok) throw new Error(await extractError(res, 'Failed to export model'))
    return res.text()
  },

  async getInsights(modelId: string, domain: string): Promise<InsightsResponse> {
    const res = await fetch(`${API_BASE}/models/${encodeURIComponent(modelId)}/insights/${domain}`)
    if (!res.ok) throw new Error(await extractError(res, `HTTP ${res.status}`))
    return res.json()
  },

  async getInsightsStatus(modelId: string): Promise<InsightsStatusResponse> {
    const res = await fetch(`${API_BASE}/models/${encodeURIComponent(modelId)}/insights/status`)
    if (!res.ok) throw new Error(await extractError(res, `HTTP ${res.status}`))
    return res.json()
  },

  async getNeedView(modelId: string): Promise<NeedViewResponse> {
    const r = await fetch(`/api/models/${encodeURIComponent(modelId)}/views/need`)
    if (!r.ok) throw new Error(await extractError(r, 'Failed to fetch need view'))
    return r.json()
  },

  async getCapabilityView(modelId: string): Promise<CapabilityViewResponse> {
    const r = await fetch(`/api/models/${encodeURIComponent(modelId)}/views/capability`)
    if (!r.ok) throw new Error(await extractError(r, 'Failed to fetch capability view'))
    return r.json()
  },

  async getOwnershipView(modelId: string): Promise<OwnershipViewResponse> {
    const r = await fetch(`${API_BASE}/models/${encodeURIComponent(modelId)}/views/ownership`)
    if (!r.ok) throw new Error(await extractError(r, 'Failed to fetch ownership view'))
    return r.json()
  },

  async getTeamTopologyView(modelId: string): Promise<TeamTopologyViewResponse> {
    const r = await fetch(`${API_BASE}/models/${encodeURIComponent(modelId)}/views/team-topology`)
    if (!r.ok) throw new Error(await extractError(r, 'Failed to fetch team topology view'))
    return r.json()
  },

  async getCognitiveLoadView(modelId: string): Promise<CognitiveLoadViewResponse> {
    const r = await fetch(`${API_BASE}/models/${encodeURIComponent(modelId)}/views/cognitive-load`)
    if (!r.ok) throw new Error(await extractError(r, 'Failed to fetch cognitive load view'))
    return r.json()
  },

  async getRealizationView(modelId: string): Promise<RealizationViewResponse> {
    const r = await fetch(`${API_BASE}/models/${encodeURIComponent(modelId)}/views/realization`)
    if (!r.ok) throw new Error(await extractError(r, 'Failed to fetch realization view'))
    return r.json()
  },

  async getUNMMapView(modelId: string): Promise<UNMMapViewResponse> {
    const r = await fetch(`${API_BASE}/models/${encodeURIComponent(modelId)}/views/unm-map`)
    if (!r.ok) throw new Error(await extractError(r, 'Failed to fetch UNM map view'))
    return r.json()
  },
}
