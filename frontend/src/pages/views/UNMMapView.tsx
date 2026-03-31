import { useState, useCallback, useMemo } from 'react'
import { ReactFlow, useNodesState, useEdgesState, type NodeMouseHandler, type EdgeMouseHandler } from '@xyflow/react'
import { useQueryClient, useQueries } from '@tanstack/react-query'
import { ModelRequired } from '@/components/ui/ModelRequired'
import { LoadingState, ErrorState } from '@/components/ViewState'
import { useModel } from '@/lib/model-context'
import { usePageInsights } from '@/hooks/usePageInsights'
import { useChangeset } from '@/lib/changeset-context'
import { viewsApi } from '@/services/api/views'
import { modelsApi } from '@/services/api/models'
import { transformMapResponse } from '@/services/transforms/map-transform'
import { buildChainData, computeChain } from '@/features/unm-map/chain'
import { buildNodePanel, buildConnPanel } from '@/features/unm-map/panels'
import { VIS_ORDER } from '@/features/unm-map/constants'
import type { PNode, Conn, PanelItem, EditState } from '@/features/unm-map/types'
import type { ChangeAction } from '@/types/changeset'
import { slug } from '@/lib/slug'
import { ActorNode } from '@/components/unm-map/ActorNode'
import { NeedNode } from '@/components/unm-map/NeedNode'
import { CapabilityNode } from '@/components/unm-map/CapabilityNode'
import { ExtDepNode } from '@/components/unm-map/ExtDepNode'
import { BandNode } from '@/components/unm-map/BandNode'
import { MapToolbar } from '@/components/unm-map/MapToolbar'
import { DetailDrawer } from '@/components/unm-map/DetailDrawer'
import { CapabilityEditForm } from '@/components/unm-map/CapabilityEditForm'

const NODE_TYPES = { actor: ActorNode, need: NeedNode, capability: CapabilityNode, 'ext-dep': ExtDepNode, band: BandNode }

