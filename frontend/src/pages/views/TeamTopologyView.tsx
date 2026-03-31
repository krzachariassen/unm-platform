import { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Users, ArrowRight, Zap, Layers } from 'lucide-react'
import { ModelRequired } from '@/components/ui/ModelRequired'
import { ContentContainer } from '@/components/ui/content-container'
import { PageHeader } from '@/components/ui/page-header'
import { StatCard } from '@/components/ui/stat-card'
import { LoadingState, ErrorState } from '@/components/ViewState'
import { useModel } from '@/lib/model-context'
import { useSearch, matchesQuery } from '@/lib/search-context'
import { usePageInsights } from '@/hooks/usePageInsights'
import { viewsApi } from '@/services/api'
import { GraphView } from '@/features/team-topology/GraphView'
import { TableView } from '@/features/team-topology/TableView'
import { TEAM_TYPES, TEAM_TYPE_DESCRIPTIONS } from '@/features/team-topology/constants'
import { cn } from '@/lib/utils'

export function TeamTopologyView() {
  const { modelId } = useModel()
  const { query, teamTypeFilter } = useSearch()
  const [viewMode, setViewMode] = useState<'graph' | 'table'>('graph')
  const [filterOverloaded, setFilterOverloaded] = useState(false)
  const [localSearch, setLocalSearch] = useState('')
  const [localTypeFilter, setLocalTypeFilter] = useState<string | null>(null)
  const { insights } = usePageInsights('topology')

  const { data: viewData, isLoading, error } = useQuery({
    queryKey: ['teamTopologyView', modelId],
    queryFn: () => viewsApi.getTeamTopologyView(modelId!),
    enabled: !!modelId,
  })

  const filteredTeams = useMemo(() => {
    if (!viewData) return []
    return viewData.teams.filter(t => {
      if (teamTypeFilter && t.type !== teamTypeFilter) return false
      if (localTypeFilter && t.type !== localTypeFilter) return false
      if (query && !matchesQuery(t.label, query)) return false
      if (localSearch && !matchesQuery(t.label, localSearch)) return false
      if (filterOverloaded && !t.is_overloaded) return false
      return true
    })
  }, [viewData, teamTypeFilter, localTypeFilter, query, localSearch, filterOverloaded])

  const filteredIds = useMemo(() => new Set(filteredTeams.map(t => t.id)), [filteredTeams])
  const filteredInteractions = useMemo(() => {
    if (!viewData) return []
    return viewData.interactions.filter(ix => filteredIds.has(ix.source_id) && filteredIds.has(ix.target_id))
  }, [viewData, filteredIds])

  if (isLoading) return <LoadingState />
  if (error) return <ErrorState message={(error as Error).message} />
  if (!viewData) return null

  const overloadedCount = viewData.teams.filter(t => t.is_overloaded).length
  const typeCounts = Object.fromEntries(
    Object.keys(TEAM_TYPES).map(k => [k, viewData.teams.filter(t => t.type === k).length])
  )

  return (
    <ModelRequired>
      <ContentContainer className="space-y-4">
        <PageHeader
          title="Team Topology"
          description={`${viewData.teams.length} teams · ${viewData.interactions.length} interactions${overloadedCount > 0 ? ` · ${overloadedCount} overloaded` : ''}`}
          actions={
            <div className="flex gap-2 items-center">
              {overloadedCount > 0 && (
                <button onClick={() => setFilterOverloaded(f => !f)}
                  className={cn('px-2.5 py-1 rounded text-xs font-medium border transition-colors',
                    filterOverloaded ? 'bg-amber-500 text-white border-amber-500' : 'bg-amber-50 text-amber-800 border-amber-200 hover:bg-amber-100'
                  )}>
                  {overloadedCount} overloaded
                </button>
              )}
              <div className="flex rounded-md overflow-hidden border border-border">
                {(['graph', 'table'] as const).map(m => (
                  <button key={m} onClick={() => setViewMode(m)}
                    className={cn('px-3 py-1.5 text-xs font-medium capitalize transition-colors border-l border-border first:border-l-0',
                      viewMode === m ? 'bg-foreground text-background' : 'bg-card text-muted-foreground hover:bg-muted'
                    )}>
                    {m}
                  </button>
                ))}
              </div>
            </div>
          }
        />

        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
          <StatCard label="Total Teams" value={viewData.teams.length} icon={<Users className="w-4 h-4" />} />
          <StatCard label="Platform" value={typeCounts['platform'] ?? 0} description="shared capability providers" icon={<Layers className="w-4 h-4 text-purple-500" />} />
          <StatCard label="Stream-aligned" value={typeCounts['stream-aligned'] ?? 0} description="aligned to value streams" icon={<ArrowRight className="w-4 h-4 text-blue-500" />} />
          <StatCard label="Interactions" value={viewData.interactions.length} description={overloadedCount > 0 ? `${overloadedCount} overloaded` : 'all healthy'} icon={<Zap className="w-4 h-4 text-green-500" />} />
        </div>

        {/* Team type legend */}
        <div className="flex gap-2 flex-wrap">
          {Object.entries(TEAM_TYPES).map(([key, cfg]) => {
            const count = typeCounts[key] ?? 0
            if (count === 0) return null
            return (
              <span key={key} className="inline-flex items-center rounded-lg px-2.5 py-1 text-[11px] font-semibold"
                style={{ background: cfg.bg, color: cfg.accent, border: `1px solid ${cfg.border}` }}
                title={TEAM_TYPE_DESCRIPTIONS[key]}>
                <span style={{ width: 6, height: 6, borderRadius: '50%', background: cfg.accent, marginRight: 5, display: 'inline-block' }} />
                {cfg.label}
                <span className="ml-1 font-extrabold">{count}</span>
              </span>
            )
          })}
        </div>

        {/* Filters */}
        <div className="flex items-center gap-2 flex-wrap">
          <input value={localSearch} onChange={e => setLocalSearch(e.target.value)}
            placeholder="Search teams..."
            className="px-3 py-1.5 border border-slate-300 rounded-lg text-sm" />
          {Object.entries(TEAM_TYPES).map(([key, cfg]) => {
            const active = localTypeFilter === key
            return (
              <button key={key} type="button"
                onClick={() => setLocalTypeFilter(active ? null : key)}
                className="px-3 py-1 rounded-full text-xs font-semibold border cursor-pointer transition-colors"
                style={{
                  background: active ? cfg.accent : 'white',
                  color: active ? 'white' : cfg.accent,
                  borderColor: active ? cfg.accent : '#d1d5db',
                }}
                title={TEAM_TYPE_DESCRIPTIONS[key]}>
                {cfg.label}
              </button>
            )
          })}
        </div>

        {viewMode === 'graph' ? (
          <GraphView teams={filteredTeams} interactions={filteredInteractions} insights={insights} filterOverloaded={filterOverloaded} />
        ) : (
          <TableView teams={filteredTeams} interactions={filteredInteractions} insights={insights} />
        )}
      </ContentContainer>
    </ModelRequired>
  )
}
