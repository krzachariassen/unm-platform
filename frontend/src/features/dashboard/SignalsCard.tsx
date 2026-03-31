import { ArrowRight, AlertTriangle, Users, Zap, Layers, Link2, Flag, TrendingUp, Activity, Network, CheckCircle2 } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import type { SignalsViewResponse } from '@/types/views'

function SignalBar({ count, max, color }: { count: number; max: number; color: string }) {
  const pct = max > 0 ? Math.min((count / max) * 100, 100) : 0
  return (
    <div className="flex-1 h-1.5 rounded-full bg-slate-100 overflow-hidden">
      <div className="h-full rounded-full transition-all duration-700" style={{ width: `${pct}%`, background: color }} />
    </div>
  )
}

type SignalItem = { label: string; count: number; color: string; icon: typeof AlertTriangle; filter: string }

function buildSignals(signals: SignalsViewResponse): SignalItem[] {
  const { user_experience_layer: ux, architecture_layer: arch, organization_layer: org } = signals
  const items: SignalItem[] = []
  if (ux.needs_at_risk.length > 0)                        items.push({ label: 'Needs at risk',                   count: ux.needs_at_risk.length,                              color: '#ef4444', icon: AlertTriangle, filter: 'needs-at-risk'  })
  if (ux.needs_requiring_3plus_teams.length > 0)          items.push({ label: 'Needs served by 3+ teams',        count: ux.needs_requiring_3plus_teams.length,                color: '#f59e0b', icon: Users,         filter: 'needs-at-risk'  })
  if (ux.needs_with_no_capability_backing.length > 0)     items.push({ label: 'Unbacked needs',                  count: ux.needs_with_no_capability_backing.length,           color: '#ef4444', icon: Zap,           filter: 'gap'            })
  if (arch.capabilities_fragmented_across_teams.length > 0)      items.push({ label: 'Fragmented capabilities',  count: arch.capabilities_fragmented_across_teams.length,     color: '#f59e0b', icon: Layers,        filter: 'fragmentation'  })
  if (arch.capabilities_not_connected_to_any_need.length > 0)    items.push({ label: 'Disconnected capabilities',count: arch.capabilities_not_connected_to_any_need.length,   color: '#6366f1', icon: Link2,         filter: 'gap'            })
  if (arch.user_facing_caps_with_cross_team_services.length > 0) items.push({ label: 'Cross-team user-facing',   count: arch.user_facing_caps_with_cross_team_services.length,color: '#f97316', icon: Flag,          filter: 'fragmentation'  })
  if (org.critical_bottleneck_services.length > 0)        items.push({ label: 'Bottleneck services',             count: org.critical_bottleneck_services.length,              color: '#ef4444', icon: TrendingUp,    filter: 'bottleneck'     })
  if (org.top_teams_by_structural_load.length > 0)        items.push({ label: 'High-load teams',                 count: org.top_teams_by_structural_load.length,              color: '#ec4899', icon: Activity,      filter: 'cognitive-load' })
  if (org.low_coherence_teams.length > 0)                 items.push({ label: 'Low-coherence teams',             count: org.low_coherence_teams.length,                       color: '#8b5cf6', icon: Network,       filter: 'cognitive-load' })
  return items
}

export function SignalsCard({ signals }: { signals: SignalsViewResponse }) {
  const navigate = useNavigate()
  const items = buildSignals(signals)
  const max = Math.max(...items.map(s => s.count), 1)
  const total = items.reduce((s, i) => s + i.count, 0)

  return (
    <div className="rounded-lg border border-border bg-card p-6">
      <div className="flex items-center justify-between mb-4">
        <div>
          <h2 className="text-base font-bold text-slate-800 cursor-pointer hover:text-blue-600 transition-colors" onClick={() => navigate('/signals')}>
            Architecture Signals
          </h2>
          <p className="text-xs text-slate-400 mt-0.5">
            {total > 0 ? `${total} finding${total === 1 ? '' : 's'} detected` : 'No issues detected'}
          </p>
        </div>
        <button onClick={() => navigate('/signals')} className="text-xs font-semibold text-indigo-600 bg-indigo-50 hover:bg-indigo-100 px-3 py-1.5 rounded-lg flex items-center gap-1 transition-colors">
          Details <ArrowRight className="w-3 h-3" />
        </button>
      </div>
      {items.length > 0 ? (
        <div className="space-y-2">
          {items.map((item, i) => {
            const Icon = item.icon
            return (
              <div key={i} onClick={() => navigate(`/signals?filter=${item.filter}`)}
                className="flex items-center gap-2.5 cursor-pointer rounded px-1 py-0.5 hover:bg-slate-50 transition-colors">
                <Icon className="w-3.5 h-3.5 shrink-0" style={{ color: item.color }} />
                <span className="text-xs text-slate-500 w-44 shrink-0">{item.label}</span>
                <SignalBar count={item.count} max={max} color={item.color} />
                <span className="text-xs font-bold w-5 text-right" style={{ color: item.color }}>{item.count}</span>
              </div>
            )
          })}
        </div>
      ) : (
        <div className="flex items-center gap-2.5 p-4 rounded-xl bg-green-50 border border-green-200">
          <CheckCircle2 className="w-5 h-5 text-green-500" />
          <div>
            <div className="text-sm font-semibold text-green-700">All clear</div>
            <div className="text-xs text-green-600">No significant architecture signals detected</div>
          </div>
        </div>
      )}
    </div>
  )
}
