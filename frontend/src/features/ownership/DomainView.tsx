import { VIS_BADGE } from '@/lib/visibility-styles'
import { TEAM_TYPE_BADGE } from '@/lib/team-type-styles'
import { matchesQuery } from '@/lib/search-context'
import { LoadingState } from '@/components/ViewState'
import type { CapabilityViewResponse, OwnershipViewResponse } from '@/types/views'
import type { NodeDetails } from './TeamLane'

export function DomainView({ capViewData, viewData, query, onSelectNode, onTabSwitch, onSetSearch }: {
  capViewData: CapabilityViewResponse | null
  viewData: OwnershipViewResponse
  query: string
  onSelectNode: (n: NodeDetails) => void
  onTabSwitch: (tab: 'team') => void
  onSetSearch: (s: string) => void
}) {
  if (!capViewData) return <LoadingState />

  // Build capOwnership: cap id → team labels
  const capOwnership = new Map<string, string[]>()
  viewData.lanes.forEach(lane => {
    lane.caps.forEach(cg => {
      const existing = capOwnership.get(cg.cap.id) ?? []
      capOwnership.set(cg.cap.id, [...existing, lane.team.label])
    })
  })

  const UnownedList = ({ caps, subLabel }: { caps: Array<{ id: string; label: string; data: { visibility: string } }>; subLabel: string }) => {
    const filtered = caps.filter(c => !query || matchesQuery(c.label, query))
    if (filtered.length === 0) return null
    return (
      <div className="rounded-2xl overflow-hidden border border-red-200" style={{ background: 'linear-gradient(135deg, #fff1f2 0%, #ffffff 100%)' }}>
        <div className="h-1 w-full" style={{ background: 'linear-gradient(90deg, #ef4444 0%, #fb7185 100%)' }} />
        <div className="flex items-center gap-2 px-5 py-4 border-b border-red-200">
          <span className="text-base font-bold text-rose-900">Unowned Capabilities</span>
          <span className="text-xs font-semibold text-red-600">{subLabel}</span>
        </div>
        <div className="px-5 py-4 flex flex-wrap gap-2">
          {filtered.map(c => (
            <button key={c.id} type="button"
              onClick={() => onSelectNode({ id: c.id, label: c.label, nodeType: 'capability', data: { ...c.data, nodeType: 'capability' } })}
              className="inline-flex items-center text-[11px] font-semibold rounded-full px-3.5 py-2 bg-white border border-dashed border-red-300 text-rose-700 transition-transform hover:-translate-y-px">
              {c.label}
            </button>
          ))}
        </div>
      </div>
    )
  }

  const allGroupedIds = new Set(capViewData.parent_groups.flatMap(g => g.children))
  const ungroupedUnowned = viewData.unowned_capabilities.filter(c => !allGroupedIds.has(c.id))

  return (
    <div className="space-y-3">
      {capViewData.parent_groups.map(group => {
        const childCaps = group.children.map(id => {
          const capInfo = capViewData.capabilities.find(c => c.id === id)
          const owners = capOwnership.get(id) ?? []
          return { id, label: capInfo?.label ?? id, visibility: capInfo?.visibility ?? '', owners }
        }).filter(c => !query || matchesQuery(c.label, query))
        if (childCaps.length === 0) return null

        const allOwners = Array.from(new Set(childCaps.flatMap(c => c.owners)))
        const hasCrossTeamCap = childCaps.some(c => c.owners.length > 1)
        let groupAccent = '#22c55e'
        if (allOwners.length >= 3) groupAccent = '#ef4444'
        else if (allOwners.length === 2) groupAccent = '#f59e0b'

        return (
          <div key={group.id} className="rounded-2xl overflow-hidden border border-slate-200 bg-gradient-to-br from-white to-slate-50 shadow-sm">
            <div className="h-1 w-full" style={{ background: `linear-gradient(90deg, ${groupAccent} 0%, #6366f1 100%)` }} />
            <div className="px-5 py-4 flex items-center gap-3 flex-wrap border-b border-slate-200 bg-gradient-to-r from-slate-50 to-white">
              <span className="font-bold text-base text-slate-900">{group.label}</span>
              <span className="text-[11px] font-semibold rounded-full px-2.5 py-0.5 bg-slate-100 text-slate-500">{childCaps.length} caps</span>
              <span className="text-[11px] font-semibold rounded-full px-2.5 py-0.5 bg-slate-100 text-slate-500">{allOwners.length} {allOwners.length === 1 ? 'team' : 'teams'}</span>
              {hasCrossTeamCap && <span className="text-[11px] font-semibold rounded-full px-2.5 py-0.5 bg-amber-50 text-amber-800 border border-amber-200">cross-team</span>}
              <div className="flex gap-1.5 flex-wrap ml-auto">
                {allOwners.map(owner => {
                  const lane = viewData.lanes.find(l => l.team.label === owner)
                  const tt = lane?.team.data.type ?? ''
                  const tsBadge = TEAM_TYPE_BADGE[tt] ?? { bg: '#f3f4f6', text: '#374151' }
                  return <span key={owner} className="text-[11px] font-semibold rounded-full px-2.5 py-0.5" style={{ background: tsBadge.bg, color: tsBadge.text, border: `1px solid ${tsBadge.text}22` }}>{owner}</span>
                })}
                {allOwners.length === 0 && <span className="text-xs italic text-slate-400">unowned</span>}
              </div>
            </div>
            <div>
              {childCaps.map((cap, idx) => (
                <div key={cap.id} className="px-5 py-3 flex items-center gap-3 flex-wrap sm:flex-nowrap"
                  style={{ background: cap.owners.length > 1 ? 'linear-gradient(90deg, #fffbeb 0%, #ffffff 100%)' : '#ffffff', borderBottom: idx < childCaps.length - 1 ? '1px solid #f1f5f9' : 'none' }}>
                  <button type="button" className="text-sm font-bold flex-1 min-w-0 text-left text-blue-500 underline bg-transparent border-0 cursor-pointer p-0"
                    onClick={() => onSelectNode({ id: cap.id, label: cap.label, nodeType: 'capability', data: { visibility: cap.visibility, nodeType: 'capability' } })}>
                    {cap.label}
                  </button>
                  {cap.visibility && (() => {
                    const b = VIS_BADGE[cap.visibility] ?? { bg: '#f3f4f6', text: '#374151' }
                    return <span className="text-[11px] font-semibold rounded-full px-2.5 py-0.5 shrink-0" style={{ background: b.bg, color: b.text }}>{cap.visibility}</span>
                  })()}
                  <div className="flex gap-1.5 flex-wrap shrink-0 justify-end min-w-[120px]">
                    {cap.owners.length === 0
                      ? <span className="text-xs italic text-slate-400">unowned</span>
                      : cap.owners.map(owner => {
                          const lane = viewData.lanes.find(l => l.team.label === owner)
                          const tt = lane?.team.data.type ?? ''
                          const tsBadge = TEAM_TYPE_BADGE[tt] ?? { bg: '#f3f4f6', text: '#374151' }
                          return (
                            <button key={owner} type="button" className="text-[11px] font-semibold rounded-full px-2.5 py-0.5 border-0 cursor-pointer"
                              style={{ background: tsBadge.bg, color: tsBadge.text }}
                              onClick={() => { onTabSwitch('team'); onSetSearch(owner) }}>
                              {owner}
                            </button>
                          )
                        })
                    }
                  </div>
                </div>
              ))}
            </div>
          </div>
        )
      })}
      <UnownedList caps={ungroupedUnowned} subLabel="no team or group assigned" />
    </div>
  )
}
