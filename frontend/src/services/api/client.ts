import { config } from '@/lib/config'

const BASE = config.apiBaseUrl

async function extractError(res: Response, fallback: string): Promise<string> {
  try {
    const body = await res.json()
    return body?.error ?? fallback
  } catch {
    return `${fallback} (${res.status} ${res.statusText})`
  }
}

export async function apiFetch<T>(
  path: string,
  options?: RequestInit & { signal?: AbortSignal }
): Promise<T> {
  const res = await fetch(`${BASE}${path}`, options)
  if (!res.ok) throw new Error(await extractError(res, `Request failed: ${path}`))
  return res.json() as Promise<T>
}

export async function apiFetchText(
  path: string,
  options?: RequestInit & { signal?: AbortSignal }
): Promise<string> {
  const res = await fetch(`${BASE}${path}`, options)
  if (!res.ok) throw new Error(await extractError(res, `Request failed: ${path}`))
  return res.text()
}

// Special case: 409 returns JSON with valid:false (not an error)
export async function apiFetchAllowConflict<T>(
  path: string,
  options?: RequestInit & { signal?: AbortSignal }
): Promise<T> {
  const res = await fetch(`${BASE}${path}`, options)
  if (res.status === 409) return res.json() as Promise<T>
  if (!res.ok) throw new Error(await extractError(res, `Request failed: ${path}`))
  return res.json() as Promise<T>
}

/**
 * Workspace-scoped fetch helper.
 * Prepends /orgs/{orgSlug}/ws/{wsSlug} to workspace-scoped API calls.
 * Use this for endpoints that operate within a specific org + workspace context.
 */
export async function workspacedFetch<T>(
  orgSlug: string,
  wsSlug: string,
  path: string,
  options?: RequestInit & { signal?: AbortSignal }
): Promise<T> {
  const prefix = `/orgs/${encodeURIComponent(orgSlug)}/ws/${encodeURIComponent(wsSlug)}`
  return apiFetch<T>(`${prefix}${path}`, options)
}
