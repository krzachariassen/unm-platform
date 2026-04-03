import { useState, useCallback, useEffect, useRef } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { Upload, AlertCircle, CheckCircle2, FlaskConical, Loader2, Sparkles } from 'lucide-react'
import { modelsApi, insightsApi } from '@/services/api'
import { useModel } from '@/lib/model-context'
import { useAIEnabled } from '@/hooks/useAIEnabled'
import { getRuntimeConfig } from '@/lib/runtimeConfig'
import { Button } from '@/components/ui/button'
import { PageHeader } from '@/components/ui/page-header'
import { cn } from '@/lib/utils'
import type { ParseResponse } from '@/types/model'

type StepStatus = 'idle' | 'active' | 'done' | 'error'

const BASE_STEPS = [
  { id: 'read',    label: 'Reading file' },
  { id: 'parse',   label: 'Parsing & validating model' },
  { id: 'analyse', label: 'Analysing architecture' },
]
const AI_STEP = { id: 'ai', label: 'Generating AI insights', sublabel: 'Signals, needs, capabilities, ownership…' }

function StepRow({ step, status }: { step: { id: string; label: string; sublabel?: string }; status: StepStatus }) {
  return (
    <div className="flex items-start gap-3">
      <div className="shrink-0 mt-0.5">
        {status === 'done'   && <CheckCircle2 className="w-4 h-4 text-green-500" />}
        {status === 'active' && <Loader2 className="w-4 h-4 animate-spin text-blue-600" />}
        {status === 'error'  && <AlertCircle className="w-4 h-4 text-red-500" />}
        {status === 'idle'   && <div className="w-4 h-4 rounded-full border-2 border-gray-300" />}
      </div>
      <div>
        <p className={cn('text-sm font-medium',
          status === 'active' ? 'text-blue-700' : status === 'done' ? 'text-green-700' : status === 'error' ? 'text-red-700' : 'text-gray-400'
        )}>{step.label}</p>
        {step.sublabel && status === 'active' && <p className="text-xs mt-0.5 text-blue-400">{step.sublabel}</p>}
      </div>
    </div>
  )
}

