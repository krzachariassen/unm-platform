import { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { ModelRequired } from '@/components/ui/ModelRequired'
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

type ViewMode = 'visibility' | 'domain' | 'team'

const CARD_SHELL = 'rounded-2xl border border-slate-200 bg-gradient-to-br from-white to-slate-50 overflow-hidden shadow-sm'

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
    select: (data) => {
      // Pre-expand all groups on first load
      const groups = new Set(data.parent_groups.map(g => g.id))
      setExpandedGroups(prev => prev.size === 0 ? groups : prev)
      return data
    },
  })

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

  const GroupHeader = ({ label, count, fragmented, colorAccent, onClick, isExpanded }: { label: string; count: number; fragmented?: number; colorAccent?: string; onClick: () => void; isExpanded: boolean }) => (
    <button type="button" onClick={onClick} className="flex items-center gap-2 w-full text-left px-4 py-3.5 hover:bg-slate-50/80 transition-colors"
      style={{ background: 'linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%)', borderBottom: isExpanded ? '1px solid #e2e8f0' : 'none' }}>
      <div className="w-1 self-stretch rounded-full shrink-0" style={{ background: colorAccent ? `linear-gradient(180deg, ${colorAccent} 0%, ${colorAccent}88 100%)` : 'linear-gradient(180deg, #6366f1 0%, #8b5cf6 100%)', minHeight: 24 }} />
      <span className="text-sm font-bold text-slate-900">{label}</span>
      <span className="text-xs text-slate-400">{count}</span>
      {(fragmented ?? 0) > 0 && <span className="text-[11px] font-semibold rounded-full px-2.5 py-0.5 bg-red-100 text-red-700">{fragmented} fragmented</span>}
      <span className="ml-auto text-xs text-slate-400">{isExpanded ? '▾' : '▸'}</span>
    </button>
  )

  const CapGrid = ({ caps }: { caps: CapabilityType[] }) => (
    <div className="grid gap-3 p-4" style={{ gridTemplateColumns: 'repeat(auto-fill, minmax(260px, 1fr))' }}>
      {caps.map(cap => <CapabilityCard key={cap.id} cap={cap} onClick={() => setSelectedCap(cap)} />)}
    </div>
  )

  return (
    <ModelRequired>
      <div className="space-y-6">
        <PageHeader
          title="Capability View"
          description={`${viewData.parent_groups.length} domain groups · ${viewData.leaf_capability_count} capabilities`}
          actions={
            <div className="flex rounded-lg overflow-hidden border border-border">
              {(['visibility', 'domain', 'team'] as ViewMode[]).map(m => (
                <button key={m} onClick={() => setViewMode(m)}
                  className={cn('px-4 py-2 text-sm font-medium capitalize transition-colors border-l border-border first:border-l-0',
                    viewMode === m ? 'bg-slate-900 text-white' : 'bg-white text-slate-600 hover:bg-slate-50')}>
                  {m === 'visibility' ? 'By Visibility' : m === 'domain' ? 'By Domain' : 'By Team'}
                </button>
              ))}
            </div>
          }
        />

        {/* Signals */}
        {fragmentedCount === 0 && disconnectedCount === 0 && highSpanCount === 0 && atRiskUserFacing.length === 0 ? (
          <div className="flex items-center gap-2 px-4 py-3 rounded-lg bg-green-50 border border-green-200">
            <span className="text-green-600">✓</span>
            <span className="text-sm text-green-700 font-medium">No architecture issues detected — capabilities are well-structured</span>
          </div>
        ) : (
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
            {[
              { value: fragmentedCount, label: 'Fragmented', alertColor: '#be123c', okColor: '#047857' },
              { value: disconnectedCount, label: 'Unowned', alertColor: '#c2410c', okColor: '#047857' },
              { value: highSpanCount, label: 'High-span', alertColor: '#a16207', okColor: '#047857' },
              { value: atRiskUserFacing.length, label: 'User-facing at risk', alertColor: '#be185d', okColor: '#047857' },
            ].map(({ value, label, alertColor, okColor }) => (
              <div key={label} className="rounded-2xl p-4 border border-slate-200" style={{ background: value > 0 ? 'linear-gradient(135deg, #fef2f2 0%, #ffe4e6 100%)' : 'linear-gradient(135deg, #ecfdf5 0%, #d1fae5 100%)' }}>
                <div className="text-2xl font-extrabold tabular-nums" style={{ color: value > 0 ? alertColor : okColor }}>{value}</div>
                <div className="text-[11px] font-semibold uppercase tracking-wider mt-1" style={{ color: value > 0 ? alertColor : okColor, opacity: 0.7 }}>{label}</div>
              </div>
            ))}
          </div>
        )}

        {/* AI insight */}
        {dashInsight && (
          <div className="rounded-2xl p-5 bg-gradient-to-br from-indigo-50 to-slate-50 border border-indigo-100">
            <p className="text-[11px] font-semibold text-indigo-600 uppercase tracking-wider mb-2">AI Capability Analysis</p>
            <p className="text-sm leading-relaxed text-slate-700">{dashInsight.explanation}</p>
            {dashInsight.suggestion && <p className="text-sm leading-relaxed text-indigo-800 mt-2 font-medium">{dashInsight.suggestion}</p>}
          </div>
        )}

        {/* At-risk user-facing */}
        {atRiskUserFacing.length > 0 && (
          <div className="rounded-2xl p-5 bg-gradient-to-br from-red-50 to-white border border-red-200 relative overflow-hidden">
            <div className="absolute left-0 top-0 bottom-0 w-1 bg-gradient-to-b from-red-500 to-orange-500" />
            <div className="pl-3">
              <h3 className="text-sm font-bold text-rose-800 mb-3">User-Facing Capabilities Served by Multiple Teams</h3>
              <div className="space-y-3">
                {atRiskUserFacing.map(cap => (
                  <div key={cap.id} className="flex items-center gap-2 flex-wrap">
                    <span className="text-sm font-semibold text-slate-900">{cap.label}</span>
                    {cap.teams.map(team => (
                      <span key={team.id} className="text-[11px] font-semibold rounded-full px-3 py-1 bg-indigo-600 text-white">{team.label}</span>
                    ))}
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}

        {/* Fragmented/high-span pills */}
        {(fragmentedCount > 0 || highSpanCount > 0) && (
          <div className="flex flex-wrap gap-2">
            {viewData.fragmented_capabilities.map(fc => (
              <div key={fc.id} className="flex items-center gap-1.5 text-[11px] font-semibold rounded-full px-3.5 py-1.5 bg-red-100 text-red-700 border border-red-200">
                <span>{fc.label}</span><span className="text-red-400">· {fc.team_count} teams</span>
              </div>
            ))}
            {viewData.high_span_services.map(hs => (
              <div key={hs.name} className="flex items-center gap-1.5 text-[11px] font-semibold font-mono rounded-full px-3.5 py-1.5 bg-amber-50 text-amber-700 border border-amber-200">
                <span>{hs.name}</span><span className="text-amber-500">· {hs.capability_count} caps</span>
              </div>
            ))}
          </div>
        )}

        {/* Visibility view */}
        {viewMode === 'visibility' && (
          <div className="space-y-10">
            {VIS_BANDS.map(band => {
              const bandCaps = viewData.capabilities.filter(c => c.visibility === band.key && c.is_leaf && matchesCap(c))
              if (bandCaps.length === 0) return null
              return (
                <div key={band.key}>
                  <div className="flex items-center gap-3 mb-4">
                    <div className="h-px flex-1" style={{ background: `linear-gradient(90deg, transparent, ${band.border}, transparent)` }} />
                    <span className="text-[11px] font-bold uppercase tracking-wider rounded-full px-3.5 py-1.5 whitespace-nowrap"
                      style={{ background: band.bg, color: band.accent, border: `1px solid ${band.border}` }}>
                      {band.label} · {bandCaps.length}
                    </span>
                    <div className="h-px flex-1" style={{ background: `linear-gradient(90deg, transparent, ${band.border}, transparent)` }} />
                  </div>
                  <div className="grid gap-3" style={{ gridTemplateColumns: 'repeat(auto-fill, minmax(260px, 1fr))' }}>
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
              const isExpanded = expandedGroups.has(pg.id)
              return (
                <div key={pg.id} className={CARD_SHELL}>
                  <GroupHeader label={pg.label} count={groupCaps.length} fragmented={groupCaps.filter(c => c.is_fragmented).length} onClick={() => toggleGroup(pg.id)} isExpanded={isExpanded} />
                  {isExpanded && <CapGrid caps={groupCaps} />}
                </div>
              )
            })}
            {(() => {
              const uncategorized = viewData.capabilities.filter(c => c.is_leaf && !capToParent.has(c.id) && matchesCap(c))
              if (uncategorized.length === 0) return null
              const isExpanded = expandedGroups.has('__uncategorized__')
              return (
                <div className={CARD_SHELL}>
                  <GroupHeader label="Uncategorized" count={uncategorized.length} colorAccent="#94a3b8" onClick={() => toggleGroup('__uncategorized__')} isExpanded={isExpanded} />
                  {isExpanded && <CapGrid caps={uncategorized} />}
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
                const isExpanded = expandedGroups.has(`team:${teamName}`)
                return (
                  <div key={teamName} className={CARD_SHELL}>
                    <button type="button" onClick={() => toggleGroup(`team:${teamName}`)} className="flex items-center gap-2 w-full text-left px-4 py-3.5 hover:bg-slate-50/80 transition-colors"
                      style={{ background: 'linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%)', borderBottom: isExpanded ? '1px solid #e2e8f0' : 'none' }}>
                      <div className="w-1 self-stretch rounded-full shrink-0" style={{ background: `linear-gradient(180deg, ${badge.accent} 0%, ${badge.accent}88 100%)`, minHeight: 24 }} />
                      <span className="text-sm font-bold text-slate-900">{teamName}</span>
                      <span className="text-xs font-semibold px-2 py-0.5 rounded-full" style={{ background: badge.bg, color: badge.text }}>{type}</span>
                      <span className="text-xs text-slate-400">{caps.length} cap{caps.length !== 1 ? 's' : ''}</span>
                      {caps.filter(c => c.is_fragmented).length > 0 && <span className="text-[11px] font-semibold rounded-full px-2.5 py-0.5 bg-red-100 text-red-700">{caps.filter(c => c.is_fragmented).length} fragmented</span>}
                      <span className="ml-auto text-xs text-slate-400">{isExpanded ? '▾' : '▸'}</span>
                    </button>
                    {isExpanded && <CapGrid caps={caps} />}
                  </div>
                )
              })}
              {unowned.length > 0 && (() => {
                const isExpanded = expandedGroups.has('team:__unowned__')
                return (
                  <div className={CARD_SHELL} style={{ border: '1px solid #fecaca' }}>
                    <button type="button" onClick={() => toggleGroup('team:__unowned__')} className="flex items-center gap-2 w-full text-left px-4 py-3.5 hover:bg-red-50/60 transition-colors"
                      style={{ background: 'linear-gradient(135deg, #fff1f2 0%, #ffe4e6 100%)', borderBottom: isExpanded ? '1px solid #fecaca' : 'none' }}>
                      <div className="w-1 self-stretch rounded-full shrink-0" style={{ background: 'linear-gradient(180deg, #ef4444, #f87171)', minHeight: 24 }} />
                      <span className="text-sm font-bold text-rose-800">Unowned</span>
                      <span className="text-xs text-slate-400">{unowned.length} cap{unowned.length !== 1 ? 's' : ''}</span>
                      <span className="ml-auto text-xs text-slate-400">{isExpanded ? '▾' : '▸'}</span>
                    </button>
                    {isExpanded && <CapGrid caps={unowned} />}
                  </div>
                )
              })()}
            </div>
          )
        })()}

        {selectedCap && (
          <DetailPanel cap={selectedCap} allCaps={viewData.capabilities} onClose={() => setSelectedCap(null)}
            insight={insights[`cap:${slug(selectedCap.label)}`]} />
        )}
      </div>
    </ModelRequired>
  )
}
