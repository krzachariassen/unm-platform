import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { ModelRequired } from '@/components/ui/ModelRequired'
import { useModel } from '@/lib/model-context'
import { useSearch, matchesQuery } from '@/lib/search-context'
import { usePageInsights } from '@/hooks/usePageInsights'
import { viewsApi } from '@/services/api'
import { LoadingState, ErrorState } from '@/components/ViewState'
import { PageHeader } from '@/components/ui/page-header'
import { StatCard } from '@/components/ui/stat-card'
import { slug } from '@/lib/slug'
import { VIS_BADGE } from '@/lib/visibility-styles'
import { Users, Layers, AlertTriangle, ChevronDown, ChevronUp, Sparkles, Info, Lightbulb } from 'lucide-react'
import type { NeedViewResponse } from '@/types/views'
import type { InsightStatus } from '@/hooks/usePageInsights'

function atRiskReason(teamSpan: number): string {
  if (teamSpan >= 3) return `Spans ${teamSpan} teams — coordinating delivery across this many teams introduces handoff overhead.`
  if (teamSpan > 0) return `A team in the delivery chain is under high cognitive load.`
  return `Flagged at risk — a team in the delivery chain has high cognitive load or spans too many boundaries.`
}

type NeedViewNeed = NeedViewResponse['groups'][number]['needs'][number]

