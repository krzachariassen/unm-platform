import { useState, useCallback } from 'react'
import { Search, X, ChevronDown, ChevronUp, Check, Loader2, AlertTriangle, Trash2, ArrowRight } from 'lucide-react'
import { useModel } from '@/lib/model-context'
import { useSearch } from '@/lib/search-context'
import { useChangeset } from '@/lib/changeset-context'
import { api } from '@/lib/api'
import type { ImpactDelta } from '@/types/changeset'
import { ImpactPanel } from '@/components/changeset/ImpactPanel'
import { cn } from '@/lib/utils'

type Phase = 'list' | 'saving' | 'confirm' | 'committing' | 'done'

function actionLabel(a: { type: string; service_name?: string; capability_name?: string; team_name?: string; need_name?: string; from_team_name?: string; to_team_name?: string }): string {
  const verb = a.type.replace(/_/g, ' ')
  const entity = a.service_name || a.capability_name || a.team_name || a.need_name || ''
  if (a.from_team_name && a.to_team_name) return `${verb}: ${entity} (${a.from_team_name} → ${a.to_team_name})`
  return entity ? `${verb}: ${entity}` : verb
}

export function TopBar() {
  const { modelId, parseResult, setModel } = useModel()
  const { query, setQuery } = useSearch()
  const { actions, removeAction, discardAll, description } = useChangeset()

  const [open, setOpen] = useState(false)
  const [phase, setPhase] = useState<Phase>('list')
  const [changesetId, setChangesetId] = useState<string | null>(null)
  const [impact, setImpact] = useState<ImpactDelta[] | null>(null)
  const [error, setError] = useState<string | null>(null)

  const reset = useCallback(() => {
    setPhase('list'); setChangesetId(null); setImpact(null); setError(null)
  }, [])

  const handleSave = useCallback(async () => {
    if (!modelId || actions.length === 0) return
    setPhase('saving'); setError(null); setImpact(null)
    try {
      const csId = `batch-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`
      const cs = await api.createChangeset(modelId, { id: csId, description: description || 'Batch edit', actions })
      const result = await api.getChangesetImpact(modelId, cs.id)
      setChangesetId(cs.id); setImpact(result.deltas); setPhase('confirm')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create changeset'); setPhase('list')
    }
  }, [modelId, actions, description])

  const handleConfirm = useCallback(async () => {
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
        setTimeout(() => { discardAll(); reset(); setOpen(false) }, 1200)
      } else {
        setError(result.validation.errors?.join('; ') || 'Validation failed')
        setPhase('confirm')
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Commit failed'); setPhase('confirm')
    }
  }, [modelId, changesetId, parseResult, setModel, discardAll, reset])

  const handleDiscard = useCallback(() => {
    if (actions.length > 0 && !window.confirm(`Discard ${actions.length} pending change${actions.length === 1 ? '' : 's'}?`)) return
    discardAll(); reset(); setOpen(false)
  }, [actions.length, discardAll, reset])

  const handleRemoveAction = useCallback((i: number) => {
    removeAction(i); reset()
  }, [removeAction, reset])

  const isWorking = phase === 'saving' || phase === 'committing'

  return (
    <header className="flex items-center h-14 px-5 gap-3 shrink-0" style={{ background: '#ffffff', borderBottom: '1px solid #e5e7eb' }}>
      {parseResult ? (
        <div className="flex items-center gap-2">
          <span className="text-sm font-semibold" style={{ color: '#111827' }}>{parseResult.system_name}</span>
          <span className={cn(
            'inline-flex items-center px-2 py-0.5 rounded text-[10px] font-semibold',
            parseResult.validation.is_valid ? 'text-green-700' : 'text-red-700'
          )} style={{ background: parseResult.validation.is_valid ? '#dcfce7' : '#fef2f2' }}>
            {parseResult.validation.is_valid ? 'Valid' : 'Invalid'}
          </span>
          {parseResult.validation.warnings.length > 0 && (
            <span className="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-semibold" style={{ background: '#fef3c7', color: '#92400e' }}>
              {parseResult.validation.warnings.length} {parseResult.validation.warnings.length === 1 ? 'warning' : 'warnings'}
            </span>
          )}
        </div>
      ) : (
        <span className="text-sm" style={{ color: '#9ca3af' }}>No model loaded</span>
      )}

      <div className="flex items-center gap-3 ml-auto">
        {actions.length > 0 && (
          <div className="relative">
            <button
              type="button"
              onClick={() => setOpen(o => !o)}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-colors"
              style={phase === 'done'
                ? { background: '#dcfce7', color: '#15803d', border: '1px solid #bbf7d0' }
                : { background: '#111827', color: '#ffffff', border: '1px solid #111827' }
              }
            >
              {phase === 'done' ? (
                <><Check size={12} /> Saved</>
              ) : (
                <>
                  <span className="inline-flex items-center justify-center w-[18px] h-[18px] rounded-full text-[10px] font-bold" style={{ background: 'rgba(255,255,255,0.2)' }}>
                    {actions.length}
                  </span>
                  {actions.length === 1 ? 'change' : 'changes'}
                  {open ? <ChevronUp size={11} /> : <ChevronDown size={11} />}
                </>
              )}
            </button>

            {open && phase !== 'done' && (
              <>
                <div className="fixed inset-0 z-40" onClick={() => { setOpen(false); if (phase === 'list') reset() }} />
                <div className="absolute right-0 top-full mt-1.5 z-50 w-[400px] rounded-xl border shadow-2xl overflow-hidden" style={{ background: '#ffffff', borderColor: '#e5e7eb' }}>

                  {/* Phase: list — show changes + save button */}
                  {(phase === 'list' || phase === 'saving') && (
                    <>
                      <div className="px-4 pt-3 pb-2 flex items-center justify-between">
                        <span className="text-[11px] font-semibold" style={{ color: '#111827' }}>
                          {actions.length} pending {actions.length === 1 ? 'change' : 'changes'}
                        </span>
                        <button type="button" onClick={handleDiscard} className="text-[10px] transition-colors hover:underline" style={{ color: '#dc2626' }}>
                          Discard all
                        </button>
                      </div>

                      <div className="max-h-[260px] overflow-y-auto px-4 pb-3">
                        <div className="space-y-1">
                          {actions.map((a, i) => (
                            <div key={i} className="flex items-center gap-2 rounded-lg px-2.5 py-2 text-[11px] group" style={{ background: '#f9fafb', border: '1px solid #f3f4f6' }}>
                              <span className="flex-1 truncate" style={{ color: '#374151' }}>{actionLabel(a)}</span>
                              <button type="button" onClick={() => handleRemoveAction(i)}
                                className="shrink-0 rounded p-0.5 opacity-0 group-hover:opacity-100 transition-opacity" style={{ color: '#9ca3af' }}>
                                <X size={10} />
                              </button>
                            </div>
                          ))}
                        </div>
                      </div>

                      {error && (
                        <div className="mx-4 mb-3 flex items-start gap-2 rounded-lg px-3 py-2" style={{ background: '#fef2f2', border: '1px solid #fca5a5' }}>
                          <AlertTriangle size={11} className="shrink-0 mt-0.5" style={{ color: '#dc2626' }} />
                          <span className="text-[11px]" style={{ color: '#991b1b' }}>{error}</span>
                        </div>
                      )}

                      <div className="px-4 py-3 flex items-center justify-between" style={{ borderTop: '1px solid #f3f4f6' }}>
                        <button type="button" onClick={handleDiscard}
                          className="flex items-center gap-1 rounded-md px-3 py-1.5 text-[11px] transition-colors hover:bg-gray-100" style={{ color: '#6b7280' }}>
                          <Trash2 size={11} /> Discard
                        </button>
                        <button type="button" onClick={handleSave} disabled={isWorking}
                          className="flex items-center gap-1.5 rounded-md px-4 py-1.5 text-[11px] font-semibold text-white transition-colors disabled:opacity-50"
                          style={{ background: '#111827' }}>
                          {phase === 'saving' ? <Loader2 size={12} className="animate-spin" /> : <ArrowRight size={12} />}
                          {phase === 'saving' ? 'Checking impact…' : 'Save changes'}
                        </button>
                      </div>
                    </>
                  )}

                  {/* Phase: confirm — show impact + confirm/back */}
                  {(phase === 'confirm' || phase === 'committing') && (
                    <>
                      <div className="px-4 pt-3 pb-2">
                        <span className="text-[11px] font-semibold" style={{ color: '#111827' }}>Impact of {actions.length} {actions.length === 1 ? 'change' : 'changes'}</span>
                      </div>

                      <div className="max-h-[320px] overflow-y-auto px-4 pb-3">
                        {impact && <ImpactPanel deltas={impact} changesetId={changesetId ?? ''} />}
                        {!impact && (
                          <p className="text-[11px] py-4 text-center" style={{ color: '#9ca3af' }}>No impact data available</p>
                        )}
                      </div>

                      {error && (
                        <div className="mx-4 mb-3 flex items-start gap-2 rounded-lg px-3 py-2" style={{ background: '#fef2f2', border: '1px solid #fca5a5' }}>
                          <AlertTriangle size={11} className="shrink-0 mt-0.5" style={{ color: '#dc2626' }} />
                          <span className="text-[11px]" style={{ color: '#991b1b' }}>{error}</span>
                        </div>
                      )}

                      <div className="px-4 py-3 flex items-center justify-between" style={{ borderTop: '1px solid #f3f4f6' }}>
                        <button type="button" onClick={reset} disabled={isWorking}
                          className="flex items-center gap-1 rounded-md px-3 py-1.5 text-[11px] transition-colors hover:bg-gray-100 disabled:opacity-50" style={{ color: '#6b7280' }}>
                          Back to changes
                        </button>
                        <button type="button" onClick={handleConfirm} disabled={isWorking}
                          className="flex items-center gap-1.5 rounded-md px-4 py-1.5 text-[11px] font-semibold text-white transition-colors disabled:opacity-50"
                          style={{ background: '#059669' }}>
                          {phase === 'committing' ? <Loader2 size={12} className="animate-spin" /> : <Check size={12} />}
                          {phase === 'committing' ? 'Committing…' : 'Confirm & commit'}
                        </button>
                      </div>
                    </>
                  )}
                </div>
              </>
            )}
          </div>
        )}

        <div className="relative flex items-center">
          <Search className="absolute left-2.5 w-3.5 h-3.5 pointer-events-none" style={{ color: '#9ca3af' }} />
          <input
            type="text"
            value={query}
            onChange={e => setQuery(e.target.value)}
            placeholder="Search entities..."
            className="pl-8 pr-8 py-1.5 text-sm rounded-md w-52 focus:outline-none focus:ring-2"
            style={{ background: '#f9fafb', border: '1px solid #e5e7eb', color: '#111827' }}
          />
          {query && (
            <button onClick={() => setQuery('')} className="absolute right-2.5 hover:text-foreground" style={{ color: '#9ca3af' }} aria-label="Clear search">
              <X className="w-3 h-3" />
            </button>
          )}
        </div>
      </div>
    </header>
  )
}
