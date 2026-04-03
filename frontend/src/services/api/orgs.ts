import type { OrgInfo, WorkspaceInfo } from '@/types/org'
import { apiFetch } from './client'

export const orgsApi = {
  async listOrgs(signal?: AbortSignal): Promise<OrgInfo[]> {
    return apiFetch<OrgInfo[]>('/orgs', { signal })
  },

  async createOrg(name: string, signal?: AbortSignal): Promise<OrgInfo> {
    return apiFetch<OrgInfo>('/orgs', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name }),
      signal,
    })
  },

  async listWorkspaces(orgSlug: string, signal?: AbortSignal): Promise<WorkspaceInfo[]> {
    return apiFetch<WorkspaceInfo[]>(`/orgs/${encodeURIComponent(orgSlug)}/workspaces`, { signal })
  },

  async createWorkspace(
    orgSlug: string,
    name: string,
    visibility: 'private' | 'org-visible' = 'private',
    signal?: AbortSignal,
  ): Promise<WorkspaceInfo> {
    return apiFetch<WorkspaceInfo>(`/orgs/${encodeURIComponent(orgSlug)}/workspaces`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name, visibility }),
      signal,
    })
  },

  async getWorkspace(orgSlug: string, wsSlug: string, signal?: AbortSignal): Promise<WorkspaceInfo> {
    return apiFetch<WorkspaceInfo>(
      `/orgs/${encodeURIComponent(orgSlug)}/ws/${encodeURIComponent(wsSlug)}`,
      { signal },
    )
  },
}
