import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { AlertTriangle, ArrowRight, ChevronDown, ChevronUp } from 'lucide-react'
import { VIS_BADGE } from '@/lib/visibility-styles'
import { TEAM_TYPE_BADGE } from '@/lib/team-type-styles'
import { matchesQuery } from '@/lib/search-context'
import type { GroupedNeed } from './utils'
import type { RealizationViewResponse } from '@/types/views'
import { cn } from '@/lib/utils'

type FilterMode = 'all' | 'cross-team' | 'unbacked'

function NeedCard({ n, capToSvcTeam }: {
  n: GroupedNeed
  capToSvcTeam: Map<string, { services: string[]; teams: Array<{ label: string; type: string }> }>
}) {
  const [expanded, setExpanded] = useState(false)

  return (
    <div className="rounded-lg border border-border bg-card overflow-hidden">
      <div onClick={() => setExpanded(e => !e)} className="px-3 py-2.5 cursor-pointer hover:bg-muted/50 transition-colors">
        <div className="flex items-start gap-2">
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-1.5 flex-wrap mb-1">
              <span className="text-sm font-semibold text-foreground">{n.need}</span>
              {n.isUnbacked && <span className="text-[10px] font-semibold text-red-700 bg-red-50 border border-red-200 rounded px-1.5 py-0.5">Unbacked</span>}
              {n.isCrossTeam && (
                <span className="text-[10px] font-semibold text-amber-800 bg-amber-50 border border-amber-200 rounded px-1.5 py-0.5 flex items-center gap-1">
                  <AlertTriangle className="w-2.5 h-2.5" />{n.teams.length} teams
                </span>
              )}
            </div>
            <div className="flex items-center gap-1 flex-wrap text-[10px]">
              {n.capabilities.length > 0 ? n.capabilities.map(cap => {
                const b = VIS_BADGE[cap.visibility] ?? { bg: '#f1f5f9', text: '#475569' }
                return <span key={cap.id} className="font-semibold px-1.5 py-0.5 rounded" style={{ background: b.bg, color: b.text }}>{cap.label}</span>
              }) : <span className="italic text-red-500">no capability</span>}
              {n.capabilities.length > 0 && n.services.length > 0 && <ArrowRight className="w-3 h-3 text-muted-foreground/50 shrink-0" />}
              {n.services.slice(0, 2).map(s => <span key={s} className="font-mono bg-muted border border-border rounded px-1.5 py-0.5 text-muted-foreground">{s}</span>)}
              {n.services.length > 2 && <span className="text-muted-foreground">+{n.services.length - 2}</span>}
              {n.services.length > 0 && n.teams.length > 0 && <ArrowRight className="w-3 h-3 text-muted-foreground/50 shrink-0" />}
              {n.teams.map(t => {
                const badge = TEAM_TYPE_BADGE[t.type] ?? { bg: '#f1f5f9', text: '#475569' }
                return <span key={t.label} className="font-semibold px-1.5 py-0.5 rounded-full" style={{ background: badge.bg, color: badge.text }}>{t.label}</span>
              })}
            </div>
          </div>
          <span className="text-muted-foreground shrink-0 mt-1">{expanded ? <ChevronUp className="w-3.5 h-3.5" /> : <ChevronDown className="w-3.5 h-3.5" />}</span>
        </div>
      </div>
      {expanded && (
        <div className="border-t border-border bg-muted/30">
          <table className="w-full text-xs">
            <thead className="border-b border-border">
              <tr>{['Capability', 'Visibility', 'Services', 'Teams'].map(h => <th key={h} className="px-3 py-1.5 text-left text-[10px] font-semibold text-muted-foreground uppercase tracking-wide">{h}</th>)}</tr>
            </thead>
            <tbody>
              {n.capabilities.length === 0 ? (
                <tr><td colSpan={4} className="px-3 py-2 text-center italic text-red-500 text-[11px]">No capability mapped</td></tr>
              ) : n.capabilities.map((cap, ci) => {
                const st = capToSvcTeam.get(cap.id)
                const vis = VIS_BADGE[cap.visibility] ?? { bg: '#f1f5f9', text: '#475569' }
                return (
                  <tr key={ci} className="border-t border-border">
                    <td className="px-3 py-1.5 font-semibold text-foreground">{cap.label}</td>
                    <td className="px-3 py-1.5"><span className="text-[10px] font-semibold px-1.5 py-0.5 rounded" style={{ background: vis.bg, color: vis.text }}>{cap.visibility || '—'}</span></td>
                    <td className="px-3 py-1.5"><div className="flex flex-wrap gap-1">{st?.services.map(s => <span key={s} className="font-mono text-[10px] bg-muted border border-border rounded px-1 py-0.5 text-muted-foreground">{s}</span>) ?? <span className="italic text-muted-foreground">—</span>}</div></td>
                    <td className="px-3 py-1.5"><div className="flex flex-wrap gap-1">{st?.teams.map(t => { const b = TEAM_TYPE_BADGE[t.type] ?? { bg: '#f1f5f9', text: '#475569' }; return <span key={t.label} className="text-[10px] font-semibold px-1.5 py-0.5 rounded-full" style={{ background: b.bg, color: b.text }}>{t.label}</span> }) ?? <span className="italic text-muted-foreground">unowned</span>}</div></td>
                  </tr>
                )
              })}
            </tbody>
          </table>
          {n.isCrossTeam && (
            <div className="mx-3 mb-2 mt-1.5 flex items-start gap-2 rounded-md bg-amber-50 border border-amber-200 px-2.5 py-1.5">
              <AlertTriangle className="w-3 h-3 text-amber-500 shrink-0 mt-0.5" />
              <span className="text-[11px] text-amber-800">Spans <strong>{n.teams.length} teams</strong> ({n.teams.map(t => t.label).join(', ')})</span>
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

  const filtered = groupedNeeds
    .filter(n => filterMode === 'cross-team' ? n.isCrossTeam : filterMode === 'unbacked' ? n.isUnbacked : true)
    .filter(n => !query || matchesQuery(n.actor, query) || matchesQuery(n.need, query) || n.capabilities.some(c => matchesQuery(c.label, query)) || n.services.some(s => matchesQuery(s, query)) || n.teams.some(t => matchesQuery(t.label, query)))

  const crossTeamCount = groupedNeeds.filter(n => n.isCrossTeam).length
  const unbackedCount = groupedNeeds.filter(n => n.isUnbacked).length

  const filterBtn = (mode: FilterMode, label: string, active: string) => (
    <button onClick={() => setFilterMode(m => m === mode ? 'all' : mode)}
      className={cn('px-2.5 py-1 rounded text-xs font-medium border transition-colors', filterMode === mode ? active : 'bg-card text-muted-foreground border-border hover:bg-muted')}>
      {label}
    </button>
  )

  return (
    <div className="space-y-3">
      <div className="flex items-center gap-2 flex-wrap">
        {filterBtn('all', `All (${groupedNeeds.length})`, 'bg-foreground text-background border-transparent')}
        {crossTeamCount > 0 && filterBtn('cross-team', `Cross-team (${crossTeamCount})`, 'bg-amber-500 text-white border-transparent')}
        {unbackedCount > 0 && filterBtn('unbacked', `Unbacked (${unbackedCount})`, 'bg-red-500 text-white border-transparent')}
        <span className="ml-auto text-xs text-muted-foreground">{filtered.length} needs</span>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-3">
        {filtered.map(n => <NeedCard key={`${n.actor}:${n.need}`} n={n} capToSvcTeam={capToSvcTeam} />)}
      </div>
      {filtered.length === 0 && <div className="text-center py-8 text-sm text-muted-foreground">No needs match the current filter.</div>}
    </div>
  )
}

export function ServiceTableView({ viewData, query }: { viewData: RealizationViewResponse; query: string }) {
  const navigate = useNavigate()
  const rows = query
    ? viewData.service_rows.filter(row => matchesQuery(row.service.label, query) || (row.team && matchesQuery(row.team.label, query)) || row.capabilities.some(c => matchesQuery(c.label, query)))
    : viewData.service_rows

  return (
    <div className="rounded-lg border border-border overflow-hidden">
      <table className="w-full text-sm">
        <thead className="bg-muted border-b border-border">
          <tr>{['Service', 'Owning Team', 'Capabilities', 'Ext. Deps'].map(h => (
            <th key={h} className="px-3 py-2 text-left text-[10px] font-semibold text-muted-foreground uppercase tracking-wide">{h}</th>
          ))}</tr>
        </thead>
        <tbody>
          {rows.map((row, idx) => {
            const teamBadge = TEAM_TYPE_BADGE[row.team?.data.type ?? ''] ?? { bg: '#f1f5f9', text: '#475569' }
            return (
              <tr key={row.service.id} className={cn('border-t border-border hover:bg-muted/50 transition-colors', idx % 2 === 1 && 'bg-muted/20')}>
                <td className="px-3 py-2"><span className="font-mono text-xs font-semibold text-foreground">{row.service.label}</span></td>
                <td className="px-3 py-2">{row.team ? <span className="text-[10px] font-semibold px-2 py-0.5 rounded-full" style={{ background: teamBadge.bg, color: teamBadge.text }}>{row.team.label}</span> : <span className="text-xs italic text-muted-foreground">unowned</span>}</td>
                <td className="px-3 py-2">
                  <div className="flex flex-wrap gap-1">
                    {row.capabilities.length === 0 ? <span className="text-xs italic text-muted-foreground">none</span> : row.capabilities.map(cap => {
                      const b = VIS_BADGE[cap.data.visibility ?? ''] ?? { bg: '#f1f5f9', text: '#475569' }
                      return <button key={cap.id} onClick={() => navigate(`/capability?highlight=${encodeURIComponent(cap.label)}`)} className="text-[10px] font-semibold px-1.5 py-0.5 rounded cursor-pointer border-0" style={{ background: b.bg, color: b.text }}>{cap.label}</button>
                    })}
                  </div>
                </td>
                <td className="px-3 py-2"><div className="flex flex-wrap gap-1">{(row.external_deps ?? []).length === 0 ? <span className="text-xs text-muted-foreground">—</span> : (row.external_deps ?? []).map((dep, i) => <span key={i} className="font-mono text-[10px] bg-muted border border-border rounded px-1 py-0.5 text-muted-foreground">{dep}</span>)}</div></td>
              </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}
