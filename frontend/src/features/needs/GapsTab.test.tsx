import { describe, it, expect, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { GapsView } from '@/types/views'

vi.mock('@/lib/model-context', () => ({
  useModel: () => ({ modelId: 'model-1', isHydrating: false }),
}))

vi.mock('@/services/api', () => ({
  viewsApi: {
    getGaps: vi.fn(),
  },
}))

import { GapsTab } from './GapsTab'

function renderGapsTab() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>
        <GapsTab />
      </MemoryRouter>
    </QueryClientProvider>
  )
}

const cleanGaps: GapsView = {
  model_id: 'model-1',
  unmapped_needs: [],
  unrealized_capabilities: [],
  unowned_services: [],
  unneeded_capabilities: [],
  orphan_services: [],
}

const gapsWithIssues: GapsView = {
  model_id: 'model-1',
  unmapped_needs: ['Need A', 'Need B'],
  unrealized_capabilities: ['Cap X'],
  unowned_services: ['Svc Y'],
  unneeded_capabilities: ['Cap Z'],
  orphan_services: ['Orphan Svc'],
}

describe('GapsTab', () => {
  it('shows empty state when no gaps', async () => {
    const { viewsApi } = await import('@/services/api')
    vi.mocked(viewsApi.getGaps).mockResolvedValue(cleanGaps)
    renderGapsTab()
    await waitFor(() => expect(screen.getByText(/no gaps/i)).toBeInTheDocument())
  })

  it('renders all 5 gap sections', async () => {
    const { viewsApi } = await import('@/services/api')
    vi.mocked(viewsApi.getGaps).mockResolvedValue(gapsWithIssues)
    renderGapsTab()
    await waitFor(() => expect(screen.getByText('Need A')).toBeInTheDocument())
    expect(screen.getByText('Cap X')).toBeInTheDocument()
    expect(screen.getByText('Svc Y')).toBeInTheDocument()
    expect(screen.getByText('Cap Z')).toBeInTheDocument()
    expect(screen.getByText('Orphan Svc')).toBeInTheDocument()
  })

  it('shows section headers', async () => {
    const { viewsApi } = await import('@/services/api')
    vi.mocked(viewsApi.getGaps).mockResolvedValue(gapsWithIssues)
    renderGapsTab()
    await waitFor(() => expect(screen.getByText(/Unmapped Needs/i)).toBeInTheDocument())
    expect(screen.getByText(/Unrealized Capabilities/i)).toBeInTheDocument()
    expect(screen.getByText(/Unowned Services/i)).toBeInTheDocument()
  })

  it('shows all clear for empty sections', async () => {
    const { viewsApi } = await import('@/services/api')
    vi.mocked(viewsApi.getGaps).mockResolvedValue({
      ...cleanGaps,
      unmapped_needs: ['Need A'],
    })
    renderGapsTab()
    await waitFor(() => expect(screen.getAllByText(/all clear/i).length).toBeGreaterThan(0))
  })
})
