import { config } from '@/lib/config'

export interface OrgMembership {
  id: string
  name: string
  slug: string
  role: string
}

export interface AuthUser {
  id: string
  email: string
  name: string
  avatar_url: string
  orgs: OrgMembership[]
}

const BASE = config.apiBaseUrl

export const authApi = {
  async getMe(signal?: AbortSignal): Promise<AuthUser | null> {
    const res = await fetch(`${BASE}/me`, { signal, credentials: 'include' })
    if (res.status === 401) return null
    if (!res.ok) return null
    return res.json() as Promise<AuthUser>
  },

  async logout(): Promise<void> {
    await fetch('/auth/logout', {
      method: 'POST',
      credentials: 'include',
    })
  },

  async devLogin(): Promise<boolean> {
    const res = await fetch('/auth/dev-login', {
      method: 'POST',
      credentials: 'include',
    })
    return res.ok
  },

  loginUrl(): string {
    return `${BASE.replace('/api', '')}/auth/google`
  },
}