export function UNMMapView() {
  const { modelId } = useModel()
  const { isEditMode, actions, addAction, enterEditMode } = useChangeset()
  const queryClient = useQueryClient()
  const { insights } = usePageInsights('dashboard')

  const [panel, setPanel] = useState<PanelItem | null>(null)
  const [highlight, setHighlight] = useState<Set<string> | null>(null)
  const [editState, setEditState] = useState<EditState | null>(null)
  const [stagedCaps, setStagedCaps] = useState<Set<string>>(new Set())

  const [{ data: mapData, isLoading, error }, teamsQ, servicesQ] = useQueries({
    queries: [
      {
        queryKey: ['unmMapView', modelId],
        queryFn: () => viewsApi.getUNMMapView(modelId!),
        enabled: !!modelId,
      },
      {
        queryKey: ['teams', modelId],
        queryFn: () => modelsApi.getTeams(modelId!),
        enabled: !!modelId,
      },
      {
        queryKey: ['services', modelId],
        queryFn: () => modelsApi.getServices(modelId!),
        enabled: !!modelId,
      },
    ],
  })

  const teams = useMemo(() => teamsQ.data?.teams.map(t => t.name) ?? [], [teamsQ.data])
  const apiServices = useMemo(() => servicesQ.data?.services.map(s => s.name) ?? [], [servicesQ.data])

  const services = useMemo(() => {
    const pending = actions
      .filter(a => a.type === 'add_service')
      .map(a => String((a as unknown as Record<string, unknown>).service_name ?? ''))
      .filter(Boolean)
    return [...new Set([...apiServices, ...pending])].sort((a, b) => a.localeCompare(b))
  }, [apiServices, actions])

  const pendingCapNodes = useMemo(() => actions
    .filter(a => a.type === 'add_capability')
    .map(a => {
      const ac = a as unknown as Record<string, unknown>
      const name = String(ac.capability_name ?? '')
      const vis = String(ac.visibility || 'domain')
      return { id: `pending:${name}`, label: name, type: 'capability', data: { visibility: vis, services: [], isPending: true } }
    }), [actions])

  const mapResult = useMemo(() => {
    if (!mapData) return null
    return transformMapResponse(mapData, pendingCapNodes)
  }, [mapData, pendingCapNodes])

  const chainData = useMemo(() => {
    if (!mapResult) return null
    return buildChainData(
      mapData!.edges.filter(e => e.label === 'has need'),
      mapData!.edges.filter(e => e.label === 'supportedBy'),
      mapData!.edges.filter(e => e.label === 'dependsOn').map(e => ({ from: e.source, to: e.target, description: e.description })),
      mapResult.extDepConns,
    )
  }, [mapResult, mapData])

  // Build RF nodes with band backgrounds + highlight state
  const baseNodes = useMemo(() => {
    if (!mapResult) return []
    const canvasWidth = mapResult.canvasWidth
    const bandNodes = mapResult.bands.map(band => ({
      id: `band:${band.vis}`,
      type: 'band' as const,
      position: { x: 0, y: band.y },
      data: { ...band, canvasWidth },
      style: { width: canvasWidth, height: band.h },
      selectable: false, draggable: false, focusable: false,
      zIndex: -1,
    }))
    return [...bandNodes, ...mapResult.rfNodes]
  }, [mapResult])

  const [rfNodes, setRfNodes, onNodesChange] = useNodesState(baseNodes)
  const [rfEdges, setRfEdges, onEdgesChange] = useEdgesState(mapResult?.rfEdges ?? [])

  // Sync RF nodes/edges when mapResult changes (refresh after changeset commit)
  useMemo(() => {
    if (!mapResult) return
    setRfNodes(baseNodes)
    setRfEdges(mapResult.rfEdges)
  }, [mapResult, baseNodes]) // eslint-disable-line

  // Apply highlight by setting dimmed on nodes/edges
  useMemo(() => {
    setRfNodes(prev => prev.map(n => {
      if (n.type === 'band') return n
      const dimmed = highlight ? !highlight.has(n.id) : false
      return { ...n, data: { ...n.data, dimmed } }
    }))
    setRfEdges(prev => prev.map(e => {
      const src = e.data?.sourceId as string; const tgt = e.data?.targetId as string
      const dimmed = highlight ? !(highlight.has(src) && highlight.has(tgt)) : false
      return { ...e, style: { ...e.style, opacity: dimmed ? 0.03 : (e.style?.strokeDasharray ? 0.5 : 0.35) } }
    }))
  }, [highlight]) // eslint-disable-line

  const openNodePanel = useCallback((node: PNode) => {
    if (!chainData || !mapResult) return
    setEditState(null)
    const panel = buildNodePanel(
      node,
      mapData?.edges.filter(e => e.label === 'has need').map(e => ({ source: e.source, target: e.target })) ?? [],
      mapData?.edges.filter(e => e.label === 'supportedBy').map(e => ({ source: e.source, target: e.target })) ?? [],
      mapResult.nodePos,
      insights,
      slug,
    )
    setPanel(panel)
    setHighlight(computeChain(node.id, node.type as 'actor' | 'need' | 'capability' | 'ext-dep', chainData))
    if (node.type === 'capability' && !node.id.startsWith('pending:')) {
      setEditState({
        capLabel: node.label, description: node.description ?? '', visibility: node.vis ?? 'foundational',
        teamName: node.team?.label ?? '', origDescription: node.description ?? '',
        origVisibility: node.vis ?? 'foundational', origTeam: node.team?.label ?? '',
        svcs: node.svcs ?? [], isPendingNode: false, linkSvcName: '', newSvcName: '',
      })
    }
  }, [chainData, mapResult, mapData, insights])

  const handleNodeClick: NodeMouseHandler = useCallback((_evt, rfNode) => {
    if (rfNode.type === 'band') return
    const node = mapResult?.nodePos.get(rfNode.id)
    if (node) openNodePanel(node)
  }, [mapResult, openNodePanel])

  const handleEdgeClick: EdgeMouseHandler = useCallback((_evt, edge) => {
    if (!mapResult) return
    const allConns: Conn[] = [...mapResult.conns, ...mapResult.depConns, ...mapResult.extDepConns]
    const conn = allConns.find(c => c.sourceId === edge.data?.sourceId && c.targetId === edge.data?.targetId)
    if (!conn) return
    setHighlight(new Set([conn.sourceId, conn.targetId]))
    setPanel(buildConnPanel(conn, mapResult.nodePos))
    setEditState(null)
  }, [mapResult])

  const clearSelection = useCallback(() => {
    setPanel(null); setHighlight(null); setEditState(null)
  }, [])

  const handleSaveEdit = useCallback(() => {
    if (!editState) return
    const changes: ChangeAction[] = []
    if (editState.description !== editState.origDescription)
      changes.push({ type: 'update_description', entity_type: 'capability', entity_name: editState.capLabel, description: editState.description })
    if (editState.visibility !== editState.origVisibility)
      changes.push({ type: 'update_capability_visibility', capability_name: editState.capLabel, visibility: editState.visibility })
    if (editState.teamName !== editState.origTeam)
      changes.push({ type: 'reassign_capability', capability_name: editState.capLabel, from_team_name: editState.origTeam || undefined, to_team_name: editState.teamName || undefined })
    if (changes.length === 0) return
    if (!isEditMode) enterEditMode()
    changes.forEach(a => addAction(a))
    const capName = editState.capLabel
    setStagedCaps(prev => new Set([...prev, capName]))
    setTimeout(() => setStagedCaps(prev => { const n = new Set(prev); n.delete(capName); return n }), 2000)
    setPanel(null); setHighlight(null); setEditState(null)
  }, [editState, isEditMode, enterEditMode, addAction])

  // Invalidate query when changeset is committed (refreshKey bumps via useChangeset)
  useMemo(() => {
    if (!modelId) return
    queryClient.invalidateQueries({ queryKey: ['unmMapView', modelId] })
  }, [stagedCaps.size]) // eslint-disable-line

  if (isLoading) return <LoadingState message="Building UNM map…" />
  if (error) return <ErrorState message={(error as Error).message} />
  if (!mapResult) return null

  return (
    <ModelRequired>
      <div className="h-full flex flex-col">
        <MapToolbar highlighted={!!highlight} onClearHighlight={clearSelection} />

        <div
          className="flex-1 relative rounded-xl overflow-hidden"
          style={{ border: isEditMode ? '2px solid #3b82f6' : '1px solid #e5e7eb', transition: 'border-color 0.2s' }}
        >
          <ReactFlow
            nodes={rfNodes}
            edges={rfEdges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            nodeTypes={NODE_TYPES}
            onNodeClick={handleNodeClick}
            onEdgeClick={handleEdgeClick}
            onPaneClick={clearSelection}
            fitView
            fitViewOptions={{ padding: 0.05 }}
            minZoom={0.1}
            maxZoom={3}
            panOnScroll
            panOnDrag
            nodesDraggable={false}
            nodesConnectable={false}
            elementsSelectable={false}
            style={{ background: '#f9fafb' }}
          >
            {VIS_ORDER.map(_vis => null)}
          </ReactFlow>

          <DetailDrawer panel={panel} isEditMode={isEditMode} onClose={clearSelection}>
            {editState && (
              <CapabilityEditForm
                editState={editState}
                teams={teams}
                services={services}
                isEditMode={isEditMode}
                onUpdateState={updater => setEditState(s => s ? updater(s) : s)}
                onSave={handleSaveEdit}
                onMoveService={(svc, toTeam) => {
                  if (!isEditMode) enterEditMode()
                  addAction({ type: 'move_service', service_name: svc.label, from_team_name: svc.teamName || undefined, to_team_name: toTeam })
                }}
                onUnlinkService={svcLabel => {
                  if (!isEditMode) enterEditMode()
                  addAction({ type: 'unlink_capability_service', capability_name: editState.capLabel, service_name: svcLabel })
                }}
                onLinkService={svcName => {
                  if (!isEditMode) enterEditMode()
                  addAction({ type: 'link_capability_service', capability_name: editState.capLabel, service_name: svcName })
                }}
                onAddService={svcName => {
                  if (!isEditMode) enterEditMode()
                  addAction({ type: 'add_service', service_name: svcName, owner_team_name: editState.teamName || undefined })
                  addAction({ type: 'link_capability_service', capability_name: editState.capLabel, service_name: svcName })
                }}
              />
            )}
          </DetailDrawer>
        </div>
      </div>
    </ModelRequired>
  )
}
