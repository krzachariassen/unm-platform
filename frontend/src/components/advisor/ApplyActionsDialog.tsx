import { useState, useCallback, useRef, useEffect } from 'react'
import { X, Loader2, Sparkles, Check, AlertTriangle, ChevronRight } from 'lucide-react'
import { useModel } from '@/lib/model-context'
import { useChangeset } from '@/lib/changeset-context'
import { advisorApi } from '@/services/api'
import type { ExtractedAction } from '@/types/changeset'
import type { ChangeAction } from '@/lib/api'

type Phase = 'idle' | 'extracting' | 'review' | 'done'

function actionIcon(type: string): string {
  if (type.startsWith('add_') || type.startsWith('link_')) return '+'
  if (type.startsWith('remove_') || type.startsWith('unlink_')) return '−'
  if (type.startsWith('move_') || type.startsWith('reassign_')) return '→'
  if (type.startsWith('update_') || type.startsWith('rename_')) return '✎'
  if (type.startsWith('split_')) return '⑂'
  if (type.startsWith('merge_')) return '⊕'
  return '•'
}

function actionLabel(a: ExtractedAction): string {
  const verb = a.type.replace(/_/g, ' ')
  const entity = a.service_name || a.capability_name || a.team_name || a.need_name || ''
  if (a.from_team_name && a.to_team_name) return `${entity} (${a.from_team_name} → ${a.to_team_name})`
  if (a.original_team_name) return `${a.original_team_name} → ${a.new_team_a_name}, ${a.new_team_b_name}`
  return entity || verb
}

function actionCategory(type: string): string {
  if (type.includes('service')) return 'Services'
  if (type.includes('team') || type.includes('merge') || type.includes('split')) return 'Teams'
  if (type.includes('capability')) return 'Capabilities'
  if (type.includes('interaction')) return 'Interactions'
  if (type.includes('need')) return 'Needs'
  return 'Other'
}

interface ApplyActionsDialogProps {
  open: boolean
  onClose: () => void
  advisorResponse: string
}

