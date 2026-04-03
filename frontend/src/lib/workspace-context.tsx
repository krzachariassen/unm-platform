import {
  createContext,
  useContext,
  useState,
  useCallback,
  type ReactNode,
} from 'react'
import { useQuery } from '@tanstack/react-query'
import { orgsApi } from '@/services/api/orgs'
import type { OrgInfo, WorkspaceInfo } from '@/types/org'

interface WorkspaceContextValue {
  /** Current org slug — null if user has no orgs or none selected */
  orgSlug: string | null
  /** Current workspace slug — null if none selected */
  wsSlug: string | null
  /** Current org info (resolved from orgSlug) */
  org: OrgInfo | null
  /** Current workspace info (resolved from orgSlug + wsSlug) */
  workspace: WorkspaceInfo | null
  /** All orgs accessible to the current user */
  orgs: OrgInfo[]
  /** Workspaces in the current org */
  workspaces: WorkspaceInfo[]
  /** Switch to a different org + workspace */
  setWorkspace: (orgSlug: string, wsSlug: string) => void
  /** True while loading orgs or workspace list */
  loading: boolean
  /** True if the user has no orgs at all */
  hasNoOrgs: boolean
}

const WorkspaceContext = createContext<WorkspaceContextValue | null>(null)

const LS_ORG_SLUG = 'unm_org_slug'
const LS_WS_SLUG = 'unm_ws_slug'

export function WorkspaceProvider({ children }: { children: ReactNode }) {
  const [orgSlug, setOrgSlug] = useState<string | null>(
    () => localStorage.getItem(LS_ORG_SLUG),
  )
  const [wsSlug, setWsSlug] = useState<string | null>(
    () => localStorage.getItem(LS_WS_SLUG),
  )

  // Load all orgs
  const { data: orgs = [], isLoading: orgsLoading } = useQuery({
    queryKey: ['orgs'],
    queryFn: () => orgsApi.listOrgs(),
    staleTime: 5 * 60 * 1000,
  })

  // Auto-select first org if none stored
  const resolvedOrgSlug: string | null = (() => {
    if (orgSlug && orgs.some(o => o.slug === orgSlug)) return orgSlug
    if (orgs.length > 0) return orgs[0].slug
    return null
  })()

  // Load workspaces for the resolved org
  const { data: workspaces = [], isLoading: wsLoading } = useQuery({
    queryKey: ['workspaces', resolvedOrgSlug],
    queryFn: () => orgsApi.listWorkspaces(resolvedOrgSlug!),
    enabled: !!resolvedOrgSlug,
    staleTime: 5 * 60 * 1000,
  })

  // Auto-select first workspace if none stored
  const resolvedWsSlug: string | null = (() => {
    if (wsSlug && workspaces.some(w => w.slug === wsSlug)) return wsSlug
    if (workspaces.length > 0) return workspaces[0].slug
    return null
  })()

  const org = orgs.find(o => o.slug === resolvedOrgSlug) ?? null
  const workspace = workspaces.find(w => w.slug === resolvedWsSlug) ?? null

  const setWorkspace = useCallback((newOrgSlug: string, newWsSlug: string) => {
    setOrgSlug(newOrgSlug)
    setWsSlug(newWsSlug)
    localStorage.setItem(LS_ORG_SLUG, newOrgSlug)
    localStorage.setItem(LS_WS_SLUG, newWsSlug)
  }, [])

  const loading = orgsLoading || wsLoading
  const hasNoOrgs = !orgsLoading && orgs.length === 0

  return (
    <WorkspaceContext.Provider
      value={{
        orgSlug: resolvedOrgSlug,
        wsSlug: resolvedWsSlug,
        org,
        workspace,
        orgs,
        workspaces,
        setWorkspace,
        loading,
        hasNoOrgs,
      }}
    >
      {children}
    </WorkspaceContext.Provider>
  )
}

export function useWorkspace() {
  const ctx = useContext(WorkspaceContext)
  if (!ctx) throw new Error('useWorkspace must be used within WorkspaceProvider')
  return ctx
}
