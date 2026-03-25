import { useEffect, useState, useMemo, type CSSProperties } from 'react'
import { useNavigate } from 'react-router-dom'
import { GitBranch, Layers, Server, AlertTriangle, Users, ChevronDown, ChevronUp, ArrowRight } from 'lucide-react'
import { api } from '@/lib/api'
import { useRequireModel } from '@/lib/model-context'
import { useSearch, matchesQuery } from '@/lib/search-context'

const VIS_BADGE: Record<string, { bg: string; text: string }> = {
  'user-facing':    { bg: '#dbeafe', text: '#1e40af' },
  'domain':         { bg: '#ede9fe', text: '#5b21b6' },
  'foundational':   { bg: '#d1fae5', text: '#065f46' },
  'infrastructure': { bg: '#f1f5f9', text: '#475569' },
}

const TEAM_TYPE_BADGE: Record<string, { bg: string; text: string }> = {
  'stream-aligned':        { bg: '#dbeafe', text: '#1e40af' },
  'platform':              { bg: '#ede9fe', text: '#5b21b6' },
  'enabling':              { bg: '#d1fae5', text: '#065f46' },
  'complicated-subsystem': { bg: '#fef3c7', text: '#92400e' },
}

interface RealizationViewResponse {
  view_type: string
  no_cap_count: number
  multi_cap_count: number
  service_rows: Array<{
    service: { id: string; label: string; description?: string }
    team: { id: string; label: string; data: { type: string } } | null
    capabilities: Array<{ id: string; label: string; data: { visibility: string } }>
    external_deps?: string[]
  }>
}

interface NeedViewResponse {
  view_type: string
  groups: Array<{
    actor: { id: string; label: string }
    needs: Array<{
      need: { id: string; label: string }
      capabilities: Array<{ id: string; label: string }>
    }>
  }>
}

interface GroupedNeed {
  actor: string
  need: string
  capabilities: Array<{ id: string; label: string; visibility: string }>
  services: string[]
  teams: Array<{ label: string; type: string }>
  isCrossTeam: boolean
  isUnbacked: boolean
}

const thStyle: CSSProperties = {
  textAlign: 'left', padding: '12px 16px', fontSize: 11, fontWeight: 700,
  color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.06em',
  borderBottom: '1px solid #e2e8f0', background: '#f8fafc',
}

