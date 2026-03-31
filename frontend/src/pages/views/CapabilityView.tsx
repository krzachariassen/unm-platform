import { useEffect, useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { ModelRequired } from '@/components/ui/ModelRequired'
import { ContentContainer } from '@/components/ui/content-container'
import { PageHeader } from '@/components/ui/page-header'
import { LoadingState, ErrorState } from '@/components/ViewState'
import { useModel } from '@/lib/model-context'
import { useSearch, matchesQuery } from '@/lib/search-context'
import { usePageInsights } from '@/hooks/usePageInsights'
import { viewsApi } from '@/services/api'
import { slug } from '@/lib/slug'
import { CapabilityCard } from '@/features/capability/CapabilityCard'
import { DetailPanel } from '@/features/capability/DetailPanel'
import { VIS_BANDS, TEAM_TYPE_CAP_BADGE } from '@/features/capability/constants'
import type { CapabilityType } from '@/features/capability/CapabilityCard'
import { cn } from '@/lib/utils'
import { ChevronDown, ChevronUp } from 'lucide-react'

type ViewMode = 'visibility' | 'domain' | 'team'

export function CapabilityView() {
  const { modelId } = useModel()
  const { query } = useSearch()
  const [viewMode, setViewMode] = useState<ViewMode>('visibility')
  const [expandedGroups, setExpandedGroups] = useState<Set<string>>(new Set())
  const [selectedCap, setSelectedCap] = useState<CapabilityType | null>(null)
  const { insights } = usePageInsights('capabilities')

  const { data: viewData, isLoading, error } = useQuery({
    queryKey: ['capabilityView', modelId],
    queryFn: () => viewsApi.getCapabilityView(modelId!),
    enabled: !!modelId,
  })

  useEffect(() => {
    if (!viewData) return
    setExpandedGroups(prev => {
      if (prev.size > 0) return prev
      return new Set(viewData.parent_groups.map(g => g.id))
    })
  }, [viewData])

  const capById = useMemo(() => new Map(viewData?.capabilities.map(c => [c.id, c]) ?? []), [viewData])
  const capToParent = useMemo(() => {
    const map = new Map<string, string>()
    viewData?.parent_groups.forEach(pg => pg.children.forEach(cid => map.set(cid, pg.id)))
    return map
  }, [viewData])

  if (isLoading) return <LoadingState />
  if (error) return <ErrorState message={(error as Error).message} />
  if (!viewData) return null

  const matchesCap = (cap: CapabilityType) =>
    !query || matchesQuery(cap.label, query) || cap.services.some(s => matchesQuery(s.label, query))

  const toggleGroup = (id: string) => setExpandedGroups(prev => {
    const next = new Set(prev); if (next.has(id)) next.delete(id); else next.add(id); return next
  })

  const fragmentedCount = viewData.fragmented_capabilities.length
  const highSpanCount = viewData.high_span_services.length
  const disconnectedCount = viewData.capabilities.filter(c => c.is_leaf && c.teams.length === 0).length
  const atRiskUserFacing = viewData.capabilities.filter(c => c.visibility === 'user-facing' && c.teams.length > 1)
  const dashInsight = insights['summary'] ?? insights['dashboard']

  const CapGrid = ({ caps }: { caps: CapabilityType[] }) => (
    <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-3 p-3">
      {caps.map(cap => <CapabilityCard key={cap.id} cap={cap} onClick={() => setSelectedCap(cap)} />)}
    </div>
  )

  const GroupToggle = ({ label, count, fragmented, id }: { label: string; count: number; fragmented?: number; id: string }) => {
    const isExpanded = expandedGroups.has(id)
    return (
      <button type="button" onClick={() => toggleGroup(id)} className="flex items-center gap-2 w-full text-left px-3 py-2.5 hover:bg-muted/50 transition-colors border-b border-border">
        <span className="text-sm font-semibold text-foreground">{label}</span>
        <span className="text-xs text-muted-foreground">{count}</span>
        {(fragmented ?? 0) > 0 && <span className="text-[10px] font-semibold rounded px-1.5 py-0.5 bg-red-100 text-red-700">{fragmented} fragmented</span>}
        <span className="ml-auto text-muted-foreground">{isExpanded ? <ChevronUp className="w-3.5 h-3.5" /> : <ChevronDown className="w-3.5 h-3.5" />}</span>
      </button>
    )
  }

  return (
    <ModelRequired>
      <ContentContainer className="space-y-4">
        <PageHeader
          title="Capability View"
          description={`${viewData.parent_groups.length} domain groups · ${viewData.leaf_capability_count} capabilities`}
          actions={
            <div className="flex rounded-md overflow-hidden border border-border">
              {(['visibility', 'domain', 'team'] as ViewMode[]).map(m => (
                <button key={m} onClick={() => setViewMode(m)}
                  className={cn('px-3 py-1.5 text-xs font-medium capitalize transition-colors border-l border-border first:border-l-0',
                    viewMode === m ? 'bg-foreground text-background' : 'bg-card text-muted-foreground hover:bg-muted')}>
                  {m === 'visibility' ? 'By Visibility' : m === 'domain' ? 'By Domain' : 'By Team'}
                </button>
              ))}
            </div>
          }
        />

        {/* Signals summary */}
        {fragmentedCount === 0 && disconnectedCount === 0 && highSpanCount === 0 && atRiskUserFacing.length === 0 ? (
          <div className="flex items-center gap-2 px-3 py-2 rounded-md bg-green-50 border border-green-200">
            <span className="text-green-600 text-sm">✓</span>
            <span className="text-xs text-green-700 font-medium">No architecture issues detected</span>
          </div>
        ) : (
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
            {[
              { value: fragmentedCount, label: 'Fragmented' },
              { value: disconnectedCount, label: 'Unowned' },
              { value: highSpanCount, label: 'High-span' },
              { value: atRiskUserFacing.length, label: 'User-facing at risk' },
            ].map(({ value, label }) => (
              <div key={label} className={cn('rounded-lg border p-3', value > 0 ? 'border-red-200 bg-red-50' : 'border-green-200 bg-green-50')}>
                <div className={cn('text-xl font-bold tabular-nums', value > 0 ? 'text-red-700' : 'text-green-700')}>{value}</div>
                <div className={cn('text-[10px] font-semibold uppercase tracking-wide mt-0.5', value > 0 ? 'text-red-600' : 'text-green-600')}>{label}</div>
              </div>
            ))}
          </div>
        )}

        {/* AI insight */}
        {dashInsight && (
          <div className="rounded-lg p-4 bg-sky-50 border border-sky-200">
            <p className="text-[10px] font-semibold text-sky-600 uppercase tracking-wide mb-1.5">AI Capability Analysis</p>
            <p className="text-xs leading-relaxed text-foreground">{dashInsight.explanation}</p>
            {dashInsight.suggestion && <p className="text-xs leading-relaxed text-sky-800 mt-1.5 font-medium">{dashInsight.suggestion}</p>}
          </div>
        )}

        {/* At-risk user-facing */}
        {atRiskUserFacing.length > 0 && (
          <div className="rounded-lg p-3 bg-red-50 border border-red-200">
            <h3 className="text-xs font-semibold text-red-800 mb-2">User-Facing Capabilities Served by Multiple Teams</h3>
            <div className="space-y-1.5">
              {atRiskUserFacing.map(cap => (
                <div key={cap.id} className="flex items-center gap-2 flex-wrap">
                  <span className="text-xs font-semibold text-foreground">{cap.label}</span>
                  {cap.teams.map(team => (
                    <span key={team.id} className="text-[10px] font-semibold rounded px-1.5 py-0.5 bg-foreground text-background">{team.label}</span>
                  ))}
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Fragmented/high-span pills */}
        {(fragmentedCount > 0 || highSpanCount > 0) && (
          <div className="flex flex-wrap gap-1.5">
            {viewData.fragmented_capabilities.map(fc => (
              <span key={fc.id} className="text-[10px] font-semibold rounded px-2 py-1 bg-red-100 text-red-700 border border-red-200">
                {fc.label} · {fc.team_count} teams
              </span>
            ))}
            {viewData.high_span_services.map(hs => (
              <span key={hs.name} className="text-[10px] font-semibold font-mono rounded px-2 py-1 bg-amber-50 text-amber-700 border border-amber-200">
                {hs.name} · {hs.capability_count} caps
              </span>
            ))}
          </div>
        )}

        {/* Visibility view */}
        {viewMode === 'visibility' && (
          <div className="space-y-6">
            {VIS_BANDS.map(band => {
              const bandCaps = viewData.capabilities.filter(c => c.visibility === band.key && c.is_leaf && matchesCap(c))
              if (bandCaps.length === 0) return null
              return (
                <div key={band.key}>
                  <div className="flex items-center gap-2 mb-3">
                    <span className="text-xs font-semibold uppercase tracking-wide px-2 py-0.5 rounded"
                      style={{ background: band.bg, color: band.accent, border: `1px solid ${band.border}` }}>
                      {band.label} · {bandCaps.length}
                    </span>
                    <div className="h-px flex-1 bg-border" />
                  </div>
                  <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-3">
                    {bandCaps.map(cap => <CapabilityCard key={cap.id} cap={cap} onClick={() => setSelectedCap(cap)} />)}
                  </div>
                </div>
              )
            })}
          </div>
        )}

        {/* Domain view */}
        {viewMode === 'domain' && (
          <div className="space-y-3">
            {viewData.parent_groups.map(pg => {
              const groupCaps = pg.children.map(id => capById.get(id)).filter((c): c is CapabilityType => c != null && c.is_leaf && matchesCap(c))
              if (groupCaps.length === 0) return null
              return (
                <div key={pg.id} className="rounded-lg border border-border bg-card overflow-hidden">
                  <GroupToggle label={pg.label} count={groupCaps.length} fragmented={groupCaps.filter(c => c.is_fragmented).length} id={pg.id} />
                  {expandedGroups.has(pg.id) && <CapGrid caps={groupCaps} />}
                </div>
              )
            })}
            {(() => {
              const uncategorized = viewData.capabilities.filter(c => c.is_leaf && !capToParent.has(c.id) && matchesCap(c))
              if (uncategorized.length === 0) return null
              return (
                <div className="rounded-lg border border-border bg-card overflow-hidden">
                  <GroupToggle label="Uncategorized" count={uncategorized.length} id="__uncategorized__" />
                  {expandedGroups.has('__uncategorized__') && <CapGrid caps={uncategorized} />}
                </div>
              )
            })()}
          </div>
        )}

        {/* Team view */}
        {viewMode === 'team' && (() => {
          const teamMap = new Map<string, { type: string; caps: CapabilityType[] }>()
          const unowned: CapabilityType[] = []
          for (const cap of viewData.capabilities) {
            if (!cap.is_leaf || !matchesCap(cap)) continue
            if (cap.teams.length === 0) unowned.push(cap)
            else for (const t of cap.teams) {
              const entry = teamMap.get(t.label) ?? { type: t.type, caps: [] }
              entry.caps.push(cap); teamMap.set(t.label, entry)
            }
          }
          const sortedTeams = Array.from(teamMap.entries()).sort((a, b) => b[1].caps.length - a[1].caps.length)
          return (
            <div className="space-y-3">
              {sortedTeams.map(([teamName, { type, caps }]) => {
                const badge = TEAM_TYPE_CAP_BADGE[type] ?? { bg: '#f1f5f9', text: '#475569', accent: '#6b7280' }
                return (
                  <div key={teamName} className="rounded-lg border border-border bg-card overflow-hidden">
                    <button type="button" onClick={() => toggleGroup(`team:${teamName}`)} className="flex items-center gap-2 w-full text-left px-3 py-2.5 hover:bg-muted/50 transition-colors border-b border-border">
                      <span className="text-sm font-semibold text-foreground">{teamName}</span>
                      <span className="text-[10px] font-semibold px-1.5 py-0.5 rounded" style={{ background: badge.bg, color: badge.text }}>{type}</span>
                      <span className="text-xs text-muted-foreground">{caps.length} cap{caps.length !== 1 ? 's' : ''}</span>
                      {caps.filter(c => c.is_fragmented).length > 0 && <span className="text-[10px] font-semibold rounded px-1.5 py-0.5 bg-red-100 text-red-700">{caps.filter(c => c.is_fragmented).length} fragmented</span>}
                      <span className="ml-auto text-muted-foreground">{expandedGroups.has(`team:${teamName}`) ? <ChevronUp className="w-3.5 h-3.5" /> : <ChevronDown className="w-3.5 h-3.5" />}</span>
                    </button>
                    {expandedGroups.has(`team:${teamName}`) && <CapGrid caps={caps} />}
                  </div>
                )
              })}
              {unowned.length > 0 && (
                <div className="rounded-lg border border-red-200 bg-card overflow-hidden">
                  <button type="button" onClick={() => toggleGroup('team:__unowned__')} className="flex items-center gap-2 w-full text-left px-3 py-2.5 hover:bg-red-50/50 transition-colors border-b border-red-200">
                    <span className="text-sm font-semibold text-red-800">Unowned</span>
                    <span className="text-xs text-muted-foreground">{unowned.length} cap{unowned.length !== 1 ? 's' : ''}</span>
                    <span className="ml-auto text-muted-foreground">{expandedGroups.has('team:__unowned__') ? <ChevronUp className="w-3.5 h-3.5" /> : <ChevronDown className="w-3.5 h-3.5" />}</span>
                  </button>
                  {expandedGroups.has('team:__unowned__') && <CapGrid caps={unowned} />}
                </div>
              )}
            </div>
          )
        })()}

        {selectedCap && (
          <DetailPanel cap={selectedCap} allCaps={viewData.capabilities} onClose={() => setSelectedCap(null)}
            insight={insights[`cap:${slug(selectedCap.label)}`]} />
        )}
      </ContentContainer>
    </ModelRequired>
  )
}
