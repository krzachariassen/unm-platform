import { useEffect, useState, useCallback, useRef } from 'react'
import { useSearchParams } from 'react-router-dom'
import { api } from '@/lib/api'
import { useRequireModel } from '@/lib/model-context'
import { useAIEnabled } from '@/hooks/useAIEnabled'
import type { SignalsViewResponse, SignalsNeedRisk, SignalsCapItem, SignalsTeamItem, SignalsServiceItem } from '@/lib/api'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import {
  AlertTriangle, CheckCircle, Users, Layers, Building2,
  Zap, Link2Off, GitMerge, Server, TrendingDown, ChevronDown, ChevronUp,
  Info, Lightbulb, Link2, Sparkles, X,
} from 'lucide-react'

// ── Local types for new backend fields ─────────────────────────────────────────

interface ExtDepItem {
  dep_name: string
  service_count: number
  services: string[]
  is_critical: boolean
  is_warning: boolean
}

interface OrgLayerWithExtDeps {
  top_teams_by_structural_load: import('@/lib/api').SignalsTeamItem[]
  critical_bottleneck_services: import('@/lib/api').SignalsServiceItem[]
  low_coherence_teams: import('@/lib/api').SignalsTeamItem[]
  critical_external_deps?: ExtDepItem[]
}

// ── Risk styling ────────────────────────────────────────────────────────────────

const RISK = {
  red: {
    bg: '#fef2f2',
    text: '#b91c1c',
    dot: '#ef4444',
    border: '#fca5a5',
    strip: '#ef4444',
    stripGradient: 'linear-gradient(90deg, #ef4444 0%, #f87171 100%)',
    label: 'High Risk',
    badgeBg: '#fee2e2',
    cardGradient: 'linear-gradient(135deg, #fecaca 0%, #fef2f2 55%, #ffffff 100%)',
    iconWrap: 'linear-gradient(135deg, #fee2e2 0%, #fecaca 100%)',
  },
  amber: {
    bg: '#fffbeb',
    text: '#b45309',
    dot: '#f59e0b',
    border: '#fcd34d',
    strip: '#f59e0b',
    stripGradient: 'linear-gradient(90deg, #f59e0b 0%, #fbbf24 100%)',
    label: 'Elevated',
    badgeBg: '#fef3c7',
    cardGradient: 'linear-gradient(135deg, #fde68a 0%, #fffbeb 50%, #ffffff 100%)',
    iconWrap: 'linear-gradient(135deg, #fef3c7 0%, #fde68a 100%)',
  },
  green: {
    bg: '#f0fdf4',
    text: '#15803d',
    dot: '#22c55e',
    border: '#86efac',
    strip: '#22c55e',
    stripGradient: 'linear-gradient(90deg, #22c55e 0%, #4ade80 100%)',
    label: 'Healthy',
    badgeBg: '#dcfce7',
    cardGradient: 'linear-gradient(135deg, #bbf7d0 0%, #f0fdf4 55%, #ffffff 100%)',
    iconWrap: 'linear-gradient(135deg, #dcfce7 0%, #bbf7d0 100%)',
  },
}
function rs(risk: string) { return RISK[risk as keyof typeof RISK] ?? RISK.green }

const gradientTitle: React.CSSProperties = {
  fontSize: 30,
  fontWeight: 800,
  letterSpacing: '-0.025em',
  background: 'linear-gradient(135deg, #1e293b 0%, #475569 100%)',
  WebkitBackgroundClip: 'text',
  WebkitTextFillColor: 'transparent',
  backgroundClip: 'text',
}

const cardShell: React.CSSProperties = {
  borderRadius: 20,
  background: 'linear-gradient(135deg, #ffffff 0%, #f8fafc 100%)',
  border: '1px solid #e2e8f0',
  boxShadow: '0 1px 3px rgba(0,0,0,0.04)',
}

const tagBase: React.CSSProperties = {
  borderRadius: 8,
  padding: '4px 10px',
  fontSize: 11,
  fontWeight: 600,
}

// ── Atoms ───────────────────────────────────────────────────────────────────────

function RiskPill({ risk }: { risk: string }) {
  const s = rs(risk)
  return (
    <span className="inline-flex items-center gap-1.5" style={{ ...tagBase, background: s.badgeBg, color: s.text }}>
      <span className="w-1.5 h-1.5 rounded-full flex-shrink-0" style={{ background: s.dot }} />
      {s.label}
    </span>
  )
}

function CountBadge({ n }: { n: number }) {
  return (
    <span className="ml-2 inline-flex items-center justify-center min-w-[22px] h-[22px] px-1.5 rounded-lg text-[11px] font-bold" style={{ background: '#f1f5f9', color: '#64748b', border: '1px solid #e2e8f0' }}>
      {n}
    </span>
  )
}

function Tag({ text, color }: { text: string; color?: 'amber' | 'red' }) {
  const styles = {
    amber: { background: '#fef3c7', color: '#92400e' },
    red:   { background: '#fee2e2', color: '#b91c1c' },
    default: { background: '#f1f5f9', color: '#64748b' },
  }
  const style = color ? styles[color] : styles.default
  return (
    <span className="inline-flex items-center" style={{ ...tagBase, ...style }}>
      {text}
    </span>
  )
}

// ── Expandable row wrapper ───────────────────────────────────────────────────────

