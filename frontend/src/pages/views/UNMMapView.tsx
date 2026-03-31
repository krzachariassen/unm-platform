import { useEffect, useState, useCallback, useRef, useMemo } from 'react'
import { ZoomIn, ZoomOut, Maximize2 } from 'lucide-react'
import { api, type ViewNode, type ViewEdge, type ChangeAction, type UNMMapExtDep } from '@/lib/api'
import { useRequireModel } from '@/lib/model-context'
import { ModelRequired } from '@/components/ui/ModelRequired'
import { usePageInsights } from '@/hooks/usePageInsights'
import { LoadingState, ErrorState } from '@/components/ViewState'
import { slug } from '@/lib/slug'
import { useChangeset } from '@/lib/changeset-context'

// ─── Layout constants ──────────────────────────────────────────────────────────
const PAD_X = 60
const ACTOR_Y = 52;  const ACTOR_W = 138; const ACTOR_H = 46
const NEED_Y  = 188; const NEED_W  = 150; const NEED_H  = 52; const NEED_GAP = 16
const ACTOR_SECTION_GAP = 48
const CAP_W = 172; const CAP_H = 96; const CAP_GAP = 14; const CAP_BAND_PAD = 20
const VIS_ORDER = ['user-facing', 'domain', 'foundational', 'infrastructure'] as const
const VIS_FIRST_Y = 340
const MIN_BAND_H = 130
const EXT_DEP_W = 150; const EXT_DEP_H = 40; const EXT_DEP_GAP = 14

const VIS: Record<string, {
  nodeBg: string; border: string; text: string; bandBg: string; label: string; line: string
}> = {
  'user-facing':    { nodeBg: '#fffbeb', border: '#d97706', text: '#92400e', bandBg: 'rgba(217,119,6,0.04)',   label: 'User-facing',    line: '#fbbf24' },
  'domain':         { nodeBg: '#f5f3ff', border: '#7c3aed', text: '#5b21b6', bandBg: 'rgba(124,58,237,0.04)',  label: 'Domain',         line: '#a78bfa' },
  'foundational':   { nodeBg: '#f0fdf4', border: '#059669', text: '#065f46', bandBg: 'rgba(5,150,105,0.04)',   label: 'Foundational',   line: '#6ee7b7' },
  'infrastructure': { nodeBg: '#f8fafc', border: '#94a3b8', text: '#475569', bandBg: 'rgba(148,163,184,0.06)', label: 'Infrastructure', line: '#94a3b8' },
}

function teamColor(name: string): string {
  let h = 0
  for (let i = 0; i < name.length; i++) h = (h * 31 + name.charCodeAt(i)) & 0x7fffffff
  const palette = ['#3b82f6', '#8b5cf6', '#06b6d4', '#10b981', '#f59e0b', '#ef4444', '#ec4899', '#14b8a6', '#f97316', '#6366f1']
  return palette[h % palette.length]
}

// ─── Data types ────────────────────────────────────────────────────────────────
interface SvcInfo { id: string; label: string; teamName: string }

interface PNode {
  id: string; label: string; type: 'actor' | 'need' | 'capability' | 'ext-dep'
  x: number; y: number; w: number; h: number
  vis?: string; team?: { label: string; type: string }
  isMapped?: boolean; isFragmented?: boolean; crossTeam?: boolean
  svcs?: SvcInfo[]
  description?: string; outcome?: string
  extDep?: UNMMapExtDep
}
interface Conn {
  x1: number; y1: number; x2: number; y2: number
  color: string; dashed?: boolean
  sourceId: string; targetId: string; edgeType: string
  description?: string
}
interface ActorGroup { id: string; label: string; centerX: number; secStart: number; secEnd: number; description?: string }
interface BandInfo { vis: string; y: number; h: number }

interface ChainData {
  actorToNeed: ViewEdge[]
  needToCap: ViewEdge[]
  capDepEdges: Array<{ from: string; to: string; description?: string }>
  extDepToCapIds: Map<string, string[]>
  capToExtDepIds: Map<string, string[]>
}

