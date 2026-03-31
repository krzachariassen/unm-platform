import { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Users, AlertTriangle, Gauge, Shield } from 'lucide-react'
import { ModelRequired } from '@/components/ui/ModelRequired'
import { useModel } from '@/lib/model-context'
import { usePageInsights } from '@/hooks/usePageInsights'
import { viewsApi } from '@/services/api'
import { LoadingState, ErrorState } from '@/components/ViewState'
import { PageHeader } from '@/components/ui/page-header'
import { StatCard } from '@/components/ui/stat-card'
import { slug } from '@/lib/slug'
import { TeamCard } from '@/features/cognitive-load/TeamCard'
import { cn } from '@/lib/utils'

type SortKey = 'level' | 'name' | 'type'
const TEAM_TYPES = ['stream-aligned', 'platform', 'enabling', 'complicated-subsystem']
const levelRank = (l: string) => l === 'high' ? 3 : l === 'medium' ? 2 : 1

const DIMENSIONS_INFO = [
  { key: 'domain_spread'    as const, label: 'Domain Spread',      thresholds: '1-3 low · 4-5 med · 6+ high' },
  { key: 'service_load'     as const, label: 'Service Load',       thresholds: '≤2 low · 2-3 med · >3 high' },
  { key: 'interaction_load' as const, label: 'Interaction Load',   thresholds: '≤3 low · 4-6 med · 7+ high' },
  { key: 'dependency_load'  as const, label: 'Dependency Fan-out', thresholds: '≤4 low · 5-8 med · 9+ high' },
]

