import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

vi.mock('@/lib/model-context', () => ({
  useModel: () => ({ modelId: 'model-1', isHydrating: false, parseResult: null }),
}))
vi.mock('@/pages/views/CapabilityView', () => ({
  CapabilityView: () => <div>CapabilityView content</div>,
}))
vi.mock('@/pages/views/RealizationView', () => ({
  RealizationView: () => <div>RealizationView content</div>,
}))
vi.mock('@/components/ui/ModelRequired', () => ({
  ModelRequired: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}))

import { CapabilitiesPage } from './CapabilitiesPage'

function renderPage(url = '/capabilities') {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={[url]}>
        <CapabilitiesPage />
      </MemoryRouter>
    </QueryClientProvider>
  )
}

describe('CapabilitiesPage', () => {
  it('renders Hierarchy tab by default', () => {
    renderPage()
    expect(screen.getByText('CapabilityView content')).toBeInTheDocument()
  })

  it('renders Services tab when ?tab=services', () => {
    renderPage('/capabilities?tab=services')
    expect(screen.getByText('RealizationView content')).toBeInTheDocument()
  })
})