function ExpandableRow({ summary, explanation, suggestion }: {
  summary: React.ReactNode
  explanation: string
  suggestion: string
}) {
  const [open, setOpen] = useState(false)
  const [hover, setHover] = useState(false)
  return (
    <div
      className="overflow-hidden transition-all duration-200 ease-out"
      style={{
        borderRadius: 16,
        border: '1px solid #f1f5f9',
        boxShadow: hover && !open ? '0 4px 12px rgba(15, 23, 42, 0.06)' : '0 1px 2px rgba(0,0,0,0.03)',
        transform: hover && !open ? 'translateY(-1px)' : 'translateY(0)',
      }}
    >
      <button
        type="button"
        onClick={() => setOpen(o => !o)}
        className="w-full flex items-center justify-between gap-3 px-4 py-3.5 text-left transition-colors duration-200"
        style={{ background: open ? '#f8fafc' : '#ffffff' }}
        onMouseEnter={() => setHover(true)}
        onMouseLeave={() => setHover(false)}
      >
        <div className="flex-1 min-w-0">{summary}</div>
        <div className="flex-shrink-0 ml-2" style={{ color: '#94a3b8' }}>
          {open ? <ChevronUp size={16} /> : <ChevronDown size={16} />}
        </div>
      </button>

      {open && (
        <div className="px-4 pb-4 pt-2 space-y-3" style={{ background: '#f8fafc', borderTop: '1px solid #f1f5f9' }}>
          <div className="flex gap-2.5">
            <Info size={14} className="flex-shrink-0 mt-0.5" style={{ color: '#64748b' }} />
            <p className="text-xs leading-relaxed" style={{ color: '#475569' }}>{explanation || 'No additional detail available for this signal'}</p>
          </div>
          <div className="flex gap-2.5">
            <Lightbulb size={14} className="flex-shrink-0 mt-0.5" style={{ color: '#6366f1' }} />
            <p className="text-xs leading-relaxed" style={{ color: '#4338ca' }}>{suggestion || 'No additional detail available for this signal'}</p>
          </div>
        </div>
      )}
    </div>
  )
}

// ── List renderers ───────────────────────────────────────────────────────────────

function NeedRow({ item, variant }: { item: SignalsNeedRisk; variant: 'cross-team' | 'unbacked' | 'at-risk' }) {
  const explanation =
    variant === 'unbacked'
      ? `"${item.need_name}" (actor: ${item.actor_name}) has no capability mapped to it. This means there is no part of the system responsible for delivering this user need — it exists in the model but has no implementation path.`
      : variant === 'cross-team'
      ? `"${item.need_name}" (actor: ${item.actor_name}) requires ${item.team_span ?? 'multiple'} different teams to coordinate before it can be delivered. Each additional team adds handoff overhead, release coordination risk, and potential for misalignment.`
      : `"${item.need_name}" (actor: ${item.actor_name}) is flagged at-risk because it spans ${item.team_span ?? 'multiple'} teams${item.teams?.length ? ` (${item.teams.join(', ')})` : ''}, exceeding the recommended threshold.`

  const suggestion =
    variant === 'unbacked'
      ? 'Add a capability that addresses this need, or explicitly retire the need if it is no longer relevant.'
      : `Consider restructuring ownership so this need is primarily served by a single stream-aligned team. Use platform capabilities to reduce cross-team dependencies.`

  const summary = (
    <div className="w-full space-y-2">
      <div>
        <div className="text-sm font-semibold leading-snug" style={{ color: '#1e293b' }}>{item.need_name}</div>
        <div className="text-xs mt-1" style={{ color: '#94a3b8' }}>actor: {item.actor_name}</div>
      </div>
      <div className="flex items-center gap-1.5 flex-wrap">
        {variant === 'unbacked' && <Tag text="No backing" color="red" />}
        {(item.team_span ?? 0) > 0 && <Tag text={`${item.team_span} team${item.team_span !== 1 ? 's' : ''}`} color="amber" />}
        {item.teams?.slice(0, 3).map(t => <Tag key={t} text={t} />)}
        {(item.teams?.length ?? 0) > 3 && <Tag text={`+${(item.teams?.length ?? 0) - 3} more`} />}
      </div>
    </div>
  )
  return <ExpandableRow summary={summary} explanation={explanation} suggestion={suggestion} />
}

function CapRow({ item, variant }: { item: SignalsCapItem; variant: 'cross-team' | 'unconnected' | 'fragmented' }) {
  const explanation =
    variant === 'unconnected'
      ? `"${item.capability_name}" exists in the system but is not connected to any user need. It may be a legacy capability with no current user value, or a need for it was never modelled.`
      : variant === 'cross-team'
      ? `"${item.capability_name}" is user-facing but its underlying services are owned by ${item.team_count ?? 'multiple'} different teams. Delivering a change to this capability requires coordinating across team boundaries.`
      : `"${item.capability_name}" is realised by services owned across ${item.team_count ?? 'multiple'} teams. No single team has clear accountability for this capability, which leads to duplicated effort and unclear ownership.`

  const suggestion =
    variant === 'unconnected'
      ? 'Identify which user need this capability supports and add the relationship, or remove it if it is genuinely unused.'
      : variant === 'cross-team'
      ? 'Move the services that realise this capability under a single team, or extract shared logic into a platform service.'
      : 'Consolidate ownership by moving services to one team, or split the capability into sub-capabilities each owned by one team.'

  const summary = (
    <div className="w-full space-y-2">
      <div className="text-sm font-semibold leading-snug" style={{ color: '#1e293b' }}>{item.capability_name}</div>
      <div className="flex items-center gap-1.5 flex-wrap">
        {item.visibility && <Tag text={item.visibility} />}
        {item.team_count != null && item.team_count > 0 && <Tag text={`${item.team_count} teams`} color="amber" />}
        {item.teams?.slice(0, 2).map(t => <Tag key={t} text={t} />)}
      </div>
    </div>
  )
  return <ExpandableRow summary={summary} explanation={explanation} suggestion={suggestion} />
}

