import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, act } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { SignalsView } from './SignalsView'

// ── Mocks ────────────────────────────────────────────────────────────────────
// vi.mock is hoisted to top of file — do not reference top-level variables inside factories.

vi.mock('@/lib/model-context', () => ({
  useRequireModel: () => ({
    modelId: 'test-model-id',
    isHydrating: false,
    parseResult: null,
    loadedAt: null,
    setModel: vi.fn(),
    clearModel: vi.fn(),
  }),
  useModel: () => ({
    modelId: 'test-model-id',
    isHydrating: false,
    parseResult: null,
    loadedAt: null,
    setModel: vi.fn(),
    clearModel: vi.fn(),
  }),
}))

vi.mock('@/hooks/useAIEnabled', () => ({
  useAIEnabled: () => false,
}))

vi.mock('@/lib/api', () => ({
  api: {
    getSignals: vi.fn().mockResolvedValue({
      view_type: 'signals',
      health: { ux_risk: 'green', architecture_risk: 'amber', org_risk: 'red' },
      user_experience_layer: {
        needs_requiring_3plus_teams: [],
        needs_with_no_capability_backing: [],
        needs_at_risk: [],
      },
      architecture_layer: {
        user_facing_caps_with_cross_team_services: [],
        capabilities_not_connected_to_any_need: [],
        capabilities_fragmented_across_teams: [],
      },
      organization_layer: {
        top_teams_by_structural_load: [],
        critical_bottleneck_services: [],
        low_coherence_teams: [],
        critical_external_deps: [],
      },
    }),
  },
}))

// ── Tests ─────────────────────────────────────────────────────────────────────

function renderSignalsView() {
  return render(
    <MemoryRouter>
      <SignalsView />
    </MemoryRouter>
  )
}

describe('SignalsView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders without crashing', async () => {
    await act(async () => {
      renderSignalsView()
    })
    expect(document.body).toBeTruthy()
  })

  it('shows loading state initially then resolves', async () => {
    let container: ReturnType<typeof renderSignalsView>
    // Capture loading state before data resolves
    act(() => {
      container = renderSignalsView()
    })
    expect(screen.getByText(/loading/i)).toBeInTheDocument()
    // Wait for async resolution
    await act(async () => {})
    expect(container!.container).toBeTruthy()
  })
})
