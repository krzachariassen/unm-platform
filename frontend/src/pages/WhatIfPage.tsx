import { useState, useRef, useCallback, useEffect, useMemo } from 'react'
import { useModel, useRequireModel } from '@/lib/model-context'
import { useAIEnabled } from '@/hooks/useAIEnabled'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Send, Loader2, Sparkles, Wrench, RotateCcw, AlertTriangle } from 'lucide-react'
import { ActionForm } from '@/components/changeset/ActionForm'
import { ActionList } from '@/components/changeset/ActionList'
import { ImpactPanel } from '@/components/changeset/ImpactPanel'
import { api } from '@/lib/api'
import type { ChangeAction, ImpactDelta } from '@/lib/api'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'

interface ConversationEntry {
  role: 'user' | 'ai'
  content: string
}

function AIWhatIfTab({ modelId, aiEnabled }: { modelId: string; aiEnabled: boolean }) {
  const { parseResult } = useModel()
  const [question, setQuestion] = useState('')
  const [conversation, setConversation] = useState<ConversationEntry[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const chatEndRef = useRef<HTMLDivElement>(null)

  const suggestions = useMemo(() => {
    const chips: string[] = []
    const summary = parseResult?.summary
    if (summary) {
      if (summary.teams > 3) {
        chips.push('What if we consolidated teams to reduce coordination overhead?')
      }
      if (summary.services > summary.teams * 3) {
        chips.push('How would you reduce the number of services per team?')
      }
      if (summary.capabilities > 10) {
        chips.push('Which capabilities should be consolidated?')
      }
    }
    const warnings = parseResult?.validation?.warnings ?? []
    if (warnings.length > 0) {
      chips.push('How should we address the validation warnings?')
    }
    // Fallback chips
    if (chips.length === 0) {
      chips.push('What would happen if we split the largest team?')
      chips.push('How could we reduce cross-team dependencies?')
      chips.push('Should we create a platform team for shared infrastructure?')
      chips.push('How would you reduce cognitive load for the most overloaded team?')
    }
    return chips.slice(0, 4)
  }, [parseResult])

  useEffect(() => {
    chatEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [conversation])

  const askQuestion = useCallback(async () => {
    const q = question.trim()
    if (!q || loading) return
    setQuestion('')
    setConversation(prev => [...prev, { role: 'user', content: q }])
    setLoading(true)
    setError(null)
    try {
      const resp = await api.askAdvisor(modelId, q, 'whatif-scenario')
      setConversation(prev => [...prev, { role: 'ai', content: resp.answer }])
    } catch (err) {
      setError(err instanceof Error ? err.message : 'AI request failed')
      setConversation(prev => [...prev, { role: 'ai', content: 'Failed to generate scenario. Please try again.' }])
    } finally {
      setLoading(false)
    }
  }, [modelId, question, loading])

  return (
    <div className="flex flex-col h-full" style={{ minHeight: 500 }}>
      {/* AI not configured banner */}
      {!aiEnabled && (
        <div style={{
          background: '#fef3c7', border: '1px solid #fcd34d',
          borderRadius: 8, padding: '10px 14px', marginBottom: 16,
          display: 'flex', alignItems: 'center', gap: 8,
        }}>
          <span title="AI not configured" aria-label="AI not configured" style={{ cursor: 'help' }}>
            <AlertTriangle size={14} style={{ color: '#92400e' }} />
          </span>
          <span style={{ fontSize: 13, color: '#92400e' }}>
            AI features are not configured. Ask your administrator to set up the AI API key.
          </span>
        </div>
      )}
      {/* Conversation area */}
      <div className="flex-1 overflow-auto space-y-4 pb-4">
        {conversation.length === 0 && !loading && (
          <div className="text-center py-16">
            <Sparkles size={32} style={{ color: '#2563eb', margin: '0 auto 12px' }} />
            <h3 className="text-lg font-semibold" style={{ color: '#111827' }}>AI What-If Scenarios</h3>
            <p className="text-sm mt-2 max-w-md mx-auto" style={{ color: '#6b7280' }}>
              Ask a natural language question about restructuring your architecture. The AI will analyze the full model and propose specific changes with impact assessment.
            </p>
            <div className="mt-6 flex flex-wrap gap-2 justify-center">
              {suggestions.map(suggestion => (
                <button key={suggestion} onClick={() => { setQuestion(suggestion) }}
                  className="text-xs px-3 py-1.5 rounded-full"
                  style={{ background: '#eff6ff', color: '#1d4ed8', border: '1px solid #bfdbfe' }}
                  disabled={!aiEnabled}>
                  {suggestion}
                </button>
              ))}
            </div>
          </div>
        )}

        {conversation.map((entry, i) => (
          <div key={i} className={`flex ${entry.role === 'user' ? 'justify-end' : 'justify-start'}`}>
            <div className={`max-w-3xl rounded-xl px-4 py-3 ${entry.role === 'user' ? '' : 'w-full'}`}
              style={{
                background: entry.role === 'user' ? '#2563eb' : '#ffffff',
                color: entry.role === 'user' ? '#ffffff' : '#111827',
                border: entry.role === 'ai' ? '1px solid #e5e7eb' : undefined,
              }}>
              {entry.role === 'user' ? (
                <p className="text-sm">{entry.content}</p>
              ) : (
                <div className="prose prose-sm max-w-none">
                  <ReactMarkdown
                    remarkPlugins={[remarkGfm]}
                    components={{
                      h2: ({ children }) => <h2 className="text-base font-semibold mt-4 mb-2 pb-1" style={{ borderBottom: '1px solid #e5e7eb', color: '#1d4ed8' }}>{children}</h2>,
                      h3: ({ children }) => <h3 className="text-sm font-semibold mt-3 mb-1" style={{ color: '#111827' }}>{children}</h3>,
                      strong: ({ children }) => <strong style={{ color: '#111827' }}>{children}</strong>,
                      li: ({ children }) => <li className="text-sm leading-relaxed" style={{ color: '#374151' }}>{children}</li>,
                      p: ({ children }) => <p className="text-sm leading-relaxed mb-2" style={{ color: '#374151' }}>{children}</p>,
                      table: ({ children }) => <table className="w-full text-xs border-collapse my-3">{children}</table>,
                      th: ({ children }) => <th className="text-left px-2 py-1.5 font-semibold" style={{ background: '#f3f4f6', borderBottom: '2px solid #d1d5db', color: '#111827' }}>{children}</th>,
                      td: ({ children }) => <td className="px-2 py-1.5" style={{ borderBottom: '1px solid #e5e7eb', color: '#374151' }}>{children}</td>,
                    }}
                  >{entry.content}</ReactMarkdown>
                </div>
              )}
            </div>
          </div>
        ))}

        {loading && (
          <div className="flex justify-start">
            <div className="rounded-xl px-4 py-3 flex items-center gap-2" style={{ background: '#f9fafb', border: '1px solid #e5e7eb' }}>
              <Loader2 size={14} className="animate-spin" style={{ color: '#2563eb' }} />
              <span className="text-sm" style={{ color: '#6b7280' }}>Analyzing scenario...</span>
            </div>
          </div>
        )}

        {error && (
          <div className="rounded-lg p-3" style={{ border: '1px solid #fca5a5', background: '#fef2f2' }}>
            <p className="text-xs" style={{ color: '#b91c1c' }}>{error}</p>
          </div>
        )}

        <div ref={chatEndRef} />
      </div>

      {/* Input area */}
      <div className="flex-shrink-0 pt-3" style={{ borderTop: '1px solid #e5e7eb' }}>
        <div className="flex gap-2">
          <input
            className="flex-1 rounded-lg border px-4 py-2.5 text-sm"
            style={{ borderColor: '#d1d5db', background: '#ffffff', color: '#111827' }}
            placeholder="How would you restructure..."
            value={question}
            onChange={e => setQuestion(e.target.value)}
            onKeyDown={e => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); askQuestion() } }}
            disabled={loading || !aiEnabled}
          />
          <Button onClick={askQuestion} disabled={loading || !question.trim() || !aiEnabled} className="gap-2 px-4">
            {loading ? <Loader2 size={14} className="animate-spin" /> : <Send size={14} />}
            Ask
          </Button>
          {conversation.length > 0 && (
            <Button variant="outline" size="icon" onClick={() => { setConversation([]); setError(null) }} title="New conversation">
              <RotateCcw size={14} />
            </Button>
          )}
        </div>
      </div>
    </div>
  )
}

