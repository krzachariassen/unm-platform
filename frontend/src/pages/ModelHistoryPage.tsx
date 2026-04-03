import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { GitBranch, Clock, MessageSquare, Loader2, ChevronDown, ChevronUp, History } from 'lucide-react'
import { ModelRequired } from '@/components/ui/ModelRequired'
import { useModel } from '@/lib/model-context'
import { modelsApi } from '@/services/api'
import { PageHeader } from '@/components/ui/page-header'
import { Button } from '@/components/ui/button'
import { DiffViewer } from '@/components/model/DiffViewer'
import { cn } from '@/lib/utils'
import type { VersionMeta } from '@/types/model'

function formatDate(iso: string): string {
  return new Date(iso).toLocaleString(undefined, {
    year: 'numeric', month: 'short', day: 'numeric',
    hour: '2-digit', minute: '2-digit',
  })
}

function VersionRow({
  version,
  modelId,
  isLatest,
  isSelected,
  compareFrom,
  onSelect,
  onCompareFrom,
}: {
  version: VersionMeta
  modelId: string
  isLatest: boolean
  isSelected: boolean
  compareFrom: number | null
  onSelect: () => void
  onCompareFrom: () => void
}) {
  return (
    <div className={cn(
      'rounded-lg border p-3 transition-colors',
      isSelected ? 'border-foreground/30 bg-muted/50' : 'border-border bg-card hover:bg-muted/30'
    )}>
      <div className="flex items-start justify-between gap-3">
        <div className="flex items-start gap-2.5 min-w-0">
          <div className="mt-0.5 flex items-center gap-1.5 shrink-0">
            <GitBranch className="w-3.5 h-3.5 text-muted-foreground" />
            <span className="text-xs font-mono font-semibold text-foreground">v{version.version}</span>
            {isLatest && (
              <span className="px-1.5 py-0.5 rounded-full bg-green-100 text-green-700 text-[10px] font-medium">
                latest
              </span>
            )}
          </div>
          <div className="min-w-0 space-y-0.5">
            {version.commit_message ? (
              <div className="flex items-center gap-1.5">
                <MessageSquare className="w-3 h-3 text-muted-foreground shrink-0" />
                <span className="text-xs text-foreground truncate">{version.commit_message}</span>
              </div>
            ) : (
              <span className="text-xs text-muted-foreground italic">No commit message</span>
            )}
            <div className="flex items-center gap-1 text-[11px] text-muted-foreground">
              <Clock className="w-3 h-3" />
              {formatDate(version.committed_at)}
            </div>
          </div>
        </div>
        <div className="flex items-center gap-1.5 shrink-0">
          <Button
            size="sm"
            variant="ghost"
            className="h-6 px-2 text-[11px]"
            onClick={onCompareFrom}
            title="Set as diff base version"
          >
            {compareFrom === version.version ? 'Base ✓' : 'Set base'}
          </Button>
          <Button
            size="sm"
            variant="ghost"
            className="h-6 px-2 text-[11px]"
            onClick={onSelect}
          >
            {isSelected ? <ChevronUp className="w-3 h-3" /> : <ChevronDown className="w-3 h-3" />}
            Compare
          </Button>
        </div>
      </div>

      {isSelected && compareFrom !== null && compareFrom !== version.version && (
        <div className="mt-3 pt-3 border-t border-border">
          <DiffViewer
            modelId={modelId}
            fromVersion={Math.min(compareFrom, version.version)}
            toVersion={Math.max(compareFrom, version.version)}
          />
        </div>
      )}
    </div>
  )
}

export function ModelHistoryPage() {
  const { modelId } = useModel()
  const [selectedVersion, setSelectedVersion] = useState<number | null>(null)
  const [compareFrom, setCompareFrom] = useState<number | null>(null)

  const { data, isLoading, error } = useQuery({
    queryKey: ['modelHistory', modelId],
    queryFn: () => modelsApi.getHistory(modelId!),
    enabled: !!modelId,
  })

  return (
    <ModelRequired>
      <div className="max-w-screen-lg mx-auto space-y-4">
        <PageHeader
          title="Version History"
          description="Browse and compare model versions committed via changesets"
        />

        {isLoading && (
          <div className="flex items-center gap-3 py-8 justify-center text-muted-foreground">
            <Loader2 className="w-4 h-4 animate-spin" />
            <span className="text-sm">Loading history…</span>
          </div>
        )}

        {error && (
          <div className="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-700">
            Failed to load history: {(error as Error).message}
          </div>
        )}

        {data && data.versions.length === 0 && (
          <div className="flex flex-col items-center justify-center py-16 gap-3 rounded-lg border border-dashed border-border text-center">
            <History className="w-8 h-8 text-muted-foreground/40" />
            <p className="text-sm font-medium text-muted-foreground">No version history yet</p>
            <p className="text-xs text-muted-foreground/70">
              Commit a changeset in What-If Explorer to create a version
            </p>
          </div>
        )}

        {data && data.versions.length > 0 && (
          <div className="space-y-2">
            {compareFrom !== null && (
              <p className="text-xs text-muted-foreground bg-muted/50 rounded px-3 py-2 border border-border">
                Base set to <span className="font-semibold text-foreground">v{compareFrom}</span>.
                Click "Compare" on another version to see the diff.{' '}
                <button
                  className="underline hover:no-underline"
                  onClick={() => { setCompareFrom(null); setSelectedVersion(null) }}
                >
                  Clear
                </button>
              </p>
            )}
            {[...data.versions].reverse().map((v, i) => (
              <VersionRow
                key={v.id}
                version={v}
                modelId={modelId!}
                isLatest={i === 0}
                isSelected={selectedVersion === v.version}
                compareFrom={compareFrom}
                onSelect={() => setSelectedVersion(prev => prev === v.version ? null : v.version)}
                onCompareFrom={() => setCompareFrom(prev => prev === v.version ? null : v.version)}
              />
            ))}
          </div>
        )}
      </div>
    </ModelRequired>
  )
}
