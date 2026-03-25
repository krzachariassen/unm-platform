import { useRequireModel } from '@/lib/model-context'
import { ModelRequired } from '@/components/ui/ModelRequired'
import { useNavigate } from 'react-router-dom'
import { useEffect, useState, useMemo } from 'react'
import {
  Users, Layers, Flag, Network, Activity, Map, GitBranch,
  AlertTriangle, CheckCircle2, ArrowRight, TrendingUp,
  Shield, Box, Link2, Zap,
} from 'lucide-react'
import { api, SignalsViewResponse, CognitiveLoadViewResponse, TeamLoad } from '@/lib/api'

const VIEW_CARDS = [
  { id: 'unm-map', label: 'UNM Map', icon: Map, desc: 'Full Actor → Need → Capability map', gradient: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)' },
  { id: 'need', label: 'Need View', icon: Users, desc: 'Actor → Need → Capability chains', gradient: 'linear-gradient(135deg, #3b82f6 0%, #2dd4bf 100%)' },
  { id: 'capability', label: 'Capability View', icon: Layers, desc: 'Hierarchy and dependencies', gradient: 'linear-gradient(135deg, #10b981 0%, #34d399 100%)' },
  { id: 'ownership', label: 'Ownership View', icon: Flag, desc: 'Team ownership + service matrix', gradient: 'linear-gradient(135deg, #f59e0b 0%, #fbbf24 100%)' },
  { id: 'team-topology', label: 'Team Topology', icon: Network, desc: 'Team interactions', gradient: 'linear-gradient(135deg, #ef4444 0%, #f97316 100%)' },
  { id: 'cognitive-load', label: 'Cognitive Load', icon: Activity, desc: 'Per-team load metrics', gradient: 'linear-gradient(135deg, #ec4899 0%, #a855f7 100%)' },
  { id: 'realization', label: 'Realization View', icon: GitBranch, desc: 'Service → capability mapping', gradient: 'linear-gradient(135deg, #0ea5e9 0%, #6366f1 100%)' },
]

const STAT_CONFIG: Array<{ key: string; label: string; icon: typeof Users; gradient: string; iconBg: string }> = [
  { key: 'actors', label: 'Actors', icon: Users, gradient: 'linear-gradient(135deg, #ede9fe 0%, #e0e7ff 100%)', iconBg: '#8b5cf6' },
  { key: 'needs', label: 'Needs', icon: Zap, gradient: 'linear-gradient(135deg, #dbeafe 0%, #e0f2fe 100%)', iconBg: '#3b82f6' },
  { key: 'capabilities', label: 'Capabilities', icon: Layers, gradient: 'linear-gradient(135deg, #d1fae5 0%, #ccfbf1 100%)', iconBg: '#10b981' },
  { key: 'services', label: 'Services', icon: Box, gradient: 'linear-gradient(135deg, #fef3c7 0%, #fef9c3 100%)', iconBg: '#f59e0b' },
  { key: 'teams', label: 'Teams', icon: Network, gradient: 'linear-gradient(135deg, #fce7f3 0%, #fce4ec 100%)', iconBg: '#ec4899' },
  { key: 'external_dependencies', label: 'External Deps', icon: Link2, gradient: 'linear-gradient(135deg, #cffafe 0%, #e0f2fe 100%)', iconBg: '#06b6d4' },
]

const HEALTH_CONFIG: Record<string, { color: string; bg: string; border: string; ring: string; label: string; score: number }> = {
  red:   { color: '#dc2626', bg: '#fef2f2', border: '#fecaca', ring: '#ef4444', label: 'Critical', score: 33 },
  amber: { color: '#d97706', bg: '#fffbeb', border: '#fde68a', ring: '#f59e0b', label: 'Warning', score: 66 },
  green: { color: '#16a34a', bg: '#f0fdf4', border: '#bbf7d0', ring: '#22c55e', label: 'Healthy', score: 100 },
}

const LEVEL_COLORS: Record<string, string> = { low: '#22c55e', medium: '#f59e0b', high: '#ef4444' }

