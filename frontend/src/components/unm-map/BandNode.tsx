import { type NodeProps, type Node } from '@xyflow/react'
import type { BandInfo } from '@/features/unm-map/types'
import { VIS } from '@/features/unm-map/constants'

type BandNodeData = BandInfo & { canvasWidth: number }
type BandNodeType = Node<BandNodeData>

export function BandNode({ data }: NodeProps<BandNodeType>) {
  const cfg = VIS[data.vis]
  if (!cfg) return null
  return (
    <div
      style={{
        width: data.canvasWidth, height: data.h,
        background: cfg.bandBg,
        borderTop: `0.5px solid ${cfg.border}`,
        position: 'relative', pointerEvents: 'none',
      }}
    >
      <span style={{
        position: 'absolute', right: 14, top: 10,
        fontSize: 9, fontWeight: 700, textTransform: 'uppercase',
        letterSpacing: '0.1em', color: cfg.border, opacity: 0.6,
        fontFamily: 'ui-monospace, monospace',
      }}>
        {cfg.label}
      </span>
    </div>
  )
}
