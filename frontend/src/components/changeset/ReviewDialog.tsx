import { useState, useCallback, useEffect, useRef, useMemo } from 'react'
import { X, Trash2, ArrowRight, Check, Loader2, AlertTriangle, Sparkles, Bot } from 'lucide-react'
import { useModel } from '@/lib/model-context'
import { useChangeset } from '@/lib/changeset-context'
import { useAIEnabled } from '@/hooks/useAIEnabled'
import { api } from '@/lib/api'
import type { ImpactDelta } from '@/types/changeset'
import { ImpactPanel } from './ImpactPanel'
import { Prose } from '@/components/ui/prose'

type Tab = 'changes' | 'impact' | 'ai'
type CommitPhase = 'idle' | 'creating' | 'ready' | 'committing' | 'done'

function actionLabel(a: { type: string; service_name?: string; capability_name?: string; team_name?: string; need_name?: string; from_team_name?: string; to_team_name?: string }): string {
  const verb = a.type.replace(/_/g, ' ')
  const entity = a.service_name || a.capability_name || a.team_name || a.need_name || ''
  if (a.from_team_name && a.to_team_name) return `${verb}: ${entity} (${a.from_team_name} → ${a.to_team_name})`
  return entity ? `${verb}: ${entity}` : verb
}

function actionIcon(type: string): string {
  if (type.startsWith('add_') || type.startsWith('link_')) return '+'
  if (type.startsWith('remove_') || type.startsWith('unlink_')) return '−'
  if (type.startsWith('move_') || type.startsWith('reassign_')) return '→'
  if (type.startsWith('update_') || type.startsWith('rename_')) return '✎'
  if (type.startsWith('split_')) return '⑂'
  if (type.startsWith('merge_')) return '⊕'
  return '•'
}

interface ReviewDialogProps {
  open: boolean
  onClose: () => void
}

