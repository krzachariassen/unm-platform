import type { AdvisorResponse } from '@/types/insights'
import { apiFetch } from './client'

export const advisorApi = {
  ask: (modelId: string, question: string, category?: string): Promise<AdvisorResponse> =>
    apiFetch(`/models/${modelId}/ask`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ question, category: category ?? 'general' }),
    }),
}
