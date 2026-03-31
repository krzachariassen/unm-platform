import { SlidePanel, PanelSection } from '@/components/ui/slide-panel'
import type { CapabilityType } from './CapabilityCard'
import type { InsightItem } from '@/types/insights'
import { VIS_BANDS } from './constants'
import { Sparkles, Lightbulb } from 'lucide-react'

export function DetailPanel({ cap, allCaps, onClose, insight }: {
  cap: CapabilityType
  allCaps: CapabilityType[]
  onClose: () => void
  insight?: InsightItem
}) {
  const visBand = VIS_BANDS.find(b => b.key === cap.visibility)
  const dependedOnBy = allCaps.filter(c => c.depends_on.some(d => d.id === cap.id))
  const vb = visBand ?? { label: cap.visibility, accent: '#475569', border: '#e2e8f0', bg: '#f1f5f9' }

  const pillCls = 'text-[10px] font-semibold px-2 py-0.5 rounded'

  return (
    <SlidePanel
      open
      onClose={onClose}
      title={cap.label}
      badge={
        <span className={pillCls} style={{ background: vb.bg, color: vb.accent, border: `1px solid ${vb.border}` }}>
          {vb.label}
        </span>
      }
    >
      <div className="space-y-3">
        <PanelSection label="Description">
          {cap.description
            ? <p className="text-xs leading-relaxed text-muted-foreground">{cap.description}</p>
            : <p className="text-xs italic text-muted-foreground/60">None</p>}
        </PanelSection>

        {insight && (
          <div className="rounded-lg p-3 border border-border bg-muted/50">
            <div className="flex items-center gap-1.5 mb-1.5">
              <Sparkles size={11} className="text-primary" />
              <span className="text-[10px] font-semibold text-primary uppercase tracking-wide">AI Insight</span>
            </div>
            <p className="text-xs leading-relaxed text-foreground">{insight.explanation}</p>
            {insight.suggestion && (
              <div className="flex items-start gap-1.5 mt-1.5">
                <Lightbulb size={11} className="text-primary shrink-0 mt-0.5" />
                <p className="text-xs leading-relaxed text-primary font-medium">{insight.suggestion}</p>
              </div>
            )}
          </div>
        )}

        <PanelSection label="Teams">
          {cap.teams.length > 0
            ? <div className="flex flex-wrap gap-1">{cap.teams.map(t => <span key={t.id} className="text-[10px] rounded px-2 py-0.5 bg-muted text-foreground border border-border">{t.label} <span className="text-muted-foreground">({t.type})</span></span>)}</div>
            : <p className="text-xs italic text-destructive/70">No team assigned</p>}
        </PanelSection>

        <PanelSection label="Services">
          {cap.services.length > 0
            ? <div className="flex flex-wrap gap-1">{cap.services.map(s => <span key={s.id} className="font-mono text-[10px] rounded px-2 py-0.5 bg-muted text-foreground border border-border">{s.label}</span>)}</div>
            : <p className="text-xs italic text-muted-foreground/60">None</p>}
        </PanelSection>

        <PanelSection label="Depends on">
          {cap.depends_on.length > 0
            ? <div className="flex flex-wrap gap-1">{cap.depends_on.map(d => <span key={d.id} className="text-[10px] rounded px-2 py-0.5 bg-muted text-foreground border border-border">{d.label}</span>)}</div>
            : <p className="text-xs italic text-muted-foreground/60">None</p>}
        </PanelSection>

        <PanelSection label="Depended on by">
          {dependedOnBy.length > 0
            ? <ul className="text-xs space-y-0.5 text-foreground">{dependedOnBy.map(c => <li key={c.id} className="text-muted-foreground">• {c.label}</li>)}</ul>
            : <p className="text-xs italic text-muted-foreground/60">None</p>}
        </PanelSection>

        {cap.external_deps && cap.external_deps.length > 0 && (
          <PanelSection label="External Dependencies">
            <div className="space-y-1">
              {cap.external_deps.map(dep => (
                <div key={dep.name} className="flex items-center gap-2 px-2.5 py-1.5 rounded-lg bg-muted border border-border">
                  <span className="text-xs font-medium text-foreground">{dep.name}</span>
                  {dep.description && <span className="text-[10px] text-muted-foreground">{dep.description}</span>}
                </div>
              ))}
            </div>
          </PanelSection>
        )}

        {(cap.is_fragmented || cap.depended_on_by_count >= 3 || (cap.anti_patterns?.length ?? 0) > 0) && (
          <PanelSection label="Anti-patterns">
            <div className="space-y-1.5">
              {cap.is_fragmented && <div className="text-[10px] rounded-lg px-2.5 py-1.5 bg-destructive/10 text-destructive border border-destructive/20">Fragmented — owned by multiple teams</div>}
              {cap.depended_on_by_count >= 3 && <div className="text-[10px] rounded-lg px-2.5 py-1.5 bg-amber-50 text-amber-800 border border-amber-200">High fan-in — {cap.depended_on_by_count} capabilities depend on this</div>}
              {cap.anti_patterns?.map((ap, i) => <div key={i} className="text-[10px] rounded-lg px-2.5 py-1.5 bg-orange-50 text-orange-700 border border-orange-200">{ap.message}</div>)}
            </div>
          </PanelSection>
        )}

        {!cap.is_leaf && cap.children.length > 0 && (
          <PanelSection label="Sub-capabilities">
            <ul className="text-xs space-y-0.5 text-muted-foreground">{cap.children.map(ch => <li key={ch.id}>• {ch.label}</li>)}</ul>
          </PanelSection>
        )}
      </div>
    </SlidePanel>
  )
}
