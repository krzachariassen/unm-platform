import { useState, useRef, useEffect } from 'react'
import { useModel } from '@/lib/model-context'
import { useAIEnabled } from '@/hooks/useAIEnabled'
import { api } from '@/lib/api'
import { AlertTriangle, Bot } from 'lucide-react'
import { ChatMessage, type ChatEntry } from '@/components/advisor/ChatMessage'
import { QuickActions } from '@/components/advisor/QuickActions'
import { AdvisorInput } from '@/components/advisor/AdvisorInput'

export function AdvisorPage() {
  const { modelId, parseResult } = useModel()
  const aiEnabled = useAIEnabled()
  const [history, setHistory] = useState<ChatEntry[]>([])
  const [loading, setLoading] = useState(false)
  const [category, setCategory] = useState('general')
  const [aiConfigured, setAiConfigured] = useState<boolean | null>(null)
  const scrollRef = useRef<HTMLDivElement>(null)

  const hasModel = !!modelId

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [history, loading])

  const handleAsk = async (question: string, cat: string) => {
    if (!modelId) return
    setCategory(cat)
    setLoading(true)
    try {
      const resp = await api.askAdvisor(modelId, question, cat)
      setAiConfigured(resp.ai_configured)
      setHistory(prev => [...prev, {
        question,
        answer: resp.answer,
        category: resp.category,
        aiConfigured: resp.ai_configured,
      }])
    } catch (err) {
      setHistory(prev => [...prev, {
        question,
        answer: `Error: ${err instanceof Error ? err.message : 'Request failed'}`,
        category: cat,
        aiConfigured: false,
      }])
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex flex-col h-full max-w-3xl">
      {/* Header */}
      <div className="flex-shrink-0 mb-4">
        <div className="flex items-center gap-2">
          <div className="rounded-lg p-2" style={{ background: '#dbeafe' }}>
            <Bot size={18} style={{ color: '#2563eb' }} />
          </div>
          <div>
            <h1 className="text-lg font-semibold tracking-tight" style={{ color: '#111827' }}>AI Advisor</h1>
            <p className="text-xs" style={{ color: '#6b7280' }}>
              Ask questions about your architecture model
            </p>
          </div>
        </div>

        {/* Model context indicator */}
        <div className="mt-3 flex items-center gap-2 text-xs" style={{ color: hasModel ? '#374151' : '#9ca3af' }}>
          <span className="w-2 h-2 rounded-full" style={{ background: hasModel ? '#22c55e' : '#d1d5db' }} />
          {hasModel
            ? <span>Model loaded: <strong>{parseResult?.system_name}</strong></span>
            : <span>No model loaded — upload or load an example first</span>
          }
        </div>

        {/* AI not configured banner — proactive check */}
        {!aiEnabled && (
          <div
            className="mt-3 flex items-center gap-2 rounded-lg px-3 py-2 text-xs"
            style={{ background: '#fef3c7', border: '1px solid #fcd34d', color: '#92400e' }}
          >
            <span title="AI not configured" aria-label="AI not configured" style={{ cursor: 'help' }}>
              <AlertTriangle size={14} />
            </span>
            AI Advisor is not configured. Contact your administrator to enable AI features.
          </div>
        )}

        {/* AI not configured banner — from API response */}
        {aiEnabled && aiConfigured === false && (
          <div
            className="mt-3 flex items-center gap-2 rounded-lg px-3 py-2 text-xs"
            style={{ background: '#fffbeb', border: '1px solid #fde68a', color: '#92400e' }}
          >
            <AlertTriangle size={14} />
            AI advisor not configured — set UNM_OPENAI_API_KEY to enable
          </div>
        )}
      </div>

      {/* Quick actions */}
      <div className="flex-shrink-0 mb-4">
        <QuickActions onSelect={handleAsk} disabled={!hasModel || loading || !aiEnabled} />
      </div>

      {/* Chat history */}
      <div
        ref={scrollRef}
        className="flex-1 overflow-auto space-y-6 mb-4"
        style={{ minHeight: 0 }}
      >
        {history.length === 0 && !loading && (
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <Bot size={40} style={{ color: '#d1d5db' }} />
            <p className="mt-3 text-sm" style={{ color: '#9ca3af' }}>
              {hasModel
                ? 'Ask a question or click a quick action to get started'
                : 'Load a model to start asking questions'
              }
            </p>
          </div>
        )}

        {history.map((entry, i) => (
          <ChatMessage key={i} entry={entry} />
        ))}

        {/* Loading indicator */}
        {loading && (
          <div className="flex items-center gap-2 text-sm" style={{ color: '#6b7280' }}>
            <span className="w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin" />
            Thinking...
          </div>
        )}
      </div>

      {/* Input area */}
      <div className="flex-shrink-0 pt-3" style={{ borderTop: '1px solid #e5e7eb' }}>
        <AdvisorInput
          onSend={handleAsk}
          disabled={!hasModel || !aiEnabled}
          loading={loading}
          category={category}
          onCategoryChange={setCategory}
        />
      </div>
    </div>
  )
}
