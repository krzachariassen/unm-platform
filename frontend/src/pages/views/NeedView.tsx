import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { ModelRequired } from '@/components/ui/ModelRequired'
import { ContentContainer } from '@/components/ui/content-container'
import { useModel } from '@/lib/model-context'
import { useSearch, matchesQuery } from '@/lib/search-context'
import { usePageInsights } from '@/hooks/usePageInsights'
import { viewsApi } from '@/services/api'
import { LoadingState, ErrorState } from '@/components/ViewState'
import { PageHeader } from '@/components/ui/page-header'
import { StatCard } from '@/components/ui/stat-card'
import { slug } from '@/lib/slug'
import { pl } from '@/lib/format'
import { VIS_BADGE } from '@/lib/visibility-styles'
import { Users, Layers, AlertTriangle, ChevronDown, ChevronUp, Sparkles, Lightbulb } from 'lucide-react'
import type { NeedViewResponse } from '@/types/views'
import type { InsightStatus } from '@/hooks/usePageInsights'

function atRiskReason(teamSpan: number): string {
  if (teamSpan >= 3) return `Spans ${pl(teamSpan, 'team')} — coordinating delivery across this many teams introduces handoff overhead.`
  if (teamSpan > 0) return `A team in the delivery chain is under high cognitive load.`
  return `Flagged at risk — a team in the delivery chain has high cognitive load or spans too many boundaries.`
}

type NeedViewNeed = NeedViewResponse['groups'][number]['needs'][number]

