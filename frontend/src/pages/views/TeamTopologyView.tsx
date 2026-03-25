import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '@/lib/api'
import { useRequireModel } from '@/lib/model-context'
import { ModelRequired } from '@/components/ui/ModelRequired'
import { useSearch, matchesQuery } from '@/lib/search-context'
import { usePageInsights } from '@/hooks/usePageInsights'
import { Users, ArrowRight, Zap, Layers, X, Info, Lightbulb, Sparkles, ChevronRight } from 'lucide-react'
import { QuickAction } from '@/components/changeset/QuickAction'

const slug = (s: string) => s.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '')

interface TeamTopologyInteraction {
  source_id: string
  target_id: string
  mode: string
  via: string
  description: string
}

interface TeamTopologyTeam {
  id: string
  label: string
  description: string
  type: string
  is_overloaded: boolean
  capability_count: number
  service_count: number
  interactions: TeamTopologyInteraction[]
  anti_patterns?: Array<{ code: string; message: string; severity: string }>
  services: string[]
  capabilities: string[]
}

interface TeamTopologyViewResponse {
  view_type: string
  teams: TeamTopologyTeam[]
  interactions: TeamTopologyInteraction[]
}

// ── Design tokens ─────────────────────────────────────────────────────────────

const TEAM_TYPES: Record<string, {
  label: string; accent: string; bg: string; border: string
  gradientFrom: string; gradientTo: string; zoneBg: string
}> = {
  'platform': {
    label: 'Platform', accent: '#7c3aed', bg: '#faf5ff', border: '#ddd6fe',
    gradientFrom: '#7c3aed', gradientTo: '#a78bfa', zoneBg: 'rgba(124,58,237,0.03)',
  },
  'stream-aligned': {
    label: 'Stream-aligned', accent: '#1d4ed8', bg: '#eff6ff', border: '#bfdbfe',
    gradientFrom: '#1d4ed8', gradientTo: '#60a5fa', zoneBg: 'rgba(29,78,216,0.03)',
  },
  'complicated-subsystem': {
    label: 'Complicated Subsystem', accent: '#b45309', bg: '#fffbeb', border: '#fde68a',
    gradientFrom: '#b45309', gradientTo: '#fbbf24', zoneBg: 'rgba(180,83,9,0.03)',
  },
  'enabling': {
    label: 'Enabling', accent: '#15803d', bg: '#f0fdf4', border: '#bbf7d0',
    gradientFrom: '#15803d', gradientTo: '#4ade80', zoneBg: 'rgba(21,128,61,0.03)',
  },
}

const TEAM_TYPE_DESCRIPTIONS: Record<string, string> = {
  'stream-aligned': 'Aligned to a flow of work from a business domain segment',
  'platform': 'Provides internal services to reduce cognitive load of other teams',
  'enabling': 'Helps other teams adopt new practices or technologies',
  'complicated-subsystem': 'Owns a subsystem requiring deep specialist knowledge',
}

const INTERACTION_STYLE: Record<string, { label: string; bg: string; text: string; border: string; color: string }> = {
  'collaboration':  { label: 'Collaboration',  bg: '#dbeafe', text: '#1e40af', border: '#bfdbfe', color: '#1d4ed8' },
  'x-as-a-service': { label: 'X-as-a-Service', bg: '#ede9fe', text: '#5b21b6', border: '#ddd6fe', color: '#7c3aed' },
  'facilitating':   { label: 'Facilitating',   bg: '#d1fae5', text: '#065f46', border: '#a7f3d0', color: '#15803d' },
}

function getType(t: string) { return TEAM_TYPES[t] ?? TEAM_TYPES['stream-aligned'] }
function getIx(m: string)   { return INTERACTION_STYLE[m] ?? { label: m, bg: '#f1f5f9', text: '#475569', border: '#e2e8f0', color: '#64748b' } }

// ── Layout constants ───────────────────────────────────────────────────────────

const NODE_W = 220
const NODE_H = 108
const COL_PAD = 24
const COL_GAP = 180  // gap between columns
const ROW_GAP = 18

// 3 columns: Platform | Stream-aligned | Complicated/Enabling
const COLUMNS = [
  { types: ['platform'],              label: 'Platform',               x: COL_PAD },
  { types: ['stream-aligned'],        label: 'Stream-aligned',         x: COL_PAD + NODE_W + COL_GAP },
  { types: ['complicated-subsystem', 'enabling'], label: 'Subsystem / Enabling', x: COL_PAD + (NODE_W + COL_GAP) * 2 },
]

const HEADER_H = 56  // space for column header + gap

// ── Shared styles ──────────────────────────────────────────────────────────────

const gradientTitle: React.CSSProperties = {
  fontSize: 30, fontWeight: 800, letterSpacing: '-0.025em',
  background: 'linear-gradient(135deg, #1e293b 0%, #475569 100%)',
  WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent', backgroundClip: 'text',
}

const pill = (extra?: React.CSSProperties): React.CSSProperties => ({
  display: 'inline-flex', alignItems: 'center', borderRadius: 8,
  padding: '3px 10px', fontSize: 11, fontWeight: 600, ...extra,
})

// ── Stat card ─────────────────────────────────────────────────────────────────

