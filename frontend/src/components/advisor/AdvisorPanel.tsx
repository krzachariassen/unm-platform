import { useState, useRef, useEffect } from 'react'
import { useLocation } from 'react-router-dom'
import { Bot, X, Send, AlertTriangle, Sparkles } from 'lucide-react'
import { useModel } from '@/lib/model-context'
import { api } from '@/lib/api'
import { type ChatEntry } from './ChatMessage'

interface SuggestedPrompt {
  label: string
  question: string
  category: string
}

interface PageConfig {
  title: string
  description: string
  prompts: SuggestedPrompt[]
  defaultCategory: string
}

const PAGE_CONFIGS: Record<string, PageConfig> = {
  '/need': {
    title: 'Need Analysis',
    description: 'Ask about user needs, value delivery and delivery risks',
    defaultCategory: 'need-delivery-risk',
    prompts: [
      { label: 'Which needs span many teams?', question: 'Which user needs span the most teams and are at highest delivery risk?', category: 'need-delivery-risk' },
      { label: 'Unbacked needs?', question: 'Which user needs have no capability backing them?', category: 'need-delivery-risk' },
      { label: 'Value stream health', question: 'How healthy is our value delivery from user needs to capabilities to services?', category: 'value-stream' },
      { label: 'Actor coverage', question: 'Are all actors well-served by capabilities or are some underserved?', category: 'need-delivery-risk' },
    ],
  },
  '/capability': {
    title: 'Capability Analysis',
    description: 'Ask about capability ownership, fragmentation and gaps',
    defaultCategory: 'fragmentation',
    prompts: [
      { label: 'Fragmented capabilities?', question: 'Which capabilities are fragmented across too many teams?', category: 'fragmentation' },
      { label: 'Optimal service placement', question: 'Are services placed in the right capabilities or should any be moved?', category: 'service-placement' },
      { label: 'Capability hierarchy issues', question: 'Are there issues with how capabilities are decomposed or nested?', category: 'fragmentation' },
      { label: 'Unowned capabilities', question: 'Are there capabilities with no clear team ownership?', category: 'team-boundary' },
    ],
  },
  '/ownership': {
    title: 'Ownership Analysis',
    description: 'Ask about team ownership, boundaries and responsibilities',
    defaultCategory: 'team-boundary',
    prompts: [
      { label: 'Team boundary issues', question: 'Which teams have unclear or overlapping ownership boundaries?', category: 'team-boundary' },
      { label: 'Cross-team services', question: 'Are there services that should move to a different team?', category: 'service-placement' },
      { label: 'Overloaded teams', question: 'Which teams own too many capabilities or responsibilities?', category: 'structural-load' },
      { label: 'Ownership gaps', question: 'Are there capabilities or services with no clear owner?', category: 'team-boundary' },
    ],
  },
  '/team-topology': {
    title: 'Team Topology Analysis',
    description: 'Ask about team types, interactions and collaboration patterns',
    defaultCategory: 'interaction-mode',
    prompts: [
      { label: 'Interaction anti-patterns', question: 'Are there team interactions that should be changed to a different mode?', category: 'interaction-mode' },
      { label: 'Platform team effectiveness', question: 'Are platform teams effectively reducing cognitive load for stream-aligned teams?', category: 'structural-load' },
      { label: 'Team coupling', question: 'Which teams are too tightly coupled and need more independence?', category: 'coupling' },
      { label: 'Missing interactions', question: 'Are there teams that should be interacting but are not?', category: 'interaction-mode' },
    ],
  },
  '/cognitive-load': {
    title: 'Cognitive Load Analysis',
    description: 'Ask about team overload, domain spread and service load',
    defaultCategory: 'structural-load',
    prompts: [
      { label: 'Most overloaded teams', question: 'Which teams have the highest cognitive load and why?', category: 'structural-load' },
      { label: 'Domain spread issues', question: 'Which teams own capabilities across too many domains?', category: 'structural-load' },
      { label: 'How to reduce load', question: 'What structural changes would most reduce cognitive load?', category: 'structural-load' },
      { label: 'Team size vs load', question: 'Are team sizes appropriate for their cognitive load levels?', category: 'structural-load' },
    ],
  },
  '/signals': {
    title: 'Architecture Signals',
    description: 'Ask about architectural health, risks and patterns detected',
    defaultCategory: 'health-summary',
    prompts: [
      { label: 'Overall health summary', question: "What is the overall architectural health and what are the top risks?", category: 'health-summary' },
      { label: 'Critical bottlenecks', question: 'What are the most critical bottleneck services and what should we do about them?', category: 'bottleneck' },
      { label: 'UX risk drivers', question: 'What is driving the UX risk signals and how can we improve them?', category: 'need-delivery-risk' },
      { label: 'Where to start improving', question: 'If we could only fix one thing, what would have the most impact on architecture health?', category: 'health-summary' },
    ],
  },
  '/dashboard': {
    title: 'Architecture Overview',
    description: 'Ask for an overview, summary or strategic insights',
    defaultCategory: 'model-summary',
    prompts: [
      { label: 'Architecture summary', question: 'Summarize this architecture — what is it, how is it structured, and what stands out?', category: 'model-summary' },
      { label: 'Biggest risks', question: "What are this architecture's biggest structural risks right now?", category: 'health-summary' },
      { label: 'Strategic recommendations', question: 'What are the top 3 structural changes that would have the most impact?', category: 'health-summary' },
      { label: 'How mature is it', question: 'How mature and well-structured is this architecture compared to Team Topologies best practices?', category: 'model-summary' },
    ],
  },
  '/what-if': {
    title: 'What-If Analysis',
    description: 'Ask about the impact and risks of proposed changes',
    defaultCategory: 'value-stream',
    prompts: [
      { label: 'Impact of changes', question: 'What is the overall impact of the changes I am considering?', category: 'value-stream' },
      { label: 'Risks of reorganization', question: 'What are the risks of this team reorganization?', category: 'coupling' },
      { label: 'Better alternatives', question: 'Are there better ways to achieve the same outcome with fewer disruptions?', category: 'service-placement' },
      { label: 'Transition complexity', question: 'How complex would this transition be and what are the key dependencies?', category: 'coupling' },
    ],
  },
  '/unm-map': {
    title: 'UNM Map Analysis',
    description: 'Ask about the end-to-end value chain from needs to services',
    defaultCategory: 'value-stream',
    prompts: [
      { label: 'Value chain integrity', question: 'Is the value chain intact from user needs through capabilities to services?', category: 'value-stream' },
      { label: 'Missing links', question: 'Where are there gaps or missing links in the value chain?', category: 'need-delivery-risk' },
      { label: 'Longest chains', question: 'Which value chains span the most teams and have the highest delivery risk?', category: 'need-delivery-risk' },
      { label: 'Simplify the map', question: 'How could we simplify this UNM map to reduce complexity?', category: 'model-summary' },
    ],
  },
  '/realization': {
    title: 'Realization Analysis',
    description: 'Ask about how services realize capabilities',
    defaultCategory: 'service-placement',
    prompts: [
      { label: 'Multi-capability services', question: 'Which services realize too many capabilities and should be split?', category: 'service-placement' },
      { label: 'Unowned services', question: 'Are there services with no clear team owner?', category: 'team-boundary' },
      { label: 'Service-capability fit', question: 'Are services well-matched to the capabilities they realize?', category: 'service-placement' },
      { label: 'Realization gaps', question: 'Which capabilities have no services realizing them?', category: 'fragmentation' },
    ],
  },
}