function NeedRow({ nr, isLast, insights, aiStatus }: { nr: NeedViewNeed; isLast: boolean; insights: Record<string, { explanation: string; suggestion: string }>; aiStatus: InsightStatus }) {
  const [open, setOpen] = useState(false)
  const { need, capabilities } = nr
  const d = need.data
  const isMapped = d.is_mapped !== false
  const teamSpan = (d.team_span as number) ?? 0
  const aiInsight = insights[`need:${slug(need.label)}`]
  const borderB = open || !isLast ? 'border-b border-slate-100' : ''

  return (
    <>
      <tr className={`cursor-pointer select-none transition-colors hover:bg-slate-50 ${open ? 'bg-blue-50/50' : ''}`} onClick={() => setOpen(o => !o)}>
        <td className={`px-5 py-3.5 align-top ${borderB} border-r border-slate-100`}>
          <div className="flex items-start gap-2">
            <div className="flex-1 min-w-0">
              <span className="font-semibold text-sm text-slate-800">{need.label}</span>
              {d.outcome && <div className="text-xs text-gray-500 mt-0.5 truncate max-w-sm" title={d.outcome as string}>{d.outcome as string}</div>}
              <div className="flex flex-wrap gap-1.5 mt-1.5">
                {d.unbacked && <span className="px-2 py-0.5 rounded text-[11px] font-semibold bg-amber-100 text-amber-800 cursor-help" title="No capability has services backing this need">Unbacked</span>}
                {d.at_risk && !d.unbacked && <span className="px-2 py-0.5 rounded text-[11px] font-semibold bg-red-100 text-red-700 cursor-help" title={atRiskReason(teamSpan)}>At risk · {teamSpan} team{teamSpan !== 1 ? 's' : ''}</span>}
                {!d.unbacked && !d.at_risk && teamSpan > 1 && <span className="px-2 py-0.5 rounded text-[11px] bg-slate-100 text-slate-600">{teamSpan} teams</span>}
              </div>
            </div>
            <span className="text-slate-400 mt-0.5 shrink-0">{open ? <ChevronUp className="w-3 h-3" /> : <ChevronDown className="w-3 h-3" />}</span>
          </div>
        </td>
        <td className={`px-5 py-3.5 align-top ${borderB} border-r border-slate-100`}>
          {capabilities.length === 0 ? (
            <span className="text-xs italic text-slate-400">no capabilities linked</span>
          ) : (
            <div className="flex flex-wrap gap-1.5">
              {capabilities.map(cap => {
                const badge = VIS_BADGE[cap.data.visibility ?? ''] ?? { bg: '#f1f5f9', text: '#475569' }
                return <span key={cap.id} className="px-2 py-0.5 rounded text-[11px] font-semibold" style={{ background: badge.bg, color: badge.text }}>{cap.label}</span>
              })}
            </div>
          )}
        </td>
        <td className={`px-5 py-3.5 align-top ${borderB}`}>
          {isMapped
            ? <span className="px-2 py-0.5 rounded text-[11px] font-semibold bg-green-100 text-green-700">Mapped</span>
            : <span className="px-2 py-0.5 rounded text-[11px] font-semibold bg-red-100 text-red-700">Unmapped</span>}
        </td>
      </tr>
      {open && (
        <tr>
          <td colSpan={3} className={!isLast ? 'border-b border-slate-100' : ''}>
            <div className="px-5 py-4 space-y-3 bg-slate-50 border-t border-slate-200">
              {d.outcome && <p className="text-xs leading-relaxed text-slate-600">{d.outcome as string}</p>}
              {d.at_risk && (
                <div className="flex gap-2 rounded-lg px-3 py-2.5 bg-red-50 border border-red-200">
                  <AlertTriangle className="w-3 h-3 text-red-600 shrink-0 mt-0.5" aria-label="At risk" />
                  <p className="text-xs leading-relaxed text-red-700"><strong>Why at risk: </strong>{atRiskReason(teamSpan)}</p>
                </div>
              )}
              {((d.teams as string[]) ?? []).length > 0 && (
                <div>
                  <p className="text-xs font-semibold text-slate-500 mb-1.5">Delivery teams</p>
                  <div className="flex flex-wrap gap-1.5">
                    {(d.teams as string[]).map(t => <span key={t} className="px-2 py-0.5 rounded text-[11px] bg-slate-100 text-slate-600">{t}</span>)}
                  </div>
                </div>
              )}
              {aiStatus === 'loading' && <div className="flex items-center gap-2 text-xs text-slate-400"><span className="w-3 h-3 border-2 border-current border-t-transparent rounded-full animate-spin shrink-0" />Loading AI insight…</div>}
              {aiInsight && (
                <div className="rounded-lg px-3 py-3 bg-sky-50 border border-sky-200 space-y-2">
                  <div className="flex items-center gap-1.5"><Sparkles className="w-3 h-3 text-sky-600" /><span className="text-xs font-semibold text-sky-600">AI Insight</span></div>
                  <div className="flex gap-2"><Info className="w-3 h-3 text-sky-700 shrink-0 mt-0.5" /><p className="text-xs text-sky-900">{aiInsight.explanation}</p></div>
                  <div className="flex gap-2"><Lightbulb className="w-3 h-3 text-sky-700 shrink-0 mt-0.5" /><p className="text-xs text-sky-900">{aiInsight.suggestion}</p></div>
                </div>
              )}
            </div>
          </td>
        </tr>
      )}
    </>
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
  const unmappedPct = totalNeeds > 0 ? Math.round((unmapped / totalNeeds) * 100) : 0
  const atRiskPct = totalNeeds > 0 ? Math.round((atRisk / totalNeeds) * 100) : 0

  const filtered = viewData.groups
    .map(g => ({ ...g, needs: g.needs.filter(nr => !query || matchesQuery(g.actor.label, query) || matchesQuery(nr.need.label, query) || nr.capabilities.some(c => matchesQuery(c.label, query))) }))
    .filter(g => g.needs.length > 0)

  return (
    <ModelRequired>
      <div className="space-y-6">
        <PageHeader
          title="Need View"
          description={`${totalNeeds} needs across ${viewData.groups.length} actors`}
          actions={
            <div className="flex flex-wrap gap-2">
              {Object.entries(VIS_BADGE).map(([k, s]) => (
                <span key={k} className="px-2 py-0.5 rounded text-[11px] font-semibold capitalize" style={{ background: s.bg, color: s.text }}>{k.replace(/-/g, ' ')}</span>
              ))}
            </div>
          }
        />

        <div className="grid gap-4 grid-cols-2 sm:grid-cols-4">
          <StatCard label="Total Needs" value={totalNeeds} icon={<Layers className="w-4 h-4" />} />
          <StatCard label="Actors" value={viewData.groups.length} icon={<Users className="w-4 h-4" />} />
          <StatCard label="Unmapped" value={`${unmapped} (${unmappedPct}%)`} icon={<AlertTriangle className="w-4 h-4" />} />
          <StatCard label="At Risk" value={`${atRisk} (${atRiskPct}%)`} icon={<AlertTriangle className="w-4 h-4" />} />
        </div>

        <div className="space-y-6">
          {filtered.map(({ actor, needs }) => (
            <div key={actor.id} className="rounded-lg border border-border bg-card overflow-hidden">
              <div className="h-0.5 bg-gradient-to-r from-indigo-500 via-purple-500 to-purple-300" />
              <div className="px-6 py-4 flex items-center justify-between bg-gradient-to-r from-violet-50 to-white border-b border-slate-100">
                <div>
                  <div className="font-bold text-slate-800">{actor.label}</div>
                  <div className="text-xs text-slate-400 mt-0.5">Actor · {needs.length} need{needs.length === 1 ? '' : 's'}</div>
                </div>
                <span className="px-2 py-0.5 rounded text-[11px] font-semibold bg-violet-100 text-violet-700">{needs.length} needs</span>
              </div>
              <div className="p-4 bg-gradient-to-b from-white to-slate-50">
                <div className="rounded-lg border border-slate-200 overflow-hidden">
                  <table className="w-full text-sm bg-white">
                    <thead className="bg-slate-50">
                      <tr>
                        {[['Need / Outcome', '35%'], ['Supported by capabilities', ''], ['Status', '100px']].map(([label, w]) => (
                          <th key={label} className="px-5 py-3 text-left text-[11px] font-semibold text-slate-500 uppercase tracking-wide border-b border-slate-100 border-r last:border-r-0" style={w ? { width: w } : undefined}>{label}</th>
                        ))}
                      </tr>
                    </thead>
                    <tbody>
                      {needs.length === 0 ? (
                        <tr><td colSpan={3} className="px-5 py-5 text-sm italic text-slate-400">No needs match this filter</td></tr>
                      ) : needs.map((nr, idx) => (
                        <NeedRow key={nr.need.id} nr={nr} isLast={idx === needs.length - 1} insights={insights} aiStatus={aiStatus} />
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </ModelRequired>
  )
}
