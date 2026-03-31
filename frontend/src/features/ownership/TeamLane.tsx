import { TEAM_TYPE_BADGE } from '@/lib/team-type-styles'
import { QuickAction } from '@/components/changeset/QuickAction'
import { matchesQuery } from '@/lib/search-context'
import { slug } from '@/lib/slug'
import type { OwnershipViewResponse } from '@/types/views'

type Lane = OwnershipViewResponse['lanes'][number]

interface NodeDetails { id: string; label: string; nodeType: string; data: Record<string, unknown> }

interface SvcPopoverData {
  label: string; teamLabel: string; teamType: string
  x: number; y: number
  capList: Array<{ label: string; visibility: string }>
  isHighSpan: boolean; isFromOtherTeam: boolean
}

function laneAccentGradient(isOverloaded: boolean, hasCrossTeam: boolean): string {
  if (isOverloaded) return 'linear-gradient(90deg, #ef4444 0%, #f97316 100%)'
  if (hasCrossTeam)  return 'linear-gradient(90deg, #f59e0b 0%, #fbbf24 100%)'
  return 'linear-gradient(90deg, #22c55e 0%, #4ade80 100%)'
}

export function TeamLane({ lane, query, insights, crossTeamCaps, onSelectNode, onOpenSvcPopover }: {
  lane: Lane
  query: string
  insights: Record<string, { explanation: string; suggestion: string }>
  crossTeamCaps: OwnershipViewResponse['cross_team_capabilities']
  onSelectNode: (n: NodeDetails) => void
  onOpenSvcPopover: (e: React.MouseEvent, svc: Lane['caps'][number]['services'][number], laneTeamId: string) => void
}) {
  const teamType = lane.team.data.type ?? ''
  const isOverloaded = lane.team.data.is_overloaded
  const hasCrossTeam = lane.caps.some(cg => cg.cross_team)
  const topAccent = laneAccentGradient(isOverloaded, hasCrossTeam)
  const badge = TEAM_TYPE_BADGE[teamType] ?? { bg: '#f3f4f6', text: '#374151' }
  const teamInsight = insights[`team:${slug(lane.team.label)}`]

  const visibleCaps = lane.caps.filter(cg =>
    !query || matchesQuery(cg.cap.label, query) || cg.services.some(s => matchesQuery(s.label, query))
  )
  const totalSvcCount = visibleCaps.reduce((sum, cg) => sum + cg.services.length, 0)

  return (
    <div className="rounded-2xl overflow-hidden flex flex-col border border-slate-200 bg-gradient-to-br from-white to-slate-50 shadow-sm transition-all"
      onMouseEnter={e => { (e.currentTarget as HTMLDivElement).style.transform = 'translateY(-1px)'; (e.currentTarget as HTMLDivElement).style.boxShadow = '0 12px 32px rgba(15,23,42,0.08)' }}
      onMouseLeave={e => { (e.currentTarget as HTMLDivElement).style.transform = ''; (e.currentTarget as HTMLDivElement).style.boxShadow = '' }}
    >
      <div className="h-1 w-full shrink-0" style={{ background: topAccent }} />

      {/* Team header */}
      <div className="p-5 pb-3 border-b border-slate-200 bg-gradient-to-b from-slate-50 to-white">
        <div className="flex items-start gap-2 flex-wrap mb-2"
          onMouseEnter={e => { const btn = e.currentTarget.querySelector('.qa-team') as HTMLElement; if (btn) btn.style.opacity = '1' }}
          onMouseLeave={e => { const btn = e.currentTarget.querySelector('.qa-team') as HTMLElement; if (btn) btn.style.opacity = '0.35' }}>
          <span className="font-bold text-base tracking-tight text-slate-900">{lane.team.label}</span>
          <span className="qa-team" style={{ opacity: 0.35, transition: 'opacity 0.15s' }}>
            <QuickAction size={12} options={[
              { label: 'Change team type', action: { type: 'update_team_type', team_name: lane.team.label } },
              { label: 'Update team size', action: { type: 'update_team_size', team_name: lane.team.label } },
            ]} />
          </span>
          <span className="text-[11px] font-semibold rounded-full px-2.5 py-0.5 shrink-0" style={{ background: badge.bg, color: badge.text, border: `1px solid ${badge.text}22` }}>
            {teamType || 'unknown'}
          </span>
          {isOverloaded && <span className="text-[11px] font-semibold rounded-full px-2.5 py-0.5 shrink-0 bg-orange-50 text-orange-700 border border-orange-200">overloaded</span>}
          {hasCrossTeam && <span className="text-[11px] font-semibold rounded-full px-2.5 py-0.5 shrink-0 bg-amber-50 text-amber-800 border border-amber-200">cross-team</span>}
          <button type="button" className="text-xs hover:opacity-80 ml-auto text-slate-400" onClick={() => onSelectNode({ id: lane.team.id, label: lane.team.label, nodeType: 'team', data: { ...lane.team.data, nodeType: 'team' } })}>ℹ</button>
        </div>
        {lane.team.data.description && <p className="text-sm line-clamp-2 mb-2 leading-relaxed text-slate-500">{lane.team.data.description}</p>}
        <div className="flex items-center gap-3 text-xs font-medium text-slate-400">
          <span>{visibleCaps.length} {visibleCaps.length === 1 ? 'cap' : 'caps'}</span>
          <span>{totalSvcCount} {totalSvcCount === 1 ? 'service' : 'services'}</span>
          {(lane.external_deps ?? []).length > 0 && <span>{(lane.external_deps ?? []).length} ext deps</span>}
        </div>
      </div>

      {/* AI insight */}
      {teamInsight && (
        <div className="px-5 py-3 bg-gradient-to-r from-indigo-50 to-slate-50 border-b border-slate-200">
          <p className="text-xs leading-relaxed text-slate-700">{teamInsight.explanation}</p>
          {teamInsight.suggestion && <p className="text-xs leading-relaxed mt-1.5 font-medium text-indigo-700">{teamInsight.suggestion}</p>}
        </div>
      )}

      {/* Capabilities */}
      <div className="flex-1">
        {visibleCaps.length === 0
          ? <div className="px-5 py-4 text-sm italic text-slate-400">no capabilities owned</div>
          : visibleCaps.map((cg, idx) => {
            const isLeaf = cg.cap.data.is_leaf !== false
            return (
              <div key={cg.cap.id} className="px-5 py-3"
                style={{ background: cg.cross_team ? 'linear-gradient(90deg, #fffbeb 0%, #ffffff 100%)' : '#ffffff', borderBottom: idx < visibleCaps.length - 1 ? '1px solid #f1f5f9' : 'none' }}>
                <div className="flex items-center gap-2 mb-2">
                  <button type="button" className="text-left min-w-0 flex-1"
                    onClick={() => onSelectNode({ id: cg.cap.id, label: cg.cap.label, nodeType: 'capability', data: { ...cg.cap.data, nodeType: 'capability' } })}>
                    <span className={`text-sm font-bold hover:underline ${cg.cross_team ? 'text-amber-700' : 'text-slate-900'}`}>{cg.cap.label}</span>
                  </button>
                  {!isLeaf && <span className="text-xs text-slate-300 shrink-0">parent</span>}
                  {cg.cross_team && (() => {
                    const teams = crossTeamCaps.find(ct => ct.cap_id === cg.cap.id)?.team_labels ?? []
                    return <span className="text-xs text-amber-600 shrink-0 cursor-help" title={`Cross-team: ${teams.join(', ')}`}>⚠</span>
                  })()}
                </div>
                <div className="flex flex-wrap gap-1.5">
                  {cg.services.length === 0
                    ? <span className="text-xs italic text-slate-400">{isLeaf ? 'no services' : 'groups sub-capabilities'}</span>
                    : cg.services.map(svc => {
                        const isFromOtherTeam = Boolean(svc.team_id && svc.team_id !== lane.team.id)
                        return (
                          <span key={svc.id} className="inline-flex items-center gap-0.5"
                            onMouseEnter={e => { const btn = e.currentTarget.querySelector('.qa-svc') as HTMLElement; if (btn) btn.style.opacity = '1' }}
                            onMouseLeave={e => { const btn = e.currentTarget.querySelector('.qa-svc') as HTMLElement; if (btn) btn.style.opacity = '0.35' }}>
                            <button type="button" onClick={(e) => onOpenSvcPopover(e, svc, lane.team.id)}
                              className="font-mono text-[11px] rounded-md px-2.5 py-1 cursor-pointer transition-transform hover:scale-[1.02]"
                              style={{ background: isFromOtherTeam ? '#fef3c7' : '#f1f5f9', border: isFromOtherTeam ? '1px solid #fde68a' : '1px solid #e2e8f0', color: isFromOtherTeam ? '#92400e' : '#334155' }}
                              title={isFromOtherTeam ? `Owned by ${svc.team_label}` : undefined}>
                              {svc.label}
                            </button>
                            <span className="qa-svc" style={{ opacity: 0.35, transition: 'opacity 0.15s' }}>
                              <QuickAction size={11} options={[
                                { label: `Move ${svc.label} to another team`, action: { type: 'move_service', service_name: svc.label, from_team_name: svc.team_label } },
                                { label: `Rename ${svc.label}`, action: { type: 'rename_service', service_name: svc.label } },
                              ]} />
                            </span>
                          </span>
                        )
                      })
                  }
                </div>
              </div>
            )
          })
        }

        {(lane.external_deps ?? []).length > 0 && (
          <div className="px-5 py-3 flex items-center gap-2 flex-wrap border-t border-slate-100 bg-slate-50">
            <span className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider shrink-0">External →</span>
            {(lane.external_deps ?? []).map(dep => (
              <span key={dep.id} className="font-mono text-[11px] rounded-md px-2 py-0.5 bg-slate-100 text-slate-600 border border-slate-200" title={dep.description}>
                {dep.label}
                {dep.service_count > 1 && <span className="text-slate-400 ml-1">({dep.service_count} svcs)</span>}
              </span>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

export type { SvcPopoverData }
export type { NodeDetails }
