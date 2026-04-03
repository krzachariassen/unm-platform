import { useQuery } from '@tanstack/react-query'
import { Loader2, Plus, Minus, RefreshCw } from 'lucide-react'
import { modelsApi } from '@/services/api'
import type { DiffEntities } from '@/types/model'

export interface DiffViewerProps {
  modelId: string
  fromVersion: number
  toVersion: number
}

const ENTITY_LABELS: Record<keyof DiffEntities, string> = {
  actors: 'Actors',
  needs: 'Needs',
  capabilities: 'Capabilities',
  services: 'Services',
  teams: 'Teams',
}

function EntityGroup({ label, added, removed, changed }: {
  label: string
  added: string[]
  removed: string[]
  changed: string[]
}) {
  if (added.length === 0 && removed.length === 0 && changed.length === 0) return null
  return (
    <div className="space-y-1.5">
      <p className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">{label}</p>
      {added.map(name => (
        <div key={`a-${name}`} className="flex items-center gap-2 px-2 py-1 rounded bg-green-50 border border-green-200">
          <Plus className="w-3 h-3 text-green-600 shrink-0" />
          <span className="text-xs text-green-800">{name}</span>
        </div>
      ))}
      {removed.map(name => (
        <div key={`r-${name}`} className="flex items-center gap-2 px-2 py-1 rounded bg-red-50 border border-red-200">
          <Minus className="w-3 h-3 text-red-600 shrink-0" />
          <span className="text-xs text-red-800">{name}</span>
        </div>
      ))}
      {changed.map(name => (
        <div key={`c-${name}`} className="flex items-center gap-2 px-2 py-1 rounded bg-amber-50 border border-amber-200">
          <RefreshCw className="w-3 h-3 text-amber-600 shrink-0" />
          <span className="text-xs text-amber-800">{name}</span>
        </div>
      ))}
    </div>
  )
}

export function DiffViewer({ modelId, fromVersion, toVersion }: DiffViewerProps) {
  const { data, isLoading, error } = useQuery({
    queryKey: ['diff', modelId, fromVersion, toVersion],
    queryFn: () => modelsApi.getDiff(modelId, fromVersion, toVersion),
    enabled: !!modelId && fromVersion > 0 && toVersion > 0,
  })

  if (isLoading) {
    return (
      <div className="flex items-center gap-2 py-4 text-muted-foreground">
        <Loader2 className="w-3.5 h-3.5 animate-spin" />
        <span className="text-xs">Computing diff…</span>
      </div>
    )
  }

  if (error) {
    return (
      <div className="rounded-md bg-red-50 border border-red-200 p-3 text-xs text-red-700">
        Failed to load diff: {(error as Error).message}
      </div>
    )
  }

  if (!data) return null

  const hasChanges = (Object.keys(ENTITY_LABELS) as Array<keyof DiffEntities>).some(
    key => data.added[key].length > 0 || data.removed[key].length > 0 || data.changed[key].length > 0
  )

  return (
    <div className="space-y-3">
      <p className="text-xs text-muted-foreground font-medium">
        Comparing <span className="font-semibold text-foreground">v{fromVersion}</span>
        {' → '}
        <span className="font-semibold text-foreground">v{toVersion}</span>
      </p>

      {!hasChanges && (
        <div className="rounded-md border border-border bg-muted/40 p-4 text-center text-sm text-muted-foreground">
          No changes between these versions
        </div>
      )}

      {hasChanges && (
        <div className="space-y-3">
          {(Object.keys(ENTITY_LABELS) as Array<keyof DiffEntities>).map(key => (
            <EntityGroup
              key={key}
              label={ENTITY_LABELS[key]}
              added={data.added[key]}
              removed={data.removed[key]}
              changed={data.changed[key]}
            />
          ))}
        </div>
      )}
    </div>
  )
}
