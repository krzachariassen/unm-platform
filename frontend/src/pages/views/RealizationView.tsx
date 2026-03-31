import { useMemo, useState } from 'react'
import { useQueries } from '@tanstack/react-query'
import { GitBranch, Layers, Server, AlertTriangle } from 'lucide-react'
import { ModelRequired } from '@/components/ui/ModelRequired'
import { PageHeader } from '@/components/ui/page-header'
import { StatCard } from '@/components/ui/stat-card'
import { LoadingState, ErrorState } from '@/components/ViewState'
import { useModel } from '@/lib/model-context'
import { useSearch } from '@/lib/search-context'
import { viewsApi } from '@/services/api'
import { buildCapToSvcTeam, buildCapVisibility, buildGroupedNeeds } from '@/features/realization/utils'
import { ValueChainView, ServiceTableView } from '@/features/realization/RealizationTabs'
import { cn } from '@/lib/utils'

type Tab = 'chain' | 'service'

export function RealizationView() {
  const { modelId } = useModel()
  const { query } = useSearch()
  const [tab, setTab] = useState<Tab>('chain')

  const [realizationQ, needQ] = useQueries({
    queries: [
      { queryKey: ['realizationView', modelId], queryFn: () => viewsApi.getRealizationView(modelId!), enabled: !!modelId },
      { queryKey: ['needView', modelId],        queryFn: () => viewsApi.getNeedView(modelId!),        enabled: !!modelId },
    ],
  })

  const capToSvcTeam = useMemo(() => realizationQ.data ? buildCapToSvcTeam(realizationQ.data) : new Map(), [realizationQ.data])
  const capVisibility  = useMemo(() => realizationQ.data ? buildCapVisibility(realizationQ.data)  : new Map(), [realizationQ.data])
  const groupedNeeds  = useMemo(() => needQ.data ? buildGroupedNeeds(needQ.data, capToSvcTeam, capVisibility) : [], [needQ.data, capToSvcTeam, capVisibility])

  if (realizationQ.isLoading || needQ.isLoading) return <LoadingState message="Loading traceability…" />
  if (realizationQ.error) return <ErrorState message={(realizationQ.error as Error).message} />
  if (!realizationQ.data) return null

  const viewData = realizationQ.data
  const crossTeamCount = groupedNeeds.filter(n => n.isCrossTeam).length
  const unbackedCount  = groupedNeeds.filter(n => n.isUnbacked).length
  const mappedCount    = groupedNeeds.filter(n => !n.isUnbacked).length

  return (
    <ModelRequired>
      <div className="space-y-5">
        <PageHeader
          title="Realization View"
          description="End-to-end value chain traceability — Need → Capability → Service → Team"
          actions={
            <div className="flex rounded-lg overflow-hidden border border-border">
              <button onClick={() => setTab('chain')} className={cn('px-4 py-2 text-sm font-medium transition-colors', tab === 'chain' ? 'bg-slate-900 text-white' : 'bg-white text-slate-600 hover:bg-slate-50')}>Value Chain</button>
              <button onClick={() => setTab('service')} className={cn('px-4 py-2 text-sm font-medium border-l border-border transition-colors', tab === 'service' ? 'bg-slate-900 text-white' : 'bg-white text-slate-600 hover:bg-slate-50')}>By Service</button>
            </div>
          }
        />

        <div className="grid grid-cols-2 sm:grid-cols-5 gap-3">
          <StatCard label="Needs"       value={groupedNeeds.length} icon={<Layers className="w-4 h-4" />} />
          <StatCard label="Cross-Team"  value={crossTeamCount}      icon={<GitBranch className="w-4 h-4 text-amber-500" />} />
          <StatCard label="Unbacked"    value={unbackedCount}       icon={<AlertTriangle className="w-4 h-4 text-red-500" />} />
          <StatCard label="Services"    value={viewData.service_rows.length} icon={<Server className="w-4 h-4" />} />
          <StatCard label="Mapped"      value={`${mappedCount} (${groupedNeeds.length > 0 ? Math.round(mappedCount / groupedNeeds.length * 100) : 0}%)`} icon={<Layers className="w-4 h-4 text-green-500" />} />
        </div>

        {tab === 'chain' && (
          <ValueChainView groupedNeeds={groupedNeeds} capToSvcTeam={capToSvcTeam} query={query} />
        )}
        {tab === 'service' && (
          <ServiceTableView viewData={viewData} query={query} />
        )}
      </div>
    </ModelRequired>
  )
}
