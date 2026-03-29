import { useEffect, useMemo, useRef, useState } from 'react'
import { api, type OwnershipViewResponse } from '@/lib/api'
import { useRequireModel } from '@/lib/model-context'
import { ModelRequired } from '@/components/ui/ModelRequired'
import { useSearch, matchesQuery } from '@/lib/search-context'
import { AntiPatternPanel } from '@/components/AntiPatternPanel'
import { usePageInsights } from '@/hooks/usePageInsights'
import { LoadingState, ErrorState } from '@/components/ViewState'
import { slug } from '@/lib/slug'
import { VIS_BADGE } from '@/lib/visibility-styles'
import { TEAM_TYPE_BADGE } from '@/lib/team-type-styles'
import { QuickAction } from '@/components/changeset/QuickAction'

const H1_GRADIENT = {
  fontSize: 30,
  fontWeight: 800,
  letterSpacing: '-0.025em',
  lineHeight: 1.15,
  background: 'linear-gradient(135deg, #1e293b 0%, #475569 100%)',
  WebkitBackgroundClip: 'text' as const,
  WebkitTextFillColor: 'transparent' as const,
}

const SUBTITLE = { fontSize: 14, color: '#64748b' } as const

const CARD_SHELL = {
  borderRadius: 20,
  background: 'linear-gradient(135deg, #ffffff 0%, #f8fafc 100%)',
  border: '1px solid #e2e8f0',
  boxShadow: '0 1px 3px rgba(0,0,0,0.04)',
} as const

const SECTION_LABEL = {
  fontSize: 11,
  fontWeight: 600,
  color: '#64748b',
  textTransform: 'uppercase' as const,
  letterSpacing: '0.05em',
} as const

interface NodeDetails {
  id: string; label: string; nodeType: string; data: Record<string, unknown>
}

interface CapViewData {
  parent_groups: Array<{ id: string; label: string; children: string[] }>
  capabilities: Array<{ id: string; label: string; visibility: string }>
}

interface SvcPopover {
  label: string
  teamLabel: string
  teamType: string
  x: number
  y: number
  capList: Array<{ label: string; visibility: string }>
  isHighSpan: boolean
  isFromOtherTeam: boolean
}

const ALL_TEAM_TYPES = ['stream-aligned', 'platform', 'enabling', 'complicated-subsystem'] as const

function laneAccentGradient(isOverloaded: boolean, hasCrossTeam: boolean): string {
  if (isOverloaded) return 'linear-gradient(90deg, #ef4444 0%, #f97316 100%)'
  if (hasCrossTeam) return 'linear-gradient(90deg, #f59e0b 0%, #fbbf24 100%)'
  return 'linear-gradient(90deg, #22c55e 0%, #4ade80 100%)'
}

function PillTabs({
  value,
  onChange,
}: {
  value: 'team' | 'domain'
  onChange: (v: 'team' | 'domain') => void
}) {
  return (
    <div
      className="inline-flex p-1"
      style={{
        borderRadius: 12,
        padding: 4,
        background: '#f1f5f9',
        border: '1px solid #e2e8f0',
      }}
    >
      {([
        { key: 'team' as const, label: 'By Team' },
        { key: 'domain' as const, label: 'By Domain Area' },
      ]).map(opt => {
        const active = value === opt.key
        return (
          <button
            key={opt.key}
            type="button"
            onClick={() => onChange(opt.key)}
            className="px-3 py-2 text-xs font-semibold transition-all whitespace-nowrap"
            style={{
              borderRadius: 8,
              background: active ? 'linear-gradient(135deg, #6366f1 0%, #4f46e5 100%)' : 'transparent',
              color: active ? '#ffffff' : '#64748b',
              boxShadow: active ? '0 2px 8px rgba(99,102,241,0.35)' : 'none',
            }}
          >
            {opt.label}
          </button>
        )
      })}
    </div>
  )
}