function HealthRing({ value, color, size = 80, strokeWidth = 6 }: { value: number; color: string; size?: number; strokeWidth?: number }) {
  const radius = (size - strokeWidth) / 2
  const circumference = 2 * Math.PI * radius
  const offset = circumference - (value / 100) * circumference
  return (
    <svg width={size} height={size} style={{ transform: 'rotate(-90deg)' }}>
      <circle cx={size / 2} cy={size / 2} r={radius} fill="none" stroke="#f1f5f9" strokeWidth={strokeWidth} />
      <circle
        cx={size / 2} cy={size / 2} r={radius} fill="none" stroke={color} strokeWidth={strokeWidth}
        strokeDasharray={circumference} strokeDashoffset={offset} strokeLinecap="round"
        style={{ transition: 'stroke-dashoffset 1s ease-out' }}
      />
    </svg>
  )
}

function AnimatedNumber({ value }: { value: number }) {
  const [display, setDisplay] = useState(0)
  useEffect(() => {
    if (value === 0) { setDisplay(0); return }
    let frame: number
    const start = performance.now()
    const duration = 600
    const animate = (now: number) => {
      const progress = Math.min((now - start) / duration, 1)
      const eased = 1 - Math.pow(1 - progress, 3)
      setDisplay(Math.round(eased * value))
      if (progress < 1) frame = requestAnimationFrame(animate)
    }
    frame = requestAnimationFrame(animate)
    return () => cancelAnimationFrame(frame)
  }, [value])
  return <>{display}</>
}

function SignalBar({ count, max, color }: { count: number; max: number; color: string }) {
  const pct = max > 0 ? Math.min((count / max) * 100, 100) : 0
  return (
    <div style={{ height: 6, borderRadius: 3, background: '#f1f5f9', overflow: 'hidden', flex: 1 }}>
      <div style={{ height: '100%', borderRadius: 3, background: color, width: `${pct}%`, transition: 'width 0.8s ease-out' }} />
    </div>
  )
}

