import type { PanelItem } from '@/features/unm-map/types'

interface DetailDrawerProps {
  panel: PanelItem | null
  isEditMode: boolean
  children?: React.ReactNode
  onClose: () => void
}

export function DetailDrawer({ panel, isEditMode, children, onClose }: DetailDrawerProps) {
  return (
    <div
      style={{
        position: 'fixed', right: 0, top: 56, bottom: 0, width: 320,
        background: 'white', borderLeft: '1px solid #e5e7eb',
        overflowY: 'auto', zIndex: 50,
        transform: panel ? 'translateX(0)' : 'translateX(100%)',
        transition: 'transform 0.2s ease',
        boxShadow: '-4px 0 12px rgba(0,0,0,0.08)',
      }}
      onClick={e => e.stopPropagation()}
    >
      {panel && (
        <>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '12px 16px', borderBottom: '1px solid #e5e7eb' }}>
            <div style={{ flex: 1, minWidth: 0 }}>
              {panel.badge && (
                <span style={{
                  display: 'inline-block', marginBottom: 6, padding: '2px 8px', borderRadius: 4,
                  fontSize: 12, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.05em',
                  background: panel.badge.color + '18', color: panel.badge.color,
                  border: `1px solid ${panel.badge.color}44`,
                }}>
                  {panel.badge.text}
                </span>
              )}
              <div style={{ fontWeight: 600, fontSize: 14, color: '#111827', lineHeight: 1.3 }}>{panel.title}</div>
              {panel.subtitle && <div style={{ fontSize: 12, color: '#6b7280', marginTop: 2 }}>{panel.subtitle}</div>}
            </div>
            <button
              onClick={onClose}
              style={{ background: 'none', border: 'none', cursor: 'pointer', color: '#6b7280', fontSize: 18, lineHeight: 1, marginLeft: 8, flexShrink: 0 }}
            >
              ✕
            </button>
          </div>
          <div style={{ padding: 16 }}>
            {children}
            {(!isEditMode || !children) && panel.fields.map((f, i) => (
              <div key={i} style={{ marginBottom: 12 }}>
                <div style={{ fontSize: 11, fontWeight: 500, textTransform: 'uppercase', letterSpacing: '0.05em', color: f.label.startsWith('⚠') ? '#ef4444' : '#9ca3af', marginBottom: 4 }}>{f.label}</div>
                <div style={{ fontSize: 13, lineHeight: 1.5, color: f.label.startsWith('⚠') ? '#dc2626' : '#374151', whiteSpace: 'pre-line' }}>{f.value || '—'}</div>
              </div>
            ))}
            {isEditMode && children && (
              <details style={{ marginTop: 4 }}>
                <summary style={{ fontSize: 11, color: '#9ca3af', cursor: 'pointer', userSelect: 'none' }}>▸ View details</summary>
                <div style={{ marginTop: 8 }}>
                  {panel.fields.map((f, i) => (
                    <div key={i} style={{ marginBottom: 12 }}>
                      <div style={{ fontSize: 11, fontWeight: 500, textTransform: 'uppercase', letterSpacing: '0.05em', color: f.label.startsWith('⚠') ? '#ef4444' : '#9ca3af', marginBottom: 4 }}>{f.label}</div>
                      <div style={{ fontSize: 13, lineHeight: 1.5, color: f.label.startsWith('⚠') ? '#dc2626' : '#374151', whiteSpace: 'pre-line' }}>{f.value || '—'}</div>
                    </div>
                  ))}
                </div>
              </details>
            )}
          </div>
        </>
      )}
    </div>
  )
}
