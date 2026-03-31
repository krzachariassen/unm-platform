import { useState, useRef, useEffect, useCallback } from 'react'
import { useLocation } from 'react-router-dom'
import { Bot, X, Send, AlertTriangle, Sparkles, RotateCcw } from 'lucide-react'
import { useModel } from '@/lib/model-context'
import { api } from '@/lib/api'
import { Prose } from '@/components/ui/prose'
import { type ChatEntry } from './ChatMessage'
import { ApplyActionsDialog } from './ApplyActionsDialog'

function buildConversationPrompt(history: ChatEntry[], newQuestion: string): string {
  if (history.length === 0) return newQuestion
  const context = history
    .map(e => `User: ${e.question}\nAdvisor: ${e.answer}`)
    .join('\n\n')
  return `Previous conversation:\n${context}\n\nUser: ${newQuestion}\n\nContinue the conversation. Answer the latest question, referencing prior context where relevant.`
}

interface SuggestedPrompt {
  label: string
  question: string
}

interface PageConfig {
  title: string
  description: string
  prompts: SuggestedPrompt[]
}

const PAGE_CONFIGS: Record<string, PageConfig> = {
  '/need': {
    title: 'Need Analysis',
    description: 'Ask about user needs, value delivery and delivery risks',
    prompts: [
      { label: 'Which needs span many teams?', question: 'Which user needs span the most teams and are at highest delivery risk?' },
      { label: 'Unbacked needs?', question: 'Which user needs have no capability backing them?' },
      { label: 'Value stream health', question: 'How healthy is our value delivery from user needs to capabilities to services?' },
      { label: 'Actor coverage', question: 'Are all actors well-served by capabilities or are some underserved?' },
    ],
  },
  '/capability': {
    title: 'Capability Analysis',
    description: 'Ask about capability ownership, fragmentation and gaps',
    prompts: [
      { label: 'Fragmented capabilities?', question: 'Which capabilities are fragmented across too many teams?' },
      { label: 'Optimal service placement', question: 'Are services placed in the right capabilities or should any be moved?' },
      { label: 'Capability hierarchy issues', question: 'Are there issues with how capabilities are decomposed or nested?' },
      { label: 'Unowned capabilities', question: 'Are there capabilities with no clear team ownership?' },
    ],
  },
  '/ownership': {
    title: 'Ownership Analysis',
    description: 'Ask about team ownership, boundaries and responsibilities',
    prompts: [
      { label: 'Team boundary issues', question: 'Which teams have unclear or overlapping ownership boundaries?' },
      { label: 'Cross-team services', question: 'Are there services that should move to a different team?' },
      { label: 'Overloaded teams', question: 'Which teams own too many capabilities or responsibilities?' },
      { label: 'Ownership gaps', question: 'Are there capabilities or services with no clear owner?' },
    ],
  },
  '/team-topology': {
    title: 'Team Topology Analysis',
    description: 'Ask about team types, interactions and collaboration patterns',
    prompts: [
      { label: 'Interaction anti-patterns', question: 'Are there team interactions that should be changed to a different mode?' },
      { label: 'Platform team effectiveness', question: 'Are platform teams effectively reducing cognitive load for stream-aligned teams?' },
      { label: 'Team coupling', question: 'Which teams are too tightly coupled and need more independence?' },
      { label: 'Missing interactions', question: 'Are there teams that should be interacting but are not?' },
    ],
  },
  '/cognitive-load': {
    title: 'Cognitive Load Analysis',
    description: 'Ask about team overload, domain spread and service load',
    prompts: [
      { label: 'Most overloaded teams', question: 'Which teams have the highest cognitive load and why?' },
      { label: 'Domain spread issues', question: 'Which teams own capabilities across too many domains?' },
      { label: 'How to reduce load', question: 'What structural changes would most reduce cognitive load?' },
      { label: 'Team size vs load', question: 'Are team sizes appropriate for their cognitive load levels?' },
    ],
  },
  '/signals': {
    title: 'Architecture Signals',
    description: 'Ask about architectural health, risks and patterns detected',
    prompts: [
      { label: 'Overall health summary', question: "What is the overall architectural health and what are the top risks?" },
      { label: 'Critical bottlenecks', question: 'What are the most critical bottleneck services and what should we do about them?' },
      { label: 'UX risk drivers', question: 'What is driving the UX risk signals and how can we improve them?' },
      { label: 'Where to start improving', question: 'If we could only fix one thing, what would have the most impact on architecture health?' },
    ],
  },
  '/dashboard': {
    title: 'Architecture Overview',
    description: 'Ask for an overview, summary or strategic insights',
    prompts: [
      { label: 'Architecture summary', question: 'Summarize this architecture — what is it, how is it structured, and what stands out?' },
      { label: 'Biggest risks', question: "What are this architecture's biggest structural risks right now?" },
      { label: 'Strategic recommendations', question: 'What are the top 3 structural changes that would have the most impact?' },
      { label: 'How mature is it', question: 'How mature and well-structured is this architecture compared to Team Topologies best practices?' },
    ],
  },
  '/what-if': {
    title: 'What-If Analysis',
    description: 'Ask about the impact and risks of proposed changes',
    prompts: [
      { label: 'Impact of changes', question: 'What is the overall impact of the changes I am considering?' },
      { label: 'Risks of reorganization', question: 'What are the risks of this team reorganization?' },
      { label: 'Better alternatives', question: 'Are there better ways to achieve the same outcome with fewer disruptions?' },
      { label: 'Transition complexity', question: 'How complex would this transition be and what are the key dependencies?' },
    ],
  },
  '/unm-map': {
    title: 'UNM Map Analysis',
    description: 'Ask about the end-to-end value chain from needs to services',
    prompts: [
      { label: 'Value chain integrity', question: 'Is the value chain intact from user needs through capabilities to services?' },
      { label: 'Missing links', question: 'Where are there gaps or missing links in the value chain?' },
      { label: 'Longest chains', question: 'Which value chains span the most teams and have the highest delivery risk?' },
      { label: 'Simplify the map', question: 'How could we simplify this UNM map to reduce complexity?' },
    ],
  },
  '/realization': {
    title: 'Realization Analysis',
    description: 'Ask about how services realize capabilities',
    prompts: [
      { label: 'Multi-capability services', question: 'Which services realize too many capabilities and should be split?' },
      { label: 'Unowned services', question: 'Are there services with no clear team owner?' },
      { label: 'Service-capability fit', question: 'Are services well-matched to the capabilities they realize?' },
      { label: 'Realization gaps', question: 'Which capabilities have no services realizing them?' },
    ],
  },
}

