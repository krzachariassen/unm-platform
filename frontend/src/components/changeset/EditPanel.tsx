import { useState, useCallback } from 'react'
import { X, Send, Loader2, Download, Check, AlertTriangle, RotateCcw, ChevronDown, ChevronUp } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { ActionForm, useModelEntities } from './ActionForm'
import { ActionList } from './ActionList'
import { ImpactPanel } from './ImpactPanel'
import { api } from '@/lib/api'
import { useModel } from '@/lib/model-context'
import type { ChangeAction, ImpactDelta, CommitResponse } from '@/lib/api'

type EditPhase = 'editing' | 'previewing' | 'committed'

interface EditPanelProps {
  onClose: () => void
  onCommitted?: () => void
}

export function EditPanel({ onClose, onCommitted }: EditPanelProps) {
  const { modelId, parseResult, setModel } = useModel()
  const entities = useModelEntities()

  const [actions, setActions] = useState<ChangeAction[]>([])
  const [description, setDescription] = useState('')
  const [phase, setPhase] = useState<EditPhase>('editing')

  const [submitting, setSubmitting] = useState(false)
  const [committing, setCommitting] = useState(false)
  const [exporting, setExporting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const [changesetId, setChangesetId] = useState<string | null>(null)
  const [impact, setImpact] = useState<ImpactDelta[] | null>(null)
  const [commitResult, setCommitResult] = useState<CommitResponse | null>(null)

  const [showForm, setShowForm] = useState(true)
  const [showImpact, setShowImpact] = useState(true)

  const handleAddAction = (action: ChangeAction) => {
    setActions(prev => [...prev, action])
    setError(null)
    setPhase('editing')
    setChangesetId(null)
    setImpact(null)
    setCommitResult(null)
  }

  const handleRemoveAction = (index: number) => {
    setActions(prev => prev.filter((_, i) => i !== index))
    setChangesetId(null)
    setImpact(null)
    setPhase('editing')
  }

  const handleClear = () => {
    setActions([])
    setChangesetId(null)
    setImpact(null)
    setCommitResult(null)
    setError(null)
    setPhase('editing')
    setDescription('')
  }

  const handlePreview = useCallback(async () => {
    if (actions.length === 0 || !modelId) return
    setSubmitting(true)
    setError(null)
    try {
      const csId = `edit-${Date.now()}`
      const result = await api.createChangeset(modelId, {
        id: csId,
        description: description || 'Model edit',
        actions,
      })
      setChangesetId(result.id)
      const impactResult = await api.getChangesetImpact(modelId, result.id)
      setImpact(impactResult.deltas)
      setPhase('previewing')
      setShowImpact(true)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Preview failed')
    } finally {
      setSubmitting(false)
    }
  }, [actions, modelId, description])

  const handleCommit = useCallback(async () => {
    if (!changesetId || !modelId) return
    setCommitting(true)
    setError(null)
    try {
      const result = await api.commitChangeset(modelId, changesetId)
      setCommitResult(result)
      if (result.validation.valid) {
        setPhase('committed')
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
        onCommitted?.()
      } else {
        setError('Commit rejected — validation errors found.')
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Commit failed')
    } finally {
      setCommitting(false)
    }
  }, [changesetId, modelId, parseResult, setModel, onCommitted])

  const handleExport = useCallback(async (format: 'yaml' | 'dsl' = 'dsl') => {
    if (!modelId) return
    setExporting(true)
    try {
      const content = await api.exportModel(modelId, format)
      const ext = format === 'dsl' ? '.unm' : '.unm.yaml'
      const mimeType = format === 'dsl' ? 'text/plain' : 'application/x-yaml'
      const blob = new Blob([content], { type: mimeType })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `${parseResult?.system_name?.replace(/\s+/g, '-').toLowerCase() ?? 'model'}${ext}`
      a.click()
      URL.revokeObjectURL(url)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Export failed')
    } finally {
      setExporting(false)
    }
  }, [modelId, parseResult])

  if (!modelId) return null

  return (
    <div className="flex flex-col h-full" style={{ background: '#ffffff' }}>
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 flex-shrink-0"
        style={{ borderBottom: '1px solid #e5e7eb' }}>
        <div className="flex items-center gap-2">
          <h3 className="text-sm font-semibold" style={{ color: '#111827' }}>Edit Model</h3>
          {/* Phase pills */}
          <div className="flex gap-0.5">
            {(['editing', 'previewing', 'committed'] as EditPhase[]).map(p => (
              <div key={p} className="px-1.5 py-0.5 rounded text-[10px] font-medium"
                style={{
                  background: phase === p ? '#111827' : '#f3f4f6',
                  color: phase === p ? '#ffffff' : '#9ca3af',
                }}>
                {p === 'editing' ? 'Edit' : p === 'previewing' ? 'Preview' : 'Done'}
              </div>
            ))}
          </div>
        </div>
        <div className="flex items-center gap-1">
          <button onClick={() => handleExport('dsl')} disabled={exporting}
            className="p-1.5 rounded hover:bg-gray-100 disabled:opacity-40" title="Export .unm">
            {exporting ? <Loader2 size={14} className="animate-spin" style={{ color: '#9ca3af' }} /> :
              <Download size={14} style={{ color: '#6b7280' }} />}
          </button>
          <button onClick={onClose} className="p-1.5 rounded hover:bg-gray-100" title="Close editor">
            <X size={14} style={{ color: '#6b7280' }} />
          </button>
        </div>
      </div>

      {/* Scrollable content */}
      <div className="flex-1 overflow-y-auto">
        {/* Error */}
        {error && (
          <div className="mx-3 mt-3 rounded-lg p-2.5 flex items-start gap-2" style={{ border: '1px solid #fca5a5', background: '#fef2f2' }}>
            <AlertTriangle size={13} className="flex-shrink-0 mt-0.5" style={{ color: '#b91c1c' }} />
            <p className="text-xs" style={{ color: '#b91c1c' }}>{error}</p>
          </div>
        )}

        {/* Commit validation errors */}
        {commitResult && !commitResult.validation.valid && commitResult.validation.errors && (
          <div className="mx-3 mt-3 rounded-lg p-2.5 space-y-1" style={{ border: '1px solid #fca5a5', background: '#fef2f2' }}>
            <p className="text-xs font-medium" style={{ color: '#b91c1c' }}>Validation errors:</p>
            {commitResult.validation.errors.map((e, i) => (
              <p key={i} className="text-[11px]" style={{ color: '#991b1b' }}>{e}</p>
            ))}
          </div>
        )}

        {/* Success banner */}
        {phase === 'committed' && commitResult?.validation.valid && (
          <div className="mx-3 mt-3 rounded-lg p-3 flex items-start gap-2" style={{ border: '1px solid #86efac', background: '#f0fdf4' }}>
            <Check size={14} className="flex-shrink-0 mt-0.5" style={{ color: '#15803d' }} />
            <div className="flex-1">
              <p className="text-xs font-medium" style={{ color: '#15803d' }}>Changes committed</p>
              <p className="text-[11px] mt-0.5" style={{ color: '#166534' }}>
                Model updated. The map will refresh when you close this panel.
              </p>
              {commitResult.validation.warnings && commitResult.validation.warnings.length > 0 && (
                <div className="mt-1.5 space-y-0.5">
                  {commitResult.validation.warnings.map((w, i) => (
                    <p key={i} className="text-[11px]" style={{ color: '#854d0e' }}>{w}</p>
                  ))}
                </div>
              )}
              <Button variant="outline" size="sm" onClick={handleClear} className="mt-2 h-7 text-xs gap-1">
                <RotateCcw size={11} /> New Edit
              </Button>
            </div>
          </div>
        )}

        {/* Section: Add Action */}
        <div className="px-3 pt-3">
          <button onClick={() => setShowForm(!showForm)}
            className="flex items-center justify-between w-full text-xs font-semibold mb-2 group"
            style={{ color: '#374151' }}>
            <span>Add Action</span>
            {showForm ? <ChevronUp size={13} style={{ color: '#9ca3af' }} /> : <ChevronDown size={13} style={{ color: '#9ca3af' }} />}
          </button>
          {showForm && (
            <ActionForm onAdd={handleAddAction} entities={entities} compact />
          )}
        </div>

        {/* Section: Pending Changes */}
        <div className="px-3 pt-4" style={{ borderTop: actions.length > 0 ? undefined : undefined }}>
          <div className="text-xs font-semibold mb-2" style={{ color: '#374151' }}>
            Pending Changes
            {actions.length > 0 && <span className="ml-1.5 font-normal" style={{ color: '#9ca3af' }}>({actions.length})</span>}
          </div>
          <ActionList actions={actions} onRemove={handleRemoveAction} onClear={handleClear} />

          {actions.length > 0 && (
            <div className="mt-3 space-y-2">
              <input className="w-full rounded-md border px-2.5 py-1.5 text-xs"
                style={{ borderColor: '#d1d5db', background: '#ffffff', color: '#111827' }}
                placeholder="Description (optional)"
                value={description} onChange={e => setDescription(e.target.value)} />
              <Button size="sm" className="w-full gap-1.5 h-8 text-xs" disabled={submitting || actions.length === 0}
                onClick={handlePreview}>
                {submitting ? <Loader2 size={13} className="animate-spin" /> : <Send size={13} />}
                Preview Impact
              </Button>
            </div>
          )}
        </div>

        {/* Section: Impact */}
        {(impact || submitting) && (
          <div className="px-3 pt-4 pb-4">
            <button onClick={() => setShowImpact(!showImpact)}
              className="flex items-center justify-between w-full text-xs font-semibold mb-2"
              style={{ color: '#374151' }}>
              <span>Impact</span>
              {showImpact ? <ChevronUp size={13} style={{ color: '#9ca3af' }} /> : <ChevronDown size={13} style={{ color: '#9ca3af' }} />}
            </button>
            {showImpact && (
              <>
                {submitting && (
                  <div className="flex items-center justify-center py-6">
                    <Loader2 size={16} className="animate-spin" style={{ color: '#9ca3af' }} />
                  </div>
                )}

                {impact && changesetId && !submitting && (
                  <div className="space-y-3">
                    <ImpactPanel deltas={impact} changesetId={changesetId} />

                    {phase === 'previewing' && (
                      <Button size="sm" className="w-full gap-1.5 h-8 text-xs" disabled={committing}
                        onClick={handleCommit}
                        style={{ background: '#15803d' }}>
                        {committing ? <Loader2 size={13} className="animate-spin" /> : <Check size={13} />}
                        Commit Changes
                      </Button>
                    )}
                  </div>
                )}
              </>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
