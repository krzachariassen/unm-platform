/**
 * CapabilityView regression tests.
 * Verifies that CapabilityView renders correctly after the infinite re-render fix.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { CapabilityView } from './CapabilityView'

// ── Mocks ──────────────────────────────────────────────────────────────────────

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

vi.mock('@/components/ui/ModelRequired', () => ({
  ModelRequired: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}))

vi.mock('@/services/api', () => ({
  viewsApi: {
    getCapabilityView: vi.fn().mockResolvedValue({
      view_type: 'capability',
      leaf_capability_count: 2,
      high_span_services: [],
      fragmented_capabilities: [],
      parent_groups: [
        { id: 'pg1', label: 'Core Domain', children: ['cap1', 'cap2'] },
      ],
      capabilities: [
        {
          id: 'cap1', label: 'Authentication', description: 'Auth cap',
          visibility: 'user-facing', is_leaf: true, is_fragmented: false,
          depended_on_by_count: 0, services: [], teams: [{ id: 't1', label: 'Team A', type: 'stream-aligned' }],
          depends_on: [], children: [],
        },
        {
          id: 'cap2', label: 'Authorization', description: 'Authz cap',
          visibility: 'domain', is_leaf: true, is_fragmented: false,
          depended_on_by_count: 1, services: [], teams: [],
          depends_on: [], children: [],
        },
      ],
    }),
  },
}))

// ── Helpers ────────────────────────────────────────────────────────────────────

function makeClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } })
}

function renderCapabilityView() {
  return render(
    <QueryClientProvider client={makeClient()}>
      <MemoryRouter>
        <CapabilityView />
      </MemoryRouter>
    </QueryClientProvider>
  )
}

// ── Tests ──────────────────────────────────────────────────────────────────────

describe('CapabilityView — crash regression', () => {
  beforeEach(() => { vi.clearAllMocks() })

  it('renders the page title after data loads without crashing', async () => {
    renderCapabilityView()
    await waitFor(() => {
      expect(screen.getByText('Capability View')).toBeInTheDocument()
    })
  })

  it('renders capability cards after data loads', async () => {
    renderCapabilityView()
    await waitFor(() => {
      expect(screen.getByText('Authentication')).toBeInTheDocument()
    })
  })

  it('renders description in default visibility view mode', async () => {
    renderCapabilityView()
    await waitFor(() => {
      expect(screen.getByText(/1 domain groups/)).toBeInTheDocument()
    })
  })
})