function NeedCard({ nr, insights, aiStatus }: { nr: NeedViewNeed; insights: Record<string, { explanation: string; suggestion: string }>; aiStatus: InsightStatus }) {
  const [open, setOpen] = useState(false)
  const { need, capabilities } = nr
  const d = need.data
  const isMapped = d.is_mapped !== false
  const teamSpan = (d.team_span as number) ?? 0
  const aiInsight = insights[`need:${slug(need.label)}`]

  return (
    <div className="rounded-lg border border-border bg-card overflow-hidden">
      <div onClick={() => setOpen(o => !o)} className="px-3 py-2.5 cursor-pointer hover:bg-muted/50 transition-colors flex items-start gap-2">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <span className="text-sm font-semibold text-foreground">{need.label}</span>
            {d.unbacked && <span className="px-1.5 py-0.5 rounded text-[10px] font-semibold bg-amber-100 text-amber-800">Unbacked</span>}
            {d.at_risk && !d.unbacked && <span className="px-1.5 py-0.5 rounded text-[10px] font-semibold bg-red-100 text-red-700">At risk</span>}
            {!isMapped && <span className="px-1.5 py-0.5 rounded text-[10px] font-semibold bg-red-100 text-red-700">Unmapped</span>}
          </div>
          {d.outcome && <p className="text-xs text-muted-foreground mt-1 line-clamp-1">{d.outcome as string}</p>}
          {capabilities.length > 0 && (
            <div className="flex flex-wrap gap-1 mt-1.5">
              {capabilities.map(cap => {
                const badge = VIS_BADGE[cap.data.visibility ?? ''] ?? { bg: '#f1f5f9', text: '#475569' }
                return <span key={cap.id} className="px-1.5 py-0.5 rounded text-[10px] font-semibold" style={{ background: badge.bg, color: badge.text }}>{cap.label}</span>
              })}
            </div>
          )}
        </div>
        <span className="text-muted-foreground shrink-0 mt-0.5">{open ? <ChevronUp className="w-3.5 h-3.5" /> : <ChevronDown className="w-3.5 h-3.5" />}</span>
      </div>
      {open && (
        <div className="px-3 pb-3 space-y-2 border-t border-border bg-muted/30">
          <div className="pt-2 space-y-2">
            {d.outcome && <p className="text-xs text-muted-foreground leading-relaxed">{d.outcome as string}</p>}
            {d.at_risk && (
              <div className="flex gap-2 rounded-md px-2.5 py-2 bg-red-50 border border-red-200">
                <AlertTriangle className="w-3 h-3 text-red-600 shrink-0 mt-0.5" />
                <p className="text-xs text-red-700"><strong>Risk: </strong>{atRiskReason(teamSpan)}</p>
              </div>
            )}
            {((d.teams as string[]) ?? []).length > 0 && (
              <div>
                <p className="text-[10px] font-semibold text-muted-foreground uppercase tracking-wide mb-1">Teams</p>
                <div className="flex flex-wrap gap-1">
                  {(d.teams as string[]).map(t => <span key={t} className="px-1.5 py-0.5 rounded text-[10px] bg-muted text-muted-foreground">{t}</span>)}
                </div>
              </div>
            )}
            {aiStatus === 'loading' && <div className="flex items-center gap-2 text-xs text-muted-foreground"><span className="w-3 h-3 border-2 border-current border-t-transparent rounded-full animate-spin" />Loading AI insight…</div>}
            {aiInsight && (
              <div className="rounded-md px-2.5 py-2 bg-sky-50 border border-sky-200 space-y-1.5">
                <div className="flex items-center gap-1.5"><Sparkles className="w-3 h-3 text-sky-600" /><span className="text-[10px] font-semibold text-sky-600 uppercase tracking-wide">AI Insight</span></div>
                <p className="text-xs text-sky-900">{aiInsight.explanation}</p>
                <div className="flex gap-1.5"><Lightbulb className="w-3 h-3 text-sky-700 shrink-0 mt-0.5" /><p className="text-xs text-sky-800">{aiInsight.suggestion}</p></div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}

export function NeedView() {
  const { modelId } = useModel()
  const { query } = useSearch()
  const { data: viewData, isLoading, error } = useQuery({
    queryKey: ['needView', modelId],
    queryFn: () => viewsApi.getNeedView(modelId!),
    enabled: !!modelId,
  })
  const { insights, status: aiStatus } = usePageInsights('needs')

  if (isLoading) return <LoadingState message="Loading need view…" />
  if (error) return <ErrorState message={(error as Error).message} />
  if (!viewData) return null

  const totalNeeds = viewData.total_needs
  const unmapped = viewData.unmapped_count
  const allNeedItems = viewData.groups.flatMap(g => g.needs)
  const atRisk = allNeedItems.filter(nr => nr.need.data.at_risk).length

  const filtered = viewData.groups
    .map(g => ({ ...g, needs: g.needs.filter(nr => !query || matchesQuery(g.actor.label, query) || matchesQuery(nr.need.label, query) || nr.capabilities.some(c => matchesQuery(c.label, query))) }))
    .filter(g => g.needs.length > 0)

  return (
    <ModelRequired>
      <ContentContainer className="space-y-4">
        <PageHeader title="Need View" description={`${totalNeeds} needs across ${viewData.groups.length} actors`} />

        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
          <StatCard label="Total Needs" value={totalNeeds} icon={<Layers className="w-4 h-4" />} />
          <StatCard label="Actors" value={viewData.groups.length} icon={<Users className="w-4 h-4" />} />
          <StatCard label="Unmapped" value={unmapped} icon={<AlertTriangle className="w-4 h-4" />} />
          <StatCard label="At Risk" value={atRisk} icon={<AlertTriangle className="w-4 h-4" />} />
        </div>

        {filtered.map(({ actor, needs }) => (
          <div key={actor.id}>
            <div className="flex items-center gap-2 mb-2 mt-4">
              <Users className="w-3.5 h-3.5 text-muted-foreground" />
              <span className="text-sm font-semibold text-foreground">{actor.label}</span>
              <span className="text-xs text-muted-foreground">{needs.length} need{needs.length !== 1 ? 's' : ''}</span>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-2">
              {needs.map(nr => <NeedCard key={nr.need.id} nr={nr} insights={insights} aiStatus={aiStatus} />)}
            </div>
          </div>
        ))}
      </ContentContainer>
    </ModelRequired>
  )
}
