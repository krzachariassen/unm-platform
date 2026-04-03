/**
 * Minimal auth-context stub for Phase 15B.
 * Phase 15A will provide the real implementation when merged.
 * This stub allows 15B components to import from @/lib/auth-context
 * without a hard dependency on the 15A branch.
 */
import { createContext, useContext, type ReactNode } from 'react'

export interface AuthUser {
  id: string
  name: string
  email: string
  avatar_url?: string
}

interface AuthContextValue {
  user: AuthUser | null
  loading: boolean
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue>({
  user: null,
  loading: false,
  logout: async () => {},
})

export function AuthProvider({ children }: { children: ReactNode }) {
  // Stub: pass through. Phase 15A replaces this with the real implementation.
  return <AuthContext.Provider value={{ user: null, loading: false, logout: async () => {} }}>{children}</AuthContext.Provider>
}

export function useAuth() {
  return useContext(AuthContext)
}