function TeamRow({ item, variant }: { item: SignalsTeamItem; variant: 'load' | 'coherence' }) {
  const loadRisk = item.overall_level === 'high' ? 'red' : item.overall_level === 'medium' ? 'amber' : 'green'

  const explanation =
    variant === 'load'
      ? `${item.team_name} (${item.team_type ?? 'team'}) owns ${item.capability_count ?? '?'} capabilities and ${item.service_count ?? '?'} services. The structural load score combines domain spread, service count, dependency count, and team interaction volume relative to team size. A high score means the team is cognitively overloaded and will struggle to maintain quality and velocity.`
      : `${item.team_name} has a coherence score of ${item.coherence_score != null ? Math.round(item.coherence_score * 100) : '?'}%. Coherence measures how much of a team's capabilities connect to the same user needs and value chains. A low score means the team owns capabilities across many unrelated domains — they are not aligned to a clear mission.`

  const suggestion =
    variant === 'load'
      ? 'Consider splitting this team, moving lower-priority capabilities to another team, or converting some capabilities into platform services to reduce active ownership burden.'
      : 'Redistribute capabilities so each team owns a coherent slice of the value chain. Teams should own capabilities that serve the same actors and needs.'

  const summary = (
    <div className="w-full space-y-2">
      <div>
        <div className="text-sm font-semibold leading-snug" style={{ color: '#1e293b' }}>{item.team_name}</div>
        {item.team_type && <div className="text-xs mt-1" style={{ color: '#94a3b8' }}>{item.team_type}</div>}
      </div>
      <div className="flex items-center gap-1.5 flex-wrap">
        {variant === 'load' && item.capability_count != null && item.capability_count > 0 && <Tag text={`${item.capability_count} caps`} />}
        {variant === 'load' && item.service_count != null && item.service_count > 0 && <Tag text={`${item.service_count} svcs`} />}
        {variant === 'load' && item.overall_level && <RiskPill risk={loadRisk} />}
        {variant === 'coherence' && item.coherence_score != null && (
          <span className="inline-flex items-center" style={{ ...tagBase, background: '#fee2e2', color: '#b91c1c' }}>
            {Math.round(item.coherence_score * 100)}% coherent
          </span>
        )}
      </div>
    </div>
  )
  return <ExpandableRow summary={summary} explanation={explanation} suggestion={suggestion} />
}

function ServiceRow({ item }: { item: SignalsServiceItem }) {
  const explanation = `${item.service_name} has ${item.fan_in} other services depending on it directly. Fan-in is the number of upstream callers. A high fan-in makes this service a single point of failure — if it goes down, has a breaking API change, or degrades, all ${item.fan_in} dependents are affected.`
  const suggestion = `Consider whether ${item.service_name} should be split into smaller, more focused services. Alternatively, introduce an API gateway or facade to decouple consumers from the implementation, reducing blast radius.`

  const summary = (
    <div className="w-full space-y-2">
      <div>
        <div className="text-sm font-semibold leading-snug" style={{ color: '#1e293b' }}>{item.service_name}</div>
        <div className="text-xs mt-1" style={{ color: '#94a3b8' }}>{item.fan_in} other services depend on this</div>
      </div>
      <div className="flex items-center gap-1.5 flex-wrap">
        <span className="inline-flex items-center gap-1 font-bold" style={{ ...tagBase, background: '#fee2e2', color: '#b91c1c' }}>
          {item.fan_in} dependents
        </span>
      </div>
    </div>
  )
  return <ExpandableRow summary={summary} explanation={explanation} suggestion={suggestion} />
}

function ExtDepRow({ item }: { item: ExtDepItem }) {
  const color = item.is_critical ? 'red' : 'amber'
  const summary = (
    <div className="w-full space-y-2">
      <div className="text-sm font-semibold leading-snug" style={{ color: '#1e293b' }}>{item.dep_name}</div>
      <div className="flex items-center gap-1.5 flex-wrap">
        <Tag text={`${item.service_count} services`} color={color} />
      </div>
    </div>
  )
  const explanation = `${item.dep_name} is used by ${item.service_count} services in this platform. When this external system is unavailable or changes its API, all ${item.service_count} dependent services are affected simultaneously — making it a concentration risk.`
  const suggestion = item.is_critical
    ? `Consider introducing an abstraction layer or adapter around ${item.dep_name} to decouple services from the external system. This reduces blast radius and enables easier migration.`
    : `Monitor ${item.dep_name}'s reliability closely. Document the dependency in runbooks and ensure ownership is clear across all dependent teams.`
  return <ExpandableRow summary={summary} explanation={explanation} suggestion={suggestion} />
}

// ── Layout components ────────────────────────────────────────────────────────────