function relativeTime(date: Date): string {
  const mins = Math.floor((Date.now() - date.getTime()) / 60000)
  if (mins < 1) return 'just now'
  if (mins < 60) return `${mins} minute${mins === 1 ? '' : 's'} ago`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours} hour${hours === 1 ? '' : 's'} ago`
  return date.toLocaleDateString()
}

export function DashboardPage() {
  const { parseResult, modelId, loadedAt, isHydrating } = useRequireModel()
  const navigate = useNavigate()
  const [signals, setSignals] = useState<SignalsViewResponse | null>(null)
  const [signalsLoading, setSignalsLoading] = useState(false)
  const [teamLoads, setTeamLoads] = useState<TeamLoad[]>([])
  const [teamLoadsLoading, setTeamLoadsLoading] = useState(false)

  useEffect(() => {
    if (isHydrating || !modelId) return
    setSignalsLoading(true)
    api.getSignals(modelId)
      .then(data => setSignals(data))
      .catch(() => setSignals(null))
      .finally(() => setSignalsLoading(false))

    setTeamLoadsLoading(true)
    api.getView(modelId, 'cognitive-load')
      .then(data => setTeamLoads((data as unknown as CognitiveLoadViewResponse).team_loads ?? []))
      .catch(() => setTeamLoads([]))
      .finally(() => setTeamLoadsLoading(false))
  }, [isHydrating, modelId])

  const signalItems = useMemo(() => {
    if (!signals) return []
    const { user_experience_layer: ux, architecture_layer: arch, organization_layer: org } = signals
    const items: Array<{ label: string; count: number; color: string; icon: typeof AlertTriangle; filter: string }> = []
    if (ux.needs_at_risk.length > 0)
      items.push({ label: 'Needs at risk', count: ux.needs_at_risk.length, color: '#ef4444', icon: AlertTriangle, filter: 'needs-at-risk' })
    if (ux.needs_requiring_3plus_teams.length > 0)
      items.push({ label: 'Needs served by 3+ teams', count: ux.needs_requiring_3plus_teams.length, color: '#f59e0b', icon: Users, filter: 'needs-at-risk' })
    if (ux.needs_with_no_capability_backing.length > 0)
      items.push({ label: 'Unbacked needs', count: ux.needs_with_no_capability_backing.length, color: '#ef4444', icon: Zap, filter: 'gap' })
    if (arch.capabilities_fragmented_across_teams.length > 0)
      items.push({ label: 'Fragmented capabilities', count: arch.capabilities_fragmented_across_teams.length, color: '#f59e0b', icon: Layers, filter: 'fragmentation' })
    if (arch.capabilities_not_connected_to_any_need.length > 0)
      items.push({ label: 'Disconnected capabilities', count: arch.capabilities_not_connected_to_any_need.length, color: '#6366f1', icon: Link2, filter: 'gap' })
    if (arch.user_facing_caps_with_cross_team_services.length > 0)
      items.push({ label: 'Cross-team user-facing caps', count: arch.user_facing_caps_with_cross_team_services.length, color: '#f97316', icon: Flag, filter: 'fragmentation' })
    if (org.critical_bottleneck_services.length > 0)
      items.push({ label: 'Bottleneck services', count: org.critical_bottleneck_services.length, color: '#ef4444', icon: TrendingUp, filter: 'bottleneck' })
    if (org.top_teams_by_structural_load.length > 0)
      items.push({ label: 'High-load teams', count: org.top_teams_by_structural_load.length, color: '#ec4899', icon: Activity, filter: 'cognitive-load' })
    if (org.low_coherence_teams.length > 0)
      items.push({ label: 'Low-coherence teams', count: org.low_coherence_teams.length, color: '#8b5cf6', icon: Network, filter: 'cognitive-load' })
    return items
  }, [signals])

  const maxSignalCount = useMemo(() => Math.max(...signalItems.map(s => s.count), 1), [signalItems])

  const topLoadedTeams = useMemo(
    () => [...teamLoads].sort((a, b) => {
      const rank = (l: string) => l === 'high' ? 3 : l === 'medium' ? 2 : 1
      const scoreDiff = rank(b.overall_level) - rank(a.overall_level)
      return scoreDiff !== 0 ? scoreDiff : a.team.name.localeCompare(b.team.name)
    }).slice(0, 5),
    [teamLoads],
  )

  if (!parseResult || !modelId) return null

  const isLoading = signalsLoading || teamLoadsLoading
  const { summary, validation } = parseResult
  const hasErrors = validation.errors.length > 0
  const totalSignals = signalItems.reduce((sum, s) => sum + s.count, 0)

  return (
    <ModelRequired>
      <div style={{ maxWidth: 1200, margin: '0 auto' }}>
      {/* ─── Hero Header ─── */}
      <div style={{ marginBottom: 32 }}>
        <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 16 }}>
          <div style={{ flex: 1 }}>
            <h1
              style={{
                fontSize: 32, fontWeight: 800, letterSpacing: '-0.025em', lineHeight: 1.1, margin: 0,
                background: 'linear-gradient(135deg, #1e293b 0%, #475569 100%)',
                WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent',
              }}
            >
              {parseResult.system_name}
            </h1>
            <p style={{ fontSize: 15, color: '#64748b', marginTop: 6, lineHeight: 1.5 }}>
              {parseResult.system_description || 'Architecture model overview'}
              {loadedAt && (
                <span style={{ fontSize: 12, color: '#94a3b8', marginLeft: 12 }}>
                  Last loaded: {relativeTime(loadedAt)}
                </span>
              )}
            </p>
          </div>
          <div
            style={{
              display: 'flex', alignItems: 'center', gap: 8, padding: '8px 16px', borderRadius: 12,
              background: hasErrors ? '#fef2f2' : '#f0fdf4', border: `1px solid ${hasErrors ? '#fecaca' : '#bbf7d0'}`,
            }}
          >
            {hasErrors
              ? <AlertTriangle size={16} style={{ color: '#ef4444' }} />
              : <CheckCircle2 size={16} style={{ color: '#22c55e' }} />
            }
            <span style={{ fontSize: 13, fontWeight: 600, color: hasErrors ? '#dc2626' : '#16a34a' }}>
              {hasErrors ? `${validation.errors.length} Error${validation.errors.length === 1 ? '' : 's'}` : 'Model Valid'}
            </span>
            {validation.warnings.length > 0 && (
              <span style={{ fontSize: 12, color: '#d97706', fontWeight: 500, marginLeft: 4 }}>
                · {validation.warnings.length} warning{validation.warnings.length === 1 ? '' : 's'}
              </span>
            )}
          </div>
        </div>
      </div>

      {/* ─── Stats Grid ─── */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(160px, 1fr))', gap: 12, marginBottom: 28 }}>
        {STAT_CONFIG.map(({ key, label, icon: Icon, gradient, iconBg }) => {
          const val = (summary as Record<string, number>)[key]
          if (val == null) return null
          return (
            <div
              key={key}
              style={{
                padding: '20px 16px', borderRadius: 16, background: gradient,
                border: '1px solid rgba(0,0,0,0.04)', position: 'relative', overflow: 'hidden',
              }}
            >
              <div aria-hidden="true" style={{
                position: 'absolute', top: -8, right: -8, width: 56, height: 56, borderRadius: '50%',
                background: iconBg, opacity: 0.08,
              }} />
              <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                <div style={{
                  width: 36, height: 36, borderRadius: 10, display: 'flex', alignItems: 'center', justifyContent: 'center',
                  background: `${iconBg}18`,
                }}>
                  <Icon size={18} style={{ color: iconBg }} />
                </div>
                <div>
                  <div style={{ fontSize: 26, fontWeight: 800, color: '#1e293b', letterSpacing: '-0.02em', lineHeight: 1 }}>
                    <AnimatedNumber value={val} />
                  </div>
                  <div style={{ fontSize: 11, fontWeight: 600, color: '#64748b', marginTop: 2, textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                    {label}
                  </div>
                </div>
              </div>
            </div>
          )
        })}
      </div>

      {/* ─── Health & Signals Row ─── */}
      {!signalsLoading && signals && (
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16, marginBottom: 28 }}>
          {/* Health Overview Card */}
          <div style={{
            padding: 24, borderRadius: 20,
            background: 'linear-gradient(135deg, #ffffff 0%, #f8fafc 100%)',
            border: '1px solid #e2e8f0', boxShadow: '0 1px 3px rgba(0,0,0,0.04)',
          }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 20 }}>
              <div>
                <h2 style={{ fontSize: 16, fontWeight: 700, color: '#1e293b', margin: 0 }}>Platform Health</h2>
                <p style={{ fontSize: 12, color: '#94a3b8', marginTop: 2 }}>Across three architecture layers</p>
              </div>
              <Shield size={20} style={{ color: '#94a3b8' }} />
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-around', gap: 12 }}>
              {([
                { key: 'ux_risk' as const, label: 'UX', sublabel: 'User Experience', tooltip: `Based on ${signals.user_experience_layer.needs_at_risk.length} needs at risk, ${signals.user_experience_layer.needs_with_no_capability_backing.length} unbacked needs` },
                { key: 'architecture_risk' as const, label: 'Arch', sublabel: 'Architecture', tooltip: `Based on ${signals.architecture_layer.capabilities_fragmented_across_teams.length} fragmented capabilities, ${signals.architecture_layer.capabilities_not_connected_to_any_need.length} disconnected capabilities` },
                { key: 'org_risk' as const, label: 'Org', sublabel: 'Organization', tooltip: `Based on ${signals.organization_layer.critical_bottleneck_services.length} bottleneck services, ${signals.organization_layer.top_teams_by_structural_load.length} high-load teams` },
              ]).map(({ key, label, sublabel, tooltip }) => {
                const level = signals.health[key]
                const cfg = HEALTH_CONFIG[level]
                return (
                  <div key={key} style={{ textAlign: 'center' }} title={tooltip}>
                    <div style={{ position: 'relative', display: 'inline-block' }}>
                      <HealthRing value={cfg.score} color={cfg.ring} />
                      <div style={{
                        position: 'absolute', inset: 0, display: 'flex', flexDirection: 'column',
                        alignItems: 'center', justifyContent: 'center', transform: 'rotate(0deg)',
                      }}>
                        <span style={{ fontSize: 16, fontWeight: 800, color: cfg.color }}>{label}</span>
                      </div>
                    </div>
                    <div style={{ marginTop: 8 }}>
                      <div style={{
                        fontSize: 11, fontWeight: 700, color: cfg.color,
                        padding: '2px 10px', borderRadius: 6, background: cfg.bg, display: 'inline-block',
                      }}>
                        {cfg.label}
                      </div>
                      <div style={{ fontSize: 10, color: '#94a3b8', marginTop: 4 }}>{sublabel}</div>
                    </div>
                  </div>
                )
              })}
            </div>
          </div>

          {/* Signals Breakdown Card */}
          <div style={{
            padding: 24, borderRadius: 20,
            background: 'linear-gradient(135deg, #ffffff 0%, #f8fafc 100%)',
            border: '1px solid #e2e8f0', boxShadow: '0 1px 3px rgba(0,0,0,0.04)',
          }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 16 }}>
              <div>
                <h2
                  onClick={() => navigate('/signals')}
                  style={{ fontSize: 16, fontWeight: 700, color: '#1e293b', margin: 0, cursor: 'pointer' }}
                  onMouseEnter={e => { e.currentTarget.style.color = '#3b82f6' }}
                  onMouseLeave={e => { e.currentTarget.style.color = '#1e293b' }}
                >Architecture Signals</h2>
                <p style={{ fontSize: 12, color: '#94a3b8', marginTop: 2 }}>
                  {totalSignals > 0 ? `${totalSignals} finding${totalSignals === 1 ? '' : 's'} detected` : 'No issues detected'}
                </p>
              </div>
              <button
                onClick={() => navigate('/signals')}
                style={{
                  fontSize: 12, fontWeight: 600, color: '#6366f1', background: '#eef2ff',
                  border: 'none', borderRadius: 8, padding: '6px 12px', cursor: 'pointer',
                  display: 'flex', alignItems: 'center', gap: 4,
                }}
                onMouseEnter={e => { e.currentTarget.style.background = '#e0e7ff' }}
                onMouseLeave={e => { e.currentTarget.style.background = '#eef2ff' }}
              >
                Details <ArrowRight size={12} />
              </button>
            </div>
            {signalItems.length > 0 ? (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                {signalItems.map((item, i) => {
                  const Icon = item.icon
                  return (
                    <div
                      key={i}
                      onClick={() => navigate(`/signals?filter=${item.filter}`)}
                      style={{ display: 'flex', alignItems: 'center', gap: 10, cursor: 'pointer', borderRadius: 6, padding: '2px 4px' }}
                      onMouseEnter={e => { e.currentTarget.style.background = '#f8fafc' }}
                      onMouseLeave={e => { e.currentTarget.style.background = 'transparent' }}
                    >
                      <Icon size={14} style={{ color: item.color, flexShrink: 0 }} />
                      <span style={{ fontSize: 12, color: '#475569', flex: '0 0 auto', minWidth: 160 }}>{item.label}</span>
                      <SignalBar count={item.count} max={maxSignalCount} color={item.color} />
                      <span style={{
                        fontSize: 12, fontWeight: 700, color: item.color, minWidth: 20, textAlign: 'right',
                        display: 'flex', alignItems: 'center', gap: 4,
                      }}>
                        {item.count}
                        <span style={{ color: '#9ca3af', fontSize: 12 }}>&rarr;</span>
                      </span>
                    </div>
                  )
                })}
              </div>
            ) : (
              <div style={{
                display: 'flex', alignItems: 'center', gap: 10, padding: 16, borderRadius: 12,
                background: '#f0fdf4', border: '1px solid #bbf7d0',
              }}>
                <CheckCircle2 size={20} style={{ color: '#22c55e' }} />
                <div>
                  <div style={{ fontSize: 14, fontWeight: 600, color: '#15803d' }}>All clear</div>
                  <div style={{ fontSize: 12, color: '#16a34a' }}>No significant architecture signals detected</div>
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Loading state for health */}
      {isLoading && (
        <div style={{
          display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 10, padding: 40,
          borderRadius: 20, background: '#f8fafc', border: '1px solid #e2e8f0', marginBottom: 28,
        }}>
          <span style={{
            width: 16, height: 16, border: '2px solid #e2e8f0', borderTopColor: '#6366f1',
            borderRadius: '50%', animation: 'spin 0.8s linear infinite', display: 'inline-block',
          }} />
          <span style={{ fontSize: 13, color: '#94a3b8' }}>Loading health analysis…</span>
        </div>
      )}

      {/* ─── Team Cognitive Load (top 5) ─── */}
      {topLoadedTeams.length > 0 && (
        <div style={{
          padding: 24, borderRadius: 20, marginBottom: 28,
          background: 'linear-gradient(135deg, #ffffff 0%, #f8fafc 100%)',
          border: '1px solid #e2e8f0', boxShadow: '0 1px 3px rgba(0,0,0,0.04)',
        }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 16 }}>
            <div>
              <h2 style={{ fontSize: 16, fontWeight: 700, color: '#1e293b', margin: 0 }}>Team Load Overview</h2>
              <p style={{ fontSize: 12, color: '#94a3b8', marginTop: 2 }}>
                Top {topLoadedTeams.length} teams by structural cognitive load
              </p>
            </div>
            <button
              onClick={() => navigate('/cognitive-load')}
              style={{
                fontSize: 12, fontWeight: 600, color: '#6366f1', background: '#eef2ff',
                border: 'none', borderRadius: 8, padding: '6px 12px', cursor: 'pointer',
                display: 'flex', alignItems: 'center', gap: 4,
              }}
              onMouseEnter={e => { e.currentTarget.style.background = '#e0e7ff' }}
              onMouseLeave={e => { e.currentTarget.style.background = '#eef2ff' }}
            >
              Full View <ArrowRight size={12} />
            </button>
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
            {topLoadedTeams.map(team => {
              const levelColor = LEVEL_COLORS[team.overall_level] ?? '#94a3b8'
              return (
                <div key={team.team.name} style={{
                  display: 'grid', gridTemplateColumns: '1fr 80px 64px 64px 64px 80px', alignItems: 'center', gap: 8,
                  padding: '10px 14px', borderRadius: 12, background: '#f8fafc',
                }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                    <span style={{
                      width: 8, height: 8, borderRadius: '50%', background: levelColor, flexShrink: 0,
                    }} />
                    <span style={{ fontSize: 13, fontWeight: 600, color: '#1e293b', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                      {team.team.name}
                    </span>
                  </div>
                  <span style={{
                    fontSize: 10, fontWeight: 600, color: '#64748b', textTransform: 'uppercase',
                    letterSpacing: '0.03em', textAlign: 'center', padding: '2px 6px', borderRadius: 4,
                    background: '#e2e8f0',
                  }}>
                    {team.team.type.replace('complicated-subsystem', 'subsystem').replace('stream-aligned', 'stream')}
                  </span>
                  <div style={{ textAlign: 'center' }}>
                    <div style={{ fontSize: 14, fontWeight: 700, color: '#1e293b' }}>{team.capability_count}</div>
                    <div style={{ fontSize: 9, color: '#94a3b8', textTransform: 'uppercase' }}>Caps</div>
                  </div>
                  <div style={{ textAlign: 'center' }}>
                    <div style={{ fontSize: 14, fontWeight: 700, color: '#1e293b' }}>{team.service_count}</div>
                    <div style={{ fontSize: 9, color: '#94a3b8', textTransform: 'uppercase' }}>Svcs</div>
                  </div>
                  <div style={{ textAlign: 'center' }}>
                    <div style={{ fontSize: 14, fontWeight: 700, color: '#1e293b' }}>{team.dependency_count}</div>
                    <div style={{ fontSize: 9, color: '#94a3b8', textTransform: 'uppercase' }}>Deps</div>
                  </div>
                  <div style={{ textAlign: 'right' }}>
                    <span style={{
                      fontSize: 11, fontWeight: 700, color: levelColor,
                      padding: '3px 10px', borderRadius: 6,
                      background: team.overall_level === 'high' ? '#fef2f2' : team.overall_level === 'medium' ? '#fffbeb' : '#f0fdf4',
                    }}>
                      {team.overall_level.charAt(0).toUpperCase() + team.overall_level.slice(1)}
                    </span>
                  </div>
                </div>
              )
            })}
          </div>
        </div>
      )}

      {/* ─── Validation Details (only if issues) ─── */}
      {(hasErrors || validation.warnings.length > 0) && (
        <div style={{
          padding: 24, borderRadius: 20, marginBottom: 28,
          background: 'linear-gradient(135deg, #ffffff 0%, #f8fafc 100%)',
          border: '1px solid #e2e8f0', boxShadow: '0 1px 3px rgba(0,0,0,0.04)',
        }}>
          <h2 style={{ fontSize: 16, fontWeight: 700, color: '#1e293b', margin: '0 0 12px 0' }}>Validation Issues</h2>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
            {hasErrors && (
              <div style={{ padding: 16, borderRadius: 14, background: '#fef2f2', border: '1px solid #fecaca' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 8 }}>
                  <AlertTriangle size={14} style={{ color: '#ef4444' }} />
                  <span style={{ fontSize: 13, fontWeight: 700, color: '#dc2626' }}>
                    {validation.errors.length} Error{validation.errors.length === 1 ? '' : 's'}
                  </span>
                </div>
                {validation.errors.slice(0, 4).map((e, i) => (
                  <p key={i} style={{ fontSize: 12, color: '#991b1b', margin: '4px 0 0 20px', lineHeight: 1.4 }}>{e.message}</p>
                ))}
              </div>
            )}
            {validation.warnings.length > 0 && (
              <div style={{ padding: 16, borderRadius: 14, background: '#fffbeb', border: '1px solid #fde68a' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 8 }}>
                  <AlertTriangle size={14} style={{ color: '#f59e0b' }} />
                  <span style={{ fontSize: 13, fontWeight: 700, color: '#92400e' }}>
                    {validation.warnings.length} Warning{validation.warnings.length === 1 ? '' : 's'}
                  </span>
                </div>
                {validation.warnings.slice(0, 4).map((w, i) => (
                  <p key={i} style={{ fontSize: 12, color: '#78350f', margin: '4px 0 0 20px', lineHeight: 1.4, display: 'flex', alignItems: 'center', gap: 4 }}>
                    <span style={{ flexShrink: 0 }}>&#9888;</span>
                    <span style={{ flex: 1 }}>{w.message}</span>
                    {w.message.toLowerCase().includes('cognitive load') && (
                      <button onClick={() => navigate('/cognitive-load')}
                        style={{color:'#3b82f6', background:'none', border:'none', cursor:'pointer', fontSize:'inherit', flexShrink: 0}}>
                        View &rarr;
                      </button>
                    )}
                  </p>
                ))}
              </div>
            )}
          </div>
        </div>
      )}

      {/* ─── Explore Views ─── */}
      <div style={{ marginBottom: 20 }}>
        <h2 style={{ fontSize: 16, fontWeight: 700, color: '#1e293b', margin: '0 0 16px 0' }}>Explore Views</h2>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(240px, 1fr))', gap: 12 }}>
          {VIEW_CARDS.map(({ id, label, icon: Icon, desc, gradient }) => (
            <button
              key={id}
              onClick={() => navigate(`/${id}`)}
              style={{
                padding: 20, borderRadius: 16, border: '1px solid #e2e8f0', background: '#ffffff',
                textAlign: 'left', cursor: 'pointer', position: 'relative', overflow: 'hidden',
                transition: 'all 0.2s ease', boxShadow: '0 1px 3px rgba(0,0,0,0.02)',
              }}
              onMouseEnter={e => {
                const el = e.currentTarget
                el.style.transform = 'translateY(-2px)'
                el.style.boxShadow = '0 8px 25px rgba(0,0,0,0.08)'
                el.style.borderColor = '#c7d2fe'
              }}
              onMouseLeave={e => {
                const el = e.currentTarget
                el.style.transform = 'translateY(0)'
                el.style.boxShadow = '0 1px 3px rgba(0,0,0,0.02)'
                el.style.borderColor = '#e2e8f0'
              }}
            >
              <div style={{
                position: 'absolute', top: 0, left: 0, right: 0, height: 3, background: gradient, borderRadius: '16px 16px 0 0',
              }} />
              <div style={{ display: 'flex', alignItems: 'flex-start', gap: 12, marginTop: 2 }}>
                <div style={{
                  width: 40, height: 40, borderRadius: 12, display: 'flex', alignItems: 'center', justifyContent: 'center',
                  background: gradient, flexShrink: 0,
                }}>
                  <Icon size={18} style={{ color: '#ffffff' }} />
                </div>
                <div>
                  <div style={{ fontSize: 14, fontWeight: 700, color: '#1e293b', marginBottom: 3 }}>{label}</div>
                  <div style={{ fontSize: 12, color: '#94a3b8', lineHeight: 1.4 }}>{desc}</div>
                </div>
              </div>
              <div style={{
                position: 'absolute', bottom: 12, right: 14, opacity: 0.3,
              }}>
                <ArrowRight size={14} style={{ color: '#94a3b8' }} />
              </div>
            </button>
          ))}
        </div>
      </div>

      {/* CSS animation for spinner */}
      <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
      </div>
    </ModelRequired>
  )
}
