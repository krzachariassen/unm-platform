import { Info, Lightbulb, Sparkles } from 'lucide-react'
import { SlidePanel, PanelSection } from '@/components/ui/slide-panel'
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

  const pillCls = 'inline-flex items-center rounded px-2 py-0.5 text-[10px] font-semibold'

  function IxList({ items, getOther }: { items: TeamTopologyInteraction[]; getOther: (ix: TeamTopologyInteraction) => string }) {
    return (
      <div className="flex flex-col gap-1">
        {items.map((ix, i) => {
          const s = getIx(ix.mode)
          const other = teamById.get(getOther(ix))
          return (
            <div key={i} className="flex items-start gap-2 px-2.5 py-1.5 rounded-lg bg-muted border border-border">
              <span className={pillCls} style={{ background: s.bg, color: s.text, border: `1px solid ${s.border}`, flexShrink: 0 }}>{s.label}</span>
              <div className="min-w-0">
                <div className="text-xs font-medium text-foreground">{other?.label ?? getOther(ix)}</div>
                {ix.via && <div className="text-[10px] text-muted-foreground">via {ix.via}</div>}
                {ix.description && <div className="text-[10px] text-muted-foreground mt-0.5">{ix.description}</div>}
              </div>
            </div>
          )
        })}
      </div>
    )
  }

  return (
    <SlidePanel
      open
      onClose={onClose}
      title={team.label}
      badge={
        <div className="flex items-center gap-2">
          <span className={pillCls} style={{ background: cfg.bg, color: cfg.accent, border: `1px solid ${cfg.border}` }}
            title={TEAM_TYPE_DESCRIPTIONS[team.type]}>
            {cfg.label}
          </span>
          <QuickAction size={11} options={[
            { label: 'Change team type', action: { type: 'update_team_type', team_name: team.label } },
            { label: 'Update team size', action: { type: 'update_team_size', team_name: team.label } },
            { label: 'Add service to team', action: { type: 'add_service', owner_team_name: team.label } },
          ]} />
          {team.is_overloaded && (
            <span className={pillCls} style={{ background: '#fff7ed', color: '#c2410c', border: '1px solid #fed7aa' }}
              title="This team owns too many services or capabilities">
              Overloaded
            </span>
          )}
        </div>
      }
    >
      <div className="space-y-3">
        {team.description && <p className="text-xs text-muted-foreground leading-relaxed">{team.description}</p>}

        {/* Metric pills */}
        <div className="flex gap-2 flex-wrap">
          {[
            { v: team.capability_count, l: 'capabilities' },
            { v: team.service_count,    l: 'services' },
            { v: team.interactions?.length ?? 0, l: 'interactions' },
          ].map(({ v, l }) => (
            <div key={l} className="rounded-lg px-3 py-1.5 text-center bg-muted border border-border">
              <div className="text-lg font-bold text-foreground">{v}</div>
              <div className="text-[9px] font-semibold uppercase tracking-wider text-muted-foreground">{l}</div>
            </div>
          ))}
        </div>

        {teamInsight && (
          <div className="rounded-lg p-3 border border-border bg-muted/50">
            <div className="flex items-center gap-1.5 mb-1.5">
              <Sparkles size={11} className="text-primary" />
              <span className="text-[10px] font-semibold text-primary uppercase tracking-wide">AI Insight</span>
            </div>
            <div className="flex gap-2 mb-1.5">
              <Info size={11} className="text-primary shrink-0 mt-0.5" />
              <p className="text-xs text-foreground leading-relaxed">{teamInsight.explanation}</p>
            </div>
            {teamInsight.suggestion && (
              <div className="flex gap-2">
                <Lightbulb size={11} className="text-primary shrink-0 mt-0.5" />
                <p className="text-xs text-primary leading-relaxed font-medium">{teamInsight.suggestion}</p>
              </div>
            )}
          </div>
        )}

        {inbound.length > 0 && (
          <PanelSection label={`Inbound (${inbound.length})`}>
            <IxList items={inbound} getOther={ix => ix.source_id} />
          </PanelSection>
        )}
        {outbound.length > 0 && (
          <PanelSection label={`Outbound (${outbound.length})`}>
            <IxList items={outbound} getOther={ix => ix.target_id} />
          </PanelSection>
        )}

        {team.capabilities?.length > 0 && (
          <PanelSection label={`Capabilities (${team.capabilities.length})`}>
            <div className="flex flex-col gap-0.5">
              {team.capabilities.map((cap, i) => (
                <div key={i} className="text-xs text-foreground px-2 py-1 rounded bg-muted border border-border">{cap}</div>
              ))}
            </div>
          </PanelSection>
        )}

        {team.services?.length > 0 && (
          <PanelSection label={`Services (${team.services.length})`}>
            <div className="flex flex-wrap gap-1">
              {team.services.map((svc, i) => (
                <span key={i} className="font-mono text-[10px] px-2 py-0.5 rounded bg-muted text-muted-foreground border border-border">{svc}</span>
              ))}
            </div>
          </PanelSection>
        )}

        {team.anti_patterns && team.anti_patterns.length > 0 && (
          <PanelSection label={`Anti-patterns (${team.anti_patterns.length})`}>
            {team.anti_patterns.map((ap, i) => (
              <div key={i} className="px-2.5 py-1.5 rounded-lg bg-orange-50 border border-orange-200 mb-1">
                <div className="text-[10px] font-bold text-orange-700 font-mono mb-0.5">{ap.code}</div>
                <div className="text-xs text-orange-800">{ap.message}</div>
              </div>
            ))}
          </PanelSection>
        )}
      </div>
    </SlidePanel>
  )
}

export type { TeamTopologyTeam, TeamTopologyInteraction }
export { INTERACTION_STYLE }