function ManualWhatIfTab({ modelId, parseResult: _parseResult }: { modelId: string; parseResult: { system_name: string } }) {
  const [actions, setActions] = useState<ChangeAction[]>([])
  const [description, setDescription] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [changesetId, setChangesetId] = useState<string | null>(null)
  const [impact, setImpact] = useState<ImpactDelta[] | null>(null)
  const [loadingImpact, setLoadingImpact] = useState(false)

  const handleAddAction = (action: ChangeAction) => {
    setActions(prev => [...prev, action])
    setError(null)
  }
  const handleRemoveAction = (index: number) => {
    setActions(prev => prev.filter((_, i) => i !== index))
  }
  const handleClear = () => {
    setActions([]); setChangesetId(null); setImpact(null); setError(null)
  }
  const handleSubmit = async () => {
    if (actions.length === 0) return
    setSubmitting(true); setError(null)
    try {
      const csId = `cs-${Date.now()}`
      const result = await api.createChangeset(modelId, { id: csId, description: description || 'What-if scenario', actions })
      setChangesetId(result.id)
      setLoadingImpact(true)
      const impactResult = await api.getChangesetImpact(modelId, result.id)
      setImpact(impactResult.deltas)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Submission failed')
    } finally {
      setSubmitting(false); setLoadingImpact(false)
    }
  }

  return (
    <div className="grid grid-cols-3 gap-6">
      <Card>
        <CardHeader className="pb-3"><CardTitle className="text-sm">Add Action</CardTitle></CardHeader>
        <CardContent><ActionForm onAdd={handleAddAction} /></CardContent>
      </Card>
      <Card>
        <CardHeader className="pb-3"><CardTitle className="text-sm">Changeset Actions</CardTitle></CardHeader>
        <CardContent className="space-y-4">
          <ActionList actions={actions} onRemove={handleRemoveAction} onClear={handleClear} />
          {actions.length > 0 && (
            <>
              <div>
                <label className="text-xs font-medium block mb-1" style={{ color: '#6b7280' }}>Description (optional)</label>
                <input className="w-full rounded-md border px-3 py-2 text-sm"
                  style={{ borderColor: '#d1d5db', background: '#ffffff', color: '#111827' }}
                  placeholder="Describe this what-if scenario..." value={description} onChange={e => setDescription(e.target.value)} />
              </div>
              <Button className="w-full" disabled={submitting || actions.length === 0} onClick={handleSubmit}>
                {submitting ? (<><Loader2 size={14} className="animate-spin" />Submitting...</>) : (<><Send size={14} />Submit & Analyze</>)}
              </Button>
            </>
          )}
          {error && (
            <div className="rounded-lg p-3" style={{ border: '1px solid #fca5a5', background: '#fef2f2' }}>
              <p className="text-xs" style={{ color: '#b91c1c' }}>{error}</p>
            </div>
          )}
        </CardContent>
      </Card>
      <Card>
        <CardHeader className="pb-3"><CardTitle className="text-sm">Impact Analysis</CardTitle></CardHeader>
        <CardContent>
          {loadingImpact ? (
            <div className="flex items-center justify-center py-8"><Loader2 size={20} className="animate-spin" style={{ color: '#9ca3af' }} /></div>
          ) : impact && changesetId ? (
            <ImpactPanel deltas={impact} changesetId={changesetId} />
          ) : (
            <div className="text-center py-8">
              <p className="text-sm" style={{ color: '#9ca3af' }}>Submit a changeset to see impact</p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

export function WhatIfPage() {
  const { modelId, parseResult, isHydrating } = useRequireModel()
  const aiEnabled = useAIEnabled()
  const [tab, setTab] = useState<'ai' | 'manual'>(aiEnabled ? 'ai' : 'manual')


  if (isHydrating || !modelId || !parseResult) return null

  return (
    <div className="max-w-6xl space-y-6 h-full flex flex-col">
      <div className="flex items-center justify-between flex-shrink-0">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight" style={{ color: '#111827' }}>What-If Explorer</h1>
          <p className="text-sm mt-1" style={{ color: '#6b7280' }}>
            Model: {parseResult.system_name}
          </p>
        </div>
        <div className="flex rounded-lg overflow-hidden" style={{ border: '1px solid #d1d5db' }}>
          {aiEnabled && (
            <button className="px-4 py-2 text-sm font-medium flex items-center gap-2"
              style={{ background: tab === 'ai' ? '#111827' : 'white', color: tab === 'ai' ? 'white' : '#374151' }}
              onClick={() => setTab('ai')}>
              <Sparkles size={14} /> AI Scenarios
            </button>
          )}
          <button className="px-4 py-2 text-sm font-medium flex items-center gap-2"
            style={{ background: tab === 'manual' ? '#111827' : 'white', color: tab === 'manual' ? 'white' : '#374151', borderLeft: aiEnabled ? '1px solid #d1d5db' : 'none' }}
            onClick={() => setTab('manual')}>
            <Wrench size={14} /> Manual Mode
          </button>
        </div>
      </div>

      <div className="flex-1 min-h-0">
        {tab === 'ai' ? (
          <AIWhatIfTab modelId={modelId} aiEnabled={aiEnabled} />
        ) : (
          <ManualWhatIfTab modelId={modelId} parseResult={parseResult} />
        )}
      </div>
    </div>
  )
}
