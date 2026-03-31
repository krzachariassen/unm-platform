import type { AdvisorResponse } from '@/types/insights'
import type { ExtractActionsResponse } from '@/types/changeset'
import { apiFetch } from './client'

export const advisorApi = {
  ask: (modelId: string, question: string, category?: string): Promise<AdvisorResponse> =>
    apiFetch(`/models/${modelId}/ask`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ question, category: category ?? 'general' }),
    }),

  extractActions: (modelId: string, advisorResponse: string, signal?: AbortSignal): Promise<ExtractActionsResponse> =>
    apiFetch(`/models/${modelId}/extract-actions`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ advisor_response: advisorResponse }),
      signal,
    }),
}