const DEFAULT_CONFIG: PageConfig = {
  title: 'AI Advisor',
  description: 'Ask questions about your architecture model',
  prompts: [
    { label: 'Architecture summary', question: 'Summarize this architecture' },
    { label: 'Biggest risk', question: "What's the biggest structural risk?" },
    { label: 'Overloaded teams', question: 'Which teams are overloaded?' },
    { label: 'Bottlenecks', question: 'Where are the bottlenecks?' },
  ],
}

function getPageConfig(pathname: string): PageConfig {
  const path = pathname.replace(/^\/views/, '')
  return PAGE_CONFIGS[path] ?? DEFAULT_CONFIG
}

export function AdvisorPanel() {
  const location = useLocation()
  const { modelId, parseResult } = useModel()
  const [open, setOpen] = useState(false)
  const [history, setHistory] = useState<ChatEntry[]>([])
  const [loading, setLoading] = useState(false)
  const [pendingQuestion, setPendingQuestion] = useState<string | null>(null)
  const [question, setQuestion] = useState('')
  const [aiConfigured, setAiConfigured] = useState<boolean | null>(null)
  const scrollRef = useRef<HTMLDivElement>(null)
  const askingRef = useRef(false)

  const [applyDialogOpen, setApplyDialogOpen] = useState(false)
  const [applyResponse, setApplyResponse] = useState('')

  const config = getPageConfig(location.pathname)
  const hasModel = !!modelId

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [history, loading, pendingQuestion])

  const handleAsk = useCallback(async (q: string) => {
    if (!modelId || askingRef.current) return
    askingRef.current = true
    const trimmed = q.trim()
    if (!trimmed) return
    setPendingQuestion(trimmed)
    setLoading(true)
    const prompt = buildConversationPrompt(history, trimmed)
    try {
      const resp = await api.askAdvisor(modelId, prompt)
      setAiConfigured(resp.ai_configured)
      setHistory(prev => [...prev, {
        question: trimmed,
        answer: resp.answer,
        aiConfigured: resp.ai_configured,
      }])
    } catch (err) {
      setHistory(prev => [...prev, {
        question: trimmed,
        answer: `Error: ${err instanceof Error ? err.message : 'Request failed'}`,
        aiConfigured: false,
      }])
    } finally {
      setLoading(false)
      setPendingQuestion(null)
      askingRef.current = false
    }
  }, [modelId, history])

  const handleSend = () => {
    handleAsk(question)
    setQuestion('')
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  return (
    <>
      {!open && (
        <button
          onClick={() => setOpen(true)}
          title="Open AI Advisor"
          className="fixed z-40 flex items-center gap-2 rounded-full px-4 py-2.5 text-sm font-medium shadow-lg transition-all hover:shadow-xl"
          style={{ background: '#2563eb', color: '#ffffff', bottom: 24, right: 24 }}
        >
          <Bot size={16} />
          Ask AI
        </button>
      )}

      {open && (
        <div
          className="fixed right-0 top-0 bottom-0 z-30 flex flex-col"
          style={{ width: 380, background: '#ffffff', borderLeft: '1px solid #e5e7eb', boxShadow: '-4px 0 24px rgba(0,0,0,0.08)' }}
        >
          <div
            className="flex items-center gap-3 px-4 flex-shrink-0"
            style={{ height: 56, borderBottom: '1px solid #e5e7eb' }}
          >
            <div className="flex items-center gap-2 flex-1 min-w-0">
              <div className="rounded-md p-1.5 flex-shrink-0" style={{ background: '#dbeafe' }}>
                <Bot size={14} style={{ color: '#2563eb' }} />
              </div>
              <div className="min-w-0">
                <div className="text-sm font-semibold truncate" style={{ color: '#111827' }}>{config.title}</div>
                <div className="text-xs truncate" style={{ color: '#9ca3af' }}>{config.description}</div>
              </div>
            </div>
            {history.length > 0 && (
              <button
                onClick={() => setHistory([])}
                title="New conversation"
                className="flex-shrink-0 rounded-md p-1.5 transition-colors hover:bg-gray-100"
                style={{ color: '#9ca3af' }}
              >
                <RotateCcw size={13} />
              </button>
            )}
            <button
              onClick={() => setOpen(false)}
              className="flex-shrink-0 rounded-md p-1.5 transition-colors hover:bg-gray-100"
              style={{ color: '#9ca3af' }}
            >
              <X size={15} />
            </button>
          </div>

          <div className="px-4 py-2 flex-shrink-0" style={{ borderBottom: '1px solid #f3f4f6' }}>
            <div className="flex items-center gap-2 text-xs" style={{ color: hasModel ? '#374151' : '#9ca3af' }}>
              <span className="w-1.5 h-1.5 rounded-full flex-shrink-0" style={{ background: hasModel ? '#22c55e' : '#d1d5db' }} />
              {hasModel
                ? <span className="truncate">Model: <strong>{parseResult?.system_name}</strong></span>
                : <span>No model loaded</span>
              }
            </div>
            {aiConfigured === false && (
              <div className="mt-1.5 flex items-center gap-1.5 text-xs" style={{ color: '#92400e' }}>
                <AlertTriangle size={11} />
                Set UNM_OPENAI_API_KEY to enable AI
              </div>
            )}
          </div>

          {hasModel && history.length === 0 && (
            <div className="px-4 py-3 flex-shrink-0" style={{ borderBottom: '1px solid #f3f4f6' }}>
              <div className="flex items-center gap-1.5 mb-2">
                <Sparkles size={11} style={{ color: '#9ca3af' }} />
                <span className="text-xs font-medium" style={{ color: '#9ca3af' }}>Suggested for this page</span>
              </div>
              <div className="flex flex-col gap-1.5">
                {config.prompts.map((prompt) => (
                  <button
                    key={prompt.label}
                    onClick={() => { if (hasModel && !loading) handleAsk(prompt.question) }}
                    disabled={!hasModel || loading}
                    className="text-left text-xs px-3 py-2 rounded-md transition-colors disabled:opacity-40 disabled:pointer-events-none"
                    style={{ background: '#f8fafc', border: '1px solid #e2e8f0', color: '#374151' }}
                    onMouseEnter={e => { (e.currentTarget as HTMLElement).style.background = '#f1f5f9'; (e.currentTarget as HTMLElement).style.borderColor = '#cbd5e1' }}
                    onMouseLeave={e => { (e.currentTarget as HTMLElement).style.background = '#f8fafc'; (e.currentTarget as HTMLElement).style.borderColor = '#e2e8f0' }}
                  >
                    {prompt.label}
                  </button>
                ))}
              </div>
            </div>
          )}

          <div ref={scrollRef} className="flex-1 overflow-auto px-4 py-3 space-y-4" style={{ minHeight: 0 }}>
            {history.length === 0 && !loading && (
              <div className="flex flex-col items-center justify-center py-10 text-center">
                <Bot size={28} style={{ color: '#e5e7eb' }} />
                <p className="mt-2 text-xs" style={{ color: '#9ca3af' }}>
                  {hasModel
                    ? 'Click a suggestion or ask a question'
                    : 'Load a model to start'
                  }
                </p>
              </div>
            )}
            {history.map((entry, i) => (
              <div key={i} className="space-y-2">
                <div className="flex gap-2">
                  <div
                    className="flex-1 text-xs px-3 py-2 rounded-lg"
                    style={{ background: '#f3f4f6', color: '#374151' }}
                  >
                    {entry.question}
                  </div>
                </div>
                <div className="flex gap-2">
                  <div className="flex-shrink-0 mt-1">
                    <Bot size={12} style={{ color: '#2563eb' }} />
                  </div>
                  <div
                    className="flex-1 px-3 py-2 rounded-lg"
                    style={{ background: '#eff6ff', border: '1px solid #dbeafe' }}
                  >
                    <Prose compact>{entry.answer}</Prose>
                    {entry.aiConfigured && (
                      <button
                        type="button"
                        onClick={() => { setApplyResponse(entry.answer); setApplyDialogOpen(true) }}
                        className="mt-2 flex items-center gap-1.5 text-[10px] font-medium px-2 py-1 rounded transition-all hover:opacity-90"
                        style={{ background: '#111827', color: '#ffffff' }}
                      >
                        <Sparkles size={9} /> Apply
                      </button>
                    )}
                  </div>
                </div>
              </div>
            ))}
            {loading && (
              <div className="space-y-2">
                {pendingQuestion && (
                  <div className="flex gap-2">
                    <div
                      className="flex-1 text-xs px-3 py-2 rounded-lg"
                      style={{ background: '#f3f4f6', color: '#374151' }}
                    >
                      {pendingQuestion}
                    </div>
                  </div>
                )}
                <div className="flex gap-2">
                  <div className="flex-shrink-0 mt-1">
                    <Bot size={12} className="animate-pulse" style={{ color: '#2563eb' }} />
                  </div>
                  <div className="flex items-center gap-2 px-3 py-2 rounded-lg" style={{ background: '#eff6ff', border: '1px solid #dbeafe' }}>
                    <span className="text-xs" style={{ color: '#6b7280' }}>Thinking</span>
                    <span className="flex gap-0.5">
                      <span className="w-1 h-1 rounded-full animate-bounce" style={{ background: '#2563eb', animationDelay: '0ms', animationDuration: '1.2s' }} />
                      <span className="w-1 h-1 rounded-full animate-bounce" style={{ background: '#2563eb', animationDelay: '200ms', animationDuration: '1.2s' }} />
                      <span className="w-1 h-1 rounded-full animate-bounce" style={{ background: '#2563eb', animationDelay: '400ms', animationDuration: '1.2s' }} />
                    </span>
                  </div>
                </div>
              </div>
            )}
          </div>

          <div className="px-4 py-3 flex-shrink-0" style={{ borderTop: '1px solid #e5e7eb' }}>
            <div className="flex gap-2">
              <textarea
                value={question}
                onChange={e => setQuestion(e.target.value)}
                onKeyDown={handleKeyDown}
                disabled={!hasModel || loading}
                placeholder={!hasModel ? 'Load a model first...' : `Ask about ${config.title.toLowerCase()}...`}
                rows={2}
                className="flex-1 text-xs rounded-lg px-3 py-2 resize-none disabled:opacity-40 focus:outline-none focus:ring-1 focus:ring-blue-300"
                style={{ border: '1px solid #e5e7eb', color: '#111827', background: '#ffffff' }}
              />
              <button
                onClick={handleSend}
                disabled={!hasModel || loading || !question.trim()}
                className="self-end rounded-lg p-2 transition-colors disabled:opacity-40 disabled:pointer-events-none"
                style={{ background: '#2563eb', color: '#ffffff' }}
              >
                {loading
                  ? <span className="w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin block" />
                  : <Send size={14} />
                }
              </button>
            </div>
            <p className="mt-1.5 text-xs" style={{ color: '#d1d5db' }}>
              Enter to send · Shift+Enter for newline
            </p>
          </div>
        </div>
      )}
      <ApplyActionsDialog
        open={applyDialogOpen}
        onClose={() => setApplyDialogOpen(false)}
        advisorResponse={applyResponse}
      />
    </>
  )
}
