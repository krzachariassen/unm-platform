import { useState, useRef, useEffect, useCallback } from 'react'
import { useModel } from '@/lib/model-context'
import { ModelRequired } from '@/components/ui/ModelRequired'
import { useAIEnabled } from '@/hooks/useAIEnabled'
import { advisorApi } from '@/services/api'
import { AlertTriangle, Bot, RotateCcw } from 'lucide-react'
import { ChatMessage, type ChatEntry } from '@/components/advisor/ChatMessage'
import { QuickActions } from '@/components/advisor/QuickActions'
import { AdvisorInput } from '@/components/advisor/AdvisorInput'
import { ApplyActionsDialog } from '@/components/advisor/ApplyActionsDialog'
import { PageHeader } from '@/components/ui/page-header'

function buildConversationPrompt(history: ChatEntry[], newQuestion: string): string {
  if (history.length === 0) return newQuestion
  const context = history
    .map(e => `User: ${e.question}\nAdvisor: ${e.answer}`)
    .join('\n\n')
  return `Previous conversation:\n${context}\n\nUser: ${newQuestion}\n\nContinue the conversation. Answer the latest question, referencing prior context where relevant.`
}

export function AdvisorPage() {
  const { modelId, parseResult } = useModel()
  const aiEnabled = useAIEnabled()
  const [history, setHistory] = useState<ChatEntry[]>([])
  const [loading, setLoading] = useState(false)
  const [pendingQuestion, setPendingQuestion] = useState<string | null>(null)
  const [aiConfigured, setAiConfigured] = useState<boolean | null>(null)
  const scrollRef = useRef<HTMLDivElement>(null)
  const askingRef = useRef(false)

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [history, loading, pendingQuestion])

  const handleAsk = useCallback(async (question: string) => {
    if (!modelId || askingRef.current) return
    askingRef.current = true
    setPendingQuestion(question)
    setLoading(true)
    const prompt = buildConversationPrompt(history, question)
    try {
      const resp = await advisorApi.ask(modelId, prompt)
      setAiConfigured(resp.ai_configured)
      setHistory(prev => [...prev, { question, answer: resp.answer, aiConfigured: resp.ai_configured, routing: resp.routing }])
    } catch (err) {
      setHistory(prev => [...prev, { question, answer: `Error: ${err instanceof Error ? err.message : 'Request failed'}`, aiConfigured: false }])
    } finally {
      setLoading(false)
      setPendingQuestion(null)
      askingRef.current = false
    }
  }, [modelId, history])

  const handleNewConversation = useCallback(() => {
    setHistory([])
  }, [])

  const [applyDialogOpen, setApplyDialogOpen] = useState(false)
  const [applyResponse, setApplyResponse] = useState('')

  const handleApply = useCallback((answer: string) => {
    setApplyResponse(answer)
    setApplyDialogOpen(true)
  }, [])

  const handleRetryWithTier = useCallback(async (question: string, tier: string) => {
    if (!modelId || askingRef.current) return
    askingRef.current = true
    setPendingQuestion(question)
    setLoading(true)
    const prompt = buildConversationPrompt(
      history.filter(e => e.question !== question),
      question,
    )
    try {
      const resp = await advisorApi.ask(modelId, prompt, 'general', tier)
      setAiConfigured(resp.ai_configured)
      setHistory(prev => {
        const idx = [...prev].reverse().findIndex((e: ChatEntry) => e.question === question)
        const lastIdx = idx === -1 ? -1 : prev.length - 1 - idx
        if (lastIdx === -1) return [...prev, { question, answer: resp.answer, aiConfigured: resp.ai_configured, routing: resp.routing }]
        const updated = [...prev]
        updated[lastIdx] = { question, answer: resp.answer, aiConfigured: resp.ai_configured, routing: resp.routing }
        return updated
      })
    } catch (err) {
      setHistory(prev => {
        const idx = [...prev].reverse().findIndex((e: ChatEntry) => e.question === question)
        const lastIdx = idx === -1 ? -1 : prev.length - 1 - idx
        const entry: ChatEntry = { question, answer: `Error: ${err instanceof Error ? err.message : 'Request failed'}`, aiConfigured: false }
        if (lastIdx === -1) return [...prev, entry]
        const updated = [...prev]
        updated[lastIdx] = entry
        return updated
      })
    } finally {
      setLoading(false)
      setPendingQuestion(null)
      askingRef.current = false
    }
  }, [modelId, history])

  return (
    <ModelRequired>
      <div className="flex flex-col h-full max-w-screen-xl mx-auto">
        <div className="max-w-3xl flex flex-col h-full w-full">
        <PageHeader
          title="AI Advisor"
          description="Ask questions about your architecture model"
          actions={history.length > 0 ? (
            <button
              type="button"
              onClick={handleNewConversation}
              className="flex items-center gap-1.5 rounded-md px-3 py-1.5 text-xs font-medium transition-colors hover:bg-gray-100"
              style={{ color: '#6b7280', border: '1px solid #e5e7eb' }}
            >
              <RotateCcw size={12} /> New conversation
            </button>
          ) : undefined}
        />

        <div className="flex-shrink-0 mb-4 space-y-2">
          <div className="flex items-center gap-2 text-xs text-muted-foreground">
            <span className={`w-2 h-2 rounded-full ${modelId ? 'bg-green-500' : 'bg-gray-300'}`} />
            {modelId
              ? <span>Model loaded: <strong>{parseResult?.system_name}</strong></span>
              : <span>No model loaded — upload or load an example first</span>
            }
          </div>
          {!aiEnabled && (
            <div className="flex items-center gap-2 rounded-lg px-3 py-2 text-xs bg-amber-50 border border-amber-200 text-amber-800">
              <AlertTriangle className="w-3.5 h-3.5" />
              AI Advisor is not configured. Contact your administrator to enable AI features.
            </div>
          )}
          {aiEnabled && aiConfigured === false && (
            <div className="flex items-center gap-2 rounded-lg px-3 py-2 text-xs bg-amber-50 border border-amber-200 text-amber-800">
              <AlertTriangle className="w-3.5 h-3.5" />
              AI advisor not configured — set UNM_OPENAI_API_KEY to enable
            </div>
          )}
        </div>

        {history.length === 0 && (
          <div className="flex-shrink-0 mb-4">
            <QuickActions onSelect={handleAsk} disabled={!modelId || loading || !aiEnabled} />
          </div>
        )}

        <div ref={scrollRef} className="flex-1 overflow-auto space-y-6 mb-4 min-h-0">
          {history.length === 0 && !loading && (
            <div className="flex flex-col items-center justify-center py-16 text-center">
              <Bot className="w-10 h-10 text-muted-foreground/40" />
              <p className="mt-3 text-sm text-muted-foreground">
                {modelId ? 'Ask a question or click a quick action to get started' : 'Load a model to start asking questions'}
              </p>
            </div>
          )}
          {history.map((entry, i) => <ChatMessage key={i} entry={entry} onApply={handleApply} onRetryWithTier={handleRetryWithTier} />)}
          {loading && pendingQuestion && (
            <div className="space-y-4">
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
                    <span className="opacity-90" aria-hidden>You</span>
                  </div>
                  {pendingQuestion}
                </div>
              </div>
              <div className="flex justify-start gap-2">
                <div className="flex items-center gap-3 px-4 py-3 rounded-2xl rounded-bl-sm" style={{ background: '#ffffff', border: '1px solid #e2e8f0', boxShadow: '0 1px 3px rgba(0,0,0,0.06)' }}>
                  <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg" style={{ background: 'linear-gradient(135deg, #eef2ff 0%, #f5f3ff 100%)' }}>
                    <Bot size={14} className="animate-pulse" style={{ color: '#6366f1' }} />
                  </div>
                  <div className="flex items-center gap-1">
                    <span className="text-xs font-medium" style={{ color: '#6b7280' }}>Thinking</span>
                    <span className="flex gap-0.5 ml-0.5">
                      <span className="w-1 h-1 rounded-full animate-bounce" style={{ background: '#6366f1', animationDelay: '0ms', animationDuration: '1.2s' }} />
                      <span className="w-1 h-1 rounded-full animate-bounce" style={{ background: '#6366f1', animationDelay: '200ms', animationDuration: '1.2s' }} />
                      <span className="w-1 h-1 rounded-full animate-bounce" style={{ background: '#6366f1', animationDelay: '400ms', animationDuration: '1.2s' }} />
                    </span>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>

        <div className="flex-shrink-0 pt-3 border-t border-border">
          <AdvisorInput onSend={handleAsk} disabled={!modelId || !aiEnabled} loading={loading} />
        </div>
        </div>
      </div>

      <ApplyActionsDialog
        open={applyDialogOpen}
        onClose={() => setApplyDialogOpen(false)}
        advisorResponse={applyResponse}
      />
    </ModelRequired>
  )
}