function HealthCard({ label, risk, icon: Icon, subtitle, findingsCount }: { label: string; risk: string; icon: React.ElementType; subtitle: string; findingsCount: number }) {
  const s = rs(risk)
  const criticalCount = risk === 'red' ? findingsCount : 0
  const mediumCount = risk === 'amber' ? findingsCount : 0
  const tooltipText = risk === 'green'
    ? `${label}: No signals detected — healthy`
    : `Score based on ${criticalCount > 0 ? `${criticalCount} critical` : ''}${criticalCount > 0 && mediumCount > 0 ? ' and ' : ''}${mediumCount > 0 ? `${mediumCount} medium` : ''} signal${findingsCount !== 1 ? 's' : ''} in this layer`
  return (
    <div
      title={tooltipText}
      className="flex-1 relative overflow-hidden transition-all duration-200 ease-out hover:-translate-y-px"
      style={{
        cursor: 'help', padding: '14px 16px',
        borderRadius: 16,
        background: s.cardGradient,
        border: `1px solid ${s.border}`,
        boxShadow: '0 2px 8px rgba(15, 23, 42, 0.06), inset 0 1px 0 rgba(255,255,255,0.85)',
      }}
    >
      <div
        className="pointer-events-none absolute inset-0 opacity-[0.35]"
        style={{ background: `radial-gradient(120% 80% at 100% 0%, ${s.dot}22 0%, transparent 55%)` }}
      />
      <div className="relative flex items-center justify-between mb-2">
        <div
          className="rounded-xl p-2 flex items-center justify-center"
          style={{
            background: s.iconWrap,
            border: `1px solid ${s.border}`,
          }}
        >
          <Icon size={16} style={{ color: s.dot }} strokeWidth={2.25} />
        </div>
        <RiskPill risk={risk} />
      </div>
      <div className="relative">
        <div className="font-bold leading-snug" style={{ fontSize: 13, fontWeight: 700, color: '#1e293b' }}>{label}</div>
        <div className="mt-1 flex items-baseline gap-1.5">
          <span className="tabular-nums" style={{ fontSize: 20, fontWeight: 800, color: '#1e293b' }}>{findingsCount}</span>
          <span className="uppercase font-semibold" style={{ fontSize: 10, fontWeight: 600, color: '#64748b', letterSpacing: '0.05em' }}>findings</span>
        </div>
        <div className="text-xs mt-1 font-medium leading-snug" style={{ color: s.text }}>{subtitle}</div>
      </div>
    </div>
  )
}

// ── Rich markdown components for AI modal ────────────────────────────────────────

const mdComponents = {
  h1: ({ children }: { children?: React.ReactNode }) => (
    <h1
      className="text-2xl font-extrabold mb-5 pb-3 tracking-tight"
      style={{
        borderBottom: '2px solid transparent',
        borderImage: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%) 1',
        color: '#1e293b',
        letterSpacing: '-0.02em',
      }}
    >
      {children}
    </h1>
  ),
  h2: ({ children }: { children?: React.ReactNode }) => (
    <h2
      className="text-lg font-bold mt-8 mb-3 pb-2 flex items-center gap-2"
      style={{ color: '#6366f1', borderBottom: '1px solid #e2e8f0' }}
    >
      <span
        className="inline-block h-2 w-2 rounded-full shrink-0"
        style={{ background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)' }}
        aria-hidden
      />
      {children}
    </h2>
  ),
  h3: ({ children }: { children?: React.ReactNode }) => (
    <h3 className="text-base font-bold mt-6 mb-2" style={{ color: '#334155', letterSpacing: '-0.01em' }}>
      {children}
    </h3>
  ),
  strong: ({ children }: { children?: React.ReactNode }) => <strong style={{ color: '#1e293b', fontWeight: 700 }}>{children}</strong>,
  li: ({ children }: { children?: React.ReactNode }) => <li className="text-sm leading-relaxed" style={{ color: '#475569' }}>{children}</li>,
  p: ({ children }: { children?: React.ReactNode }) => <p className="text-sm leading-relaxed mb-4" style={{ color: '#475569' }}>{children}</p>,
  table: ({ children }: { children?: React.ReactNode }) => (
    <div className="my-6 overflow-x-auto rounded-xl border" style={{ borderColor: '#e2e8f0', boxShadow: '0 1px 3px rgba(0,0,0,0.04)' }}>
      <table className="w-full text-xs border-collapse min-w-full">{children}</table>
    </div>
  ),
  thead: ({ children }: { children?: React.ReactNode }) => <thead style={{ background: 'linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%)' }}>{children}</thead>,
  th: ({ children }: { children?: React.ReactNode }) => (
    <th className="text-left px-4 py-3 font-bold text-xs uppercase tracking-wide" style={{ borderBottom: '2px solid #e2e8f0', color: '#475569' }}>
      {children}
    </th>
  ),
  td: ({ children }: { children?: React.ReactNode }) => (
    <td className="px-4 py-3 text-sm" style={{ borderBottom: '1px solid #f1f5f9', color: '#475569', background: '#ffffff' }}>
      {children}
    </td>
  ),
  tr: ({ children }: { children?: React.ReactNode }) => <tr className="transition-colors hover:bg-slate-50/80">{children}</tr>,
  code: ({ children }: { children?: React.ReactNode }) => (
    <code className="px-1.5 py-0.5 rounded-md text-xs font-mono" style={{ background: '#f1f5f9', color: '#6366f1', border: '1px solid #e2e8f0' }}>
      {children}
    </code>
  ),
  blockquote: ({ children }: { children?: React.ReactNode }) => (
    <blockquote className="border-l-4 pl-4 my-4 italic text-sm" style={{ borderColor: '#a5b4fc', color: '#64748b' }}>
      {children}
    </blockquote>
  ),
  hr: () => (
    <hr className="my-6" style={{ border: 'none', height: 1, background: 'linear-gradient(90deg, #e2e8f0 0%, #c7d2fe 50%, #e2e8f0 100%)' }} />
  ),
}

