import { useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Upload, Box, AlertCircle, Loader2 } from 'lucide-react'
import { useWorkspace } from '@/lib/workspace-context'
import { workspacedFetch } from '@/services/api/client'
import type { ModelListItem } from '@/types/model'
import { cn } from '@/lib/utils'

interface WsModelList {
  models: ModelListItem[]
  total: number
}

export function WorkspaceDashboardPage() {
  const navigate = useNavigate()
  const { orgSlug, wsSlug, org, workspace, loading: wsLoading, hasNoOrgs } = useWorkspace()

  const { data, isLoading: modelsLoading, error } = useQuery({
    queryKey: ['ws-models', orgSlug, wsSlug],
    queryFn: () =>
      workspacedFetch<WsModelList>(orgSlug!, wsSlug!, '/models'),
    enabled: !!orgSlug && !!wsSlug,
    staleTime: 60 * 1000,
  })

  const models = data?.models ?? []

  if (wsLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-6 h-6 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (hasNoOrgs) {
    return (
      <div className="max-w-lg mx-auto mt-24 text-center">
        <Box className="w-12 h-12 mx-auto text-muted-foreground mb-4" />
        <h2 className="text-xl font-semibold text-foreground mb-2">No organisation yet</h2>
        <p className="text-sm text-muted-foreground">
          Create an organisation to start mapping your architecture.
        </p>
      </div>
    )
  }

  if (!workspace || !org) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-6 h-6 animate-spin text-muted-foreground" />
      </div>
    )
  }

  return (
    <div className="max-w-4xl mx-auto">
      {/* Header */}
      <div className="mb-8">
        <p className="text-xs text-muted-foreground uppercase tracking-wider font-medium mb-1">
          {org.name}
        </p>
        <h1 className="text-2xl font-bold text-foreground">{workspace.name}</h1>
        <p className="text-sm text-muted-foreground mt-1">
          {workspace.visibility === 'org-visible' ? 'Visible to organisation' : 'Private workspace'}
        </p>
      </div>

      {/* Quick actions */}
      <div className="flex gap-3 mb-8">
        <button
          type="button"
          onClick={() => navigate('/')}
          className="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium bg-foreground text-background hover:bg-foreground/90 transition-colors"
        >
          <Upload className="w-4 h-4" />
          Upload / Import Model
        </button>
      </div>

      {/* Models list */}
      <div>
        <h2 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider mb-3">
          Models
        </h2>

        {modelsLoading && (
          <div className="flex items-center gap-2 py-8 text-muted-foreground">
            <Loader2 className="w-4 h-4 animate-spin" />
            <span className="text-sm">Loading models…</span>
          </div>
        )}

        {error && (
          <div className="flex items-center gap-2 py-4 text-red-600 text-sm">
            <AlertCircle className="w-4 h-4" />
            Failed to load models: {(error as Error).message}
          </div>
        )}

        {!modelsLoading && !error && models.length === 0 && (
          <div className="border border-dashed border-border rounded-xl p-12 text-center">
            <Box className="w-10 h-10 mx-auto text-muted-foreground mb-3" />
            <p className="text-sm font-medium text-foreground mb-1">No models yet</p>
            <p className="text-xs text-muted-foreground">
              Upload a <code className="font-mono">.unm.yaml</code> file to get started.
            </p>
          </div>
        )}

        {!modelsLoading && models.length > 0 && (
          <div className="divide-y divide-border border border-border rounded-xl overflow-hidden">
            {models.map((model) => (
              <div
                key={model.id}
                className={cn(
                  'flex items-center justify-between px-4 py-3 bg-background',
                  'hover:bg-muted/40 cursor-pointer transition-colors',
                )}
                onClick={() => navigate(`/dashboard`)}
                role="button"
                tabIndex={0}
                onKeyDown={(e) => e.key === 'Enter' && navigate('/dashboard')}
              >
                <div>
                  <p className="text-sm font-medium text-foreground">{model.name ?? model.id}</p>
                  <p className="text-xs text-muted-foreground font-mono">{model.id}</p>
                </div>
                <span className="text-xs text-muted-foreground">
                  {model.version_count} version{model.version_count !== 1 ? 's' : ''}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
