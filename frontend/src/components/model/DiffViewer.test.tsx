import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, act, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { DiffViewer } from './DiffViewer'

vi.mock('@/services/api', () => ({
  modelsApi: {
    getDiff: vi.fn(),
  },
}))

const emptyDiff = {
  model_id: 'test-model',
  from_version: 1,
  to_version: 2,
  added: { actors: [], needs: [], capabilities: [], services: [], teams: [] },
  removed: { actors: [], needs: [], capabilities: [], services: [], teams: [] },
  changed: { actors: [], needs: [], capabilities: [], services: [], teams: [] },
}

function renderDiff(props = { modelId: 'test-model', fromVersion: 1, toVersion: 2 }) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={qc}>
      <DiffViewer {...props} />
    </QueryClientProvider>
  )
}

describe('DiffViewer', () => {
  beforeEach(() => { vi.clearAllMocks() })

  it('renders without crashing', async () => {
    const { modelsApi } = await import('@/services/api')
    vi.mocked(modelsApi.getDiff).mockResolvedValue(emptyDiff)
    await act(async () => { renderDiff() })
    expect(document.body).toBeTruthy()
  })

  it('shows version range label', async () => {
    const { modelsApi } = await import('@/services/api')
    vi.mocked(modelsApi.getDiff).mockResolvedValue(emptyDiff)
    renderDiff()
    await waitFor(() => expect(screen.getByText('v1')).toBeInTheDocument())
    expect(screen.getByText('v2')).toBeInTheDocument()
  })

  it('shows no changes message when diff is empty', async () => {
    const { modelsApi } = await import('@/services/api')
    vi.mocked(modelsApi.getDiff).mockResolvedValue(emptyDiff)
    renderDiff()
    await waitFor(() => expect(screen.getByText(/no changes/i)).toBeInTheDocument())
  })

  it('shows added entities with green styling', async () => {
    const { modelsApi } = await import('@/services/api')
    vi.mocked(modelsApi.getDiff).mockResolvedValue({
      ...emptyDiff,
      added: { actors: ['New Actor'], needs: [], capabilities: [], services: [], teams: ['Team B'] },
    })
    renderDiff()
    await waitFor(() => expect(screen.getByText('New Actor')).toBeInTheDocument())
    expect(screen.getByText('Team B')).toBeInTheDocument()
  })

  it('shows removed entities with red styling', async () => {
    const { modelsApi } = await import('@/services/api')
    vi.mocked(modelsApi.getDiff).mockResolvedValue({
      ...emptyDiff,
      removed: { actors: [], needs: [], capabilities: ['Old Cap'], services: [], teams: [] },
    })
    renderDiff()
    await waitFor(() => expect(screen.getByText('Old Cap')).toBeInTheDocument())
  })
})
