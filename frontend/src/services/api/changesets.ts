import type { ChangeAction, ChangesetSummary, Changeset, ImpactResponse, CommitResponse } from '@/types/changeset'
import { apiFetch, apiFetchAllowConflict } from './client'

export const changesetsApi = {
  createChangeset: (
    modelId: string,
    body: { id: string; description: string; actions: ChangeAction[] }
  ): Promise<ChangesetSummary> =>
    apiFetch(`/models/${modelId}/changesets`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    }),

  getChangeset: (modelId: string, changesetId: string): Promise<Changeset> =>
    apiFetch(`/models/${modelId}/changesets/${changesetId}`),

  getImpact: (modelId: string, changesetId: string): Promise<ImpactResponse> =>
    apiFetch(`/models/${modelId}/changesets/${changesetId}/impact`),

  applyChangeset: (modelId: string, changesetId: string): Promise<Record<string, unknown>> =>
    apiFetch(`/models/${modelId}/changesets/${changesetId}/apply`, { method: 'POST' }),

  commitChangeset: (modelId: string, changesetId: string): Promise<CommitResponse> =>
    apiFetchAllowConflict(`/models/${modelId}/changesets/${changesetId}/commit`, { method: 'POST' }),
}
