import type { ViewEdge } from '@/types/views'
import type { ChainData } from './types'

export function computeChain(
  clickedId: string,
  clickedType: 'actor' | 'need' | 'capability' | 'ext-dep',
  { actorToNeed, needToCap, capDepEdges, extDepToCapIds, capToExtDepIds }: ChainData,
): Set<string> {
  const result = new Set<string>([clickedId])

  if (clickedType === 'ext-dep') {
    for (const capId of extDepToCapIds.get(clickedId) ?? []) result.add(capId)
    return result
  }

  // Build lookup maps
  const actorToNeedIds = new Map<string, string[]>()
  const needToActorId = new Map<string, string>()
  for (const e of actorToNeed) {
    if (!actorToNeedIds.has(e.source)) actorToNeedIds.set(e.source, [])
    actorToNeedIds.get(e.source)!.push(e.target)
    needToActorId.set(e.target, e.source)
  }
  const needToCapIds = new Map<string, string[]>()
  const capToNeedIds = new Map<string, string[]>()
  for (const e of needToCap) {
    if (!needToCapIds.has(e.source)) needToCapIds.set(e.source, [])
    needToCapIds.get(e.source)!.push(e.target)
    if (!capToNeedIds.has(e.target)) capToNeedIds.set(e.target, [])
    capToNeedIds.get(e.target)!.push(e.source)
  }
  const capDownIds = new Map<string, string[]>()
  const capUpIds = new Map<string, string[]>()
  for (const { from, to } of capDepEdges) {
    if (!capDownIds.has(from)) capDownIds.set(from, [])
    capDownIds.get(from)!.push(to)
    if (!capUpIds.has(to)) capUpIds.set(to, [])
    capUpIds.get(to)!.push(from)
  }

  function bfsDown(startId: string) {
    const q = [startId]
    if (!result.has(startId)) result.add(startId)
    while (q.length > 0) {
      const id = q.shift()!
      for (const next of capDownIds.get(id) ?? []) {
        if (!result.has(next)) { result.add(next); q.push(next) }
      }
    }
  }

  function bfsUp(startId: string) {
    const q = [startId]
    if (!result.has(startId)) result.add(startId)
    while (q.length > 0) {
      const id = q.shift()!
      for (const next of capUpIds.get(id) ?? []) {
        if (!result.has(next)) { result.add(next); q.push(next) }
      }
    }
  }

  if (clickedType === 'actor') {
    for (const nid of actorToNeedIds.get(clickedId) ?? []) {
      result.add(nid)
      for (const cid of needToCapIds.get(nid) ?? []) bfsDown(cid)
    }
  } else if (clickedType === 'need') {
    const aid = needToActorId.get(clickedId)
    if (aid) result.add(aid)
    for (const cid of needToCapIds.get(clickedId) ?? []) bfsDown(cid)
  } else {
    for (const nid of capToNeedIds.get(clickedId) ?? []) {
      result.add(nid)
      const aid = needToActorId.get(nid)
      if (aid) result.add(aid)
    }
    bfsDown(clickedId)
    bfsUp(clickedId)
  }

  for (const capId of result) {
    for (const extDepId of capToExtDepIds.get(capId) ?? []) result.add(extDepId)
  }

  return result
}

export function buildChainData(
  actorToNeed: ViewEdge[],
  needToCap: ViewEdge[],
  capDepEdges: Array<{ from: string; to: string; description?: string }>,
  extDepConns: Array<{ sourceId: string; targetId: string }>,
): ChainData {
  const extDepToCapIds = new Map<string, string[]>()
  const capToExtDepIds = new Map<string, string[]>()
  for (const conn of extDepConns) {
    if (!extDepToCapIds.has(conn.targetId)) extDepToCapIds.set(conn.targetId, [])
    extDepToCapIds.get(conn.targetId)!.push(conn.sourceId)
    if (!capToExtDepIds.has(conn.sourceId)) capToExtDepIds.set(conn.sourceId, [])
    capToExtDepIds.get(conn.sourceId)!.push(conn.targetId)
  }
  return { actorToNeed, needToCap, capDepEdges, extDepToCapIds, capToExtDepIds }
}
