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
