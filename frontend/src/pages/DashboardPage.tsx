import { useEffect, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import {
  Users, Layers, Network, Activity, Map, GitBranch,
  AlertTriangle, CheckCircle2, ArrowRight, Box, Link2, Zap, Flag,
} from 'lucide-react'
import { ModelRequired } from '@/components/ui/ModelRequired'
import { StatCard } from '@/components/ui/stat-card'
import { useModel } from '@/lib/model-context'
import { viewsApi } from '@/services/api'
import { HealthCard } from '@/features/dashboard/HealthCard'
import { SignalsCard } from '@/features/dashboard/SignalsCard'
import { TeamLoadCard } from '@/features/dashboard/TeamLoadCard'
import { cn } from '@/lib/utils'

const VIEW_CARDS = [
  { id: 'unm-map',        label: 'UNM Map',         icon: Map,       desc: 'Full Actor → Need → Capability map', color: 'from-indigo-500 to-purple-600' },
  { id: 'need',           label: 'Need View',        icon: Users,     desc: 'Actor → Need → Capability chains',   color: 'from-blue-500 to-teal-400' },
  { id: 'capability',     label: 'Capability View',  icon: Layers,    desc: 'Hierarchy and dependencies',         color: 'from-emerald-500 to-green-400' },
  { id: 'ownership',      label: 'Ownership View',   icon: Flag,      desc: 'Team ownership + service matrix',    color: 'from-amber-400 to-yellow-300' },
  { id: 'team-topology',  label: 'Team Topology',    icon: Network,   desc: 'Team interactions',                  color: 'from-red-500 to-orange-400' },
  { id: 'cognitive-load', label: 'Cognitive Load',   icon: Activity,  desc: 'Per-team load metrics',              color: 'from-pink-500 to-purple-400' },
  { id: 'realization',    label: 'Realization View', icon: GitBranch, desc: 'Service → capability mapping',       color: 'from-sky-500 to-indigo-400' },
]

const STAT_KEYS: Array<{ key: string; label: string; icon: typeof Users; iconClass: string }> = [
  { key: 'actors',               label: 'Actors',       icon: Users,   iconClass: 'text-purple-500' },
  { key: 'needs',                label: 'Needs',        icon: Zap,     iconClass: 'text-blue-500' },
  { key: 'capabilities',         label: 'Capabilities', icon: Layers,  iconClass: 'text-emerald-500' },
  { key: 'services',             label: 'Services',     icon: Box,     iconClass: 'text-amber-500' },
  { key: 'teams',                label: 'Teams',        icon: Network, iconClass: 'text-pink-500' },
  { key: 'external_dependencies',label: 'Ext Deps',     icon: Link2,   iconClass: 'text-cyan-500' },
]

function AnimatedNumber({ value }: { value: number }) {
  const [display, setDisplay] = useState(0)
  useEffect(() => {
    if (!value) { setDisplay(0); return }
    let frame: number
    const start = performance.now()
    const animate = (now: number) => {
      const t = Math.min((now - start) / 600, 1)
      setDisplay(Math.round((1 - Math.pow(1 - t, 3)) * value))
      if (t < 1) frame = requestAnimationFrame(animate)
    }
    frame = requestAnimationFrame(animate)
    return () => cancelAnimationFrame(frame)
  }, [value])
  return <>{display}</>
}

function relativeTime(date: Date): string {
  const mins = Math.floor((Date.now() - date.getTime()) / 60000)
  if (mins < 1) return 'just now'
  if (mins < 60) return `${mins}m ago`
  const h = Math.floor(mins / 60)
  if (h < 24) return `${h}h ago`
  return date.toLocaleDateString()
}

export function DashboardPage() {
  const { modelId, parseResult, loadedAt } = useModel()
  const navigate = useNavigate()

  const { data: signals, isLoading: signalsLoading } = useQuery({
    queryKey: ['signalsView', modelId],
    queryFn: () => viewsApi.getSignalsView(modelId!),
    enabled: !!modelId,
  })
  const { data: cogLoad } = useQuery({
    queryKey: ['cognitiveLoadView', modelId],
    queryFn: () => viewsApi.getCognitiveLoadView(modelId!),
    enabled: !!modelId,
  })

  if (!parseResult || !modelId) return null
  const { summary, validation } = parseResult
  const hasErrors = validation.errors.length > 0

  return (
    <ModelRequired>
      <div className="space-y-6 max-w-6xl">
        {/* Hero */}
        <div className="flex items-start justify-between gap-4">
          <div>
            <h1 className="text-3xl font-extrabold tracking-tight text-slate-800 leading-tight">{parseResult.system_name}</h1>
            <p className="text-sm text-slate-500 mt-1.5 leading-relaxed">
              {parseResult.system_description || 'Architecture model overview'}
              {loadedAt && <span className="text-slate-400 ml-3 text-xs">Loaded {relativeTime(loadedAt)}</span>}
            </p>
          </div>
          <div className={cn('flex items-center gap-2 px-3 py-1.5 rounded-xl border text-sm font-semibold shrink-0',
            hasErrors ? 'bg-red-50 border-red-200 text-red-700' : 'bg-green-50 border-green-200 text-green-700')}>
            {hasErrors ? <AlertTriangle className="w-4 h-4" /> : <CheckCircle2 className="w-4 h-4" />}
            {hasErrors ? `${validation.errors.length} error${validation.errors.length === 1 ? '' : 's'}` : 'Model valid'}
            {validation.warnings.length > 0 && (
              <span className="text-amber-600 font-medium ml-1">· {validation.warnings.length} warning{validation.warnings.length === 1 ? '' : 's'}</span>
            )}
          </div>
        </div>

        {/* Stats */}
        <div className="grid grid-cols-3 sm:grid-cols-6 gap-3">
          {STAT_KEYS.map(({ key, label, icon: Icon, iconClass }) => {
            const val = (summary as Record<string, number>)[key]
            if (val == null) return null
            return (
              <StatCard key={key} label={label} value={<AnimatedNumber value={val} />} icon={<Icon className={cn('w-4 h-4', iconClass)} />} />
            )
          })}
        </div>

        {/* Health + Signals */}
        {signalsLoading && (
          <div className="flex items-center gap-3 p-8 rounded-2xl bg-slate-50 border border-slate-200">
            <span className="w-4 h-4 border-2 border-slate-200 border-t-indigo-500 rounded-full animate-spin" />
            <span className="text-sm text-slate-400">Loading health analysis…</span>
          </div>
        )}
        {signals && (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            <HealthCard signals={signals} />
            <SignalsCard signals={signals} />
          </div>
        )}

        {/* Team Load */}
        {cogLoad && <TeamLoadCard teams={cogLoad.team_loads} />}

        {/* Validation issues */}
        {(hasErrors || validation.warnings.length > 0) && (
          <div className="rounded-2xl border border-slate-200 bg-white p-6">
            <h2 className="text-base font-bold text-slate-800 mb-3">Validation Issues</h2>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              {hasErrors && (
                <div className="rounded-xl bg-red-50 border border-red-200 p-4">
                  <div className="flex items-center gap-1.5 mb-2">
                    <AlertTriangle className="w-3.5 h-3.5 text-red-500" />
                    <span className="text-sm font-bold text-red-700">{validation.errors.length} Error{validation.errors.length === 1 ? '' : 's'}</span>
                  </div>
                  {validation.errors.slice(0, 4).map((e, i) => (
                    <p key={i} className="text-xs text-red-800 mt-1 ml-5 leading-relaxed">{e.message}</p>
                  ))}
                </div>
              )}
              {validation.warnings.length > 0 && (
                <div className="rounded-xl bg-amber-50 border border-amber-200 p-4">
                  <div className="flex items-center gap-1.5 mb-2">
                    <AlertTriangle className="w-3.5 h-3.5 text-amber-500" />
                    <span className="text-sm font-bold text-amber-800">{validation.warnings.length} Warning{validation.warnings.length === 1 ? '' : 's'}</span>
                  </div>
                  {validation.warnings.slice(0, 4).map((w, i) => (
                    <p key={i} className="text-xs text-amber-900 mt-1 ml-5 leading-relaxed">{w.message}</p>
                  ))}
                </div>
              )}
            </div>
          </div>
        )}

        {/* Explore Views */}
        <div>
          <h2 className="text-base font-bold text-slate-800 mb-3">Explore Views</h2>
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-3">
            {VIEW_CARDS.map(({ id, label, icon: Icon, desc, color }) => (
              <button key={id} onClick={() => navigate(`/${id}`)}
                className="group p-4 rounded-xl border border-slate-200 bg-white text-left hover:-translate-y-0.5 hover:shadow-md hover:border-indigo-200 transition-all duration-200 relative overflow-hidden">
                <div className={cn('absolute top-0 left-0 right-0 h-0.5 bg-gradient-to-r', color)} />
                <div className="flex items-start gap-3 mt-1">
                  <div className={cn('w-9 h-9 rounded-xl bg-gradient-to-br flex items-center justify-center shrink-0', color)}>
                    <Icon className="w-4 h-4 text-white" />
                  </div>
                  <div className="min-w-0">
                    <div className="text-sm font-bold text-slate-800 mb-0.5">{label}</div>
                    <div className="text-xs text-slate-400 leading-snug">{desc}</div>
                  </div>
                </div>
                <ArrowRight className="absolute bottom-3 right-3 w-3.5 h-3.5 text-slate-300 group-hover:text-slate-400 transition-colors" />
              </button>
            ))}
          </div>
        </div>
      </div>
    </ModelRequired>
  )
}
