import { useState, useRef, useEffect } from 'react'
import { useModel } from '@/lib/model-context'
import { ModelRequired } from '@/components/ui/ModelRequired'
import { useAIEnabled } from '@/hooks/useAIEnabled'
import { advisorApi } from '@/services/api'
import { AlertTriangle, Bot } from 'lucide-react'
import { ChatMessage, type ChatEntry } from '@/components/advisor/ChatMessage'
import { QuickActions } from '@/components/advisor/QuickActions'
import { AdvisorInput } from '@/components/advisor/AdvisorInput'
import { PageHeader } from '@/components/ui/page-header'

export function AdvisorPage() {
  const { modelId, parseResult } = useModel()
  const aiEnabled = useAIEnabled()
  const [history, setHistory] = useState<ChatEntry[]>([])
  const [loading, setLoading] = useState(false)
  const [category, setCategory] = useState('general')
  const [aiConfigured, setAiConfigured] = useState<boolean | null>(null)
  const scrollRef = useRef<HTMLDivElement>(null)

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
      const resp = await advisorApi.ask(modelId, question, cat)
      setAiConfigured(resp.ai_configured)
      setHistory(prev => [...prev, { question, answer: resp.answer, category: resp.category, aiConfigured: resp.ai_configured }])
    } catch (err) {
      setHistory(prev => [...prev, { question, answer: `Error: ${err instanceof Error ? err.message : 'Request failed'}`, category: cat, aiConfigured: false }])
    } finally {
      setLoading(false)
    }
  }

  return (
    <ModelRequired>
      <div className="flex flex-col h-full max-w-3xl">
        <PageHeader
          title="AI Advisor"
          description="Ask questions about your architecture model"
        />

        {/* Model context & AI status */}
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

        <div className="flex-shrink-0 mb-4">
          <QuickActions onSelect={handleAsk} disabled={!modelId || loading || !aiEnabled} />
        </div>

        <div ref={scrollRef} className="flex-1 overflow-auto space-y-6 mb-4 min-h-0">
          {history.length === 0 && !loading && (
            <div className="flex flex-col items-center justify-center py-16 text-center">
              <Bot className="w-10 h-10 text-muted-foreground/40" />
              <p className="mt-3 text-sm text-muted-foreground">
                {modelId ? 'Ask a question or click a quick action to get started' : 'Load a model to start asking questions'}
              </p>
            </div>
          )}
          {history.map((entry, i) => <ChatMessage key={i} entry={entry} />)}
          {loading && (
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <span className="w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin" />
              Thinking...
            </div>
          )}
        </div>

        <div className="flex-shrink-0 pt-3 border-t border-border">
          <AdvisorInput onSend={handleAsk} disabled={!modelId || !aiEnabled} loading={loading} category={category} onCategoryChange={setCategory} />
        </div>
      </div>
    </ModelRequired>
  )
}
