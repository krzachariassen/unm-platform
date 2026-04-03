import { describe, it, expect, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { InteractionsView } from '@/types/views'

vi.mock('@/lib/model-context', () => ({
  useModel: () => ({ modelId: 'model-1', isHydrating: false }),
}))

vi.mock('@/services/api', () => ({
  viewsApi: {
    getInteractions: vi.fn(),
  },
}))

import { InteractionsTab } from './InteractionsTab'

function renderInteractionsTab() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>
        <InteractionsTab />
      </MemoryRouter>
    </QueryClientProvider>
  )
}

const emptyInteractions: InteractionsView = {
  model_id: 'model-1',
  mode_distribution: {},
  isolated_teams: [],
  over_reliant_teams: [],
  all_modes_same: false,
}

const richInteractions: InteractionsView = {
  model_id: 'model-1',
  mode_distribution: { 'x-as-a-service': 3, 'facilitating': 1 },
  isolated_teams: ['Team Iso'],
  over_reliant_teams: [{ team_name: 'Team Over', mode: 'x-as-a-service', count: 5 }],
  all_modes_same: true,
}

describe('InteractionsTab', () => {
  it('shows empty state when no interactions', async () => {
    const { viewsApi } = await import('@/services/api')
    vi.mocked(viewsApi.getInteractions).mockResolvedValue(emptyInteractions)
    renderInteractionsTab()
    await waitFor(() => expect(screen.getByText(/no interactions/i)).toBeInTheDocument())
  })

  it('renders mode distribution', async () => {
    const { viewsApi } = await import('@/services/api')
    vi.mocked(viewsApi.getInteractions).mockResolvedValue(richInteractions)
    renderInteractionsTab()
    // formatMode converts 'x-as-a-service' → 'X As A Service'
    await waitFor(() => expect(screen.getAllByText(/X As A Service/i).length).toBeGreaterThan(0))
    expect(screen.getByText(/Facilitating/i)).toBeInTheDocument()
  })

  it('shows isolated teams warning', async () => {
    const { viewsApi } = await import('@/services/api')
    vi.mocked(viewsApi.getInteractions).mockResolvedValue(richInteractions)
    renderInteractionsTab()
    await waitFor(() => expect(screen.getByText(/Team Iso/i)).toBeInTheDocument())
  })

  it('shows all_modes_same warning banner', async () => {
    const { viewsApi } = await import('@/services/api')
    vi.mocked(viewsApi.getInteractions).mockResolvedValue(richInteractions)
    renderInteractionsTab()
    await waitFor(() => expect(screen.getByText(/same interaction mode/i)).toBeInTheDocument())
  })

  it('shows over-reliant teams', async () => {
    const { viewsApi } = await import('@/services/api')
    vi.mocked(viewsApi.getInteractions).mockResolvedValue(richInteractions)
    renderInteractionsTab()
    await waitFor(() => expect(screen.getByText(/Team Over/i)).toBeInTheDocument())
  })
})