export function OwnershipView() {
  const { modelId, isHydrating } = useRequireModel()
  const { query, teamFilter } = useSearch()
  const [viewData, setViewData] = useState<OwnershipViewResponse | null>(null)
  const [capViewData, setCapViewData] = useState<CapViewData | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedNode, setSelectedNode] = useState<NodeDetails | null>(null)
  const [tab, setTab] = useState<'team' | 'domain'>('team')
  const [svcPopover, setSvcPopover] = useState<SvcPopover | null>(null)

  // 4.8.3 — signal filter pills
  const [filterCrossTeam, setFilterCrossTeam] = useState(false)
  const [filterOverloaded, setFilterOverloaded] = useState(false)
  const [filterUnowned, setFilterUnowned] = useState(false)
  const [showProblemsOnly, setShowProblemsOnly] = useState(false)
  const [activeTypeFilters, setActiveTypeFilters] = useState<Set<string>>(new Set())
  const [localSearch, setLocalSearch] = useState('')

  const unownedRef = useRef<HTMLDivElement>(null)
  const { insights } = usePageInsights('ownership')

  useEffect(() => {
    if (isHydrating || !modelId) return
    Promise.all([
      api.getOwnershipView(modelId),
      api.getCapabilityView(modelId),
    ]).then(([ownershipData, capData]) => {
      setViewData(ownershipData)
      setCapViewData(capData as unknown as CapViewData)
    }).catch((e: unknown) => setError((e as Error).message))
      .finally(() => setLoading(false))
  }, [isHydrating, modelId])

  // Scroll unowned into view when unowned filter is toggled on
  useEffect(() => {
    if (filterUnowned && unownedRef.current) {
      unownedRef.current.scrollIntoView({ behavior: 'smooth', block: 'start' })
    }
  }, [filterUnowned])

  const capOwnership = useMemo(() => {
    const map = new Map<string, string[]>()
    viewData?.lanes.forEach(lane => {
      lane.caps.forEach(cg => {
        const existing = map.get(cg.cap.id) ?? []
        map.set(cg.cap.id, [...existing, lane.team.label])
      })
    })
    return map
  }, [viewData])

  // Build a map: service id → all caps that reference it (for enhanced popover)
  const svcCapMap = useMemo(() => {
    const map = new Map<string, Array<{ label: string; visibility: string }>>()
    viewData?.lanes.forEach(lane => {
      lane.caps.forEach(cg => {
        cg.services.forEach(svc => {
          const existing = map.get(svc.id) ?? []
          if (!existing.some(e => e.label === cg.cap.label)) {
            existing.push({ label: cg.cap.label, visibility: cg.cap.data.visibility ?? '' })
          }
          map.set(svc.id, existing)
        })
      })
    })
    return map
  }, [viewData])

  if (loading) return <LoadingState />
  if (error)   return <ErrorState message={error} />
  if (!viewData) return null

  const filteredLanes = viewData.lanes.filter(lane => {
    // Search/team filter from context
    if (teamFilter && !matchesQuery(lane.team.label, teamFilter)) return false
    if (query && !(matchesQuery(lane.team.label, query) || lane.caps.some(cg =>
      matchesQuery(cg.cap.label, query) || cg.services.some(s => matchesQuery(s.label, query))
    ))) return false

    // Local search filter
    if (localSearch) {
      const q = localSearch.toLowerCase()
      const matchesLocal = lane.team.label.toLowerCase().includes(q)
        || lane.caps.some(cg => cg.cap.label.toLowerCase().includes(q))
        || lane.caps.some(cg => cg.services.some(s => s.label.toLowerCase().includes(q)))
      if (!matchesLocal) return false
    }

    // 4.8.3 signal filters
    const isOverloaded = lane.team.data.is_overloaded
    const hasCrossTeam = lane.caps.some(cg => cg.cross_team)
    const hasSignal = isOverloaded || hasCrossTeam

    if (filterCrossTeam && !hasCrossTeam) return false
    if (filterOverloaded && !isOverloaded) return false
    if (showProblemsOnly && !hasSignal) return false

    // 4.8.3 team type filter
    if (activeTypeFilters.size > 0 && !activeTypeFilters.has(lane.team.data.type ?? '')) return false

    return true
  })

  const totalCaps = viewData.lanes.reduce((sum, l) => sum + l.caps.length, 0)
  const totalServices = viewData.service_rows.length
  const crossTeamCount = viewData.cross_team_capabilities.length
  const unownedCount = viewData.unowned_capabilities.length
  const overloadedCount = viewData.overloaded_teams.length

  function toggleTypeFilter(t: string) {
    setActiveTypeFilters(prev => {
      const next = new Set(prev)
      if (next.has(t)) next.delete(t)
      else next.add(t)
      return next
    })
  }

  function openSvcPopover(
    e: React.MouseEvent,
    svc: { id: string; label: string; team_label: string; cap_count: number; team_id: string },
    currentLaneTeamId: string,
  ) {
    e.stopPropagation()
    const rect = (e.currentTarget as HTMLElement).getBoundingClientRect()

    const capList = svcCapMap.get(svc.id) ?? []
    const isHighSpan = svc.cap_count >= 3
    const isFromOtherTeam = Boolean(svc.team_id && svc.team_id !== currentLaneTeamId)

    // Find team type for the owning team
    const owningLane = viewData?.lanes.find(l => l.team.label === svc.team_label)
    const teamType = owningLane?.team.data.type ?? ''

    // Smart positioning
    let x = rect.left
    let y = rect.bottom + 4
    if (y > window.innerHeight * 0.7) y = rect.top - 154
    if (x > window.innerWidth * 0.75) x = rect.right - 280

    setSvcPopover({ label: svc.label, teamLabel: svc.team_label, teamType, x, y, capList, isHighSpan, isFromOtherTeam })
  }

  return (
    <ModelRequired>
      <div className="space-y-6 relative">
      <div className="flex flex-col lg:flex-row lg:items-start lg:justify-between gap-4">
        <div>
          <h1 style={H1_GRADIENT}>Ownership View</h1>
          <p className="mt-1" style={SUBTITLE}>
            Teams, capabilities, and service ownership across the model
          </p>
        </div>
        <PillTabs value={tab} onChange={setTab} />
      </div>

      <div className="rounded-[20px] p-4 space-y-3" style={{ ...CARD_SHELL }}>
        <div className="flex items-center gap-2 flex-wrap" style={{ fontSize: 14, color: '#64748b' }}>
          <span>{viewData.lanes.length} teams</span>
          <span style={{ color: '#cbd5e1' }}>·</span>
          <span>{totalCaps} caps</span>
          <span style={{ color: '#cbd5e1' }}>·</span>
          <span>{totalServices} services</span>

          {crossTeamCount > 0 && (
            <>
              <span style={{ color: '#cbd5e1' }}>·</span>
              <button
                type="button"
                className="font-semibold transition-all"
                style={{
                  fontSize: 11,
                  fontWeight: 600,
                  borderRadius: 20,
                  padding: '5px 12px',
                  background: filterCrossTeam ? '#ea580c' : '#f1f5f9',
                  color: filterCrossTeam ? '#ffffff' : '#64748b',
                  border: filterCrossTeam ? '1px solid #ea580c' : '1px solid #e2e8f0',
                }}
                onClick={() => setFilterCrossTeam(v => !v)}
                title="Filter to cross-team caps"
              >
                {crossTeamCount} cross-team
              </button>
            </>
          )}

          {overloadedCount > 0 && (
            <>
              <span style={{ color: '#cbd5e1' }}>·</span>
              <button
                type="button"
                className="font-semibold transition-all"
                style={{
                  fontSize: 11,
                  fontWeight: 600,
                  borderRadius: 20,
                  padding: '5px 12px',
                  background: filterOverloaded ? '#dc2626' : '#f1f5f9',
                  color: filterOverloaded ? '#ffffff' : '#64748b',
                  border: filterOverloaded ? '1px solid #dc2626' : '1px solid #e2e8f0',
                }}
                onClick={() => setFilterOverloaded(v => !v)}
                title="Filter to overloaded teams"
              >
                {overloadedCount} overloaded
              </button>
            </>
          )}

          {unownedCount > 0 && (
            <>
              <span style={{ color: '#cbd5e1' }}>·</span>
              <button
                type="button"
                className="font-semibold transition-all"
                style={{
                  fontSize: 11,
                  fontWeight: 600,
                  borderRadius: 20,
                  padding: '5px 12px',
                  background: filterUnowned ? '#dc2626' : '#f1f5f9',
                  color: filterUnowned ? '#ffffff' : '#64748b',
                  border: filterUnowned ? '1px solid #dc2626' : '1px solid #e2e8f0',
                }}
                onClick={() => setFilterUnowned(v => !v)}
                title="Scroll to unowned capabilities"
              >
                {unownedCount} unowned
              </button>
            </>
          )}

          {(viewData.external_dependency_count ?? 0) > 0 && (
            <>
              <span style={{ color: '#cbd5e1' }}>·</span>
              <span
                className="font-semibold"
                style={{
                  fontSize: 11,
                  fontWeight: 600,
                  borderRadius: 20,
                  padding: '5px 12px',
                  background: '#f1f5f9',
                  color: '#475569',
                  border: '1px solid #e2e8f0',
                }}
              >
                {viewData.external_dependency_count} external deps
              </span>
            </>
          )}

          <span style={{ flex: 1, minWidth: 8 }} />
          <button
            type="button"
            className="font-semibold text-xs underline-offset-2 hover:underline"
            style={{ color: showProblemsOnly ? '#0f172a' : '#94a3b8' }}
            onClick={() => setShowProblemsOnly(v => !v)}
            title="Show teams with overloaded capabilities, cross-team fragmentation, or cognitive load issues"
          >
            {showProblemsOnly ? 'Show all' : 'Show problems only'}
          </button>
        </div>

        <div className="flex items-center gap-2 flex-wrap pt-1 border-t" style={{ borderColor: '#e2e8f0' }}>
          <span style={{ ...SECTION_LABEL, marginRight: 4 }}>Team type</span>
          {ALL_TEAM_TYPES.map(t => {
            const isActive = activeTypeFilters.has(t)
            return (
              <button
                key={t}
                type="button"
                className="font-semibold transition-all"
                style={{
                  fontSize: 11,
                  fontWeight: 600,
                  borderRadius: 20,
                  padding: '3px 10px',
                  background: isActive ? '#1d4ed8' : 'white',
                  color: isActive ? 'white' : '#374151',
                  border: isActive ? '1px solid #1d4ed8' : '1px solid #d1d5db',
                  cursor: 'pointer',
                }}
                onClick={() => toggleTypeFilter(t)}
              >
                {t.split('-').map(w => w.charAt(0).toUpperCase() + w.slice(1)).join(' ')}
              </button>
            )
          })}
          {activeTypeFilters.size > 0 && (
            <button
              type="button"
              className="text-xs font-medium ml-1"
              style={{ color: '#94a3b8', textDecoration: 'underline' }}
              onClick={() => setActiveTypeFilters(new Set())}
            >
              clear
            </button>
          )}
        </div>
      </div>

      {tab === 'team' && (
        <div>
          <input
            value={localSearch}
            onChange={e => setLocalSearch(e.target.value)}
            placeholder="Filter by team, service, or capability..."
            style={{
              width: '100%', maxWidth: 360, padding: '6px 12px',
              border: '1px solid #d1d5db', borderRadius: 6, fontSize: 13,
              marginBottom: 16, outline: 'none',
            }}
          />
          {filteredLanes.length === 0 && localSearch && (
            <div style={{ textAlign: 'center', color: '#6b7280', padding: 32 }}>
              No teams match &quot;{localSearch}&quot;
              <button onClick={() => setLocalSearch('')} style={{ marginLeft: 8, color: '#3b82f6', background: 'none', border: 'none', cursor: 'pointer' }}>
                Clear
              </button>
            </div>
          )}
          <div className="grid gap-4" style={{ gridTemplateColumns: 'repeat(auto-fill, minmax(380px, 1fr))' }}>
            {filteredLanes.length === 0 && activeTypeFilters.size > 0 && !localSearch && (
              <div style={{ textAlign: 'center', color: '#6b7280', padding: 48, gridColumn: '1/-1' }}>
                <div style={{ fontSize: 14, marginBottom: 8 }}>
                  No {Array.from(activeTypeFilters).join(', ')} teams in this model
                </div>
                <button onClick={() => setActiveTypeFilters(new Set())}
                  style={{ color: '#3b82f6', background: 'none', border: 'none', cursor: 'pointer', fontSize: 13 }}>
                  Clear filter
                </button>
              </div>
            )}
            {filteredLanes.map(lane => {
              const teamType = lane.team.data.type ?? ''
              const isOverloaded = lane.team.data.is_overloaded
              const hasCrossTeam = lane.caps.some(cg => cg.cross_team)
              const topAccent = laneAccentGradient(isOverloaded, hasCrossTeam)
              const badge = TEAM_TYPE_BADGE[teamType] ?? { bg: '#f3f4f6', text: '#374151' }
              const description = lane.team.data.description ?? ''

              const visibleCaps = lane.caps.filter(cg =>
                !query || matchesQuery(cg.cap.label, query) || cg.services.some(s => matchesQuery(s.label, query))
              )

              const totalSvcCount = visibleCaps.reduce((sum, cg) => sum + cg.services.length, 0)

              const teamInsight = insights[`team:${slug(lane.team.label)}`]

              return (
                <div
                  key={lane.team.id}
                  className="overflow-hidden flex flex-col transition-all"
                  style={{
                    borderRadius: 20,
                    background: 'linear-gradient(135deg, #ffffff 0%, #f8fafc 100%)',
                    border: '1px solid #e2e8f0',
                    boxShadow: '0 1px 3px rgba(0,0,0,0.04)',
                  }}
                  onMouseEnter={e => {
                    e.currentTarget.style.transform = 'translateY(-1px)'
                    e.currentTarget.style.boxShadow = '0 12px 32px rgba(15,23,42,0.08)'
                  }}
                  onMouseLeave={e => {
                    e.currentTarget.style.transform = 'translateY(0)'
                    e.currentTarget.style.boxShadow = '0 1px 3px rgba(0,0,0,0.04)'
                  }}
                >
                  <div className="h-1 w-full flex-shrink-0" style={{ background: topAccent }} />

                  <div className="p-5 pb-3" style={{ borderBottom: '1px solid #e2e8f0', background: 'linear-gradient(180deg, #fafafa 0%, #ffffff 100%)' }}>
                    <div className="flex items-start gap-2 flex-wrap mb-2"
                      onMouseEnter={e => { const btn = e.currentTarget.querySelector('.qa-team') as HTMLElement; if (btn) btn.style.opacity = '1' }}
                      onMouseLeave={e => { const btn = e.currentTarget.querySelector('.qa-team') as HTMLElement; if (btn) btn.style.opacity = '0.35' }}
                    >
                      <span className="font-bold text-base tracking-tight" style={{ color: '#0f172a' }}>{lane.team.label}</span>
                      <span className="qa-team" style={{ opacity: 0.35, transition: 'opacity 0.15s' }}>
                        <QuickAction size={12} options={[
                          { label: 'Change team type', action: { type: 'update_team_type', team_name: lane.team.label } },
                          { label: 'Update team size', action: { type: 'update_team_size', team_name: lane.team.label } },
                        ]} />
                      </span>
                      <span
                        className="font-semibold flex-shrink-0"
                        style={{
                          fontSize: 11,
                          fontWeight: 600,
                          borderRadius: 20,
                          padding: '3px 10px',
                          background: badge.bg,
                          color: badge.text,
                          border: `1px solid ${badge.text}22`,
                        }}
                      >
                        {teamType || 'unknown'}
                      </span>
                      {isOverloaded && (
                        <span
                          className="font-semibold flex-shrink-0"
                          style={{
                            fontSize: 11,
                            fontWeight: 600,
                            borderRadius: 20,
                            padding: '3px 10px',
                            background: '#fff7ed',
                            color: '#c2410c',
                            border: '1px solid #fed7aa',
                          }}
                        >
                          overloaded
                        </span>
                      )}
                      {hasCrossTeam && (
                        <span
                          className="font-semibold flex-shrink-0"
                          style={{
                            fontSize: 11,
                            fontWeight: 600,
                            borderRadius: 20,
                            padding: '3px 10px',
                            background: '#fef3c7',
                            color: '#92400e',
                            border: '1px solid #fde68a',
                          }}
                        >
                          cross-team
                        </span>
                      )}
                      <button
                        type="button"
                        className="text-xs hover:opacity-80 ml-auto font-medium"
                        style={{ color: '#94a3b8' }}
                        title="View details"
                        onClick={() => setSelectedNode({ id: lane.team.id, label: lane.team.label, nodeType: 'team', data: { ...lane.team.data, nodeType: 'team' } })}
                      >
                        &#9432;
                      </button>
                    </div>

                    {description && (
                      <p className="text-sm line-clamp-2 mb-2 leading-relaxed" style={{ color: '#64748b' }}>{description}</p>
                    )}

                    <div className="flex items-center gap-3 text-xs font-medium" style={{ color: '#94a3b8' }}>
                      <span>{visibleCaps.length} {visibleCaps.length === 1 ? 'cap' : 'caps'}</span>
                      <span>{totalSvcCount} {totalSvcCount === 1 ? 'service' : 'services'}</span>
                      {(lane.external_deps ?? []).length > 0 && (
                        <span>{(lane.external_deps ?? []).length} ext deps</span>
                      )}
                    </div>
                  </div>

                  {teamInsight && (
                    <div className="px-5 py-3" style={{ background: 'linear-gradient(135deg, #eef2ff 0%, #f8fafc 100%)', borderBottom: '1px solid #e2e8f0' }}>
                      <p className="text-xs leading-relaxed" style={{ color: '#334155' }}>{teamInsight.explanation}</p>
                      {teamInsight.suggestion && (
                        <p className="text-xs leading-relaxed mt-1.5 font-medium" style={{ color: '#4338ca' }}>{teamInsight.suggestion}</p>
                      )}
                    </div>
                  )}

                  <div className="flex-1">
                    {visibleCaps.length === 0 ? (
                      <div className="px-5 py-4 text-sm italic" style={{ color: '#94a3b8' }}>no capabilities owned</div>
                    ) : (
                      visibleCaps.map((cg, idx) => {
                        const isLeaf = cg.cap.data.is_leaf !== false
                        return (
                          <div
                            key={cg.cap.id}
                            className="px-5 py-3"
                            style={{
                              background: cg.cross_team ? 'linear-gradient(90deg, #fffbeb 0%, #ffffff 100%)' : '#ffffff',
                              borderBottom: idx < visibleCaps.length - 1 ? '1px solid #f1f5f9' : 'none',
                            }}
                          >
                            <div className="flex items-center gap-2 mb-2">
                              <button
                                type="button"
                                className="text-left min-w-0 flex-1"
                                title={(cg.cap.data.description as string) || cg.cap.label}
                                onClick={() => setSelectedNode({ id: cg.cap.id, label: cg.cap.label, nodeType: 'capability', data: { ...cg.cap.data, nodeType: 'capability' } })}
                              >
                                <span
                                  className="text-sm font-bold hover:underline"
                                  style={{ color: cg.cross_team ? '#b45309' : '#0f172a' }}
                                >
                                  {cg.cap.label}
                                </span>
                              </button>
                              {!isLeaf && <span className="text-xs flex-shrink-0 font-medium" style={{ color: '#cbd5e1' }}>parent</span>}
                              {cg.cross_team && (() => {
                                const teams = viewData?.cross_team_capabilities.find(ct => ct.cap_id === cg.cap.id)?.team_labels ?? []
                                const teamList = teams.length > 0 ? teams.join(', ') : 'multiple teams'
                                return (
                                  <span
                                    className="text-xs flex-shrink-0"
                                    style={{ color: '#d97706', cursor: 'help' }}
                                    title={`This capability is realized by services from multiple teams: ${teamList}`}
                                    aria-label={`Cross-team warning: realized by ${teamList}`}
                                  >⚠</span>
                                )
                              })()}
                            </div>
                            <div className="flex flex-wrap gap-1.5">
                              {cg.services.length === 0 ? (
                                <span className="text-xs italic" style={{ color: '#94a3b8' }}>
                                  {isLeaf ? 'no services' : 'groups sub-capabilities'}
                                </span>
                              ) : (
                                cg.services.map(svc => {
                                  const isFromOtherTeam = Boolean(svc.team_id && svc.team_id !== lane.team.id)
                                  return (
                                    <span key={svc.id} className="inline-flex items-center gap-0.5"
                                      onMouseEnter={e => { const btn = e.currentTarget.querySelector('.qa-svc') as HTMLElement; if (btn) btn.style.opacity = '1' }}
                                      onMouseLeave={e => { const btn = e.currentTarget.querySelector('.qa-svc') as HTMLElement; if (btn) btn.style.opacity = '0.35' }}
                                    >
                                      <button
                                        type="button"
                                        className="inline-flex items-center font-mono cursor-pointer transition-transform hover:scale-[1.02]"
                                        style={{
                                          fontSize: 11,
                                          fontWeight: 500,
                                          borderRadius: 6,
                                          padding: '4px 10px',
                                          background: isFromOtherTeam ? '#fef3c7' : '#f1f5f9',
                                          border: isFromOtherTeam ? '1px solid #fde68a' : '1px solid #e2e8f0',
                                          color: isFromOtherTeam ? '#92400e' : '#334155',
                                        }}
                                        title={isFromOtherTeam ? `Owned by ${svc.team_label}` : undefined}
                                        onClick={(e) => openSvcPopover(e, svc, lane.team.id)}
                                      >
                                        {svc.label}
                                      </button>
                                      <span className="qa-svc" style={{ opacity: 0.35, transition: 'opacity 0.15s' }}>
                                        <QuickAction size={11} options={[
                                          { label: `Move ${svc.label} to another team`, action: { type: 'move_service', service_name: svc.label, from_team_name: svc.team_label } },
                                          { label: `Rename ${svc.label}`, action: { type: 'rename_service', service_name: svc.label } },
                                        ]} />
                                      </span>
                                    </span>
                                  )
                                })
                              )}
                            </div>
                          </div>
                        )
                      })
                    )}
                    {(lane.external_deps ?? []).length > 0 && (
                      <div className="px-5 py-3 flex items-center gap-2 flex-wrap border-t" style={{ borderColor: '#f1f5f9', background: '#fafafa' }}>
                        <span className="text-xs flex-shrink-0 font-semibold" style={{ ...SECTION_LABEL }}>External →</span>
                        {(lane.external_deps ?? []).map(dep => (
                          <span
                            key={dep.id}
                            className="inline-flex items-center font-mono"
                            style={{
                              fontSize: 11,
                              fontWeight: 500,
                              borderRadius: 6,
                              padding: '3px 8px',
                              background: '#f1f5f9',
                              color: '#334155',
                              border: '1px solid #e2e8f0',
                            }}
                            title={dep.description}
                          >
                            {dep.label}
                            {dep.service_count > 1 && (
                              <span
                                title={dep.description || `Used by ${dep.service_count} services`}
                                style={{ cursor: 'help', color: '#6b7280', fontSize: 12, marginLeft: 4 }}
                              >
                                ({dep.service_count} svc{dep.service_count !== 1 ? 's' : ''})
                              </span>
                            )}
                          </span>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
              )
            })}
          </div>

          {viewData.unowned_capabilities.length > 0 && (
            <div
              ref={unownedRef}
              className="overflow-hidden mt-6"
              style={{
                borderRadius: 20,
                border: '1px solid #fecaca',
                background: 'linear-gradient(135deg, #fff1f2 0%, #ffe4e6 30%, #ffffff 100%)',
                boxShadow: filterUnowned ? '0 0 0 2px #ef4444, 0 8px 30px rgba(239,68,68,0.12)' : '0 1px 3px rgba(0,0,0,0.04)',
                transition: 'box-shadow 0.2s ease',
              }}
            >
              <div className="h-1 w-full" style={{ background: 'linear-gradient(90deg, #ef4444 0%, #f97316 50%, #fb7185 100%)' }} />
              <div className="flex items-center gap-2 px-5 py-4" style={{ borderBottom: '1px solid #fecaca' }}>
                <span className="text-base font-bold" style={{ color: '#9f1239' }}>Unowned Capabilities</span>
                <span className="text-xs font-semibold" style={{ color: '#e11d48' }}>no team assigned</span>
              </div>
              <div className="px-5 py-4 flex flex-wrap gap-2">
                {viewData.unowned_capabilities.filter(c => !query || matchesQuery(c.label, query)).map(c => (
                  <button
                    key={c.id}
                    type="button"
                    onClick={() => setSelectedNode({ id: c.id, label: c.label, nodeType: 'capability', data: { ...c.data, nodeType: 'capability' } })}
                    className="inline-flex items-center text-left font-semibold transition-all hover:translate-y-[-1px]"
                    style={{
                      fontSize: 11,
                      fontWeight: 600,
                      borderRadius: 20,
                      padding: '8px 14px',
                      background: '#ffffff',
                      border: '1px dashed #fca5a5',
                      color: '#be123c',
                      boxShadow: '0 1px 2px rgba(0,0,0,0.04)',
                    }}
                  >
                    {c.label}
                  </button>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {tab === 'domain' && (
        capViewData === null ? (
          <LoadingState />
        ) : (
          <div className="space-y-3">
            {capViewData.parent_groups.map(group => {
              const childCaps = group.children.map(id => {
                const capInfo = capViewData.capabilities.find(c => c.id === id)
                const owners = capOwnership.get(id) ?? []
                return { id, label: capInfo?.label ?? id, visibility: capInfo?.visibility ?? '', owners }
              }).filter(c => !query || matchesQuery(c.label, query))

              if (childCaps.length === 0) return null

              const allOwners = Array.from(new Set(childCaps.flatMap(c => c.owners)))
              const ownerCount = allOwners.length

              let groupAccent = '#22c55e'
              if (ownerCount >= 3) groupAccent = '#ef4444'
              else if (ownerCount === 2) groupAccent = '#f59e0b'

              const hasCrossTeamCap = childCaps.some(c => c.owners.length > 1)

              return (
                <div
                  key={group.id}
                  className="overflow-hidden"
                  style={{ ...CARD_SHELL }}
                >
                  <div className="h-1 w-full" style={{ background: `linear-gradient(90deg, ${groupAccent} 0%, #6366f1 100%)` }} />
                  <div
                    className="px-5 py-4 flex items-center gap-3 flex-wrap"
                    style={{
                      background: 'linear-gradient(135deg, #f8fafc 0%, #ffffff 100%)',
                      borderBottom: '1px solid #e2e8f0',
                    }}
                  >
                    <span className="font-bold text-base" style={{ color: '#0f172a' }}>{group.label}</span>
                    <span
                      className="font-semibold"
                      style={{ fontSize: 11, fontWeight: 600, borderRadius: 20, padding: '3px 10px', background: '#f1f5f9', color: '#64748b' }}
                    >
                      {childCaps.length} caps
                    </span>
                    <span
                      className="font-semibold"
                      style={{ fontSize: 11, fontWeight: 600, borderRadius: 20, padding: '3px 10px', background: '#f1f5f9', color: '#64748b' }}
                    >
                      {ownerCount} {ownerCount === 1 ? 'team' : 'teams'}
                    </span>

                    {hasCrossTeamCap && (
                      <span
                        className="font-semibold"
                        style={{ fontSize: 11, fontWeight: 600, borderRadius: 20, padding: '3px 10px', background: '#fef3c7', color: '#92400e', border: '1px solid #fde68a' }}
                      >
                        cross-team
                      </span>
                    )}

                    <div className="flex gap-1.5 flex-wrap ml-auto">
                      {allOwners.map(owner => {
                        const lane = viewData?.lanes.find(l => l.team.label === owner)
                        const tt = lane?.team.data.type ?? ''
                        const tsBadge = TEAM_TYPE_BADGE[tt] ?? { bg: '#f3f4f6', text: '#374151' }
                        return (
                          <span
                            key={owner}
                            className="font-semibold"
                            style={{ fontSize: 11, fontWeight: 600, borderRadius: 20, padding: '3px 10px', background: tsBadge.bg, color: tsBadge.text, border: `1px solid ${tsBadge.text}22` }}
                          >
                            {owner}
                          </span>
                        )
                      })}
                      {ownerCount === 0 && (
                        <span className="text-xs italic font-medium" style={{ color: '#94a3b8' }}>unowned</span>
                      )}
                    </div>
                  </div>

                  <div>
                    {childCaps.map((cap, idx) => (
                      <div
                        key={cap.id}
                        className="px-5 py-3 flex items-center gap-3 flex-wrap sm:flex-nowrap"
                        style={{
                          background: cap.owners.length > 1 ? 'linear-gradient(90deg, #fffbeb 0%, #ffffff 100%)' : '#ffffff',
                          borderBottom: idx < childCaps.length - 1 ? '1px solid #f1f5f9' : 'none',
                        }}
                      >
                        <button
                          type="button"
                          className="text-sm font-bold flex-1 min-w-0 text-left"
                          style={{ color: '#3b82f6', background: 'none', border: 'none', cursor: 'pointer', textDecoration: 'underline', padding: 0, fontSize: 13 }}
                          onClick={() => setSelectedNode({ id: cap.id, label: cap.label, nodeType: 'capability', data: { visibility: cap.visibility, nodeType: 'capability' } })}
                        >
                          {cap.label}
                        </button>

                        {cap.visibility && (() => {
                          const b = VIS_BADGE[cap.visibility] ?? { bg: '#f3f4f6', text: '#374151' }
                          return (
                            <span
                              className="font-semibold flex-shrink-0"
                              style={{ fontSize: 11, fontWeight: 600, borderRadius: 20, padding: '3px 10px', background: b.bg, color: b.text }}
                            >
                              {cap.visibility}
                            </span>
                          )
                        })()}

                        <div className="flex gap-1.5 flex-wrap flex-shrink-0 justify-end" style={{ minWidth: 120 }}>
                          {cap.owners.length === 0
                            ? <span className="text-xs italic" style={{ color: '#94a3b8' }}>unowned</span>
                            : cap.owners.map(owner => {
                                const lane = viewData?.lanes.find(l => l.team.label === owner)
                                const tt = lane?.team.data.type ?? ''
                                const tsBadge = TEAM_TYPE_BADGE[tt] ?? { bg: '#f3f4f6', text: '#374151' }
                                return (
                                  <button
                                    key={owner}
                                    type="button"
                                    className="font-semibold"
                                    style={{ fontSize: 11, fontWeight: 600, borderRadius: 20, padding: '3px 10px', background: tsBadge.bg, color: tsBadge.text, border: 'none', cursor: 'pointer' }}
                                    onClick={() => { setTab('team'); setLocalSearch(owner) }}
                                  >
                                    {owner}
                                  </button>
                                )
                              })
                          }
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )
            })}

            {(() => {
              const allGroupedIds = new Set(capViewData.parent_groups.flatMap(g => g.children))
              const ungroupedUnowned = viewData.unowned_capabilities.filter(
                c => !allGroupedIds.has(c.id) && (!query || matchesQuery(c.label, query))
              )
              if (ungroupedUnowned.length === 0) return null
              return (
                <div
                  className="overflow-hidden"
                  style={{
                    borderRadius: 20,
                    border: '1px solid #fecaca',
                    background: 'linear-gradient(135deg, #fff1f2 0%, #ffffff 100%)',
                    boxShadow: '0 1px 3px rgba(0,0,0,0.04)',
                  }}
                >
                  <div className="h-1 w-full" style={{ background: 'linear-gradient(90deg, #ef4444 0%, #fb7185 100%)' }} />
                  <div className="flex items-center gap-2 px-5 py-4" style={{ borderBottom: '1px solid #fecaca' }}>
                    <span className="text-base font-bold" style={{ color: '#9f1239' }}>Unowned Capabilities</span>
                    <span className="text-xs font-semibold" style={{ color: '#e11d48' }}>no team or group assigned</span>
                  </div>
                  <div className="px-5 py-4 flex flex-wrap gap-2">
                    {ungroupedUnowned.map(c => (
                      <button
                        key={c.id}
                        type="button"
                        onClick={() => setSelectedNode({ id: c.id, label: c.label, nodeType: 'capability', data: { ...c.data, nodeType: 'capability' } })}
                        className="inline-flex items-center font-semibold transition-all hover:translate-y-[-1px]"
                        style={{
                          fontSize: 11,
                          fontWeight: 600,
                          borderRadius: 20,
                          padding: '8px 14px',
                          background: '#ffffff',
                          border: '1px dashed #fca5a5',
                          color: '#be123c',
                        }}
                      >
                        {c.label}
                      </button>
                    ))}
                  </div>
                </div>
              )
            })()}
          </div>
        )
      )}

      {svcPopover && (
        <>
          <div
            style={{ position: 'fixed', inset: 0, zIndex: 39, background: 'rgba(0,0,0,0.1)', backdropFilter: 'blur(2px)' }}
            onClick={() => setSvcPopover(null)}
            aria-hidden
          />
          <div
            style={{
              position: 'fixed',
              left: svcPopover.x,
              top: svcPopover.y,
              background: 'linear-gradient(135deg, #ffffff 0%, #f8fafc 100%)',
              border: '1px solid #e2e8f0',
              borderRadius: 20,
              boxShadow: '0 12px 40px rgba(15,23,42,0.12), 0 1px 3px rgba(0,0,0,0.04)',
              padding: '14px 16px',
              zIndex: 40,
              minWidth: 240,
              maxWidth: 340,
              fontSize: 12,
            }}
            role="dialog"
            aria-label="Service details"
          >
            <div className="h-0.5 w-full -mt-2 mb-3 rounded-full" style={{ background: 'linear-gradient(90deg, #6366f1, #a855f7)' }} />
            <div className="font-mono font-bold mb-2" style={{ color: '#0f172a', fontSize: 13 }}>
              {svcPopover.label}
            </div>

            <div className="flex items-center gap-1.5 mb-3 flex-wrap" style={{ color: '#64748b' }}>
              <span className="text-xs">Owned by:</span>
              <span className="text-xs font-semibold" style={{ color: '#334155' }}>{svcPopover.teamLabel}</span>
              {svcPopover.teamType && (() => {
                const b = TEAM_TYPE_BADGE[svcPopover.teamType] ?? { bg: '#f3f4f6', text: '#374151' }
                return (
                  <span
                    className="font-semibold"
                    style={{ fontSize: 11, fontWeight: 600, borderRadius: 20, padding: '2px 8px', background: b.bg, color: b.text }}
                  >
                    {svcPopover.teamType}
                  </span>
                )
              })()}
            </div>

            {svcPopover.capList.length > 0 && (
              <div className="mb-2">
                <div className="mb-2" style={{ ...SECTION_LABEL }}>Capabilities ({svcPopover.capList.length})</div>
                <ul className="space-y-1.5">
                  {svcPopover.capList.map((cap, i) => {
                    const b = VIS_BADGE[cap.visibility] ?? { bg: '#f3f4f6', text: '#374151' }
                    return (
                      <li key={i} className="flex items-center gap-2">
                        <span style={{ color: '#cbd5e1' }}>•</span>
                        <span className="text-xs flex-1 min-w-0 font-medium" style={{ color: '#0f172a' }}>{cap.label}</span>
                        {cap.visibility && (
                          <span
                            className="font-semibold flex-shrink-0"
                            style={{ fontSize: 10, fontWeight: 600, borderRadius: 20, padding: '2px 6px', background: b.bg, color: b.text }}
                          >
                            {cap.visibility}
                          </span>
                        )}
                      </li>
                    )
                  })}
                </ul>
              </div>
            )}

            {svcPopover.isHighSpan && (
              <div className="flex items-center gap-1 mt-2 font-semibold" style={{ color: '#b45309', fontSize: 11 }}>
                <span title="This service supports 3 or more capabilities, indicating high span" aria-label="High-span service warning" style={{ cursor: 'help' }}>⚠</span>
                <span>high-span service</span>
              </div>
            )}
            {svcPopover.isFromOtherTeam && (
              <div className="flex items-center gap-1 mt-1 font-semibold" style={{ color: '#b45309', fontSize: 11 }}>
                <span title="This service is owned by a different team than the one owning this capability" aria-label="Cross-team dependency warning" style={{ cursor: 'help' }}>⚠</span>
                <span>cross-team dependency</span>
              </div>
            )}
          </div>
        </>
      )}

      <AntiPatternPanel node={selectedNode} onClose={() => setSelectedNode(null)} />
    </div>
    </ModelRequired>
  )
}
