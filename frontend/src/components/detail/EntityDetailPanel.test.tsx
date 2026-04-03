import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { EntityDetailPanel } from './EntityDetailPanel'

const baseTeam = {
  id: 'team-1',
  label: 'Core Team',
  nodeType: 'team' as const,
  data: {
    type: 'stream-aligned',
    description: 'Core platform team',
    is_overloaded: false,
    anti_patterns: [] as Array<{ code: string; message: string; severity: string }>,
  },
}

describe('EntityDetailPanel', () => {
  it('renders team node with type badge', () => {
    render(<EntityDetailPanel entity={baseTeam} onClose={vi.fn()} />)
    expect(screen.getByText('Core Team')).toBeInTheDocument()
    expect(screen.getByText('stream-aligned')).toBeInTheDocument()
    expect(screen.getByText('Core platform team')).toBeInTheDocument()
  })

  it('shows overloaded badge for overloaded team', () => {
    const team = { ...baseTeam, data: { ...baseTeam.data, is_overloaded: true } }
    render(<EntityDetailPanel entity={team} onClose={vi.fn()} />)
    expect(screen.getByText('Overloaded')).toBeInTheDocument()
  })

  it('renders capability node with visibility', () => {
    const cap = {
      id: 'cap-1',
      label: 'Search',
      nodeType: 'capability' as const,
      data: {
        visibility: 'core',
        is_fragmented: false,
        anti_patterns: [] as Array<{ code: string; message: string; severity: string }>,
      },
    }
    render(<EntityDetailPanel entity={cap} onClose={vi.fn()} />)
    expect(screen.getByText('Search')).toBeInTheDocument()
    expect(screen.getByText('core')).toBeInTheDocument()
  })

  it('shows AI insight when provided', () => {
    const insight = { explanation: 'This team is overloaded.', suggestion: 'Split responsibilities.' }
    render(<EntityDetailPanel entity={baseTeam} insight={insight} onClose={vi.fn()} />)
    expect(screen.getByText('This team is overloaded.')).toBeInTheDocument()
    expect(screen.getByText('Split responsibilities.')).toBeInTheDocument()
  })

  it('renders anti-patterns when present', () => {
    const team = {
      ...baseTeam,
      data: {
        ...baseTeam.data,
        anti_patterns: [{ code: 'AP-01', message: 'Too many services', severity: 'error' }],
      },
    }
    render(<EntityDetailPanel entity={team} onClose={vi.fn()} />)
    expect(screen.getByText('Too many services')).toBeInTheDocument()
  })

  it('calls onClose when panel closes', async () => {
    const onClose = vi.fn()
    render(<EntityDetailPanel entity={baseTeam} onClose={onClose} />)
    const closeBtn = screen.getByRole('button', { name: /close/i })
    await userEvent.click(closeBtn)
    expect(onClose).toHaveBeenCalled()
  })
})
