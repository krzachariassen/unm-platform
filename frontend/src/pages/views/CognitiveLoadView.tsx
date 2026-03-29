import { useMemo, useState, type CSSProperties } from 'react'
import {
  AlertTriangle, Gauge, Users, Box, GitBranch, Zap,
  ChevronDown, ChevronUp, Shield, Layers, ArrowRight,
} from 'lucide-react'
import { api } from '@/lib/api'
import { ModelRequired } from '@/components/ui/ModelRequired'
import { usePageInsights } from '@/hooks/usePageInsights'
import { useModelView } from '@/hooks/useModelView'
import { LoadingState, ErrorState } from '@/components/ViewState'
import { slug } from '@/lib/slug'
import type { TeamLoad, CognitiveLoadViewResponse, InsightItem } from '@/lib/api'

const TEAM_TYPE_BADGE: Record<string, { bg: string; text: string; gradient: string }> = {
  'stream-aligned':        { bg: '#dbeafe', text: '#1e40af', gradient: 'linear-gradient(135deg, #3b82f6 0%, #2563eb 100%)' },
  'platform':              { bg: '#ede9fe', text: '#5b21b6', gradient: 'linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%)' },
  'enabling':              { bg: '#d1fae5', text: '#065f46', gradient: 'linear-gradient(135deg, #10b981 0%, #059669 100%)' },
  'complicated-subsystem': { bg: '#fef3c7', text: '#92400e', gradient: 'linear-gradient(135deg, #f59e0b 0%, #d97706 100%)' },
}

const LEVEL = {
  low:    { bg: '#f0fdf4', text: '#15803d', border: '#86efac', dot: '#22c55e', label: 'Low',    barColor: '#22c55e', barBg: '#dcfce7', gradient: 'linear-gradient(135deg, #22c55e 0%, #16a34a 100%)' },
  medium: { bg: '#fffbeb', text: '#92400e', border: '#fcd34d', dot: '#f59e0b', label: 'Medium', barColor: '#f59e0b', barBg: '#fef3c7', gradient: 'linear-gradient(135deg, #f59e0b 0%, #d97706 100%)' },
  high:   { bg: '#fef2f2', text: '#b91c1c', border: '#fca5a5', dot: '#ef4444', label: 'High',   barColor: '#ef4444', barBg: '#fee2e2', gradient: 'linear-gradient(135deg, #ef4444 0%, #dc2626 100%)' },
} as const

const DIMENSIONS = [
  { key: 'domain_spread'    as const, abbr: 'Dom',  label: 'Domain Spread',      desc: 'Capability count',                  icon: Layers,    thresholds: '1-3 low · 4-5 med · 6+ high', maxVal: 10 },
  { key: 'service_load'     as const, abbr: 'Svc',  label: 'Service Load',       desc: 'Services ÷ team size (per person)', icon: Box,       thresholds: '≤2 low · 2-3 med · >3 high',  maxVal: 6  },
  { key: 'interaction_load' as const, abbr: 'Ixn',  label: 'Interaction Load',   desc: 'Weighted interaction score',        icon: GitBranch, thresholds: '≤3 low · 4-6 med · 7+ high',  maxVal: 12 },
  { key: 'dependency_load'  as const, abbr: 'Deps', label: 'Dependency Fan-out', desc: 'Outbound service dependencies',     icon: Zap,       thresholds: '≤4 low · 5-8 med · 9+ high',  maxVal: 15 },
] as const

type SortKey = 'level' | 'name' | 'type'
const TEAM_TYPES = ['stream-aligned', 'platform', 'enabling', 'complicated-subsystem']
const levelRank = (l: string) => l === 'high' ? 3 : l === 'medium' ? 2 : 1




