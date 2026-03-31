import { AlertTriangle, AlertCircle } from 'lucide-react'
import { SlidePanel, PanelSection } from '@/components/ui/slide-panel'

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

const NODE_TYPE_BADGE: Record<string, { bg: string; text: string }> = {
  actor:       { bg: '#ede9fe', text: '#5b21b6' },
  need:        { bg: '#dbeafe', text: '#1e40af' },
  capability:  { bg: '#d1fae5', text: '#065f46' },
  service:     { bg: '#f3f4f6', text: '#374151' },
  team:        { bg: '#fef3c7', text: '#92400e' },
}

export function AntiPatternPanel({ node, onClose }: AntiPatternPanelProps) {
  const antiPatterns = (node?.data.anti_patterns as Array<{ code: string; message: string; severity: string }> | undefined) ?? []
  const skipKeys = new Set(['nodeType', 'label', 'is_fragmented', 'is_overloaded', 'is_mapped', 'is_leaf', 'anti_patterns'])
  const dataEntries = node ? Object.entries(node.data).filter(([k]) => !skipKeys.has(k)) : []
  const typeBadge = node ? (NODE_TYPE_BADGE[node.nodeType] ?? { bg: '#f3f4f6', text: '#374151' }) : { bg: '#f3f4f6', text: '#374151' }

  return (
    <SlidePanel
      open={!!node}
      onClose={onClose}
      title={node?.label ?? ''}
      badge={node && (
        <span
          className="inline-flex items-center rounded px-2 py-0.5 text-[10px] font-semibold"
          style={{ background: typeBadge.bg, color: typeBadge.text }}
        >
          {node.nodeType}
        </span>
      )}
    >
      <div className="space-y-3">
        {antiPatterns.length > 0 && (
          <PanelSection label="Anti-patterns">
            <div className="space-y-1.5 rounded-lg border border-destructive/25 bg-destructive/10 p-2.5">
              {antiPatterns.map((ap, i) => (
                <div key={i} className="flex items-start gap-2 text-xs" style={{ color: ap.severity === 'error' ? '#b91c1c' : '#c2410c' }}>
                  {ap.severity === 'error' ? <AlertCircle size={12} className="mt-0.5 shrink-0" /> : <AlertTriangle size={12} className="mt-0.5 shrink-0" />}
                  <span className="leading-normal">{ap.message}</span>
                </div>
              ))}
            </div>
          </PanelSection>
        )}

        {dataEntries.length > 0 && (
          <div className="space-y-1">
            {dataEntries.map(([k, v]) => (
              <div key={k} className="flex justify-between gap-4 text-xs">
                <span className="text-muted-foreground">{k.replace(/_/g, ' ')}</span>
                <span className="text-right font-medium text-foreground">{String(v)}</span>
              </div>
            ))}
          </div>
        )}

        {antiPatterns.length === 0 && dataEntries.length === 0 && (
          <p className="text-xs text-muted-foreground">No additional details.</p>
        )}
      </div>
    </SlidePanel>
  )
}
