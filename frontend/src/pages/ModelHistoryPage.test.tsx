import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, act, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ModelHistoryPage } from './ModelHistoryPage'

vi.mock('@/lib/model-context', () => ({
  useModel: () => ({ modelId: 'test-model-id', isHydrating: false, parseResult: null }),
  useRequireModel: () => ({ modelId: 'test-model-id', isHydrating: false, parseResult: null }),
}))

vi.mock('@/components/ui/ModelRequired', () => ({
  ModelRequired: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}))

vi.mock('@/services/api', () => ({
  modelsApi: {
    getHistory: vi.fn(),
    getDiff: vi.fn(),
  },
}))

function renderPage() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>
        <ModelHistoryPage />
      </MemoryRouter>
    </QueryClientProvider>
  )
}

describe('ModelHistoryPage', () => {
  beforeEach(() => { vi.clearAllMocks() })

  it('renders without crashing', async () => {
    const { modelsApi } = await import('@/services/api')
    vi.mocked(modelsApi.getHistory).mockResolvedValue({ model_id: 'test-model-id', versions: [] })
    await act(async () => { renderPage() })
    expect(document.body).toBeTruthy()
  })

  it('shows the page heading', async () => {
    const { modelsApi } = await import('@/services/api')
    vi.mocked(modelsApi.getHistory).mockResolvedValue({ model_id: 'test-model-id', versions: [] })
    await act(async () => { renderPage() })
    expect(screen.getByText('Version History')).toBeInTheDocument()
  })

  it('shows version entries when history exists', async () => {
    const { modelsApi } = await import('@/services/api')
    vi.mocked(modelsApi.getHistory).mockResolvedValue({
      model_id: 'test-model-id',
      versions: [
        { id: 'v1', version: 1, commit_message: 'Initial commit', committed_at: '2026-04-03T10:00:00Z' },
        { id: 'v2', version: 2, commit_message: 'Added Team B', committed_at: '2026-04-03T11:00:00Z' },
      ],
    })
    renderPage()
    await waitFor(() => expect(screen.getByText('Initial commit')).toBeInTheDocument())
    expect(screen.getByText('Added Team B')).toBeInTheDocument()
  })

  it('shows empty state when no versions', async () => {
    const { modelsApi } = await import('@/services/api')
    vi.mocked(modelsApi.getHistory).mockResolvedValue({ model_id: 'test-model-id', versions: [] })
    renderPage()
    await waitFor(() => expect(screen.getByText(/no version history/i)).toBeInTheDocument())
  })
})
