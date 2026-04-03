import { useMemo, useRef, useState } from 'react'
import { useQueries } from '@tanstack/react-query'
import { ModelRequired } from '@/components/ui/ModelRequired'
import { ContentContainer } from '@/components/ui/content-container'
import { PageHeader } from '@/components/ui/page-header'
import { LoadingState, ErrorState } from '@/components/ViewState'
import { AntiPatternPanel } from '@/components/AntiPatternPanel'
import { SlidePanel, PanelSection, PanelField } from '@/components/ui/slide-panel'
import { useModel } from '@/lib/model-context'
import { useSearch, matchesQuery } from '@/lib/search-context'
import { usePageInsights } from '@/hooks/usePageInsights'
import { VIS_BADGE } from '@/lib/visibility-styles'
import { TEAM_TYPE_BADGE } from '@/lib/team-type-styles'
import { viewsApi } from '@/services/api'
import { TeamLane } from '@/features/ownership/TeamLane'
import { DomainView } from '@/features/ownership/DomainView'
import type { NodeDetails, SvcPanelData } from '@/features/ownership/TeamLane'
import { cn } from '@/lib/utils'

const ALL_TEAM_TYPES = ['stream-aligned', 'platform', 'enabling', 'complicated-subsystem'] as const

export function OwnershipView() {
  const { modelId } = useModel()
  const { query, teamFilter } = useSearch()
  const [selectedNode, setSelectedNode] = useState<NodeDetails | null>(null)
  const [tab, setTab] = useState<'team' | 'domain'>('team')
  const [selectedSvc, setSelectedSvc] = useState<SvcPanelData | null>(null)
  const [filterCrossTeam, setFilterCrossTeam] = useState(false)
  const [filterOverloaded, setFilterOverloaded] = useState(false)
  const [filterUnowned, setFilterUnowned] = useState(false)
  const [showProblemsOnly, setShowProblemsOnly] = useState(false)
  const [activeTypeFilters, setActiveTypeFilters] = useState<Set<string>>(new Set())
  const [localSearch, setLocalSearch] = useState('')
  const unownedRef = useRef<HTMLDivElement>(null)
  const { insights } = usePageInsights('ownership')

  const [ownershipQ, capabilityQ] = useQueries({
    queries: [
      { queryKey: ['ownershipView', modelId], queryFn: () => viewsApi.getOwnershipView(modelId!), enabled: !!modelId },
      { queryKey: ['capabilityView', modelId], queryFn: () => viewsApi.getCapabilityView(modelId!), enabled: !!modelId },
    ],
  })

  const svcCapMap = useMemo(() => {
    const map = new Map<string, Array<{ label: string; visibility: string }>>()
    ownershipQ.data?.lanes.forEach(lane => {
      lane.caps.forEach(cg => {
        cg.services.forEach(svc => {
          const existing = map.get(svc.id) ?? []
          if (!existing.some(e => e.label === cg.cap.label)) existing.push({ label: cg.cap.label, visibility: cg.cap.data.visibility ?? '' })
          map.set(svc.id, existing)
        })
      })
    })
    return map
  }, [ownershipQ.data])

  const filteredLanes = useMemo(() => {
    if (!ownershipQ.data) return []
    return ownershipQ.data.lanes.filter(lane => {
      if (teamFilter && !matchesQuery(lane.team.label, teamFilter)) return false
      if (query && !(matchesQuery(lane.team.label, query) || lane.caps.some(cg =>
        matchesQuery(cg.cap.label, query) || cg.services.some(s => matchesQuery(s.label, query))
      ))) return false
      if (localSearch) {
        const q = localSearch.toLowerCase()
        if (!(lane.team.label.toLowerCase().includes(q) ||
          lane.caps.some(cg => cg.cap.label.toLowerCase().includes(q) || cg.services.some(s => s.label.toLowerCase().includes(q))))) return false
      }
      const isOverloaded = lane.team.data.is_overloaded
      const hasCrossTeam = lane.caps.some(cg => cg.cross_team)
      if (filterCrossTeam && !hasCrossTeam) return false
      if (filterOverloaded && !isOverloaded) return false
      if (showProblemsOnly && !(isOverloaded || hasCrossTeam)) return false
      if (activeTypeFilters.size > 0 && !activeTypeFilters.has(lane.team.data.type ?? '')) return false
      return true
    })
  }, [ownershipQ.data, teamFilter, query, localSearch, filterCrossTeam, filterOverloaded, showProblemsOnly, activeTypeFilters])

  if (ownershipQ.isLoading) return <LoadingState />
  if (ownershipQ.error) return <ErrorState message={(ownershipQ.error as Error).message} />
  if (!ownershipQ.data) return null

  const viewData = ownershipQ.data
  const totalCaps = viewData.lanes.reduce((sum, l) => sum + l.caps.length, 0)
  const crossTeamCount = viewData.cross_team_capabilities.length
  const unownedCount = viewData.unowned_capabilities.length
  const overloadedCount = viewData.overloaded_teams.length

  function openSvcPanel(svc: { id: string; label: string; team_label: string; cap_count: number; team_id: string }, currentLaneTeamId: string) {
    const capList = svcCapMap.get(svc.id) ?? []
    const isHighSpan = svc.cap_count >= 3
    const isFromOtherTeam = Boolean(svc.team_id && svc.team_id !== currentLaneTeamId)
    const owningLane = viewData.lanes.find(l => l.team.label === svc.team_label)
    const teamType = owningLane?.team.data.type ?? ''
    setSelectedSvc({ label: svc.label, teamLabel: svc.team_label, teamType, capList, isHighSpan, isFromOtherTeam })
  }

  const filterBtnCls = (active: boolean, activeColor: string) =>
    `text-[11px] font-semibold rounded-full px-3 py-1 border cursor-pointer transition-colors ${active ? activeColor : 'bg-slate-100 text-slate-500 border-slate-200 hover:bg-slate-200'}`

  return (
    <ModelRequired>
      <ContentContainer className="space-y-4 relative">
        <PageHeader
          title="Ownership View"
          description="Teams, capabilities, and service ownership across the model"
          actions={
            <div className="flex rounded-md overflow-hidden border border-border">
              {(['team', 'domain'] as const).map(t => (
                <button key={t} onClick={() => setTab(t)}
                  className={cn('px-3 py-1.5 text-xs font-medium capitalize transition-colors border-l border-border first:border-l-0',
                    tab === t ? 'bg-foreground text-background' : 'bg-card text-muted-foreground hover:bg-muted')}>
                  {t === 'team' ? 'By Team' : 'By Domain Area'}
                </button>
              ))}
            </div>
          }
        />

        {/* Stats + filter bar */}
        <div className="rounded-lg p-3 space-y-2.5 border border-border bg-card">
          <div className="flex items-center gap-2 flex-wrap text-sm text-slate-500">
            <span>{viewData.lanes.length} teams</span>
            <span className="text-slate-300">·</span>
            <span>{totalCaps} caps</span>
            <span className="text-slate-300">·</span>
            <span>{viewData.service_rows.length} services</span>
            {crossTeamCount > 0 && <>
              <span className="text-slate-300">·</span>
              <button onClick={() => setFilterCrossTeam(v => !v)} className={filterBtnCls(filterCrossTeam, 'bg-orange-600 text-white border-orange-600')}>{crossTeamCount} cross-team</button>
            </>}
            {overloadedCount > 0 && <>
              <span className="text-slate-300">·</span>
              <button onClick={() => setFilterOverloaded(v => !v)} className={filterBtnCls(filterOverloaded, 'bg-red-600 text-white border-red-600')}>{overloadedCount} overloaded</button>
            </>}
            {unownedCount > 0 && <>
              <span className="text-slate-300">·</span>
              <button onClick={() => { setFilterUnowned(v => !v); setTimeout(() => unownedRef.current?.scrollIntoView({ behavior: 'smooth', block: 'start' }), 50) }}
                className={filterBtnCls(filterUnowned, 'bg-red-600 text-white border-red-600')}>{unownedCount} unowned</button>
            </>}
            {(viewData.external_dependency_count ?? 0) > 0 && <>
              <span className="text-slate-300">·</span>
              <span className="text-[11px] font-semibold rounded-full px-3 py-1 bg-slate-100 text-slate-500">{viewData.external_dependency_count} external deps</span>
            </>}
            <span className="flex-1 min-w-2" />
            <button onClick={() => setShowProblemsOnly(v => !v)} className={`text-xs font-medium underline-offset-2 hover:underline ${showProblemsOnly ? 'text-slate-900' : 'text-slate-400'}`}>
              {showProblemsOnly ? 'Show all' : 'Show problems only'}
            </button>
          </div>
          <div className="flex items-center gap-2 flex-wrap pt-2 border-t border-slate-200">
            <span className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider mr-1">Team type</span>
            {ALL_TEAM_TYPES.map(t => {
              const active = activeTypeFilters.has(t)
              return (
                <button key={t} type="button" onClick={() => setActiveTypeFilters(prev => { const n = new Set(prev); if (n.has(t)) n.delete(t); else n.add(t); return n })}
                  className="text-[11px] font-semibold rounded-full px-2.5 py-0.5 border cursor-pointer transition-colors"
                  style={{ background: active ? '#1d4ed8' : 'white', color: active ? 'white' : '#374151', borderColor: active ? '#1d4ed8' : '#d1d5db' }}>
                  {t.split('-').map(w => w.charAt(0).toUpperCase() + w.slice(1)).join(' ')}
                </button>
              )
            })}
            {activeTypeFilters.size > 0 && <button type="button" onClick={() => setActiveTypeFilters(new Set())} className="text-xs text-slate-400 underline ml-1">clear</button>}
          </div>
        </div>

        {tab === 'team' && (
          <>
            <input value={localSearch} onChange={e => setLocalSearch(e.target.value)}
              placeholder="Filter by team, service, or capability..."
              className="w-full max-w-sm px-3 py-1.5 border border-slate-300 rounded-lg text-sm" />
            {filteredLanes.length === 0 && (localSearch || activeTypeFilters.size > 0) && (
              <div className="text-center text-slate-500 py-12 text-sm">
                No teams match current filters.
                <button onClick={() => { setLocalSearch(''); setActiveTypeFilters(new Set()) }} className="ml-2 text-blue-500 underline">Clear</button>
              </div>
            )}
            <div className="grid gap-4" style={{ gridTemplateColumns: 'repeat(auto-fill, minmax(380px, 1fr))' }}>
              {filteredLanes.map(lane => (
                <TeamLane key={lane.team.id} lane={lane} query={query} insights={insights}
                  crossTeamCaps={viewData.cross_team_capabilities} onSelectNode={setSelectedNode} onOpenSvcPanel={openSvcPanel} />
              ))}
            </div>
            {viewData.unowned_capabilities.length > 0 && (
              <div ref={unownedRef} className="rounded-lg overflow-hidden mt-4 border border-red-200 bg-card transition-shadow"
                style={filterUnowned ? { boxShadow: '0 0 0 2px #ef4444' } : undefined}>
                <div className="flex items-center gap-2 px-3 py-2.5 border-b border-red-200">
                  <span className="text-sm font-semibold text-red-800">Unowned Capabilities</span>
                  <span className="text-xs text-red-600">no team assigned</span>
                </div>
                <div className="px-3 py-3 flex flex-wrap gap-1.5">
                  {viewData.unowned_capabilities.filter(c => !query || matchesQuery(c.label, query)).map(c => (
                    <button key={c.id} type="button" onClick={() => setSelectedNode({ id: c.id, label: c.label, nodeType: 'capability', data: { ...c.data, nodeType: 'capability' } })}
                      className="text-[11px] font-semibold rounded-full px-3.5 py-2 bg-white border border-dashed border-red-300 text-rose-700 hover:-translate-y-px transition-transform">
                      {c.label}
                    </button>
                  ))}
                </div>
              </div>
            )}
          </>
        )}

        {tab === 'domain' && (
          <DomainView capViewData={capabilityQ.data ?? null} viewData={viewData} query={query}
            onSelectNode={setSelectedNode} onTabSwitch={setTab} onSetSearch={setLocalSearch} />
        )}

        {/* Service detail panel (replaces popover) */}
        <SlidePanel
          open={!!selectedSvc}
          onClose={() => setSelectedSvc(null)}
          title={selectedSvc?.label ?? ''}
          subtitle={selectedSvc ? `Owned by ${selectedSvc.teamLabel}` : undefined}
          badge={selectedSvc?.teamType ? (() => {
            const b = TEAM_TYPE_BADGE[selectedSvc.teamType] ?? { bg: '#f3f4f6', text: '#374151' }
            return <span className="text-[11px] font-semibold rounded-full px-2 py-0.5" style={{ background: b.bg, color: b.text }}>{selectedSvc.teamType}</span>
          })() : undefined}
        >
          {selectedSvc && (
            <>
              {selectedSvc.capList.length > 0 && (
                <PanelSection label={`Capabilities (${selectedSvc.capList.length})`}>
                  <ul className="space-y-1.5">
                    {selectedSvc.capList.map((cap, i) => {
                      const b = VIS_BADGE[cap.visibility] ?? { bg: '#f3f4f6', text: '#374151' }
                      return (
                        <li key={i} className="flex items-center gap-2">
                          <span className="text-slate-300">•</span>
                          <span className="flex-1 min-w-0 text-xs font-medium text-slate-900">{cap.label}</span>
                          {cap.visibility && <span className="text-[10px] font-semibold rounded-full px-1.5 py-0.5 shrink-0" style={{ background: b.bg, color: b.text }}>{cap.visibility}</span>}
                        </li>
                      )
                    })}
                  </ul>
                </PanelSection>
              )}
              {(selectedSvc.isHighSpan || selectedSvc.isFromOtherTeam) && (
                <PanelSection label="Signals">
                  {selectedSvc.isHighSpan && (
                    <PanelField label="High-span service" value="This service realizes 3+ capabilities — consider splitting responsibilities." />
                  )}
                  {selectedSvc.isFromOtherTeam && (
                    <PanelField label="Cross-team dependency" value="This service is owned by a different team than the capability it realizes." />
                  )}
                </PanelSection>
              )}
            </>
          )}
        </SlidePanel>

        <AntiPatternPanel node={selectedNode} onClose={() => setSelectedNode(null)} />
      </ContentContainer>
    </ModelRequired>
  )
}
