import { X, Info, Lightbulb, Sparkles } from 'lucide-react'
import { slug } from '@/lib/slug'
import { QuickAction } from '@/components/changeset/QuickAction'
import { getType, getIx, TEAM_TYPE_DESCRIPTIONS, INTERACTION_STYLE } from './constants'
import type { TeamTopologyTeam, TeamTopologyInteraction } from './constants'

export function DetailPanel({ team, teamById, allInteractions, insights, onClose }: {
  team: TeamTopologyTeam
  teamById: Map<string, TeamTopologyTeam>
  allInteractions: TeamTopologyInteraction[]
  insights: Record<string, { explanation: string; suggestion: string }>
  onClose: () => void
}) {
  const cfg = getType(team.type)
  const inbound  = allInteractions.filter(ix => ix.target_id === team.id)
  const outbound = allInteractions.filter(ix => ix.source_id === team.id)

  const aiInsight = team.interactions
    .map(ix => {
      const a = slug(team.id), b = slug(ix.source_id === team.id ? ix.target_id : ix.source_id)
      return insights[`interaction:${a}:${b}`] ?? insights[`interaction:${b}:${a}`]
    })
    .find(Boolean)
  const teamInsight = insights[`team:${slug(team.label)}`] ?? aiInsight

  const pillCls = 'inline-flex items-center rounded-lg px-2.5 py-0.5 text-[11px] font-semibold'

  function IxList({ items, getOther }: { items: TeamTopologyInteraction[]; getOther: (ix: TeamTopologyInteraction) => string }) {
    return (
      <div className="flex flex-col gap-1.5">
        {items.map((ix, i) => {
          const s = getIx(ix.mode)
          const other = teamById.get(getOther(ix))
          return (
            <div key={i} className="flex items-start gap-2 px-3 py-2 rounded-xl bg-slate-50 border border-slate-100">
              <span className={pillCls} style={{ background: s.bg, color: s.text, border: `1px solid ${s.border}`, flexShrink: 0 }}>{s.label}</span>
              <div className="min-w-0">
                <div className="text-xs font-semibold text-slate-700">{other?.label ?? getOther(ix)}</div>
                {ix.via && <div className="text-[11px] text-slate-400">via {ix.via}</div>}
                {ix.description && <div className="text-[11px] text-slate-500 mt-0.5">{ix.description}</div>}
              </div>
            </div>
          )
        })}
      </div>
    )
  }

  return (
    <>
      <div onClick={onClose} className="fixed inset-0 bg-slate-900/15 backdrop-blur-sm z-40" />
      <div className="fixed top-0 right-0 w-[360px] h-full bg-white shadow-2xl z-50 overflow-y-auto flex flex-col">
        <div style={{ height: 4, background: `linear-gradient(90deg, ${cfg.gradientFrom} 0%, ${cfg.gradientTo} 100%)`, flexShrink: 0 }} />

        <div className="p-5 pb-8 flex flex-col gap-5">
          {/* Header */}
          <div className="flex items-start justify-between gap-3">
            <div className="min-w-0">
              <span className={pillCls} style={{ background: cfg.bg, color: cfg.accent, border: `1px solid ${cfg.border}` }}
                title={TEAM_TYPE_DESCRIPTIONS[team.type]}>
                {cfg.label}
              </span>
              <div className="flex items-center gap-2 mt-2">
                <h2 className="text-lg font-extrabold text-slate-900 tracking-tight">{team.label}</h2>
                <QuickAction size={11} options={[
                  { label: 'Change team type', action: { type: 'update_team_type', team_name: team.label } },
                  { label: 'Update team size', action: { type: 'update_team_size', team_name: team.label } },
                  { label: 'Add service to team', action: { type: 'add_service', owner_team_name: team.label } },
                ]} />
              </div>
              {team.is_overloaded && (
                <span className={pillCls + ' mt-2'} style={{ background: '#fff7ed', color: '#c2410c', border: '1px solid #fed7aa' }}
                  title="This team owns too many services or capabilities, increasing cognitive load">
                  ⚠ Overloaded
                </span>
              )}
            </div>
            <button onClick={onClose} className="rounded-xl w-8 h-8 bg-slate-100 border-0 cursor-pointer flex items-center justify-center text-slate-500 shrink-0">
              <X size={16} />
            </button>
          </div>

          {team.description && <p className="text-sm text-slate-500 leading-relaxed m-0">{team.description}</p>}

          {/* Metric pills */}
          <div className="flex gap-2 flex-wrap">
            {[
              { v: team.capability_count, l: 'capabilities', color: '#5b21b6', bg: '#ede9fe' },
              { v: team.service_count,    l: 'services',     color: '#1d4ed8', bg: '#dbeafe' },
              { v: team.interactions?.length ?? 0, l: 'interactions', color: '#15803d', bg: '#d1fae5' },
            ].map(({ v, l, color, bg }) => (
              <div key={l} className="rounded-xl px-3.5 py-2 text-center" style={{ background: bg }}>
                <div className="text-xl font-extrabold" style={{ color }}>{v}</div>
                <div className="text-[10px] font-semibold uppercase tracking-wider" style={{ color, opacity: 0.8 }}>{l}</div>
              </div>
            ))}
          </div>

          {/* AI insight */}
          {teamInsight && (
            <div className="rounded-2xl p-4 bg-gradient-to-br from-indigo-50 to-sky-50 border border-indigo-200">
              <div className="flex items-center gap-1.5 mb-2.5">
                <Sparkles size={13} className="text-indigo-600" />
                <span className="text-[11px] font-bold text-indigo-600 uppercase tracking-wider">AI Insight</span>
              </div>
              <div className="flex gap-2 mb-2">
                <Info size={12} className="text-indigo-700 shrink-0 mt-0.5" />
                <p className="text-xs text-indigo-900 leading-relaxed m-0">{teamInsight.explanation}</p>
              </div>
              {teamInsight.suggestion && (
                <div className="flex gap-2">
                  <Lightbulb size={12} className="text-indigo-700 shrink-0 mt-0.5" />
                  <p className="text-xs text-indigo-800 leading-relaxed m-0 font-medium">{teamInsight.suggestion}</p>
                </div>
              )}
            </div>
          )}

          {inbound.length > 0 && (
            <div>
              <div className="text-[11px] font-bold text-slate-500 uppercase tracking-wider mb-2">Inbound ({inbound.length})</div>
              <IxList items={inbound} getOther={ix => ix.source_id} />
            </div>
          )}
          {outbound.length > 0 && (
            <div>
              <div className="text-[11px] font-bold text-slate-500 uppercase tracking-wider mb-2">Outbound ({outbound.length})</div>
              <IxList items={outbound} getOther={ix => ix.target_id} />
            </div>
          )}

          {team.capabilities?.length > 0 && (
            <div>
              <div className="text-[11px] font-bold text-slate-500 uppercase tracking-wider mb-2">Capabilities ({team.capabilities.length})</div>
              <div className="flex flex-col gap-1">
                {team.capabilities.map((cap, i) => (
                  <div key={i} className="text-xs text-slate-700 px-2.5 py-1.5 rounded-lg bg-slate-50 border border-slate-100">{cap}</div>
                ))}
              </div>
            </div>
          )}

          {team.services?.length > 0 && (
            <div>
              <div className="text-[11px] font-bold text-slate-500 uppercase tracking-wider mb-2">Services ({team.services.length})</div>
              <div className="flex flex-wrap gap-1.5">
                {team.services.map((svc, i) => (
                  <span key={i} className="font-mono text-[11px] px-2.5 py-0.5 rounded-lg bg-slate-100 text-slate-500 border border-slate-200">{svc}</span>
                ))}
              </div>
            </div>
          )}

          {team.anti_patterns && team.anti_patterns.length > 0 && (
            <div>
              <div className="text-[11px] font-bold text-orange-700 uppercase tracking-wider mb-2"
                title="Detected organizational anti-patterns">
                ⚠ Anti-patterns ({team.anti_patterns.length})
              </div>
              {team.anti_patterns.map((ap, i) => (
                <div key={i} className="px-3 py-2.5 rounded-xl bg-orange-50 border border-orange-200 mb-1.5">
                  <div className="text-[11px] font-bold text-orange-700 font-mono mb-1">{ap.code}</div>
                  <div className="text-xs text-orange-800">{ap.message}</div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </>
  )
}

// Re-export for use in GraphView/TableView
export type { TeamTopologyTeam, TeamTopologyInteraction }
export { INTERACTION_STYLE }
