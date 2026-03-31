import { useState, useRef, useEffect, useCallback } from 'react'
import { Pencil } from 'lucide-react'
import type { ChangeAction } from '@/lib/api'
import { useChangeset } from '@/lib/changeset-context'

interface QuickActionOption {
  label: string
  action: ChangeAction
}

interface QuickActionProps {
  options: QuickActionOption[]
  size?: number
  onOpen?: () => void  // callback to open EditPanel if the caller wants to show it
}

export function QuickAction({ options, size = 13, onOpen }: QuickActionProps) {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)
  const { addAction, enterEditMode } = useChangeset()

  useEffect(() => {
    if (!open) return
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [open])

  const handleSelect = useCallback((action: ChangeAction) => {
    enterEditMode()
    addAction(action)
    setOpen(false)
    onOpen?.()
  }, [enterEditMode, addAction, onOpen])

  if (options.length === 0) return null

  if (options.length === 1) {
    return (
      <button
        onClick={(e) => { e.stopPropagation(); handleSelect(options[0].action) }}
        className="inline-flex items-center justify-center rounded p-0.5 transition-colors"
        style={{ color: '#9ca3af' }}
        onMouseEnter={e => { (e.currentTarget).style.color = '#2563eb' }}
        onMouseLeave={e => { (e.currentTarget).style.color = '#9ca3af' }}
        title={options[0].label}
      >
        <Pencil size={size} />
      </button>
    )
  }

  return (
    <div ref={ref} className="relative inline-block">
      <button
        onClick={(e) => { e.stopPropagation(); setOpen(!open) }}
        className="inline-flex items-center justify-center rounded p-0.5 transition-colors"
        style={{ color: '#9ca3af' }}
        onMouseEnter={e => { (e.currentTarget).style.color = '#2563eb' }}
        onMouseLeave={e => { if (!open) (e.currentTarget).style.color = '#9ca3af' }}
        title="Quick edit"
      >
        <Pencil size={size} />
      </button>

      {open && (
        <div
          className="absolute right-0 top-full mt-1 z-50 min-w-48 rounded-lg py-1 shadow-lg"
          style={{ background: '#ffffff', border: '1px solid #e5e7eb' }}
        >
          {options.map((opt, i) => (
            <button
              key={i}
              className="w-full text-left px-3 py-1.5 text-xs transition-colors"
              style={{ color: '#374151' }}
              onMouseEnter={e => { (e.currentTarget).style.background = '#f3f4f6' }}
              onMouseLeave={e => { (e.currentTarget).style.background = 'transparent' }}
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
