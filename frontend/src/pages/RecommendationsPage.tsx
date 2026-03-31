import { useState, useCallback, useRef, useEffect } from 'react'
import { RefreshCw, Sparkles, Loader2, FileText, Download } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { ModelRequired } from '@/components/ui/ModelRequired'
import { useModel } from '@/lib/model-context'
import { advisorApi } from '@/services/api'
import { PageHeader } from '@/components/ui/page-header'
import { Prose } from '@/components/ui/prose'
import { exportToPdf } from '@/lib/export-pdf'
import { ApplyActionsDialog } from '@/components/advisor/ApplyActionsDialog'

export function RecommendationsPage() {
  const { modelId, parseResult } = useModel()
  const [report, setReport] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [aiConfigured, setAiConfigured] = useState(true)
  const [applyDialogOpen, setApplyDialogOpen] = useState(false)
  const cacheRef = useRef<Map<string, string>>(new Map())
  const abortRef = useRef<AbortController | null>(null)

  useEffect(() => {
    if (!modelId) return
    if (cacheRef.current.has(modelId)) {
      setReport(cacheRef.current.get(modelId)!)
    }
  }, [modelId])

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

  return (
    <ModelRequired>
      <div className="max-w-screen-xl mx-auto"><div className="max-w-4xl space-y-4">
        <PageHeader
          title="AI Recommendations"
          description={`Comprehensive restructuring report for ${parseResult?.system_name ?? ''}`}
          actions={report && !loading ? (
            <div className="flex items-center gap-2">
              <button
                type="button"
                onClick={() => setApplyDialogOpen(true)}
                className="flex items-center gap-1.5 rounded-md px-3 py-1.5 text-xs font-semibold transition-all hover:opacity-90"
                style={{ background: '#111827', color: '#ffffff' }}
              >
                <Sparkles size={13} />
                Apply this report
              </button>
              <Button variant="outline" size="sm" onClick={() => exportToPdf(report, `AI Recommendations — ${parseResult?.system_name ?? 'UNM'}`)} className="gap-2">
                <Download className="w-3.5 h-3.5" />
                Save as PDF
              </Button>
              <Button variant="outline" size="sm" onClick={() => generate(true)} disabled={loading} className="gap-2">
                <RefreshCw className="w-3.5 h-3.5" />
                Regenerate
              </Button>
            </div>
          ) : undefined}
        />

        {!aiConfigured && (
          <div className="rounded-lg p-4 bg-amber-50 border border-amber-200 text-sm text-amber-800">
            AI advisor is not configured. Set the UNM_OPENAI_API_KEY environment variable to enable AI features.
          </div>
        )}

        {/* Empty state — explicit generate */}
        {!report && !loading && !error && (
          <div className="flex flex-col items-center justify-center py-20 gap-5 rounded-xl border" style={{ background: '#fafafa', borderColor: '#e5e7eb' }}>
            <div className="rounded-xl p-4" style={{ background: '#eff6ff' }}>
              <FileText size={28} style={{ color: '#2563eb' }} />
            </div>
            <div className="text-center max-w-md">
              <h3 className="text-sm font-semibold mb-1" style={{ color: '#111827' }}>Generate Restructuring Report</h3>
              <p className="text-xs leading-relaxed" style={{ color: '#6b7280' }}>
                AI will analyze your architecture model and produce a comprehensive report with specific service moves, team restructuring suggestions, and a prioritized action roadmap. This may take 1–3 minutes.
              </p>
            </div>
            <button
              type="button"
              onClick={() => generate()}
              className="flex items-center gap-2 rounded-lg px-5 py-2.5 text-sm font-semibold text-white transition-all hover:opacity-90"
              style={{ background: '#111827' }}
            >
              <Sparkles size={15} /> Generate Report
            </button>
          </div>
        )}

        {loading && (
          <div className="flex flex-col items-center justify-center py-20 gap-4 rounded-xl border" style={{ background: '#fafafa', borderColor: '#e5e7eb' }}>
            <Loader2 size={28} className="animate-spin" style={{ color: '#2563eb' }} />
            <div className="text-center">
              <p className="text-sm font-semibold" style={{ color: '#111827' }}>Generating the report…</p>
              <p className="text-[11px] mt-1" style={{ color: '#9ca3af' }}>This may take up to 2–3 minutes</p>
            </div>
            <button
              type="button"
              onClick={() => { abortRef.current?.abort(); setLoading(false) }}
              className="text-xs px-3 py-1.5 rounded-md transition-colors hover:bg-gray-100"
              style={{ color: '#6b7280', border: '1px solid #e5e7eb' }}
            >
              Cancel
            </button>
          </div>
        )}

        {error && !loading && (
          <div className="rounded-lg p-6 space-y-3" style={{ background: '#fef2f2', border: '1px solid #fecaca' }}>
            <p className="text-sm" style={{ color: '#b91c1c' }}>{error}</p>
            <button
              type="button"
              onClick={() => generate(true)}
              className="text-xs font-medium px-3 py-1.5 rounded-md transition-colors hover:bg-red-100"
              style={{ color: '#dc2626', border: '1px solid #fca5a5' }}
            >
              Retry
            </button>
          </div>
        )}

        {report && !loading && (
          <div className="p-8 rounded-lg border border-border" style={{ background: '#ffffff' }}>
            <Prose>{report}</Prose>
          </div>
        )}
      </div></div>

      {report && (
        <ApplyActionsDialog
          open={applyDialogOpen}
          onClose={() => setApplyDialogOpen(false)}
          advisorResponse={report}
        />
      )}
    </ModelRequired>
  )
}