export function ApplyActionsDialog({ open, onClose, advisorResponse }: ApplyActionsDialogProps) {
  const { modelId } = useModel()
  const { addAction } = useChangeset()

  const [phase, setPhase] = useState<Phase>('idle')
  const [actions, setActions] = useState<ExtractedAction[]>([])
  const [summary, setSummary] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [selected, setSelected] = useState<Set<number>>(new Set())
  const [stagedCount, setStagedCount] = useState(0)
  const abortRef = useRef<AbortController | null>(null)

  useEffect(() => {
    if (open && phase === 'idle' && modelId && advisorResponse) {
      extract()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open])

  useEffect(() => {
    if (!open) {
      setPhase('idle')
      setActions([])
      setSummary('')
      setError(null)
      setSelected(new Set())
      setStagedCount(0)
    }
  }, [open])

  const extract = useCallback(async () => {
    if (!modelId) return
    setPhase('extracting')
    setError(null)
    const abort = new AbortController()
    abortRef.current = abort
    try {
      const resp = await advisorApi.extractActions(modelId, advisorResponse, abort.signal)
      if (!resp.ai_configured) {
        setError('AI is not configured.')
        setPhase('idle')
        return
      }
      setActions(resp.actions ?? [])
      setSummary(resp.summary ?? '')
      setSelected(new Set((resp.actions ?? []).map((_, i) => i)))
      setPhase('review')
    } catch (err) {
      if ((err as Error).name === 'AbortError') return
      setError(err instanceof Error ? err.message : 'Extraction failed')
      setPhase('idle')
    }
  }, [modelId, advisorResponse])

  const toggleAction = useCallback((index: number) => {
    setSelected(prev => {
      const next = new Set(prev)
      if (next.has(index)) next.delete(index)
      else next.add(index)
      return next
    })
  }, [])

  const selectAll = useCallback(() => {
    setSelected(new Set(actions.map((_, i) => i)))
  }, [actions])

  const selectNone = useCallback(() => {
    setSelected(new Set())
  }, [])

  const stageSelected = useCallback(() => {
    let count = 0
    for (const idx of Array.from(selected).sort((a, b) => a - b)) {
      const { reason: _, ...action } = actions[idx]
      addAction(action as ChangeAction)
      count++
    }
    setStagedCount(count)
    setPhase('done')
  }, [selected, actions, addAction])

  if (!open) return null

  const grouped = actions.reduce<Record<string, { action: ExtractedAction; index: number }[]>>((acc, action, i) => {
    const cat = actionCategory(action.type)
    if (!acc[cat]) acc[cat] = []
    acc[cat].push({ action, index: i })
    return acc
  }, {})

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center" style={{ background: 'rgba(0,0,0,0.5)' }}>
      <div
        className="relative w-full max-w-2xl max-h-[80vh] flex flex-col rounded-2xl overflow-hidden"
        style={{ background: '#ffffff', boxShadow: '0 25px 50px -12px rgba(0,0,0,0.25)' }}
      >
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 shrink-0" style={{ borderBottom: '1px solid #e5e7eb' }}>
          <div className="flex items-center gap-3">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg" style={{ background: 'linear-gradient(135deg, #eef2ff 0%, #f5f3ff 100%)' }}>
              <Sparkles size={16} style={{ color: '#6366f1' }} />
            </div>
            <div>
              <h2 className="text-sm font-bold" style={{ color: '#111827' }}>Apply AI Recommendations</h2>
              <p className="text-[11px]" style={{ color: '#9ca3af' }}>
                {phase === 'extracting' ? 'Analyzing recommendations...' :
                 phase === 'review' ? `${actions.length} actions extracted — select which to stage` :
                 phase === 'done' ? `${stagedCount} actions staged` :
                 'Extract and apply structural changes'}
              </p>
            </div>
          </div>
          <button
            type="button"
            onClick={() => { abortRef.current?.abort(); onClose() }}
            className="rounded-md p-1.5 transition-colors hover:bg-gray-100"
            style={{ color: '#9ca3af' }}
          >
            <X size={16} />
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto min-h-0">
          {phase === 'extracting' && (
            <div className="flex flex-col items-center justify-center py-16 gap-4">
              <div className="relative">
                <div className="absolute inset-0 rounded-full animate-ping" style={{ background: '#6366f1', opacity: 0.1 }} />
                <Loader2 size={28} className="animate-spin relative" style={{ color: '#6366f1' }} />
              </div>
              <div className="text-center">
                <p className="text-sm font-semibold" style={{ color: '#111827' }}>Extracting actions from AI response</p>
                <p className="text-[11px] mt-1" style={{ color: '#9ca3af' }}>
                  The AI is parsing recommendations into concrete model changes...
                </p>
              </div>
              <button
                type="button"
                onClick={() => { abortRef.current?.abort(); setPhase('idle'); onClose() }}
                className="text-xs px-3 py-1.5 rounded-md transition-colors hover:bg-gray-100"
                style={{ color: '#6b7280', border: '1px solid #e5e7eb' }}
              >
                Cancel
              </button>
            </div>
          )}

          {phase === 'idle' && error && (
            <div className="flex flex-col items-center justify-center py-16 gap-4">
              <AlertTriangle size={28} style={{ color: '#ef4444' }} />
              <p className="text-sm" style={{ color: '#ef4444' }}>{error}</p>
              <button
                type="button"
                onClick={extract}
                className="text-xs px-4 py-2 rounded-lg font-medium text-white"
                style={{ background: '#111827' }}
              >
                Retry
              </button>
            </div>
          )}

          {phase === 'review' && actions.length === 0 && (
            <div className="flex flex-col items-center justify-center py-16 gap-3">
              <AlertTriangle size={24} style={{ color: '#f59e0b' }} />
              <p className="text-sm font-medium" style={{ color: '#111827' }}>No actionable changes found</p>
              <p className="text-xs text-center max-w-sm" style={{ color: '#6b7280' }}>
                The AI response did not contain specific enough recommendations to extract as structured changes.
                Try asking for more prescriptive advice with specific service and team names.
              </p>
            </div>
          )}

          {phase === 'review' && actions.length > 0 && (
            <div className="px-6 py-4">
              {summary && (
                <div className="mb-4 px-3 py-2.5 rounded-lg text-xs leading-relaxed" style={{ background: '#f0fdf4', border: '1px solid #bbf7d0', color: '#166534' }}>
                  {summary}
                </div>
              )}

              <div className="flex items-center justify-between mb-3">
                <span className="text-xs font-medium" style={{ color: '#6b7280' }}>
                  {selected.size} of {actions.length} selected
                </span>
                <div className="flex items-center gap-2">
                  <button type="button" onClick={selectAll} className="text-[11px] font-medium hover:underline" style={{ color: '#6366f1' }}>
                    Select all
                  </button>
                  <span className="text-[10px]" style={{ color: '#d1d5db' }}>|</span>
                  <button type="button" onClick={selectNone} className="text-[11px] font-medium hover:underline" style={{ color: '#6366f1' }}>
                    Deselect all
                  </button>
                </div>
              </div>

              <div className="space-y-4">
                {Object.entries(grouped).map(([category, items]) => (
                  <div key={category}>
                    <div className="flex items-center gap-2 mb-2">
                      <span className="text-[10px] font-semibold uppercase tracking-wide" style={{ color: '#9ca3af' }}>{category}</span>
                      <div className="flex-1 h-px" style={{ background: '#f3f4f6' }} />
                    </div>
                    <div className="space-y-1">
                      {items.map(({ action, index }) => (
                        <button
                          key={index}
                          type="button"
                          onClick={() => toggleAction(index)}
                          className="w-full flex items-start gap-3 px-3 py-2.5 rounded-lg text-left transition-colors"
                          style={{
                            background: selected.has(index) ? '#f0f9ff' : '#ffffff',
                            border: `1px solid ${selected.has(index) ? '#bfdbfe' : '#f3f4f6'}`,
                          }}
                        >
                          <div
                            className="mt-0.5 shrink-0 flex items-center justify-center w-4 h-4 rounded border transition-colors"
                            style={{
                              background: selected.has(index) ? '#6366f1' : '#ffffff',
                              borderColor: selected.has(index) ? '#6366f1' : '#d1d5db',
                            }}
                          >
                            {selected.has(index) && <Check size={10} className="text-white" strokeWidth={3} />}
                          </div>
                          <div className="min-w-0 flex-1">
                            <div className="flex items-center gap-2">
                              <span
                                className="shrink-0 flex items-center justify-center w-5 h-5 rounded text-[11px] font-mono font-bold"
                                style={{ background: '#f3f4f6', color: '#6b7280' }}
                              >
                                {actionIcon(action.type)}
                              </span>
                              <span className="text-xs font-semibold truncate" style={{ color: '#111827' }}>
                                {action.type.replace(/_/g, ' ')}
                              </span>
                              <ChevronRight size={10} style={{ color: '#d1d5db' }} />
                              <span className="text-xs truncate" style={{ color: '#374151' }}>
                                {actionLabel(action)}
                              </span>
                            </div>
                            {action.reason && (
                              <p className="text-[11px] leading-relaxed mt-1 ml-7" style={{ color: '#6b7280' }}>
                                {action.reason}
                              </p>
                            )}
                          </div>
                        </button>
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {phase === 'done' && (
            <div className="flex flex-col items-center justify-center py-16 gap-4">
              <div className="flex h-12 w-12 items-center justify-center rounded-full" style={{ background: '#dcfce7' }}>
                <Check size={24} style={{ color: '#16a34a' }} />
              </div>
              <div className="text-center">
                <p className="text-sm font-bold" style={{ color: '#111827' }}>
                  {stagedCount} {stagedCount === 1 ? 'action' : 'actions'} staged
                </p>
                <p className="text-xs mt-1" style={{ color: '#6b7280' }}>
                  Review your pending changes in the top bar, then commit when ready.
                </p>
              </div>
            </div>
          )}
        </div>

        {/* Footer */}
        {phase === 'review' && actions.length > 0 && (
          <div className="flex items-center justify-between px-6 py-4 shrink-0" style={{ borderTop: '1px solid #e5e7eb' }}>
            <button
              type="button"
              onClick={onClose}
              className="text-xs px-3 py-1.5 rounded-md transition-colors hover:bg-gray-100"
              style={{ color: '#6b7280', border: '1px solid #e5e7eb' }}
            >
              Cancel
            </button>
            <button
              type="button"
              onClick={stageSelected}
              disabled={selected.size === 0}
              className="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-semibold text-white transition-all disabled:opacity-40"
              style={{ background: '#111827' }}
            >
              <Sparkles size={14} />
              Stage {selected.size} {selected.size === 1 ? 'action' : 'actions'}
            </button>
          </div>
        )}

        {phase === 'done' && (
          <div className="flex items-center justify-end px-6 py-4 shrink-0" style={{ borderTop: '1px solid #e5e7eb' }}>
            <button
              type="button"
              onClick={onClose}
              className="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-semibold text-white"
              style={{ background: '#111827' }}
            >
              Done
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
