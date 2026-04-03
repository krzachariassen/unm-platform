import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, act } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { WhatIfPage } from './WhatIfPage'

vi.mock('@/lib/model-context', () => ({
  useModel: () => ({ modelId: 'test-model-id', isHydrating: false, parseResult: null }),
  useRequireModel: () => ({ modelId: 'test-model-id', isHydrating: false, parseResult: null }),
}))

vi.mock('@/hooks/useAIEnabled', () => ({ useAIEnabled: () => false }))

vi.mock('@/components/ui/ModelRequired', () => ({
  ModelRequired: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}))

vi.mock('@/features/whatif/AIWhatIfTab', () => ({
  AIWhatIfTab: () => <div data-testid="ai-whatif-tab">AI What-If Tab</div>,
}))

vi.mock('@/features/whatif/ManualWhatIfTab', () => ({
  ManualWhatIfTab: () => <div data-testid="manual-whatif-tab">Manual What-If Tab</div>,
}))

function renderWhatIfPage() {
  return render(
    <MemoryRouter>
      <WhatIfPage />
    </MemoryRouter>
  )
}

describe('WhatIfPage', () => {
  beforeEach(() => { vi.clearAllMocks() })

  it('renders without crashing', async () => {
    await act(async () => { renderWhatIfPage() })
    expect(document.body).toBeTruthy()
  })

  it('shows the page heading', async () => {
    await act(async () => { renderWhatIfPage() })
    expect(screen.getByText('What-If Explorer')).toBeInTheDocument()
  })

  it('shows the manual mode tab by default when AI is disabled', async () => {
    await act(async () => { renderWhatIfPage() })
    expect(screen.getByText('Manual Mode')).toBeInTheDocument()
  })

  it('renders the manual what-if tab when AI is disabled', async () => {
    await act(async () => { renderWhatIfPage() })
    expect(screen.getByTestId('manual-whatif-tab')).toBeInTheDocument()
  })
})
