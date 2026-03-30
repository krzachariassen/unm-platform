import { useState, useCallback, useEffect, useRef } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { Upload, AlertCircle, CheckCircle2, FlaskConical, Loader2, Sparkles } from 'lucide-react'
import { api } from '@/lib/api'
import { useModel } from '@/lib/model-context'
import { useAIEnabled } from '@/hooks/useAIEnabled'
import { getRuntimeConfig } from '@/lib/runtimeConfig'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import type { ParseResponse } from '@/lib/api'

// ── Progress step definitions ───────────────────────────────────────────────

type StepStatus = 'idle' | 'active' | 'done' | 'error'

interface Step {
  id: string
  label: string
  sublabel?: string
}

const BASE_STEPS: Step[] = [
  { id: 'read',    label: 'Reading file' },
  { id: 'parse',   label: 'Parsing & validating model' },
  { id: 'analyse', label: 'Analysing architecture' },
]

const AI_STEP: Step = { id: 'ai', label: 'Generating AI insights', sublabel: 'Signals, needs, capabilities, ownership…' }

function StepRow({ step, status }: { step: Step; status: StepStatus }) {
  return (
    <div className="flex items-start gap-3">
      <div className="flex-shrink-0 mt-0.5">
        {status === 'done' && (
          <CheckCircle2 size={16} style={{ color: '#22c55e' }} />
        )}
        {status === 'active' && (
          <Loader2 size={16} className="animate-spin" style={{ color: '#2563eb' }} />
        )}
        {status === 'error' && (
          <AlertCircle size={16} style={{ color: '#ef4444' }} />
        )}
        {status === 'idle' && (
          <div className="w-4 h-4 rounded-full border-2" style={{ borderColor: '#d1d5db' }} />
        )}
      </div>
      <div>
        <p className="text-sm font-medium" style={{
          color: status === 'active' ? '#1d4ed8'
               : status === 'done'   ? '#15803d'
               : status === 'error'  ? '#b91c1c'
               : '#9ca3af',
        }}>
          {step.label}
        </p>
        {step.sublabel && status === 'active' && (
          <p className="text-xs mt-0.5" style={{ color: '#93c5fd' }}>{step.sublabel}</p>
        )}
      </div>
    </div>
  )
}

// ── Main component ───────────────────────────────────────────────────────────

