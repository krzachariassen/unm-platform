export interface InsightItem {
  explanation: string
  suggestion: string
}

export interface InsightsStatusResponse {
  domains: Record<string, string>
  all_ready: boolean
}

export interface InsightsResponse {
  domain: string
  status?: string
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
