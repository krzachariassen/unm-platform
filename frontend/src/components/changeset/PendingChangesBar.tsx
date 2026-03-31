import { useState, useCallback } from 'react'
import { Loader2, Send, Check, AlertTriangle, ChevronUp, ChevronDown, X, List } from 'lucide-react'
import { api } from '@/lib/api'
import type { ImpactDelta } from '@/lib/api'
import { useModel } from '@/lib/model-context'
import { useChangeset } from '@/lib/changeset-context'
import { ImpactPanel } from './ImpactPanel'
import { ActionForm } from './ActionForm'

type CommitPhase = 'idle' | 'previewing' | 'previewed' | 'committing' | 'committed'

function actionSummary(a: { type: string; service_name?: string; capability_name?: string; team_name?: string; need_name?: string; from_team_name?: string; to_team_name?: string }): string {
  const parts: string[] = [a.type.replace(/_/g, ' ')]
  if (a.service_name) parts.push(a.service_name)
  else if (a.capability_name) parts.push(a.capability_name)
  else if (a.team_name) parts.push(a.team_name)
  else if (a.need_name) parts.push(a.need_name)
  if (a.from_team_name && a.to_team_name) parts.push(`${a.from_team_name} → ${a.to_team_name}`)
  return parts.join(': ')
}

