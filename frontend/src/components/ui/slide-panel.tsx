import { useEffect, type ReactNode } from 'react'
import { X } from 'lucide-react'
import { cn } from '@/lib/utils'

export interface SlidePanelProps {
  open: boolean
  onClose: () => void
  title: string
  subtitle?: string
  badge?: ReactNode
  children: ReactNode
  className?: string
  /** When false, no backdrop overlay is shown and the content behind remains interactive. */
  backdrop?: boolean
}

export function SlidePanel({ open, onClose, title, subtitle, badge, children, className, backdrop = true }: SlidePanelProps) {
  useEffect(() => {
    if (!open) return
    const handler = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose() }
    document.addEventListener('keydown', handler)
    return () => document.removeEventListener('keydown', handler)
  }, [open, onClose])

  return (
    <>
      {backdrop && (
        <div
          className={cn('fixed inset-0 z-40 bg-black/10 backdrop-blur-[2px] transition-opacity duration-200', open ? 'opacity-100' : 'opacity-0 pointer-events-none')}
          onClick={onClose}
          aria-hidden
        />
      )}

      <div
        className={cn(
          'fixed top-0 right-0 h-full w-[380px] z-50 border-l flex flex-col transition-transform duration-200 ease-out',
          open ? 'translate-x-0' : 'translate-x-full',
          className,
        )}
        style={{ background: '#ffffff', borderColor: '#e5e7eb' }}
      >
        <div className="flex items-start justify-between gap-3 px-5 py-4 shrink-0" style={{ borderBottom: '1px solid #e5e7eb' }}>
          <div className="min-w-0 flex-1">
            {badge && <div className="mb-1.5">{badge}</div>}
            <h3 className="text-sm font-semibold leading-snug truncate" style={{ color: '#111827' }}>{title}</h3>
            {subtitle && <p className="text-xs mt-0.5 truncate" style={{ color: '#9ca3af' }}>{subtitle}</p>}
          </div>
          <button
            type="button"
            onClick={onClose}
            className="shrink-0 rounded-md p-1.5 transition-colors hover:bg-gray-100"
            style={{ color: '#9ca3af' }}
            aria-label="Close"
          >
            <X size={16} />
          </button>
        </div>

        <div className="flex-1 overflow-y-auto px-5 py-4">
          {children}
        </div>
      </div>
    </>
  )
}

export function PanelSection({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="mb-4">
      <p className="text-[10px] font-semibold uppercase tracking-wider mb-2" style={{ color: '#9ca3af' }}>{label}</p>
      {children}
    </div>
  )
}

export function PanelField({ label, value }: { label: string; value: string | ReactNode }) {
  return (
    <div className="mb-3">
      <p className="text-[10px] font-semibold uppercase tracking-wider mb-0.5" style={{ color: '#9ca3af' }}>{label}</p>
      <div className="text-xs leading-relaxed whitespace-pre-line" style={{ color: '#374151' }}>{value || '—'}</div>
    </div>
  )
}
