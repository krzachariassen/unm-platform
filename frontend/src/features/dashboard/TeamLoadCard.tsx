import { ArrowRight } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import type { TeamLoad } from '@/types/views'
import { cn } from '@/lib/utils'

const LEVEL_CLASS: Record<string, { dot: string; badge: string; text: string }> = {
  high:   { dot: 'bg-red-500',   badge: 'bg-red-50',   text: 'text-red-600' },
  medium: { dot: 'bg-amber-500', badge: 'bg-amber-50', text: 'text-amber-600' },
  low:    { dot: 'bg-green-500', badge: 'bg-green-50', text: 'text-green-600' },
}

export function TeamLoadCard({ teams }: { teams: TeamLoad[] }) {
  const navigate = useNavigate()
  const top5 = [...teams]
    .sort((a, b) => {
      const r = (l: string) => l === 'high' ? 3 : l === 'medium' ? 2 : 1
      return r(b.overall_level) - r(a.overall_level) || a.team.name.localeCompare(b.team.name)
    })
    .slice(0, 5)

  if (top5.length === 0) return null

  return (
    <div className="rounded-2xl border border-slate-200 bg-white p-6">
      <div className="flex items-center justify-between mb-4">
        <div>
          <h2 className="text-base font-bold text-slate-800">Team Load Overview</h2>
          <p className="text-xs text-slate-400 mt-0.5">Top {top5.length} teams by structural cognitive load</p>
        </div>
        <button onClick={() => navigate('/cognitive-load')}
          className="text-xs font-semibold text-indigo-600 bg-indigo-50 hover:bg-indigo-100 px-3 py-1.5 rounded-lg flex items-center gap-1 transition-colors">
          Full View <ArrowRight className="w-3 h-3" />
        </button>
      </div>
      <div className="space-y-1.5">
        {top5.map(team => {
          const lc = LEVEL_CLASS[team.overall_level] ?? LEVEL_CLASS.low
          const shortType = team.team.type.replace('complicated-subsystem', 'subsystem').replace('stream-aligned', 'stream')
          return (
            <div key={team.team.name} className="grid grid-cols-[1fr_80px_3rem_3rem_3rem_4rem] items-center gap-2 px-3.5 py-2.5 rounded-xl bg-slate-50">
              <div className="flex items-center gap-2 min-w-0">
                <span className={cn('w-2 h-2 rounded-full shrink-0', lc.dot)} />
                <span className="text-sm font-semibold text-slate-800 truncate">{team.team.name}</span>
              </div>
              <span className="text-[10px] font-semibold text-slate-500 uppercase tracking-wide text-center px-1.5 py-0.5 rounded bg-slate-200">
                {shortType}
              </span>
              <div className="text-center">
                <div className="text-sm font-bold text-slate-800">{team.capability_count}</div>
                <div className="text-[9px] text-slate-400 uppercase">Caps</div>
              </div>
              <div className="text-center">
                <div className="text-sm font-bold text-slate-800">{team.service_count}</div>
                <div className="text-[9px] text-slate-400 uppercase">Svcs</div>
              </div>
              <div className="text-center">
                <div className="text-sm font-bold text-slate-800">{team.dependency_count}</div>
                <div className="text-[9px] text-slate-400 uppercase">Deps</div>
              </div>
              <div className="text-right">
                <span className={cn('text-[11px] font-bold px-2.5 py-0.5 rounded-md', lc.badge, lc.text)}>
                  {team.overall_level.charAt(0).toUpperCase() + team.overall_level.slice(1)}
                </span>
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
