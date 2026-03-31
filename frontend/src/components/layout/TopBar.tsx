import { useState, useCallback } from 'react'
import { Search, X, ChevronDown, ChevronUp, Send, Check, Loader2, AlertTriangle, Trash2 } from 'lucide-react'
import { useModel } from '@/lib/model-context'
import { useSearch } from '@/lib/search-context'
import { useChangeset } from '@/lib/changeset-context'
import { api } from '@/lib/api'
import type { ImpactDelta } from '@/types/changeset'
import { ImpactPanel } from '@/components/changeset/ImpactPanel'
import { cn } from '@/lib/utils'

type Phase = 'idle' | 'previewing' | 'previewed' | 'committing' | 'committed'

function actionLabel(a: { type: string; service_name?: string; capability_name?: string; team_name?: string; need_name?: string; from_team_name?: string; to_team_name?: string }): string {
  const parts: string[] = [a.type.replace(/_/g, ' ')]
  if (a.service_name) parts.push(a.service_name)
  else if (a.capability_name) parts.push(a.capability_name)
  else if (a.team_name) parts.push(a.team_name)
  else if (a.need_name) parts.push(a.need_name)
  if (a.from_team_name && a.to_team_name) parts.push(`${a.from_team_name} → ${a.to_team_name}`)
  return parts.join(': ')
}

