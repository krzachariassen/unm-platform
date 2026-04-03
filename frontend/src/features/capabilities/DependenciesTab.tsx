import { useQuery } from '@tanstack/react-query'
import { useModel } from '@/lib/model-context'
import { viewsApi } from '@/services/api'
import { LoadingState, ErrorState } from '@/components/ViewState'
import { EmptyState } from '@/components/ui/empty-state'
import { StatCard } from '@/components/ui/stat-card'
import { ContentContainer } from '@/components/ui/content-container'
import { CheckCircle, AlertTriangle, ArrowRight } from 'lucide-react'

function CyclePath({ path }: { path: string[] }) {
  return (
    <div className="flex items-center flex-wrap gap-1 text-sm font-mono text-red-700">
      {path.map((node, i) => (
        <span key={i} className="flex items-center gap-1">
          <span className="rounded px-2 py-0.5 bg-red-50 border border-red-200">{node}</span>
          {i < path.length - 1 && <ArrowRight className="w-3 h-3 text-red-400 shrink-0" />}
        </span>
      ))}
    </div>
  )
}

export function DependenciesTab() {
  const { modelId } = useModel()
  const { data, isLoading, error } = useQuery({
    queryKey: ['dependencies', modelId],
    queryFn: () => viewsApi.getDependencies(modelId!),
    enabled: !!modelId,
  })

  if (isLoading) return <LoadingState />
  if (error) return <ErrorState message={(error as Error).message} />
  if (!data) return null

  const totalCycles = data.service_cycles.length + data.capability_cycles.length

  return (
    <ContentContainer className="space-y-6">
      {/* Stat cards */}
      <div className="grid grid-cols-3 gap-3">
        <StatCard label="Max Service Depth" value={data.max_service_depth} description="Longest service dependency chain" />
        <StatCard label="Max Capability Depth" value={data.max_capability_depth} description="Longest capability dependency chain" />
        <StatCard
          label="Dependency Cycles"
          value={totalCycles}
          description={totalCycles === 0 ? 'No cycles detected' : 'Cycles found — see below'}
          trend={totalCycles > 0 ? { direction: 'down', label: 'Action needed' } : { direction: 'up', label: 'Healthy' }}
        />
      </div>

      {/* No cycles empty state */}
      {totalCycles === 0 && (
        <EmptyState
          title="No dependency cycles detected"
          description="Good architecture hygiene — no circular dependencies between services or capabilities."
          icon={<CheckCircle className="w-12 h-12 text-green-500" />}
        />
      )}

      {/* Service cycles */}
      {data.service_cycles.length > 0 && (
        <div>
          <div className="flex items-center gap-2 mb-3">
            <AlertTriangle className="w-4 h-4 text-red-500 shrink-0" />
            <h3 className="text-sm font-semibold text-foreground">
              Service Cycles ({data.service_cycles.length})
            </h3>
          </div>
          <div className="space-y-2">
            {data.service_cycles.map((cycle, i) => (
              <div key={i} className="rounded-lg border border-red-200 bg-red-50 px-4 py-3">
                <CyclePath path={cycle.path} />
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Capability cycles */}
      {data.capability_cycles.length > 0 && (
        <div>
          <div className="flex items-center gap-2 mb-3">
            <AlertTriangle className="w-4 h-4 text-orange-500 shrink-0" />
            <h3 className="text-sm font-semibold text-foreground">
              Capability Cycles ({data.capability_cycles.length})
            </h3>
          </div>
          <div className="space-y-2">
            {data.capability_cycles.map((cycle, i) => (
              <div key={i} className="rounded-lg border border-orange-200 bg-orange-50 px-4 py-3">
                <CyclePath path={cycle.path} />
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Critical path */}
      {data.critical_service_path.length > 0 && (
        <div>
          <h3 className="text-sm font-semibold text-foreground mb-3">Critical Service Path</h3>
          <div className="rounded-lg border border-border bg-card px-4 py-3">
            <p className="text-xs text-muted-foreground mb-2">
              Longest dependency chain — services here are most impactful to system stability.
            </p>
            <div className="flex items-center flex-wrap gap-1">
              {data.critical_service_path.map((node, i) => (
                <span key={i} className="flex items-center gap-1 text-sm font-mono">
                  <span className="rounded px-2 py-0.5 bg-muted border border-border text-foreground">{node}</span>
                  {i < data.critical_service_path.length - 1 && (
                    <ArrowRight className="w-3 h-3 text-muted-foreground shrink-0" />
                  )}
                </span>
              ))}
            </div>
          </div>
        </div>
      )}
    </ContentContainer>
  )
}
