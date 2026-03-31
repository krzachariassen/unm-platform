import { SlidePanel, PanelField } from '@/components/ui/slide-panel'
import type { PanelItem } from '@/features/unm-map/types'

interface DetailDrawerProps {
  panel: PanelItem | null
  children?: React.ReactNode
  onClose: () => void
}

export function DetailDrawer({ panel, children, onClose }: DetailDrawerProps) {
  return (
    <SlidePanel
      open={!!panel}
      onClose={onClose}
      title={panel?.title ?? ''}
      subtitle={panel?.subtitle}
      backdrop={false}
      badge={panel?.badge && (
        <span
          className="text-[10px] font-semibold px-2 py-0.5 rounded uppercase tracking-wide"
          style={{ background: panel.badge.color + '18', color: panel.badge.color, border: `1px solid ${panel.badge.color}44` }}
        >
          {panel.badge.text}
        </span>
      )}
    >
      {panel && (
        <div>
          {/* Edit form when available */}
          {children && (
            <div className="pb-4 mb-4" style={{ borderBottom: '1px solid #f3f4f6' }}>
              {children}
            </div>
          )}

          {/* Entity detail fields */}
          <div>
            {children && (
              <p className="text-[10px] font-semibold uppercase tracking-wider mb-2" style={{ color: '#9ca3af' }}>Details</p>
            )}
            <div className="space-y-1">
              {panel.fields.map((f, i) => (
                <PanelField key={i} label={f.label} value={f.value || '—'} />
              ))}
            </div>
          </div>
        </div>
      )}
    </SlidePanel>
  )
}
