import { useState, useCallback, useRef, useEffect } from 'react'
import { RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { ModelRequired } from '@/components/ui/ModelRequired'
import { useModel } from '@/lib/model-context'
import { advisorApi } from '@/services/api'
import { PageHeader } from '@/components/ui/page-header'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'

const PHASES = [
  'Analyzing team structure...',
  'Checking capability coverage...',
  'Identifying fragmentation patterns...',
  'Evaluating cognitive load...',
  'Generating recommendations...',
]

export function RecommendationsPage() {
  const { modelId, parseResult } = useModel()
  const [report, setReport] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [aiConfigured, setAiConfigured] = useState(true)
  const [phase, setPhase] = useState(0)
  const cacheRef = useRef<Map<string, string>>(new Map())
  const abortRef = useRef<AbortController | null>(null)

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
      const resp = await advisorApi.ask(
        modelId,
        'Generate a comprehensive UNM restructuring report with specific service moves, team restructuring, and a prioritized action roadmap.',
        'recommendations'
      )
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

  useEffect(() => {
    if (!modelId) return
    if (cacheRef.current.has(modelId)) {
      setReport(cacheRef.current.get(modelId)!)
    } else if (!report && !loading) {
      generate()
    }
  }, [modelId]) // eslint-disable-line react-hooks/exhaustive-deps

  return (
    <ModelRequired>
      <div className="max-w-screen-xl mx-auto"><div className="max-w-4xl space-y-4">
        <PageHeader
          title="AI Recommendations"
          description={`Comprehensive restructuring report for ${parseResult?.system_name ?? ''}`}
          actions={
            <Button variant="outline" size="sm" onClick={() => generate(true)} disabled={loading} className="gap-2">
              <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin' : ''}`} />
              Regenerate
            </Button>
          }
        />

        {!aiConfigured && (
          <div className="rounded-lg p-4 bg-amber-50 border border-amber-200 text-sm text-amber-800">
            AI advisor is not configured. Set the UNM_OPENAI_API_KEY environment variable to enable AI features.
          </div>
        )}

        {loading && (
          <div className="flex flex-col items-center justify-center py-20 gap-5 rounded-lg border border-border bg-card">
            <div className="w-10 h-10 rounded-full animate-spin border-2 border-muted border-t-primary" aria-hidden />
            <p className="text-sm font-semibold text-primary animate-pulse">{PHASES[phase]}</p>
            <p className="text-xs text-muted-foreground text-center max-w-sm">This may take up to 2–3 minutes for deep analysis</p>
            <button
              onClick={() => { abortRef.current?.abort(); setLoading(false) }}
              className="mt-2 px-3 py-1 text-sm text-muted-foreground border border-border rounded-md hover:bg-muted transition-colors"
            >
              Cancel
            </button>
          </div>
        )}

        {error && !loading && (
          <div className="rounded-lg p-6 bg-red-50 border border-red-200 space-y-3">
            <p className="text-sm text-red-700">{error}</p>
            <Button variant="outline" size="sm" onClick={() => generate(true)}>Retry</Button>
          </div>
        )}

        {report && !loading && (
          <div className="p-8 rounded-lg border border-border bg-card prose prose-sm max-w-none">
            <ReactMarkdown
              remarkPlugins={[remarkGfm]}
              components={{
                h1: ({ children }) => <h1 className="text-2xl font-extrabold mb-5 pb-3 tracking-tight text-foreground border-b border-border">{children}</h1>,
                h2: ({ children }) => <h2 className="text-lg font-bold mt-8 mb-3 pb-2 text-primary border-b border-border">{children}</h2>,
                h3: ({ children }) => <h3 className="text-base font-bold mt-6 mb-2 text-foreground">{children}</h3>,
                strong: ({ children }) => <strong className="font-bold text-foreground">{children}</strong>,
                li: ({ children }) => <li className="text-sm leading-relaxed text-muted-foreground">{children}</li>,
                p: ({ children }) => <p className="text-sm leading-relaxed mb-4 text-muted-foreground">{children}</p>,
                code: ({ children }) => <code className="px-1.5 py-0.5 rounded text-xs font-mono bg-muted text-primary border border-border">{children}</code>,
                blockquote: ({ children }) => <blockquote className="border-l-4 border-primary/40 pl-4 my-4 italic text-sm text-muted-foreground">{children}</blockquote>,
              }}
            >
              {report}
            </ReactMarkdown>
          </div>
        )}
      </div></div>
    </ModelRequired>
  )
}
