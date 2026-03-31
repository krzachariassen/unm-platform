import { useState } from 'react'
import { Send } from 'lucide-react'

interface AdvisorInputProps {
  onSend: (question: string) => void
  disabled: boolean
  loading: boolean
}

export function AdvisorInput({ onSend, disabled, loading }: AdvisorInputProps) {
  const [question, setQuestion] = useState('')

  const handleSubmit = () => {
    const trimmed = question.trim()
    if (!trimmed) return
    onSend(trimmed)
    setQuestion('')
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSubmit()
    }
  }

  return (
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
  )
}
