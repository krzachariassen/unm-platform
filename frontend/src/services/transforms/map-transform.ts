import type { Node, Edge } from '@xyflow/react'
import type { ViewNode, ViewEdge, UNMMapExtDep, UNMMapViewResponse } from '@/types/views'
import { buildLayout } from '@/features/unm-map/layout'
import type { PNode, Conn, LayoutResult } from '@/features/unm-map/types'

export interface MapTransformResult extends LayoutResult {
  rfNodes: Node[]
  rfEdges: Edge[]
}

function connToEdge(conn: Conn, index: number, prefix: string): Edge {
  const isDashed = conn.dashed
  const isExtDep = conn.edgeType === 'ext-dep'
  const isCapDep = conn.edgeType === 'cap-dep'

  return {
    id: `${prefix}-${index}`,
    source: conn.sourceId,
    target: conn.targetId,
    type: 'smoothstep',
    style: {
      stroke: conn.color,
      strokeWidth: isCapDep ? 1.5 : 1.2,
      strokeDasharray: isDashed ? (isExtDep ? '4 3' : '5 4') : undefined,
      opacity: 0.5,
    },
    markerEnd: {
      type: 'arrowclosed' as const,
      color: conn.color,
      width: 12,
      height: 12,
    },
    data: {
      edgeType: conn.edgeType,
      sourceId: conn.sourceId,
      targetId: conn.targetId,
      description: conn.description,
      color: conn.color,
    },
    selectable: true,
    focusable: false,
  }
}

function pnodeToRfNode(node: PNode): Node {
  return {
    id: node.id,
    type: node.type,
    position: { x: node.x, y: node.y },
    data: { ...node },
    style: { width: node.w, height: node.h },
    selectable: true,
    draggable: false,
    focusable: false,
  }
}

export function transformMapResponse(
  data: UNMMapViewResponse,
  pendingCaps: ViewNode[] = [],
): MapTransformResult {
  const actors = data.nodes.filter(n => n.type === 'actor')
  const needs = data.nodes.filter(n => n.type === 'need')
  const caps = [...data.nodes.filter(n => n.type === 'capability'), ...pendingCaps]
  const actorToNeed = data.edges.filter(e => e.label === 'has need')
  const needToCap = data.edges.filter(e => e.label === 'supportedBy')
  const capDepEdges: Array<{ from: string; to: string; description?: string }> = []
  for (const e of data.edges) {
    if (e.label === 'dependsOn') capDepEdges.push({ from: e.source, to: e.target, description: e.description })
  }
  const extDeps: UNMMapExtDep[] = data.external_deps ?? []

  const layout = buildLayout(actors, needs, caps, actorToNeed, needToCap, capDepEdges, extDeps)

  // Build RF nodes
  const rfNodes: Node[] = layout.pnodes.map(pnodeToRfNode)

  // Build RF edges from conns, depConns, extDepConns
  const rfEdges: Edge[] = [
    ...layout.conns.map((c, i) => connToEdge(c, i, 'conn')),
    ...layout.depConns.map((c, i) => connToEdge(c, i, 'dep')),
    ...layout.extDepConns.map((c, i) => connToEdge(c, i, 'ext')),
  ]

  return { ...layout, rfNodes, rfEdges }
}

export function extractRawData(data: UNMMapViewResponse): {
  actors: ViewNode[]; needs: ViewNode[]; caps: ViewNode[]
  actorToNeed: ViewEdge[]; needToCap: ViewEdge[]
  capDepEdges: Array<{ from: string; to: string; description?: string }>
  extDeps: UNMMapExtDep[]
} {
  return {
    actors: data.nodes.filter(n => n.type === 'actor'),
    needs: data.nodes.filter(n => n.type === 'need'),
    caps: data.nodes.filter(n => n.type === 'capability'),
    actorToNeed: data.edges.filter(e => e.label === 'has need'),
    needToCap: data.edges.filter(e => e.label === 'supportedBy'),
    capDepEdges: data.edges
      .filter(e => e.label === 'dependsOn')
      .map(e => ({ from: e.source, to: e.target, description: e.description })),
    extDeps: data.external_deps ?? [],
  }
}
