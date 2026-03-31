import { cn } from '@/lib/utils'
import type { RiskLevel } from '@/types/common'

const riskConfig: Record<RiskLevel, { label: string; className: string }> = {
  red:   { label: 'Critical', className: 'bg-red-100 text-red-700 border-red-200' },
  amber: { label: 'Warning',  className: 'bg-amber-100 text-amber-700 border-amber-200' },
  green: { label: 'Healthy',  className: 'bg-green-100 text-green-700 border-green-200' },
}

export interface RiskBadgeProps {
  level: RiskLevel
  label?: string
  className?: string
}

export function RiskBadge({ level, label, className }: RiskBadgeProps) {
  const cfg = riskConfig[level]
  return (
    <span className={cn(
      'inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium border',
      cfg.className,
      className
    )}>
      {label ?? cfg.label}
    </span>
  )
}
