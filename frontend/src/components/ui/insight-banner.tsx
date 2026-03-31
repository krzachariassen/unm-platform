import { useState, type ReactNode } from 'react'
import { X, Sparkles } from 'lucide-react'
import { cn } from '@/lib/utils'

export interface InsightBannerProps {
  title?: string
  children: ReactNode
  onDismiss?: () => void
  className?: string
}

export function InsightBanner({ title = 'AI Insight', children, onDismiss, className }: InsightBannerProps) {
  const [dismissed, setDismissed] = useState(false)

  if (dismissed) return null

  const handleDismiss = () => {
    setDismissed(true)
    onDismiss?.()
  }

  return (
    <div className={cn(
      'flex gap-3 p-3 rounded-lg border border-primary/20 bg-primary/5 text-sm',
      className
    )}>
      <Sparkles className="w-4 h-4 text-primary mt-0.5 shrink-0" />
      <div className="flex-1 min-w-0">
        <span className="font-medium text-primary mr-2">{title}:</span>
        <span className="text-foreground">{children}</span>
      </div>
      <button
        onClick={handleDismiss}
        className="text-muted-foreground hover:text-foreground shrink-0 transition-colors"
        aria-label="Dismiss insight"
      >
        <X className="w-4 h-4" />
      </button>
    </div>
  )
}
