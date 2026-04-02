import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, act, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { RealizationView } from './RealizationView'

vi.mock('@/lib/model-context', () => ({
  useModel: () => ({ modelId: 'test-model', isHydrating: false }),
  useRequireModel: () => ({ modelId: 'test-model', isHydrating: false }),
}))

vi.mock('@/lib/search-context', () => ({
  useSearch: () => ({ query: '' }),
  matchesQuery: (text: string, q: string) => text.toLowerCase().includes(q.toLowerCase()),
}))

vi.mock('@/components/ui/ModelRequired', () => ({
  ModelRequired: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}))

vi.mock('@/features/realization/RealizationTabs', () => ({
  ValueChainView: () => <div data-testid="value-chain-view" />,
  ServiceTableView: () => <div data-testid="service-table-view" />,
}))

vi.mock('@/features/realization/utils', () => ({
  buildCapToSvcTeam: () => new Map(),
  buildCapVisibility: () => new Map(),
  buildGroupedNeeds: () => [],
}))

vi.mock('@/services/api', () => ({
  viewsApi: {
    getRealizationView: vi.fn().mockResolvedValue({
      view_type: 'realization',
      service_rows: [],
      capabilities: [],
    }),
    getNeedView: vi.fn().mockResolvedValue({
      view_type: 'need',
      total_needs: 0,
      unmapped_count: 0,
      groups: [],
    }),
  },
}))

function makeClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } })
}

function renderRealizationView() {
  return render(
    <QueryClientProvider client={makeClient()}>
      <MemoryRouter>
        <RealizationView />
      </MemoryRouter>
    </QueryClientProvider>
  )
}

describe('RealizationView', () => {
  beforeEach(() => { vi.clearAllMocks() })

  it('renders without crashing', async () => {
    await act(async () => { renderRealizationView() })
    expect(document.body).toBeTruthy()
  })

  it('shows loading state initially', () => {
    act(() => { renderRealizationView() })
    expect(screen.getByText(/loading/i)).toBeInTheDocument()
  })

  it('renders the page title after data loads', async () => {
    renderRealizationView()
    await waitFor(() => {
      expect(screen.getByText('Realization View')).toBeInTheDocument()
    })
  })

  it('renders value chain tab by default', async () => {
    renderRealizationView()
    await waitFor(() => {
      expect(screen.getByTestId('value-chain-view')).toBeInTheDocument()
    })
  })
})