// ── AI Recommendation Modal ──────────────────────────────────────────────────────

function AIRecommendationModal({ area, answer, onClose }: {
  area: string
  answer: string
  onClose: () => void
}) {
  useEffect(() => {
    const handler = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose() }
    document.addEventListener('keydown', handler)
    document.body.style.overflow = 'hidden'
    return () => {
      document.removeEventListener('keydown', handler)
      document.body.style.overflow = ''
    }
  }, [onClose])

  return (
    <div
      className="fixed inset-0 z-50 flex items-start justify-center"
      style={{ background: 'rgba(15, 23, 42, 0.6)', backdropFilter: 'blur(4px)' }}
      onClick={onClose}
    >
      <div
        className="relative w-full max-w-3xl my-8 mx-4 overflow-hidden"
        style={{
          borderRadius: 24,
          background: '#ffffff',
          boxShadow: '0 25px 50px -12px rgba(0,0,0,0.25), 0 0 0 1px rgba(0,0,0,0.05)',
          maxHeight: 'calc(100vh - 64px)',
          display: 'flex',
          flexDirection: 'column',
        }}
        onClick={e => e.stopPropagation()}
      >
        {/* Header */}
        <div
          className="flex items-center justify-between px-8 py-5 flex-shrink-0"
          style={{
            background: 'linear-gradient(135deg, #eef2ff 0%, #f5f3ff 100%)',
            borderBottom: '1px solid #c7d2fe',
          }}
        >
          <div className="flex items-center gap-3">
            <div
              className="rounded-xl p-2.5"
              style={{
                background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)',
                boxShadow: '0 2px 8px rgba(99, 102, 241, 0.3)',
              }}
            >
              <Sparkles size={18} style={{ color: '#ffffff' }} />
            </div>
            <div>
              <div style={{ fontSize: 16, fontWeight: 800, color: '#1e293b', letterSpacing: '-0.02em' }}>
                AI Recommendation
              </div>
              <div style={{ fontSize: 12, color: '#6366f1', fontWeight: 600, marginTop: 2 }}>
                {area} Layer Analysis
              </div>
            </div>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="rounded-xl p-2 transition-colors duration-150 hover:bg-white/80"
            style={{ color: '#64748b' }}
          >
            <X size={20} />
          </button>
        </div>

        {/* Body */}
        <div className="overflow-y-auto flex-1 px-8 py-6 prose prose-sm max-w-none">
          <ReactMarkdown remarkPlugins={[remarkGfm]} components={mdComponents as never}>
            {answer}
          </ReactMarkdown>
        </div>
      </div>
    </div>
  )
}

// ── AI Recommendation cache (survives component unmount/remount) ─────────────────

const aiRecommendationCache = new Map<string, string>()

// ── AI Recommendation Panel (button + loading in section card) ───────────────────

function AIRecommendationPanel({ area, modelId, findingsSummary, aiEnabled }: {
  area: string
  modelId: string
  findingsSummary: string
  aiEnabled: boolean
}) {
  const cacheKey = `${modelId}:${area}`
  const cached = aiRecommendationCache.get(cacheKey)
  const [state, setState] = useState<'idle' | 'loading' | 'done' | 'error'>(cached ? 'done' : 'idle')
  const [answer, setAnswer] = useState(cached ?? '')
  const [modalOpen, setModalOpen] = useState(false)

  const generate = useCallback(() => {
    setState('loading')
    const question = `Analyze the ${area} layer signals and provide a consolidated recommendation. Findings:\n${findingsSummary}\n\nProvide: (1) a brief summary of the most critical issues in this layer, (2) a prioritized action plan with specific entity names (teams, services, capabilities), (3) expected impact of each action. Be concise and actionable.`
    api.askAdvisor(modelId, question, 'general')
      .then(resp => {
        aiRecommendationCache.set(cacheKey, resp.answer)
        setAnswer(resp.answer)
        setState('done')
        setModalOpen(true)
      })
      .catch(() => setState('error'))
  }, [area, modelId, findingsSummary, cacheKey])

  if (!aiEnabled) return null

  return (
    <>
      <div className="px-6 pb-5">
        {state === 'idle' && (
          <button
            type="button"
            onClick={generate}
            className="inline-flex items-center gap-2 px-4 py-2.5 rounded-xl text-sm font-semibold transition-all duration-200 hover:shadow-md hover:-translate-y-px"
            style={{
              background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)',
              color: '#ffffff',
              border: 'none',
              cursor: 'pointer',
            }}
          >
            <Sparkles size={15} />
            AI Recommendation
          </button>
        )}

        {state === 'loading' && (
          <div
            className="rounded-xl p-4 flex items-center gap-3"
            style={{ background: 'linear-gradient(135deg, #eef2ff 0%, #f5f3ff 100%)', border: '1px solid #c7d2fe' }}
          >
            <span
              className="animate-spin flex-shrink-0"
              style={{ width: 18, height: 18, border: '2px solid #c7d2fe', borderTopColor: '#6366f1', borderRadius: '50%' }}
            />
            <span style={{ fontSize: 13, color: '#6366f1', fontWeight: 600 }}>
              Generating {area} recommendation…
            </span>
          </div>
        )}

        {state === 'error' && (
          <div className="rounded-xl p-4 flex items-center justify-between" style={{ background: '#fef2f2', border: '1px solid #fecaca' }}>
            <span style={{ fontSize: 13, color: '#b91c1c', fontWeight: 500 }}>Failed to generate recommendation.</span>
            <button type="button" onClick={generate} className="text-xs font-semibold underline" style={{ color: '#6366f1' }}>Retry</button>
          </div>
        )}

        {state === 'done' && (
          <button
            type="button"
            onClick={() => setModalOpen(true)}
            className="inline-flex items-center gap-2 px-4 py-2.5 rounded-xl text-sm font-semibold transition-all duration-200 hover:shadow-md hover:-translate-y-px"
            style={{
              background: 'linear-gradient(135deg, #eef2ff 0%, #f5f3ff 100%)',
              color: '#6366f1',
              border: '1px solid #c7d2fe',
              cursor: 'pointer',
            }}
          >
            <Sparkles size={15} />
            View Recommendation
          </button>
        )}
      </div>

      {modalOpen && state === 'done' && (
        <AIRecommendationModal area={area} answer={answer} onClose={() => setModalOpen(false)} />
      )}
    </>
  )
}

