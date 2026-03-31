import { useState, useRef, useEffect } from 'react'
import { ZoomIn, ZoomOut, Maximize2, Plus, X, Info } from 'lucide-react'
import { useReactFlow } from '@xyflow/react'
import { VIS_ORDER, VIS } from '@/features/unm-map/constants'
import { useChangeset } from '@/lib/changeset-context'
import { ActionForm } from '@/components/changeset/ActionForm'

interface MapToolbarProps {
  highlighted: boolean
  onClearHighlight: () => void
}

export function MapToolbar({ highlighted, onClearHighlight }: MapToolbarProps) {
  const { zoomIn, zoomOut, fitView } = useReactFlow()
  const { addAction } = useChangeset()
  const [showAdd, setShowAdd] = useState(false)
  const [showLegend, setShowLegend] = useState(false)
  const addRef = useRef<HTMLDivElement>(null)
  const legendRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!showAdd) return
    const handler = (e: MouseEvent) => {
      if (addRef.current && !addRef.current.contains(e.target as Node)) setShowAdd(false)
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [showAdd])

  useEffect(() => {
    if (!showLegend) return
    const handler = (e: MouseEvent) => {
      if (legendRef.current && !legendRef.current.contains(e.target as Node)) setShowLegend(false)
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [showLegend])

  return (
    <div className="mb-2 flex items-center gap-2 flex-shrink-0">
      <h2 className="text-base font-semibold text-foreground">UNM Map</h2>

      {highlighted && (
        <button
          onClick={onClearHighlight}
          className="rounded px-2 py-0.5 text-[10px] text-muted-foreground border border-border hover:bg-muted hover:text-foreground transition-colors"
        >
          Clear highlight
        </button>
      )}

      <div className="ml-auto flex items-center gap-0.5">
        {/* Legend toggle */}
        <div ref={legendRef} className="relative">
          <button
            type="button"
            onClick={() => setShowLegend(s => !s)}
            className="rounded p-1.5 hover:bg-muted text-muted-foreground transition-colors"
            title="Legend"
          >
            <Info size={14} />
          </button>
          {showLegend && (
            <div className="absolute right-0 top-full mt-1 z-50 rounded-lg border px-3 py-2 shadow-lg whitespace-nowrap" style={{ background: '#ffffff', borderColor: '#e5e7eb' }}>
              <div className="flex flex-col gap-1.5 text-[11px]">
                <span className="flex items-center gap-1.5" style={{ color: '#2563eb' }}>
                  <span className="w-2 h-2 rounded-full" style={{ background: '#2563eb' }} /> Actor
                </span>
                <span className="flex items-center gap-1.5" style={{ color: '#3b82f6' }}>
                  <span className="w-2 h-2 rounded-sm" style={{ background: '#3b82f6' }} /> Need
                </span>
                {VIS_ORDER.map(v => (
                  <span key={v} className="flex items-center gap-1.5" style={{ color: VIS[v].border }}>
                    <span className="w-2 h-2 rounded-sm" style={{ background: VIS[v].border }} /> {VIS[v].label}
                  </span>
                ))}
                <span className="flex items-center gap-1.5 text-red-500">
                  <svg width="14" height="6"><line x1="0" y1="3" x2="14" y2="3" stroke="#ef4444" strokeWidth="1.5" strokeDasharray="3 2"/></svg>
                  External dep.
                </span>
              </div>
            </div>
          )}
        </div>

        <div className="h-4 w-px bg-border mx-0.5" />

        {/* Zoom controls */}
        <button type="button" onClick={() => zoomIn()} className="rounded p-1.5 hover:bg-muted text-muted-foreground" title="Zoom in"><ZoomIn size={14} /></button>
        <button type="button" onClick={() => zoomOut()} className="rounded p-1.5 hover:bg-muted text-muted-foreground" title="Zoom out"><ZoomOut size={14} /></button>
        <button type="button" onClick={() => fitView({ padding: 0.05 })} className="rounded p-1.5 hover:bg-muted text-muted-foreground" title="Fit view"><Maximize2 size={14} /></button>

        <div className="h-4 w-px bg-border mx-0.5" />

        {/* Add entity */}
        <div ref={addRef} className="relative">
          <button
            type="button"
            onClick={() => setShowAdd(s => !s)}
            className="flex items-center gap-1 rounded-md bg-primary px-2.5 py-1 text-[11px] font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
          >
            <Plus size={12} /> Add
          </button>
          {showAdd && (
            <div className="absolute right-0 top-full mt-1 z-50 w-[340px] max-h-[70vh] overflow-y-auto rounded-lg border p-4 shadow-xl" style={{ background: '#ffffff', borderColor: '#e5e7eb' }}>
              <div className="mb-3 flex items-center justify-between">
                <span className="text-xs font-semibold text-foreground">Add Entity</span>
                <button type="button" onClick={() => setShowAdd(false)} className="rounded p-0.5 hover:bg-muted text-muted-foreground"><X size={12} /></button>
              </div>
              <ActionForm onAdd={(action) => { addAction(action); setShowAdd(false) }} compact />
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
