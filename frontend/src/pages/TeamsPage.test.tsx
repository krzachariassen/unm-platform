import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

vi.mock('@/lib/model-context', () => ({
  useModel: () => ({ modelId: 'model-1', isHydrating: false, parseResult: null }),
}))
vi.mock('@/pages/views/TeamTopologyView', () => ({
  TeamTopologyView: () => <div>TeamTopologyView content</div>,
}))
vi.mock('@/pages/views/OwnershipView', () => ({
  OwnershipView: () => <div>OwnershipView content</div>,
}))
vi.mock('@/pages/views/CognitiveLoadView', () => ({
  CognitiveLoadView: () => <div>CognitiveLoadView content</div>,
}))
vi.mock('@/components/ui/ModelRequired', () => ({
  ModelRequired: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}))

import { TeamsPage } from './TeamsPage'

function renderPage(url = '/teams') {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={[url]}>
        <TeamsPage />
      </MemoryRouter>
    </QueryClientProvider>
  )
}

describe('TeamsPage', () => {
  it('renders Topology tab by default', () => {
    renderPage()
    expect(screen.getByText('TeamTopologyView content')).toBeInTheDocument()
  })

  it('renders Ownership tab when ?tab=ownership', () => {
    renderPage('/teams?tab=ownership')
    expect(screen.getByText('OwnershipView content')).toBeInTheDocument()
  })

  it('renders Cognitive Load tab when ?tab=cognitive-load', () => {
    renderPage('/teams?tab=cognitive-load')
    expect(screen.getByText('CognitiveLoadView content')).toBeInTheDocument()
  })

  it('shows all three tab buttons', () => {
    renderPage()
    expect(screen.getByRole('button', { name: 'Topology' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Ownership' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Cognitive Load' })).toBeInTheDocument()
  })
})