export function UploadPage() {
  const [dragging, setDragging] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [parseResult, setParseResult] = useState<ParseResponse | null>(null)
  const [parseWarnings, setParseWarnings] = useState<string[]>([])
  const [debugEnabled, setDebugEnabled] = useState(false)
  const aiEnabled = useAIEnabled()
  const STEPS = aiEnabled ? [...BASE_STEPS, AI_STEP] : BASE_STEPS

  // Step tracking
  const [stepStatuses, setStepStatuses] = useState<Record<string, StepStatus>>({
    read: 'idle', parse: 'idle', analyse: 'idle', ai: 'idle',
  })
  const [processing, setProcessing] = useState(false)

  const { modelId, parseResult: loadedParseResult, setModel } = useModel()
  const navigate = useNavigate()
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const setStep = (id: string, status: StepStatus) =>
    setStepStatuses(prev => ({ ...prev, [id]: status }))

  // Clean up polling on unmount
  useEffect(() => () => { if (pollRef.current) clearInterval(pollRef.current) }, [])

  useEffect(() => {
    getRuntimeConfig().then(cfg => setDebugEnabled(cfg.features?.debug_routes ?? false))
  }, [])

  const startPollingInsights = useCallback((id: string) => {
    setStep('ai', 'active')

    const poll = async () => {
      try {
        const status = await api.getInsightsStatus(id)
        if (status.all_ready) {
          if (pollRef.current) clearInterval(pollRef.current)
          pollRef.current = null
          setStep('ai', 'done')
          setTimeout(() => navigate('/dashboard'), 400)
        }
      } catch {
        // Transient error — keep polling
      }
    }

    poll() // immediate first check
    pollRef.current = setInterval(poll, 2000)
  }, [navigate])

  const runUploadFlow = useCallback(async (content: string, filename: string) => {
    setError(null)
    setProcessing(true)
    setStepStatuses({ read: 'done', parse: 'idle', analyse: 'idle', ai: 'idle' })
    setParseResult(null)
    setParseWarnings([])

    const format = filename.endsWith('.unm') && !filename.endsWith('.unm.yaml') ? 'dsl' as const : undefined

    // Step: Parse
    setStep('parse', 'active')
    let parsed: ParseResponse
    try {
      parsed = await api.parseModel(content, modelId ?? undefined, format)
    } catch (e) {
      setStep('parse', 'error')
      setError(e instanceof Error ? e.message : 'Upload failed')
      setProcessing(false)
      return
    }

    setParseResult(parsed)

    if (!parsed.validation.is_valid) {
      setStep('parse', 'error')
      setProcessing(false)
      return
    }

    setStep('parse', 'done')
    setModel(parsed.id, parsed)
    if (parsed.warnings && parsed.warnings.length > 0) {
      setParseWarnings(parsed.warnings)
    }

    // Step: Analyse (synchronous in backend — parse already ran all analyzers)
    setStep('analyse', 'active')
    await new Promise(r => setTimeout(r, 200)) // brief visual pause
    setStep('analyse', 'done')

    if (aiEnabled) {
      startPollingInsights(parsed.id)
    } else {
      setTimeout(() => navigate('/dashboard'), 400)
    }
  }, [modelId, setModel, startPollingInsights, aiEnabled, navigate])

  const loadExample = useCallback(async () => {
    setError(null)
    setProcessing(true)
    setStepStatuses({ read: 'done', parse: 'idle', analyse: 'idle', ai: 'idle' })
    setParseResult(null)
    setParseWarnings([])

    setStep('parse', 'active')
    let parsed: ParseResponse
    try {
      parsed = await api.loadExample(modelId ?? undefined)
    } catch (e) {
      setStep('parse', 'error')
      setError(e instanceof Error ? e.message : 'Failed to load example')
      setProcessing(false)
      return
    }

    setParseResult(parsed)
    if (!parsed.validation.is_valid) {
      setStep('parse', 'error')
      setProcessing(false)
      return
    }

    setStep('parse', 'done')
    setModel(parsed.id, parsed)
    if (parsed.warnings && parsed.warnings.length > 0) {
      setParseWarnings(parsed.warnings)
    }
    setStep('analyse', 'active')
    await new Promise(r => setTimeout(r, 200))
    setStep('analyse', 'done')
    if (aiEnabled) {
      startPollingInsights(parsed.id)
    } else {
      setTimeout(() => navigate('/dashboard'), 400)
    }
  }, [modelId, setModel, startPollingInsights, aiEnabled, navigate])

  const onDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setDragging(false)
    const file = e.dataTransfer.files[0]
    if (!file) return
    setStep('read', 'active')
    const reader = new FileReader()
    reader.onload = (ev) => {
      setStep('read', 'done')
      runUploadFlow(ev.target?.result as string, file.name)
    }
    reader.readAsText(file)
  }, [runUploadFlow])

  const onFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    setStep('read', 'active')
    const reader = new FileReader()
    reader.onload = (ev) => {
      setStep('read', 'done')
      runUploadFlow(ev.target?.result as string, file.name)
    }
    reader.readAsText(file)
  }

  const showProgress = processing || stepStatuses.parse !== 'idle'

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      <div>
        <h2 className="text-2xl font-bold">Upload Model</h2>
        <p className="text-muted-foreground mt-1">Upload a .unm.yaml or .unm file to explore your architecture</p>
      </div>

      {/* UI-11: Active model banner */}
      {modelId && loadedParseResult && !showProgress && (
        <div style={{
          background: '#f0fdf4', border: '1px solid #bbf7d0', borderRadius: 8,
          padding: '12px 16px', display: 'flex',
          alignItems: 'center', justifyContent: 'space-between'
        }}>
          <div style={{display:'flex', alignItems:'center', gap:8}}>
            <span style={{color:'#16a34a', fontSize:16}}>&#10003;</span>
            <span style={{fontSize:14, color:'#15803d', fontWeight:500}}>
              Model loaded: {loadedParseResult.system_name ?? 'Unknown'}
            </span>
          </div>
          <Link to="/dashboard" style={{fontSize:13, color:'#16a34a', fontWeight:500, textDecoration:'none'}}>
            Go to Dashboard →
          </Link>
        </div>
      )}

      {!showProgress && (
        <>
          <label
            htmlFor="file-input"
            onDragOver={(e) => { e.preventDefault(); setDragging(true) }}
            onDragLeave={() => setDragging(false)}
            onDrop={onDrop}
            className={cn(
              'block border-2 border-dashed rounded-xl p-12 text-center transition-colors cursor-pointer',
              dragging ? 'border-primary bg-primary/5' : 'border-border hover:border-primary/50',
            )}
          >
            <Upload className="mx-auto mb-3 text-muted-foreground" size={32} />
            <p className="text-sm font-medium">Drop your .unm.yaml or .unm file here</p>
            <p className="text-xs text-muted-foreground mt-1">or click to browse</p>
            <input id="file-input" type="file" accept=".yaml,.yml,.unm" className="sr-only" onChange={onFileChange} />
          </label>

          {debugEnabled && (
            <>
              <div className="flex items-center gap-3">
                <div className="flex-1 h-px bg-border" />
                <span className="text-xs text-muted-foreground">or</span>
                <div className="flex-1 h-px bg-border" />
              </div>

              <Button
                variant="outline"
                className="w-full gap-2"
                onClick={loadExample}
              >
                <FlaskConical size={16} />
                Load Example Model
              </Button>
            </>
          )}
        </>
      )}

      {showProgress && (
        <div className="rounded-xl border border-border p-6 space-y-5">
          <div className="flex items-center gap-2 mb-2">
            <Sparkles size={16} style={{ color: '#2563eb' }} />
            <span className="text-sm font-semibold">Processing model</span>
          </div>
          <div className="space-y-4">
            {STEPS.map(step => (
              <StepRow key={step.id} step={step} status={stepStatuses[step.id]} />
            ))}
          </div>

          {/* Validation summary after parse */}
          {parseResult && !parseResult.validation.is_valid && (
            <div className="mt-4 space-y-2">
              {parseResult.validation.errors.map((e, i) => (
                <div key={i} className="text-xs text-destructive/80 flex gap-1">
                  <span className="font-mono">[{e.code}]</span> {e.message}
                </div>
              ))}
              <Button variant="outline" size="sm" className="mt-2" onClick={() => {
                setProcessing(false)
                setStepStatuses({ read: 'idle', parse: 'idle', analyse: 'idle', ai: 'idle' })
                setParseResult(null)
              }}>
                Try another file
              </Button>
            </div>
          )}
        </div>
      )}

      {/* Reference warnings from parser */}
      {parseWarnings.length > 0 && (
        <div style={{
          background: '#fffbeb', border: '1px solid #fcd34d', borderRadius: 8, padding: '12px 16px'
        }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
            <span style={{ color: '#d97706', fontSize: 16 }}>⚠</span>
            <span style={{ fontSize: 14, fontWeight: 600, color: '#92400e' }}>
              Reference Warnings ({parseWarnings.length})
            </span>
          </div>
          <ul style={{ margin: 0, paddingLeft: 20, fontSize: 13, color: '#78350f' }}>
            {parseWarnings.map((w, i) => <li key={i}>{w}</li>)}
          </ul>
          <p style={{ fontSize: 12, color: '#92400e', marginTop: 8, marginBottom: 0 }}>
            These references were not found in the model. Check spelling of capability, service, or team names.
          </p>
        </div>
      )}

      {error && (
        <div
          className="rounded-xl p-4 flex gap-2"
          style={{ border: '1px solid #fca5a5', background: '#fef2f2' }}
        >
          <AlertCircle size={16} style={{ color: '#ef4444', flexShrink: 0, marginTop: 2 }} />
          <div className="flex-1">
            <span className="text-sm" style={{ color: '#b91c1c' }}>{error}</span>
            <Button variant="outline" size="sm" className="mt-2 block" onClick={() => {
              setError(null)
              setProcessing(false)
              setStepStatuses({ read: 'idle', parse: 'idle', analyse: 'idle', ai: 'idle' })
            }}>
              Try again
            </Button>
          </div>
        </div>
      )}

      {/* UI-65: Getting Started section */}
      {!showProgress && (
        <div style={{marginTop: 48, borderTop: '1px solid #e5e7eb', paddingTop: 32}}>
          <h3 style={{fontSize: 14, fontWeight: 600, color: '#374151', marginBottom: 12}}>
            What is UNM Platform?
          </h3>
          <p style={{fontSize: 13, color: '#6b7280', lineHeight: 1.6, marginBottom: 16}}>
            UNM Platform turns User Needs Mapping into an engineering tool. Upload a{' '}
            <code style={{background:'#f3f4f6', padding:'1px 4px', borderRadius:3}}>.unm.yaml</code>{' '}
            model file to visualize your architecture, analyze team cognitive load, and explore
            capability ownership across your organization.
          </p>
          <div style={{display:'flex', gap: 16, flexWrap:'wrap'}}>
            <div style={{fontSize:13, color:'#6b7280'}}>
              <strong style={{color:'#374151'}}>Accepted formats:</strong> .unm.yaml, .unm
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
