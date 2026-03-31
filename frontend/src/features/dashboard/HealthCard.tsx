import { Shield } from 'lucide-react'
import { HealthRing } from '@/components/ui/health-ring'
import type { SignalsViewResponse } from '@/types/views'
import { cn } from '@/lib/utils'

const HEALTH_CONFIG: Record<string, { color: string; bg: string; ring: string; label: string; score: number }> = {
  red:   { color: 'text-red-600',   bg: 'bg-red-50',   ring: '#ef4444', label: 'Critical', score: 33 },
  amber: { color: 'text-amber-600', bg: 'bg-amber-50', ring: '#f59e0b', label: 'Warning',  score: 66 },
  green: { color: 'text-green-600', bg: 'bg-green-50', ring: '#22c55e', label: 'Healthy',  score: 100 },
}

const LAYERS = [
  { key: 'ux_risk' as const,           label: 'UX',   sublabel: 'User Experience' },
  { key: 'architecture_risk' as const, label: 'Arch', sublabel: 'Architecture' },
  { key: 'org_risk' as const,          label: 'Org',  sublabel: 'Organization' },
]

export function HealthCard({ signals }: { signals: SignalsViewResponse }) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="flex items-center justify-between mb-5">
        <div>
          <h2 className="text-base font-bold text-slate-800">Platform Health</h2>
          <p className="text-xs text-slate-400 mt-0.5">Across three architecture layers</p>
        </div>
        <Shield className="w-5 h-5 text-slate-300" />
      </div>
      <div className="flex justify-around gap-3">
        {LAYERS.map(({ key, label, sublabel }) => {
          const level = signals.health[key]
          const cfg = HEALTH_CONFIG[level]
          return (
            <div key={key} className="text-center">
              <div className="relative inline-block">
                <HealthRing value={cfg.score} color={cfg.ring} />
                <div className="absolute inset-0 flex items-center justify-center">
                  <span className={cn('text-base font-extrabold', cfg.color)}>{label}</span>
                </div>
              </div>
              <div className="mt-2">
                <span className={cn('text-[11px] font-bold px-2.5 py-0.5 rounded-md inline-block', cfg.bg, cfg.color)}>
                  {cfg.label}
                </span>
                <div className="text-[10px] text-slate-400 mt-1">{sublabel}</div>
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
