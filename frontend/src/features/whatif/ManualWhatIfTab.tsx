import { useState } from 'react'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Send, Loader2 } from 'lucide-react'
import { ActionForm } from '@/components/changeset/ActionForm'
import { ActionList } from '@/components/changeset/ActionList'
import { ImpactPanel } from '@/components/changeset/ImpactPanel'
import { changesetsApi } from '@/services/api'
import type { ChangeAction, ImpactDelta } from '@/types/changeset'

export function ManualWhatIfTab({ modelId }: { modelId: string }) {
  const [actions, setActions] = useState<ChangeAction[]>([])
  const [description, setDescription] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [changesetId, setChangesetId] = useState<string | null>(null)
  const [impact, setImpact] = useState<ImpactDelta[] | null>(null)
  const [loadingImpact, setLoadingImpact] = useState(false)

  const handleAddAction = (action: ChangeAction) => { setActions(prev => [...prev, action]); setError(null) }
  const handleRemoveAction = (index: number) => setActions(prev => prev.filter((_, i) => i !== index))
  const handleClear = () => { setActions([]); setChangesetId(null); setImpact(null); setError(null) }

  const handleSubmit = async () => {
    if (actions.length === 0) return
    setSubmitting(true); setError(null)
    try {
      const csId = `cs-${Date.now()}`
      const result = await changesetsApi.createChangeset(modelId, { id: csId, description: description || 'What-if scenario', actions })
      setChangesetId(result.id)
      setLoadingImpact(true)
      const impactResult = await changesetsApi.getImpact(modelId, result.id)
      setImpact(impactResult.deltas)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Submission failed')
    } finally {
      setSubmitting(false); setLoadingImpact(false)
    }
  }

  return (
    <div className="grid grid-cols-3 gap-6">
      <Card>
        <CardHeader className="pb-3"><CardTitle className="text-sm">Add Action</CardTitle></CardHeader>
        <CardContent><ActionForm onAdd={handleAddAction} /></CardContent>
      </Card>
      <Card>
        <CardHeader className="pb-3"><CardTitle className="text-sm">Changeset Actions</CardTitle></CardHeader>
        <CardContent className="space-y-4">
          <ActionList actions={actions} onRemove={handleRemoveAction} onClear={handleClear} />
          {actions.length > 0 && (
            <>
              <div>
                <label className="text-xs font-medium block mb-1 text-muted-foreground">Description (optional)</label>
                <input className="w-full rounded-md border border-border px-3 py-2 text-sm bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
                  placeholder="Describe this what-if scenario..." value={description} onChange={e => setDescription(e.target.value)} />
              </div>
              <Button className="w-full gap-2" disabled={submitting || actions.length === 0} onClick={handleSubmit}>
                {submitting ? <><Loader2 className="w-3.5 h-3.5 animate-spin" />Submitting...</> : <><Send className="w-3.5 h-3.5" />Submit & Analyze</>}
              </Button>
            </>
          )}
          {error && <div className="rounded-lg p-3 bg-red-50 border border-red-200"><p className="text-xs text-red-700">{error}</p></div>}
        </CardContent>
      </Card>
      <Card>
        <CardHeader className="pb-3"><CardTitle className="text-sm">Impact Analysis</CardTitle></CardHeader>
        <CardContent>
          {loadingImpact ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="w-5 h-5 animate-spin text-muted-foreground" />
            </div>
          ) : impact && changesetId ? (
            <ImpactPanel deltas={impact} changesetId={changesetId} />
          ) : (
            <div className="text-center py-8">
              <p className="text-sm text-muted-foreground">Submit a changeset to see impact</p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
