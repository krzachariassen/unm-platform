import { useState, type ReactNode } from 'react'
import { ChevronDown, ChevronRight } from 'lucide-react'
import { cn } from '@/lib/utils'

export interface CollapsibleCardProps {
  title: string
  subtitle?: string
  badge?: ReactNode
  children: ReactNode
  defaultOpen?: boolean
  className?: string
}

export function CollapsibleCard({
  title, subtitle, badge, children, defaultOpen = true, className
}: CollapsibleCardProps) {
  const [open, setOpen] = useState(defaultOpen)

  return (
    <div className={cn('rounded-lg border border-border bg-card', className)}>
      <button
        className="w-full flex items-center gap-3 px-4 py-3 text-left hover:bg-muted/50 transition-colors rounded-lg"
        onClick={() => setOpen(!open)}
      >
        {open ? (
          <ChevronDown className="w-4 h-4 text-muted-foreground shrink-0" />
        ) : (
          <ChevronRight className="w-4 h-4 text-muted-foreground shrink-0" />
        )}
        <span className="font-medium text-foreground">{title}</span>
        {subtitle && <span className="text-sm text-muted-foreground">{subtitle}</span>}
        {badge && <span className="ml-auto">{badge}</span>}
      </button>
      {open && <div className="px-4 pb-4">{children}</div>}
    </div>
  )
}
