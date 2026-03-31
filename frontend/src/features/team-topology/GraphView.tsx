import { useMemo, useState } from 'react'
import { ChevronRight } from 'lucide-react'
import { QuickAction } from '@/components/changeset/QuickAction'
import { getType, getIx, TEAM_TYPES, TEAM_TYPE_DESCRIPTIONS, INTERACTION_STYLE, COLUMNS, NODE_W, NODE_H, COL_PAD, HEADER_H, ROW_GAP } from './constants'
import { DetailPanel } from './DetailPanel'
import type { TeamTopologyTeam, TeamTopologyInteraction } from './constants'

function pill(extra?: React.CSSProperties): React.CSSProperties {
  return { display: 'inline-flex', alignItems: 'center', borderRadius: 8, padding: '3px 10px', fontSize: 11, fontWeight: 600, ...extra }
}

export function GraphView({ teams, interactions, insights, filterOverloaded }: {
  teams: TeamTopologyTeam[]
  interactions: TeamTopologyInteraction[]
  insights: Record<string, { explanation: string; suggestion: string }>
  filterOverloaded: boolean
}) {
  const [focusedTeam, setFocusedTeam] = useState<string | null>(null)
  const [detailTeam, setDetailTeam] = useState<TeamTopologyTeam | null>(null)
  const [modeFilters, setModeFilters] = useState({ 'x-as-a-service': true, collaboration: true, facilitating: true })

  const teamById = useMemo(() => new Map(teams.map(t => [t.id, t])), [teams])

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
      <div className="flex items-center gap-2 flex-wrap mb-4">
        <span className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider mr-1">Show</span>
        {Object.entries(INTERACTION_STYLE).map(([mode, s]) => {
          const active = modeFilters[mode as keyof typeof modeFilters]
          return (
            <button key={mode} type="button"
              onClick={() => setModeFilters(p => ({ ...p, [mode]: !p[mode as keyof typeof p] }))}
              style={{ background: active ? s.bg : '#f1f5f9', color: active ? s.text : '#94a3b8', border: `1px solid ${active ? s.border : '#e2e8f0'}` }}
              className="text-[11px] font-semibold rounded-full px-3.5 py-1 cursor-pointer transition-all">
              {s.label}
            </button>
          )
        })}
        <span className="text-[11px] text-slate-400 ml-1">
          {visibleInteractions.length} interaction{visibleInteractions.length !== 1 ? 's' : ''} visible
        </span>
      </div>

      {/* Graph canvas */}
      <div className="relative overflow-x-auto rounded-2xl border border-slate-200 shadow-sm"
        style={{ background: 'linear-gradient(135deg, #fafbff 0%, #f8fafc 100%)' }}>
        {/* Column zone backgrounds */}
        {COLUMNS.map(col => (
          <div key={col.label} style={{
            position: 'absolute', top: 0, bottom: 0,
            left: col.x - 12, width: NODE_W + 24,
            background: TEAM_TYPES[col.types[0]]?.zoneBg ?? 'transparent',
            borderRight: '1px dashed #e2e8f0', pointerEvents: 'none',
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
            const isFocus = focusedTeam ? ix.source_id === focusedTeam || ix.target_id === focusedTeam : true
            const opacity = focusedTeam ? (isFocus ? 1 : 0.08) : 0.75
            const strokeW = focusedTeam && isFocus ? 2.5 : 1.5
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
                markerEnd={`url(#arrow-${ix.mode})`}>
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
                boxShadow: isFocused ? `0 0 0 3px ${cfg.accent}33, 0 12px 32px rgba(15,23,42,0.14)` : '0 2px 8px rgba(15,23,42,0.06)',
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
              <div style={{ width: 4, flexShrink: 0, background: `linear-gradient(180deg, ${cfg.gradientFrom} 0%, ${cfg.gradientTo} 100%)` }} />
              <div style={{ flex: 1, padding: '10px 12px', minWidth: 0, display: 'flex', flexDirection: 'column', justifyContent: 'space-between' }}>
                <div className="flex items-center justify-between gap-1.5">
                  <span style={{ ...pill({ textTransform: 'uppercase', letterSpacing: '0.04em', fontSize: 9 }), background: cfg.bg, color: cfg.accent, border: `1px solid ${cfg.border}` }}
                    title={TEAM_TYPE_DESCRIPTIONS[team.type]}>
                    {cfg.label}
                  </span>
                  <div className="flex items-center gap-1">
                    {overloaded && (
                      <span style={{ ...pill({ fontSize: 9, padding: '2px 7px' }), background: '#fff7ed', color: '#c2410c', border: '1px solid #fed7aa' }}
                        title="Overloaded">overloaded</span>
                    )}
                    {fi > 0 && <span className="text-[10px] font-semibold text-slate-400">← {fi}</span>}
                  </div>
                </div>
                <div className="text-[13px] font-bold text-slate-900 truncate -tracking-tight mt-1.5 flex items-center gap-1">
                  <span className="truncate">{team.label}</span>
                  <span className="qa-node" style={{ opacity: 0.35, transition: 'opacity 0.15s' }}>
                    <QuickAction size={11} options={[
                      { label: 'Change team type', action: { type: 'update_team_type', team_name: team.label } },
                      { label: 'Update team size', action: { type: 'update_team_size', team_name: team.label } },
                      { label: 'Add service to team', action: { type: 'add_service', owner_team_name: team.label } },
                    ]} />
                  </span>
                </div>
                <div className="flex gap-2.5 mt-1.5">
                  <span className="text-[11px] text-slate-500"><span className="font-bold text-slate-700">{team.capability_count}</span> caps</span>
                  <span className="text-[11px] text-slate-500"><span className="font-bold text-slate-700">{team.service_count}</span> svcs</span>
                  <span className="text-[11px] text-slate-500"><span className="font-bold text-slate-700">{team.interactions?.length ?? 0}</span> ixns</span>
                </div>
              </div>
              <div className="flex items-center pr-2 text-slate-300"><ChevronRight size={14} /></div>
            </div>
          )
        })}

        <div style={{ height: svgH, width: svgW, pointerEvents: 'none' }} />
      </div>

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
