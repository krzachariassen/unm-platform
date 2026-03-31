import { Handle, Position, type NodeProps, type Node } from '@xyflow/react'
import type { PNode } from '@/features/unm-map/types'

type NeedNodeType = Node<PNode>

export function NeedNode({ data }: NodeProps<NeedNodeType>) {
  const unmapped = data.isMapped === false
  return (
    <div
      className="flex items-center justify-center text-center rounded-lg select-none cursor-pointer w-full h-full"
      style={{
        background: unmapped ? '#fef2f2' : '#eff6ff',
        border: `1.5px solid ${unmapped ? '#ef4444' : '#3b82f6'}`,
        fontSize: 10, fontWeight: 500, lineHeight: 1.3,
        color: unmapped ? '#b91c1c' : '#1e40af',
        padding: '4px 8px',
        opacity: data.dimmed ? 0.1 : 1, transition: 'opacity 0.15s',
      }}
      title={data.label}
    >
      {data.label}
      <Handle type="target" position={Position.Top} style={{ opacity: 0 }} />
      <Handle type="source" position={Position.Bottom} style={{ opacity: 0 }} />
    </div>
  )
}