function Speedometer({ level, dimensions }: {
  level: string
  dimensions: Array<{ abbr: string; level: string; value: number }>
}) {
  const s = LEVEL[level as keyof typeof LEVEL] ?? LEVEL.low
  const needle = level === 'high' ? 0.85 : level === 'medium' ? 0.5 : 0.2
  const r = 44
  const cx = 52
  const cy = 52
  const startAngle = -210
  const sweep = 240
  const endAngle = startAngle + sweep

  const arcPath = (startDeg: number, endDeg: number, radius: number) => {
    const s1 = (startDeg * Math.PI) / 180
    const e1 = (endDeg * Math.PI) / 180
    const x1 = cx + radius * Math.cos(s1)
    const y1 = cy + radius * Math.sin(s1)
    const x2 = cx + radius * Math.cos(e1)
    const y2 = cy + radius * Math.sin(e1)
    const large = endDeg - startDeg > 180 ? 1 : 0
    return `M ${x1} ${y1} A ${radius} ${radius} 0 ${large} 1 ${x2} ${y2}`
  }

  const needleAngle = startAngle + sweep * needle
  const needleRad = (needleAngle * Math.PI) / 180
  const nx = cx + (r - 14) * Math.cos(needleRad)
  const ny = cy + (r - 14) * Math.sin(needleRad)

  const greenEnd = startAngle + sweep * 0.33
  const yellowEnd = startAngle + sweep * 0.66

  return (
    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
      <svg width={104} height={72} viewBox="0 0 104 72">
        {/* Track segments: green, yellow, red */}
        <path d={arcPath(startAngle, greenEnd, r)} fill="none" stroke="#dcfce7" strokeWidth={7} strokeLinecap="round" />
        <path d={arcPath(greenEnd, yellowEnd, r)} fill="none" stroke="#fef3c7" strokeWidth={7} strokeLinecap="round" />
        <path d={arcPath(yellowEnd, endAngle, r)} fill="none" stroke="#fee2e2" strokeWidth={7} strokeLinecap="round" />

        {/* Active arc */}
        <path d={arcPath(startAngle, needleAngle, r)} fill="none" stroke={s.dot} strokeWidth={7} strokeLinecap="round"
          style={{ filter: `drop-shadow(0 0 4px ${s.dot}50)` }} />

        {/* Needle */}
        <line x1={cx} y1={cy} x2={nx} y2={ny} stroke={s.dot} strokeWidth={2.5} strokeLinecap="round"
          style={{ filter: `drop-shadow(0 0 3px ${s.dot}40)` }} />
        <circle cx={cx} cy={cy} r={4} fill={s.dot} />
        <circle cx={cx} cy={cy} r={2} fill="#fff" />

        {/* Label */}
        <text x={cx} y={cy - 10} textAnchor="middle" style={{ fontSize: 14, fontWeight: 800, fill: s.text }}>{s.label}</text>
      </svg>

      {/* Dimension dots below gauge */}
      <div style={{ display: 'flex', gap: 8, marginTop: 2 }}>
        {dimensions.map(d => {
          const ds = LEVEL[d.level as keyof typeof LEVEL] ?? LEVEL.low
          return (
            <div key={d.abbr} style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 2 }}>
              <div style={{
                width: 22, height: 22, borderRadius: 6, background: ds.bg, border: `1.5px solid ${ds.border}`,
                display: 'flex', alignItems: 'center', justifyContent: 'center',
              }}>
                <span style={{ fontSize: 9, fontWeight: 800, color: ds.text, fontFamily: 'ui-monospace, monospace' }}>{d.value}</span>
              </div>
              <span style={{ fontSize: 7, fontWeight: 700, color: '#94a3b8', textTransform: 'uppercase', letterSpacing: '0.03em' }}>{d.abbr}</span>
            </div>
          )
        })}
      </div>
    </div>
  )
}