// ─── Chain computation ────────────────────────────────────────────────────────
function computeChain(
  clickedId: string,
  clickedType: 'actor' | 'need' | 'capability' | 'ext-dep',
  { actorToNeed, needToCap, capDepEdges, extDepToCapIds, capToExtDepIds }: ChainData,
): Set<string> {
  const result = new Set<string>([clickedId])

  if (clickedType === 'ext-dep') {
    for (const capId of extDepToCapIds.get(clickedId) ?? []) {
      result.add(capId)
    }
    return result
  }

  const actorToNeedIds = new Map<string, string[]>()
  const needToCapIds   = new Map<string, string[]>()
  const capToNeedIds   = new Map<string, string[]>()
  const needToActorId  = new Map<string, string>()
  const capDepsDown    = new Map<string, string[]>()
  const capDepsUp      = new Map<string, string[]>()

  for (const e of actorToNeed) {
    if (!actorToNeedIds.has(e.source)) actorToNeedIds.set(e.source, [])
    actorToNeedIds.get(e.source)!.push(e.target)
    needToActorId.set(e.target, e.source)
  }
  for (const e of needToCap) {
    if (!needToCapIds.has(e.source)) needToCapIds.set(e.source, [])
    needToCapIds.get(e.source)!.push(e.target)
    if (!capToNeedIds.has(e.target)) capToNeedIds.set(e.target, [])
    capToNeedIds.get(e.target)!.push(e.source)
  }
  for (const { from, to } of capDepEdges) {
    if (!capDepsDown.has(from)) capDepsDown.set(from, [])
    capDepsDown.get(from)!.push(to)
    if (!capDepsUp.has(to)) capDepsUp.set(to, [])
    capDepsUp.get(to)!.push(from)
  }

  const bfsDown = (startId: string) => {
    const q = [startId]; const vis = new Set<string>()
    while (q.length) {
      const id = q.shift()!
      if (vis.has(id)) continue; vis.add(id); result.add(id)
      for (const dep of capDepsDown.get(id) ?? []) q.push(dep)
    }
  }
  const bfsUp = (startId: string) => {
    const q = [startId]; const vis = new Set<string>()
    while (q.length) {
      const id = q.shift()!
      if (vis.has(id)) continue; vis.add(id); result.add(id)
      for (const dep of capDepsUp.get(id) ?? []) q.push(dep)
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

  // Include external deps linked to any capability in the chain
  for (const capId of result) {
    for (const extDepId of capToExtDepIds.get(capId) ?? []) {
      result.add(extDepId)
    }
  }

  return result
}

// ─── Panel ───────────────────────────────────────────────────────────────────
interface PanelField { label: string; value: string }
interface PanelItem {
  title: string
  subtitle?: string
  badge?: { text: string; color: string }
  fields: PanelField[]
}

// ─── Compute card height based on services ──────────────────────────────────
function capCardHeight(svcsCount: number, labelLen = 20): number {
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

// ─── Layout engine ────────────────────────────────────────────────────────────
function buildLayout(
  actors: ViewNode[], needs: ViewNode[], caps: ViewNode[],
  actorToNeed: ViewEdge[], needToCap: ViewEdge[],
  capDepEdges: Array<{ from: string; to: string; description?: string }>,
  extDeps: UNMMapExtDep[],
) {
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
    actorGroups.push({ id: actor.id, label: actor.label, centerX: (secStart + secEnd) / 2, secStart, secEnd,
      description: actorById.get(actor.id)?.data.description as string ?? '' })
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

  // Build service info and team info from node data
  const capToTeam = new Map<string, { label: string; type: string }>()
  const capToSvcs = new Map<string, SvcInfo[]>()
  for (const cap of caps) {
    if (cap.data.team_label) {
      capToTeam.set(cap.id, { label: cap.data.team_label as string, type: cap.data.team_type as string ?? '' })
    }
    const svcs = cap.data.services as Array<{ id: string; label: string; team_name?: string }> | undefined
    if (svcs && svcs.length > 0) {
      capToSvcs.set(cap.id, svcs.map(s => ({ id: s.id, label: s.label, teamName: s.team_name ?? '' })))
    }
  }

  // D1: Compute dynamic band heights based on tallest card in each band
  const bandMaxH: Record<string, number> = {}
  for (const vis of VIS_ORDER) {
    bandMaxH[vis] = MIN_BAND_H
    const visCaps = caps.filter(c => ((c.data.visibility as string) ?? 'foundational') === vis)
    for (const cap of visCaps) {
      const svcs = capToSvcs.get(cap.id) ?? []
      const h = capCardHeight(svcs.length, cap.label.length) + CAP_BAND_PAD * 2
      if (h > bandMaxH[vis]) bandMaxH[vis] = h
    }
  }

  // Compute cumulative band Y positions
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
      .filter(c => ((c.data.visibility as string) ?? 'foundational') === vis)
      .sort((a, b) => (capCentroid.get(a.id) ?? 0) - (capCentroid.get(b.id) ?? 0))
    let capCursor = PAD_X
    for (const cap of visCaps) {
      const ideal = Math.max(capCursor, (capCentroid.get(cap.id) ?? capCursor) - CAP_W / 2)
      capX.set(cap.id, ideal)
      capCursor = ideal + CAP_W + CAP_GAP
      if (ideal + CAP_W + PAD_X > maxCapRight) maxCapRight = ideal + CAP_W + PAD_X
    }
  }
  // Expand canvas to fit all caps (needs layout may be narrower than cap layout)
  const canvasWidth = maxCapRight

  const pnodes: PNode[] = []

  for (const g of actorGroups) {
    pnodes.push({ id: g.id, label: g.label, type: 'actor', x: g.centerX - ACTOR_W / 2, y: ACTOR_Y, w: ACTOR_W, h: ACTOR_H,
      description: g.description })
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
    const vis = (cap.data.visibility as string) ?? 'foundational'
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

  // D3: Place external deps below the last band, linked to capabilities
  const lastBand = bands[bands.length - 1]
  const extDepY = lastBand.y + lastBand.h + 60
  const extDepConns: Conn[] = []

  // Build service→capability mapping
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

    // Link to capabilities that use this ext dep via its services
    const linkedCapIds = new Set<string>()
    for (const svcName of dep.services) {
      for (const capId of svcToCapIds.get(svcName) ?? []) {
        linkedCapIds.add(capId)
      }
    }
    // Also try matching by service name in the cap's realizing services
    for (const cap of caps) {
      const capSvcs = cap.data.services as Array<{ id: string; label: string }> | undefined
      if (capSvcs) {
        for (const s of capSvcs) {
          if (dep.services.includes(s.label) || dep.services.includes(s.id)) {
            linkedCapIds.add(cap.id)
          }
        }
      }
    }

    for (const capId of linkedCapIds) {
      const capNode = pnodes.find(n => n.id === capId)
      if (capNode) {
        extDepConns.push({
          x1: capNode.x + capNode.w / 2,
          y1: capNode.y + capNode.h,
          x2: depX + EXT_DEP_W / 2,
          y2: extDepY,
          color: dep.is_critical ? '#ef4444' : dep.is_warning ? '#f59e0b' : '#94a3b8',
          dashed: true,
          sourceId: capId,
          targetId: depId,
          edgeType: 'ext-dep',
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
    if (src.y > tgt.y + (tgt.h || CAP_H)) continue
    const vis = src.vis ?? 'foundational'
    depConns.push({ x1: src.x + src.w / 2, y1: src.y + src.h, x2: tgt.x + tgt.w / 2, y2: tgt.y, color: VIS[vis]?.line ?? '#94a3b8', dashed: true, sourceId: from, targetId: to, edgeType: 'cap-dep', description })
  }

  return { pnodes, conns, depConns, extDepConns, canvasWidth, canvasH, actorGroups, nodePos, bands }
}

// ─── Component ────────────────────────────────────────────────────────────────
export function UNMMapView() {
  const { modelId, isHydrating } = useRequireModel()
  const { isEditMode, actions, addAction, enterEditMode, refreshKey } = useChangeset()
  const [loading, setLoading] = useState(true)
  const [error, setError]     = useState<string | null>(null)
  const [layout, setLayout]   = useState<ReturnType<typeof buildLayout> | null>(null)
  const [chainData, setChainData] = useState<ChainData | null>(null)
  const [panel, setPanel]     = useState<PanelItem | null>(null)
  const [highlight, setHighlight] = useState<Set<string> | null>(null)
  const [zoom, setZoom] = useState(1)
  const [stagedCaps, setStagedCaps] = useState<Set<string>>(new Set())
  const [teams, setTeams] = useState<string[]>([])
  const [editState, setEditState] = useState<{
    capLabel: string; description: string; visibility: string; teamName: string
    origDescription: string; origVisibility: string; origTeam: string
    svcs: SvcInfo[]
  } | null>(null)
  const containerRef = useRef<HTMLDivElement>(null)
  const { insights } = usePageInsights('dashboard')

  const pendingCapNames = useMemo(() => {
    const names = new Set<string>()
    for (const a of actions) {
      const an = a as unknown as Record<string, unknown>
      if (typeof an.capability_name === 'string') names.add(an.capability_name)
    }
    return names
  }, [actions])

  const isDragging = useRef(false)
  const dragStart = useRef({ x: 0, y: 0, scrollLeft: 0, scrollTop: 0 })
  const hasDragged = useRef(false)

  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    if (e.button !== 0) return
    const el = containerRef.current
    if (!el) return
    isDragging.current = true
    hasDragged.current = false
    dragStart.current = { x: e.clientX, y: e.clientY, scrollLeft: el.scrollLeft, scrollTop: el.scrollTop }
    el.style.cursor = 'grabbing'
    el.style.userSelect = 'none'
  }, [])

  const handleMouseMove = useCallback((e: React.MouseEvent) => {
    if (!isDragging.current) return
    const el = containerRef.current
    if (!el) return
    const dx = e.clientX - dragStart.current.x
    const dy = e.clientY - dragStart.current.y
    if (Math.abs(dx) > 3 || Math.abs(dy) > 3) hasDragged.current = true
    el.scrollLeft = dragStart.current.scrollLeft - dx
    el.scrollTop = dragStart.current.scrollTop - dy
  }, [])

  const handleMouseUp = useCallback(() => {
    isDragging.current = false
    const el = containerRef.current
    if (el) {
      el.style.cursor = 'grab'
      el.style.userSelect = ''
    }
  }, [])

  const clearSelection = useCallback(() => {
    if (hasDragged.current) return
    setPanel(null)
    setHighlight(null)
  }, [])

  const loadMap = useCallback(() => {
    if (isHydrating || !modelId) return
    setLoading(true)
    api.getUNMMapView(modelId)
      .then((data) => {
        const actors = data.nodes.filter(n => n.type === 'actor')
        const needs  = data.nodes.filter(n => n.type === 'need')
        const caps   = data.nodes.filter(n => n.type === 'capability')
        const actorToNeed = data.edges.filter(e => e.label === 'has need')
        const needToCap   = data.edges.filter(e => e.label === 'supportedBy')

        const capDepEdges: Array<{ from: string; to: string; description?: string }> = []
        for (const e of data.edges) {
          if (e.label === 'dependsOn') {
            capDepEdges.push({ from: e.source, to: e.target, description: e.description })
          }
        }

        const builtLayout = buildLayout(actors, needs, caps, actorToNeed, needToCap, capDepEdges, data.external_deps ?? [])

        const extDepToCapIds = new Map<string, string[]>()
        const capToExtDepIds = new Map<string, string[]>()
        for (const conn of builtLayout.extDepConns) {
          if (!extDepToCapIds.has(conn.targetId)) extDepToCapIds.set(conn.targetId, [])
          extDepToCapIds.get(conn.targetId)!.push(conn.sourceId)
          if (!capToExtDepIds.has(conn.sourceId)) capToExtDepIds.set(conn.sourceId, [])
          capToExtDepIds.get(conn.sourceId)!.push(conn.targetId)
        }

        setLayout(builtLayout)
        setChainData({ actorToNeed, needToCap, capDepEdges, extDepToCapIds, capToExtDepIds })
      })
      .catch((e: unknown) => setError((e as Error).message))
      .finally(() => setLoading(false))
  }, [isHydrating, modelId])

  useEffect(() => { loadMap() }, [loadMap])

  useEffect(() => {
    if (!modelId || isHydrating) return
    api.getTeams(modelId).then(r => setTeams(r.teams.map(t => t.name))).catch(() => {})
  }, [modelId, isHydrating])

  // Reload map when a changeset is committed (refreshKey bumps in ChangesetContext)
  useEffect(() => {
    if (refreshKey > 0) loadMap()
  }, [refreshKey, loadMap])

  const handleSaveEdit = useCallback(() => {
    if (!editState) return
    const actions: ChangeAction[] = []
    if (editState.description !== editState.origDescription)
      actions.push({ type: 'update_description', entity_type: 'capability', entity_name: editState.capLabel, description: editState.description })
    if (editState.visibility !== editState.origVisibility)
      actions.push({ type: 'update_capability_visibility', capability_name: editState.capLabel, visibility: editState.visibility })
    if (editState.teamName !== editState.origTeam)
      actions.push({ type: 'reassign_capability', capability_name: editState.capLabel, from_team_name: editState.origTeam || undefined, to_team_name: editState.teamName || undefined })
    if (actions.length === 0) return
    // Batch all actions via global context — commit from the bottom bar
    if (!isEditMode) enterEditMode()
    actions.forEach(a => addAction(a))
    const capName = editState.capLabel
    setStagedCaps(prev => new Set([...prev, capName]))
    setTimeout(() => setStagedCaps(prev => { const n = new Set(prev); n.delete(capName); return n }), 2000)
    setPanel(null)
    setHighlight(null)
    setEditState(null)
  }, [editState, isEditMode, enterEditMode, addAction])

  useEffect(() => {
    const el = containerRef.current
    if (!el) return
    const onWheel = (e: WheelEvent) => {
      if (!e.ctrlKey && !e.metaKey) return
      e.preventDefault()
      const factor = 1 - e.deltaY * 0.003
      setZoom(z => Math.min(3, Math.max(0.2, z * Math.max(0.85, Math.min(1.15, factor)))))
    }
    el.addEventListener('wheel', onWheel, { passive: false })
    return () => el.removeEventListener('wheel', onWheel)
  }, [layout])

  const openNodePanel = useCallback((node: PNode) => {
    if (!chainData) return
    setEditState(null)

    if (node.type === 'ext-dep' && node.extDep) {
      const severityColor = node.extDep.is_critical ? '#ef4444' : node.extDep.is_warning ? '#f59e0b' : '#6b7280'
      // Build list of capabilities that use this ext-dep via linked caps
      setPanel({
        title: node.label,
        badge: { text: 'External Dependency', color: severityColor },
        fields: [
          { label: 'Description', value: node.extDep.description ?? '' },
          ...(node.extDep.is_critical ? [{ label: 'Severity', value: 'Critical — this dependency is flagged as critical' }] : node.extDep.is_warning ? [{ label: 'Severity', value: 'Warning — this dependency has issues' }] : []),
          { label: 'Services using this', value: node.extDep.services.join('\n') },
          { label: 'Service count', value: String(node.extDep.service_count) },
        ],
      })
      // Highlight ext-dep + all caps that use it
      setHighlight(computeChain(node.id, 'ext-dep', chainData))
      return
    }

    if (node.type === 'actor') {
      // Build needs list for this actor
      const actorNeedEdges = chainData.actorToNeed.filter(e => e.source === node.id)
      const needFields: PanelField[] = []
      if (actorNeedEdges.length > 0) {
        for (const ne of actorNeedEdges) {
          const needNode = layout?.nodePos.get(ne.target)
          if (!needNode) continue
          const capEdges = chainData.needToCap.filter(e => e.source === ne.target)
          const capNames = capEdges.map(ce => layout?.nodePos.get(ce.target)?.label).filter(Boolean)
          const isMapped = needNode.isMapped !== false
          const statusBadge = isMapped ? '[Mapped]' : '[Unmapped]'
          const capText = capNames.length > 0 ? capNames.join(', ') : 'No capabilities linked'
          needFields.push({ label: `${statusBadge} ${needNode.label}`, value: capText })
        }
      }
      setPanel({
        title: node.label, badge: { text: 'Actor', color: '#3b82f6' },
        fields: [
          { label: 'Description', value: node.description ?? '' },
          ...(needFields.length > 0
            ? [{ label: 'Needs', value: '' }, ...needFields]
            : [{ label: 'Needs', value: 'No needs defined for this actor' }]),
        ],
      })
    } else if (node.type === 'need') {
      setPanel({
        title: node.label, badge: { text: 'Need', color: node.isMapped === false ? '#ef4444' : '#2563eb' },
        fields: [
          { label: 'Outcome', value: node.outcome ?? '' },
          { label: 'Status', value: node.isMapped === false ? 'Unmapped — no capability supports this need' : 'Mapped' },
        ],
      })
    } else {
      const cfg = VIS[node.vis ?? 'foundational']
      const svcsText = node.svcs?.map(s => `${s.label}${s.teamName ? ` (${s.teamName})` : ''}`).join('\n') ?? ''
      const nodeSlug = slug(node.label)
      const ai = insights[`cap:${nodeSlug}`] ?? insights[`cap-fragmented:${nodeSlug}`] ?? insights[`cap-disconnected:${nodeSlug}`]
      const aiFields: PanelField[] = ai
        ? [
            { label: 'AI Insight', value: ai.explanation },
            ...(ai.suggestion ? [{ label: 'Recommendation', value: ai.suggestion }] : []),
          ]
        : []
      const uniqueTeamsForPanel = [...new Set((node.svcs ?? []).map(s => s.teamName).filter(Boolean))]
      setPanel({
        title: node.label, badge: { text: cfg?.label ?? node.vis ?? '', color: cfg?.border ?? '#94a3b8' },
        fields: [
          { label: 'Description', value: node.description ?? '' },
          { label: 'Visibility', value: node.vis ?? '' },
          { label: 'Owning Team', value: node.team?.label ?? 'Unowned' },
          { label: 'Team Type', value: node.team?.type ?? '' },
          { label: 'Realized By', value: svcsText },
          ...(node.crossTeam ? [{ label: '⚠ Multi-team', value: `Services owned by ${uniqueTeamsForPanel.length} different teams: ${uniqueTeamsForPanel.join(', ')}` }] : []),
          ...(node.isFragmented ? [{ label: '⚠ Fragmented', value: 'Multiple teams own services for this capability — consider consolidating ownership' }] : []),
          ...aiFields,
        ],
      })
      setEditState({
        capLabel: node.label,
        description: node.description ?? '',
        visibility: node.vis ?? 'foundational',
        teamName: node.team?.label ?? '',
        origDescription: node.description ?? '',
        origVisibility: node.vis ?? 'foundational',
        origTeam: node.team?.label ?? '',
        svcs: node.svcs ?? [],
      })
    }

    setHighlight(computeChain(node.id, node.type, chainData))
  }, [chainData, insights, layout])

  const openConnPanel = useCallback((conn: Conn) => {
    if (!layout) return
    const src = layout.nodePos.get(conn.sourceId)
    const tgt = layout.nodePos.get(conn.targetId)
    setHighlight(new Set([conn.sourceId, conn.targetId]))
    const descField = conn.description ? [{ label: 'Description', value: conn.description }] : []
    if (conn.edgeType === 'actor-need') {
      setPanel({ title: 'Actor → Need', badge: { text: 'Demand', color: '#3b82f6' }, fields: [{ label: 'Actor', value: src?.label ?? '' }, { label: 'Need', value: tgt?.label ?? '' }, ...descField] })
    } else if (conn.edgeType === 'need-capability') {
      setPanel({ title: 'Need → Capability', badge: { text: 'Support', color: '#6366f1' }, fields: [{ label: 'Need', value: src?.label ?? '' }, { label: 'Capability', value: tgt?.label ?? '' }, ...descField] })
    } else if (conn.edgeType === 'ext-dep') {
      setPanel({ title: 'External Dependency', badge: { text: 'External', color: '#f59e0b' }, fields: [{ label: 'Capability', value: src?.label ?? '' }, { label: 'Depends on', value: tgt?.label ?? '' }, ...descField] })
    } else {
      setPanel({ title: 'Capability Dependency', badge: { text: 'Supply-side', color: '#7c3aed' }, fields: [{ label: 'From', value: src?.label ?? '' }, { label: 'Depends on', value: tgt?.label ?? '' }, ...descField] })
    }
  }, [layout])

  if (loading) return <LoadingState message="Building UNM map…" />
  if (error)   return <ErrorState message={error} />
  if (!layout) return null

  const { pnodes, conns, depConns, extDepConns, canvasWidth, canvasH, actorGroups, bands } = layout
  const hl = highlight

  const nodeOpacity  = (id: string) => hl ? (hl.has(id) ? 1 : 0.1) : 1
  const connOpacity  = (src: string, tgt: string, base: number) => hl ? (hl.has(src) && hl.has(tgt) ? base : 0.03) : base

  return (
    <ModelRequired>
      <div className="h-full flex flex-col">
      {/* Legend + Zoom controls */}
      <div className="mb-3 flex items-center gap-4 flex-wrap flex-shrink-0">
        <h2 className="text-xl font-semibold" style={{ color: '#111827' }}>UNM Map</h2>
        <div className="flex gap-3 text-xs flex-wrap" style={{ color: '#6b7280' }}>
          <span className="flex items-center gap-1" style={{ color: '#2563eb' }}>● Actor</span>
          <span className="flex items-center gap-1" style={{ color: '#3b82f6' }}>■ Need</span>
          {VIS_ORDER.map(v => (
            <span key={v} className="flex items-center gap-1" style={{ color: VIS[v].border }}>■ {VIS[v].label}</span>
          ))}
          <span className="flex items-center gap-1.5 pl-3" style={{ color: '#ef4444', borderLeft: '1px solid #e5e7eb' }}>
            <svg width="20" height="8"><line x1="0" y1="4" x2="20" y2="4" stroke="#ef4444" strokeWidth="1.5" strokeDasharray="4 3"/></svg>
            ext. dep
          </span>
          {hl && <button
            onClick={() => { setPanel(null); setHighlight(null) }}
            style={{
              cursor: 'pointer', color: '#6b7280', fontSize: 12,
              background: 'none', border: 'none', padding: '2px 8px',
              borderRadius: 4, display: 'inline-flex', alignItems: 'center', gap: 4,
              marginLeft: 12, borderLeft: '1px solid #e5e7eb',
            }}
            onMouseEnter={e => { e.currentTarget.style.color = '#111827'; e.currentTarget.style.background = '#f3f4f6' }}
            onMouseLeave={e => { e.currentTarget.style.color = '#6b7280'; e.currentTarget.style.background = 'none' }}
          >
            ✕ Clear highlight
          </button>}
        </div>
        <div className="ml-auto flex items-center gap-1">
          <div className="w-px h-5 mx-1" style={{ background: '#e5e7eb' }} />
          <button onClick={() => setZoom(z => Math.min(3, z + 0.15))} className="p-1.5 rounded hover:bg-gray-100" title="Zoom in"><ZoomIn size={15} /></button>
          <button onClick={() => setZoom(z => Math.max(0.3, z - 0.15))} className="p-1.5 rounded hover:bg-gray-100" title="Zoom out"><ZoomOut size={15} /></button>
          <button onClick={() => setZoom(1)} className="p-1.5 rounded hover:bg-gray-100" title="Reset zoom"><Maximize2 size={15} /></button>
          <span className="text-xs tabular-nums w-10 text-center" style={{ color: '#9ca3af' }}>{Math.round(zoom * 100)}%</span>
        </div>
      </div>

      {/* Canvas + optional edit panel */}
      <div className="flex-1 flex relative rounded-xl overflow-hidden" style={{ border: isEditMode ? '2px solid #3b82f6' : '1px solid #e5e7eb', transition: 'border-color 0.2s' }}>
        <div className="flex-1 relative overflow-hidden">
        <div
          ref={containerRef}
          className="absolute inset-0 overflow-auto"
          style={{ background: '#f9fafb', cursor: 'grab' }}
          onClick={clearSelection}
          onMouseDown={handleMouseDown}
          onMouseMove={handleMouseMove}
          onMouseUp={handleMouseUp}
          onMouseLeave={handleMouseUp}
        >
          <div style={{ transformOrigin: '0 0', transform: `scale(${zoom})` }}>
            <div className="relative" style={{ width: canvasWidth, height: canvasH, minWidth: '100%' }}>

              {/* ── SVG layer ── */}
            <svg className="absolute inset-0" width={canvasWidth} height={canvasH} style={{ overflow: 'visible' }}>
              <defs>
                <marker id="arrow-blue" markerWidth="8" markerHeight="8" refX="6" refY="3" orient="auto">
                  <path d="M0,0 L0,6 L8,3 z" fill="#3b82f6"/>
                </marker>
                <marker id="arrow-gray" markerWidth="8" markerHeight="8" refX="6" refY="3" orient="auto">
                  <path d="M0,0 L0,6 L8,3 z" fill="#9ca3af"/>
                </marker>
                <marker id="arrow-purple" markerWidth="8" markerHeight="8" refX="6" refY="3" orient="auto">
                  <path d="M0,0 L0,6 L8,3 z" fill="#7c3aed"/>
                </marker>
                <marker id="arrow-red" markerWidth="8" markerHeight="8" refX="6" refY="3" orient="auto">
                  <path d="M0,0 L0,6 L8,3 z" fill="#ef4444"/>
                </marker>
                <marker id="arrow-amber" markerWidth="8" markerHeight="8" refX="6" refY="3" orient="auto">
                  <path d="M0,0 L0,6 L8,3 z" fill="#f59e0b"/>
                </marker>
              </defs>
              <rect x={0} y={0} width={canvasWidth} height={canvasH} fill="transparent" onClick={clearSelection} />

              {/* Band backgrounds — dynamic heights */}
              {bands.map((band) => {
                const cfg = VIS[band.vis]
                if (!cfg) return null
                return (
                  <g key={band.vis} style={{ pointerEvents: 'none' }}>
                    <rect x={0} y={band.y} width={canvasWidth} height={band.h} fill={cfg.bandBg} />
                    <line x1={0} y1={band.y} x2={canvasWidth} y2={band.y} stroke={cfg.border} strokeWidth={0.5} opacity={0.35} />
                    <text x={canvasWidth - 14} y={band.y + 14} textAnchor="end" fontSize={9} fontWeight={700} fill={cfg.border} opacity={0.6}
                      style={{ textTransform: 'uppercase', letterSpacing: '0.1em', fontFamily: 'ui-monospace, monospace' }}>
                      {cfg.label}
                    </text>
                  </g>
                )
              })}

              {/* Actor group outlines */}
              {actorGroups.map(g => (
                <rect key={`grp-${g.id}`}
                  x={g.secStart - 10} y={NEED_Y - 18}
                  width={(g.secEnd - g.secStart) + 20} height={(VIS_FIRST_Y - 8) - (NEED_Y - 18)}
                  rx={10} fill="none" stroke="#3b82f6" strokeWidth={1} strokeDasharray="6 4"
                  opacity={hl ? 0.1 : 0.3}
                  style={{ pointerEvents: 'none' }} />
              ))}

              {/* Demand-side visual paths */}
              {conns.map((c, i) => {
                const my = (c.y1 + c.y2) / 2; const d = `M ${c.x1} ${c.y1} C ${c.x1} ${my}, ${c.x2} ${my}, ${c.x2} ${c.y2}`
                const marker = c.edgeType === 'actor-need' ? 'url(#arrow-gray)' : 'url(#arrow-blue)'
                return <path key={`dv-${i}`} d={d} fill="none" stroke={c.color} strokeWidth={1.2} opacity={connOpacity(c.sourceId, c.targetId, 0.35)} markerEnd={marker} style={{ pointerEvents: 'none' }}>
                  {c.description && <title>{c.description}</title>}
                </path>
              })}

              {/* Supply-side visual paths */}
              {depConns.map((c, i) => {
                const my = (c.y1 + c.y2) / 2; const d = `M ${c.x1} ${c.y1} C ${c.x1} ${my}, ${c.x2} ${my}, ${c.x2} ${c.y2}`
                return <path key={`sv-${i}`} d={d} fill="none" stroke={c.color} strokeWidth={1.5} strokeDasharray="5 4" opacity={connOpacity(c.sourceId, c.targetId, 0.5)} markerEnd="url(#arrow-purple)" style={{ pointerEvents: 'none' }}>
                  {c.description && <title>{c.description}</title>}
                </path>
              })}

              {/* External dependency edges */}
              {extDepConns.map((c, i) => {
                const my = (c.y1 + c.y2) / 2; const d = `M ${c.x1} ${c.y1} C ${c.x1} ${my}, ${c.x2} ${my}, ${c.x2} ${c.y2}`
                const markerColor = c.color === '#ef4444' ? 'url(#arrow-red)' : c.color === '#f59e0b' ? 'url(#arrow-amber)' : 'url(#arrow-gray)'
                return <path key={`ed-${i}`} d={d} fill="none" stroke={c.color} strokeWidth={1.5} strokeDasharray="4 3" opacity={connOpacity(c.sourceId, c.targetId, 0.5)} markerEnd={markerColor} style={{ pointerEvents: 'none' }}>
                  {c.description && <title>{c.description}</title>}
                </path>
              })}

              {/* Hit areas — demand */}
              {conns.map((c, i) => {
                const my = (c.y1 + c.y2) / 2; const d = `M ${c.x1} ${c.y1} C ${c.x1} ${my}, ${c.x2} ${my}, ${c.x2} ${c.y2}`
                return <path key={`dh-${i}`} d={d} fill="none" stroke="transparent" strokeWidth={14}
                  style={{ pointerEvents: 'stroke', cursor: 'pointer' }}
                  onClick={e => { e.stopPropagation(); openConnPanel(c) }} />
              })}

              {/* Hit areas — supply */}
              {depConns.map((c, i) => {
                const my = (c.y1 + c.y2) / 2; const d = `M ${c.x1} ${c.y1} C ${c.x1} ${my}, ${c.x2} ${my}, ${c.x2} ${c.y2}`
                return <path key={`sh-${i}`} d={d} fill="none" stroke="transparent" strokeWidth={14}
                  style={{ pointerEvents: 'stroke', cursor: 'pointer' }}
                  onClick={e => { e.stopPropagation(); openConnPanel(c) }} />
              })}

              {/* Hit areas — ext deps */}
              {extDepConns.map((c, i) => {
                const my = (c.y1 + c.y2) / 2; const d = `M ${c.x1} ${c.y1} C ${c.x1} ${my}, ${c.x2} ${my}, ${c.x2} ${c.y2}`
                return <path key={`eh-${i}`} d={d} fill="none" stroke="transparent" strokeWidth={14}
                  style={{ pointerEvents: 'stroke', cursor: 'pointer' }}
                  onClick={e => { e.stopPropagation(); openConnPanel(c) }} />
              })}
            </svg>

            {/* ── HTML node layer ── */}
            {pnodes.map(node => {
              const op = nodeOpacity(node.id)

              if (node.type === 'actor') {
                return (
                  <div key={node.id} className="absolute flex items-center justify-center gap-1.5 font-semibold rounded-xl select-none cursor-pointer"
                    style={{ left: node.x, top: node.y, width: node.w, height: node.h, zIndex: 10, opacity: op,
                      background: '#eff6ff', border: '2px solid #3b82f6', boxShadow: '0 1px 6px rgba(59,130,246,0.15)',
                      fontSize: 11, color: '#1d4ed8', transition: 'opacity 0.15s' }}
                    title={node.label}
                    onClick={e => { e.stopPropagation(); openNodePanel(node) }}
                  >
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#3b82f6" strokeWidth="2.5">
                      <circle cx="12" cy="7" r="4" /><path d="M4 21c0-5 3.6-8 8-8s8 3 8 8" />
                    </svg>
                    <span style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', maxWidth: node.w - 36 }}>{node.label}</span>
                  </div>
                )
              }

              if (node.type === 'need') {
                const unmapped = node.isMapped === false
                return (
                  <div key={node.id} className="absolute flex items-center justify-center text-center rounded-lg select-none cursor-pointer"
                    style={{ left: node.x, top: node.y, width: node.w, height: node.h, zIndex: 10, opacity: op,
                      background: unmapped ? '#fef2f2' : '#eff6ff', border: `1.5px solid ${unmapped ? '#ef4444' : '#3b82f6'}`,
                      fontSize: 10, fontWeight: 500, lineHeight: 1.3,
                      color: unmapped ? '#b91c1c' : '#1e40af', padding: '4px 8px', transition: 'opacity 0.15s' }}
                    title={node.label}
                    onClick={e => { e.stopPropagation(); openNodePanel(node) }}
                  >
                    {node.label}
                  </div>
                )
              }

              if (node.type === 'ext-dep') {
                const dep = node.extDep!
                return (
                  <div key={node.id} className="absolute rounded-lg select-none cursor-pointer flex items-center justify-center"
                    style={{ left: node.x, top: node.y, width: node.w, height: node.h, zIndex: 10, opacity: op,
                      background: dep.is_critical ? '#fef2f2' : dep.is_warning ? '#fffbeb' : '#f1f5f9',
                      border: `1.5px dashed ${dep.is_critical ? '#ef4444' : dep.is_warning ? '#f59e0b' : '#94a3b8'}`,
                      fontSize: 9, fontWeight: 600, transition: 'opacity 0.15s',
                      color: dep.is_critical ? '#b91c1c' : dep.is_warning ? '#92400e' : '#475569' }}
                    title={node.label}
                    onClick={e => { e.stopPropagation(); openNodePanel(node) }}
                  >
                    <span style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', padding: '0 6px' }}>{node.label}</span>
                  </div>
                )
              }

              if (node.type === 'capability') {
                const cfg = VIS[node.vis ?? 'foundational'] ?? VIS.foundational
                const svcs = node.svcs ?? []
                const MAX_SHOWN = 3
                const shown = svcs.slice(0, MAX_SHOWN)
                const extra = svcs.length - MAX_SHOWN

                const uniqueTeams = new Set(svcs.map(s => s.teamName).filter(Boolean))
                const crossTeam = uniqueTeams.size > 1

                const isPending = pendingCapNames.has(node.label)
                const justStaged = stagedCaps.has(node.label)

                return (
                  <div key={node.id} className="absolute rounded-lg select-none cursor-pointer"
                    style={{ left: node.x, top: node.y, width: node.w, height: node.h, zIndex: 10, opacity: op,
                      background: cfg.nodeBg, overflow: 'hidden', transition: 'opacity 0.15s, border-color 0.2s, box-shadow 0.2s',
                      border: justStaged
                        ? '2px solid #059669'
                        : isPending
                        ? '2px solid #3b82f6'
                        : `1.5px solid ${(node.isFragmented || crossTeam) ? '#ef4444' : cfg.border}`,
                      boxShadow: justStaged
                        ? '0 0 0 3px rgba(5,150,105,0.15)'
                        : isPending
                        ? '0 0 0 3px rgba(59,130,246,0.15)'
                        : (node.isFragmented || crossTeam)
                        ? '0 0 8px rgba(239,68,68,0.3)'
                        : undefined }}
                    title={isEditMode ? 'Click to edit' : node.label}
                    onClick={e => { e.stopPropagation(); openNodePanel(node) }}
                  >
                    <div style={{ fontSize: 10, fontWeight: 600, color: cfg.text, padding: '5px 8px 2px', lineHeight: 1.3 }}>
                      {node.label}
                    </div>
                    {node.team && (
                      <div style={{ fontSize: 9, color: '#6b7280', paddingLeft: 8, lineHeight: 1.2, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                        {node.team.label}
                      </div>
                    )}

                    {svcs.length > 0 && (
                      <div style={{ padding: '3px 6px 0', borderTop: `1px solid ${cfg.border}33`, marginTop: 3 }}>
                        {shown.map(s => (
                          <div key={s.id} style={{ display: 'flex', alignItems: 'center', gap: 4, marginBottom: 2 }}>
                            <div style={{ width: 5, height: 5, borderRadius: '50%', background: teamColor(s.teamName), flexShrink: 0 }} />
                            <span style={{ fontSize: 8.5, color: '#6b7280', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', flex: 1 }}>
                              {s.label}
                            </span>
                          </div>
                        ))}
                        {extra > 0 && (
                          <div style={{ fontSize: 8, color: '#9ca3af', paddingLeft: 9 }}>+{extra} more</div>
                        )}
                      </div>
                    )}

                    {(node.isFragmented || crossTeam) && (
                      <div style={{ position: 'absolute', top: 4, right: isPending ? 18 : 6, fontSize: 9, color: '#ef4444' }}>⚠</div>
                    )}
                    {justStaged && (
                      <div style={{ position: 'absolute', top: 3, right: 5, fontSize: 8, color: '#059669', fontWeight: 700, background: 'rgba(5,150,105,0.1)', borderRadius: 3, padding: '1px 3px' }}>✓</div>
                    )}
                    {isPending && !justStaged && (
                      <div style={{ position: 'absolute', top: 4, right: 5, width: 7, height: 7, borderRadius: '50%', background: '#3b82f6' }} />
                    )}
                  </div>
                )
              }
              return null
            })}
            </div>
          </div>
        </div>

        {/* Detail panel — outside scroll area, fixed to container edge */}
        </div>

        {/* Detail drawer — slides in from right, fixed position */}
        <div
          style={{
            position: 'fixed', right: 0, top: 56, bottom: 0, width: 320,
            background: 'white', borderLeft: '1px solid #e5e7eb',
            overflowY: 'auto', zIndex: 50,
            transform: panel ? 'translateX(0)' : 'translateX(100%)',
            transition: 'transform 0.2s ease',
            boxShadow: '-4px 0 12px rgba(0,0,0,0.08)',
          }}
          onClick={e => e.stopPropagation()}
        >
          {panel && (
            <>
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '12px 16px', borderBottom: '1px solid #e5e7eb' }}>
                <div style={{ flex: 1, minWidth: 0 }}>
                  {panel.badge && (
                    <span style={{
                      display: 'inline-block', marginBottom: 6, padding: '2px 8px', borderRadius: 4,
                      fontSize: 12, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.05em',
                      background: panel.badge.color + '18', color: panel.badge.color,
                      border: `1px solid ${panel.badge.color}44`,
                    }}>
                      {panel.badge.text}
                    </span>
                  )}
                  <div style={{ fontWeight: 600, fontSize: 14, color: '#111827', lineHeight: 1.3 }}>{panel.title}</div>
                  {panel.subtitle && <div style={{ fontSize: 12, color: '#6b7280', marginTop: 2 }}>{panel.subtitle}</div>}
                </div>
                <button
                  onClick={() => { setPanel(null); setHighlight(null) }}
                  style={{ background: 'none', border: 'none', cursor: 'pointer', color: '#6b7280', fontSize: 18, lineHeight: 1, marginLeft: 8, flexShrink: 0 }}
                >
                  ✕
                </button>
              </div>
              <div style={{ padding: 16 }}>

                {/* Edit form — shown FIRST in edit mode for capabilities */}
                {editState && (
                  <div style={{ borderBottom: isEditMode ? '1px solid #e5e7eb' : 'none', marginBottom: isEditMode ? 16 : 0, paddingBottom: isEditMode ? 4 : 0 }}>
                    <div style={{ fontSize: 12, fontWeight: 600, color: '#374151', marginBottom: 12 }}>Edit this capability</div>

                    <div style={{ marginBottom: 10 }}>
                      <div style={{ fontSize: 11, color: '#6b7280', marginBottom: 4 }}>Description</div>
                      <textarea
                        value={editState.description}
                        onChange={e => setEditState(s => s && { ...s, description: e.target.value })}
                        rows={3}
                        style={{ width: '100%', fontSize: 12, padding: '6px 8px', borderRadius: 6, border: '1px solid #d1d5db', resize: 'vertical', minHeight: 56, boxSizing: 'border-box', fontFamily: 'inherit', lineHeight: 1.4 }}
                      />
                    </div>

                    <div style={{ marginBottom: 10 }}>
                      <div style={{ fontSize: 11, color: '#6b7280', marginBottom: 4 }}>Visibility</div>
                      <select
                        value={editState.visibility}
                        onChange={e => setEditState(s => s && { ...s, visibility: e.target.value })}
                        style={{ width: '100%', fontSize: 12, padding: '5px 8px', borderRadius: 6, border: '1px solid #d1d5db', background: '#fff' }}
                      >
                        <option value="user-facing">User-facing</option>
                        <option value="domain">Domain</option>
                        <option value="foundational">Foundational</option>
                        <option value="infrastructure">Infrastructure</option>
                      </select>
                    </div>

                    <div style={{ marginBottom: 14 }}>
                      <div style={{ fontSize: 11, color: '#6b7280', marginBottom: 4 }}>Owning Team</div>
                      <select
                        value={editState.teamName}
                        onChange={e => setEditState(s => s && { ...s, teamName: e.target.value })}
                        style={{ width: '100%', fontSize: 12, padding: '5px 8px', borderRadius: 6, border: '1px solid #d1d5db', background: '#fff' }}
                      >
                        <option value="">— Unowned —</option>
                        {teams.map(t => <option key={t} value={t}>{t}</option>)}
                      </select>
                    </div>

                    {editState.svcs.length > 0 && (
                      <div style={{ marginBottom: 14 }}>
                        <div style={{ fontSize: 11, color: '#6b7280', marginBottom: 6 }}>Move Service to Team</div>
                        {editState.svcs.map(svc => (
                          <div key={svc.id} style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 6 }}>
                            <span style={{ fontSize: 11, color: '#374151', flex: 1, minWidth: 0, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{svc.label}</span>
                            <select
                              defaultValue=""
                              onChange={e => {
                                const toTeam = e.target.value
                                if (!toTeam) return
                                if (!isEditMode) enterEditMode()
                                addAction({ type: 'move_service', service_name: svc.label, from_team_name: svc.teamName || undefined, to_team_name: toTeam })
                                e.target.value = ''
                              }}
                              style={{ fontSize: 11, padding: '3px 6px', borderRadius: 5, border: '1px solid #d1d5db', background: '#fff', maxWidth: 130 }}
                            >
                              <option value="">Move to…</option>
                              {teams.filter(t => t !== svc.teamName).map(t => (
                                <option key={t} value={t}>{t}</option>
                              ))}
                            </select>
                          </div>
                        ))}
                      </div>
                    )}

                    <button
                      onClick={handleSaveEdit}
                      style={{ width: '100%', padding: '7px', borderRadius: 6, background: '#111827', color: '#fff', border: 'none', fontSize: 12, fontWeight: 500, cursor: 'pointer' }}
                    >
                      Stage changes →
                    </button>
                  </div>
                )}

                {/* Info fields — below edit form in edit mode; full view in view mode */}
                {(!isEditMode || !editState) && panel.fields.map((f, i) => (
                  <div key={i} style={{ marginBottom: 12 }}>
                    <div style={{ fontSize: 11, fontWeight: 500, textTransform: 'uppercase', letterSpacing: '0.05em', color: f.label.startsWith('⚠') ? '#ef4444' : '#9ca3af', marginBottom: 4 }}>{f.label}</div>
                    <div style={{ fontSize: 13, lineHeight: 1.5, color: f.label.startsWith('⚠') ? '#dc2626' : '#374151', whiteSpace: 'pre-line' }}>{f.value || '—'}</div>
                  </div>
                ))}
                {isEditMode && editState && (
                  <details style={{ marginTop: 4 }}>
                    <summary style={{ fontSize: 11, color: '#9ca3af', cursor: 'pointer', userSelect: 'none' }}>▸ View details</summary>
                    <div style={{ marginTop: 8 }}>
                      {panel.fields.map((f, i) => (
                        <div key={i} style={{ marginBottom: 12 }}>
                          <div style={{ fontSize: 11, fontWeight: 500, textTransform: 'uppercase', letterSpacing: '0.05em', color: f.label.startsWith('⚠') ? '#ef4444' : '#9ca3af', marginBottom: 4 }}>{f.label}</div>
                          <div style={{ fontSize: 13, lineHeight: 1.5, color: f.label.startsWith('⚠') ? '#dc2626' : '#374151', whiteSpace: 'pre-line' }}>{f.value || '—'}</div>
                        </div>
                      ))}
                    </div>
                  </details>
                )}
              </div>
            </>
          )}
        </div>
      </div>
    </div>
    </ModelRequired>
  )
}

