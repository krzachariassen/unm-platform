import { X, Download, Loader2, Pencil } from 'lucide-react'
import { useState, useCallback } from 'react'
import { ActionForm, useModelEntities } from './ActionForm'
import { ActionList } from './ActionList'
import { useChangeset } from '@/lib/changeset-context'
import { useModel } from '@/lib/model-context'
import { api } from '@/lib/api'

interface EditPanelProps {
  open: boolean
  onClose: () => void
}

export function EditPanel({ open, onClose }: EditPanelProps) {
  const { modelId, parseResult } = useModel()
  const { isEditMode, actions, addAction, removeAction, clearActions, enterEditMode } = useChangeset()
  const entities = useModelEntities()
  const [exporting, setExporting] = useState(false)
  const [exportError, setExportError] = useState<string | null>(null)

  const handleAddAction = useCallback((action: Parameters<typeof addAction>[0]) => {
    if (!isEditMode) enterEditMode()
    addAction(action)
  }, [isEditMode, enterEditMode, addAction])

  const handleExport = useCallback(async (format: 'yaml' | 'dsl' = 'dsl') => {
    if (!modelId) return
    setExporting(true)
    setExportError(null)
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
      setExportError(err instanceof Error ? err.message : 'Export failed')
    } finally {
      setExporting(false)
    }
  }, [modelId, parseResult])

  if (!open || !modelId) return null

  return (
    <div
      className="flex flex-col h-full"
      style={{ background: '#ffffff' }}
      onClick={e => e.stopPropagation()}
    >
      {/* Header */}
      <div
        className="flex items-center justify-between px-4 py-3 flex-shrink-0"
        style={{ borderBottom: '1px solid #e5e7eb' }}
      >
        <div className="flex items-center gap-2">
          <Pencil size={13} style={{ color: '#6b7280' }} />
          <h3 className="text-sm font-semibold" style={{ color: '#111827' }}>Add Changes</h3>
          {actions.length > 0 && (
            <span
              className="inline-flex items-center px-1.5 py-0.5 rounded-full text-[10px] font-semibold"
              style={{ background: '#dbeafe', color: '#1d4ed8' }}
            >
              {actions.length}
            </span>
          )}
        </div>
        <div className="flex items-center gap-1">
          <button
            onClick={() => handleExport('dsl')}
            disabled={exporting}
            className="p-1.5 rounded hover:bg-gray-100 disabled:opacity-40"
            title="Export .unm"
          >
            {exporting
              ? <Loader2 size={14} className="animate-spin" style={{ color: '#9ca3af' }} />
              : <Download size={14} style={{ color: '#6b7280' }} />
            }
          </button>
          <button
            onClick={onClose}
            className="p-1.5 rounded hover:bg-gray-100"
            title="Close"
          >
            <X size={14} style={{ color: '#6b7280' }} />
          </button>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto px-3 pt-3 pb-4 space-y-4">
        {exportError && (
          <p className="text-xs rounded px-2 py-1.5" style={{ background: '#fef2f2', color: '#b91c1c' }}>
            {exportError}
          </p>
        )}

        <div>
          <div className="text-xs font-semibold mb-2" style={{ color: '#374151' }}>Add Action</div>
          <ActionForm onAdd={handleAddAction} entities={entities} compact />
        </div>

        {actions.length > 0 && (
          <div>
            <div className="text-xs font-semibold mb-2" style={{ color: '#374151' }}>
              Pending Changes
              <span className="ml-1 font-normal" style={{ color: '#9ca3af' }}>({actions.length}) — commit from the bottom bar</span>
            </div>
            <ActionList actions={actions} onRemove={removeAction} onClear={clearActions} />
          </div>
        )}

        {actions.length === 0 && (
          <p className="text-xs text-center py-4" style={{ color: '#9ca3af' }}>
            Add actions above. Use the bar at the bottom to preview impact and commit.
          </p>
        )}
      </div>
    </div>
  )
}
