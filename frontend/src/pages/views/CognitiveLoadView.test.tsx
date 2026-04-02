import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, act, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { CognitiveLoadView } from './CognitiveLoadView'

vi.mock('@/lib/model-context', () => ({
  useModel: () => ({ modelId: 'test-model', isHydrating: false }),
  useRequireModel: () => ({ modelId: 'test-model', isHydrating: false }),
}))

vi.mock('@/hooks/usePageInsights', () => ({
  usePageInsights: () => ({ insights: {}, loading: false, aiError: false, status: 'idle' }),
}))

vi.mock('@/components/ui/ModelRequired', () => ({
  ModelRequired: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}))

vi.mock('@/features/cognitive-load/TeamCard', () => ({
  TeamCard: ({ tl }: { tl: { team: { name: string } } }) => (
    <div data-testid={`team-card-${tl.team.name}`}>{tl.team.name}</div>
  ),
}))

vi.mock('@/services/api', () => ({
  viewsApi: {
    getCognitiveLoadView: vi.fn().mockResolvedValue({
      view_type: 'cognitive-load',
      team_loads: [
        {
          team: { name: 'Order Team', type: 'stream-aligned' },
          overall_level: 'medium',
          domain_spread: { value: 2, level: 'low' },
          service_load: { value: 2, level: 'medium' },
          interaction_load: { value: 3, level: 'low' },
          dependency_load: { value: 4, level: 'low' },
        },
      ],
    }),
  },
}))

function makeClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } })
}

function renderCognitiveLoadView() {
  return render(
    <QueryClientProvider client={makeClient()}>
      <MemoryRouter>
        <CognitiveLoadView />
      </MemoryRouter>
    </QueryClientProvider>
  )
}

describe('CognitiveLoadView', () => {
  beforeEach(() => { vi.clearAllMocks() })

  it('renders without crashing', async () => {
    await act(async () => { renderCognitiveLoadView() })
    expect(document.body).toBeTruthy()
  })

  it('shows loading state initially', () => {
    act(() => { renderCognitiveLoadView() })
    expect(screen.getByText(/analyzing structural cognitive load/i)).toBeInTheDocument()
  })

  it('renders the page title after data loads', async () => {
    renderCognitiveLoadView()
    await waitFor(() => {
      expect(screen.getByText('Structural Cognitive Load')).toBeInTheDocument()
    })
  })

  it('renders team cards after data loads', async () => {
    renderCognitiveLoadView()
    await waitFor(() => {
      expect(screen.getByTestId('team-card-Order Team')).toBeInTheDocument()
    })
  })
})