function SectionCard({ title, risk, icon: Icon, children, empty, aiArea, modelId, findingsSummary, aiEnabled, highlighted, sectionRef }: {
  title: string; risk: string; icon: React.ElementType; children: React.ReactNode; empty: boolean
  aiArea?: string; modelId?: string; findingsSummary?: string; aiEnabled?: boolean
  highlighted?: boolean; sectionRef?: React.RefObject<HTMLDivElement | null>
}) {
  const s = rs(risk)
  return (
    <div ref={sectionRef} className="overflow-hidden transition-all duration-200 ease-out hover:shadow-md" style={{ ...cardShell, ...(highlighted ? { boxShadow: '0 0 0 3px #6366f1, 0 4px 16px rgba(99,102,241,0.2)', borderColor: '#6366f1' } : {}) }}>
      <div style={{ height: 3, background: s.stripGradient }} />
      <div className="flex items-center justify-between px-4 py-3" style={{ borderBottom: '1px solid #f1f5f9' }}>
        <div className="flex items-center gap-2.5 min-w-0">
          <div
            className="rounded-lg p-2 flex-shrink-0"
            style={{
              background: s.bg,
              border: `1px solid ${s.border}`,
            }}
          >
            <Icon size={14} style={{ color: s.dot }} strokeWidth={2.25} />
          </div>
          <div className="min-w-0">
            <div className="font-bold truncate" style={{ fontSize: 13, fontWeight: 700, color: '#1e293b' }}>{title}</div>
            <div style={{ fontSize: 11, color: '#94a3b8' }}>Click any row for details</div>
          </div>
        </div>
        <div className="flex-shrink-0 ml-2">
          <RiskPill risk={risk} />
        </div>
      </div>
      <div className="px-4 py-4">
        {empty
          ? (
            <div className="flex items-center gap-2.5 py-1">
              <div className="rounded-lg p-1.5" style={{ background: '#dcfce7', border: '1px solid #86efac' }}>
                <CheckCircle size={14} style={{ color: '#22c55e' }} strokeWidth={2.25} />
              </div>
              <span className="text-xs font-semibold" style={{ color: '#15803d' }}>No risks detected in this layer</span>
            </div>
          )
          : <div className="space-y-5">{children}</div>
        }
      </div>
      {aiArea && modelId && findingsSummary && !empty && (
        <AIRecommendationPanel area={aiArea} modelId={modelId} findingsSummary={findingsSummary} aiEnabled={!!aiEnabled} />
      )}
    </div>
  )
}

function SubSection({ title, icon: Icon, count, children }: { title: string; icon: React.ElementType; count: number; children: React.ReactNode }) {
  return (
    <div>
      <div className="flex items-center gap-1.5 mb-2">
        <Icon size={12} style={{ color: '#94a3b8' }} strokeWidth={2} />
        <span
          className="font-semibold uppercase"
          style={{ fontSize: 10, fontWeight: 600, color: '#64748b', letterSpacing: '0.05em' }}
        >
          {title}
        </span>
        <CountBadge n={count} />
      </div>
      <div className="space-y-2">{children}</div>
    </div>
  )
}

function riskSubtitle(risk: string, findings: number) {
  if (risk === 'green') return 'No issues detected'
  return `${findings} finding${findings !== 1 ? 's' : ''} require${findings === 1 ? 's' : ''} attention`
}

const spinnerEl = (
  <span
    className="animate-spin flex-shrink-0"
    style={{
      width: 20,
      height: 20,
      border: '2px solid #e2e8f0',
      borderTopColor: '#6366f1',
      borderRadius: '50%',
    }}
  />
)

// ── Main view ───────────────────────────────────────────────────────────────────

function buildUxSummary(ux: SignalsViewResponse['user_experience_layer']): string {
  const lines: string[] = []
  if (ux.needs_requiring_3plus_teams.length > 0) {
    lines.push(`Cross-team needs (${ux.needs_requiring_3plus_teams.length}): ${ux.needs_requiring_3plus_teams.map(n => `"${n.need_name}" (span ${n.team_span}, teams: ${n.teams?.join(', ') ?? '?'})`).join('; ')}`)
  }
  if (ux.needs_with_no_capability_backing.length > 0) {
    lines.push(`Unbacked needs (${ux.needs_with_no_capability_backing.length}): ${ux.needs_with_no_capability_backing.map(n => `"${n.need_name}"`).join(', ')}`)
  }
  if (ux.needs_at_risk.length > 0) {
    lines.push(`At-risk needs (${ux.needs_at_risk.length}): ${ux.needs_at_risk.map(n => `"${n.need_name}" (span ${n.team_span})`).join(', ')}`)
  }
  return lines.join('\n')
}

