import { type ReactNode } from 'react'
import { TrendingUp, TrendingDown, Minus } from 'lucide-react'
import { cn } from '@/lib/utils'

export type TrendDirection = 'up' | 'down' | 'neutral'

export interface StatCardProps {
  label: string
  value: string | number | ReactNode
  description?: string
  icon?: ReactNode
  trend?: { direction: TrendDirection; label: string }
  className?: string
}

const trendIcon = { up: TrendingUp, down: TrendingDown, neutral: Minus }
const trendColor = { up: 'text-green-600', down: 'text-red-600', neutral: 'text-muted-foreground' }

export function StatCard({ label, value, description, icon, trend, className }: StatCardProps) {
  const TrendIcon = trend ? trendIcon[trend.direction] : null
  return (
    <div className={cn('rounded-lg border border-border bg-card p-4 flex flex-col gap-2', className)}>
      <div className="flex items-center justify-between">
        <span className="text-sm font-medium text-muted-foreground">{label}</span>
        {icon && <span className="text-muted-foreground">{icon}</span>}
      </div>
      <span className="text-2xl font-bold text-foreground">{value}</span>
      {description && <span className="text-xs text-muted-foreground">{description}</span>}
      {trend && TrendIcon && (
        <div className={cn('flex items-center gap-1 text-xs', trendColor[trend.direction])}>
          <TrendIcon className="w-3 h-3" />
          <span>{trend.label}</span>
        </div>
      )}
    </div>
  )
}
