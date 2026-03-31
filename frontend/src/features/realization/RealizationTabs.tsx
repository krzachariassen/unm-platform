import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { AlertTriangle, ArrowRight, Users, ChevronDown, ChevronUp } from 'lucide-react'
import { VIS_BADGE } from '@/lib/visibility-styles'
import { TEAM_TYPE_BADGE } from '@/lib/team-type-styles'
import { matchesQuery } from '@/lib/search-context'
import type { GroupedNeed } from './utils'
import type { RealizationViewResponse } from '@/types/views'
import { cn } from '@/lib/utils'

type FilterMode = 'all' | 'cross-team' | 'unbacked'

function NeedRow({ n, capToSvcTeam }: {
  n: GroupedNeed
  capToSvcTeam: Map<string, { services: string[]; teams: Array<{ label: string; type: string }> }>
}) {
  const [expanded, setExpanded] = useState(false)
  const [showAllSvc, setShowAllSvc] = useState(false)
  const svcs = showAllSvc ? n.services : n.services.slice(0, 1)
  const remaining = n.services.length - 1

  return (
    <div className={cn('border-b border-slate-50 last:border-b-0', n.isCrossTeam && 'bg-amber-50/40')}>
      <div onClick={() => setExpanded(e => !e)} className="px-3.5 py-2.5 cursor-pointer hover:bg-slate-50/60 transition-colors">
        <div className="flex items-start gap-3">
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 flex-wrap mb-1.5">
              <span className="text-sm font-semibold text-slate-800">{n.need}</span>
              {n.isUnbacked && <span className="text-[10px] font-bold text-red-700 bg-red-50 border border-red-200 rounded-full px-2 py-0.5">UNBACKED</span>}
              {n.isCrossTeam && <span className="text-[10px] font-bold text-amber-800 bg-amber-50 border border-amber-200 rounded-full px-2 py-0.5 flex items-center gap-1"><AlertTriangle className="w-2.5 h-2.5" />{n.teams.map(t => t.label).join(', ')}</span>}
            </div>
            <div className="flex items-center gap-1.5 flex-wrap">
              {n.capabilities.length > 0 ? n.capabilities.map(cap => {
                const b = VIS_BADGE[cap.visibility] ?? { bg: '#f1f5f9', text: '#475569' }
                return <span key={cap.id} className="text-[11px] font-semibold px-2 py-0.5 rounded-md" style={{ background: b.bg, color: b.text }}>{cap.label}</span>
              }) : <span className="text-[11px] italic text-red-500">no capability</span>}
              {n.capabilities.length > 0 && n.services.length > 0 && <ArrowRight className="w-3 h-3 text-slate-300 shrink-0" />}
              {svcs.map(s => <span key={s} className="text-[10px] font-mono bg-slate-100 border border-slate-200 rounded px-1.5 py-0.5 text-slate-500">{s}</span>)}
              {!showAllSvc && remaining > 0 && <button onClick={e => { e.stopPropagation(); setShowAllSvc(true) }} className="text-[11px] text-blue-500 hover:underline">+{remaining} more</button>}
              {showAllSvc && remaining > 0 && <button onClick={e => { e.stopPropagation(); setShowAllSvc(false) }} className="text-[11px] text-slate-400 hover:underline">less</button>}
              {n.services.length > 0 && n.teams.length > 0 && <ArrowRight className="w-3 h-3 text-slate-300 shrink-0" />}
              {n.teams.map(t => {
                const badge = TEAM_TYPE_BADGE[t.type] ?? { bg: '#f1f5f9', text: '#475569' }
                return <span key={t.label} className="text-[10px] font-semibold px-2 py-0.5 rounded-full" style={{ background: badge.bg, color: badge.text }}>{t.label}</span>
              })}
            </div>
          </div>
          <span className="text-slate-300 shrink-0 mt-1">{expanded ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}</span>
        </div>
      </div>
      {expanded && (
        <div className="px-3.5 pb-3 bg-slate-50/50">
          <div className="rounded-lg border border-slate-200 overflow-hidden">
            <table className="w-full text-xs">
              <thead className="bg-slate-50 border-b border-slate-200">
                <tr>{['Capability', 'Visibility', 'Services', 'Teams'].map(h => <th key={h} className="px-3 py-2 text-left text-[10px] font-bold text-slate-500 uppercase tracking-wide">{h}</th>)}</tr>
              </thead>
              <tbody>
                {n.capabilities.length === 0 ? (
                  <tr><td colSpan={4} className="px-3 py-3 text-center italic text-red-500">No capability mapped — no implementation path</td></tr>
                ) : n.capabilities.map((cap, ci) => {
                  const st = capToSvcTeam.get(cap.id)
                  const vis = VIS_BADGE[cap.visibility] ?? { bg: '#f1f5f9', text: '#475569' }
                  return (
                    <tr key={ci} className="border-t border-slate-100">
                      <td className="px-3 py-2 font-semibold text-slate-700">{cap.label}</td>
                      <td className="px-3 py-2"><span className="text-[10px] font-semibold px-2 py-0.5 rounded-full" style={{ background: vis.bg, color: vis.text }}>{cap.visibility || '—'}</span></td>
                      <td className="px-3 py-2"><div className="flex flex-wrap gap-1">{st?.services.map(s => <span key={s} className="font-mono text-[10px] bg-slate-100 border border-slate-200 rounded px-1.5 py-0.5 text-slate-500">{s}</span>) ?? <span className="italic text-slate-400">—</span>}</div></td>
                      <td className="px-3 py-2"><div className="flex flex-wrap gap-1">{st?.teams.map(t => { const b = TEAM_TYPE_BADGE[t.type] ?? { bg: '#f1f5f9', text: '#475569' }; return <span key={t.label} className="text-[10px] font-semibold px-2 py-0.5 rounded-full" style={{ background: b.bg, color: b.text }}>{t.label}</span> }) ?? <span className="italic text-slate-400">unowned</span>}</div></td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
          {n.isCrossTeam && (
            <div className="mt-2 flex items-start gap-2 rounded-lg bg-amber-50 border border-amber-200 px-3 py-2">
              <AlertTriangle className="w-3.5 h-3.5 text-amber-500 shrink-0 mt-0.5" />
              <span className="text-[11px] text-amber-800 leading-snug">Coordination across <strong>{n.teams.length} teams</strong> ({n.teams.map(t => t.label).join(', ')}) — adds handoff overhead.</span>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

export function ValueChainView({ groupedNeeds, capToSvcTeam, query }: {
  groupedNeeds: GroupedNeed[]
  capToSvcTeam: Map<string, { services: string[]; teams: Array<{ label: string; type: string }> }>
  query: string
}) {
  const [filterMode, setFilterMode] = useState<FilterMode>('all')
  const [collapsedActors, setCollapsedActors] = useState<Set<string>>(new Set())

  const filtered = groupedNeeds
    .filter(n => filterMode === 'cross-team' ? n.isCrossTeam : filterMode === 'unbacked' ? n.isUnbacked : true)
    .filter(n => !query || matchesQuery(n.actor, query) || matchesQuery(n.need, query) || n.capabilities.some(c => matchesQuery(c.label, query)) || n.services.some(s => matchesQuery(s, query)) || n.teams.some(t => matchesQuery(t.label, query)))

  const byActor = new Map<string, GroupedNeed[]>()
  for (const n of filtered) { const arr = byActor.get(n.actor) ?? []; arr.push(n); byActor.set(n.actor, arr) }

  const crossTeamCount = groupedNeeds.filter(n => n.isCrossTeam).length
  const unbackedCount = groupedNeeds.filter(n => n.isUnbacked).length

  const filterBtn = (mode: FilterMode, label: string, activeClass: string) => (
    <button onClick={() => setFilterMode(m => m === mode ? 'all' : mode)}
      className={cn('px-3 py-1 rounded-full text-xs font-semibold border transition-colors', filterMode === mode ? activeClass : 'bg-white text-slate-500 border-slate-200 hover:bg-slate-50')}>
      {label}
    </button>
  )

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2 flex-wrap">
        {filterBtn('all', `All (${groupedNeeds.length})`, 'bg-indigo-600 text-white border-transparent')}
        {crossTeamCount > 0 && filterBtn('cross-team', `Cross-team (${crossTeamCount})`, 'bg-amber-500 text-white border-transparent')}
        {unbackedCount > 0 && filterBtn('unbacked', `Unbacked (${unbackedCount})`, 'bg-red-500 text-white border-transparent')}
        <span className="ml-auto text-xs text-slate-400">{filtered.length} needs · {byActor.size} actors</span>
      </div>
      <div className="space-y-2.5">
        {Array.from(byActor.entries()).map(([actor, needs]) => (
          <div key={actor} className="rounded-xl border border-slate-200 bg-white overflow-hidden">
            <div onClick={() => setCollapsedActors(prev => { const next = new Set(prev); if (next.has(actor)) next.delete(actor); else next.add(actor); return next })}
              className="px-3.5 py-2 flex items-center gap-2.5 bg-gradient-to-r from-indigo-50 to-purple-50 border-b border-indigo-100 cursor-pointer hover:from-indigo-100/60 transition-colors">
              <span className="text-[10px] text-indigo-400">{collapsedActors.has(actor) ? '▶' : '▼'}</span>
              <div className="w-5 h-5 rounded-md bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center shrink-0"><Users className="w-3 h-3 text-white" /></div>
              <span className="text-sm font-bold text-indigo-900">{actor}</span>
              <span className="text-xs text-indigo-500">{needs.length} need{needs.length !== 1 ? 's' : ''}</span>
            </div>
            {!collapsedActors.has(actor) && needs.map(n => (
              <NeedRow key={`${actor}:${n.need}`} n={n} capToSvcTeam={capToSvcTeam} />
            ))}
          </div>
        ))}
        {filtered.length === 0 && <div className="text-center py-12 text-sm text-slate-400">No needs match the current filter.</div>}
      </div>
    </div>
  )
}

export function ServiceTableView({ viewData, query }: { viewData: RealizationViewResponse; query: string }) {
  const navigate = useNavigate()
  const rows = query
    ? viewData.service_rows.filter(row => matchesQuery(row.service.label, query) || (row.team && matchesQuery(row.team.label, query)) || row.capabilities.some(c => matchesQuery(c.label, query)))
    : viewData.service_rows
  return (
    <div className="rounded-xl border border-slate-200 overflow-hidden">
      <table className="w-full text-sm">
        <thead className="bg-slate-50 border-b border-slate-200">
          <tr>{['Service', 'Owning Team', 'Capabilities Realized', 'External Deps'].map(h => (
            <th key={h} className="px-4 py-3 text-left text-[11px] font-bold text-slate-500 uppercase tracking-wide border-r border-slate-200 last:border-r-0">{h}</th>
          ))}</tr>
        </thead>
        <tbody>
          {rows.map((row, idx) => {
            const teamBadge = TEAM_TYPE_BADGE[row.team?.data.type ?? ''] ?? { bg: '#f1f5f9', text: '#475569' }
            return (
              <tr key={row.service.id} className={cn('border-t border-slate-100 hover:bg-slate-50/60 transition-colors', idx % 2 === 1 && 'bg-slate-50/30')}>
                <td className="px-4 py-2.5"><span className="font-mono text-xs font-semibold text-slate-700">{row.service.label}</span></td>
                <td className="px-4 py-2.5">{row.team ? <span className="text-[11px] font-semibold px-2.5 py-0.5 rounded-full" style={{ background: teamBadge.bg, color: teamBadge.text }}>{row.team.label}</span> : <span className="text-xs italic text-slate-400">unowned</span>}</td>
                <td className="px-4 py-2.5">
                  <div className="flex flex-wrap gap-1">
                    {row.capabilities.length === 0 ? <span className="text-xs italic text-slate-400">none</span> : row.capabilities.map(cap => {
                      const b = VIS_BADGE[cap.data.visibility ?? ''] ?? { bg: '#f1f5f9', text: '#475569' }
                      return <button key={cap.id} onClick={() => navigate(`/capability?highlight=${encodeURIComponent(cap.label)}`)} className="text-[11px] font-semibold px-2 py-0.5 rounded-full cursor-pointer border-0" style={{ background: b.bg, color: b.text }}>{cap.label}</button>
                    })}
                  </div>
                </td>
                <td className="px-4 py-2.5"><div className="flex flex-wrap gap-1">{(row.external_deps ?? []).length === 0 ? <span className="text-xs text-slate-400">—</span> : (row.external_deps ?? []).map((dep, i) => <span key={i} className="font-mono text-[10px] bg-slate-100 border border-slate-200 rounded px-1.5 py-0.5 text-slate-500">{dep}</span>)}</div></td>
              </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}