export function TopBar() {
  const { modelId, parseResult, setModel } = useModel()
  const { query, setQuery } = useSearch()
  const { actions, description, removeAction, discardAll } = useChangeset()

  const [open, setOpen] = useState(false)
  const [phase, setPhase] = useState<Phase>('idle')
  const [changesetId, setChangesetId] = useState<string | null>(null)
  const [impact, setImpact] = useState<ImpactDelta[] | null>(null)
  const [error, setError] = useState<string | null>(null)

  const handlePreview = useCallback(async () => {
    if (!modelId || actions.length === 0) return
    setPhase('previewing'); setError(null); setImpact(null)
    try {
      const csId = `batch-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`
      const cs = await api.createChangeset(modelId, { id: csId, description: description || 'Batch edit', actions })
      const result = await api.getChangesetImpact(modelId, cs.id)
      setChangesetId(cs.id); setImpact(result.deltas); setPhase('previewed')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Preview failed'); setPhase('idle')
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
        setPhase('committed')
        setTimeout(() => {
          discardAll(); setPhase('idle'); setImpact(null); setChangesetId(null); setOpen(false)
        }, 1200)
      } else {
        setError(result.validation.errors?.join('; ') || 'Validation failed')
        setPhase('previewed')
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Commit failed'); setPhase('previewed')
    }
  }, [modelId, changesetId, parseResult, setModel, discardAll])

  const handleDiscard = useCallback(() => {
    if (actions.length > 0 && !window.confirm(`Discard ${actions.length} pending change${actions.length === 1 ? '' : 's'}?`)) return
    discardAll(); setPhase('idle'); setImpact(null); setChangesetId(null); setOpen(false); setError(null)
  }, [actions.length, discardAll])

  const isWorking = phase === 'previewing' || phase === 'committing'

  return (
    <header className="flex items-center h-14 px-5 gap-3 bg-white border-b border-border shrink-0">
      {parseResult ? (
        <div className="flex items-center gap-2">
          <span className="text-sm font-semibold text-foreground">{parseResult.system_name}</span>
          <span className={cn(
            'inline-flex items-center px-2 py-0.5 rounded text-xs font-medium',
            parseResult.validation.is_valid ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'
          )}>
            {parseResult.validation.is_valid ? 'Valid' : 'Invalid'}
          </span>
          {parseResult.validation.warnings.length > 0 && (
            <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-amber-100 text-amber-700">
              {parseResult.validation.warnings.length} {parseResult.validation.warnings.length === 1 ? 'warning' : 'warnings'}
            </span>
          )}
        </div>
      ) : (
        <span className="text-sm text-muted-foreground">No model loaded</span>
      )}

      <div className="flex items-center gap-3 ml-auto">
        {/* Pending changes indicator */}
        {actions.length > 0 && (
          <div className="relative">
            <button
              type="button"
              onClick={() => setOpen(o => !o)}
              className={cn(
                'flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-colors border',
                phase === 'committed'
                  ? 'bg-green-50 text-green-700 border-green-200'
                  : 'bg-primary text-primary-foreground border-primary'
              )}
            >
              {phase === 'committed' ? (
                <><Check size={12} /> Saved</>
              ) : (
                <>
                  <span className="inline-flex items-center justify-center w-4 h-4 rounded-full text-[9px] font-bold bg-white/20">
                    {actions.length}
                  </span>
                  {actions.length === 1 ? 'change' : 'changes'}
                  {open ? <ChevronUp size={12} /> : <ChevronDown size={12} />}
                </>
              )}
            </button>

            {/* Dropdown */}
            {open && phase !== 'committed' && (
              <>
                <div className="fixed inset-0 z-40" onClick={() => setOpen(false)} />
                <div className="absolute right-0 top-full mt-1 z-50 w-[380px] rounded-lg border shadow-xl" style={{ background: '#ffffff', borderColor: '#e5e7eb' }}>
                  {/* Change list */}
                  <div className="max-h-[240px] overflow-y-auto p-3">
                    <div className="flex items-center justify-between mb-2">
                      <span className="text-xs font-semibold text-foreground">Pending Changes</span>
                      <button type="button" onClick={handleDiscard} className="text-[10px] text-destructive hover:underline">Discard all</button>
                    </div>
                    <div className="space-y-1">
                      {actions.map((a, i) => (
                        <div key={i} className="flex items-center justify-between rounded border border-border bg-muted/50 px-2.5 py-1.5 text-xs">
                          <span className="text-foreground truncate">{actionLabel(a)}</span>
                          <button type="button" onClick={() => { removeAction(i); setPhase('idle'); setImpact(null); setChangesetId(null) }}
                            className="ml-2 shrink-0 rounded p-0.5 text-muted-foreground hover:text-destructive">
                            <X size={10} />
                          </button>
                        </div>
                      ))}
                    </div>
                  </div>

                  {/* Impact preview */}
                  {(impact || phase === 'previewing') && (
                    <div className="border-t border-border p-3 max-h-[200px] overflow-y-auto">
                      <span className="text-xs font-semibold text-foreground mb-2 block">Impact Preview</span>
                      {phase === 'previewing' && (
                        <div className="flex items-center gap-2 py-2 text-xs text-muted-foreground"><Loader2 size={12} className="animate-spin" /> Computing…</div>
                      )}
                      {impact && <ImpactPanel deltas={impact} changesetId={changesetId ?? ''} />}
                    </div>
                  )}

                  {/* Error */}
                  {error && (
                    <div className="border-t border-destructive/30 px-3 py-2 flex items-start gap-2">
                      <AlertTriangle size={12} className="shrink-0 text-destructive mt-0.5" />
                      <span className="text-xs text-destructive">{error}</span>
                    </div>
                  )}

                  {/* Actions */}
                  <div className="border-t border-border px-3 py-2 flex items-center justify-end gap-1.5">
                    <button type="button" onClick={handleDiscard}
                      className="flex items-center gap-1 rounded px-2.5 py-1 text-[10px] text-muted-foreground hover:bg-muted hover:text-destructive transition-colors">
                      <Trash2 size={10} /> Discard
                    </button>
                    <button type="button" onClick={handlePreview} disabled={isWorking}
                      className={cn('flex items-center gap-1 rounded px-2.5 py-1 text-[10px] font-medium transition-colors disabled:opacity-30',
                        'bg-muted text-foreground hover:bg-muted/80')}>
                      {phase === 'previewing' ? <Loader2 size={10} className="animate-spin" /> : <Send size={10} />}
                      Preview
                    </button>
                    <button type="button" onClick={handleCommit}
                      disabled={!(phase === 'previewed' && changesetId && !isWorking)}
                      className={cn('flex items-center gap-1 rounded px-2.5 py-1 text-[10px] font-medium text-white transition-colors disabled:opacity-30',
                        phase === 'previewed' && changesetId ? 'bg-emerald-600 hover:bg-emerald-700' : 'bg-muted text-muted-foreground')}>
                      {phase === 'committing' ? <Loader2 size={10} className="animate-spin" /> : <Check size={10} />}
                      Commit
                    </button>
                  </div>
                </div>
              </>
            )}
          </div>
        )}

        {/* Search */}
        <div className="relative flex items-center">
          <Search className="absolute left-2.5 w-3.5 h-3.5 text-muted-foreground pointer-events-none" />
          <input
            type="text"
            value={query}
            onChange={e => setQuery(e.target.value)}
            placeholder="Search entities..."
            className="pl-8 pr-8 py-1.5 text-sm rounded-md w-52 bg-gray-50 border border-gray-200 text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
          />
          {query && (
            <button onClick={() => setQuery('')} className="absolute right-2.5 text-muted-foreground hover:text-foreground" aria-label="Clear search">
              <X className="w-3 h-3" />
            </button>
          )}
        </div>
      </div>
    </header>
  )
}
