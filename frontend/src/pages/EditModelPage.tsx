import { useState, useCallback, useEffect } from 'react'
import { useRequireModel } from '@/lib/model-context'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Send, Loader2, Download, Check, AlertTriangle, RotateCcw } from 'lucide-react'
import { ActionForm } from '@/components/changeset/ActionForm'
import { ActionList } from '@/components/changeset/ActionList'
import { ImpactPanel } from '@/components/changeset/ImpactPanel'
import { api } from '@/lib/api'
import type { ChangeAction, ImpactDelta, CommitResponse } from '@/lib/api'
import { getAndClearPendingAction } from '@/components/changeset/QuickAction'

type EditPhase = 'editing' | 'previewing' | 'committed'

export function EditModelPage() {
  const { modelId, parseResult, setModel, isHydrating } = useRequireModel()

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

  useEffect(() => {
    const pending = getAndClearPendingAction()
    if (pending) {
      setActions([pending])
    }
  }, [])

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
      } else {
        setError('Commit rejected — validation errors found. See details below.')
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Commit failed')
    } finally {
      setCommitting(false)
    }
  }, [changesetId, modelId, parseResult, setModel])

  const handleExport = useCallback(async () => {
    if (!modelId) return
    setExporting(true)
    try {
      const yaml = await api.exportModel(modelId)
      const blob = new Blob([yaml], { type: 'application/x-yaml' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `${parseResult?.system_name?.replace(/\s+/g, '-').toLowerCase() ?? 'model'}.unm.yaml`
      a.click()
      URL.revokeObjectURL(url)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Export failed')
    } finally {
      setExporting(false)
    }
  }, [modelId, parseResult])

  if (isHydrating || !modelId || !parseResult) return null

  return (
    <div className="max-w-7xl space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight" style={{ color: '#111827' }}>Edit Model</h1>
          <p className="text-sm mt-1" style={{ color: '#6b7280' }}>
            {parseResult.system_name} &mdash; Make structural changes, preview impact, and commit
          </p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={handleExport} disabled={exporting} className="gap-2">
            {exporting ? <Loader2 size={14} className="animate-spin" /> : <Download size={14} />}
            {phase === 'committed' ? 'Export Updated Model as YAML' : 'Export Current Model as YAML'}
          </Button>
        </div>
      </div>

      {/* Phase indicator */}
      <div className="flex gap-1">
        {(['editing', 'previewing', 'committed'] as EditPhase[]).map((p, i) => (
          <div key={p} className="flex items-center gap-1">
            {i > 0 && <div className="w-8 h-px" style={{ background: '#d1d5db' }} />}
            <div className="flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-medium"
              style={{
                background: phase === p ? '#111827' : '#f3f4f6',
                color: phase === p ? '#ffffff' : '#9ca3af',
              }}>
              <span>{i + 1}.</span>
              {p === 'editing' && 'Build Changes'}
              {p === 'previewing' && 'Preview Impact'}
              {p === 'committed' && 'Committed'}
            </div>
          </div>
        ))}
      </div>

      {error && (
        <div className="rounded-lg p-3 flex items-start gap-2" style={{ border: '1px solid #fca5a5', background: '#fef2f2' }}>
          <AlertTriangle size={14} className="flex-shrink-0 mt-0.5" style={{ color: '#b91c1c' }} />
          <p className="text-sm" style={{ color: '#b91c1c' }}>{error}</p>
        </div>
      )}

      {/* Commit validation errors */}
      {commitResult && !commitResult.validation.valid && commitResult.validation.errors && (
        <div className="rounded-lg p-4 space-y-2" style={{ border: '1px solid #fca5a5', background: '#fef2f2' }}>
          <p className="text-sm font-medium" style={{ color: '#b91c1c' }}>Validation errors prevent commit:</p>
          {commitResult.validation.errors.map((e, i) => (
            <p key={i} className="text-xs" style={{ color: '#991b1b' }}>{e}</p>
          ))}
        </div>
      )}

      {/* Success banner */}
      {phase === 'committed' && commitResult?.validation.valid && (
        <div className="rounded-lg p-4 flex items-start gap-3" style={{ border: '1px solid #86efac', background: '#f0fdf4' }}>
          <Check size={18} className="flex-shrink-0 mt-0.5" style={{ color: '#15803d' }} />
          <div>
            <p className="text-sm font-medium" style={{ color: '#15803d' }}>Changes committed successfully</p>
            <p className="text-xs mt-1" style={{ color: '#166534' }}>
              The model has been updated. You can export the YAML to save it to your git repository.
            </p>
            {commitResult.validation.warnings && commitResult.validation.warnings.length > 0 && (
              <div className="mt-2 space-y-1">
                <p className="text-xs font-medium" style={{ color: '#854d0e' }}>Warnings:</p>
                {commitResult.validation.warnings.map((w, i) => (
                  <p key={i} className="text-xs" style={{ color: '#854d0e' }}>{w}</p>
                ))}
              </div>
            )}
          </div>
        </div>
      )}

      <div className="grid grid-cols-3 gap-6">
        {/* Column 1: Add Action */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm">Add Action</CardTitle>
          </CardHeader>
          <CardContent>
            <ActionForm onAdd={handleAddAction} />
          </CardContent>
        </Card>

        {/* Column 2: Changeset */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm">Pending Changes</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <ActionList actions={actions} onRemove={handleRemoveAction} onClear={handleClear} />
            {actions.length > 0 && (
              <>
                <div>
                  <label className="text-xs font-medium block mb-1" style={{ color: '#6b7280' }}>Description (optional)</label>
                  <input className="w-full rounded-md border px-3 py-2 text-sm"
                    style={{ borderColor: '#d1d5db', background: '#ffffff', color: '#111827' }}
                    placeholder="Describe these changes..."
                    value={description} onChange={e => setDescription(e.target.value)} />
                </div>
                <Button className="w-full gap-2" disabled={submitting || actions.length === 0}
                  onClick={handlePreview}>
                  {submitting ? <Loader2 size={14} className="animate-spin" /> : <Send size={14} />}
                  Preview Impact
                </Button>
              </>
            )}
          </CardContent>
        </Card>

        {/* Column 3: Impact & Commit */}
        <Card>
          <CardHeader className="pb-3">
            <div className="flex items-center justify-between">
              <CardTitle className="text-sm">Impact & Commit</CardTitle>
              {phase === 'committed' && (
                <Button variant="ghost" size="sm" onClick={handleClear} className="h-6 px-2 text-xs gap-1">
                  <RotateCcw size={12} /> New Edit
                </Button>
              )}
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            {phase === 'editing' && !impact && (
              <div className="text-center py-8">
                <p className="text-sm" style={{ color: '#9ca3af' }}>Add actions and click "Preview Impact"</p>
              </div>
            )}

            {submitting && (
              <div className="flex items-center justify-center py-8">
                <Loader2 size={20} className="animate-spin" style={{ color: '#9ca3af' }} />
              </div>
            )}

            {impact && changesetId && (
              <>
                <ImpactPanel deltas={impact} changesetId={changesetId} />

                {phase === 'previewing' && (
                  <Button className="w-full gap-2" disabled={committing}
                    onClick={handleCommit}
                    style={{ background: '#15803d' }}>
                    {committing ? <Loader2 size={14} className="animate-spin" /> : <Check size={14} />}
                    Commit Changes
                  </Button>
                )}
              </>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