export function UploadPage() {
  const [dragging, setDragging] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [parseResult, setParseResult] = useState<ParseResponse | null>(null)
  const [parseWarnings, setParseWarnings] = useState<string[]>([])
  const [debugEnabled, setDebugEnabled] = useState(false)
  const aiEnabled = useAIEnabled()
  const STEPS = aiEnabled ? [...BASE_STEPS, AI_STEP] : BASE_STEPS
  const [stepStatuses, setStepStatuses] = useState<Record<string, StepStatus>>({ read: 'idle', parse: 'idle', analyse: 'idle', ai: 'idle' })
  const [processing, setProcessing] = useState(false)
  const { modelId, parseResult: loadedParseResult, setModel } = useModel()
  const navigate = useNavigate()
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const setStep = (id: string, status: StepStatus) => setStepStatuses(prev => ({ ...prev, [id]: status }))

  useEffect(() => () => { if (pollRef.current) clearInterval(pollRef.current) }, [])
  useEffect(() => { getRuntimeConfig().then(cfg => setDebugEnabled(cfg.features?.debug_routes ?? false)) }, [])

  const startPollingInsights = useCallback((id: string) => {
    setStep('ai', 'active')
    const poll = async () => {
      try {
        const status = await insightsApi.getInsightsStatus(id)
        if (status.all_ready) { clearInterval(pollRef.current!); pollRef.current = null; setStep('ai', 'done'); setTimeout(() => navigate('/dashboard'), 400) }
      } catch { /* transient error — keep polling */ }
    }
    poll()
    pollRef.current = setInterval(poll, 2000)
  }, [navigate])

  const runUploadFlow = useCallback(async (content: string, filename: string) => {
    setError(null); setProcessing(true); setStepStatuses({ read: 'done', parse: 'idle', analyse: 'idle', ai: 'idle' }); setParseResult(null); setParseWarnings([])
    const format = filename.endsWith('.unm') && !filename.endsWith('.unm.yaml') ? 'dsl' as const : undefined
    setStep('parse', 'active')
    let parsed: ParseResponse
    try {
      parsed = await modelsApi.parseModel(content, modelId ?? undefined, format)
    } catch (e) { setStep('parse', 'error'); setError(e instanceof Error ? e.message : 'Upload failed'); setProcessing(false); return }
    setParseResult(parsed)
    if (!parsed.validation.is_valid) { setStep('parse', 'error'); setProcessing(false); return }
    setStep('parse', 'done'); setModel(parsed.id, parsed)
    if (parsed.warnings?.length) setParseWarnings(parsed.warnings)
    setStep('analyse', 'active'); await new Promise(r => setTimeout(r, 200)); setStep('analyse', 'done')
    aiEnabled ? startPollingInsights(parsed.id) : setTimeout(() => navigate('/dashboard'), 400)
  }, [modelId, setModel, startPollingInsights, aiEnabled, navigate])

  const loadExample = useCallback(async () => {
    setError(null); setProcessing(true); setStepStatuses({ read: 'done', parse: 'idle', analyse: 'idle', ai: 'idle' }); setParseResult(null); setParseWarnings([])
    setStep('parse', 'active')
    let parsed: ParseResponse
    try {
      parsed = await modelsApi.loadExample(modelId ?? undefined)
    } catch (e) { setStep('parse', 'error'); setError(e instanceof Error ? e.message : 'Failed to load example'); setProcessing(false); return }
    setParseResult(parsed)
    if (!parsed.validation.is_valid) { setStep('parse', 'error'); setProcessing(false); return }
    setStep('parse', 'done'); setModel(parsed.id, parsed)
    if (parsed.warnings?.length) setParseWarnings(parsed.warnings)
    setStep('analyse', 'active'); await new Promise(r => setTimeout(r, 200)); setStep('analyse', 'done')
    aiEnabled ? startPollingInsights(parsed.id) : setTimeout(() => navigate('/dashboard'), 400)
  }, [modelId, setModel, startPollingInsights, aiEnabled, navigate])

  const onDrop = useCallback((e: React.DragEvent) => {
    const file = e.dataTransfer.files[0]; if (!file) return
    setStep('read', 'active')
    const reader = new FileReader()
    reader.onload = (ev) => { setStep('read', 'done'); runUploadFlow(ev.target?.result as string, file.name) }
    reader.readAsText(file)
  }, [runUploadFlow])

  const onFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]; if (!file) return
    setStep('read', 'active')
    const reader = new FileReader()
    reader.onload = (ev) => { setStep('read', 'done'); runUploadFlow(ev.target?.result as string, file.name) }
    reader.readAsText(file)
  }

  const showProgress = processing || stepStatuses.parse !== 'idle'

  return (
    <div className="max-w-screen-xl mx-auto">
      <div className="max-w-2xl space-y-4">
      <PageHeader title="Upload Model" description="Upload a .unm or .unm.yaml file to explore your architecture" />

      {modelId && loadedParseResult && !showProgress && (
        <div className="flex items-center justify-between p-4 rounded-lg bg-green-50 border border-green-200">
          <span className="text-sm font-medium text-green-700">Model loaded: {loadedParseResult.system_name}</span>
          <Link to="/dashboard" className="text-sm font-medium text-green-600 hover:text-green-700">Go to Dashboard →</Link>
        </div>
      )}

      {!showProgress && (
        <>
          <label
            htmlFor="file-input"
            onDragOver={(e) => { e.preventDefault(); setDragging(true) }}
            onDragLeave={() => setDragging(false)}
            onDrop={onDrop}
            className={cn('block border-2 border-dashed rounded-xl p-12 text-center transition-colors cursor-pointer',
              dragging ? 'border-primary bg-primary/5' : 'border-border hover:border-primary/50')}
          >
            <Upload className="mx-auto mb-3 text-muted-foreground" size={32} />
            <p className="text-sm font-medium">Drop your .unm or .unm.yaml file here</p>
            <p className="text-xs text-muted-foreground mt-1">or click to browse</p>
            <input id="file-input" type="file" accept=".yaml,.yml,.unm" className="sr-only" onChange={onFileChange} />
          </label>

          {debugEnabled && (
            <>
              <div className="flex items-center gap-3">
                <div className="flex-1 h-px bg-border" /><span className="text-xs text-muted-foreground">or</span><div className="flex-1 h-px bg-border" />
              </div>
              <Button variant="outline" className="w-full gap-2" onClick={loadExample}>
                <FlaskConical size={16} /> Load Example Model
              </Button>
            </>
          )}
        </>
      )}

      {showProgress && (
        <div className="rounded-xl border border-border p-6 space-y-5">
          <div className="flex items-center gap-2 mb-2">
            <Sparkles className="w-4 h-4 text-blue-600" />
            <span className="text-sm font-semibold">Processing model</span>
          </div>
          <div className="space-y-4">
            {STEPS.map(step => <StepRow key={step.id} step={step} status={stepStatuses[step.id]} />)}
          </div>
          {parseResult && !parseResult.validation.is_valid && (
            <div className="mt-4 space-y-2">
              {parseResult.validation.errors.map((e, i) => (
                <div key={i} className="text-xs text-destructive/80 flex gap-1"><span className="font-mono">[{e.code}]</span> {e.message}</div>
              ))}
              <Button variant="outline" size="sm" className="mt-2" onClick={() => { setProcessing(false); setStepStatuses({ read: 'idle', parse: 'idle', analyse: 'idle', ai: 'idle' }); setParseResult(null) }}>Try another file</Button>
            </div>
          )}
        </div>
      )}

      {parseWarnings.length > 0 && (
        <div className="p-4 rounded-lg bg-amber-50 border border-amber-200">
          <div className="flex items-center gap-2 mb-2">
            <AlertCircle className="w-4 h-4 text-amber-600" />
            <span className="text-sm font-semibold text-amber-800">Reference Warnings ({parseWarnings.length})</span>
          </div>
          <ul className="ml-5 space-y-0.5 text-sm text-amber-800 list-disc">
            {parseWarnings.map((w, i) => <li key={i}>{w}</li>)}
          </ul>
          <p className="text-xs text-amber-700 mt-2">These references were not found in the model. Check spelling of capability, service, or team names.</p>
        </div>
      )}

      {error && (
        <div className="rounded-xl p-4 flex gap-2 bg-red-50 border border-red-200">
          <AlertCircle className="w-4 h-4 text-red-500 shrink-0 mt-0.5" />
          <div className="flex-1">
            <span className="text-sm text-red-700">{error}</span>
            <Button variant="outline" size="sm" className="mt-2 block" onClick={() => { setError(null); setProcessing(false); setStepStatuses({ read: 'idle', parse: 'idle', analyse: 'idle', ai: 'idle' }) }}>Try again</Button>
          </div>
        </div>
      )}

      {!showProgress && (
        <div className="mt-12 pt-8 border-t border-border">
          <h3 className="text-sm font-semibold text-gray-700 mb-3">What is UNM Platform?</h3>
          <p className="text-sm text-muted-foreground leading-relaxed mb-4">
            UNM Platform turns User Needs Mapping into an engineering tool. Upload a{' '}
            <code className="bg-muted px-1 py-0.5 rounded text-xs">.unm</code>{' '}or{' '}
            <code className="bg-muted px-1 py-0.5 rounded text-xs">.unm.yaml</code>{' '}
            model file to visualize your architecture, analyze team cognitive load, and explore capability ownership.
          </p>
          <p className="text-sm text-muted-foreground"><strong className="text-gray-700">Accepted formats:</strong> .unm, .unm.yaml</p>
        </div>
      )}
      </div>
    </div>
  )
}
