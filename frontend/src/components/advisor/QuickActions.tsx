interface QuickAction {
  label: string
  question: string
  category: string
}

const QUICK_ACTIONS: QuickAction[] = [
  { label: 'Summarize this architecture', question: 'Summarize this architecture', category: 'model-summary' },
  { label: "What's the biggest risk?", question: "What's the biggest risk?", category: 'health-summary' },
  { label: 'Which teams are overloaded?', question: 'Which teams are overloaded?', category: 'structural-load' },
  { label: 'Where are the bottlenecks?', question: 'Where are the bottlenecks?', category: 'bottleneck' },
  { label: 'How is our value delivery?', question: 'How is our value delivery?', category: 'value-stream' },
]

interface QuickActionsProps {
  onSelect: (question: string, category: string) => void
  disabled: boolean
}

export function QuickActions({ onSelect, disabled }: QuickActionsProps) {
  return (
    <div className="flex flex-wrap gap-2">
      {QUICK_ACTIONS.map(action => (
        <button
          key={action.label}
          disabled={disabled}
          onClick={() => onSelect(action.question, action.category)}
          type="button"
          className="px-3.5 py-2 font-semibold transition-all disabled:opacity-40 disabled:pointer-events-none"
          style={{
            fontSize: 12,
            borderRadius: 20,
            background: '#f0f0ff',
            border: '1px solid #e0e0ff',
            color: '#6366f1',
          }}
          onMouseEnter={e => {
            if (!disabled) {
              (e.currentTarget as HTMLElement).style.background = '#e8e8ff'
              ;(e.currentTarget as HTMLElement).style.borderColor = '#d0d0ff'
            }
          }}
          onMouseLeave={e => {
            (e.currentTarget as HTMLElement).style.background = '#f0f0ff'
            ;(e.currentTarget as HTMLElement).style.borderColor = '#e0e0ff'
          }}
        >
          {action.label}
        </button>
      ))}
    </div>
  )
}
