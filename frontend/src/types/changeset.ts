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
  service_name?: string
  from_team_name?: string
  to_team_name?: string
  original_team_name?: string
  new_team_a_name?: string
  new_team_b_name?: string
  service_assignment?: Record<string, string>
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
