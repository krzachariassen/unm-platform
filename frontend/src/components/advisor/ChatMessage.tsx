import { User, Bot, Download, Sparkles } from 'lucide-react'
import { Prose } from '@/components/ui/prose'
import { exportToPdf } from '@/lib/export-pdf'

export interface ChatEntry {
  question: string
  answer: string
  aiConfigured: boolean
}

interface ChatMessageProps {
  entry: ChatEntry
  onApply?: (answer: string) => void
}

export function ChatMessage({ entry, onApply }: ChatMessageProps) {
  return (
    <div className="space-y-4">
      {/* User question */}
      <div className="flex justify-end gap-2">
        <div
          className="max-w-[min(100%,28rem)] px-4 py-3 text-sm leading-relaxed"
          style={{
            background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)',
            color: '#ffffff',
            borderRadius: '20px 20px 4px 20px',
            boxShadow: '0 4px 14px rgba(99, 102, 241, 0.25)',
          }}
        >
          <div className="text-[10px] font-bold uppercase tracking-wider opacity-90 mb-1.5 flex items-center gap-1.5">
            <User size={12} strokeWidth={2.5} className="opacity-90" aria-hidden />
            You
          </div>
          {entry.question}
        </div>
      </div>

      {/* AI answer */}
      <div className="flex justify-start gap-2">
        <div
          className="flex max-w-full min-w-0 flex-1 flex-col px-4 py-3"
          style={{
            background: '#ffffff',
            border: '1px solid #e2e8f0',
            borderRadius: '20px 20px 20px 4px',
            boxShadow: '0 1px 3px rgba(0,0,0,0.06)',
          }}
        >
          <div className="flex items-center gap-2 mb-2">
            <div
              className="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg"
              style={{ background: 'linear-gradient(135deg, #eef2ff 0%, #f5f3ff 100%)' }}
            >
              <Bot size={14} style={{ color: '#6366f1' }} />
            </div>
            <span className="text-xs font-bold" style={{ color: '#6366f1' }}>Advisor</span>
            <div className="ml-auto flex items-center gap-1">
              {onApply && entry.aiConfigured && (
                <button
                  type="button"
                  onClick={() => onApply(entry.answer)}
                  title="Apply recommendations as changes"
                  className="flex items-center gap-1.5 rounded-md px-2.5 py-1 text-[11px] font-medium transition-all hover:opacity-90"
                  style={{ background: '#111827', color: '#ffffff' }}
                >
                  <Sparkles size={11} />
                  Apply
                </button>
              )}
              <button
                type="button"
                onClick={() => exportToPdf(entry.answer, `AI Advisor — ${entry.question.slice(0, 60)}`)}
                title="Save as PDF"
                className="rounded-md p-1.5 transition-colors hover:bg-gray-100"
                style={{ color: '#9ca3af' }}
              >
                <Download size={12} />
              </button>
            </div>
          </div>
          <div className="pl-0.5">
            <Prose>{entry.answer}</Prose>
          </div>
        </div>
      </div>
    </div>
  )
}
