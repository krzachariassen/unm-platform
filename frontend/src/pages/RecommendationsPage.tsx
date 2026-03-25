import { useState, useCallback, useRef, useEffect } from 'react'
import { RefreshCw, FileText } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useRequireModel } from '@/lib/model-context'
import { ModelRequired } from '@/components/ui/ModelRequired'
import { api } from '@/lib/api'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'

const subtitleStyle = { fontSize: 14, color: '#64748b' } as const

const gradientH1Style: React.CSSProperties = {
  fontSize: 30,
  fontWeight: 800,
  letterSpacing: '-0.025em',
  background: 'linear-gradient(135deg, #1e293b 0%, #475569 100%)',
  WebkitBackgroundClip: 'text',
  WebkitTextFillColor: 'transparent',
  backgroundClip: 'text',
}

const reportCardStyle: React.CSSProperties = {
  borderRadius: 20,
  background: 'linear-gradient(135deg, #ffffff 0%, #f8fafc 100%)',
  border: '1px solid #e2e8f0',
  boxShadow: '0 1px 3px rgba(0,0,0,0.04)',
}

export function RecommendationsPage() {
  const { modelId, parseResult, isHydrating } = useRequireModel()
  const [report, setReport] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [aiConfigured, setAiConfigured] = useState(true)
  const cacheRef = useRef<Map<string, string>>(new Map())
  const abortRef = useRef<AbortController | null>(null)
  const [phase, setPhase] = useState(0)

  const phases = [
    'Analyzing team structure...',
    'Checking capability coverage...',
    'Identifying fragmentation patterns...',
    'Evaluating cognitive load...',
    'Generating recommendations...',
  ]

  // Guard handled by useRequireModel — no manual redirect needed

  // Cycle progress phases during loading
  useEffect(() => {
    if (!loading) return
    setPhase(0)
    const interval = setInterval(() => setPhase(p => (p + 1) % 5), 3000)
    return () => clearInterval(interval)
  }, [loading])

  const generate = useCallback(async (force = false) => {
    if (!modelId) return
    if (!force && cacheRef.current.has(modelId)) {
      setReport(cacheRef.current.get(modelId)!)
      return
    }
    const controller = new AbortController()
    abortRef.current = controller
    setLoading(true)
    setError(null)
    try {
      const resp = await api.askAdvisor(modelId, 'Generate a comprehensive UNM restructuring report with specific service moves, team restructuring, and a prioritized action roadmap.', 'recommendations')
      if (controller.signal.aborted) return
      setAiConfigured(resp.ai_configured)
      setReport(resp.answer)
      cacheRef.current.set(modelId, resp.answer)
    } catch (err) {
      if (controller.signal.aborted) return
      setError(err instanceof Error ? err.message : 'Failed to generate recommendations')
    } finally {
      setLoading(false)
      abortRef.current = null
    }
  }, [modelId])

  // On mount, show cached report if available, otherwise auto-generate
  useEffect(() => {
    if (modelId && cacheRef.current.has(modelId)) {
      setReport(cacheRef.current.get(modelId)!)
    } else if (modelId && !report && !loading) {
      generate()
    }
  }, [modelId]) // eslint-disable-line react-hooks/exhaustive-deps

  if (isHydrating || !modelId || !parseResult) return null

  return (
    <ModelRequired>
      <div className="max-w-4xl mx-auto space-y-6">
      <style>{`
        @keyframes rec-progress {
          0% { transform: translateX(-100%); opacity: 0.6; }
          50% { opacity: 1; }
          100% { transform: translateX(200%); opacity: 0.6; }
        }
        @keyframes rec-pulse-text {
          0%, 100% { opacity: 1; }
          50% { opacity: 0.55; }
        }
        .rec-progress-bar {
          animation: rec-progress 1.8s ease-in-out infinite;
        }
        .rec-pulse-text {
          animation: rec-pulse-text 1.4s ease-in-out infinite;
        }
      `}</style>

      <div className="flex items-center justify-between flex-wrap gap-4">
        <div className="flex items-start gap-4">
          <div
            className="flex h-14 w-14 shrink-0 items-center justify-center rounded-2xl"
            style={{
              background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)',
              boxShadow: '0 8px 24px rgba(99, 102, 241, 0.22)',
            }}
          >
            <FileText size={26} strokeWidth={1.75} className="text-white" />
          </div>
          <div>
            <h1 className="leading-tight" style={gradientH1Style}>AI Recommendations</h1>
            <p className="mt-1.5" style={subtitleStyle}>
              Comprehensive restructuring report for {parseResult.system_name}
            </p>
          </div>
        </div>
        <Button
          variant="outline"
          size="sm"
          className="gap-2 font-semibold transition-all hover:bg-slate-50 hover:border-slate-300"
          style={{
            background: '#ffffff',
            border: '1px solid #e2e8f0',
            borderRadius: 12,
            color: '#475569',
          }}
          onClick={() => generate(true)}
          disabled={loading}
        >
          <RefreshCw size={14} className={loading ? 'animate-spin' : ''} />
          Regenerate
        </Button>
      </div>

      {!aiConfigured && (
        <div
          className="rounded-[14px] p-4"
          style={{
            border: '1px solid #fde68a',
            background: 'linear-gradient(135deg, #fffbeb 0%, #fef9c3 100%)',
            boxShadow: '0 1px 3px rgba(0,0,0,0.04)',
          }}
        >
          <p className="text-sm leading-relaxed" style={{ color: '#92400e' }}>
            AI advisor is not configured. Set the UNM_OPENAI_API_KEY environment variable to enable AI features.
          </p>
        </div>
      )}

      {loading && (
        <div
          className="flex flex-col items-center justify-center py-20 gap-5 rounded-[20px]"
          style={reportCardStyle}
        >
          <div
            className="h-12 w-12 rounded-full animate-spin"
            style={{
              border: '2px solid #e2e8f0',
              borderTopColor: '#6366f1',
              borderRightColor: '#8b5cf6',
            }}
            aria-hidden
          />
          <p className="text-sm font-semibold rec-pulse-text" style={{ color: '#6366f1' }}>
            {phases[phase]}
          </p>
          <p className="text-xs text-center max-w-sm" style={subtitleStyle}>
            This may take up to 2-3 minutes for deep analysis
          </p>
          <div className="w-56 h-1.5 rounded-full overflow-hidden mt-1" style={{ background: '#f1f5f9' }}>
            <div
              className="h-full w-1/3 rounded-full rec-progress-bar"
              style={{ background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)' }}
            />
          </div>
          <button
            onClick={() => { abortRef.current?.abort(); setLoading(false) }}
            style={{ color: '#6b7280', background: 'none', border: '1px solid #d1d5db', borderRadius: 6,
              padding: '4px 12px', cursor: 'pointer', fontSize: 13, marginTop: 8 }}
          >
            Cancel
          </button>
        </div>
      )}

      {error && (
        <div
          className="rounded-[20px] p-6 space-y-3"
          style={{
            ...reportCardStyle,
            border: '1px solid #fecaca',
            background: 'linear-gradient(135deg, #ffffff 0%, #fef2f2 100%)',
          }}
        >
          <p className="text-sm font-medium leading-relaxed" style={{ color: '#b91c1c' }}>{error}</p>
          <Button
            variant="outline"
            size="sm"
            className="font-semibold transition-all hover:bg-white"
            style={{
              background: '#ffffff',
              border: '1px solid #e2e8f0',
              borderRadius: 12,
              color: '#475569',
            }}
            onClick={() => generate(true)}
          >
            Retry
          </Button>
        </div>
      )}

      {report && !loading && (
        <div className="p-8 md:p-10 prose prose-sm max-w-none" style={reportCardStyle}>
          <ReactMarkdown
            remarkPlugins={[remarkGfm]}
            components={{
              h1: ({ children }) => (
                <h1
                  className="text-2xl font-extrabold mb-5 pb-3 tracking-tight"
                  style={{
                    borderBottom: '2px solid transparent',
                    borderImage: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%) 1',
                    color: '#1e293b',
                    letterSpacing: '-0.02em',
                  }}
                >
                  {children}
                </h1>
              ),
              h2: ({ children }) => (
                <h2
                  className="text-lg font-bold mt-8 mb-3 pb-2 flex items-center gap-2"
                  style={{
                    color: '#6366f1',
                    borderBottom: '1px solid #e2e8f0',
                  }}
                >
                  <span
                    className="inline-block h-2 w-2 rounded-full shrink-0"
                    style={{ background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)' }}
                    aria-hidden
                  />
                  {children}
                </h2>
              ),
              h3: ({ children }) => (
                <h3 className="text-base font-bold mt-6 mb-2" style={{ color: '#334155', letterSpacing: '-0.01em' }}>
                  {children}
                </h3>
              ),
              strong: ({ children }) => <strong style={{ color: '#1e293b', fontWeight: 700 }}>{children}</strong>,
              li: ({ children }) => <li className="text-sm leading-relaxed" style={{ color: '#475569' }}>{children}</li>,
              p: ({ children }) => <p className="text-sm leading-relaxed mb-4" style={{ color: '#475569' }}>{children}</p>,
              table: ({ children }) => (
                <div className="my-6 overflow-x-auto rounded-xl border" style={{ borderColor: '#e2e8f0', boxShadow: '0 1px 3px rgba(0,0,0,0.04)' }}>
                  <table className="w-full text-xs border-collapse min-w-full">{children}</table>
                </div>
              ),
              thead: ({ children }) => <thead style={{ background: 'linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%)' }}>{children}</thead>,
              th: ({ children }) => (
                <th
                  className="text-left px-4 py-3 font-bold text-xs uppercase tracking-wide"
                  style={{ borderBottom: '2px solid #e2e8f0', color: '#475569' }}
                >
                  {children}
                </th>
              ),
              td: ({ children }) => (
                <td className="px-4 py-3 text-sm" style={{ borderBottom: '1px solid #f1f5f9', color: '#475569', background: '#ffffff' }}>
                  {children}
                </td>
              ),
              tr: ({ children }) => <tr className="transition-colors hover:bg-slate-50/80">{children}</tr>,
              code: ({ children }) => (
                <code
                  className="px-1.5 py-0.5 rounded-md text-xs font-mono"
                  style={{ background: '#f1f5f9', color: '#6366f1', border: '1px solid #e2e8f0' }}
                >
                  {children}
                </code>
              ),
              blockquote: ({ children }) => (
                <blockquote
                  className="border-l-4 pl-4 my-4 italic text-sm"
                  style={{ borderColor: '#a5b4fc', color: '#64748b' }}
                >
                  {children}
                </blockquote>
              ),
            }}
          >
            {report}
          </ReactMarkdown>
        </div>
      )}
    </div>
    </ModelRequired>
  )
}