function buildArchSummary(arch: SignalsViewResponse['architecture_layer']): string {
  const lines: string[] = []
  if (arch.capabilities_not_connected_to_any_need.length > 0) {
    lines.push(`Unlinked capabilities (${arch.capabilities_not_connected_to_any_need.length}): ${arch.capabilities_not_connected_to_any_need.map(c => `"${c.capability_name}"`).join(', ')}`)
  }
  if (arch.capabilities_fragmented_across_teams.length > 0) {
    lines.push(`Fragmented capabilities (${arch.capabilities_fragmented_across_teams.length}): ${arch.capabilities_fragmented_across_teams.map(c => `"${c.capability_name}" (${c.team_count} teams: ${c.teams?.join(', ') ?? '?'})`).join('; ')}`)
  }
  if (arch.user_facing_caps_with_cross_team_services.length > 0) {
    lines.push(`Cross-team user-facing caps (${arch.user_facing_caps_with_cross_team_services.length}): ${arch.user_facing_caps_with_cross_team_services.map(c => `"${c.capability_name}"`).join(', ')}`)
  }
  return lines.join('\n')
}

function buildOrgSummary(org: OrgLayerWithExtDeps): string {
  const lines: string[] = []
  if (org.top_teams_by_structural_load.length > 0) {
    lines.push(`High structural load teams (${org.top_teams_by_structural_load.length}): ${org.top_teams_by_structural_load.map(t => `"${t.team_name}" (${t.overall_level}, ${t.service_count} svcs, ${t.capability_count} caps)`).join('; ')}`)
  }
  if (org.critical_bottleneck_services.length > 0) {
    lines.push(`Critical bottleneck services (${org.critical_bottleneck_services.length}): ${org.critical_bottleneck_services.map(s => `"${s.service_name}" (fan-in: ${s.fan_in})`).join(', ')}`)
  }
  if (org.low_coherence_teams.length > 0) {
    lines.push(`Low coherence teams (${org.low_coherence_teams.length}): ${org.low_coherence_teams.map(t => `"${t.team_name}" (${t.coherence_score != null ? Math.round(t.coherence_score * 100) : '?'}%)`).join(', ')}`)
  }
  return lines.join('\n')
}

