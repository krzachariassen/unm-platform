import type { ParseResponse, ModelListItem, VersionMeta, DiffResult } from '@/types/model'
import { apiFetch, apiFetchText } from './client'
import { config } from '@/lib/config'

const BASE = config.apiBaseUrl

export const modelsApi = {
  async parseModel(content: string, previousModelId?: string, format?: 'dsl' | 'yaml'): Promise<ParseResponse> {
    const headers: Record<string, string> = { 'Content-Type': 'application/yaml' }
    if (previousModelId) headers['X-Replace-Model'] = previousModelId
    const url = format === 'dsl'
      ? `${BASE}/models/parse?format=dsl`
      : `${BASE}/models/parse`
    const res = await fetch(url, { method: 'POST', headers, body: content })
    if (!res.ok) {
      let msg = 'Parse failed'
      try { const b = await res.json(); msg = b?.error ?? msg } catch { /* ignore */ }
      throw new Error(msg)
    }
    return res.json()
  },

  async loadExample(previousModelId?: string): Promise<ParseResponse> {
    const headers: Record<string, string> = {}
    if (previousModelId) headers['X-Replace-Model'] = previousModelId
    const res = await fetch(`${BASE}/debug/load-example`, { method: 'POST', headers })
    if (!res.ok) throw new Error('Failed to load example')
    return res.json()
  },

  async exportModel(modelId: string, format?: 'yaml' | 'dsl'): Promise<string> {
    const query = format === 'dsl' ? '?format=dsl' : ''
    return apiFetchText(`/models/${encodeURIComponent(modelId)}/export${query}`)
  },

  async getTeams(modelId: string): Promise<{ teams: Array<{ name: string; type: string }> }> {
    const res = await fetch(`${BASE}/models/${encodeURIComponent(modelId)}/teams`)
    if (!res.ok) throw new Error('Failed to fetch teams')
    return res.json()
  },

  async getServices(modelId: string): Promise<{ services: Array<{ name: string }> }> {
    const res = await fetch(`${BASE}/models/${encodeURIComponent(modelId)}/services`)
    if (!res.ok) throw new Error('Failed to fetch services')
    return res.json()
  },

  async listModels(): Promise<{ models: ModelListItem[]; total: number }> {
    return apiFetch('/models')
  },

  async loadStoredModel(modelId: string, version: number): Promise<ParseResponse> {
    return apiFetch(`/models/${encodeURIComponent(modelId)}/versions/${version}`)
  },

  async getHistory(modelId: string): Promise<{ model_id: string; versions: VersionMeta[] }> {
    return apiFetch(`/models/${encodeURIComponent(modelId)}/history`)
  },

  async getDiff(modelId: string, from: number, to: number): Promise<DiffResult> {
    return apiFetch(`/models/${encodeURIComponent(modelId)}/diff?from=${from}&to=${to}`)
  },
}
