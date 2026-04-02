import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, act, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { NeedView } from './NeedView'

vi.mock('@/lib/model-context', () => ({
  useModel: () => ({ modelId: 'test-model', isHydrating: false }),
  useRequireModel: () => ({ modelId: 'test-model', isHydrating: false }),
}))

vi.mock('@/lib/search-context', () => ({
  useSearch: () => ({ query: '' }),
  matchesQuery: (text: string, q: string) => text.toLowerCase().includes(q.toLowerCase()),
}))

vi.mock('@/hooks/usePageInsights', () => ({
  usePageInsights: () => ({ insights: {}, loading: false, aiError: false, status: 'idle' }),
}))

vi.mock('@/components/ui/ModelRequired', () => ({
  ModelRequired: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}))

vi.mock('@/services/api', () => ({
  viewsApi: {
    getNeedView: vi.fn().mockResolvedValue({
      view_type: 'need',
      total_needs: 3,
      unmapped_count: 1,
      groups: [
        {
          actor: { id: 'actor1', label: 'Customer', type: 'actor', data: {} },
          needs: [
            {
              need: { id: 'need1', label: 'Place Order', type: 'need', data: { is_mapped: true, at_risk: false, unbacked: false, outcome: 'Order placed', teams: [] } },
              capabilities: [],
            },
          ],
        },
      ],
    }),
  },
}))

function makeClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } })
}

function renderNeedView() {
  return render(
    <QueryClientProvider client={makeClient()}>
      <MemoryRouter>
        <NeedView />
      </MemoryRouter>
    </QueryClientProvider>
  )
}

describe('NeedView', () => {
  beforeEach(() => { vi.clearAllMocks() })

  it('renders without crashing', async () => {
    await act(async () => { renderNeedView() })
    expect(document.body).toBeTruthy()
  })

  it('shows loading state initially', () => {
    act(() => { renderNeedView() })
    expect(screen.getByText(/loading/i)).toBeInTheDocument()
  })

  it('renders the page title after data loads', async () => {
    renderNeedView()
    await waitFor(() => {
      expect(screen.getByText('Need View')).toBeInTheDocument()
    })
  })

  it('renders need cards after data loads', async () => {
    renderNeedView()
    await waitFor(() => {
      expect(screen.getByText('Place Order')).toBeInTheDocument()
    })
  })
})
