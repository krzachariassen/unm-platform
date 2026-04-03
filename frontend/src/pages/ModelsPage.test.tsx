import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, act, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ModelsPage } from './ModelsPage'

vi.mock('@/lib/model-context', () => ({
  useModel: () => ({ modelId: null, isHydrating: false, parseResult: null, setModel: vi.fn() }),
}))

vi.mock('@/services/api', () => ({
  modelsApi: {
    listModels: vi.fn(),
    loadStoredModel: vi.fn(),
  },
}))

function renderPage() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>
        <ModelsPage />
      </MemoryRouter>
    </QueryClientProvider>
  )
}

describe('ModelsPage', () => {
  beforeEach(() => { vi.clearAllMocks() })

  it('renders the page heading', async () => {
    const { modelsApi } = await import('@/services/api')
    vi.mocked(modelsApi.listModels).mockResolvedValue({ models: [], total: 0 })
    await act(async () => { renderPage() })
    expect(screen.getByText('Models')).toBeInTheDocument()
  })

  it('shows empty state when no models', async () => {
    const { modelsApi } = await import('@/services/api')
    vi.mocked(modelsApi.listModels).mockResolvedValue({ models: [], total: 0 })
    renderPage()
    await waitFor(() => expect(screen.getByText(/no models/i)).toBeInTheDocument())
  })

  it('shows model cards when models exist', async () => {
    const { modelsApi } = await import('@/services/api')
    vi.mocked(modelsApi.listModels).mockResolvedValue({
      models: [
        { id: 'model-1', name: 'Test System', created_at: '2026-04-03T10:00:00Z', version_count: 3 },
      ],
      total: 1,
    })
    renderPage()
    await waitFor(() => expect(screen.getByText('Test System')).toBeInTheDocument())
    expect(screen.getByText('3')).toBeInTheDocument()
  })

  it('shows a Load button per model card', async () => {
    const { modelsApi } = await import('@/services/api')
    vi.mocked(modelsApi.listModels).mockResolvedValue({
      models: [
        { id: 'model-1', name: 'My System', created_at: '2026-04-03T10:00:00Z', version_count: 1 },
      ],
      total: 1,
    })
    renderPage()
    await waitFor(() => expect(screen.getByRole('button', { name: /load/i })).toBeInTheDocument())
  })
})
