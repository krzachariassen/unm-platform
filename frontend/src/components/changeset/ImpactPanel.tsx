import type { ImpactDelta } from '@/lib/api'

const CHANGE_STYLES = {
  improved: { color: '#15803d', bg: '#f0fdf4', border: '#bbf7d0', label: 'Improved' },
  regressed: { color: '#b91c1c', bg: '#fef2f2', border: '#fca5a5', label: 'Regressed' },
  unchanged: { color: '#6b7280', bg: '#f9fafb', border: '#e5e7eb', label: 'Unchanged' },
}

interface ImpactPanelProps {
  deltas: ImpactDelta[]
  changesetId: string
}

export function ImpactPanel({ deltas, changesetId }: ImpactPanelProps) {
  if (deltas.length === 0) {
    return (
      <div className="text-center py-6">
        <p className="text-sm" style={{ color: '#9ca3af' }}>No impact data available</p>
      </div>
    )
  }

  const improved = deltas.filter(d => d.change === 'improved').length
  const regressed = deltas.filter(d => d.change === 'regressed').length

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <span className="text-xs font-medium" style={{ color: '#6b7280' }}>
          Changeset: {changesetId.slice(0, 8)}...
        </span>
        <div className="flex gap-2 text-xs">
          {improved > 0 && <span style={{ color: '#15803d' }}>{improved} improved</span>}
          {regressed > 0 && <span style={{ color: '#b91c1c' }}>{regressed} regressed</span>}
        </div>
      </div>

      {deltas.map((delta, i) => {
        const style = CHANGE_STYLES[delta.change]
        return (
          <div
            key={i}
            className="rounded-lg p-3"
            style={{ border: `1px solid ${style.border}`, background: style.bg }}
          >
            <div className="flex items-center justify-between mb-1">
              <span className="text-sm font-medium" style={{ color: '#111827' }}>
                {delta.dimension}
              </span>
              <span className="text-xs font-semibold px-2 py-0.5 rounded-full" style={{ color: style.color, background: `${style.color}15` }}>
                {style.label}
              </span>
            </div>
            <div className="flex items-center gap-2 text-xs" style={{ color: '#374151' }}>
              <span>{delta.before}</span>
              <span style={{ color: '#9ca3af' }}>&rarr;</span>
              <span style={{ color: style.color, fontWeight: 600 }}>{delta.after}</span>
            </div>
            {delta.detail && (
              <p className="text-xs mt-1" style={{ color: '#6b7280' }}>{delta.detail}</p>
            )}
          </div>
        )
      })}
    </div>
  )
}
