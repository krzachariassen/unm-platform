import { useQuery } from '@tanstack/react-query'
import { useModel } from '@/lib/model-context'
import { viewsApi } from '@/services/api'
import { LoadingState, ErrorState } from '@/components/ViewState'
import { EmptyState } from '@/components/ui/empty-state'
import { ContentContainer } from '@/components/ui/content-container'
import { AlertTriangle, Users } from 'lucide-react'

function formatMode(mode: string): string {
  return mode.split('-').map(w => w.charAt(0).toUpperCase() + w.slice(1)).join(' ')
}

export function InteractionsTab() {
  const { modelId } = useModel()
  const { data, isLoading, error } = useQuery({
    queryKey: ['interactions', modelId],
    queryFn: () => viewsApi.getInteractions(modelId!),
    enabled: !!modelId,
  })

  if (isLoading) return <LoadingState />
  if (error) return <ErrorState message={(error as Error).message} />
  if (!data) return null

  const totalInteractions = Object.values(data.mode_distribution).reduce((s, v) => s + v, 0)

  if (totalInteractions === 0 && data.isolated_teams.length === 0) {
    return (
      <ContentContainer>
        <EmptyState
          title="No interactions defined"
          description="Add team interaction modes to your model to see diversity analysis here."
          icon={<Users className="w-12 h-12 text-muted-foreground" />}
        />
      </ContentContainer>
    )
  }

  return (
    <ContentContainer className="space-y-6">
      {/* all_modes_same warning */}
      {data.all_modes_same && (
        <div className="flex items-start gap-3 rounded-lg border border-amber-200 bg-amber-50 px-4 py-3">
          <AlertTriangle className="w-4 h-4 text-amber-600 shrink-0 mt-0.5" />
          <p className="text-sm text-amber-800">
            All teams use the same interaction mode — consider diversifying to improve collaboration patterns.
          </p>
        </div>
      )}

      {/* Mode distribution */}
      {Object.keys(data.mode_distribution).length > 0 && (
        <div>
          <h3 className="text-sm font-semibold text-foreground mb-3">Mode Distribution</h3>
          <div className="grid gap-2">
            {Object.entries(data.mode_distribution)
              .sort(([, a], [, b]) => b - a)
              .map(([mode, count]) => (
                <div key={mode} className="flex items-center gap-3">
                  <span className="text-xs font-medium text-muted-foreground w-36 shrink-0">
                    {formatMode(mode)}
                  </span>
                  <div className="flex-1 h-5 bg-muted rounded-full overflow-hidden">
                    <div
                      className="h-full bg-primary/70 rounded-full transition-all"
                      style={{ width: totalInteractions > 0 ? `${(count / totalInteractions) * 100}%` : '0%' }}
                    />
                  </div>
                  <span className="text-xs font-bold text-foreground w-6 text-right shrink-0">{count}</span>
                </div>
              ))}
          </div>
        </div>
      )}

      {/* Isolated teams */}
      {data.isolated_teams.length > 0 && (
        <div>
          <div className="flex items-center gap-2 mb-3">
            <AlertTriangle className="w-4 h-4 text-amber-500 shrink-0" />
            <h3 className="text-sm font-semibold text-foreground">
              Isolated Teams ({data.isolated_teams.length})
            </h3>
          </div>
          <p className="text-xs text-muted-foreground mb-2">
            These teams have zero declared interactions — potential silos.
          </p>
          <div className="flex flex-wrap gap-1.5">
            {data.isolated_teams.map(team => (
              <span key={team} className="text-xs font-medium rounded-full px-3 py-1 bg-amber-100 text-amber-800 border border-amber-200">
                {team}
              </span>
            ))}
          </div>
        </div>
      )}

      {/* Over-reliant teams */}
      {data.over_reliant_teams.length > 0 && (
        <div>
          <h3 className="text-sm font-semibold text-foreground mb-3">Over-Reliant Teams</h3>
          <p className="text-xs text-muted-foreground mb-2">
            Teams that rely heavily on a single interaction mode (4+ interactions of the same type).
          </p>
          <div className="space-y-2">
            {data.over_reliant_teams.map((entry, i) => (
              <div key={i} className="flex items-center gap-3 rounded-lg border border-border bg-card px-4 py-2.5">
                <span className="text-sm font-semibold text-foreground flex-1">{entry.team_name}</span>
                <span className="text-xs rounded-full px-2.5 py-0.5 bg-muted text-muted-foreground">
                  {formatMode(entry.mode)}
                </span>
                <span className="text-xs font-bold text-foreground">×{entry.count}</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </ContentContainer>
  )
}