export function RealizationView() {
  const { modelId, isHydrating } = useRequireModel()
  const navigate = useNavigate()
  const { query } = useSearch()
  const [viewData, setViewData] = useState<RealizationViewResponse | null>(null)
  const [needData, setNeedData] = useState<NeedViewResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [tab, setTab] = useState<'chain' | 'service'>('chain')
  const [expandedNeed, setExpandedNeed] = useState<string | null>(null)
  const [filterMode, setFilterMode] = useState<'all' | 'cross-team' | 'unbacked'>('all')
  const [collapsedActors, setCollapsedActors] = useState<Set<string>>(new Set())
  const [expandedServiceNeeds, setExpandedServiceNeeds] = useState<Set<string>>(new Set())

  useEffect(() => {
    if (isHydrating || !modelId) { return }
    Promise.all([
      api.getView(modelId, 'realization'),
      api.getView(modelId, 'need'),
    ]).then(([realData, nData]) => {
      setViewData(realData as unknown as RealizationViewResponse)
      setNeedData(nData as unknown as NeedViewResponse)
    }).catch(e => setError((e as Error).message))
      .finally(() => setLoading(false))
  }, [isHydrating, modelId])

  const capToSvcTeam = useMemo(() => {
    if (!viewData) return new Map<string, { services: string[]; teams: Array<{ label: string; type: string }> }>()
    const map = new Map<string, { services: string[]; teams: Array<{ label: string; type: string }> }>()
    for (const row of viewData.service_rows) {
      for (const cap of row.capabilities) {
        const existing = map.get(cap.id) ?? { services: [], teams: [] }
        if (!existing.services.includes(row.service.label)) existing.services.push(row.service.label)
        if (row.team && !existing.teams.some(t => t.label === row.team!.label)) {
          existing.teams.push({ label: row.team.label, type: row.team.data.type })
        }
        map.set(cap.id, existing)
      }
    }
    return map
  }, [viewData])

  const capVisibility = useMemo(() => {
    if (!viewData) return new Map<string, string>()
    const map = new Map<string, string>()
    for (const row of viewData.service_rows) {
      for (const cap of row.capabilities) {
        map.set(cap.id, cap.data.visibility ?? '')
      }
    }
    return map
  }, [viewData])

  const groupedNeeds = useMemo<GroupedNeed[]>(() => {
    if (!needData) return []
    const result: GroupedNeed[] = []
    for (const group of needData.groups) {
      for (const needItem of group.needs) {
        const caps: GroupedNeed['capabilities'] = []
        const allServices = new Set<string>()
        const teamMap = new Map<string, string>()

        for (const cap of needItem.capabilities) {
          const vis = capVisibility.get(cap.id) ?? ''
          caps.push({ id: cap.id, label: cap.label, visibility: vis })

          const svcTeam = capToSvcTeam.get(cap.id)
          if (svcTeam) {
            svcTeam.services.forEach(s => allServices.add(s))
            svcTeam.teams.forEach(t => teamMap.set(t.label, t.type))
          }
        }

        const teams = Array.from(teamMap.entries()).map(([label, type]) => ({ label, type }))
        result.push({
          actor: group.actor.label,
          need: needItem.need.label,
          capabilities: caps,
          services: Array.from(allServices),
          teams,
          isCrossTeam: teams.length > 1,
          isUnbacked: caps.length === 0,
        })
      }
    }
    return result
  }, [needData, capToSvcTeam, capVisibility])

  const filteredNeeds = useMemo(() => {
    let needs = groupedNeeds
    if (filterMode === 'cross-team') needs = needs.filter(n => n.isCrossTeam)
    if (filterMode === 'unbacked') needs = needs.filter(n => n.isUnbacked)
    if (query) {
      needs = needs.filter(n =>
        matchesQuery(n.actor, query) || matchesQuery(n.need, query) ||
        n.capabilities.some(c => matchesQuery(c.label, query)) ||
        n.services.some(s => matchesQuery(s, query)) ||
        n.teams.some(t => matchesQuery(t.label, query))
      )
    }
    return needs
  }, [groupedNeeds, filterMode, query])

  const needsByActor = useMemo(() => {
    const map = new Map<string, GroupedNeed[]>()
    for (const n of filteredNeeds) {
      const arr = map.get(n.actor) ?? []
      arr.push(n)
      map.set(n.actor, arr)
    }
    return map
  }, [filteredNeeds])

  const stats = useMemo(() => {
    const total = groupedNeeds.length
    const mapped = groupedNeeds.filter(n => n.capabilities.length > 0).length
    return {
      total,
      crossTeam: groupedNeeds.filter(n => n.isCrossTeam).length,
      unbacked: groupedNeeds.filter(n => n.isUnbacked).length,
      services: viewData?.service_rows.length ?? 0,
      actors: new Set(groupedNeeds.map(n => n.actor)).size,
      mapped,
      mappedPct: total > 0 ? Math.round(mapped / total * 100) : 0,
    }
  }, [groupedNeeds, viewData])

  const filteredSvc = useMemo(() => {
    if (!viewData) return []
    if (!query) return viewData.service_rows
    return viewData.service_rows.filter(row =>
      matchesQuery(row.service.label, query) ||
      (row.team && matchesQuery(row.team.label, query)) ||
      row.capabilities.some(c => matchesQuery(c.label, query))
    )
  }, [viewData, query])

  if (loading) return (
    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', gap: 14, height: '100%', minHeight: 300 }}>
      <div className="animate-spin" style={{ width: 36, height: 36, borderRadius: '50%', border: '2px solid #e2e8f0', borderTopColor: '#6366f1' }} />
      <span style={{ fontSize: 14, color: '#94a3b8' }}>Loading traceability...</span>
    </div>
  )
  if (error) return (
    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', height: '100%', gap: 8, minHeight: 300 }}>
      <AlertTriangle size={22} style={{ color: '#ef4444' }} />
      <span style={{ fontSize: 14, color: '#ef4444', fontWeight: 600 }}>{error}</span>
    </div>
  )
  if (!viewData) return null

  const pillBase: CSSProperties = { borderRadius: 8, padding: '7px 16px', fontSize: 12, fontWeight: 600, border: 'none', cursor: 'pointer', transition: 'all 0.15s' }
  const pillActive: CSSProperties = { ...pillBase, background: 'linear-gradient(135deg, #6366f1, #8b5cf6)', color: '#fff', boxShadow: '0 2px 6px rgba(99,102,241,0.25)' }
  const pillInactive: CSSProperties = { ...pillBase, background: 'transparent', color: '#64748b' }

  const filterPill = (mode: typeof filterMode, color?: string): CSSProperties => {
    const isActive = filterMode === mode
    return {
      borderRadius: 20, padding: '5px 14px', fontSize: 11, fontWeight: 600,
      cursor: 'pointer', transition: 'all 0.15s', border: 'none',
      background: isActive ? (color ?? '#6366f1') : '#f8fafc',
      color: isActive ? '#fff' : '#64748b',
      boxShadow: isActive ? `0 2px 6px ${color ?? '#6366f1'}40` : 'none',
    }
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16, height: '100%' }}>

      {/* Header */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
        <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 16, flexWrap: 'wrap' }}>
          <div>
            <h1 style={{
              fontSize: 22, fontWeight: 800, letterSpacing: '-0.03em', margin: 0, lineHeight: 1.2,
              background: 'linear-gradient(135deg, #1e293b, #475569)',
              WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent', backgroundClip: 'text',
            }}>
              Realization View
            </h1>
            <p style={{ fontSize: 14, color: '#64748b', marginTop: 6, marginBottom: 0 }}>
              End-to-end value chain traceability
            </p>
          </div>
          <div style={{ display: 'flex', borderRadius: 12, padding: 4, background: '#f1f5f9', border: '1px solid #e2e8f0', gap: 2, flexShrink: 0 }}>
            <button type="button" style={tab === 'chain' ? pillActive : pillInactive} onClick={() => setTab('chain')}>Value Chain</button>
            <button type="button" style={tab === 'service' ? pillActive : pillInactive} onClick={() => setTab('service')}>By Service</button>
          </div>
        </div>

        {/* Stat cards */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(5, 1fr)', gap: 8 }}>
          {[
            { label: 'Needs', value: stats.total, gradient: 'linear-gradient(135deg, #e0e7ff, #f5f3ff)', border: '#c7d2fe', icon: Layers, iconGrad: 'linear-gradient(135deg, #6366f1, #4f46e5)', textColor: '#312e81', labelColor: '#4338ca' },
            { label: 'Cross-Team', value: stats.crossTeam,
              gradient: stats.crossTeam > 0 ? 'linear-gradient(135deg, #fde68a, #fffbeb)' : 'linear-gradient(135deg, #f1f5f9, #f8fafc)',
              border: stats.crossTeam > 0 ? '#fcd34d' : '#e2e8f0',
              icon: GitBranch,
              iconGrad: stats.crossTeam > 0 ? 'linear-gradient(135deg, #f59e0b, #d97706)' : 'linear-gradient(135deg, #94a3b8, #64748b)',
              textColor: stats.crossTeam > 0 ? '#92400e' : '#64748b',
              labelColor: stats.crossTeam > 0 ? '#b45309' : '#94a3b8',
            },
            { label: 'Unbacked', value: stats.unbacked,
              gradient: stats.unbacked > 0 ? 'linear-gradient(135deg, #fecaca, #fef2f2)' : 'linear-gradient(135deg, #d1fae5, #f0fdf4)',
              border: stats.unbacked > 0 ? '#fca5a5' : '#86efac',
              icon: stats.unbacked > 0 ? AlertTriangle : AlertTriangle,
              iconGrad: stats.unbacked > 0 ? 'linear-gradient(135deg, #ef4444, #dc2626)' : 'linear-gradient(135deg, #22c55e, #16a34a)',
              textColor: stats.unbacked > 0 ? '#991b1b' : '#065f46',
              labelColor: stats.unbacked > 0 ? '#b91c1c' : '#047857',
            },
            { label: 'Services', value: stats.services, gradient: 'linear-gradient(135deg, #a7f3d0, #ecfdf5)', border: '#6ee7b7', icon: Server, iconGrad: 'linear-gradient(135deg, #10b981, #059669)', textColor: '#065f46', labelColor: '#047857' },
            { label: 'Mapped', value: stats.mapped,
              gradient: 'linear-gradient(135deg, #d1fae5, #f0fdf4)',
              border: '#86efac',
              icon: Layers,
              iconGrad: 'linear-gradient(135deg, #22c55e, #16a34a)',
              textColor: '#065f46',
              labelColor: '#047857',
              subtitle: `${stats.mappedPct}%`,
            },
          ].map(c => (
            <div key={c.label} style={{
              borderRadius: 12, padding: '10px 12px', background: c.gradient, border: `1px solid ${c.border}`,
              boxShadow: '0 1px 3px rgba(0,0,0,0.04)', display: 'flex', alignItems: 'center', gap: 10,
            }}>
              <div style={{
                width: 32, height: 32, borderRadius: 10, background: c.iconGrad,
                display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0,
              }}>
                <c.icon size={15} color="#fff" strokeWidth={2.25} />
              </div>
              <div style={{ minWidth: 0 }}>
                <div style={{ fontSize: 20, fontWeight: 800, color: c.textColor, lineHeight: 1 }}>{c.value}</div>
                <div style={{ fontSize: 10, fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.04em', color: c.labelColor, marginTop: 2 }}>
                  {c.label}{'subtitle' in c && c.subtitle && <span style={{ fontWeight: 400, color: '#6b7280' }}> · {c.subtitle}</span>}
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Value Chain tab */}
      {tab === 'chain' && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 16, flex: 1, minHeight: 0 }}>
          {/* Filters */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
            <button type="button" style={filterPill('all')} onClick={() => setFilterMode('all')}>
              All needs ({stats.total})
            </button>
            {stats.crossTeam > 0 && (
              <button type="button" style={filterPill('cross-team', '#d97706')} onClick={() => setFilterMode(filterMode === 'cross-team' ? 'all' : 'cross-team')}>
                Cross-team ({stats.crossTeam})
              </button>
            )}
            {stats.unbacked > 0 && (
              <button type="button" style={filterPill('unbacked', '#ef4444')} onClick={() => setFilterMode(filterMode === 'unbacked' ? 'all' : 'unbacked')}>
                Unbacked ({stats.unbacked})
              </button>
            )}
            <span style={{ marginLeft: 'auto', fontSize: 12, color: '#94a3b8' }}>
              {filteredNeeds.length} need{filteredNeeds.length !== 1 ? 's' : ''} across {needsByActor.size} actor{needsByActor.size !== 1 ? 's' : ''}
            </span>
          </div>

          {/* Grouped need cards by actor */}
          <div style={{ flex: 1, overflow: 'auto', display: 'flex', flexDirection: 'column', gap: 10 }}>
            {Array.from(needsByActor.entries()).map(([actor, needs]) => (
              <div key={actor} style={{
                borderRadius: 12, overflow: 'hidden',
                border: '1px solid #e2e8f0',
                background: '#fff',
                boxShadow: '0 1px 3px rgba(0,0,0,0.04)',
              }}>
                {/* Actor header */}
                <div
                  onClick={() => setCollapsedActors(prev => {
                    const next = new Set(prev)
                    if (next.has(actor)) next.delete(actor); else next.add(actor)
                    return next
                  })}
                  style={{
                    padding: '9px 14px', display: 'flex', alignItems: 'center', gap: 8,
                    background: 'linear-gradient(135deg, #eef2ff, #f5f3ff)',
                    borderBottom: '1px solid #e0e7ff',
                    cursor: 'pointer', userSelect: 'none',
                  }}>
                  <span style={{ marginRight: 2, display: 'inline-block', transition: 'transform 0.15s', transform: collapsedActors.has(actor) ? 'rotate(0deg)' : 'rotate(90deg)', fontSize: 10, color: '#6366f1' }}>&#9654;</span>
                  <div style={{
                    width: 24, height: 24, borderRadius: 7,
                    background: 'linear-gradient(135deg, #6366f1, #8b5cf6)',
                    display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0,
                  }}>
                    <Users size={12} color="#fff" />
                  </div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                    <span style={{ fontSize: 13, fontWeight: 700, color: '#312e81' }}>{actor}</span>
                    <span style={{ fontSize: 11, color: '#6366f1' }}>{needs.length} need{needs.length !== 1 ? 's' : ''}</span>
                  </div>
                </div>

                {/* Need rows */}
                {!collapsedActors.has(actor) && <div>
                  {needs.map((n, idx) => {
                    const isExpanded = expandedNeed === `${actor}:${n.need}`
                    const needKey = `${actor}:${n.need}`
                    const isLast = idx === needs.length - 1
                    return (
                      <div key={needKey} style={{ borderBottom: isLast ? 'none' : '1px solid #f1f5f9' }}>
                        {/* Need summary row */}
                        <div
                          onClick={() => setExpandedNeed(isExpanded ? null : needKey)}
                          style={{
                            padding: '9px 14px', cursor: 'pointer', transition: 'background 0.15s',
                            background: n.isCrossTeam ? 'rgba(251, 191, 36, 0.06)' : 'transparent',
                          }}
                          onMouseEnter={e => { (e.currentTarget as HTMLDivElement).style.background = n.isCrossTeam ? 'rgba(251, 191, 36, 0.1)' : 'rgba(99, 102, 241, 0.03)' }}
                          onMouseLeave={e => { (e.currentTarget as HTMLDivElement).style.background = n.isCrossTeam ? 'rgba(251, 191, 36, 0.06)' : 'transparent' }}
                        >
                          <div style={{ display: 'flex', alignItems: 'flex-start', gap: 12 }}>
                            {/* Left: need info */}
                            <div style={{ flex: 1, minWidth: 0 }}>
                              <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 6, flexWrap: 'wrap' }}>
                                <span style={{ fontSize: 14, fontWeight: 600, color: '#1e293b' }}>{n.need}</span>
                                {n.isUnbacked && (
                                  <span style={{ fontSize: 10, fontWeight: 700, color: '#b91c1c', background: '#fee2e2', borderRadius: 20, padding: '2px 8px', border: '1px solid #fca5a5' }}>
                                    UNBACKED
                                  </span>
                                )}
                                {n.isCrossTeam && (
                                  <span style={{ fontSize: 10, fontWeight: 700, color: '#92400e', background: '#fef3c7', borderRadius: 20, padding: '2px 8px', border: '1px solid #fde68a', display: 'inline-flex', alignItems: 'center', gap: 3 }}>
                                    {n.teams.length <= 2
                                      ? n.teams.map(t => t.label).join(', ')
                                      : `${n.teams.slice(0, 2).map(t => t.label).join(', ')}...`}
                                    {' '}
                                    <span title="This need is served by multiple teams — potential fragmentation risk" aria-label="Cross-team warning" style={{ cursor: 'help', display: 'inline-flex' }}><AlertTriangle size={10} /></span>
                                  </span>
                                )}
                              </div>

                              {/* Capability + Service + Team flow */}
                              <div style={{ display: 'flex', alignItems: 'center', gap: 6, flexWrap: 'wrap' }}>
                                {n.capabilities.length > 0 ? (
                                  n.capabilities.map(cap => {
                                    const b = VIS_BADGE[cap.visibility] ?? { bg: '#f1f5f9', text: '#475569' }
                                    return (
                                      <span key={cap.id ?? cap.label} style={{
                                        fontSize: 11, fontWeight: 600, borderRadius: 8, padding: '3px 9px',
                                        background: b.bg, color: b.text, border: `1px solid ${b.text}15`,
                                      }}>{cap.label}</span>
                                    )
                                  })
                                ) : (
                                  <span style={{ fontSize: 11, fontStyle: 'italic', color: '#ef4444' }}>no capability mapped</span>
                                )}
                                {n.capabilities.length > 0 && n.services.length > 0 && (
                                  <ArrowRight size={12} style={{ color: '#cbd5e1', flexShrink: 0 }} />
                                )}
                                {(() => {
                                  const showAllSvc = expandedServiceNeeds.has(needKey)
                                  const svcsToShow = showAllSvc ? n.services : n.services.slice(0, 1)
                                  const remaining = n.services.length - 1
                                  return <>
                                    {svcsToShow.map(s => (
                                      <span key={s} style={{
                                        fontSize: 10, fontWeight: 500, fontFamily: 'ui-monospace, monospace',
                                        borderRadius: 6, padding: '2px 7px', background: '#f1f5f9', color: '#475569', border: '1px solid #e2e8f0',
                                      }}>{s}</span>
                                    ))}
                                    {!showAllSvc && remaining > 0 && (
                                      <button type="button" onClick={e => { e.stopPropagation(); setExpandedServiceNeeds(prev => { const next = new Set(prev); next.add(needKey); return next }) }}
                                        style={{ color: '#3b82f6', background: 'none', border: 'none', cursor: 'pointer', fontSize: 12, padding: 0 }}>
                                        +{remaining} more
                                      </button>
                                    )}
                                    {showAllSvc && remaining > 0 && (
                                      <button type="button" onClick={e => { e.stopPropagation(); setExpandedServiceNeeds(prev => { const next = new Set(prev); next.delete(needKey); return next }) }}
                                        style={{ color: '#94a3b8', background: 'none', border: 'none', cursor: 'pointer', fontSize: 12, padding: 0 }}>
                                        show less
                                      </button>
                                    )}
                                  </>
                                })()}
                                {n.services.length > 0 && n.teams.length > 0 && (
                                  <ArrowRight size={12} style={{ color: '#cbd5e1', flexShrink: 0 }} />
                                )}
                                {n.teams.map(t => {
                                  const badge = TEAM_TYPE_BADGE[t.type] ?? { bg: '#f1f5f9', text: '#475569' }
                                  return (
                                    <span key={t.label} style={{
                                      fontSize: 10, fontWeight: 600, borderRadius: 20, padding: '2px 9px',
                                      background: badge.bg, color: badge.text, border: `1px solid ${badge.text}15`,
                                    }}>{t.label}</span>
                                  )
                                })}
                              </div>
                            </div>

                            {/* Expand chevron */}
                            <div style={{ color: '#94a3b8', flexShrink: 0, marginTop: 4 }}>
                              {isExpanded ? <ChevronUp size={16} /> : <ChevronDown size={16} />}
                            </div>
                          </div>
                        </div>

                        {/* Expanded detail */}
                        {isExpanded && (
                          <div style={{
                            padding: '0 14px 12px', display: 'flex', flexDirection: 'column', gap: 10,
                            background: 'linear-gradient(180deg, rgba(248,250,252,0.5), #fff)',
                          }}>
                            {/* Full traceability table */}
                            <div style={{ borderRadius: 12, overflow: 'hidden', border: '1px solid #e2e8f0' }}>
                              <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 12 }}>
                                <thead>
                                  <tr>
                                    <th style={{ ...thStyle, padding: '8px 12px', fontSize: 10 }}>Capability</th>
                                    <th style={{ ...thStyle, padding: '8px 12px', fontSize: 10 }}>Visibility</th>
                                    <th style={{ ...thStyle, padding: '8px 12px', fontSize: 10 }}>Services</th>
                                    <th style={{ ...thStyle, padding: '8px 12px', fontSize: 10 }}>Owning Teams</th>
                                  </tr>
                                </thead>
                                <tbody>
                                  {n.capabilities.length === 0 ? (
                                    <tr>
                                      <td colSpan={4} style={{ padding: '12px', fontSize: 12, fontStyle: 'italic', color: '#ef4444', textAlign: 'center' }}>
                                        No capability is mapped to this need — it has no implementation path
                                      </td>
                                    </tr>
                                  ) : (
                                    n.capabilities.map((cap, ci) => {
                                      const svcTeam = capToSvcTeam.get(cap.id)
                                      const vis = VIS_BADGE[cap.visibility] ?? { bg: '#f1f5f9', text: '#475569' }
                                      return (
                                        <tr key={ci} style={{ borderBottom: ci < n.capabilities.length - 1 ? '1px solid #f1f5f9' : 'none' }}>
                                          <td style={{ padding: '8px 12px', fontWeight: 600, color: '#1e293b' }}>{cap.label}</td>
                                          <td style={{ padding: '8px 12px' }}>
                                            <span style={{ fontSize: 10, fontWeight: 600, borderRadius: 20, padding: '2px 8px', background: vis.bg, color: vis.text }}>{cap.visibility || '—'}</span>
                                          </td>
                                          <td style={{ padding: '8px 12px' }}>
                                            <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4 }}>
                                              {(svcTeam?.services ?? []).map(s => (
                                                <span key={s} style={{ fontSize: 10, fontFamily: 'ui-monospace, monospace', background: '#f1f5f9', border: '1px solid #e2e8f0', borderRadius: 6, padding: '2px 6px', color: '#475569' }}>{s}</span>
                                              ))}
                                              {(!svcTeam || svcTeam.services.length === 0) && <span style={{ fontSize: 10, color: '#94a3b8', fontStyle: 'italic' }}>—</span>}
                                            </div>
                                          </td>
                                          <td style={{ padding: '8px 12px' }}>
                                            <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4 }}>
                                              {(svcTeam?.teams ?? []).map(t => {
                                                const badge = TEAM_TYPE_BADGE[t.type] ?? { bg: '#f1f5f9', text: '#475569' }
                                                return (
                                                  <span key={t.label} style={{ fontSize: 10, fontWeight: 600, borderRadius: 20, padding: '2px 8px', background: badge.bg, color: badge.text }}>{t.label}</span>
                                                )
                                              })}
                                              {(!svcTeam || svcTeam.teams.length === 0) && <span style={{ fontSize: 10, color: '#94a3b8', fontStyle: 'italic' }}>unowned</span>}
                                            </div>
                                          </td>
                                        </tr>
                                      )
                                    })
                                  )}
                                </tbody>
                              </table>
                            </div>

                            {n.isCrossTeam && (
                              <div style={{
                                borderRadius: 10, padding: '10px 14px', display: 'flex', alignItems: 'flex-start', gap: 8,
                                background: 'linear-gradient(135deg, #fffbeb, #fef3c7)',
                                border: '1px solid #fde68a',
                              }}>
                                <span title="Cross-team coordination risk — multiple teams serve this need" aria-label="Cross-team coordination warning" style={{ cursor: 'help', display: 'inline-flex' }}><AlertTriangle size={14} style={{ color: '#d97706', marginTop: 1, flexShrink: 0 }} /></span>
                                <span style={{ fontSize: 12, color: '#78350f', lineHeight: 1.5 }}>
                                  This need requires coordination across <strong>{n.teams.length} teams</strong> ({n.teams.map(t => t.label).join(', ')}).
                                  Each additional team adds handoff overhead and release coordination risk.
                                </span>
                              </div>
                            )}
                          </div>
                        )}
                      </div>
                    )
                  })}
                </div>}
              </div>
            ))}

            {filteredNeeds.length === 0 && (
              <div style={{
                textAlign: 'center', padding: '48px 0', fontSize: 14, color: '#94a3b8',
                borderRadius: 20, background: 'linear-gradient(135deg, #fff, #f8fafc)',
                border: '1px solid #e2e8f0',
              }}>
                No needs match the current filter.
              </div>
            )}
          </div>
        </div>
      )}

      {/* Service tab */}
      {tab === 'service' && (
        <div style={{
          flex: 1, overflow: 'auto', borderRadius: 16, border: '1px solid #e2e8f0',
          background: 'linear-gradient(135deg, #ffffff, #f8fafc)', boxShadow: '0 1px 3px rgba(0,0,0,0.04)',
        }}>
          <table style={{ width: '100%', fontSize: 13, borderCollapse: 'collapse' }}>
            <thead>
              <tr style={{ position: 'sticky', top: 0, zIndex: 1 }}>
                <th style={{ ...thStyle, width: 200 }}>Service</th>
                <th style={{ ...thStyle, width: 180 }}>Owning Team</th>
                <th style={thStyle}>Capabilities Realized</th>
                <th style={thStyle}>External Deps</th>
              </tr>
            </thead>
            <tbody>
              {filteredSvc.map((row, idx) => {
                const teamType = row.team?.data.type ?? ''
                const teamBadge = TEAM_TYPE_BADGE[teamType] ?? { bg: '#f1f5f9', text: '#475569' }
                const isHighSpan = row.capabilities.length >= 3
                const baseBg = idx % 2 === 0 ? '#ffffff' : '#fafbfc'
                return (
                  <tr key={row.service.id}
                    style={{ borderBottom: '1px solid #f1f5f9', background: baseBg, transition: 'background 0.15s' }}
                    onMouseEnter={e => { (e.currentTarget as HTMLTableRowElement).style.background = 'rgba(99, 102, 241, 0.03)' }}
                    onMouseLeave={e => { (e.currentTarget as HTMLTableRowElement).style.background = baseBg }}
                  >
                    <td style={{ padding: '10px 16px' }}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                        <span style={{ fontFamily: 'ui-monospace, monospace', fontSize: 12, fontWeight: 600, color: '#334155' }}>{row.service.label}</span>
                        {isHighSpan && <span style={{ fontSize: 10, fontWeight: 700, color: '#7c3aed' }}>x{row.capabilities.length}</span>}
                      </div>
                    </td>
                    <td style={{ padding: '10px 16px' }}>
                      {row.team ? (
                        <span style={{ fontSize: 11, fontWeight: 600, borderRadius: 20, padding: '3px 10px', background: teamBadge.bg, color: teamBadge.text }}>{row.team.label}</span>
                      ) : (
                        <span style={{ fontSize: 11, fontStyle: 'italic', color: '#94a3b8' }}>unowned</span>
                      )}
                    </td>
                    <td style={{ padding: '10px 16px' }}>
                      {row.capabilities.length === 0 ? (
                        <span style={{ fontSize: 11, fontStyle: 'italic', color: '#94a3b8' }}>no capabilities</span>
                      ) : (
                        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4 }}>
                          {row.capabilities.map(cap => {
                            const b = VIS_BADGE[cap.data.visibility ?? ''] ?? { bg: '#f1f5f9', text: '#475569' }
                            return <button key={cap.id} type="button"
                              onClick={() => navigate(`/capability?highlight=${encodeURIComponent(cap.label)}`)}
                              title={`${cap.label} (${cap.data.visibility ?? 'unknown'}) — click to view in Capability View`}
                              style={{ display: 'inline-flex', padding: '2px 8px', borderRadius: 9999, fontSize: 11, fontWeight: 600, background: b.bg, color: b.text, border: `1px solid ${b.text}30`, cursor: 'pointer', margin: '2px' }}>
                              {cap.label}
                            </button>
                          })}
                        </div>
                      )}
                    </td>
                    <td style={{ padding: '10px 16px' }}>
                      {(row.external_deps ?? []).length === 0 ? (
                        <span style={{ fontSize: 11, color: '#94a3b8' }}>—</span>
                      ) : (
                        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4 }}>
                          {(row.external_deps ?? []).map((dep, i) => (
                            <span key={i} style={{ fontSize: 10, fontFamily: 'ui-monospace, monospace', background: '#f1f5f9', border: '1px solid #e2e8f0', borderRadius: 6, padding: '2px 6px', color: '#475569' }}>{dep}</span>
                          ))}
                        </div>
                      )}
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
