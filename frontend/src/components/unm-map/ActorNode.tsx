import { Handle, Position, type NodeProps, type Node } from '@xyflow/react'
import type { PNode } from '@/features/unm-map/types'

type ActorNodeType = Node<PNode>

export function ActorNode({ data }: NodeProps<ActorNodeType>) {
  return (
    <div
      className="flex items-center justify-center gap-1.5 font-semibold rounded-xl select-none cursor-pointer w-full h-full"
      style={{
        background: '#eff6ff', border: '2px solid #3b82f6',
        boxShadow: '0 1px 6px rgba(59,130,246,0.15)',
        fontSize: 11, color: '#1d4ed8',
        opacity: data.dimmed ? 0.1 : 1, transition: 'opacity 0.15s',
      }}
      title={data.label}
    >
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#3b82f6" strokeWidth="2.5">
        <circle cx="12" cy="7" r="4" /><path d="M4 21c0-5 3.6-8 8-8s8 3 8 8" />
      </svg>
      <span style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{data.label}</span>
      <Handle type="source" position={Position.Bottom} style={{ opacity: 0 }} />
    </div>
  )
}
