import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, act } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { AdvisorPage } from './AdvisorPage'

vi.mock('@/lib/model-context', () => ({
  useModel: () => ({
    modelId: 'test-model-id',
    isHydrating: false,
    parseResult: { system_name: 'Test System' },
  }),
  useRequireModel: () => ({
    modelId: 'test-model-id',
    isHydrating: false,
    parseResult: { system_name: 'Test System' },
  }),
}))

vi.mock('@/hooks/useAIEnabled', () => ({ useAIEnabled: () => true }))

vi.mock('@/components/ui/ModelRequired', () => ({
  ModelRequired: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}))

vi.mock('@/services/api', () => ({
  advisorApi: {
    ask: vi.fn().mockResolvedValue({ answer: 'AI response', ai_configured: true, routing: null }),
  },
}))

vi.mock('@/components/advisor/ChatMessage', () => ({
  ChatMessage: () => <div data-testid="chat-message">Chat Message</div>,
}))

vi.mock('@/components/advisor/QuickActions', () => ({
  QuickActions: () => <div data-testid="quick-actions">Quick Actions</div>,
}))

vi.mock('@/components/advisor/AdvisorInput', () => ({
  AdvisorInput: () => <div data-testid="advisor-input">Advisor Input</div>,
}))

vi.mock('@/components/advisor/ApplyActionsDialog', () => ({
  ApplyActionsDialog: () => null,
}))

function renderAdvisorPage() {
  return render(
    <MemoryRouter>
      <AdvisorPage />
    </MemoryRouter>
  )
}

describe('AdvisorPage', () => {
  beforeEach(() => { vi.clearAllMocks() })

  it('renders without crashing', async () => {
    await act(async () => { renderAdvisorPage() })
    expect(document.body).toBeTruthy()
  })

  it('shows the page heading', async () => {
    await act(async () => { renderAdvisorPage() })
    expect(screen.getByText('AI Advisor')).toBeInTheDocument()
  })

  it('shows the advisor input', async () => {
    await act(async () => { renderAdvisorPage() })
    expect(screen.getByTestId('advisor-input')).toBeInTheDocument()
  })

  it('shows model status when model is loaded', async () => {
    await act(async () => { renderAdvisorPage() })
    expect(screen.getByText(/Test System/)).toBeInTheDocument()
  })

  it('shows quick actions when conversation is empty', async () => {
    await act(async () => { renderAdvisorPage() })
    expect(screen.getByTestId('quick-actions')).toBeInTheDocument()
  })
})
