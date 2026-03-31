import { useState, useRef, useCallback, useEffect, useMemo } from 'react'
import { useModel } from '@/lib/model-context'
import { useAIEnabled } from '@/hooks/useAIEnabled'
import { advisorApi } from '@/services/api'
import { Send, Loader2, Sparkles, RotateCcw, AlertTriangle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Prose } from '@/components/ui/prose'

interface ConversationEntry { role: 'user' | 'ai'; content: string }

export function AIWhatIfTab({ modelId }: { modelId: string }) {
  const { parseResult } = useModel()
  const aiEnabled = useAIEnabled()
  const [question, setQuestion] = useState('')
  const [conversation, setConversation] = useState<ConversationEntry[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const chatEndRef = useRef<HTMLDivElement>(null)

  const suggestions = useMemo(() => {
    const chips: string[] = []
    const s = parseResult?.summary
    if (s) {
      if (s.teams > 3) chips.push('What if we consolidated teams to reduce coordination overhead?')
      if (s.services > s.teams * 3) chips.push('How would you reduce the number of services per team?')
      if (s.capabilities > 10) chips.push('Which capabilities should be consolidated?')
    }
    if ((parseResult?.validation?.warnings ?? []).length > 0) chips.push('How should we address the validation warnings?')
    if (chips.length === 0) {
      chips.push('What would happen if we split the largest team?', 'How could we reduce cross-team dependencies?',
        'Should we create a platform team for shared infrastructure?', 'How would you reduce cognitive load for the most overloaded team?')
    }
    return chips.slice(0, 4)
  }, [parseResult])

  useEffect(() => { chatEndRef.current?.scrollIntoView({ behavior: 'smooth' }) }, [conversation])

  const askQuestion = useCallback(async () => {
    const q = question.trim()
    if (!q || loading) return
    setQuestion('')
    setConversation(prev => [...prev, { role: 'user', content: q }])
    setLoading(true)
    setError(null)
    try {
      const resp = await advisorApi.ask(modelId, q, 'whatif-scenario')
      setConversation(prev => [...prev, { role: 'ai', content: resp.answer }])
    } catch (err) {
      setError(err instanceof Error ? err.message : 'AI request failed')
      setConversation(prev => [...prev, { role: 'ai', content: 'Failed to generate scenario. Please try again.' }])
    } finally {
      setLoading(false)
    }
  }, [modelId, question, loading])

  return (
    <div className="flex flex-col h-full min-h-[500px]">
      {!aiEnabled && (
        <div className="flex items-center gap-2 rounded-lg px-3 py-2 mb-4 bg-amber-50 border border-amber-200 text-xs text-amber-800">
          <AlertTriangle className="w-3.5 h-3.5 shrink-0" />
          AI features are not configured. Ask your administrator to set up the AI API key.
        </div>
      )}
      <div className="flex-1 overflow-auto space-y-4 pb-4">
        {conversation.length === 0 && !loading && (
          <div className="text-center py-16">
            <Sparkles className="w-8 h-8 text-primary mx-auto mb-3" />
            <h3 className="text-lg font-semibold text-foreground">AI What-If Scenarios</h3>
            <p className="text-sm mt-2 max-w-md mx-auto text-muted-foreground">
              Ask a natural language question about restructuring your architecture.
            </p>
            <div className="mt-6 flex flex-wrap gap-2 justify-center">
              {suggestions.map(s => (
                <button key={s} onClick={() => setQuestion(s)} disabled={!aiEnabled}
                  className="text-xs px-3 py-1.5 rounded-full bg-blue-50 text-blue-700 border border-blue-200 hover:bg-blue-100 transition-colors disabled:opacity-50">
                  {s}
                </button>
              ))}
            </div>
          </div>
        )}
        {conversation.map((entry, i) => (
          <div key={i} className={`flex ${entry.role === 'user' ? 'justify-end' : 'justify-start'}`}>
            <div className={`max-w-3xl rounded-xl px-4 py-3 ${entry.role === 'user' ? 'bg-blue-600 text-white' : 'w-full bg-white border border-border text-foreground'}`}>
              {entry.role === 'user' ? (
                <p className="text-sm">{entry.content}</p>
              ) : (
                <Prose>{entry.content}</Prose>
              )}
            </div>
          </div>
        ))}
        {loading && (
          <div className="flex justify-start">
            <div className="rounded-xl px-4 py-3 flex items-center gap-2 bg-muted border border-border">
              <Loader2 className="w-3.5 h-3.5 animate-spin text-primary" />
              <span className="text-sm text-muted-foreground">Analyzing scenario...</span>
            </div>
          </div>
        )}
        {error && <div className="rounded-lg p-3 bg-red-50 border border-red-200"><p className="text-xs text-red-700">{error}</p></div>}
        <div ref={chatEndRef} />
      </div>
      <div className="flex-shrink-0 pt-3 border-t border-border">
        <div className="flex gap-2">
          <input
            className="flex-1 rounded-lg border border-border px-4 py-2.5 text-sm bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
            placeholder="How would you restructure..."
            value={question}
            onChange={e => setQuestion(e.target.value)}
            onKeyDown={e => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); askQuestion() } }}
            disabled={loading || !aiEnabled}
          />
          <Button onClick={askQuestion} disabled={loading || !question.trim() || !aiEnabled} className="gap-2 px-4">
            {loading ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <Send className="w-3.5 h-3.5" />}
            Ask
          </Button>
          {conversation.length > 0 && (
            <Button variant="outline" size="icon" onClick={() => { setConversation([]); setError(null) }} title="New conversation">
              <RotateCcw className="w-3.5 h-3.5" />
            </Button>
          )}
        </div>
      </div>
    </div>
  )
}
