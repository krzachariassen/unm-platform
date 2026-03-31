import type { RealizationViewResponse, NeedViewResponse } from '@/types/views'

export interface GroupedNeed {
  actor: string
  need: string
  capabilities: Array<{ id: string; label: string; visibility: string }>
  services: string[]
  teams: Array<{ label: string; type: string }>
  isCrossTeam: boolean
  isUnbacked: boolean
}

export function buildCapToSvcTeam(viewData: RealizationViewResponse) {
  const map = new Map<string, { services: string[]; teams: Array<{ label: string; type: string }> }>()
  for (const row of viewData.service_rows) {
    for (const cap of row.capabilities ?? []) {
      const existing = map.get(cap.id) ?? { services: [], teams: [] }
      if (!existing.services.includes(row.service.label)) existing.services.push(row.service.label)
      if (row.team && !existing.teams.some(t => t.label === row.team!.label)) {
        existing.teams.push({ label: row.team.label, type: row.team.data.type })
      }
      map.set(cap.id, existing)
    }
  }
  return map
}

export function buildCapVisibility(viewData: RealizationViewResponse) {
  const map = new Map<string, string>()
  for (const row of viewData.service_rows) {
    for (const cap of row.capabilities ?? []) map.set(cap.id, cap.data.visibility ?? '')
  }
  return map
}

export function buildGroupedNeeds(
  needData: NeedViewResponse,
  capToSvcTeam: ReturnType<typeof buildCapToSvcTeam>,
  capVisibility: ReturnType<typeof buildCapVisibility>,
): GroupedNeed[] {
  return needData.groups.flatMap(group =>
    group.needs.map(needItem => {
      const caps = needItem.capabilities.map(cap => ({
        id: cap.id, label: cap.label, visibility: capVisibility.get(cap.id) ?? '',
      }))
      const allServices = new Set<string>()
      const teamMap = new Map<string, string>()
      for (const cap of needItem.capabilities) {
        const st = capToSvcTeam.get(cap.id)
        if (st) { st.services.forEach(s => allServices.add(s)); st.teams.forEach(t => teamMap.set(t.label, t.type)) }
      }
      const teams = Array.from(teamMap.entries()).map(([label, type]) => ({ label, type }))
      return { actor: group.actor.label, need: needItem.need.label, caps, capabilities: caps, services: Array.from(allServices), teams, isCrossTeam: teams.length > 1, isUnbacked: caps.length === 0 }
    })
  )
}
