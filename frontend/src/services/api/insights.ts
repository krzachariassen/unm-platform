import type { InsightsResponse, InsightsStatusResponse } from '@/types/insights'
import { apiFetch } from './client'

export const insightsApi = {
  getInsights: (modelId: string, domain: string, signal?: AbortSignal): Promise<InsightsResponse> =>
    apiFetch(`/models/${encodeURIComponent(modelId)}/insights/${domain}`, { signal }),

  getInsightsStatus: (modelId: string, signal?: AbortSignal): Promise<InsightsStatusResponse> =>
    apiFetch(`/models/${encodeURIComponent(modelId)}/insights/status`, { signal }),
}
