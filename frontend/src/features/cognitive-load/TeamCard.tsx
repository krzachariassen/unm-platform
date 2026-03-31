import { ChevronDown, ChevronUp, Shield, ArrowRight } from 'lucide-react'
import { Speedometer } from '@/components/ui/speedometer'
import { TEAM_TYPE_BADGE } from '@/lib/team-type-styles'
import type { TeamLoad } from '@/types/views'
import type { InsightItem } from '@/types/insights'
import { cn } from '@/lib/utils'

const LEVEL = {
  low:    { label: 'Low',    border: 'border-green-200',  bg: 'bg-green-50',  text: 'text-green-700',  dot: '#22c55e' },
  medium: { label: 'Medium', border: 'border-amber-200',  bg: 'bg-amber-50',  text: 'text-amber-700',  dot: '#f59e0b' },
  high:   { label: 'High',   border: 'border-red-200',    bg: 'bg-red-50',    text: 'text-red-700',    dot: '#ef4444' },
} as const

const DIMENSIONS = [
  { key: 'domain_spread'    as const, abbr: 'Dom',  label: 'Domain Spread',      thresholds: '1-3 low · 4-5 med · 6+ high' },
  { key: 'service_load'     as const, abbr: 'Svc',  label: 'Service Load',       thresholds: '≤2 low · 2-3 med · >3 high' },
  { key: 'interaction_load' as const, abbr: 'Ixn',  label: 'Interaction Load',   thresholds: '≤3 low · 4-6 med · 7+ high' },
  { key: 'dependency_load'  as const, abbr: 'Deps', label: 'Dependency Fan-out', thresholds: '≤4 low · 5-8 med · 9+ high' },
] as const

const levelBadge: Record<string, string> = {
  low:    'bg-green-100 text-green-700',
  medium: 'bg-amber-100 text-amber-700',
  high:   'bg-red-100 text-red-700',
}

export function TeamCard({ tl, insight, isExpanded, onToggle }: {
  tl: TeamLoad; insight?: InsightItem; isExpanded: boolean; onToggle: () => void
}) {
  const ls = LEVEL[tl.overall_level as keyof typeof LEVEL] ?? LEVEL.low
  const typeStyle = TEAM_TYPE_BADGE[tl.team.type] ?? { bg: '#f1f5f9', text: '#475569' }
  const dims = DIMENSIONS.map(d => ({ abbr: d.abbr, level: tl[d.key].level, value: tl[d.key].value }))

  return (
    <div className={cn('rounded-lg overflow-hidden bg-card border transition-all duration-200', isExpanded ? ls.border : 'border-border')}>
      <div className={cn('h-0.5', tl.overall_level === 'high' ? 'bg-red-500' : tl.overall_level === 'medium' ? 'bg-amber-500' : 'bg-green-500')} />
      <div onClick={onToggle} className="p-4 cursor-pointer hover:bg-slate-50 transition-colors">
        <div className="flex items-center justify-between gap-2 mb-1.5">
          <h3 className="text-sm font-bold text-slate-800 font-mono truncate">{tl.team.name}</h3>
          <span className="text-slate-400 shrink-0">{isExpanded ? <ChevronUp className="w-3.5 h-3.5" /> : <ChevronDown className="w-3.5 h-3.5" />}</span>
        </div>
        <div className="flex items-center gap-1.5 mb-3 flex-wrap">
          <span className="text-[10px] font-semibold px-2 py-0.5 rounded-full" style={{ background: typeStyle.bg, color: typeStyle.text }}>{tl.team.type}</span>
          {!tl.size_is_explicit && <span className="text-[9px] font-semibold px-1.5 py-0.5 rounded-full bg-amber-50 text-amber-700 border border-amber-200">no size</span>}
          <span className="text-[10px] text-slate-400 ml-auto">
            <span className="font-bold text-slate-600 font-mono">{tl.service_count}</span> svc ·{' '}
            <span className="font-bold text-slate-600 font-mono">{tl.capability_count}</span> cap
            {tl.size_is_explicit && <> · <span className="font-bold text-slate-600 font-mono">{tl.team_size}</span> ppl</>}
          </span>
        </div>
        <Speedometer level={tl.overall_level} dimensions={dims} />
      </div>
      {isExpanded && (
        <div className={cn('border-t px-4 py-4 space-y-4 bg-muted', ls.border)}>
          {insight && (
            <div className="rounded-lg p-3.5 bg-card border border-border space-y-2">
              <div className="flex items-center gap-1.5">
                <Shield className="w-3 h-3 text-primary" />
                <span className="text-[10px] font-bold text-primary uppercase tracking-wide">AI Recommendation</span>
              </div>
              <p className="text-xs text-foreground leading-relaxed">{insight.explanation}</p>
              {insight.suggestion && (
                <div className="flex items-start gap-1.5 p-2 rounded-lg bg-muted border border-border">
                  <ArrowRight className="w-3 h-3 text-primary mt-0.5 shrink-0" />
                  <p className="text-[11px] text-foreground leading-snug font-medium">{insight.suggestion}</p>
                </div>
              )}
            </div>
          )}
          <div className="space-y-2 pt-2 border-t border-slate-100">
            <div className="text-[11px] font-semibold text-slate-400 uppercase tracking-wide mb-2">Load Dimensions</div>
            {DIMENSIONS.map(d => {
              const dim = tl[d.key]
              return (
                <div key={d.abbr} className="flex items-center justify-between">
                  <span className="text-sm text-slate-700">{d.label}</span>
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-semibold font-mono">{dim.value}</span>
                    <span className={cn('text-[11px] px-1.5 py-0.5 rounded-full', levelBadge[dim.level] ?? 'bg-slate-100 text-slate-600')}>{dim.level}</span>
                    <span className="text-[10px] text-slate-400 font-mono hidden sm:inline">{d.thresholds}</span>
                  </div>
                </div>
              )
            })}
          </div>
          {tl.capabilities?.length > 0 && (
            <div>
              <div className="text-[10px] font-bold text-slate-500 uppercase tracking-wide mb-1.5">Capabilities ({tl.capabilities.length})</div>
              <div className="flex flex-wrap gap-1">
                {tl.capabilities.map(c => <span key={c} className="text-[10px] px-2 py-0.5 rounded bg-slate-100 border border-slate-200 text-slate-600">{c}</span>)}
              </div>
            </div>
          )}
          {tl.services?.length > 0 && (
            <div>
              <div className="text-[10px] font-bold text-slate-500 uppercase tracking-wide mb-1.5">Services ({tl.services.length})</div>
              <div className="flex flex-wrap gap-1">
                {tl.services.map(s => <span key={s} className="text-[10px] px-2 py-0.5 rounded bg-slate-50 border border-slate-200 text-slate-600 font-mono">{s}</span>)}
              </div>
            </div>
          )}
          {!tl.size_is_explicit && (
            <div className="rounded-lg p-2.5 bg-amber-50 border border-amber-200">
              <p className="text-[11px] text-amber-800">Missing <code className="bg-white/60 px-1 rounded border border-amber-200 font-mono text-[10px]">size:</code> — defaults to 5 people.</p>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
