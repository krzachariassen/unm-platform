import type { ImpactDelta } from '@/types/changeset'
import { cn } from '@/lib/utils'

const CHANGE_STYLES = {
  improved: {
    label: 'Improved',
    card: 'border-green-200 bg-green-50 dark:border-green-900 dark:bg-green-950/40',
    pill: 'text-green-800 bg-green-100 dark:text-green-200 dark:bg-green-900/50',
    value: 'text-green-800 dark:text-green-300',
    summary: 'text-green-700 dark:text-green-400',
  },
  regressed: {
    label: 'Regressed',
    card: 'border-red-200 bg-red-50 dark:border-red-900 dark:bg-red-950/40',
    pill: 'text-red-800 bg-red-100 dark:text-red-200 dark:bg-red-900/50',
    value: 'text-red-800 dark:text-red-300',
    summary: 'text-red-700 dark:text-red-400',
  },
  unchanged: {
    label: 'Unchanged',
    card: 'border-border bg-muted',
    pill: 'text-muted-foreground bg-muted',
    value: 'text-muted-foreground',
    summary: 'text-muted-foreground',
  },
} as const

interface ImpactPanelProps {
  deltas: ImpactDelta[]
  changesetId: string
}

export function ImpactPanel({ deltas, changesetId }: ImpactPanelProps) {
  if (deltas.length === 0) {
    return (
      <div className="py-6 text-center">
        <p className="text-sm text-muted-foreground">No impact data available</p>
      </div>
    )
  }

  const improved = deltas.filter(d => d.change === 'improved').length
  const regressed = deltas.filter(d => d.change === 'regressed').length

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <span className="text-xs font-medium text-muted-foreground">
          Changeset: {changesetId.slice(0, 8)}...
        </span>
        <div className="flex gap-2 text-xs">
          {improved > 0 && <span className="text-green-700 dark:text-green-400">{improved} improved</span>}
          {regressed > 0 && <span className="text-destructive">{regressed} regressed</span>}
        </div>
      </div>

      {deltas.map((delta, i) => {
        const style = CHANGE_STYLES[delta.change]
        return (
          <div
            key={i}
            className={cn('rounded-lg border p-3', style.card)}
          >
            <div className="mb-1 flex items-center justify-between">
              <span className="text-sm font-medium text-foreground">
                {delta.dimension}
              </span>
              <span className={cn('rounded-full px-2 py-0.5 text-xs font-semibold', style.pill)}>
                {style.label}
              </span>
            </div>
            <div className="flex items-center gap-2 text-xs text-foreground">
              <span>{delta.before}</span>
              <span className="text-muted-foreground">&rarr;</span>
              <span className={cn('font-semibold', style.value)}>{delta.after}</span>
            </div>
            {delta.detail && (
              <p className="mt-1 text-xs text-muted-foreground">{delta.detail}</p>
            )}
          </div>
        )
      })}
    </div>
  )
}
