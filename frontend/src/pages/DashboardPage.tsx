import { useEffect, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import {
  Users, Layers, Network, Activity, Map, GitBranch,
  AlertTriangle, CheckCircle2, Box, Link2, Zap, Flag,
} from 'lucide-react'
import { ModelRequired } from '@/components/ui/ModelRequired'
import { ContentContainer } from '@/components/ui/content-container'
import { StatCard } from '@/components/ui/stat-card'
import { useModel } from '@/lib/model-context'
import { viewsApi } from '@/services/api'
import { HealthCard } from '@/features/dashboard/HealthCard'
import { SignalsCard } from '@/features/dashboard/SignalsCard'
import { TeamLoadCard } from '@/features/dashboard/TeamLoadCard'
import { cn } from '@/lib/utils'

const VIEW_CARDS = [
  { id: 'unm-map',        label: 'UNM Map',         icon: Map,       desc: 'Full Actor → Need → Capability map' },
  { id: 'need',           label: 'Need View',        icon: Users,     desc: 'Actor → Need → Capability chains' },
  { id: 'capability',     label: 'Capability View',  icon: Layers,    desc: 'Hierarchy and dependencies' },
  { id: 'ownership',      label: 'Ownership View',   icon: Flag,      desc: 'Team ownership + service matrix' },
  { id: 'team-topology',  label: 'Team Topology',    icon: Network,   desc: 'Team interactions' },
  { id: 'cognitive-load', label: 'Cognitive Load',   icon: Activity,  desc: 'Per-team load metrics' },
  { id: 'realization',    label: 'Realization View', icon: GitBranch, desc: 'Service → capability mapping' },
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
      <ContentContainer className="space-y-4">
        {/* Hero */}
        <div className="flex items-start justify-between gap-4">
          <div>
            <h1 className="text-xl font-bold text-foreground">{parseResult.system_name}</h1>
            <p className="text-xs text-muted-foreground mt-1 line-clamp-2 max-w-2xl" title={parseResult.system_description}>
              {parseResult.system_description || 'Architecture model overview'}
            </p>
            {loadedAt && <p className="text-[10px] text-muted-foreground/60 mt-0.5">Loaded {relativeTime(loadedAt)}</p>}
          </div>
          <div className={cn('flex items-center gap-1.5 px-2.5 py-1 rounded-md border text-xs font-medium shrink-0',
            hasErrors ? 'bg-red-50 border-red-200 text-red-700' : 'bg-green-50 border-green-200 text-green-700')}>
            {hasErrors ? <AlertTriangle className="w-3.5 h-3.5" /> : <CheckCircle2 className="w-3.5 h-3.5" />}
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
            return <StatCard key={key} label={label} value={<AnimatedNumber value={val} />} icon={<Icon className={cn('w-4 h-4', iconClass)} />} />
          })}
        </div>

        {/* Health + Signals */}
        {signalsLoading && (
          <div className="flex items-center gap-3 p-6 rounded-lg bg-muted border border-border">
            <span className="w-4 h-4 border-2 border-border border-t-foreground rounded-full animate-spin" />
            <span className="text-sm text-muted-foreground">Loading health analysis…</span>
          </div>
        )}
        {signals && (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-3">
            <HealthCard signals={signals} />
            <SignalsCard signals={signals} />
          </div>
        )}

        {/* Team Load */}
        {cogLoad && <TeamLoadCard teams={cogLoad.team_loads} />}

        {/* Validation issues */}
        {(hasErrors || validation.warnings.length > 0) && (
          <div className="rounded-lg border border-border bg-card p-4">
            <h2 className="text-sm font-semibold text-foreground mb-2">Validation Issues</h2>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              {hasErrors && (
                <div className="rounded-md bg-red-50 border border-red-200 p-3">
                  <div className="flex items-center gap-1.5 mb-1.5">
                    <AlertTriangle className="w-3 h-3 text-red-500" />
                    <span className="text-xs font-semibold text-red-700">{validation.errors.length} Error{validation.errors.length === 1 ? '' : 's'}</span>
                  </div>
                  {validation.errors.slice(0, 4).map((e, i) => (
                    <p key={i} className="text-xs text-red-800 mt-0.5 ml-4">{e.message}</p>
                  ))}
                </div>
              )}
              {validation.warnings.length > 0 && (
                <div className="rounded-md bg-amber-50 border border-amber-200 p-3">
                  <div className="flex items-center gap-1.5 mb-1.5">
                    <AlertTriangle className="w-3 h-3 text-amber-500" />
                    <span className="text-xs font-semibold text-amber-800">{validation.warnings.length} Warning{validation.warnings.length === 1 ? '' : 's'}</span>
                  </div>
                  {validation.warnings.slice(0, 4).map((w, i) => (
                    <p key={i} className="text-xs text-amber-900 mt-0.5 ml-4">{w.message}</p>
                  ))}
                </div>
              )}
            </div>
          </div>
        )}

        {/* Explore Views */}
        <div>
          <h2 className="text-sm font-semibold text-foreground mb-2">Explore Views</h2>
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-3">
            {VIEW_CARDS.map(({ id, label, icon: Icon, desc }) => (
              <button key={id} onClick={() => navigate(`/${id}`)}
                className="group p-3 rounded-lg border border-border bg-card text-left hover:bg-muted/50 hover:border-foreground/20 transition-colors">
                <div className="flex items-start gap-2.5">
                  <Icon className="w-4 h-4 text-muted-foreground mt-0.5 shrink-0" />
                  <div className="min-w-0">
                    <div className="text-xs font-semibold text-foreground">{label}</div>
                    <div className="text-[10px] text-muted-foreground mt-0.5 leading-snug">{desc}</div>
                  </div>
                </div>
              </button>
            ))}
          </div>
        </div>
      </ContentContainer>
    </ModelRequired>
  )
}
