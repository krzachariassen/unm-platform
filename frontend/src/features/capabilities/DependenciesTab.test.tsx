import { describe, it, expect, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { DependenciesView } from '@/types/views'

vi.mock('@/lib/model-context', () => ({
  useModel: () => ({ modelId: 'model-1', isHydrating: false }),
}))

vi.mock('@/services/api', () => ({
  viewsApi: {
    getDependencies: vi.fn(),
  },
}))

import { DependenciesTab } from './DependenciesTab'

function renderDepsTab() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>
        <DependenciesTab />
      </MemoryRouter>
    </QueryClientProvider>
  )
}

const cleanDeps: DependenciesView = {
  model_id: 'model-1',
  service_cycles: [],
  capability_cycles: [],
  max_service_depth: 2,
  max_capability_depth: 1,
  critical_service_path: [],
}

const depsWithCycles: DependenciesView = {
  model_id: 'model-1',
  service_cycles: [{ path: ['svc-a', 'svc-b', 'svc-a'] }],
  capability_cycles: [],
  max_service_depth: 3,
  max_capability_depth: 2,
  critical_service_path: ['svc-x', 'svc-y', 'svc-z'],
}

describe('DependenciesTab', () => {
  it('shows no-cycles empty state when clean', async () => {
    const { viewsApi } = await import('@/services/api')
    vi.mocked(viewsApi.getDependencies).mockResolvedValue(cleanDeps)
    renderDepsTab()
    await waitFor(() => expect(screen.getByText(/no dependency cycles/i)).toBeInTheDocument())
  })

  it('renders stat cards for max depths', async () => {
    const { viewsApi } = await import('@/services/api')
    vi.mocked(viewsApi.getDependencies).mockResolvedValue(cleanDeps)
    renderDepsTab()
    await waitFor(() => expect(screen.getByText(/Max Service Depth/i)).toBeInTheDocument())
    expect(screen.getByText(/Max Capability Depth/i)).toBeInTheDocument()
  })

  it('renders cycle paths when cycles exist', async () => {
    const { viewsApi } = await import('@/services/api')
    vi.mocked(viewsApi.getDependencies).mockResolvedValue(depsWithCycles)
    renderDepsTab()
    await waitFor(() => expect(screen.getAllByText(/svc-a/i).length).toBeGreaterThan(0))
  })

  it('renders critical path chain', async () => {
    const { viewsApi } = await import('@/services/api')
    vi.mocked(viewsApi.getDependencies).mockResolvedValue(depsWithCycles)
    renderDepsTab()
    await waitFor(() => expect(screen.getByText(/svc-x/i)).toBeInTheDocument())
    expect(screen.getByText(/svc-z/i)).toBeInTheDocument()
  })
})