export function CognitiveLoadView() {
  const { modelId } = useModel()
  const { data: viewData, isLoading, error } = useQuery({
    queryKey: ['cognitiveLoadView', modelId],
    queryFn: () => viewsApi.getCognitiveLoadView(modelId!),
    enabled: !!modelId,
  })
  const { insights } = usePageInsights('cognitive-load')
  const [sortKey, setSortKey] = useState<SortKey>('level')
  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('desc')
  const [filterType, setFilterType] = useState('')
  const [expandedTeam, setExpandedTeam] = useState<string | null>(null)
  const [showThresholds, setShowThresholds] = useState(false)

  const loads = viewData?.team_loads ?? []

  const counts = useMemo(() => ({
    total: loads.length,
    high: loads.filter(t => t.overall_level === 'high').length,
    medium: loads.filter(t => t.overall_level === 'medium').length,
    low: loads.filter(t => t.overall_level === 'low').length,
  }), [loads])

  const sorted = useMemo(() => {
    const filtered = filterType ? loads.filter(t => t.team.type === filterType) : loads
    const dir = sortDir === 'desc' ? 1 : -1
    return [...filtered].sort((a, b) => {
      if (sortKey === 'level') return dir * (levelRank(b.overall_level) - levelRank(a.overall_level))
      if (sortKey === 'name') return dir * a.team.name.localeCompare(b.team.name)
      return dir * a.team.type.localeCompare(b.team.type)
    })
  }, [loads, filterType, sortKey, sortDir])

  const handleSort = (key: SortKey) => {
    if (sortKey === key) setSortDir(p => p === 'desc' ? 'asc' : 'desc')
    else { setSortKey(key); setSortDir('desc') }
  }

  if (isLoading) return <LoadingState message="Analyzing structural cognitive load…" />
  if (error) return <ErrorState message={(error as Error).message} />

  return (
    <ModelRequired>
      <div className="space-y-6">
        <PageHeader
          title="Structural Cognitive Load"
          description="Team Topologies assessment across 4 structural dimensions — worst dimension sets overall load"
        />

        {/* Stats */}
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
          <StatCard label="Total Teams" value={counts.total} icon={<Users className="w-4 h-4" />} />
          <StatCard label="High Load"   value={counts.high}   icon={<AlertTriangle className="w-4 h-4" />} />
          <StatCard label="Medium Load" value={counts.medium} icon={<Gauge className="w-4 h-4" />} />
          <StatCard label="Low Load"    value={counts.low}    icon={<Shield className="w-4 h-4" />} />
        </div>

        {/* Distribution bar */}
        {counts.total > 0 && (
          <div className="rounded-xl p-4 bg-white border border-slate-200">
            <div className="text-[11px] font-semibold text-slate-400 uppercase tracking-wide mb-2.5">Load Distribution</div>
            <div className="flex h-3.5 rounded-full overflow-hidden bg-slate-100">
              {counts.high > 0   && <div className="bg-gradient-to-r from-red-400 to-red-500 transition-all"    style={{ width: `${(counts.high   / counts.total) * 100}%` }} title={`High: ${counts.high}`} />}
              {counts.medium > 0 && <div className="bg-gradient-to-r from-amber-400 to-amber-500 transition-all" style={{ width: `${(counts.medium / counts.total) * 100}%` }} title={`Medium: ${counts.medium}`} />}
              {counts.low > 0    && <div className="bg-gradient-to-r from-green-400 to-green-500 transition-all" style={{ width: `${(counts.low    / counts.total) * 100}%` }} title={`Low: ${counts.low}`} />}
            </div>
            <div className="flex gap-5 mt-2">
              {[{ label: 'High', color: 'bg-red-500', n: counts.high }, { label: 'Medium', color: 'bg-amber-500', n: counts.medium }, { label: 'Low', color: 'bg-green-500', n: counts.low }].map(it => (
                <div key={it.label} className="flex items-center gap-1.5 text-xs text-slate-500">
                  <span className={cn('w-2 h-2 rounded-full', it.color)} />
                  {it.label}: {it.n}
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Threshold info */}
        <div>
          <button onClick={() => setShowThresholds(p => !p)} className="text-xs text-primary font-medium hover:underline">
            ⓘ Dimension thresholds {showThresholds ? '▲' : '▼'}
          </button>
          {showThresholds && (
            <div className="mt-2 rounded-xl p-4 bg-white border border-slate-200 grid grid-cols-2 gap-3">
              {DIMENSIONS_INFO.map(d => (
                <div key={d.key}>
                  <div className="text-xs font-semibold text-slate-700">{d.label}</div>
                  <div className="text-[10px] text-slate-400 font-mono">{d.thresholds}</div>
                </div>
              ))}
              <div className="col-span-2 text-[10px] text-slate-400 pt-2 border-t border-slate-100">
                Interaction weights: collaboration = 3 · facilitating = 2 · x-as-a-service = 1 · Overall = worst dimension
              </div>
            </div>
          )}
        </div>

        {/* Controls */}
        <div className="flex flex-wrap items-center gap-3">
          <div className="flex rounded-lg border border-border overflow-hidden text-sm">
            {(['level', 'name', 'type'] as SortKey[]).map(k => (
              <button key={k} onClick={() => handleSort(k)}
                className={cn('px-3 py-1.5 font-medium capitalize transition-colors border-r last:border-r-0 border-border',
                  sortKey === k ? 'bg-blue-600 text-white' : 'bg-white text-slate-600 hover:bg-slate-50')}>
                {k} {sortKey === k && (sortDir === 'desc' ? '↓' : '↑')}
              </button>
            ))}
          </div>
          <div className="flex gap-1.5 flex-wrap">
            <button onClick={() => setFilterType('')} className={cn('px-3 py-1 rounded-full text-xs font-semibold border transition-colors', !filterType ? 'bg-gradient-to-r from-indigo-500 to-purple-500 text-white border-transparent' : 'bg-white text-slate-500 border-slate-200 hover:bg-slate-50')}>All</button>
            {TEAM_TYPES.map(t => (
              <button key={t} onClick={() => setFilterType(p => p === t ? '' : t)} className={cn('px-3 py-1 rounded-full text-xs font-semibold border transition-colors', filterType === t ? 'bg-slate-800 text-white border-transparent' : 'bg-white text-slate-500 border-slate-200 hover:bg-slate-50')}>
                {t}
              </button>
            ))}
          </div>
        </div>

        {/* Team cards grid */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          {sorted.map(tl => (
            <TeamCard
              key={tl.team.name}
              tl={tl}
              insight={insights[`team:${slug(tl.team.name)}`]}
              isExpanded={expandedTeam === tl.team.name}
              onToggle={() => setExpandedTeam(p => p === tl.team.name ? null : tl.team.name)}
            />
          ))}
        </div>
      </div>
    </ModelRequired>
  )
}
