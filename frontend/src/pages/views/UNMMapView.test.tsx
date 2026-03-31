/**
 * UNMMapView regression tests.
 * Verifies that UNMMapView renders correctly with ReactFlowProvider wrapping.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { UNMMapView } from './UNMMapView'

// ── Mocks ──────────────────────────────────────────────────────────────────────

vi.mock('@/lib/model-context', () => ({
  useModel: () => ({ modelId: 'test-model', isHydrating: false }),
  useRequireModel: () => ({ modelId: 'test-model', isHydrating: false }),
}))

vi.mock('@/lib/changeset-context', () => ({
  useChangeset: () => ({
    actions: [],
    addAction: vi.fn(),
    removeAction: vi.fn(),
    clearActions: vi.fn(),
    discardAll: vi.fn(),
    description: '',
    setDescription: vi.fn(),
    refreshKey: 0,
  }),
}))

vi.mock('@/hooks/usePageInsights', () => ({
  usePageInsights: () => ({ insights: {}, loading: false, aiError: false, status: 'idle' }),
}))

vi.mock('@/components/ui/ModelRequired', () => ({
  ModelRequired: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}))

// Mock React Flow — its DOM internals (canvas, ResizeObserver) are unavailable
// in jsdom, so we replace the visual components. HOWEVER, useNodesState and
// useEdgesState must use REAL React state — not vi.fn() no-ops — so that
// calling setRfNodes/setRfEdges inside useMemo actually triggers re-renders
// and exposes the infinite loop bug.
vi.mock('@xyflow/react', async () => {
  const { useState } = await import('react')
  return {
    ReactFlow: ({ children }: { children?: React.ReactNode }) => (
      <div data-testid="react-flow-canvas">{children}</div>
    ),
    ReactFlowProvider: ({ children }: { children?: React.ReactNode }) => <>{children}</>,
    useNodesState: (initial: unknown[]) => {
      const [nodes, setNodes] = useState(initial)
      return [nodes, setNodes, vi.fn()]
    },
    useEdgesState: (initial: unknown[]) => {
      const [edges, setEdges] = useState(initial)
      return [edges, setEdges, vi.fn()]
    },
    useReactFlow: () => ({ zoomIn: vi.fn(), zoomOut: vi.fn(), fitView: vi.fn() }),
    Background: () => null,
    Controls: () => null,
    MiniMap: () => null,
    MarkerType: { ArrowClosed: 'arrowclosed' },
  }
})

vi.mock('@/components/unm-map/MapToolbar', () => ({
  MapToolbar: () => <div data-testid="map-toolbar" />,
}))

vi.mock('@/components/unm-map/DetailDrawer', () => ({
  DetailDrawer: ({ children }: { children?: React.ReactNode }) => (
    <div data-testid="detail-drawer">{children}</div>
  ),
}))

vi.mock('@/components/unm-map/CapabilityEditForm', () => ({
  CapabilityEditForm: () => <div data-testid="capability-edit-form" />,
}))

vi.mock('@/services/api/views', () => ({
  viewsApi: {
    getUNMMapView: vi.fn().mockResolvedValue({
      view_type: 'unm-map',
      nodes: [
        { id: 'actor1', label: 'Customer', type: 'actor', data: { description: 'End customer' } },
        { id: 'need1', label: 'Place Order', type: 'need', data: { is_mapped: true, outcome: 'Order placed' } },
        {
          id: 'cap1', label: 'Order Management', type: 'capability',
          data: { visibility: 'domain', services: [], team: null, is_fragmented: false },
        },
      ],
      edges: [
        { id: 'e1', source: 'actor1', target: 'need1', label: 'has need' },
        { id: 'e2', source: 'need1', target: 'cap1', label: 'supportedBy' },
      ],
      external_deps: [],
    }),
  },
}))

vi.mock('@/services/api/models', () => ({
  modelsApi: {
    getTeams: vi.fn().mockResolvedValue({ teams: [{ name: 'Platform Team', type: 'stream-aligned' }] }),
    getServices: vi.fn().mockResolvedValue({ services: [{ name: 'order-service' }] }),
  },
}))

// ── Helpers ────────────────────────────────────────────────────────────────────

function makeClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } })
}

function renderUNMMapView() {
  return render(
    <QueryClientProvider client={makeClient()}>
      <MemoryRouter>
        <UNMMapView />
      </MemoryRouter>
    </QueryClientProvider>
  )
}

// ── Tests ──────────────────────────────────────────────────────────────────────

describe('UNMMapView — crash regression', () => {
  beforeEach(() => { vi.clearAllMocks() })

  it('renders the map canvas after data loads without crashing', async () => {
    renderUNMMapView()
    await waitFor(() => {
      expect(screen.getByTestId('react-flow-canvas')).toBeInTheDocument()
    })
  })

  it('renders the map toolbar after data loads', async () => {
    renderUNMMapView()
    await waitFor(() => {
      expect(screen.getByTestId('map-toolbar')).toBeInTheDocument()
    })
  })

  it('renders without crashing when highlight state is null (initial render)', async () => {
    renderUNMMapView()
    await waitFor(() => {
      expect(screen.getByTestId('detail-drawer')).toBeInTheDocument()
    })
  })
})
