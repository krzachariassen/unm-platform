import type { ViewNode, ViewEdge, UNMMapExtDep } from '@/types/views'
import {
  PAD_X, ACTOR_Y, ACTOR_W, ACTOR_H, NEED_Y, NEED_W, NEED_H, NEED_GAP,
  ACTOR_SECTION_GAP, VIS_FIRST_Y, CAP_W, CAP_GAP, CAP_BAND_PAD, MIN_BAND_H,
  EXT_DEP_W, EXT_DEP_H, EXT_DEP_GAP, VIS_ORDER, VIS,
} from './constants'
import type { PNode, Conn, ActorGroup, BandInfo, LayoutResult, SvcInfo } from './types'

export function capCardHeight(svcsCount: number, labelLen = 20): number {
  const charsPerLine = 22
  const labelLines = Math.max(1, Math.ceil(labelLen / charsPerLine))
  const headerH = 7 + labelLines * 13 + 2
  const teamLineH = 12
  const svcLineH = 14
  const padding = 8
  const MAX_SHOWN = 3
  const shown = Math.min(svcsCount, MAX_SHOWN)
  const extraLine = svcsCount > MAX_SHOWN ? svcLineH : 0
  return headerH + teamLineH + (shown > 0 ? padding + shown * svcLineH + extraLine : 0) + padding
}

