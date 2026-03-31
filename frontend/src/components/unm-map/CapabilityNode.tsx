import { Handle, Position, type NodeProps, type Node } from '@xyflow/react'
import type { PNode } from '@/features/unm-map/types'
import { VIS, teamColor } from '@/features/unm-map/constants'

type CapabilityNodeType = Node<PNode>

export function CapabilityNode({ data }: NodeProps<CapabilityNodeType>) {
  const cfg = VIS[data.vis ?? 'foundational'] ?? VIS.foundational
  const svcs = data.svcs ?? []
  const MAX_SHOWN = 3
  const shown = svcs.slice(0, MAX_SHOWN)
  const extra = svcs.length - MAX_SHOWN
  const isVirtualPending = data.id.startsWith('pending:')
  const crossTeam = data.crossTeam

  let borderColor = cfg.border
  let boxShadow: string | undefined
  if (isVirtualPending) {
    borderColor = cfg.border
    boxShadow = `0 0 0 3px ${cfg.border}22`
  } else if (data.isFragmented || crossTeam) {
    borderColor = '#ef4444'
    boxShadow = '0 0 8px rgba(239,68,68,0.3)'
  }

  return (
    <div
      className="absolute rounded-lg select-none cursor-pointer w-full h-full overflow-hidden"
      style={{
        background: cfg.nodeBg,
        border: `${isVirtualPending ? '2px dashed' : '1.5px solid'} ${borderColor}`,
        boxShadow,
        opacity: data.dimmed ? 0.1 : (isVirtualPending ? 0.85 : 1),
        transition: 'opacity 0.15s, border-color 0.2s',
      }}
      title={data.label}
    >
      <div style={{ fontSize: 10, fontWeight: 600, color: cfg.text, padding: '5px 8px 2px', lineHeight: 1.3 }}>
        {data.label}
      </div>
      {data.team && (
        <div style={{ fontSize: 9, color: '#6b7280', paddingLeft: 8, lineHeight: 1.2, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
          {data.team.label}
        </div>
      )}
      {svcs.length > 0 && (
        <div style={{ padding: '3px 6px 0', borderTop: `1px solid ${cfg.border}33`, marginTop: 3 }}>
          {shown.map(s => (
            <div key={s.id} style={{ display: 'flex', alignItems: 'center', gap: 4, marginBottom: 2 }}>
              <div style={{ width: 5, height: 5, borderRadius: '50%', background: teamColor(s.teamName), flexShrink: 0 }} />
              <span style={{ fontSize: 8.5, color: '#6b7280', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', flex: 1 }}>
                {s.label}
              </span>
            </div>
          ))}
          {extra > 0 && <div style={{ fontSize: 8, color: '#9ca3af', paddingLeft: 9 }}>+{extra} more</div>}
        </div>
      )}
      {(data.isFragmented || crossTeam) && (
        <div style={{ position: 'absolute', top: 4, right: 6, fontSize: 9, color: '#ef4444' }}>⚠</div>
      )}
      {isVirtualPending && (
        <div style={{ position: 'absolute', top: 3, right: 5, fontSize: 8, color: cfg.border, fontWeight: 700, background: `${cfg.border}18`, borderRadius: 3, padding: '1px 4px' }}>pending</div>
      )}
      <Handle type="target" position={Position.Top} style={{ opacity: 0 }} />
      <Handle type="source" position={Position.Bottom} style={{ opacity: 0 }} />
    </div>
  )
}