const DEFAULT_CONFIG: PageConfig = {
  title: 'AI Advisor',
  description: 'Ask questions about your architecture model',
  defaultCategory: 'general',
  prompts: [
    { label: 'Architecture summary', question: 'Summarize this architecture', category: 'model-summary' },
    { label: 'Biggest risk', question: "What's the biggest structural risk?", category: 'health-summary' },
    { label: 'Overloaded teams', question: 'Which teams are overloaded?', category: 'structural-load' },
    { label: 'Bottlenecks', question: 'Where are the bottlenecks?', category: 'bottleneck' },
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
  const [question, setQuestion] = useState('')
  const [aiConfigured, setAiConfigured] = useState<boolean | null>(null)
  const scrollRef = useRef<HTMLDivElement>(null)

  const config = getPageConfig(location.pathname)
  const hasModel = !!modelId

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [history, loading])

  const handleAsk = async (q: string, category: string) => {
    if (!modelId) return
    const trimmed = q.trim()
    if (!trimmed) return
    setLoading(true)
    try {
      const resp = await api.askAdvisor(modelId, trimmed, category)
      setAiConfigured(resp.ai_configured)
      setHistory(prev => [...prev, {
        question: trimmed,
        answer: resp.answer,
        category: resp.category,
        aiConfigured: resp.ai_configured,
      }])
    } catch (err) {
      setHistory(prev => [...prev, {
        question: trimmed,
        answer: `Error: ${err instanceof Error ? err.message : 'Request failed'}`,
        category,
        aiConfigured: false,
      }])
    } finally {
      setLoading(false)
    }
  }

  const handleSend = () => {
    handleAsk(question, config.defaultCategory)
    setQuestion('')
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  const handlePromptClick = (prompt: SuggestedPrompt) => {
    if (!hasModel || loading) return
    handleAsk(prompt.question, prompt.category)
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

          {hasModel && (
            <div className="px-4 py-3 flex-shrink-0" style={{ borderBottom: '1px solid #f3f4f6' }}>
              <div className="flex items-center gap-1.5 mb-2">
                <Sparkles size={11} style={{ color: '#9ca3af' }} />
                <span className="text-xs font-medium" style={{ color: '#9ca3af' }}>Suggested for this page</span>
              </div>
              <div className="flex flex-col gap-1.5">
                {config.prompts.map((prompt) => (
                  <button
                    key={prompt.label}
                    onClick={() => handlePromptClick(prompt)}
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
                    className="flex-1 text-xs px-3 py-2 rounded-lg leading-relaxed"
                    style={{ background: '#eff6ff', color: '#1e3a5f', border: '1px solid #dbeafe' }}
                  >
                    {entry.answer}
                    {entry.category && entry.category !== 'general' && (
                      <span
                        className="mt-2 inline-block px-1.5 py-0.5 rounded text-xs"
                        style={{ background: '#dbeafe', color: '#1d4ed8', fontSize: 10 }}
                      >
                        {entry.category}
                      </span>
                    )}
                  </div>
                </div>
              </div>
            ))}
            {loading && (
              <div className="flex items-center gap-2 text-xs" style={{ color: '#9ca3af' }}>
                <span className="w-3 h-3 border-2 border-current border-t-transparent rounded-full animate-spin" />
                Thinking...
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
    </>
  )
}
