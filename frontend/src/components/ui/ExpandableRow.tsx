import { useState } from 'react'
import { ChevronDown, ChevronUp, Info, Lightbulb } from 'lucide-react'

export function ExpandableRow({ summary, explanation, suggestion }: {
  summary: React.ReactNode
  explanation: string
  suggestion: string
}) {
  const [open, setOpen] = useState(false)
  return (
    <div className="rounded-lg overflow-hidden" style={{ border: '1px solid #f0f0f0' }}>
      <button
        onClick={() => setOpen(o => !o)}
        className="w-full flex items-center justify-between gap-3 px-4 py-3 text-left transition-colors"
        style={{ background: open ? '#f8fafc' : '#f9fafb' }}
        onMouseEnter={e => { if (!open) (e.currentTarget as HTMLElement).style.background = '#f1f5f9' }}
        onMouseLeave={e => { if (!open) (e.currentTarget as HTMLElement).style.background = '#f9fafb' }}
      >
        <div className="flex-1 min-w-0">{summary}</div>
        <div className="flex-shrink-0 ml-2" style={{ color: '#9ca3af' }}>
          {open ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
        </div>
      </button>

      {open && (
        <div className="px-4 pb-4 pt-2 space-y-3" style={{ background: '#f8fafc', borderTop: '1px solid #e2e8f0' }}>
          <div className="flex gap-2.5">
            <Info size={13} className="flex-shrink-0 mt-0.5" style={{ color: '#64748b' }} />
            <p className="text-xs leading-relaxed" style={{ color: '#475569' }}>{explanation}</p>
          </div>
          <div className="flex gap-2.5">
            <Lightbulb size={13} className="flex-shrink-0 mt-0.5" style={{ color: '#2563eb' }} />
            <p className="text-xs leading-relaxed" style={{ color: '#1e40af' }}>{suggestion}</p>
          </div>
        </div>
      )}
    </div>
  )
}