export function SignalsView() {
  const { modelId, isHydrating } = useRequireModel()
  const [data, setData] = useState<SignalsViewResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const aiEnabled = useAIEnabled()
  const [searchParams] = useSearchParams()
  const filterParam = searchParams.get('filter')
  const uxRef = useRef<HTMLDivElement | null>(null)
  const archRef = useRef<HTMLDivElement | null>(null)
  const orgRef = useRef<HTMLDivElement | null>(null)

  // Map filter param to which section gets highlighted
  const filterSection = (f: string | null) => {
    if (!f) return null
    if (f === 'needs-at-risk') return 'ux'
    if (f === 'gap') return 'arch'
    if (f === 'fragmentation') return 'arch'
    if (f === 'bottleneck' || f === 'cognitive-load') return 'org'
    return null
  }
  const highlightedSection = filterSection(filterParam)

  useEffect(() => {
    if (isHydrating || !modelId) return
    api.getSignals(modelId)
      .then(setData)
      .catch(e => setError((e as Error).message))
      .finally(() => setLoading(false))
  }, [isHydrating, modelId])

  // Scroll to highlighted section after data loads
  useEffect(() => {
    if (!data || !highlightedSection) return
    const refMap = { ux: uxRef, arch: archRef, org: orgRef }
    const ref = refMap[highlightedSection as keyof typeof refMap]
    if (ref?.current) {
      setTimeout(() => ref.current?.scrollIntoView({ behavior: 'smooth', block: 'start' }), 100)
    }
  }, [data, highlightedSection])

  if (loading) return (
    <div className="flex flex-col items-center justify-center gap-3 h-64">
      {spinnerEl}
      <span style={{ fontSize: 14, color: '#94a3b8', fontWeight: 500 }}>Loading signals…</span>
    </div>
  )
  if (error) return (
    <div
      className="flex items-center justify-center gap-3 h-64 rounded-2xl mx-auto max-w-lg px-6"
      style={{ ...cardShell, color: '#b91c1c', borderColor: '#fecaca' }}
    >
      <div className="rounded-xl p-2" style={{ background: '#fee2e2' }}>
        <span title="Error loading signals" aria-label="Error loading signals" style={{ cursor: 'help' }}>
          <AlertTriangle size={20} style={{ color: '#dc2626' }} />
        </span>
      </div>
      <span className="text-sm font-medium">{error}</span>
    </div>
  )
  if (!data) return null

  const ux = data.user_experience_layer
  const arch = data.architecture_layer
  const org = data.organization_layer as OrgLayerWithExtDeps

  const uxCount = ux.needs_requiring_3plus_teams.length + ux.needs_with_no_capability_backing.length + ux.needs_at_risk.length
  const archCount = arch.user_facing_caps_with_cross_team_services.length + arch.capabilities_not_connected_to_any_need.length + arch.capabilities_fragmented_across_teams.length
  const orgCount = org.top_teams_by_structural_load.length + org.critical_bottleneck_services.length + org.low_coherence_teams.length + (org.critical_external_deps?.length ?? 0)

  return (
    <div className="space-y-5">
      <div>
        <h1 style={{ ...gradientTitle, fontSize: 24 }}>Architecture Signals</h1>
        <p style={{ fontSize: 13, color: '#64748b', marginTop: 4 }}>
          Health signals across UX, architecture and organisational layers — click any row for details
        </p>
      </div>

      {/* Health summary */}
      <div className="grid gap-3" style={{ gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))' }}>
        <HealthCard label="User Experience"  risk={data.health.ux_risk}            icon={Users}     subtitle={riskSubtitle(data.health.ux_risk, uxCount)} findingsCount={uxCount} />
        <HealthCard label="Architecture"     risk={data.health.architecture_risk}   icon={Layers}    subtitle={riskSubtitle(data.health.architecture_risk, archCount)} findingsCount={archCount} />
        <HealthCard label="Organization"     risk={data.health.org_risk}            icon={Building2} subtitle={riskSubtitle(data.health.org_risk, orgCount)} findingsCount={orgCount} />
      </div>

      {/* Signal sections */}
      <div className="grid gap-4" style={{ gridTemplateColumns: 'repeat(auto-fit, minmax(340px, 1fr))', alignItems: 'start' }}>
        {/* UX layer */}
        <SectionCard title="User Experience Layer" risk={data.health.ux_risk} icon={Users} empty={uxCount === 0}
          aiArea="User Experience" modelId={modelId ?? ''} findingsSummary={buildUxSummary(ux)} aiEnabled={aiEnabled}
          sectionRef={uxRef} highlighted={highlightedSection === 'ux'}>
          {ux.needs_requiring_3plus_teams.length > 0 && (
            <SubSection title="Needs served by 3+ teams" icon={Zap} count={ux.needs_requiring_3plus_teams.length}>
              {ux.needs_requiring_3plus_teams.map((item, i) => <NeedRow key={i} item={item} variant="cross-team" />)}
            </SubSection>
          )}
          {ux.needs_with_no_capability_backing.length > 0 && (
            <SubSection title="Needs with no capability backing" icon={Link2Off} count={ux.needs_with_no_capability_backing.length}>
              {ux.needs_with_no_capability_backing.map((item, i) => <NeedRow key={i} item={item} variant="unbacked" />)}
            </SubSection>
          )}
          {ux.needs_at_risk.length > 0 && (
            <SubSection title="Needs at risk" icon={AlertTriangle} count={ux.needs_at_risk.length}>
              {ux.needs_at_risk.map((item, i) => <NeedRow key={i} item={item} variant="at-risk" />)}
            </SubSection>
          )}
        </SectionCard>

        {/* Architecture layer */}
        <SectionCard title="Architecture Layer" risk={data.health.architecture_risk} icon={Layers} empty={archCount === 0}
          aiArea="Architecture" modelId={modelId ?? ''} findingsSummary={buildArchSummary(arch)} aiEnabled={aiEnabled}
          sectionRef={archRef} highlighted={highlightedSection === 'arch'}>
          {arch.user_facing_caps_with_cross_team_services.length > 0 && (
            <SubSection title="User-facing caps with cross-team services" icon={GitMerge} count={arch.user_facing_caps_with_cross_team_services.length}>
              {arch.user_facing_caps_with_cross_team_services.map((item, i) => <CapRow key={i} item={item} variant="cross-team" />)}
            </SubSection>
          )}
          {arch.capabilities_not_connected_to_any_need.length > 0 && (
            <SubSection title="Caps not connected to any need" icon={Link2Off} count={arch.capabilities_not_connected_to_any_need.length}>
              {arch.capabilities_not_connected_to_any_need.map((item, i) => <CapRow key={i} item={item} variant="unconnected" />)}
            </SubSection>
          )}
          {arch.capabilities_fragmented_across_teams.length > 0 && (
            <SubSection title="Caps fragmented across teams" icon={GitMerge} count={arch.capabilities_fragmented_across_teams.length}>
              {arch.capabilities_fragmented_across_teams.map((item, i) => <CapRow key={i} item={item} variant="fragmented" />)}
            </SubSection>
          )}
        </SectionCard>

        {/* Org layer */}
        <SectionCard title="Organization Layer" risk={data.health.org_risk} icon={Building2} empty={orgCount === 0}
          aiArea="Organization" modelId={modelId ?? ''} findingsSummary={buildOrgSummary(org)} aiEnabled={aiEnabled}
          sectionRef={orgRef} highlighted={highlightedSection === 'org'}>
          {org.top_teams_by_structural_load.length > 0 && (
            <SubSection title="Teams under high structural load" icon={Zap} count={org.top_teams_by_structural_load.length}>
              {org.top_teams_by_structural_load.map((item, i) => <TeamRow key={i} item={item} variant="load" />)}
            </SubSection>
          )}
          {org.critical_bottleneck_services.length > 0 && (
            <SubSection title="Critical bottleneck services" icon={Server} count={org.critical_bottleneck_services.length}>
              {org.critical_bottleneck_services.map((item, i) => <ServiceRow key={i} item={item} />)}
            </SubSection>
          )}
          {org.low_coherence_teams.length > 0 && (
            <SubSection title="Low value stream coherence" icon={TrendingDown} count={org.low_coherence_teams.length}>
              {org.low_coherence_teams.map((item, i) => <TeamRow key={i} item={item} variant="coherence" />)}
            </SubSection>
          )}
          {org.critical_external_deps && org.critical_external_deps.length > 0 && (
            <SubSection title="External Dependency Concentration" icon={Link2} count={org.critical_external_deps.length}>
              {org.critical_external_deps.map((item, i) => <ExtDepRow key={i} item={item} />)}
            </SubSection>
          )}
        </SectionCard>
      </div>
    </div>
  )
}
