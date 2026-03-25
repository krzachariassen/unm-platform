import { X, AlertTriangle, AlertCircle } from 'lucide-react'

interface NodeDetails {
  id: string
  label: string
  nodeType: string
  data: Record<string, unknown>
}

interface AntiPatternPanelProps {
  node: NodeDetails | null
  onClose: () => void
}

function AntiPatternItem({ icon: Icon, color, text }: { icon: typeof AlertTriangle; color: string; text: string }) {
  return (
    <div className="flex items-start gap-2 text-sm" style={{ color }}>
      <Icon size={13} style={{ flexShrink: 0, marginTop: 2 }} />
      <span style={{ lineHeight: '1.5' }}>{text}</span>
    </div>
  )
}

function DataRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex justify-between gap-4" style={{ fontSize: 13 }}>
      <span style={{ color: '#6b7280' }}>{label.replace(/_/g, ' ')}</span>
      <span className="font-medium text-right" style={{ color: '#111827' }}>{value}</span>
    </div>
  )
}

export function AntiPatternPanel({ node, onClose }: AntiPatternPanelProps) {
  if (!node) return null

  const antiPatterns = (node.data.anti_patterns as Array<{ code: string; message: string; severity: string }> | undefined) ?? []
  const skipKeys = new Set(['nodeType', 'label', 'is_fragmented', 'is_overloaded', 'is_mapped', 'is_leaf', 'anti_patterns'])
  const dataEntries = Object.entries(node.data).filter(([k]) => !skipKeys.has(k))

  const NODE_TYPE_BADGE: Record<string, { bg: string; text: string }> = {
    actor:       { bg: '#ede9fe', text: '#5b21b6' },
    need:        { bg: '#dbeafe', text: '#1e40af' },
    capability:  { bg: '#d1fae5', text: '#065f46' },
    service:     { bg: '#f3f4f6', text: '#374151' },
    team:        { bg: '#fef3c7', text: '#92400e' },
  }
  const typeBadge = NODE_TYPE_BADGE[node.nodeType] ?? { bg: '#f3f4f6', text: '#374151' }

  return (
    <>
    <div className="fixed inset-0 z-40" onClick={onClose} />
    <div className="fixed right-4 top-16 z-50 w-76" style={{ width: 296 }}>
      <div
        className="rounded-xl shadow-lg overflow-hidden"
        style={{ border: '1px solid #e5e7eb', background: '#ffffff' }}
      >
        {/* Header */}
        <div className="px-4 py-3" style={{ borderBottom: '1px solid #f3f4f6' }}>
          <div className="flex items-start justify-between gap-2">
            <div className="flex-1 min-w-0">
              <span
                className="inline-flex items-center rounded px-2 py-0.5 text-xs font-medium mb-1.5"
                style={{ background: typeBadge.bg, color: typeBadge.text }}
              >
                {node.nodeType}
              </span>
              <p className="font-semibold text-sm leading-snug" style={{ color: '#111827' }}>{node.label}</p>
            </div>
            <button
              onClick={onClose}
              className="flex-shrink-0 mt-0.5 p-1 rounded hover:bg-gray-100 transition-colors"
              style={{ color: '#9ca3af' }}
            >
              <X size={14} />
            </button>
          </div>
        </div>

        {/* Body */}
        <div className="px-4 py-3 space-y-3">
          {antiPatterns.length > 0 && (
            <div className="space-y-2 p-3 rounded-lg" style={{ background: '#fef2f2', border: '1px solid #fecaca' }}>
              <p className="text-xs font-semibold" style={{ color: '#9ca3af' }}>ANTI-PATTERNS</p>
              {antiPatterns.map((ap, i) => (
                <AntiPatternItem
                  key={i}
                  icon={ap.severity === 'error' ? AlertCircle : AlertTriangle}
                  color={ap.severity === 'error' ? '#b91c1c' : '#c2410c'}
                  text={ap.message}
                />
              ))}
            </div>
          )}

          {dataEntries.length > 0 && (
            <div className="space-y-1.5">
              {dataEntries.map(([k, v]) => (
                <DataRow key={k} label={k} value={String(v)} />
              ))}
            </div>
          )}

          {antiPatterns.length === 0 && dataEntries.length === 0 && (
            <p className="text-sm" style={{ color: '#9ca3af' }}>No additional details.</p>
          )}
        </div>
      </div>
    </div>
    </>
  )
}
