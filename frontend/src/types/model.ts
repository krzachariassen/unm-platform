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
  warnings?: string[]
}

export interface ValidationItem {
  code: string
  message: string
  entity: string
}

export interface ModelListItem {
  id: string
  name: string
  created_at: string
  version_count: number
}

export interface VersionMeta {
  id: string
  version: number
  commit_message: string
  committed_at: string
}

export interface DiffEntities {
  actors: string[]
  needs: string[]
  capabilities: string[]
  services: string[]
  teams: string[]
}

export interface DiffResult {
  model_id: string
  from_version: number
  to_version: number
  added: DiffEntities
  removed: DiffEntities
  changed: DiffEntities
}
