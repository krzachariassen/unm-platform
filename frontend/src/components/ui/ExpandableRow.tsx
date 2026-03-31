import { useState } from 'react'
import { ChevronDown, ChevronUp, Info, Lightbulb } from 'lucide-react'
import { cn } from '@/lib/utils'

export function ExpandableRow({ summary, explanation, suggestion }: {
  summary: React.ReactNode
  explanation: string
  suggestion: string
}) {
  const [open, setOpen] = useState(false)
  return (
    <div className="overflow-hidden rounded-lg border border-border">
      <button
        type="button"
        onClick={() => setOpen(o => !o)}
        className={cn(
          'flex w-full items-center justify-between gap-3 px-4 py-3 text-left transition-colors',
          open ? 'bg-muted' : 'bg-muted/50 hover:bg-muted'
        )}
      >
        <div className="min-w-0 flex-1">{summary}</div>
        <div className="ml-2 shrink-0 text-muted-foreground">
          {open ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
        </div>
      </button>

      {open && (
        <div className="space-y-3 border-t border-border bg-muted px-4 pb-4 pt-2">
          <div className="flex gap-2.5">
            <Info size={13} className="mt-0.5 shrink-0 text-muted-foreground" />
            <p className="text-xs leading-relaxed text-foreground/90">{explanation}</p>
          </div>
          <div className="flex gap-2.5">
            <Lightbulb size={13} className="mt-0.5 shrink-0 text-primary" />
            <p className="text-xs leading-relaxed text-primary">{suggestion}</p>
          </div>
        </div>
      )}
    </div>
  )
}