function TeamCard({ tl, insight, isExpanded, onToggle }: {
  tl: TeamLoad; insight?: InsightItem; isExpanded: boolean; onToggle: () => void
}) {
  const badge = TEAM_TYPE_BADGE[tl.team.type] ?? { bg: '#f1f5f9', text: '#475569', gradient: 'linear-gradient(135deg, #94a3b8, #64748b)' }
  const levelStyle = LEVEL[tl.overall_level as keyof typeof LEVEL] ?? LEVEL.low

  const dims = DIMENSIONS.map(d => ({ abbr: d.abbr, level: tl[d.key].level, value: tl[d.key].value }))

  return (
    <div style={{
      borderRadius: 16,
      overflow: 'hidden',
      border: `1px solid ${isExpanded ? levelStyle.border : '#e2e8f0'}`,
      background: '#fff',
      boxShadow: isExpanded ? `0 4px 16px ${levelStyle.dot}12, 0 1px 3px rgba(0,0,0,0.04)` : '0 1px 3px rgba(0,0,0,0.04)',
      transition: 'all 0.25s ease',
    }}>
      <div style={{ height: 3, background: levelStyle.gradient }} />

      <div
        onClick={onToggle}
        style={{ padding: '14px 16px', cursor: 'pointer', transition: 'background 0.15s' }}
        onMouseEnter={e => { (e.currentTarget as HTMLDivElement).style.background = '#fafbfc' }}
        onMouseLeave={e => { (e.currentTarget as HTMLDivElement).style.background = 'transparent' }}
      >
        {/* Header: name + badges */}
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 8, marginBottom: 6 }}>
          <h3 style={{
            fontSize: 14, fontWeight: 700, color: '#1e293b', margin: 0,
            fontFamily: 'ui-monospace, monospace', letterSpacing: '-0.01em',
            overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', minWidth: 0,
          }}>
            {tl.team.name}
          </h3>
          <div style={{ color: '#c4c9d4', flexShrink: 0 }}>{isExpanded ? <ChevronUp size={14} /> : <ChevronDown size={14} />}</div>
        </div>

        {/* Tags row */}
        <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 12, flexWrap: 'wrap' }}>
          <span style={{
            display: 'inline-flex', alignItems: 'center', gap: 4,
            background: badge.bg, color: badge.text, borderRadius: 20, padding: '2px 9px', fontSize: 10, fontWeight: 700,
          }}>
            <span style={{ width: 5, height: 5, borderRadius: '50%', background: badge.text, opacity: 0.5 }} />
            {tl.team.type}
          </span>
          {!tl.size_is_explicit && (
            <span style={{ fontSize: 9, color: '#d97706', background: '#fffbeb', border: '1px solid #fde68a', borderRadius: 20, padding: '1px 6px', fontWeight: 600 }}>no size</span>
          )}
          <span style={{ fontSize: 10, color: '#94a3b8', marginLeft: 'auto' }}>
            <span style={{ fontWeight: 700, color: '#64748b', fontFamily: 'ui-monospace, monospace' }}>{tl.service_count}</span> svc
            {' · '}
            <span style={{ fontWeight: 700, color: '#64748b', fontFamily: 'ui-monospace, monospace' }}>{tl.capability_count}</span> cap
            {tl.size_is_explicit && <>
              {' · '}
              <span style={{ fontWeight: 700, color: '#64748b', fontFamily: 'ui-monospace, monospace' }}>{tl.team_size}</span> ppl
            </>}
          </span>
        </div>

        {/* Speedometer gauge */}
        <Speedometer level={tl.overall_level} dimensions={dims} />
      </div>

      {/* Expanded: dimension breakdown + AI insight + capabilities/services */}
      {isExpanded && (
        <div style={{ borderTop: `1px solid ${levelStyle.border}30`, padding: '14px 16px', background: 'linear-gradient(180deg, #fafbfc 0%, #fff 100%)' }}>
          {/* AI Insight */}
          {insight && (
            <div style={{
              borderRadius: 12, padding: 14, marginBottom: 12,
              background: 'linear-gradient(135deg, #eff6ff, #e0f2fe)',
              border: '1px solid #bae6fd',
            }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 8 }}>
                <Shield size={12} style={{ color: '#0ea5e9' }} />
                <span style={{ fontSize: 10, fontWeight: 700, color: '#0369a1', textTransform: 'uppercase', letterSpacing: '0.04em' }}>AI Recommendation</span>
              </div>
              <p style={{ fontSize: 12, color: '#334155', lineHeight: 1.6, margin: 0 }}>{insight.explanation}</p>
              {insight.suggestion && (
                <div style={{ marginTop: 8, padding: '8px 10px', borderRadius: 8, background: 'rgba(255,255,255,0.6)', border: '1px solid #bae6fd' }}>
                  <div style={{ display: 'flex', alignItems: 'flex-start', gap: 6 }}>
                    <ArrowRight size={11} style={{ color: '#1e40af', marginTop: 2, flexShrink: 0 }} />
                    <p style={{ fontSize: 11, color: '#1e40af', lineHeight: 1.55, margin: 0, fontWeight: 500 }}>{insight.suggestion}</p>
                  </div>
                </div>
              )}
            </div>
          )}

          {/* Dimension breakdown table */}
          <div style={{ marginBottom: 12, borderTop: '1px solid #f3f4f6', paddingTop: 12 }}>
            <div style={{ fontSize: 11, fontWeight: 600, color: '#9ca3af', textTransform: 'uppercase', marginBottom: 8 }}>Load Dimensions</div>
            {DIMENSIONS.map(d => {
              const dim = tl[d.key]
              const ds = LEVEL[dim.level as keyof typeof LEVEL] ?? LEVEL.low
              return (
                <div key={d.abbr} style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '4px 0' }}>
                  <span style={{ fontSize: 13, color: '#374151' }}>{d.label}</span>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                    <span style={{ fontSize: 13, fontWeight: 600 }}>{dim.value}</span>
                    <span style={{
                      fontSize: 11, padding: '1px 6px', borderRadius: 9999,
                      background: ds.bg, color: ds.text,
                    }}>{ds.label.toLowerCase()}</span>
                    <span style={{ fontSize: 11, color: '#9ca3af' }}>{d.thresholds}</span>
                  </div>
                </div>
              )
            })}
          </div>

          {/* Capabilities & Services */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
            {tl.capabilities && tl.capabilities.length > 0 && (
              <div>
                <div style={{ fontSize: 10, fontWeight: 700, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: 5 }}>
                  Capabilities ({tl.capabilities.length})
                </div>
                <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4 }}>
                  {tl.capabilities.map(c => (
                    <span key={c} style={{ fontSize: 10, fontWeight: 500, color: '#475569', background: '#f1f5f9', border: '1px solid #e2e8f0', borderRadius: 6, padding: '2px 8px' }}>{c}</span>
                  ))}
                </div>
              </div>
            )}
            {tl.services && tl.services.length > 0 && (
              <div>
                <div style={{ fontSize: 10, fontWeight: 700, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: 5 }}>
                  Services ({tl.services.length})
                </div>
                <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4 }}>
                  {tl.services.map(s => (
                    <span key={s} style={{ fontSize: 10, fontWeight: 500, color: '#475569', background: '#f8fafc', border: '1px solid #e2e8f0', borderRadius: 6, padding: '2px 8px', fontFamily: 'ui-monospace, monospace' }}>{s}</span>
                  ))}
                </div>
              </div>
            )}
          </div>

          {!tl.size_is_explicit && (
            <div style={{ marginTop: 10, borderRadius: 8, padding: '8px 10px', background: '#fffbeb', border: '1px solid #fde68a' }}>
              <p style={{ fontSize: 11, color: '#92400e', margin: 0, lineHeight: 1.45 }}>
                Missing <code style={{ background: 'rgba(255,255,255,0.6)', padding: '1px 4px', borderRadius: 3, fontFamily: 'ui-monospace, monospace', border: '1px solid #fcd34d', fontSize: 10 }}>size:</code> — defaults to 5 people.
              </p>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

export function CognitiveLoadView() {
  const { data: viewData, loading, error } = useModelView<CognitiveLoadViewResponse>(api.getCognitiveLoadView.bind(api))
  const loads = useMemo(() => viewData?.team_loads ?? [], [viewData])
  const [sortKey, setSortKey] = useState<SortKey>('level')
  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('desc')
  const [filterType, setFilterType] = useState<string>('')
  const [expandedTeam, setExpandedTeam] = useState<string | null>(null)
  const [showThresholds, setShowThresholds] = useState(false)
  const { insights } = usePageInsights('cognitive-load')

  const counts = useMemo(() => ({
    total: loads.length,
    high: loads.filter(t => t.overall_level === 'high').length,
    medium: loads.filter(t => t.overall_level === 'medium').length,
    low: loads.filter(t => t.overall_level === 'low').length,
    missingSize: loads.filter(t => !t.size_is_explicit).length,
  }), [loads])

  const teamsByLevel = useMemo(() => ({
    high: loads.filter(t => t.overall_level === 'high').map(t => t.team.name),
    medium: loads.filter(t => t.overall_level === 'medium').map(t => t.team.name),
    low: loads.filter(t => t.overall_level === 'low').map(t => t.team.name),
  }), [loads])

  const sorted = useMemo(() => {
    const filtered = filterType ? loads.filter(t => t.team.type === filterType) : loads
    const dir = sortDir === 'desc' ? 1 : -1
    return [...filtered].sort((a, b) => {
      switch (sortKey) {
        case 'level': return dir * (levelRank(b.overall_level) - levelRank(a.overall_level))
        case 'name':  return dir * a.team.name.localeCompare(b.team.name)
        case 'type':  return dir * a.team.type.localeCompare(b.team.type)
      }
    })
  }, [loads, filterType, sortKey, sortDir])

  if (loading) return <LoadingState message="Analyzing structural cognitive load…" />
  if (error) return <ErrorState message={error} />

  const pillBase: CSSProperties = { borderRadius: 8, padding: '6px 14px', fontSize: 12, fontWeight: 600, border: 'none', cursor: 'pointer', transition: 'all 0.15s' }
  const pillActiveSorted: CSSProperties = { ...pillBase, background: '#1d4ed8', color: '#fff', boxShadow: '0 2px 6px rgba(29,78,216,0.25)' }
  const pillInactive: CSSProperties = { ...pillBase, background: 'transparent', color: '#64748b' }

  const handleSortClick = (key: SortKey) => {
    if (sortKey === key) {
      setSortDir(prev => prev === 'desc' ? 'asc' : 'desc')
    } else {
      setSortKey(key)
      setSortDir('desc')
    }
  }

  const chipBase: CSSProperties = { borderRadius: 20, padding: '5px 14px', fontSize: 11, fontWeight: 600, cursor: 'pointer', border: '1px solid #e2e8f0', background: '#f8fafc', color: '#64748b', transition: 'all 0.15s' }
  const chipActiveAll: CSSProperties = { ...chipBase, border: 'none', background: 'linear-gradient(135deg, #6366f1, #8b5cf6)', color: '#fff', boxShadow: '0 1px 3px rgba(99,102,241,0.3)' }

  return (
    <ModelRequired>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 28 }}>

      {/* Hero header */}
      <div>
        <h1 style={{
          fontSize: 32, fontWeight: 800, letterSpacing: '-0.03em', margin: 0, lineHeight: 1.2,
          background: 'linear-gradient(135deg, #1e293b 0%, #475569 100%)',
          WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent', backgroundClip: 'text',
        }}>
          Structural Cognitive Load
        </h1>
        <p style={{ fontSize: 15, color: '#64748b', marginTop: 8, marginBottom: 0, lineHeight: 1.5 }}>
          Team Topologies assessment across 4 structural dimensions — the worst dimension sets the overall load level
        </p>
        <button type="button" onClick={() => setShowThresholds(prev => !prev)}
          style={{ fontSize: 12, color: '#6366f1', background: 'none', border: 'none', cursor: 'pointer', padding: '4px 0', marginTop: 6, fontWeight: 500 }}>
          {showThresholds ? '\u24D8 Dimension thresholds \u25B2' : '\u24D8 Dimension thresholds \u25BC'}
        </button>
        {showThresholds && (
          <div style={{
            borderRadius: 12, padding: '14px 20px', marginTop: 8,
            background: 'linear-gradient(135deg, #ffffff 0%, #f8fafc 100%)',
            border: '1px solid #e2e8f0',
          }}>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '8px 32px' }}>
              {DIMENSIONS.map(d => (
                <div key={d.key} style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                  <div style={{ width: 28, height: 28, borderRadius: 8, background: '#f1f5f9', border: '1px solid #e2e8f0', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                    <d.icon size={13} style={{ color: '#64748b' }} />
                  </div>
                  <div style={{ flex: 1 }}>
                    <div style={{ fontSize: 12, color: '#334155', fontWeight: 600 }}>{d.label}</div>
                    <div style={{ fontSize: 10, color: '#94a3b8', fontFamily: 'ui-monospace, monospace' }}>{d.thresholds}</div>
                  </div>
                </div>
              ))}
            </div>
            <div style={{ fontSize: 11, color: '#94a3b8', marginTop: 12, paddingTop: 12, borderTop: '1px solid #e2e8f0', lineHeight: 1.6 }}>
              Interaction weights: collaboration = 3 · facilitating = 2 · x-as-a-service = 1 · Overall = worst of the four dimensions
              <br />
              <em>Structural proxies only — qualitative team assessment remains essential.</em>
            </div>
          </div>
        )}
      </div>

      {/* Summary stat cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 14 }}>
        {[
          { label: 'Total Teams', value: counts.total, gradient: 'linear-gradient(135deg, #ede9fe 0%, #e0e7ff 100%)', border: '#c7d2fe', icon: Users, iconGrad: 'linear-gradient(135deg, #8b5cf6, #6366f1)', iconColor: '#fff' },
          { label: 'High Load', value: counts.high, gradient: 'linear-gradient(135deg, #fecaca 0%, #fee2e2 100%)', border: '#fca5a5', icon: AlertTriangle, iconGrad: 'linear-gradient(135deg, #ef4444, #dc2626)', iconColor: '#fff' },
          { label: 'Medium Load', value: counts.medium, gradient: 'linear-gradient(135deg, #fde68a 0%, #fef3c7 100%)', border: '#fcd34d', icon: Gauge, iconGrad: 'linear-gradient(135deg, #f59e0b, #d97706)', iconColor: '#fff' },
          { label: 'Low Load', value: counts.low, gradient: 'linear-gradient(135deg, #bbf7d0 0%, #dcfce7 100%)', border: '#86efac', icon: Shield, iconGrad: 'linear-gradient(135deg, #22c55e, #16a34a)', iconColor: '#fff' },
        ].map(card => (
          <div key={card.label} style={{
            borderRadius: 20, padding: '20px 22px',
            background: card.gradient, border: `1px solid ${card.border}`,
            boxShadow: '0 1px 3px rgba(0,0,0,0.04)',
            display: 'flex', alignItems: 'center', gap: 16,
          }}>
            <div style={{
              width: 48, height: 48, borderRadius: 14, background: card.iconGrad,
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              boxShadow: `0 3px 10px ${card.border}80`,
            }}>
              <card.icon size={22} color={card.iconColor} strokeWidth={2.25} />
            </div>
            <div>
              <div style={{ fontSize: 28, fontWeight: 800, color: '#1e293b', lineHeight: 1, fontFamily: 'ui-monospace, monospace' }}>{card.value}</div>
              <div style={{ fontSize: 11, fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.05em', color: '#64748b', marginTop: 4 }}>{card.label}</div>
            </div>
          </div>
        ))}
      </div>

      {/* Distribution bar */}
      {counts.total > 0 && (
        <div style={{
          borderRadius: 16, padding: '16px 20px',
          background: 'linear-gradient(135deg, #ffffff 0%, #f8fafc 100%)',
          border: '1px solid #e2e8f0', boxShadow: '0 1px 3px rgba(0,0,0,0.04)',
        }}>
          <div style={{ fontSize: 11, fontWeight: 700, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: 10 }}>Load Distribution</div>
          <div style={{ display: 'flex', height: 14, borderRadius: 7, overflow: 'hidden', background: '#f1f5f9' }}>
            {counts.high > 0 && (
              <div style={{ width: `${(counts.high / counts.total) * 100}%`, background: LEVEL.high.gradient, transition: 'width 0.6s ease', cursor: 'default' }} title={`High load (${teamsByLevel.high.length}): ${teamsByLevel.high.join(', ')}`} />
            )}
            {counts.medium > 0 && (
              <div style={{ width: `${(counts.medium / counts.total) * 100}%`, background: LEVEL.medium.gradient, transition: 'width 0.6s ease', cursor: 'default' }} title={`Medium load (${teamsByLevel.medium.length}): ${teamsByLevel.medium.join(', ')}`} />
            )}
            {counts.low > 0 && (
              <div style={{ width: `${(counts.low / counts.total) * 100}%`, background: LEVEL.low.gradient, transition: 'width 0.6s ease', cursor: 'default' }} title={`Low load (${teamsByLevel.low.length}): ${teamsByLevel.low.join(', ')}`} />
            )}
          </div>
          <div style={{ display: 'flex', gap: 20, marginTop: 8 }}>
            {[
              { label: 'High', color: '#ef4444', count: counts.high },
              { label: 'Medium', color: '#f59e0b', count: counts.medium },
              { label: 'Low', color: '#22c55e', count: counts.low },
            ].map(item => (
              <div key={item.label} style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                <span style={{ width: 8, height: 8, borderRadius: '50%', background: item.color }} />
                <span style={{ fontSize: 11, color: '#64748b', fontWeight: 500 }}>
                  {item.label}: <span style={{ fontWeight: 700, color: '#1e293b' }}>{item.count}</span> ({Math.round((item.count / counts.total) * 100)}%)
                </span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Missing size banner */}
      {counts.missingSize > 0 && (
        <div style={{
          borderRadius: 16, padding: '14px 20px',
          background: 'linear-gradient(135deg, #fef3c7 0%, #fffbeb 100%)',
          border: '1px solid #fde68a',
          display: 'flex', alignItems: 'center', gap: 10,
        }}>
          <AlertTriangle size={16} style={{ color: '#d97706', flexShrink: 0 }} />
          <span style={{ fontSize: 13, color: '#78350f' }}>
            <strong>{counts.missingSize} team{counts.missingSize > 1 ? 's' : ''}</strong> missing{' '}
            <code style={{ background: 'rgba(255,255,255,0.6)', padding: '1px 5px', borderRadius: 4, fontFamily: 'ui-monospace, monospace', border: '1px solid #fcd34d', fontSize: 12 }}>size:</code>{' '}
            — Service Load uses a default of 5 people. Results may be inaccurate.
          </span>
        </div>
      )}

      {/* Controls */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 16, flexWrap: 'wrap' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          <span style={{ fontSize: 12, color: '#9ca3af', fontWeight: 500 }}>Sort:</span>
          <div style={{ display: 'flex', borderRadius: 12, padding: 4, background: '#f1f5f9', border: '1px solid #e2e8f0', gap: 2 }}>
            <button type="button" style={sortKey === 'level' ? pillActiveSorted : pillInactive} onClick={() => handleSortClick('level')}>Severity{sortKey === 'level' ? (sortDir === 'desc' ? ' \u2193' : ' \u2191') : ''}</button>
            <button type="button" style={sortKey === 'name' ? pillActiveSorted : pillInactive} onClick={() => handleSortClick('name')}>Name{sortKey === 'name' ? (sortDir === 'desc' ? ' \u2193' : ' \u2191') : ''}</button>
            <button type="button" style={sortKey === 'type' ? pillActiveSorted : pillInactive} onClick={() => handleSortClick('type')}>Type{sortKey === 'type' ? (sortDir === 'desc' ? ' \u2193' : ' \u2191') : ''}</button>
          </div>
        </div>
        <div style={{ width: 1, height: 20, background: '#e5e7eb' }} />
        <div style={{ display: 'flex', alignItems: 'center', gap: 6, flexWrap: 'wrap' }}>
          <span style={{ fontSize: 12, color: '#9ca3af', fontWeight: 500 }}>Filter:</span>
          <button type="button" style={filterType === '' ? chipActiveAll : chipBase} onClick={() => setFilterType('')}>All teams</button>
          {TEAM_TYPES.map(t => {
            const active = filterType === t
            const b = TEAM_TYPE_BADGE[t]
            return (
              <button
                key={t} type="button"
                style={active
                  ? { ...chipBase, border: `1px solid ${b.text}40`, background: b.bg, color: b.text, boxShadow: `0 1px 3px ${b.text}15` }
                  : chipBase
                }
                onClick={() => setFilterType(filterType === t ? '' : t)}
              >{t.split('-').map(w => w.charAt(0).toUpperCase() + w.slice(1)).join(' ')}</button>
            )
          })}
        </div>
        <span style={{ fontSize: 12, color: '#94a3b8', marginLeft: 'auto' }}>
          {sorted.length} of {counts.total} teams
        </span>
      </div>

      {/* Team cards grid */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(340px, 1fr))', gap: 14 }}>
        {sorted.map(tl => (
          <TeamCard
            key={tl.team.name}
            tl={tl}
            insight={insights[`team-load:${slug(tl.team.name)}`]}
            isExpanded={expandedTeam === tl.team.name}
            onToggle={() => setExpandedTeam(prev => prev === tl.team.name ? null : tl.team.name)}
          />
        ))}
        {sorted.length === 0 && (
          <div style={{
            gridColumn: '1 / -1',
            textAlign: 'center', padding: '48px 0', fontSize: 14, color: '#94a3b8',
            borderRadius: 16, background: '#f8fafc', border: '1px solid #e2e8f0',
          }}>
            No teams match the current filter.
          </div>
        )}
      </div>

    </div>
    </ModelRequired>
  )
}