export function PendingChangesBar() {
  const { modelId, parseResult, setModel } = useModel()
  const { isEditMode, actions, description, exitEditMode, addAction, removeAction, clearActions } = useChangeset()

  const [phase, setPhase] = useState<CommitPhase>('idle')
  const [changesetId, setChangesetId] = useState<string | null>(null)
  const [impact, setImpact] = useState<ImpactDelta[] | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [showList, setShowList] = useState(false)
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [showImpact, setShowImpact] = useState(true)

  const handlePreview = useCallback(async () => {
    if (!modelId || actions.length === 0) return
    setPhase('previewing')
    setError(null)
    setImpact(null)
    setChangesetId(null)
    setShowImpact(true)
    try {
      // Always fresh ID — never reuse a stale changeset
      const csId = `batch-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`
      const cs = await api.createChangeset(modelId, {
        id: csId,
        description: description || 'Batch edit',
        actions,
      })
      const impactResult = await api.getChangesetImpact(modelId, cs.id)
      setChangesetId(cs.id)
      setImpact(impactResult.deltas)
      setPhase('previewed')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Preview failed')
      setPhase('idle')
    }
  }, [modelId, actions, description])

  const handleCommit = useCallback(async () => {
    if (!modelId || !changesetId) return
    setPhase('committing')
    setError(null)
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
        // Give the user a moment to see success, then exit edit mode (triggers map refresh)
        setTimeout(() => {
          exitEditMode()
          setPhase('idle')
          setImpact(null)
          setChangesetId(null)
        }, 1500)
      } else {
        const msgs = result.validation.errors?.length
          ? result.validation.errors.join(' · ')
          : 'Validation failed — no details returned'
        setError(msgs)
        setPhase('previewed')
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Commit failed')
      setPhase('previewed')
    }
  }, [modelId, changesetId, parseResult, setModel, exitEditMode])

  const handleDiscard = useCallback(() => {
    if (actions.length > 0 && !window.confirm(`Discard ${actions.length} pending change${actions.length === 1 ? '' : 's'}?`)) return
    exitEditMode()
    setPhase('idle')
    setImpact(null)
    setChangesetId(null)
    setError(null)
  }, [actions.length, exitEditMode])

  // Reset phase when actions change (user added/removed)
  const handleReset = useCallback(() => {
    setPhase('idle')
    setImpact(null)
    setChangesetId(null)
    setError(null)
  }, [])

  if (!isEditMode) return null

  const isWorking = phase === 'previewing' || phase === 'committing'
  const canPreview = actions.length > 0 && !isWorking && phase !== 'committed'
  const canCommit = phase === 'previewed' && changesetId && !isWorking

  return (
    <div className="fixed bottom-0 left-0 right-0 z-50" style={{ boxShadow: '0 -2px 20px rgba(0,0,0,0.12)' }}>
      {/* Advanced action modal */}
      {showAdvanced && (
        <div
          style={{ position: 'fixed', inset: 0, zIndex: 60, display: 'flex', alignItems: 'center', justifyContent: 'center', background: 'rgba(0,0,0,0.5)' }}
          onClick={() => setShowAdvanced(false)}
        >
          <div
            style={{ background: '#ffffff', borderRadius: 12, padding: 24, width: 420, maxHeight: '80vh', overflowY: 'auto', boxShadow: '0 20px 60px rgba(0,0,0,0.3)' }}
            onClick={e => e.stopPropagation()}
          >
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 16 }}>
              <span style={{ fontWeight: 600, fontSize: 14, color: '#111827' }}>Add Advanced Action</span>
              <button onClick={() => setShowAdvanced(false)} style={{ background: 'none', border: 'none', cursor: 'pointer', color: '#6b7280', fontSize: 18, lineHeight: 1 }}>✕</button>
            </div>
            <ActionForm
              onAdd={(action) => {
                addAction(action)
                setShowAdvanced(false)
              }}
              compact
            />
          </div>
        </div>
      )}
      {/* Impact panel — expands above the bar */}
      {(impact || phase === 'previewing') && showImpact && (
        <div
          className="overflow-y-auto"
          style={{
            background: '#ffffff',
            borderTop: '1px solid #e5e7eb',
            borderBottom: '1px solid #e5e7eb',
            maxHeight: 320,
            padding: '12px 20px',
          }}
        >
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-semibold" style={{ color: '#374151' }}>Impact Preview</span>
            <button onClick={() => setShowImpact(false)} className="p-0.5 rounded hover:bg-gray-100">
              <ChevronDown size={14} style={{ color: '#9ca3af' }} />
            </button>
          </div>
          {phase === 'previewing' && (
            <div className="flex items-center gap-2 py-4 text-sm" style={{ color: '#9ca3af' }}>
              <Loader2 size={14} className="animate-spin" /> Computing impact…
            </div>
          )}
          {impact && <ImpactPanel deltas={impact} changesetId={changesetId ?? ''} />}
        </div>
      )}

      {/* Action list — expands above the bar */}
      {showList && (
        <div
          className="overflow-y-auto"
          style={{
            background: '#fafafa',
            borderTop: '1px solid #e5e7eb',
            maxHeight: 220,
            padding: '10px 20px',
          }}
        >
          <div className="flex items-center justify-between mb-2">
            <span className="text-xs font-semibold" style={{ color: '#374151' }}>Pending Changes ({actions.length})</span>
            <div className="flex items-center gap-2">
              <button
                onClick={() => { clearActions(); handleReset() }}
                className="text-xs px-2 py-0.5 rounded hover:bg-gray-200"
                style={{ color: '#b91c1c' }}
              >
                Clear all
              </button>
              <button onClick={() => setShowList(false)} className="p-0.5 rounded hover:bg-gray-100">
                <ChevronDown size={14} style={{ color: '#9ca3af' }} />
              </button>
            </div>
          </div>
          <div className="space-y-1">
            {actions.map((a, i) => (
              <div
                key={i}
                className="flex items-center justify-between rounded px-2.5 py-1.5 text-xs"
                style={{ background: '#ffffff', border: '1px solid #e5e7eb' }}
              >
                <span style={{ color: '#374151' }}>{actionSummary(a)}</span>
                <button
                  onClick={() => { removeAction(i); handleReset() }}
                  className="ml-2 p-0.5 rounded hover:bg-gray-100 flex-shrink-0"
                >
                  <X size={11} style={{ color: '#9ca3af' }} />
                </button>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Main bar */}
      <div
        className="flex items-center gap-3 px-5"
        style={{
          height: 52,
          background: '#111827',
          color: '#ffffff',
        }}
      >
        {/* Edit mode indicator */}
        <div className="flex items-center gap-1.5 flex-shrink-0">
          <div className="w-2 h-2 rounded-full animate-pulse" style={{ background: '#22c55e' }} />
          <span className="text-xs font-medium" style={{ color: '#d1d5db' }}>Edit Mode</span>
        </div>

        <div className="w-px h-4 flex-shrink-0" style={{ background: '#374151' }} />

        {/* Count + list toggle */}
        <button
          onClick={() => setShowList(s => !s)}
          className="flex items-center gap-1.5 text-xs rounded px-2 py-1 transition-colors"
          style={{
            color: actions.length > 0 ? '#ffffff' : '#6b7280',
            background: showList ? '#374151' : 'transparent',
          }}
          disabled={actions.length === 0}
        >
          <List size={13} />
          <span className="font-semibold">{actions.length}</span>
          <span style={{ color: '#9ca3af' }}>pending {actions.length === 1 ? 'change' : 'changes'}</span>
          {actions.length > 0 && (showList ? <ChevronDown size={11} style={{ color: '#9ca3af' }} /> : <ChevronUp size={11} style={{ color: '#9ca3af' }} />)}
        </button>

        {/* Impact toggle (when available) */}
        {impact && !showImpact && (
          <button
            onClick={() => setShowImpact(true)}
            className="flex items-center gap-1 text-xs rounded px-2 py-1 transition-colors"
            style={{ color: '#93c5fd' }}
          >
            <ChevronUp size={11} /> Impact
          </button>
        )}

        {/* Error */}
        {error && (
          <div className="flex items-start gap-1.5 flex-1 min-w-0">
            <AlertTriangle size={13} className="flex-shrink-0 mt-0.5" style={{ color: '#f87171' }} />
            <span className="text-xs break-words" style={{ color: '#f87171' }}>{error}</span>
          </div>
        )}

        {/* Success */}
        {phase === 'committed' && (
          <div className="flex items-center gap-1.5 flex-1">
            <Check size={13} style={{ color: '#4ade80' }} />
            <span className="text-xs" style={{ color: '#4ade80' }}>Changes committed!</span>
          </div>
        )}

        {!error && phase !== 'committed' && <div className="flex-1" />}

        {/* Action buttons */}
        <div className="flex items-center gap-2 flex-shrink-0">
          <button
            onClick={() => setShowAdvanced(true)}
            className="px-2 py-1.5 rounded text-xs transition-colors"
            style={{ color: '#9ca3af', background: 'transparent', border: '1px solid #374151' }}
            title="Add advanced action (split team, merge teams, etc.)"
            onMouseEnter={e => { e.currentTarget.style.color = '#ffffff'; e.currentTarget.style.borderColor = '#6b7280' }}
            onMouseLeave={e => { e.currentTarget.style.color = '#9ca3af'; e.currentTarget.style.borderColor = '#374151' }}
          >
            + More
          </button>
          <button
            onClick={handleDiscard}
            className="px-3 py-1.5 rounded text-xs transition-colors"
            style={{ color: '#9ca3af', background: 'transparent' }}
            onMouseEnter={e => { e.currentTarget.style.color = '#f87171'; e.currentTarget.style.background = '#1f2937' }}
            onMouseLeave={e => { e.currentTarget.style.color = '#9ca3af'; e.currentTarget.style.background = 'transparent' }}
          >
            Discard
          </button>

          <button
            onClick={handlePreview}
            disabled={!canPreview}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded text-xs font-medium transition-colors disabled:opacity-40"
            style={{
              background: canPreview ? '#1d4ed8' : '#374151',
              color: '#ffffff',
            }}
          >
            {phase === 'previewing' ? <Loader2 size={12} className="animate-spin" /> : <Send size={12} />}
            Preview Impact
          </button>

          <button
            onClick={handleCommit}
            disabled={!canCommit}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded text-xs font-medium transition-colors disabled:opacity-40"
            style={{
              background: canCommit ? '#15803d' : '#374151',
              color: '#ffffff',
            }}
          >
            {phase === 'committing' ? <Loader2 size={12} className="animate-spin" /> : <Check size={12} />}
            Commit
          </button>
        </div>
      </div>
    </div>
  )
}
