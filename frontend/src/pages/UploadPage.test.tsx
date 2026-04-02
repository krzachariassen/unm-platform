import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, act, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { UploadPage } from './UploadPage'

vi.mock('@/lib/model-context', () => ({
  useModel: () => ({
    modelId: null,
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

vi.mock('@/lib/runtimeConfig', () => ({
  getRuntimeConfig: vi.fn().mockResolvedValue({ features: { debug_routes: false } }),
}))

vi.mock('@/services/api', () => ({
  modelsApi: {
    parseModel: vi.fn(),
    loadExample: vi.fn(),
  },
  insightsApi: {
    getInsightsStatus: vi.fn(),
  },
}))

function makeClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } })
}

function renderUploadPage() {
  return render(
    <QueryClientProvider client={makeClient()}>
      <MemoryRouter>
        <UploadPage />
      </MemoryRouter>
    </QueryClientProvider>
  )
}

describe('UploadPage', () => {
  beforeEach(() => { vi.clearAllMocks() })

  it('renders without crashing', async () => {
    await act(async () => { renderUploadPage() })
    expect(document.body).toBeTruthy()
  })

  it('renders the page header', async () => {
    renderUploadPage()
    await waitFor(() => {
      expect(screen.getByText('Upload Model')).toBeInTheDocument()
    })
  })

  it('renders the drop zone', async () => {
    renderUploadPage()
    await waitFor(() => {
      expect(screen.getByText(/drop your .unm.yaml or .unm file/i)).toBeInTheDocument()
    })
  })

  it('renders the file input', async () => {
    renderUploadPage()
    await waitFor(() => {
      expect(document.getElementById('file-input')).toBeInTheDocument()
    })
  })
})
