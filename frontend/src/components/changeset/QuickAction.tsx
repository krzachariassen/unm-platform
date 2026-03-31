import { useState, useRef, useEffect, useCallback } from 'react'
import { Pencil } from 'lucide-react'
import type { ChangeAction } from '@/types/changeset'
import { useChangeset } from '@/lib/changeset-context'
import { cn } from '@/lib/utils'

interface QuickActionOption {
  label: string
  action: ChangeAction
}

interface QuickActionProps {
  options: QuickActionOption[]
  size?: number
  onOpen?: () => void
}

export function QuickAction({ options, size = 13, onOpen }: QuickActionProps) {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)
  const { addAction } = useChangeset()

  useEffect(() => {
    if (!open) return
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [open])

  const handleSelect = useCallback((action: ChangeAction) => {
    addAction(action)
    setOpen(false)
    onOpen?.()
  }, [addAction, onOpen])

  if (options.length === 0) return null

  if (options.length === 1) {
    return (
      <button
        type="button"
        onClick={(e) => { e.stopPropagation(); handleSelect(options[0].action) }}
        className="inline-flex items-center justify-center rounded p-0.5 text-muted-foreground transition-colors hover:text-primary"
        title={options[0].label}
      >
        <Pencil size={size} />
      </button>
    )
  }

  return (
    <div ref={ref} className="relative inline-block">
      <button
        type="button"
        onClick={(e) => { e.stopPropagation(); setOpen(!open) }}
        className={cn(
          'inline-flex items-center justify-center rounded p-0.5 transition-colors',
          open ? 'text-primary' : 'text-muted-foreground hover:text-primary'
        )}
        title="Quick edit"
      >
        <Pencil size={size} />
      </button>

      {open && (
        <div className="absolute right-0 top-full z-50 mt-1 min-w-48 rounded-lg border border-border bg-white py-1 shadow-lg">
          {options.map((opt, i) => (
            <button
              key={i}
              type="button"
              className="w-full px-3 py-1.5 text-left text-xs text-foreground transition-colors hover:bg-muted"
              onClick={(e) => { e.stopPropagation(); handleSelect(opt.action) }}
            >
              {opt.label}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
