import { ZoomIn, ZoomOut, Maximize2 } from 'lucide-react'
import { useReactFlow } from '@xyflow/react'
import { VIS_ORDER, VIS } from '@/features/unm-map/constants'

interface MapToolbarProps {
  highlighted: boolean
  onClearHighlight: () => void
}

export function MapToolbar({ highlighted, onClearHighlight }: MapToolbarProps) {
  const { zoomIn, zoomOut, fitView } = useReactFlow()

  return (
    <div className="mb-3 flex items-center gap-4 flex-wrap flex-shrink-0">
      <h2 className="text-xl font-semibold text-gray-900">UNM Map</h2>
      <div className="flex gap-3 text-xs flex-wrap text-gray-500">
        <span className="flex items-center gap-1" style={{ color: '#2563eb' }}>● Actor</span>
        <span className="flex items-center gap-1" style={{ color: '#3b82f6' }}>■ Need</span>
        {VIS_ORDER.map(v => (
          <span key={v} className="flex items-center gap-1" style={{ color: VIS[v].border }}>■ {VIS[v].label}</span>
        ))}
        <span className="flex items-center gap-1.5 pl-3 border-l border-gray-200" style={{ color: '#ef4444' }}>
          <svg width="20" height="8"><line x1="0" y1="4" x2="20" y2="4" stroke="#ef4444" strokeWidth="1.5" strokeDasharray="4 3"/></svg>
          ext. dep
        </span>
        {highlighted && (
          <button
            onClick={onClearHighlight}
            className="ml-3 pl-3 border-l border-gray-200 text-xs text-gray-500 hover:text-gray-900 hover:bg-gray-100 rounded px-2 py-0.5 transition-colors"
          >
            ✕ Clear highlight
          </button>
        )}
      </div>
      <div className="ml-auto flex items-center gap-1">
        <div className="w-px h-5 mx-1 bg-gray-200" />
        <button onClick={() => zoomIn()} className="p-1.5 rounded hover:bg-gray-100" title="Zoom in"><ZoomIn size={15} /></button>
        <button onClick={() => zoomOut()} className="p-1.5 rounded hover:bg-gray-100" title="Zoom out"><ZoomOut size={15} /></button>
        <button onClick={() => fitView({ padding: 0.05 })} className="p-1.5 rounded hover:bg-gray-100" title="Fit view"><Maximize2 size={15} /></button>
      </div>
    </div>
  )
}
