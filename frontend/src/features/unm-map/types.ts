import type { ViewEdge, UNMMapExtDep } from '@/types/views'

export interface SvcInfo { id: string; label: string; teamName: string }

export interface PNode extends Record<string, unknown> {
  id: string
  label: string
  type: 'actor' | 'need' | 'capability' | 'ext-dep'
  x: number; y: number; w: number; h: number
  vis?: string
  team?: { label: string; type: string }
  isMapped?: boolean
  isFragmented?: boolean
  crossTeam?: boolean
  svcs?: SvcInfo[]
  description?: string
  outcome?: string
  extDep?: UNMMapExtDep
  dimmed?: boolean
}

export interface Conn {
  x1: number; y1: number; x2: number; y2: number
  color: string
  dashed?: boolean
  sourceId: string
  targetId: string
  edgeType: 'actor-need' | 'need-capability' | 'cap-dep' | 'ext-dep'
  description?: string
}

export interface ActorGroup {
  id: string
  label: string
  centerX: number
  secStart: number
  secEnd: number
  description?: string
}

export interface BandInfo extends Record<string, unknown> { vis: string; y: number; h: number }

export interface ChainData {
  actorToNeed: ViewEdge[]
  needToCap: ViewEdge[]
  capDepEdges: Array<{ from: string; to: string; description?: string }>
  extDepToCapIds: Map<string, string[]>
  capToExtDepIds: Map<string, string[]>
}

export interface PanelField { label: string; value: string }

export interface PanelItem {
  title: string
  subtitle?: string
  badge?: { text: string; color: string }
  fields: PanelField[]
}

export interface EditState {
  capLabel: string
  description: string
  visibility: string
  teamName: string
  origDescription: string
  origVisibility: string
  origTeam: string
  svcs: SvcInfo[]
  isPendingNode: boolean
  linkSvcName: string
  newSvcName: string
}

export interface LayoutResult {
  pnodes: PNode[]
  conns: Conn[]
  depConns: Conn[]
  extDepConns: Conn[]
  canvasWidth: number
  canvasH: number
  actorGroups: ActorGroup[]
  nodePos: Map<string, PNode>
  bands: BandInfo[]
}
