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
  const hasValidationError = (data as unknown as Record<string, unknown>).hasValidationError === true
  const isLeafNoService = !isVirtualPending && svcs.length === 0 && !(data as unknown as Record<string, unknown>).children

  let borderColor = cfg.border
  let boxShadow: string | undefined
  if (hasValidationError || isLeafNoService) {
    borderColor = '#ef4444'
    boxShadow = '0 0 0 2px #fecaca'
  } else if (isVirtualPending) {
    borderColor = '#f59e0b'
    boxShadow = '0 0 0 2px #fde68a'
  } else if (data.isFragmented || crossTeam) {
    borderColor = '#ef4444'
    boxShadow = '0 0 8px rgba(239,68,68,0.3)'
  }

  return (
    <div
      className="absolute rounded-lg select-none cursor-pointer w-full h-full overflow-hidden"
      style={{
        background: isVirtualPending ? '#fffbeb' : cfg.nodeBg,
        border: `${isVirtualPending ? '2px dashed' : '1.5px solid'} ${borderColor}`,
        boxShadow,
        opacity: data.dimmed ? 0.1 : 1,
        transition: 'opacity 0.15s, border-color 0.2s',
      }}
      title={data.label}
    >
      <div style={{ fontSize: 10, fontWeight: 600, color: isVirtualPending ? '#92400e' : cfg.text, padding: '5px 8px 2px', lineHeight: 1.3 }}>
        {data.label}
      </div>
      {data.team && (
        <div style={{ fontSize: 9, color: '#6b7280', paddingLeft: 8, lineHeight: 1.2, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
          {data.team.label}
        </div>
      )}
      {!data.team && isVirtualPending && (
        <div style={{ fontSize: 9, color: '#b45309', paddingLeft: 8, lineHeight: 1.2, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
          {String((data as unknown as Record<string, unknown>).pendingTeam ?? '')}
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
      {/* Error/warning indicators */}
      {(hasValidationError || isLeafNoService) && (
        <div style={{ position: 'absolute', top: 3, right: 5, fontSize: 10, color: '#dc2626' }} title="No service linked">⚠</div>
      )}
      {(data.isFragmented || crossTeam) && !hasValidationError && !isLeafNoService && (
        <div style={{ position: 'absolute', top: 4, right: 6, fontSize: 9, color: '#ef4444' }}>⚠</div>
      )}
      {isVirtualPending && !hasValidationError && (
        <div style={{ position: 'absolute', top: 3, right: 5, fontSize: 8, color: '#b45309', fontWeight: 700, background: '#fef3c7', borderRadius: 3, padding: '1px 4px', border: '1px solid #fde68a' }}>new</div>
      )}
      <Handle type="target" position={Position.Top} style={{ opacity: 0 }} />
      <Handle type="source" position={Position.Bottom} style={{ opacity: 0 }} />
    </div>
  )
}
