import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

vi.mock('@/lib/model-context', () => ({
  useModel: () => ({ modelId: 'model-1', isHydrating: false, parseResult: null }),
}))
vi.mock('@/pages/views/NeedView', () => ({
  NeedView: () => <div>NeedView content</div>,
}))
vi.mock('@/pages/views/RealizationView', () => ({
  RealizationView: () => <div>RealizationView content</div>,
}))
vi.mock('@/components/ui/ModelRequired', () => ({
  ModelRequired: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}))

import { NeedsPage } from './NeedsPage'

function renderPage(url = '/needs') {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={[url]}>
        <NeedsPage />
      </MemoryRouter>
    </QueryClientProvider>
  )
}

describe('NeedsPage', () => {
  it('renders Overview tab by default', () => {
    renderPage()
    expect(screen.getByText('NeedView content')).toBeInTheDocument()
  })

  it('renders Traceability tab when ?tab=traceability', () => {
    renderPage('/needs?tab=traceability')
    expect(screen.getByText('RealizationView content')).toBeInTheDocument()
  })
})