export function ReviewDialog({ open, onClose }: ReviewDialogProps) {
  const { modelId, parseResult, setModel } = useModel()
  const { actions, description, removeAction, discardAll } = useChangeset()
  const aiEnabled = useAIEnabled()

  const [tab, setTab] = useState<Tab>('changes')
  const [phase, setPhase] = useState<CommitPhase>('idle')
  const [changesetId, setChangesetId] = useState<string | null>(null)
  const [impact, setImpact] = useState<ImpactDelta[] | null>(null)
  const [error, setError] = useState<string | null>(null)

  const [aiAnswer, setAiAnswer] = useState<string | null>(null)
  const [aiLoading, setAiLoading] = useState(false)
  const [aiError, setAiError] = useState<string | null>(null)

  const actionsFingerprint = useMemo(
    () => actions.map(a => `${a.type}:${a.service_name || a.capability_name || a.team_name || a.need_name || ''}`).join('|'),
    [actions]
  )
  const prevFingerprint = useRef(actionsFingerprint)

  useEffect(() => {
    if (actionsFingerprint !== prevFingerprint.current) {
      prevFingerprint.current = actionsFingerprint
      setPhase('idle'); setChangesetId(null); setImpact(null); setError(null)
      setAiAnswer(null); setAiError(null)
    }
  }, [actionsFingerprint])

  useEffect(() => {
    if (!open) return
    const handler = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose() }
    document.addEventListener('keydown', handler)
    return () => document.removeEventListener('keydown', handler)
  }, [open, onClose])

  const reset = useCallback(() => {
    setPhase('idle'); setChangesetId(null); setImpact(null); setError(null)
  }, [])

  const handleCreateAndPreview = useCallback(async () => {
    if (!modelId || actions.length === 0) return
    setPhase('creating'); setError(null); setImpact(null)
    try {
      const csId = `batch-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`
      const cs = await api.createChangeset(modelId, { id: csId, description: description || 'Batch edit', actions })
      const result = await api.getChangesetImpact(modelId, cs.id)
      setChangesetId(cs.id); setImpact(result.deltas); setPhase('ready'); setTab('impact')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to compute impact'); setPhase('idle')
    }
  }, [modelId, actions, description])

  const handleCommit = useCallback(async () => {
    if (!modelId || !changesetId) return
    setPhase('committing'); setError(null)
    try {
      const result = await api.commitChangeset(modelId, changesetId)
      if (result.validation.valid) {
        if (parseResult) {
          setModel(modelId, {
            ...parseResult,
            system_name: result.system_name,
            summary: {
              actors: result.summary.actors ?? parseResult.summary.actors,
              needs: result.summary.needs ?? parseResult.summary.needs,
              capabilities: result.summary.capabilities ?? parseResult.summary.capabilities,
              services: result.summary.services ?? parseResult.summary.services,
              teams: result.summary.teams ?? parseResult.summary.teams,
            },
          })
        }
        setPhase('done')
        setTimeout(() => { discardAll(); reset(); setAiAnswer(null); setAiError(null); onClose() }, 1200)
      } else {
        setError(result.validation.errors?.join('; ') || 'Validation failed')
        setPhase('ready')
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Commit failed'); setPhase('ready')
    }
  }, [modelId, changesetId, parseResult, setModel, discardAll, reset, onClose])

  const handleDiscard = useCallback(() => {
    if (actions.length > 0 && !window.confirm(`Discard ${actions.length} pending change${actions.length === 1 ? '' : 's'}?`)) return
    discardAll(); reset(); setAiAnswer(null); setAiError(null); onClose()
  }, [actions.length, discardAll, reset, onClose])

  const handleRemoveAction = useCallback((i: number) => {
    removeAction(i)
  }, [removeAction])

  const handleAskAI = useCallback(async () => {
    if (!modelId || actions.length === 0) return
    setAiLoading(true); setAiError(null); setAiAnswer(null)
    const changeList = actions.map(a => actionLabel(a)).join('; ')
    const question = `I am about to make these changes to the architecture: ${changeList}. Analyze the structural impact: will this improve or worsen team boundaries, cognitive load, fragmentation, or coupling? Are there risks or better alternatives? Be concise.`
    try {
      const resp = await api.askAdvisor(modelId, question, 'value-stream')
      setAiAnswer(resp.answer)
    } catch (err) {
      setAiError(err instanceof Error ? err.message : 'AI analysis failed')
    } finally {
      setAiLoading(false)
    }
  }, [modelId, actions])

  if (!open) return null

  const isWorking = phase === 'creating' || phase === 'committing'
  const canCommit = phase === 'ready' && !!changesetId

  const tabs: { id: Tab; label: string; disabled: boolean }[] = [
    { id: 'changes', label: 'Changes', disabled: false },
    { id: 'impact', label: 'Impact', disabled: false },
    ...(aiEnabled ? [{ id: 'ai' as Tab, label: 'AI Insights', disabled: false }] : []),
  ]

  return (
    <>
      <div className="fixed inset-0 z-50" style={{ background: 'rgba(0,0,0,0.4)' }} onClick={onClose} />
      <div className="fixed inset-0 z-50 flex items-center justify-center pointer-events-none">
        <div
          className="pointer-events-auto w-[720px] max-h-[80vh] rounded-xl flex flex-col overflow-hidden"
          style={{ background: '#ffffff', border: '1px solid #e5e7eb', boxShadow: '0 25px 50px -12px rgba(0,0,0,0.25)' }}
          onClick={e => e.stopPropagation()}
        >
          {/* Header */}
          <div className="flex items-center justify-between px-6 py-4 shrink-0" style={{ borderBottom: '1px solid #f3f4f6' }}>
            <div className="flex items-center gap-3">
              <h2 className="text-base font-semibold" style={{ color: '#111827' }}>Review Changes</h2>
              <span className="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-bold" style={{ background: '#111827', color: '#ffffff' }}>
                {actions.length}
              </span>
              {phase === 'done' && (
                <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-semibold" style={{ background: '#dcfce7', color: '#15803d' }}>
                  <Check size={10} /> Committed
                </span>
              )}
            </div>
            <button type="button" onClick={onClose} className="rounded-md p-1.5 transition-colors hover:bg-gray-100" style={{ color: '#9ca3af' }}>
              <X size={16} />
            </button>
          </div>

          {/* Tab bar */}
          <div className="flex gap-0 px-6 shrink-0" style={{ borderBottom: '1px solid #f3f4f6' }}>
            {tabs.map(t => (
              <button
                key={t.id}
                type="button"
                onClick={() => !t.disabled && setTab(t.id)}
                disabled={t.disabled}
                className="relative px-4 py-2.5 text-xs font-medium transition-colors disabled:opacity-30 disabled:cursor-not-allowed"
                style={{ color: tab === t.id ? '#111827' : '#6b7280' }}
              >
                {t.label}
                {t.id === 'ai' && <Sparkles size={10} className="inline ml-1" style={{ color: '#2563eb' }} />}
                {tab === t.id && (
                  <span className="absolute bottom-0 left-4 right-4 h-[2px] rounded-full" style={{ background: '#111827' }} />
                )}
                {t.id === 'impact' && impact && (
                  <span className="ml-1.5 inline-flex items-center px-1.5 py-0.5 rounded-full text-[9px] font-bold" style={{ background: '#f3f4f6', color: '#6b7280' }}>
                    {impact.length}
                  </span>
                )}
              </button>
            ))}
          </div>

          {/* Content */}
          <div className="flex-1 overflow-y-auto px-6 py-4" style={{ minHeight: 200 }}>
            {/* Changes tab */}
            {tab === 'changes' && (
              <div className="space-y-1.5">
                {actions.map((a, i) => (
                  <div key={i} className="flex items-center gap-3 rounded-lg px-3 py-2.5 group" style={{ background: '#f9fafb', border: '1px solid #f3f4f6' }}>
                    <span className="shrink-0 w-5 h-5 rounded flex items-center justify-center text-[11px] font-bold" style={{ background: '#e5e7eb', color: '#6b7280' }}>
                      {actionIcon(a.type)}
                    </span>
                    <div className="flex-1 min-w-0">
                      <span className="text-xs font-medium block truncate" style={{ color: '#374151' }}>{actionLabel(a)}</span>
                    </div>
                    <button
                      type="button"
                      onClick={() => handleRemoveAction(i)}
                      className="shrink-0 rounded p-1 opacity-0 group-hover:opacity-100 transition-opacity hover:bg-gray-200"
                      style={{ color: '#9ca3af' }}
                    >
                      <X size={12} />
                    </button>
                  </div>
                ))}
                {actions.length === 0 && (
                  <div className="text-center py-10">
                    <p className="text-sm" style={{ color: '#9ca3af' }}>No pending changes</p>
                  </div>
                )}
              </div>
            )}

            {/* Impact tab */}
            {tab === 'impact' && (
              <div>
                {phase === 'creating' && (
                  <div className="flex flex-col items-center justify-center py-10 gap-2">
                    <Loader2 size={20} className="animate-spin" style={{ color: '#9ca3af' }} />
                    <p className="text-xs" style={{ color: '#9ca3af' }}>Computing impact analysis…</p>
                  </div>
                )}
                {phase === 'idle' && (
                  <div className="flex flex-col items-center justify-center py-10 gap-3">
                    <p className="text-sm" style={{ color: '#6b7280' }}>Run impact analysis to see how your changes affect the architecture.</p>
                    <button type="button" onClick={handleCreateAndPreview}
                      className="flex items-center gap-1.5 rounded-md px-4 py-2 text-xs font-semibold text-white"
                      style={{ background: '#111827' }}>
                      <ArrowRight size={13} /> Analyze impact
                    </button>
                  </div>
                )}
                {impact && <ImpactPanel deltas={impact} changesetId={changesetId ?? ''} />}
              </div>
            )}

            {/* AI Insights tab */}
            {tab === 'ai' && (
              <div>
                {!aiAnswer && !aiLoading && !aiError && (
                  <div className="flex flex-col items-center justify-center py-10 gap-3">
                    <div className="rounded-lg p-3" style={{ background: '#eff6ff' }}>
                      <Bot size={24} style={{ color: '#2563eb' }} />
                    </div>
                    <div className="text-center">
                      <p className="text-sm font-medium" style={{ color: '#111827' }}>AI Architecture Review</p>
                      <p className="text-xs mt-1 max-w-sm" style={{ color: '#6b7280' }}>
                        Get an AI analysis of your {actions.length} pending {actions.length === 1 ? 'change' : 'changes'} — covering team boundaries, cognitive load, fragmentation, coupling, and risks.
                      </p>
                    </div>
                    <button type="button" onClick={handleAskAI}
                      className="flex items-center gap-1.5 rounded-md px-4 py-2 text-xs font-semibold text-white"
                      style={{ background: '#2563eb' }}>
                      <Sparkles size={13} /> Get AI Insights
                    </button>
                  </div>
                )}
                {aiLoading && (
                  <div className="flex flex-col items-center justify-center py-10 gap-2">
                    <Loader2 size={20} className="animate-spin" style={{ color: '#2563eb' }} />
                    <p className="text-xs" style={{ color: '#6b7280' }}>AI is analyzing your changes…</p>
                  </div>
                )}
                {aiError && (
                  <div className="flex items-start gap-2 rounded-lg px-4 py-3" style={{ background: '#fef2f2', border: '1px solid #fca5a5' }}>
                    <AlertTriangle size={13} className="shrink-0 mt-0.5" style={{ color: '#dc2626' }} />
                    <div>
                      <p className="text-xs font-medium" style={{ color: '#991b1b' }}>{aiError}</p>
                      <button type="button" onClick={handleAskAI} className="text-[10px] mt-1 underline" style={{ color: '#dc2626' }}>Retry</button>
                    </div>
                  </div>
                )}
                {aiAnswer && (
                  <div>
                    <div className="flex items-center gap-2 mb-3">
                      <Bot size={14} style={{ color: '#2563eb' }} />
                      <span className="text-xs font-semibold" style={{ color: '#111827' }}>AI Analysis</span>
                      <button type="button" onClick={handleAskAI} disabled={aiLoading}
                        className="ml-auto text-[10px] rounded px-2 py-0.5 transition-colors hover:bg-blue-50" style={{ color: '#2563eb' }}>
                        Re-analyze
                      </button>
                    </div>
                    <div className="rounded-lg px-4 py-3" style={{ background: '#f8fafc', border: '1px solid #e2e8f0' }}>
                      <Prose compact>{aiAnswer}</Prose>
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>

          {/* Error bar */}
          {error && (
            <div className="mx-6 mb-3 flex items-start gap-2 rounded-lg px-3 py-2" style={{ background: '#fef2f2', border: '1px solid #fca5a5' }}>
              <AlertTriangle size={12} className="shrink-0 mt-0.5" style={{ color: '#dc2626' }} />
              <span className="text-xs" style={{ color: '#991b1b' }}>{error}</span>
            </div>
          )}

          {/* Footer */}
          <div className="flex items-center justify-between px-6 py-3 shrink-0" style={{ borderTop: '1px solid #f3f4f6' }}>
            <button type="button" onClick={handleDiscard} disabled={isWorking}
              className="flex items-center gap-1.5 rounded-md px-3 py-2 text-xs transition-colors hover:bg-gray-100 disabled:opacity-50" style={{ color: '#6b7280' }}>
              <Trash2 size={12} /> Discard all
            </button>

            <div className="flex items-center gap-2">
              {/* Save / analyze button — only on changes tab when not yet analyzed */}
              {tab === 'changes' && phase === 'idle' && (
                <button type="button" onClick={handleCreateAndPreview} disabled={isWorking || actions.length === 0}
                  className="flex items-center gap-1.5 rounded-md px-4 py-2 text-xs font-semibold text-white transition-colors disabled:opacity-50"
                  style={{ background: '#111827' }}>
                  {isWorking ? <Loader2 size={13} className="animate-spin" /> : <ArrowRight size={13} />}
                  Save changes
                </button>
              )}

              {/* Commit button — available once impact is computed */}
              {canCommit && (
                <button type="button" onClick={handleCommit} disabled={isWorking}
                  className="flex items-center gap-1.5 rounded-md px-4 py-2 text-xs font-semibold text-white transition-colors disabled:opacity-50"
                  style={{ background: '#059669' }}>
                  {isWorking ? <Loader2 size={13} className="animate-spin" /> : <Check size={13} />}
                  {isWorking ? 'Committing…' : 'Confirm & commit'}
                </button>
              )}

              {/* Done state */}
              {phase === 'done' && (
                <span className="flex items-center gap-1.5 px-4 py-2 text-xs font-semibold" style={{ color: '#059669' }}>
                  <Check size={13} /> Committed successfully
                </span>
              )}
            </div>
          </div>
        </div>
      </div>
    </>
  )
}
