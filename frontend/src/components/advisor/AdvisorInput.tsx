import { useState } from 'react'
import { Send } from 'lucide-react'

const CATEGORIES = [
  { value: 'general', label: 'General', group: 'Overview' },
  { value: 'model-summary', label: 'Model Summary', group: 'Overview' },
  { value: 'health-summary', label: 'Health Summary', group: 'Overview' },
  { value: 'structural-load', label: 'Structural Load', group: 'Teams' },
  { value: 'team-boundary', label: 'Team Boundary', group: 'Teams' },
  { value: 'interaction-mode', label: 'Interaction Mode', group: 'Teams' },
  { value: 'service-placement', label: 'Service Placement', group: 'Architecture' },
  { value: 'fragmentation', label: 'Fragmentation', group: 'Architecture' },
  { value: 'bottleneck', label: 'Bottleneck', group: 'Architecture' },
  { value: 'coupling', label: 'Coupling', group: 'Architecture' },
  { value: 'value-stream', label: 'Value Stream', group: 'Delivery' },
  { value: 'need-delivery-risk', label: 'Need Delivery Risk', group: 'Delivery' },
  { value: 'natural-language', label: 'Free-form', group: 'Other' },
]

const CATEGORY_GROUPS = ['Overview', 'Teams', 'Architecture', 'Delivery', 'Other'] as const

interface AdvisorInputProps {
  onSend: (question: string, category: string) => void
  disabled: boolean
  loading: boolean
  category: string
  onCategoryChange: (category: string) => void
}

export function AdvisorInput({ onSend, disabled, loading, category, onCategoryChange }: AdvisorInputProps) {
  const [question, setQuestion] = useState('')

  const handleSubmit = () => {
    const trimmed = question.trim()
    if (!trimmed) return
    onSend(trimmed, category)
    setQuestion('')
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSubmit()
    }
  }

  return (
    <div className="space-y-3">
      {/* Category selector */}
      <div className="flex items-center gap-2 flex-wrap">
        <label className="text-xs font-semibold" style={{ color: '#64748b' }}>Category</label>
        <select
          value={category}
          onChange={e => onCategoryChange(e.target.value)}
          disabled={disabled}
          className="text-xs px-3 py-2 disabled:opacity-40 outline-none transition-shadow focus:outline-none"
          style={{
            borderRadius: 14,
            background: '#ffffff',
            border: '1px solid #e2e8f0',
            color: '#475569',
            fontSize: 14,
          }}
          onFocus={(e) => {
            e.target.style.borderColor = '#6366f1'
            e.target.style.boxShadow = '0 0 0 3px rgba(99,102,241,0.1)'
          }}
          onBlur={(e) => {
            e.target.style.borderColor = '#e2e8f0'
            e.target.style.boxShadow = 'none'
          }}
        >
          {CATEGORY_GROUPS.map(group => (
            <optgroup key={group} label={group}>
              {CATEGORIES.filter(c => c.group === group).map(c => (
                <option key={c.value} value={c.value}>{c.label}</option>
              ))}
            </optgroup>
          ))}
        </select>
      </div>

      {/* Input area */}
      <div className="flex gap-2 items-stretch">
        <textarea
          value={question}
          onChange={e => setQuestion(e.target.value)}
          onKeyDown={handleKeyDown}
          disabled={disabled || loading}
          placeholder={disabled ? 'Load a model first...' : 'Ask about your architecture...'}
          rows={2}
          className="flex-1 resize-none disabled:opacity-40 outline-none transition-shadow focus:outline-none"
          style={{
            borderRadius: 14,
            border: '1px solid #e2e8f0',
            padding: '12px 16px',
            fontSize: 14,
            color: '#1e293b',
            background: '#ffffff',
          }}
          onFocus={(e) => {
            e.target.style.borderColor = '#6366f1'
            e.target.style.boxShadow = '0 0 0 3px rgba(99,102,241,0.1)'
          }}
          onBlur={(e) => {
            e.target.style.borderColor = '#e2e8f0'
            e.target.style.boxShadow = 'none'
          }}
        />
        <button
          type="button"
          onClick={handleSubmit}
          disabled={disabled || loading || !question.trim()}
          className="inline-flex shrink-0 self-end items-center justify-center gap-2 min-w-[52px] min-h-[52px] disabled:opacity-45 disabled:cursor-not-allowed transition-opacity hover:opacity-95 px-4"
          style={{
            background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)',
            color: '#ffffff',
            borderRadius: 12,
            fontWeight: 600,
            border: 'none',
            cursor: 'pointer',
            boxShadow: '0 4px 14px rgba(99, 102, 241, 0.3)',
          }}
          aria-label="Send message"
        >
          {loading
            ? (
                <span
                  className="h-5 w-5 rounded-full animate-spin inline-block"
                  style={{ border: '2px solid #e2e8f0', borderTopColor: '#ffffff' }}
                  aria-hidden
                />
              )
            : <Send size={18} strokeWidth={2} />}
        </button>
      </div>
    </div>
  )
}
