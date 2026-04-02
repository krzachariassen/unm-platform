import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, act, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { DashboardPage } from './DashboardPage'

vi.mock('@/lib/model-context', () => ({
  useModel: () => ({
    modelId: 'test-model',
    isHydrating: false,
    parseResult: {
      id: 'test-model',
      system_name: 'Test System',
      system_description: 'A test system',
      summary: { actors: 2, needs: 5, capabilities: 3, services: 4, teams: 2, external_dependencies: 1 },
      validation: { is_valid: true, errors: [], warnings: [] },
    },
    loadedAt: new Date(),
    setModel: vi.fn(),
    clearModel: vi.fn(),
  }),
  useRequireModel: () => ({ modelId: 'test-model', isHydrating: false }),
}))

vi.mock('@/components/ui/ModelRequired', () => ({
  ModelRequired: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}))

vi.mock('@/features/dashboard/HealthCard', () => ({
  HealthCard: () => <div data-testid="health-card" />,
}))

vi.mock('@/features/dashboard/SignalsCard', () => ({
  SignalsCard: () => <div data-testid="signals-card" />,
}))

vi.mock('@/features/dashboard/TeamLoadCard', () => ({
  TeamLoadCard: () => <div data-testid="team-load-card" />,
}))

vi.mock('@/services/api', () => ({
  viewsApi: {
    getSignalsView: vi.fn().mockResolvedValue({
      view_type: 'signals',
      health: { ux_risk: 'green', architecture_risk: 'green', org_risk: 'green' },
      user_experience_layer: { needs_requiring_3plus_teams: [], needs_with_no_capability_backing: [], needs_at_risk: [] },
      architecture_layer: { user_facing_caps_with_cross_team_services: [], capabilities_not_connected_to_any_need: [], capabilities_fragmented_across_teams: [] },
      organization_layer: { top_teams_by_structural_load: [], critical_bottleneck_services: [], low_coherence_teams: [], critical_external_deps: [] },
    }),
    getCognitiveLoadView: vi.fn().mockResolvedValue({
      view_type: 'cognitive-load',
      team_loads: [],
    }),
  },
}))

function makeClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } })
}

function renderDashboardPage() {
  return render(
    <QueryClientProvider client={makeClient()}>
      <MemoryRouter>
        <DashboardPage />
      </MemoryRouter>
    </QueryClientProvider>
  )
}

describe('DashboardPage', () => {
  beforeEach(() => { vi.clearAllMocks() })

  it('renders without crashing', async () => {
    await act(async () => { renderDashboardPage() })
    expect(document.body).toBeTruthy()
  })

  it('renders the system name', async () => {
    renderDashboardPage()
    await waitFor(() => {
      expect(screen.getByText('Test System')).toBeInTheDocument()
    })
  })

  it('renders the Explore Views section', async () => {
    renderDashboardPage()
    await waitFor(() => {
      expect(screen.getByText('Explore Views')).toBeInTheDocument()
    })
  })

  it('renders view navigation cards', async () => {
    renderDashboardPage()
    await waitFor(() => {
      expect(screen.getByText('UNM Map')).toBeInTheDocument()
      expect(screen.getByText('Need View')).toBeInTheDocument()
    })
  })
})
