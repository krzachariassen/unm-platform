import { Handle, Position, type NodeProps, type Node } from '@xyflow/react'
import type { PNode } from '@/features/unm-map/types'

type ExtDepNodeType = Node<PNode>

export function ExtDepNode({ data }: NodeProps<ExtDepNodeType>) {
  const dep = data.extDep
  const bg = dep?.is_critical ? '#fef2f2' : dep?.is_warning ? '#fffbeb' : '#f1f5f9'
  const borderColor = dep?.is_critical ? '#ef4444' : dep?.is_warning ? '#f59e0b' : '#94a3b8'
  const color = dep?.is_critical ? '#b91c1c' : dep?.is_warning ? '#92400e' : '#475569'

  return (
    <div
      className="flex items-center justify-center rounded-lg select-none cursor-pointer w-full h-full"
      style={{
        background: bg, border: `1.5px dashed ${borderColor}`,
        fontSize: 9, fontWeight: 600, color,
        opacity: data.dimmed ? 0.1 : 1, transition: 'opacity 0.15s',
      }}
      title={data.label}
    >
      <span style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', padding: '0 6px' }}>
        {data.label}
      </span>
      <Handle type="target" position={Position.Top} style={{ opacity: 0 }} />
    </div>
  )
}
