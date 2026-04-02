import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, act, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { TeamTopologyView } from './TeamTopologyView'

vi.mock('@/lib/model-context', () => ({
  useModel: () => ({ modelId: 'test-model', isHydrating: false }),
  useRequireModel: () => ({ modelId: 'test-model', isHydrating: false }),
}))

vi.mock('@/lib/search-context', () => ({
  useSearch: () => ({ query: '', teamTypeFilter: null }),
  matchesQuery: (text: string, q: string) => text.toLowerCase().includes(q.toLowerCase()),
}))

vi.mock('@/hooks/usePageInsights', () => ({
  usePageInsights: () => ({ insights: {}, loading: false, aiError: false, status: 'idle' }),
}))

vi.mock('@/components/ui/ModelRequired', () => ({
  ModelRequired: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}))

vi.mock('@/features/team-topology/GraphView', () => ({
  GraphView: () => <div data-testid="graph-view" />,
}))

vi.mock('@/features/team-topology/TableView', () => ({
  TableView: () => <div data-testid="table-view" />,
}))

vi.mock('@/services/api', () => ({
  viewsApi: {
    getTeamTopologyView: vi.fn().mockResolvedValue({
      view_type: 'team-topology',
      teams: [
        { id: 'team1', label: 'Platform Team', type: 'platform', data: { is_overloaded: false }, is_overloaded: false },
        { id: 'team2', label: 'Order Team', type: 'stream-aligned', data: { is_overloaded: false }, is_overloaded: false },
      ],
      interactions: [],
    }),
  },
}))

function makeClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } })
}

function renderTeamTopologyView() {
  return render(
    <QueryClientProvider client={makeClient()}>
      <MemoryRouter>
        <TeamTopologyView />
      </MemoryRouter>
    </QueryClientProvider>
  )
}

describe('TeamTopologyView', () => {
  beforeEach(() => { vi.clearAllMocks() })

  it('renders without crashing', async () => {
    await act(async () => { renderTeamTopologyView() })
    expect(document.body).toBeTruthy()
  })

  it('shows loading state initially', () => {
    act(() => { renderTeamTopologyView() })
    expect(screen.getByText(/loading/i)).toBeInTheDocument()
  })

  it('renders the page title after data loads', async () => {
    renderTeamTopologyView()
    await waitFor(() => {
      expect(screen.getByText('Team Topology')).toBeInTheDocument()
    })
  })

  it('renders graph view by default', async () => {
    renderTeamTopologyView()
    await waitFor(() => {
      expect(screen.getByTestId('graph-view')).toBeInTheDocument()
    })
  })
})