export function buildLayout(
  actors: ViewNode[],
  needs: ViewNode[],
  caps: ViewNode[],
  actorToNeed: ViewEdge[],
  needToCap: ViewEdge[],
  capDepEdges: Array<{ from: string; to: string; description?: string }>,
  extDeps: UNMMapExtDep[],
): LayoutResult {
  const actorNeedIds = new Map<string, string[]>()
  for (const e of actorToNeed) {
    if (!actorNeedIds.has(e.source)) actorNeedIds.set(e.source, [])
    actorNeedIds.get(e.source)!.push(e.target)
  }
  const needCapIds = new Map<string, string[]>()
  for (const e of needToCap) {
    if (!needCapIds.has(e.source)) needCapIds.set(e.source, [])
    needCapIds.get(e.source)!.push(e.target)
  }

  const sortedActors = [...actors].sort(
    (a, b) => (actorNeedIds.get(b.id)?.length ?? 0) - (actorNeedIds.get(a.id)?.length ?? 0),
  )
  const actorById = new Map(actors.map(a => [a.id, a]))

  const needX = new Map<string, number>()
  const actorGroups: ActorGroup[] = []
  let cursor = PAD_X

  for (const actor of sortedActors) {
    const myNeedIds = actorNeedIds.get(actor.id) ?? []
    const myNeeds = myNeedIds.map(id => needs.find(n => n.id === id)).filter(Boolean) as ViewNode[]
    const secStart = cursor
    for (const need of myNeeds) {
      needX.set(need.id, cursor)
      cursor += NEED_W + NEED_GAP
    }
    const secEnd = cursor - NEED_GAP
    actorGroups.push({
      id: actor.id, label: actor.label,
      centerX: (secStart + secEnd) / 2,
      secStart, secEnd,
      description: actorById.get(actor.id)?.data.description as string ?? '',
    })
    cursor += ACTOR_SECTION_GAP
  }

  const needsWidth = Math.max(cursor - ACTOR_SECTION_GAP + PAD_X, 900)

  const capCentroid = new Map<string, number>()
  for (const cap of caps) {
    const xs: number[] = []
    for (const [nid, cids] of needCapIds) {
      if (cids.includes(cap.id)) {
        const x = needX.get(nid)
        if (x !== undefined) xs.push(x + NEED_W / 2)
      }
    }
    capCentroid.set(cap.id, xs.length > 0 ? xs.reduce((a, b) => a + b) / xs.length : needsWidth / 2)
  }

  const capToTeam = new Map<string, { label: string; type: string }>()
  const capToSvcs = new Map<string, SvcInfo[]>()
  for (const cap of caps) {
    if (cap.data.team_label) {
      capToTeam.set(cap.id, { label: cap.data.team_label as string, type: (cap.data.team_type as string) ?? '' })
    }
    const svcs = cap.data.services as Array<{ id: string; label: string; team_name?: string }> | undefined
    if (svcs && svcs.length > 0) {
      capToSvcs.set(cap.id, svcs.map(s => ({ id: s.id, label: s.label, teamName: s.team_name ?? '' })))
    }
  }

  const bandMaxH: Record<string, number> = {}
  for (const vis of VIS_ORDER) {
    bandMaxH[vis] = MIN_BAND_H
    const visCaps = caps.filter(c => ((c.data.visibility as string) || 'domain') === vis)
    for (const cap of visCaps) {
      const svcs = capToSvcs.get(cap.id) ?? []
      const h = capCardHeight(svcs.length, cap.label.length) + CAP_BAND_PAD * 2
      if (h > bandMaxH[vis]) bandMaxH[vis] = h
    }
  }

  const bands: BandInfo[] = []
  let bandY = VIS_FIRST_Y
  for (const vis of VIS_ORDER) {
    bands.push({ vis, y: bandY, h: bandMaxH[vis] })
    bandY += bandMaxH[vis]
  }
  const bandByVis = Object.fromEntries(bands.map(b => [b.vis, b]))

  const capX = new Map<string, number>()
  let maxCapRight = needsWidth
  for (const vis of VIS_ORDER) {
    const visCaps = caps
      .filter(c => ((c.data.visibility as string) || 'domain') === vis)
      .sort((a, b) => (capCentroid.get(a.id) ?? 0) - (capCentroid.get(b.id) ?? 0))
    let capCursor = PAD_X
    for (const cap of visCaps) {
      const ideal = Math.max(capCursor, (capCentroid.get(cap.id) ?? capCursor) - CAP_W / 2)
      capX.set(cap.id, ideal)
      capCursor = ideal + CAP_W + CAP_GAP
      if (ideal + CAP_W + PAD_X > maxCapRight) maxCapRight = ideal + CAP_W + PAD_X
    }
  }
  const canvasWidth = maxCapRight

  const pnodes: PNode[] = []

  for (const g of actorGroups) {
    pnodes.push({
      id: g.id, label: g.label, type: 'actor',
      x: g.centerX - ACTOR_W / 2, y: ACTOR_Y, w: ACTOR_W, h: ACTOR_H,
      description: g.description,
    })
  }

  for (const need of needs) {
    const x = needX.get(need.id)
    if (x === undefined) continue
    pnodes.push({
      id: need.id, label: need.label, type: 'need',
      x, y: NEED_Y, w: NEED_W, h: NEED_H,
      isMapped: (need.data.is_mapped as boolean) !== false,
      outcome: (need.data.outcome as string) ?? '',
    })
  }

  for (const cap of caps) {
    const x = capX.get(cap.id)
    if (x === undefined) continue
    const vis = (cap.data.visibility as string) || 'domain'
    const band = bandByVis[vis] ?? bandByVis['infrastructure']
    const svcs = capToSvcs.get(cap.id) ?? []
    const cardH = capCardHeight(svcs.length, cap.label.length)
    const uniqueTeams = new Set(svcs.map(s => s.teamName).filter(Boolean))
    pnodes.push({
      id: cap.id, label: cap.label, type: 'capability',
      x, y: band.y + CAP_BAND_PAD,
      w: CAP_W, h: cardH, vis,
      team: capToTeam.get(cap.id),
      isFragmented: cap.data.is_fragmented as boolean,
      crossTeam: uniqueTeams.size > 1,
      svcs,
      description: (cap.data.description as string) ?? '',
    })
  }

  const lastBand = bands[bands.length - 1]
  const extDepY = lastBand.y + lastBand.h + 60
  const extDepConns: Conn[] = []

  const svcToCapIds = new Map<string, string[]>()
  for (const cap of caps) {
    const svcs = cap.data.services as Array<{ id: string }> | undefined
    if (svcs) {
      for (const s of svcs) {
        if (!svcToCapIds.has(s.id)) svcToCapIds.set(s.id, [])
        svcToCapIds.get(s.id)!.push(cap.id)
      }
    }
  }

  let extDepCursor = PAD_X
  for (const dep of extDeps) {
    const depId = `ext-dep:${dep.id}`
    const depX = extDepCursor
    pnodes.push({
      id: depId, label: dep.name, type: 'ext-dep',
      x: depX, y: extDepY, w: EXT_DEP_W, h: EXT_DEP_H,
      extDep: dep,
    })

    const linkedCapIds = new Set<string>()
    for (const svcName of dep.services) {
      for (const capId of svcToCapIds.get(svcName) ?? []) linkedCapIds.add(capId)
    }
    for (const cap of caps) {
      const capSvcs = cap.data.services as Array<{ id: string; label: string }> | undefined
      if (capSvcs) {
        for (const s of capSvcs) {
          if (dep.services.includes(s.label) || dep.services.includes(s.id)) linkedCapIds.add(cap.id)
        }
      }
    }

    const nodePos = new Map(pnodes.map(n => [n.id, n]))
    for (const capId of linkedCapIds) {
      const capNode = nodePos.get(capId)
      if (capNode) {
        extDepConns.push({
          x1: capNode.x + capNode.w / 2, y1: capNode.y + capNode.h,
          x2: depX + EXT_DEP_W / 2, y2: extDepY,
          color: dep.is_critical ? '#ef4444' : dep.is_warning ? '#f59e0b' : '#94a3b8',
          dashed: true, sourceId: capId, targetId: depId, edgeType: 'ext-dep',
        })
      }
    }
    extDepCursor += EXT_DEP_W + EXT_DEP_GAP
  }

  const canvasH = extDeps.length > 0 ? extDepY + EXT_DEP_H + 40 : lastBand.y + lastBand.h + 64

  const nodePos = new Map(pnodes.map(n => [n.id, n]))
  const conns: Conn[] = []

  for (const e of actorToNeed) {
    const src = nodePos.get(e.source); const tgt = nodePos.get(e.target)
    if (!src || !tgt) continue
    conns.push({ x1: src.x + src.w / 2, y1: src.y + src.h, x2: tgt.x + tgt.w / 2, y2: tgt.y, color: '#3b82f6', sourceId: e.source, targetId: e.target, edgeType: 'actor-need' })
  }

  for (const e of needToCap) {
    const src = nodePos.get(e.source); const tgt = nodePos.get(e.target)
    if (!src || !tgt) continue
    const vis = tgt.vis ?? 'domain'
    conns.push({ x1: src.x + src.w / 2, y1: src.y + src.h, x2: tgt.x + tgt.w / 2, y2: tgt.y, color: VIS[vis]?.line ?? '#94a3b8', sourceId: e.source, targetId: e.target, edgeType: 'need-capability', description: e.description })
  }

  const seen = new Set<string>()
  const depConns: Conn[] = []
  for (const { from, to, description } of capDepEdges) {
    const key = `${from}→${to}`
    if (seen.has(key)) continue; seen.add(key)
    const src = nodePos.get(from); const tgt = nodePos.get(to)
    if (!src || !tgt || src.id === tgt.id) continue
    const CAP_H_CHECK = tgt.h ?? 96
    if (src.y > tgt.y + CAP_H_CHECK) continue
    const vis = src.vis ?? 'foundational'
    depConns.push({ x1: src.x + src.w / 2, y1: src.y + src.h, x2: tgt.x + tgt.w / 2, y2: tgt.y, color: VIS[vis]?.line ?? '#94a3b8', dashed: true, sourceId: from, targetId: to, edgeType: 'cap-dep', description })
  }

  return { pnodes, conns, depConns, extDepConns, canvasWidth, canvasH, actorGroups, nodePos, bands }
}
