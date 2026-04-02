import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, act, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { OwnershipView } from './OwnershipView'

vi.mock('@/lib/model-context', () => ({
  useModel: () => ({ modelId: 'test-model', isHydrating: false }),
  useRequireModel: () => ({ modelId: 'test-model', isHydrating: false }),
}))

vi.mock('@/lib/search-context', () => ({
  useSearch: () => ({ query: '', teamFilter: '' }),
  matchesQuery: (text: string, q: string) => text.toLowerCase().includes(q.toLowerCase()),
}))

vi.mock('@/hooks/usePageInsights', () => ({
  usePageInsights: () => ({ insights: {}, loading: false, aiError: false, status: 'idle' }),
}))

vi.mock('@/components/ui/ModelRequired', () => ({
  ModelRequired: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}))

vi.mock('@/components/AntiPatternPanel', () => ({
  AntiPatternPanel: () => null,
}))

vi.mock('@/features/ownership/TeamLane', () => ({
  TeamLane: () => <div data-testid="team-lane" />,
}))

vi.mock('@/features/ownership/DomainView', () => ({
  DomainView: () => <div data-testid="domain-view" />,
}))

vi.mock('@/services/api', () => ({
  viewsApi: {
    getOwnershipView: vi.fn().mockResolvedValue({
      view_type: 'ownership',
      lanes: [
        {
          team: { id: 'team1', label: 'Order Team', type: 'team', data: { type: 'stream-aligned', is_overloaded: false } },
          caps: [],
        },
      ],
      service_rows: [],
      cross_team_capabilities: [],
      unowned_capabilities: [],
      overloaded_teams: [],
      external_dependency_count: 0,
    }),
    getCapabilityView: vi.fn().mockResolvedValue({
      view_type: 'capability',
      leaf_capability_count: 0,
      high_span_services: [],
      fragmented_capabilities: [],
      parent_groups: [],
      capabilities: [],
    }),
  },
}))

function makeClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } })
}

function renderOwnershipView() {
  return render(
    <QueryClientProvider client={makeClient()}>
      <MemoryRouter>
        <OwnershipView />
      </MemoryRouter>
    </QueryClientProvider>
  )
}

describe('OwnershipView', () => {
  beforeEach(() => { vi.clearAllMocks() })

  it('renders without crashing', async () => {
    await act(async () => { renderOwnershipView() })
    expect(document.body).toBeTruthy()
  })

  it('shows loading state initially', () => {
    act(() => { renderOwnershipView() })
    expect(screen.getByText(/loading/i)).toBeInTheDocument()
  })

  it('renders the page title after data loads', async () => {
    renderOwnershipView()
    await waitFor(() => {
      expect(screen.getByText('Ownership View')).toBeInTheDocument()
    })
  })
})
