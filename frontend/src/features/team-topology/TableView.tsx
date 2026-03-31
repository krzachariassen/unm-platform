import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { getType, getIx, TEAM_TYPES, TEAM_TYPE_DESCRIPTIONS } from './constants'
import { DetailPanel } from './DetailPanel'
import type { TeamTopologyTeam, TeamTopologyInteraction } from './constants'

function pill(extra?: React.CSSProperties): React.CSSProperties {
  return { display: 'inline-flex', alignItems: 'center', borderRadius: 8, padding: '3px 10px', fontSize: 11, fontWeight: 600, ...extra }
}

export function TableView({ teams, interactions, insights }: {
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
              <div className="flex items-center gap-3 mb-4">
                <div style={{ flex: 1, height: 1, background: `linear-gradient(90deg, transparent, ${cfg.border}, transparent)` }} />
                <span style={{ ...pill({ fontSize: 11, textTransform: 'uppercase', letterSpacing: '0.06em', whiteSpace: 'nowrap', boxShadow: '0 1px 3px rgba(0,0,0,0.04)' }), background: cfg.bg, color: cfg.accent, border: `1px solid ${cfg.border}` }}>
                  {cfg.label} · {typeTeams.length} team{typeTeams.length !== 1 ? 's' : ''}
                </span>
                <div style={{ flex: 1, height: 1, background: `linear-gradient(90deg, ${cfg.border}, transparent)` }} />
              </div>

              <div className="grid gap-4" style={{ gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))' }}>
                {typeTeams.map(team => {
                  const myIx = team.interactions ?? []
                  return (
                    <div key={team.id} onClick={() => setDetailTeam(team)}
                      className="rounded-lg border border-border bg-card cursor-pointer overflow-hidden transition-all"
                      style={{ border: `1.5px solid ${team.is_overloaded ? '#fca5a5' : '#e2e8f0'}`, boxShadow: '0 1px 4px rgba(15,23,42,0.06)' }}
                      onMouseEnter={e => { (e.currentTarget as HTMLDivElement).style.transform = 'translateY(-2px)'; (e.currentTarget as HTMLDivElement).style.boxShadow = '0 12px 32px rgba(15,23,42,0.1)' }}
                      onMouseLeave={e => { (e.currentTarget as HTMLDivElement).style.transform = ''; (e.currentTarget as HTMLDivElement).style.boxShadow = '0 1px 4px rgba(15,23,42,0.06)' }}
                    >
                      <div style={{ height: 4, background: `linear-gradient(90deg, ${cfg.gradientFrom} 0%, ${cfg.gradientTo} 100%)` }} />
                      <div className="p-4">
                        <div className="flex items-start justify-between gap-2 mb-2">
                          <span className="text-sm font-bold text-slate-900 leading-snug">{team.label}</span>
                          {team.is_overloaded && (
                            <span style={{ ...pill({ padding: '3px 9px', fontSize: 10, flexShrink: 0 }), background: '#fff7ed', color: '#c2410c', border: '1px solid #fed7aa' }}
                              title="Overloaded">overloaded</span>
                          )}
                        </div>
                        {team.description && (
                          <p className="text-xs text-slate-500 mb-2.5 leading-snug line-clamp-2">{team.description}</p>
                        )}
                        <div className="flex gap-4 pb-2.5 mb-2.5 border-b border-slate-100">
                          {[{ v: team.capability_count, l: 'caps' }, { v: team.service_count, l: 'svcs' }, { v: myIx.length, l: 'ixns' }].map(({ v, l }) => (
                            <div key={l}>
                              <span className="text-lg font-extrabold text-slate-900">{v}</span>
                              <span className="text-[11px] text-slate-400 ml-1">{l}</span>
                            </div>
                          ))}
                        </div>
                        {myIx.length > 0 ? (
                          <div className="flex flex-col gap-1.5">
                            {myIx.slice(0, 4).map((ix, i) => {
                              const isSource = ix.source_id === team.id
                              const other = teamById.get(isSource ? ix.target_id : ix.source_id)
                              const s = getIx(ix.mode)
                              return (
                                <div key={i} className="flex items-center gap-1.5 text-xs">
                                  <span style={{ ...pill({ fontSize: 10, padding: '2px 8px', flexShrink: 0 }), background: s.bg, color: s.text, border: `1px solid ${s.border}` }}>
                                    {s.label}
                                  </span>
                                  <span className="text-slate-400">{isSource ? '→' : '←'}</span>
                                  <span className="text-slate-700 font-medium truncate">{other?.label ?? (isSource ? ix.target_id : ix.source_id)}</span>
                                </div>
                              )
                            })}
                            {myIx.length > 4 && <span className="text-[11px] text-slate-400">+{myIx.length - 4} more</span>}
                          </div>
                        ) : (
                          <span className="text-xs text-slate-400 italic">No interactions modelled</span>
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
          <div className="rounded-lg border border-border overflow-hidden">
            <div style={{ height: 4, background: 'linear-gradient(90deg, #6366f1 0%, #8b5cf6 50%, #ec4899 100%)' }} />
            <div className="px-5 py-3 border-b border-slate-200 bg-gradient-to-r from-slate-50 to-white">
              <span className="text-[11px] font-bold text-slate-500 uppercase tracking-wider">All Interactions ({interactions.length})</span>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full bg-white border-collapse text-sm">
                <thead>
                  <tr className="bg-slate-50 border-b border-slate-200">
                    {['From', 'Mode', 'To', 'Via', 'Description'].map(h => (
                      <th key={h} className="text-left px-5 py-2.5 text-[11px] font-semibold text-slate-500 uppercase tracking-wider">{h}</th>
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
                      <tr key={idx} className={idx % 2 === 0 ? 'bg-white' : 'bg-slate-50/60'} style={{ borderBottom: idx < interactions.length - 1 ? '1px solid #f1f5f9' : 'none' }}>
                        <td className="px-5 py-2.5 text-xs font-semibold text-slate-700">
                          {src?.label ?? ix.source_id}
                          {src && <span className="text-[11px] ml-1 px-1.5 py-0.5 rounded-full" style={{ background: getType(src.type).accent + '20', color: getType(src.type).accent }}
                            title={TEAM_TYPE_DESCRIPTIONS[src.type]}>{getType(src.type).label}</span>}
                        </td>
                        <td className="px-5 py-2.5">
                          <span style={{ ...pill(), background: s.bg, color: s.text, border: `1px solid ${s.border}` }}>{s.label}</span>
                        </td>
                        <td className="px-5 py-2.5 text-xs font-semibold text-slate-700">
                          {tgt?.label ?? ix.target_id}
                          {tgt && <span className="text-[11px] ml-1 px-1.5 py-0.5 rounded-full" style={{ background: getType(tgt.type).accent + '20', color: getType(tgt.type).accent }}
                            title={TEAM_TYPE_DESCRIPTIONS[tgt.type]}>{getType(tgt.type).label}</span>}
                          {isSelf && <span className="text-[10px] ml-1.5 px-1.5 py-0.5 rounded-full bg-slate-100 text-slate-500">self</span>}
                        </td>
                        <td className="px-5 py-2.5 text-xs text-slate-500 font-mono">
                          {ix.via ? (
                            <button onClick={() => navigate('/capability')}
                              className="text-blue-500 bg-transparent border-0 cursor-pointer text-xs underline p-0 font-mono">
                              {ix.via}
                            </button>
                          ) : '—'}
                        </td>
                        <td className="px-5 py-2.5 text-xs text-slate-500 max-w-[200px] truncate" title={ix.description || undefined}>
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