function StatCard({ value, label, sub, icon: Icon, gradient, iconTint }: {
  value: number | string; label: string; sub?: string
  icon: React.ElementType; gradient: string; iconTint: string
}) {
  return (
    <div style={{
      flex: '1 1 160px', borderRadius: 20, padding: '18px 20px',
      background: gradient, border: '1px solid #e2e8f0',
      boxShadow: '0 4px 14px rgba(15,23,42,0.06)',
      transition: 'transform 0.15s',
    }}
      onMouseEnter={e => (e.currentTarget.style.transform = 'translateY(-1px)')}
      onMouseLeave={e => (e.currentTarget.style.transform = '')}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 12 }}>
        <div style={{ borderRadius: 12, padding: 8, background: 'rgba(255,255,255,0.75)', border: '1px solid rgba(226,232,240,0.9)', boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
          <Icon size={18} style={{ color: iconTint }} strokeWidth={2.25} />
        </div>
      </div>
      <div style={{ fontSize: 26, fontWeight: 800, color: '#1e293b', lineHeight: 1 }}>{value}</div>
      <div style={{ fontSize: 11, fontWeight: 600, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em', marginTop: 4 }}>{label}</div>
      {sub && <div style={{ fontSize: 11, color: '#94a3b8', marginTop: 2 }}>{sub}</div>}
    </div>
  )
}

// ── Graph view ────────────────────────────────────────────────────────────────

function GraphView({ teams, interactions, insights, filterOverloaded }: {
  teams: TeamTopologyTeam[]
  interactions: TeamTopologyInteraction[]
  insights: Record<string, { explanation: string; suggestion: string }>
  filterOverloaded: boolean
}) {
  const [focusedTeam, setFocusedTeam] = useState<string | null>(null)
  const [detailTeam, setDetailTeam] = useState<TeamTopologyTeam | null>(null)
  const [modeFilters, setModeFilters] = useState({ 'x-as-a-service': true, collaboration: true, facilitating: true })

  const teamById = useMemo(() => new Map(teams.map(t => [t.id, t])), [teams])

  // Assign each team to a column and row
  const nodePositions = useMemo(() => {
    const pos: Record<string, { x: number; y: number }> = {}
    const rowIdx: Record<string, number> = {}
    for (const team of teams) {
      const col = COLUMNS.find(c => c.types.includes(team.type)) ?? COLUMNS[1]
      const row = rowIdx[col.label] ?? 0
      rowIdx[col.label] = row + 1
      pos[team.id] = { x: col.x, y: HEADER_H + row * (NODE_H + ROW_GAP) }
    }
    return pos
  }, [teams])

  const fanIn = useMemo(() => {
    const m = new Map<string, number>()
    for (const ix of interactions) m.set(ix.target_id, (m.get(ix.target_id) ?? 0) + 1)
    return m
  }, [interactions])

  const connectedToFocused = useMemo(() => {
    if (!focusedTeam) return new Set<string>()
    const s = new Set<string>([focusedTeam])
    for (const ix of interactions) {
      if (ix.source_id === focusedTeam) s.add(ix.target_id)
      if (ix.target_id === focusedTeam) s.add(ix.source_id)
    }
    return s
  }, [focusedTeam, interactions])

  const visibleInteractions = useMemo(() =>
    interactions.filter(ix => modeFilters[ix.mode as keyof typeof modeFilters]),
    [interactions, modeFilters]
  )

  const svgH = useMemo(() => {
    let h = HEADER_H + 40
    for (const p of Object.values(nodePositions)) h = Math.max(h, p.y + NODE_H + 32)
    return h
  }, [nodePositions])

  const svgW = useMemo(() => {
    let w = NODE_W + COL_PAD * 2
    for (const p of Object.values(nodePositions)) w = Math.max(w, p.x + NODE_W + COL_PAD)
    return w
  }, [nodePositions])

  function handleNode(team: TeamTopologyTeam) {
    const wasFocused = focusedTeam === team.id
    setFocusedTeam(p => p === team.id ? null : team.id)
    setDetailTeam(wasFocused ? null : team)
  }

  return (
    <>
      {/* Interaction mode filters */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap', marginBottom: 16 }}>
        <span style={{ fontSize: 11, fontWeight: 600, color: '#94a3b8', textTransform: 'uppercase', letterSpacing: '0.05em', marginRight: 4 }}>Show</span>
        {Object.entries(INTERACTION_STYLE).map(([mode, s]) => {
          const active = modeFilters[mode as keyof typeof modeFilters]
          return (
            <button key={mode} type="button"
              onClick={() => setModeFilters(p => ({ ...p, [mode]: !p[mode as keyof typeof p] }))}
              style={{
                fontSize: 11, fontWeight: 600, borderRadius: 20, padding: '5px 14px', cursor: 'pointer',
                background: active ? s.bg : '#f1f5f9', color: active ? s.text : '#94a3b8',
                border: `1px solid ${active ? s.border : '#e2e8f0'}`,
                transition: 'all 0.15s',
              }}
            >
              {s.label}
            </button>
          )
        })}
        <span style={{ fontSize: 11, color: '#94a3b8', marginLeft: 4 }}>
          {visibleInteractions.length} interaction{visibleInteractions.length !== 1 ? 's' : ''} visible
        </span>
      </div>

      {/* Graph canvas */}
      <div style={{
        position: 'relative', overflowX: 'auto', overflowY: 'visible',
        borderRadius: 20, border: '1px solid #e2e8f0',
        background: 'linear-gradient(135deg, #fafbff 0%, #f8fafc 100%)',
        boxShadow: '0 1px 3px rgba(0,0,0,0.04)',
      }}>
        {/* Column zone backgrounds */}
        {COLUMNS.map(col => (
          <div key={col.label} style={{
            position: 'absolute', top: 0, bottom: 0,
            left: col.x - 12, width: NODE_W + 24,
            background: TEAM_TYPES[col.types[0]]?.zoneBg ?? 'transparent',
            borderRight: '1px dashed #e2e8f0',
            pointerEvents: 'none',
          }} />
        ))}

        {/* SVG edges */}
        <svg style={{ position: 'absolute', top: 0, left: 0, width: svgW, height: svgH, pointerEvents: 'none', overflow: 'visible' }}>
          <defs>
            {Object.entries(INTERACTION_STYLE).map(([mode, s]) => (
              <marker key={mode} id={`arrow-${mode}`} markerWidth="7" markerHeight="7" refX="5" refY="3.5" orient="auto">
                <path d="M0,0 L0,7 L7,3.5 z" fill={s.color} />
              </marker>
            ))}
          </defs>
          {visibleInteractions.map((ix, i) => {
            const sp = nodePositions[ix.source_id]
            const tp = nodePositions[ix.target_id]
            if (!sp || !tp) return null
            const s = getIx(ix.mode)
            const isFocus = focusedTeam
              ? ix.source_id === focusedTeam || ix.target_id === focusedTeam
              : true
            const opacity = focusedTeam ? (isFocus ? 1 : 0.08) : 0.75
            const strokeW = focusedTeam && isFocus ? 2.5 : 1.5

            // Route edges based on which column is left of the other,
            // so connections that cross left emit from the left side of
            // the source and enter the right side of the target —
            // preventing tangled U-turn curves.
            const goingLeft = tp.x < sp.x
            const sx = goingLeft ? sp.x        : sp.x + NODE_W
            const sy = sp.y + NODE_H / 2
            const tx = goingLeft ? tp.x + NODE_W : tp.x
            const ty = tp.y + NODE_H / 2
            const dx = tx - sx
            const cp1x = sx + dx * 0.45, cp2x = tx - dx * 0.45

            return (
              <path key={i}
                d={`M${sx},${sy} C${cp1x},${sy} ${cp2x},${ty} ${tx},${ty}`}
                stroke={s.color} strokeWidth={strokeW} fill="none"
                opacity={opacity} strokeDasharray={ix.mode === 'facilitating' ? '5,3' : undefined}
                markerEnd={`url(#arrow-${ix.mode})`}
              >
                <title>{s.label}{ix.via ? ` via ${ix.via}` : ''}{ix.description ? `\n${ix.description}` : ''}</title>
              </path>
            )
          })}
        </svg>

        {/* Column headers */}
        {COLUMNS.map(col => {
          const cfg = TEAM_TYPES[col.types[0]]
          return (
            <div key={col.label} style={{
              position: 'absolute', left: col.x, top: 14, width: NODE_W,
              textAlign: 'center', fontSize: 10, fontWeight: 700, letterSpacing: '0.06em',
              textTransform: 'uppercase', color: cfg?.accent ?? '#64748b',
              background: cfg?.bg ?? '#f8fafc', border: `1px solid ${cfg?.border ?? '#e2e8f0'}`,
              borderRadius: 20, padding: '5px 0', pointerEvents: 'none',
            }}>
              {col.label}
            </div>
          )
        })}

        {/* Team nodes */}
        {teams.map(team => {
          const pos = nodePositions[team.id]
          if (!pos) return null
          const cfg = getType(team.type)
          const fi = fanIn.get(team.id) ?? 0
          const isConn = focusedTeam ? connectedToFocused.has(team.id) : true
          const isFocused = focusedTeam === team.id
          const overloaded = team.is_overloaded

          return (
            <div key={team.id} onClick={() => handleNode(team)}
              style={{
                position: 'absolute', left: pos.x, top: pos.y, width: NODE_W, height: NODE_H,
                borderRadius: 16, cursor: 'pointer', display: 'flex', overflow: 'hidden',
                opacity: focusedTeam ? (isConn ? 1 : 0.15) : filterOverloaded ? (overloaded ? 1 : 0.2) : 1,
                transition: 'opacity 0.2s, box-shadow 0.2s, transform 0.2s',
                border: `1.5px solid ${overloaded ? '#fca5a5' : isFocused ? cfg.accent : '#e2e8f0'}`,
                boxShadow: isFocused
                  ? `0 0 0 3px ${cfg.accent}33, 0 12px 32px rgba(15,23,42,0.14)`
                  : '0 2px 8px rgba(15,23,42,0.06)',
                background: 'linear-gradient(135deg, #ffffff 0%, #fafbff 100%)',
              }}
              onMouseEnter={e => {
                if (focusedTeam && !isConn) return
                e.currentTarget.style.transform = 'translateY(-2px)'
                e.currentTarget.style.boxShadow = isFocused
                  ? `0 0 0 3px ${cfg.accent}33, 0 16px 40px rgba(15,23,42,0.16)`
                  : '0 8px 28px rgba(15,23,42,0.12)'
                const btn = e.currentTarget.querySelector('.qa-node') as HTMLElement; if (btn) btn.style.opacity = '1'
              }}
              onMouseLeave={e => {
                e.currentTarget.style.transform = ''
                e.currentTarget.style.boxShadow = isFocused
                  ? `0 0 0 3px ${cfg.accent}33, 0 12px 32px rgba(15,23,42,0.14)`
                  : '0 2px 8px rgba(15,23,42,0.06)'
                const btn = e.currentTarget.querySelector('.qa-node') as HTMLElement; if (btn) btn.style.opacity = '0.35'
              }}
            >
              {/* Color strip */}
              <div style={{ width: 4, flexShrink: 0, background: `linear-gradient(180deg, ${cfg.gradientFrom} 0%, ${cfg.gradientTo} 100%)` }} />

              {/* Content */}
              <div style={{ flex: 1, padding: '10px 12px', minWidth: 0, display: 'flex', flexDirection: 'column', justifyContent: 'space-between' }}>
                {/* Top row: type badge + fan-in */}
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 6 }}>
                  <span style={{ ...pill(), background: cfg.bg, color: cfg.accent, border: `1px solid ${cfg.border}`, textTransform: 'uppercase', letterSpacing: '0.04em', fontSize: 9 }}
                    title={TEAM_TYPE_DESCRIPTIONS[team.type]}>
                    {cfg.label}
                  </span>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                    {overloaded && (
                      <span style={{ ...pill({ fontSize: 9, padding: '2px 7px', cursor: 'help' }), background: '#fff7ed', color: '#c2410c', border: '1px solid #fed7aa' }}
                        title="This team owns too many services or capabilities, increasing cognitive load"
                        aria-label="Warning: team is overloaded">
                        overloaded
                      </span>
                    )}
                    {fi > 0 && (
                      <span style={{ fontSize: 10, fontWeight: 600, color: '#94a3b8' }}>← {fi}</span>
                    )}
                  </div>
                </div>

                {/* Team name */}
                <div style={{
                  fontSize: 13, fontWeight: 700, color: '#0f172a',
                  overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
                  letterSpacing: '-0.02em', marginTop: 6,
                  display: 'flex', alignItems: 'center', gap: 4,
                }}>
                  <span style={{ overflow: 'hidden', textOverflow: 'ellipsis' }}>{team.label}</span>
                  <span className="qa-node" style={{ opacity: 0.35, transition: 'opacity 0.15s' }}>
                    <QuickAction size={11} options={[
                      { label: 'Change team type', action: { type: 'update_team_type', team_name: team.label } },
                      { label: 'Update team size', action: { type: 'update_team_size', team_name: team.label } },
                      { label: 'Add service to team', action: { type: 'add_service', owner_team_name: team.label } },
                    ]} />
                  </span>
                </div>

                {/* Metrics */}
                <div style={{ display: 'flex', gap: 10, marginTop: 6 }}>
                  <span style={{ fontSize: 11, color: '#64748b', fontWeight: 500 }}>
                    <span style={{ fontWeight: 700, color: '#334155' }}>{team.capability_count}</span> caps
                  </span>
                  <span style={{ fontSize: 11, color: '#64748b', fontWeight: 500 }}>
                    <span style={{ fontWeight: 700, color: '#334155' }}>{team.service_count}</span> svcs
                  </span>
                  <span style={{ fontSize: 11, color: '#64748b', fontWeight: 500 }}>
                    <span style={{ fontWeight: 700, color: '#334155' }}>{team.interactions?.length ?? 0}</span> ixns
                  </span>
                </div>
              </div>

              {/* Click indicator */}
              <div style={{ display: 'flex', alignItems: 'center', paddingRight: 8, color: '#cbd5e1' }}>
                <ChevronRight size={14} />
              </div>
            </div>
          )
        })}

        {/* Invisible spacer to set container height */}
        <div style={{ height: svgH, width: svgW, pointerEvents: 'none' }} />
      </div>

      {/* Detail panel */}
      {detailTeam && (
        <DetailPanel
          team={detailTeam}
          teamById={teamById}
          allInteractions={interactions}
          insights={insights}
          onClose={() => { setDetailTeam(null); setFocusedTeam(null) }}
        />
      )}
    </>
  )
}

// ── Table view ────────────────────────────────────────────────────────────────

function TableView({ teams, interactions, insights }: {
  teams: TeamTopologyTeam[]
  interactions: TeamTopologyInteraction[]
  insights: Record<string, { explanation: string; suggestion: string }>
}) {
  const [detailTeam, setDetailTeam] = useState<TeamTopologyTeam | null>(null)
  const teamById = useMemo(() => new Map(teams.map(t => [t.id, t])), [teams])
  const navigate = useNavigate()

  const typeOrder = ['platform', 'stream-aligned', 'complicated-subsystem', 'enabling']

  return (
    <>
      <div className="space-y-10">
        {typeOrder.map(typeKey => {
          const cfg = TEAM_TYPES[typeKey]
          const typeTeams = teams.filter(t => t.type === typeKey)
          if (typeTeams.length === 0 || !cfg) return null
          return (
            <div key={typeKey}>
              {/* Section divider */}
              <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 16 }}>
                <div style={{ flex: 1, height: 1, background: `linear-gradient(90deg, transparent, ${cfg.border}, transparent)` }} />
                <span style={{
                  fontSize: 11, fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.06em',
                  padding: '6px 16px', borderRadius: 20, whiteSpace: 'nowrap',
                  background: cfg.bg, color: cfg.accent, border: `1px solid ${cfg.border}`,
                  boxShadow: '0 1px 3px rgba(0,0,0,0.04)',
                }}>
                  {cfg.label} · {typeTeams.length} team{typeTeams.length !== 1 ? 's' : ''}
                </span>
                <div style={{ flex: 1, height: 1, background: `linear-gradient(90deg, ${cfg.border}, transparent)` }} />
              </div>

              <div style={{ display: 'grid', gap: 16, gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))' }}>
                {typeTeams.map(team => {
                  const myIx = team.interactions ?? []
                  return (
                    <div key={team.id} onClick={() => setDetailTeam(team)}
                      style={{
                        borderRadius: 18, background: 'linear-gradient(135deg, #ffffff 0%, #f8fafc 100%)',
                        border: `1.5px solid ${team.is_overloaded ? '#fca5a5' : '#e2e8f0'}`,
                        boxShadow: '0 1px 4px rgba(15,23,42,0.06)', cursor: 'pointer', overflow: 'hidden',
                        transition: 'transform 0.15s, box-shadow 0.15s',
                      }}
                      onMouseEnter={e => { e.currentTarget.style.transform = 'translateY(-2px)'; e.currentTarget.style.boxShadow = '0 12px 32px rgba(15,23,42,0.1)' }}
                      onMouseLeave={e => { e.currentTarget.style.transform = ''; e.currentTarget.style.boxShadow = '0 1px 4px rgba(15,23,42,0.06)' }}
                    >
                      <div style={{ height: 4, background: `linear-gradient(90deg, ${cfg.gradientFrom} 0%, ${cfg.gradientTo} 100%)` }} />
                      <div style={{ padding: '14px 16px' }}>
                        {/* Header */}
                        <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 8, marginBottom: 8 }}>
                          <span style={{ fontSize: 14, fontWeight: 700, color: '#0f172a', lineHeight: 1.3 }}>{team.label}</span>
                          {team.is_overloaded && (
                            <span style={{ ...pill({ padding: '3px 9px', fontSize: 10, cursor: 'help' }), background: '#fff7ed', color: '#c2410c', border: '1px solid #fed7aa', flexShrink: 0 }}
                              title="This team owns too many services or capabilities, increasing cognitive load"
                              aria-label="Warning: team is overloaded">
                              overloaded
                            </span>
                          )}
                        </div>

                        {team.description && (
                          <p style={{ fontSize: 12, color: '#64748b', marginBottom: 10, lineHeight: 1.5, display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical', overflow: 'hidden' }}>
                            {team.description}
                          </p>
                        )}

                        {/* Metrics strip */}
                        <div style={{ display: 'flex', gap: 16, paddingBottom: 10, marginBottom: 10, borderBottom: '1px solid #f1f5f9' }}>
                          {[
                            { v: team.capability_count, l: 'caps' },
                            { v: team.service_count, l: 'svcs' },
                            { v: myIx.length, l: 'ixns' },
                          ].map(({ v, l }) => (
                            <div key={l}>
                              <span style={{ fontSize: 18, fontWeight: 800, color: '#0f172a' }}>{v}</span>
                              <span style={{ fontSize: 11, color: '#94a3b8', marginLeft: 3 }}>{l}</span>
                            </div>
                          ))}
                        </div>

                        {/* Interactions */}
                        {myIx.length > 0 ? (
                          <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
                            {myIx.slice(0, 4).map((ix, i) => {
                              const isSource = ix.source_id === team.id
                              const other = teamById.get(isSource ? ix.target_id : ix.source_id)
                              const s = getIx(ix.mode)
                              return (
                                <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 12 }}>
                                  <span style={{ ...pill({ fontSize: 10, padding: '2px 8px' }), background: s.bg, color: s.text, border: `1px solid ${s.border}`, flexShrink: 0 }}>
                                    {s.label}
                                  </span>
                                  <span style={{ color: '#94a3b8', flexShrink: 0 }}>{isSource ? '→' : '←'}</span>
                                  <span style={{ color: '#334155', fontWeight: 500, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                                    {other?.label ?? (isSource ? ix.target_id : ix.source_id)}
                                  </span>
                                </div>
                              )
                            })}
                            {myIx.length > 4 && (
                              <span style={{ fontSize: 11, color: '#94a3b8' }}>+{myIx.length - 4} more</span>
                            )}
                          </div>
                        ) : (
                          <span style={{ fontSize: 12, color: '#94a3b8', fontStyle: 'italic' }}>No interactions modelled</span>
                        )}
                      </div>
                    </div>
                  )
                })}
              </div>
            </div>
          )
        })}

        {/* All interactions table */}
        {interactions.length > 0 && (
          <div style={{ borderRadius: 20, border: '1px solid #e2e8f0', overflow: 'hidden', boxShadow: '0 1px 3px rgba(0,0,0,0.04)' }}>
            <div style={{ height: 4, background: 'linear-gradient(90deg, #6366f1 0%, #8b5cf6 50%, #ec4899 100%)' }} />
            <div style={{ padding: '12px 20px', borderBottom: '1px solid #e2e8f0', background: 'linear-gradient(135deg, #f8fafc 0%, #ffffff 100%)' }}>
              <span style={{ fontSize: 11, fontWeight: 700, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                All Interactions ({interactions.length})
              </span>
            </div>
            <div style={{ overflowX: 'auto' }}>
              <table style={{ width: '100%', background: '#ffffff', borderCollapse: 'collapse' }}>
                <thead>
                  <tr style={{ background: '#f8fafc', borderBottom: '1px solid #e2e8f0' }}>
                    {['From', 'Mode', 'To', 'Via', 'Description'].map(h => (
                      <th key={h} style={{ textAlign: 'left', padding: '10px 20px', fontSize: 11, fontWeight: 600, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                        {h}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {interactions.map((ix, idx) => {
                    const src = teamById.get(ix.source_id)
                    const tgt = teamById.get(ix.target_id)
                    const s = getIx(ix.mode)
                    const isSelf = ix.source_id === ix.target_id
                    return (
                      <tr key={idx} style={{ borderBottom: idx < interactions.length - 1 ? '1px solid #f1f5f9' : 'none', background: idx % 2 === 0 ? '#ffffff' : '#fafafa' }}>
                        <td style={{ padding: '10px 20px', fontSize: 12, fontWeight: 600, color: '#334155' }}>
                          {src?.label ?? ix.source_id}
                          {src && (
                            <span style={{ fontSize: 11, padding: '1px 5px', borderRadius: 9999,
                              background: (getType(src.type).accent) + '20',
                              color: getType(src.type).accent,
                              marginLeft: 4 }}
                              title={TEAM_TYPE_DESCRIPTIONS[src.type]}>
                              {getType(src.type).label}
                            </span>
                          )}
                        </td>
                        <td style={{ padding: '10px 20px' }}>
                          <span style={{ ...pill(), background: s.bg, color: s.text, border: `1px solid ${s.border}` }}>{s.label}</span>
                        </td>
                        <td style={{ padding: '10px 20px', fontSize: 12, fontWeight: 600, color: '#334155' }}>
                          {tgt?.label ?? ix.target_id}
                          {tgt && (
                            <span style={{ fontSize: 11, padding: '1px 5px', borderRadius: 9999,
                              background: (getType(tgt.type).accent) + '20',
                              color: getType(tgt.type).accent,
                              marginLeft: 4 }}
                              title={TEAM_TYPE_DESCRIPTIONS[tgt.type]}>
                              {getType(tgt.type).label}
                            </span>
                          )}
                          {isSelf && (
                            <span style={{
                              fontSize: 10, padding: '1px 5px', borderRadius: 9999,
                              background: '#f3f4f6', color: '#6b7280', marginLeft: 6,
                            }} title="This team provides a service it also consumes internally">
                              self
                            </span>
                          )}
                        </td>
                        <td style={{ padding: '10px 20px', fontSize: 12, color: '#64748b', fontFamily: 'monospace' }}>
                          {ix.via ? (
                            <button onClick={() => navigate('/capability')}
                              style={{ color: '#3b82f6', background: 'none', border: 'none', cursor: 'pointer', fontSize: 13, textDecoration: 'underline', padding: 0, fontFamily: 'monospace' }}>
                              {ix.via}
                            </button>
                          ) : '—'}
                        </td>
                        <td style={{ padding: '10px 20px', fontSize: 12, color: '#64748b', maxWidth: 200, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', cursor: 'default' }}
                          title={ix.description || undefined}>
                          {ix.description || '—'}
                        </td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </div>

      {detailTeam && (
        <DetailPanel
          team={detailTeam}
          teamById={teamById}
          allInteractions={interactions}
          insights={insights}
          onClose={() => setDetailTeam(null)}
        />
      )}
    </>
  )
}

// ── Detail side panel ─────────────────────────────────────────────────────────

function DetailPanel({ team, teamById, allInteractions, insights, onClose }: {
  team: TeamTopologyTeam
  teamById: Map<string, TeamTopologyTeam>
  allInteractions: TeamTopologyInteraction[]
  insights: Record<string, { explanation: string; suggestion: string }>
  onClose: () => void
}) {
  const cfg = getType(team.type)
  const inbound  = allInteractions.filter(ix => ix.target_id === team.id)
  const outbound = allInteractions.filter(ix => ix.source_id === team.id)

  // Find AI insight for this team's interactions
  const aiInsight = team.interactions
    .map(ix => {
      const a = slug(team.id), b = slug(ix.source_id === team.id ? ix.target_id : ix.source_id)
      return insights[`interaction:${a}:${b}`] ?? insights[`interaction:${b}:${a}`]
    })
    .find(Boolean)

  const teamInsight = insights[`team:${slug(team.label)}`] ?? aiInsight

  return (
    <>
      <div onClick={onClose} style={{
        position: 'fixed', inset: 0, background: 'rgba(15,23,42,0.15)',
        backdropFilter: 'blur(4px)', WebkitBackdropFilter: 'blur(4px)', zIndex: 40,
      }} />
      <div style={{
        position: 'fixed', top: 0, right: 0, width: 360, height: '100%',
        background: '#ffffff', boxShadow: '-8px 0 40px rgba(15,23,42,0.1)',
        zIndex: 50, overflowY: 'auto', display: 'flex', flexDirection: 'column',
      }}>
        {/* Top accent */}
        <div style={{ height: 4, background: `linear-gradient(90deg, ${cfg.gradientFrom} 0%, ${cfg.gradientTo} 100%)`, flexShrink: 0 }} />

        <div style={{ padding: '20px 20px 32px', display: 'flex', flexDirection: 'column', gap: 20 }}>
          {/* Header */}
          <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 12 }}>
            <div style={{ minWidth: 0 }}>
              <span style={{ ...pill({ fontSize: 9, textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: 8, display: 'inline-flex' }), background: cfg.bg, color: cfg.accent, border: `1px solid ${cfg.border}` }}
                title={TEAM_TYPE_DESCRIPTIONS[team.type]}>
                {cfg.label}
              </span>
              <h2 style={{ fontSize: 18, fontWeight: 800, color: '#0f172a', marginTop: 6, letterSpacing: '-0.02em', lineHeight: 1.2 }}>
                {team.label}
              </h2>
              {team.is_overloaded && (
                <span style={{ ...pill({ marginTop: 8, fontSize: 10, cursor: 'help' }), background: '#fff7ed', color: '#c2410c', border: '1px solid #fed7aa' }}
                  title="This team owns too many services or capabilities, increasing cognitive load"
                  aria-label="Warning: team is overloaded">
                  ⚠ Overloaded
                </span>
              )}
            </div>
            <button onClick={onClose} style={{
              background: '#f1f5f9', border: 'none', borderRadius: 10, width: 32, height: 32,
              cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center',
              color: '#64748b', flexShrink: 0,
            }}>
              <X size={16} />
            </button>
          </div>

          {/* Description */}
          {team.description && (
            <p style={{ fontSize: 13, color: '#64748b', lineHeight: 1.6, margin: 0 }}>{team.description}</p>
          )}

          {/* Metric pills */}
          <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
            {[
              { v: team.capability_count, l: 'capabilities', color: '#5b21b6', bg: '#ede9fe' },
              { v: team.service_count,    l: 'services',     color: '#1d4ed8', bg: '#dbeafe' },
              { v: team.interactions?.length ?? 0, l: 'interactions', color: '#15803d', bg: '#d1fae5' },
            ].map(({ v, l, color, bg }) => (
              <div key={l} style={{ borderRadius: 12, padding: '8px 14px', background: bg, textAlign: 'center' }}>
                <div style={{ fontSize: 20, fontWeight: 800, color }}>{v}</div>
                <div style={{ fontSize: 10, fontWeight: 600, color, opacity: 0.8, textTransform: 'uppercase', letterSpacing: '0.04em' }}>{l}</div>
              </div>
            ))}
          </div>

          {/* AI insight */}
          {teamInsight && (
            <div style={{ borderRadius: 16, padding: '14px 16px', background: 'linear-gradient(135deg, #eef2ff 0%, #f0f9ff 100%)', border: '1px solid #c7d2fe' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 10 }}>
                <Sparkles size={13} style={{ color: '#4f46e5' }} />
                <span style={{ fontSize: 11, fontWeight: 700, color: '#4f46e5', textTransform: 'uppercase', letterSpacing: '0.05em' }}>AI Insight</span>
              </div>
              <div style={{ display: 'flex', gap: 8, marginBottom: 8 }}>
                <Info size={12} style={{ color: '#4338ca', flexShrink: 0, marginTop: 2 }} />
                <p style={{ fontSize: 12, color: '#312e81', lineHeight: 1.6, margin: 0 }}>{teamInsight.explanation}</p>
              </div>
              {teamInsight.suggestion && (
                <div style={{ display: 'flex', gap: 8 }}>
                  <Lightbulb size={12} style={{ color: '#4338ca', flexShrink: 0, marginTop: 2 }} />
                  <p style={{ fontSize: 12, color: '#3730a3', lineHeight: 1.6, margin: 0, fontWeight: 500 }}>{teamInsight.suggestion}</p>
                </div>
              )}
            </div>
          )}

          {/* Interactions in */}
          {inbound.length > 0 && (
            <div>
              <div style={{ fontSize: 11, fontWeight: 700, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: 8 }}>
                Inbound ({inbound.length})
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
                {inbound.map((ix, i) => {
                  const s = getIx(ix.mode)
                  const src = teamById.get(ix.source_id)
                  return (
                    <div key={i} style={{ display: 'flex', alignItems: 'flex-start', gap: 8, padding: '8px 12px', borderRadius: 12, background: '#f8fafc', border: '1px solid #f1f5f9' }}>
                      <span style={{ ...pill({ fontSize: 10, flexShrink: 0 }), background: s.bg, color: s.text, border: `1px solid ${s.border}` }}>{s.label}</span>
                      <div style={{ minWidth: 0 }}>
                        <div style={{ fontSize: 12, fontWeight: 600, color: '#334155' }}>{src?.label ?? ix.source_id}</div>
                        {ix.via && <div style={{ fontSize: 11, color: '#94a3b8' }}>via {ix.via}</div>}
                        {ix.description && <div style={{ fontSize: 11, color: '#64748b', marginTop: 2 }}>{ix.description}</div>}
                      </div>
                    </div>
                  )
                })}
              </div>
            </div>
          )}

          {/* Interactions out */}
          {outbound.length > 0 && (
            <div>
              <div style={{ fontSize: 11, fontWeight: 700, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: 8 }}>
                Outbound ({outbound.length})
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
                {outbound.map((ix, i) => {
                  const s = getIx(ix.mode)
                  const tgt = teamById.get(ix.target_id)
                  return (
                    <div key={i} style={{ display: 'flex', alignItems: 'flex-start', gap: 8, padding: '8px 12px', borderRadius: 12, background: '#f8fafc', border: '1px solid #f1f5f9' }}>
                      <span style={{ ...pill({ fontSize: 10, flexShrink: 0 }), background: s.bg, color: s.text, border: `1px solid ${s.border}` }}>{s.label}</span>
                      <div style={{ minWidth: 0 }}>
                        <div style={{ fontSize: 12, fontWeight: 600, color: '#334155' }}>{tgt?.label ?? ix.target_id}</div>
                        {ix.via && <div style={{ fontSize: 11, color: '#94a3b8' }}>via {ix.via}</div>}
                        {ix.description && <div style={{ fontSize: 11, color: '#64748b', marginTop: 2 }}>{ix.description}</div>}
                      </div>
                    </div>
                  )
                })}
              </div>
            </div>
          )}

          {/* Capabilities */}
          {team.capabilities?.length > 0 && (
            <div>
              <div style={{ fontSize: 11, fontWeight: 700, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: 8 }}>
                Capabilities ({team.capabilities.length})
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
                {team.capabilities.map((cap, i) => (
                  <div key={i} style={{ fontSize: 12, color: '#334155', padding: '6px 10px', borderRadius: 8, background: '#f8fafc', border: '1px solid #f1f5f9' }}>
                    {cap}
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Services */}
          {team.services?.length > 0 && (
            <div>
              <div style={{ fontSize: 11, fontWeight: 700, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: 8 }}>
                Services ({team.services.length})
              </div>
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6 }}>
                {team.services.map((svc, i) => (
                  <span key={i} style={{ ...pill({ fontFamily: 'monospace', fontSize: 11 }), background: '#f1f5f9', color: '#475569', border: '1px solid #e2e8f0' }}>
                    {svc}
                  </span>
                ))}
              </div>
            </div>
          )}

          {/* Anti-patterns */}
          {team.anti_patterns && team.anti_patterns.length > 0 && (
            <div>
              <div style={{ fontSize: 11, fontWeight: 700, color: '#c2410c', textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: 8, cursor: 'help' }}
                title="Detected organizational anti-patterns that may indicate structural issues"
                aria-label="Warning: anti-patterns detected">
                ⚠ Anti-patterns ({team.anti_patterns.length})
              </div>
              {team.anti_patterns.map((ap, i) => (
                <div key={i} style={{ padding: '10px 12px', borderRadius: 12, background: '#fff7ed', border: '1px solid #fed7aa', marginBottom: 6 }}>
                  <div style={{ fontSize: 11, fontWeight: 700, color: '#c2410c', fontFamily: 'monospace', marginBottom: 4 }}>{ap.code}</div>
                  <div style={{ fontSize: 12, color: '#9a3412' }}>{ap.message}</div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </>
  )
}

// ── Main view ─────────────────────────────────────────────────────────────────

export function TeamTopologyView() {
  const { modelId, isHydrating } = useRequireModel()
  const { query, teamTypeFilter } = useSearch()
  const [viewData, setViewData] = useState<TeamTopologyViewResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [viewMode, setViewMode] = useState<'graph' | 'table'>('graph')
  const [filterOverloaded, setFilterOverloaded] = useState(false)
  const [localSearch, setLocalSearch] = useState('')
  const [localTypeFilter, setLocalTypeFilter] = useState<string | null>(null)
  const { insights } = usePageInsights('topology')

  useEffect(() => {
    if (isHydrating || !modelId) { return }
    api.getView(modelId, 'team-topology')
      .then(data => setViewData(data as unknown as TeamTopologyViewResponse))
      .catch(e => setError((e as Error).message))
      .finally(() => setLoading(false))
  }, [isHydrating, modelId])

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

  if (loading) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', gap: 12, minHeight: 200 }}>
        <div style={{ width: 36, height: 36, border: '2px solid #e2e8f0', borderTopColor: '#6366f1', borderRadius: '50%' }} className="animate-spin" />
        <span style={{ fontSize: 14, color: '#94a3b8' }}>Loading…</span>
      </div>
    )
  }
  if (error) return <div style={{ color: '#ef4444', padding: 20 }}>{error}</div>
  if (!viewData) return null

  // Stat breakdowns
  const overloadedCount = viewData.teams.filter(t => t.is_overloaded).length
  const typeCounts = Object.fromEntries(
    Object.keys(TEAM_TYPES).map(k => [k, viewData.teams.filter(t => t.type === k).length])
  )

  return (
    <ModelRequired>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 28 }}>

      {/* ── Header ── */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 16, flexWrap: 'wrap' }}>
          <div>
            <h1 style={gradientTitle}>Team Topology</h1>
            <p style={{ fontSize: 14, color: '#64748b', marginTop: 6 }}>
              {viewData.teams.length} teams · {viewData.interactions.length} interactions
              {overloadedCount > 0 && (
                <button onClick={() => setFilterOverloaded(f => !f)}
                  style={{
                    ...pill({ marginLeft: 10 }), background: filterOverloaded ? '#fde68a' : '#fff7ed',
                    color: '#92400e', border: `1px solid ${filterOverloaded ? '#f59e0b' : '#fcd34d'}`,
                    cursor: 'pointer', fontWeight: filterOverloaded ? 700 : 600,
                  }}
                  title="Click to filter/highlight overloaded teams">
                  {overloadedCount} overloaded
                </button>
              )}
            </p>
          </div>

          {/* View toggle */}
          <div style={{ display: 'inline-flex', background: '#f1f5f9', border: '1px solid #e2e8f0', borderRadius: 12, padding: 4 }}>
            {(['graph', 'table'] as const).map(m => (
              <button key={m} type="button" onClick={() => setViewMode(m)}
                style={{
                  borderRadius: 8, padding: '7px 16px', fontSize: 12, fontWeight: 600, cursor: 'pointer',
                  background: viewMode === m ? 'linear-gradient(135deg, #6366f1 0%, #4f46e5 100%)' : 'transparent',
                  color: viewMode === m ? '#ffffff' : '#64748b', border: 'none',
                  boxShadow: viewMode === m ? '0 2px 8px rgba(99,102,241,0.35)' : 'none',
                  textTransform: 'capitalize', transition: 'all 0.15s',
                }}>
                {m}
              </button>
            ))}
          </div>
        </div>

        {/* Stat cards */}
        <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap' }}>
          <StatCard value={viewData.teams.length} label="Total Teams" icon={Users}
            gradient="linear-gradient(135deg, #ede9fe 0%, #dbeafe 100%)" iconTint="#5b21b6" />
          <StatCard value={typeCounts['platform'] ?? 0} label="Platform"
            sub="shared capability providers"
            icon={Layers} gradient="linear-gradient(135deg, #faf5ff 0%, #f5f3ff 100%)" iconTint="#7c3aed" />
          <StatCard value={typeCounts['stream-aligned'] ?? 0} label="Stream-aligned"
            sub="aligned to value streams"
            icon={ArrowRight} gradient="linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%)" iconTint="#1d4ed8" />
          <StatCard value={viewData.interactions.length} label="Interactions"
            sub={`${overloadedCount > 0 ? overloadedCount + ' overloaded' : 'all healthy'}`}
            icon={Zap} gradient="linear-gradient(135deg, #f0fdf4 0%, #dcfce7 100%)" iconTint="#15803d" />
        </div>

        {/* Team type legend */}
        <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
          {Object.entries(TEAM_TYPES).map(([key, cfg]) => {
            const count = typeCounts[key] ?? 0
            if (count === 0) return null
            return (
              <span key={key} style={{ ...pill(), background: cfg.bg, color: cfg.accent, border: `1px solid ${cfg.border}` }}
                title={TEAM_TYPE_DESCRIPTIONS[key]}>
                <span style={{ width: 6, height: 6, borderRadius: '50%', background: cfg.accent, marginRight: 5, display: 'inline-block' }} />
                {cfg.label}
                <span style={{ marginLeft: 5, fontWeight: 800 }}>{count}</span>
              </span>
            )
          })}
          <span style={{ ...pill(), background: '#f1f5f9', color: '#475569', border: '1px solid #e2e8f0' }}>
            <span style={{ display: 'inline-block', width: 16, height: 2, background: '#7c3aed', marginRight: 5, borderRadius: 1 }} />
            X-as-a-Service
          </span>
          <span style={{ ...pill(), background: '#f1f5f9', color: '#475569', border: '1px solid #e2e8f0' }}>
            <span style={{ display: 'inline-block', width: 16, height: 2, background: '#1d4ed8', marginRight: 5, borderRadius: 1 }} />
            Collaboration
          </span>
          <span style={{ ...pill(), background: '#f1f5f9', color: '#475569', border: '1px solid #e2e8f0' }}>
            <span style={{ display: 'inline-block', width: 12, height: 2, background: '#15803d', marginRight: 2, borderRadius: 1, borderStyle: 'dashed' }} />
            <span style={{ display: 'inline-block', width: 3, height: 2, background: '#15803d', marginRight: 5, borderRadius: 1 }} />
            Facilitating
          </span>
        </div>

        {/* Search and team type filter (UI-38) */}
        <div style={{ display: 'flex', alignItems: 'center', gap: 10, flexWrap: 'wrap' }}>
          <input value={localSearch} onChange={e => setLocalSearch(e.target.value)}
            placeholder="Search teams..."
            style={{ padding: '5px 10px', border: '1px solid #d1d5db', borderRadius: 6, fontSize: 13 }} />
          {Object.entries(TEAM_TYPES).map(([key, cfg]) => {
            const active = localTypeFilter === key
            return (
              <button key={key} type="button"
                onClick={() => setLocalTypeFilter(active ? null : key)}
                style={{
                  padding: '3px 10px', borderRadius: 20, fontSize: 12, cursor: 'pointer',
                  background: active ? cfg.accent : 'white',
                  color: active ? 'white' : cfg.accent,
                  border: `1px solid ${active ? cfg.accent : '#d1d5db'}`,
                }}
                title={TEAM_TYPE_DESCRIPTIONS[key]}>
                {cfg.label}
              </button>
            )
          })}
        </div>
      </div>

      {/* ── Views ── */}
      {viewMode === 'graph' ? (
        <GraphView teams={filteredTeams} interactions={filteredInteractions} insights={insights} filterOverloaded={filterOverloaded} />
      ) : (
        <TableView teams={filteredTeams} interactions={filteredInteractions} insights={insights} />
      )}
    </div>
    </ModelRequired>
  )
}
