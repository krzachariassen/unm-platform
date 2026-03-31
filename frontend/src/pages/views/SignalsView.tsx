import { useRef, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useSearchParams } from 'react-router-dom'
import { Users, Layers, Building2, AlertTriangle, Zap, Link2Off, GitMerge, Server, TrendingDown, Link2 } from 'lucide-react'
import { ModelRequired } from '@/components/ui/ModelRequired'
import { PageHeader } from '@/components/ui/page-header'
import { LoadingState, ErrorState } from '@/components/ViewState'
import { useModel } from '@/lib/model-context'
import { useAIEnabled } from '@/hooks/useAIEnabled'
import { viewsApi } from '@/services/api'
import {
  rs, riskSubtitle, tagBase,
  buildUxSummary, buildArchSummary, buildOrgSummary,
} from '@/features/signals/risk-config'
import type { SignalsNeedRisk, SignalsCapItem, SignalsTeamItem, SignalsServiceItem, SignalsExtDepItem } from '@/types/views'

// ── Atoms ────────────────────────────────────────────────────────────────────────

import { useState, useCallback, useEffect as useEff } from 'react'
import { ChevronDown, ChevronUp, Info, Lightbulb, CheckCircle, Sparkles, X } from 'lucide-react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { advisorApi } from '@/services/api'

function RiskPill({ risk }: { risk: string }) {
  const s = rs(risk)
  return <span className="inline-flex items-center gap-1.5" style={{ ...tagBase, background: s.badgeBg, color: s.text }}><span className="w-1.5 h-1.5 rounded-full shrink-0" style={{ background: s.dot }} />{s.label}</span>
}

function CountBadge({ n }: { n: number }) {
  return <span className="ml-2 inline-flex items-center justify-center min-w-[22px] h-[22px] px-1.5 rounded-lg text-[11px] font-bold bg-slate-100 text-slate-500 border border-slate-200">{n}</span>
}

function Tag({ text, color }: { text: string; color?: 'amber' | 'red' }) {
  const style = color === 'red' ? { background: '#fee2e2', color: '#b91c1c' } : color === 'amber' ? { background: '#fef3c7', color: '#92400e' } : { background: '#f1f5f9', color: '#64748b' }
  return <span className="inline-flex items-center" style={{ ...tagBase, ...style }}>{text}</span>
}

function ExpandableRow({ summary, explanation, suggestion }: { summary: React.ReactNode; explanation: string; suggestion: string }) {
  const [open, setOpen] = useState(false)
  return (
    <div className="overflow-hidden rounded-2xl border border-slate-100 hover:shadow-md transition-all duration-200">
      <button type="button" onClick={() => setOpen(o => !o)} className="w-full flex items-center justify-between gap-3 px-4 py-3.5 text-left" style={{ background: open ? '#f8fafc' : '#ffffff' }}>
        <div className="flex-1 min-w-0">{summary}</div>
        <div className="shrink-0 text-slate-400">{open ? <ChevronUp size={16} /> : <ChevronDown size={16} />}</div>
      </button>
      {open && (
        <div className="px-4 pb-4 pt-2 space-y-3 bg-slate-50 border-t border-slate-100">
          <div className="flex gap-2.5"><Info size={14} className="shrink-0 mt-0.5 text-slate-500" /><p className="text-xs leading-relaxed text-slate-600">{explanation || 'No detail available'}</p></div>
          <div className="flex gap-2.5"><Lightbulb size={14} className="shrink-0 mt-0.5 text-indigo-500" /><p className="text-xs leading-relaxed text-indigo-700">{suggestion || 'No detail available'}</p></div>
        </div>
      )}
    </div>
  )
}

// ── Row renderers ────────────────────────────────────────────────────────────────

function NeedRow({ item, variant }: { item: SignalsNeedRisk; variant: 'cross-team' | 'unbacked' | 'at-risk' }) {
  const explanation = variant === 'unbacked'
    ? `"${item.need_name}" (actor: ${item.actor_names?.join(' & ')}) has no capability mapped to it — no implementation path.`
    : `"${item.need_name}" requires ${item.team_span ?? 'multiple'} teams to coordinate. Each adds handoff overhead and release risk.`
  const suggestion = variant === 'unbacked'
    ? 'Add a capability addressing this need, or retire it if no longer relevant.'
    : 'Restructure ownership so this need is served by a single stream-aligned team.'
  return (
    <ExpandableRow explanation={explanation} suggestion={suggestion} summary={
      <div className="space-y-2">
        <div><div className="text-sm font-semibold text-slate-800">{item.need_name}</div><div className="text-xs text-slate-400 mt-0.5">actor: {item.actor_names?.join(', ')}</div></div>
        <div className="flex flex-wrap gap-1.5">{variant === 'unbacked' && <Tag text="No backing" color="red" />}{(item.team_span ?? 0) > 0 && <Tag text={`${item.team_span} team${item.team_span !== 1 ? 's' : ''}`} color="amber" />}{item.teams?.slice(0, 3).map(t => <Tag key={t} text={t} />)}</div>
      </div>
    } />
  )
}

function CapRow({ item, variant }: { item: SignalsCapItem; variant: 'cross-team' | 'unconnected' | 'fragmented' }) {
  const explanation = variant === 'unconnected' ? `"${item.capability_name}" is not connected to any user need — may be legacy.`
    : variant === 'cross-team' ? `"${item.capability_name}" is user-facing but its services are owned by ${item.team_count ?? 'multiple'} teams.`
    : `"${item.capability_name}" is realised across ${item.team_count ?? 'multiple'} teams — unclear ownership.`
  const suggestion = variant === 'unconnected' ? 'Identify the need it supports or remove it if unused.'
    : variant === 'cross-team' ? 'Move services under a single team or extract shared logic to a platform service.'
    : 'Consolidate ownership or split into sub-capabilities each owned by one team.'
  return (
    <ExpandableRow explanation={explanation} suggestion={suggestion} summary={
      <div className="space-y-2">
        <div className="text-sm font-semibold text-slate-800">{item.capability_name}</div>
        <div className="flex flex-wrap gap-1.5">{item.visibility && <Tag text={item.visibility} />}{item.team_count != null && item.team_count > 0 && <Tag text={`${item.team_count} teams`} color="amber" />}{item.teams?.slice(0, 2).map(t => <Tag key={t} text={t} />)}</div>
      </div>
    } />
  )
}

function TeamRow({ item, variant }: { item: SignalsTeamItem; variant: 'load' | 'coherence' }) {
  const loadRisk = item.overall_level === 'high' ? 'red' : item.overall_level === 'medium' ? 'amber' : 'green'
  const explanation = variant === 'load' ? `${item.team_name} owns ${item.capability_count ?? '?'} caps and ${item.service_count ?? '?'} svcs. High structural load means cognitive overload.`
    : `${item.team_name} has ${item.coherence_score != null ? Math.round(item.coherence_score * 100) : '?'}% coherence — owns capabilities across many unrelated domains.`
  const suggestion = variant === 'load' ? 'Consider splitting this team or moving lower-priority capabilities to another team.'
    : 'Redistribute capabilities so each team owns a coherent slice of the value chain.'
  return (
    <ExpandableRow explanation={explanation} suggestion={suggestion} summary={
      <div className="space-y-2">
        <div><div className="text-sm font-semibold text-slate-800">{item.team_name}</div>{item.team_type && <div className="text-xs text-slate-400 mt-0.5">{item.team_type}</div>}</div>
        <div className="flex flex-wrap gap-1.5">
          {variant === 'load' && item.capability_count != null && <Tag text={`${item.capability_count} caps`} />}
          {variant === 'load' && item.service_count != null && <Tag text={`${item.service_count} svcs`} />}
          {variant === 'load' && item.overall_level && <RiskPill risk={loadRisk} />}
          {variant === 'coherence' && item.coherence_score != null && <span className="inline-flex items-center" style={{ ...tagBase, background: '#fee2e2', color: '#b91c1c' }}>{Math.round(item.coherence_score * 100)}% coherent</span>}
        </div>
      </div>
    } />
  )
}

function ServiceRow({ item }: { item: SignalsServiceItem }) {
  return (
    <ExpandableRow
      explanation={`${item.service_name} has ${item.fan_in} dependents. If it fails, all ${item.fan_in} dependents are affected.`}
      suggestion={`Consider splitting ${item.service_name} or introducing a facade to decouple consumers.`}
      summary={
        <div className="space-y-2">
          <div><div className="text-sm font-semibold text-slate-800">{item.service_name}</div><div className="text-xs text-slate-400 mt-0.5">{item.fan_in} services depend on this</div></div>
          <div className="flex gap-1.5"><span className="inline-flex items-center font-bold" style={{ ...tagBase, background: '#fee2e2', color: '#b91c1c' }}>{item.fan_in} dependents</span></div>
        </div>
      }
    />
  )
}

function ExtDepRow({ item }: { item: SignalsExtDepItem }) {
  const color = item.is_critical ? 'red' : 'amber'
  return (
    <ExpandableRow
      explanation={`${item.dep_name} is used by ${item.service_count} services — a concentration risk if it goes down.`}
      suggestion={item.is_critical ? `Consider an abstraction layer around ${item.dep_name} to decouple services.` : `Monitor ${item.dep_name}'s reliability closely.`}
      summary={
        <div className="space-y-2">
          <div className="text-sm font-semibold text-slate-800">{item.dep_name}</div>
          <div className="flex gap-1.5"><Tag text={`${item.service_count} services`} color={color} /></div>
        </div>
      }
    />
  )
}

// ── AI Recommendation ─────────────────────────────────────────────────────────────

const aiCache = new Map<string, string>()

function AIModal({ area, answer, onClose }: { area: string; answer: string; onClose: () => void }) {
  useEff(() => {
    const handler = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose() }
    document.addEventListener('keydown', handler)
    document.body.style.overflow = 'hidden'
    return () => { document.removeEventListener('keydown', handler); document.body.style.overflow = '' }
  }, [onClose])
  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center bg-slate-900/60 backdrop-blur-sm" onClick={onClose}>
      <div className="relative w-full max-w-3xl my-8 mx-4 bg-white rounded-3xl shadow-2xl max-h-[calc(100vh-4rem)] flex flex-col" onClick={e => e.stopPropagation()}>
        <div className="flex items-center justify-between px-8 py-5 bg-gradient-to-r from-indigo-50 to-purple-50 border-b border-indigo-100 rounded-t-3xl">
          <div className="flex items-center gap-3">
            <div className="rounded-xl p-2.5 bg-gradient-to-br from-indigo-500 to-purple-600"><Sparkles size={18} className="text-white" /></div>
            <div><div className="text-base font-extrabold text-slate-800">AI Recommendation</div><div className="text-xs text-indigo-600 font-semibold mt-0.5">{area} Layer</div></div>
          </div>
          <button type="button" onClick={onClose} className="p-2 rounded-xl hover:bg-white/80 text-slate-500 transition-colors"><X size={20} /></button>
        </div>
        <div className="overflow-y-auto flex-1 px-8 py-6 prose prose-sm max-w-none">
          <ReactMarkdown remarkPlugins={[remarkGfm]}>{answer}</ReactMarkdown>
        </div>
      </div>
    </div>
  )
}

function AIRecommendationPanel({ area, modelId, findingsSummary, aiEnabled }: { area: string; modelId: string; findingsSummary: string; aiEnabled: boolean }) {
  const cacheKey = `${modelId}:${area}`
  const cached = aiCache.get(cacheKey)
  const [state, setState] = useState<'idle' | 'loading' | 'done' | 'error'>(cached ? 'done' : 'idle')
  const [answer, setAnswer] = useState(cached ?? '')
  const [modalOpen, setModalOpen] = useState(false)

  const generate = useCallback(() => {
    setState('loading')
    const q = `Analyze the ${area} layer signals and provide a consolidated recommendation. Findings:\n${findingsSummary}\n\nProvide: (1) summary of critical issues, (2) prioritized action plan with specific entity names, (3) expected impact.`
    advisorApi.ask(modelId, q, 'general')
      .then(resp => { aiCache.set(cacheKey, resp.answer); setAnswer(resp.answer); setState('done'); setModalOpen(true) })
      .catch(() => setState('error'))
  }, [area, modelId, findingsSummary, cacheKey])

  if (!aiEnabled) return null
  return (
    <>
      <div className="px-6 pb-5">
        {state === 'idle' && <button type="button" onClick={generate} className="inline-flex items-center gap-2 px-4 py-2.5 rounded-xl text-sm font-semibold bg-gradient-to-r from-indigo-500 to-purple-600 text-white hover:shadow-md hover:-translate-y-px transition-all"><Sparkles size={15} />AI Recommendation</button>}
        {state === 'loading' && <div className="rounded-xl p-4 flex items-center gap-3 bg-indigo-50 border border-indigo-200"><span className="w-4 h-4 border-2 border-indigo-200 border-t-indigo-600 rounded-full animate-spin shrink-0" /><span className="text-sm text-indigo-700 font-semibold">Generating {area} recommendation…</span></div>}
        {state === 'error' && <div className="rounded-xl p-4 flex items-center justify-between bg-red-50 border border-red-200"><span className="text-sm text-red-700">Failed to generate.</span><button type="button" onClick={generate} className="text-xs font-semibold text-indigo-600 underline">Retry</button></div>}
        {state === 'done' && <button type="button" onClick={() => setModalOpen(true)} className="inline-flex items-center gap-2 px-4 py-2.5 rounded-xl text-sm font-semibold bg-indigo-50 text-indigo-600 border border-indigo-200 hover:shadow-md hover:-translate-y-px transition-all"><Sparkles size={15} />View Recommendation</button>}
      </div>
      {modalOpen && state === 'done' && <AIModal area={area} answer={answer} onClose={() => setModalOpen(false)} />}
    </>
  )
}

// ── Layout ─────────────────────────────────────────────────────────────────────────

function HealthCard({ label, risk, icon: Icon, subtitle, findingsCount }: { label: string; risk: string; icon: React.ElementType; subtitle: string; findingsCount: number }) {
  const s = rs(risk)
  return (
    <div className="flex-1 relative overflow-hidden rounded-2xl hover:-translate-y-px transition-all duration-200" style={{ padding: '14px 16px', background: s.cardGradient, border: `1px solid ${s.border}` }}>
      <div className="flex items-center justify-between mb-2">
        <div className="rounded-xl p-2" style={{ background: s.iconWrap, border: `1px solid ${s.border}` }}><Icon size={16} style={{ color: s.dot }} /></div>
        <RiskPill risk={risk} />
      </div>
      <div className="font-bold text-sm text-slate-800">{label}</div>
      <div className="mt-1 flex items-baseline gap-1.5"><span className="text-xl font-extrabold text-slate-800">{findingsCount}</span><span className="text-[10px] font-semibold text-slate-500 uppercase tracking-wide">findings</span></div>
      <div className="text-xs mt-1 font-medium" style={{ color: s.text }}>{subtitle}</div>
    </div>
  )
}

function SectionCard({ title, risk, icon: Icon, children, empty, aiArea, modelId, findingsSummary, aiEnabled, highlighted, sectionRef }: {
  title: string; risk: string; icon: React.ElementType; children: React.ReactNode; empty: boolean
  aiArea?: string; modelId?: string; findingsSummary?: string; aiEnabled?: boolean
  highlighted?: boolean; sectionRef?: React.RefObject<HTMLDivElement | null>
}) {
  const s = rs(risk)
  return (
    <div ref={sectionRef} className="overflow-hidden rounded-2xl border border-slate-200 hover:shadow-md transition-all duration-200" style={highlighted ? { boxShadow: '0 0 0 3px #6366f1', borderColor: '#6366f1' } : undefined}>
      <div style={{ height: 3, background: s.stripGradient }} />
      <div className="flex items-center justify-between px-4 py-3 border-b border-slate-100">
        <div className="flex items-center gap-2.5 min-w-0">
          <div className="rounded-lg p-2 shrink-0" style={{ background: s.bg, border: `1px solid ${s.border}` }}><Icon size={14} style={{ color: s.dot }} /></div>
          <div className="min-w-0"><div className="text-sm font-bold text-slate-800 truncate">{title}</div><div className="text-[11px] text-slate-400">Click any row for details</div></div>
        </div>
        <RiskPill risk={risk} />
      </div>
      <div className="px-4 py-4">
        {empty
          ? <div className="flex items-center gap-2"><div className="rounded-lg p-1.5 bg-green-50 border border-green-200"><CheckCircle size={14} className="text-green-500" /></div><span className="text-xs font-semibold text-green-700">No risks detected</span></div>
          : <div className="space-y-5">{children}</div>}
      </div>
      {aiArea && modelId && findingsSummary && !empty && <AIRecommendationPanel area={aiArea} modelId={modelId} findingsSummary={findingsSummary} aiEnabled={!!aiEnabled} />}
    </div>
  )
}

function SubSection({ title, icon: Icon, count, children }: { title: string; icon: React.ElementType; count: number; children: React.ReactNode }) {
  return (
    <div>
      <div className="flex items-center gap-1.5 mb-2"><Icon size={12} className="text-slate-400" /><span className="text-[10px] font-semibold text-slate-500 uppercase tracking-wide">{title}</span><CountBadge n={count} /></div>
      <div className="space-y-2">{children}</div>
    </div>
  )
}

// ── Main View ─────────────────────────────────────────────────────────────────────

export function SignalsView() {
  const { modelId } = useModel()
  const aiEnabled = useAIEnabled()
  const [searchParams] = useSearchParams()
  const filterParam = searchParams.get('filter')
  const uxRef = useRef<HTMLDivElement | null>(null)
  const archRef = useRef<HTMLDivElement | null>(null)
  const orgRef = useRef<HTMLDivElement | null>(null)

  const { data, isLoading, error } = useQuery({
    queryKey: ['signalsView', modelId],
    queryFn: () => viewsApi.getSignalsView(modelId!),
    enabled: !!modelId,
  })

  const highlightedSection = !filterParam ? null
    : filterParam === 'needs-at-risk' ? 'ux'
    : filterParam === 'gap' || filterParam === 'fragmentation' ? 'arch'
    : filterParam === 'bottleneck' || filterParam === 'cognitive-load' ? 'org'
    : null

  useEffect(() => {
    if (!data || !highlightedSection) return
    const refMap = { ux: uxRef, arch: archRef, org: orgRef }
    const ref = refMap[highlightedSection as keyof typeof refMap]
    setTimeout(() => ref?.current?.scrollIntoView({ behavior: 'smooth', block: 'start' }), 100)
  }, [data, highlightedSection])

  if (isLoading) return <LoadingState message="Loading signals…" />
  if (error) return <ErrorState message={(error as Error).message} />
  if (!data) return null

  const ux = data.user_experience_layer
  const arch = data.architecture_layer
  const org = data.organization_layer
  const uxCount = ux.needs_requiring_3plus_teams.length + ux.needs_with_no_capability_backing.length + ux.needs_at_risk.length
  const archCount = arch.user_facing_caps_with_cross_team_services.length + arch.capabilities_not_connected_to_any_need.length + arch.capabilities_fragmented_across_teams.length
  const orgCount = org.top_teams_by_structural_load.length + org.critical_bottleneck_services.length + org.low_coherence_teams.length + (org.critical_external_deps?.length ?? 0)

  return (
    <ModelRequired>
      <div className="space-y-5">
        <PageHeader title="Architecture Signals" description="Health signals across UX, architecture and organisational layers — click any row for details" />

        <div className="flex gap-3 flex-wrap">
          <HealthCard label="User Experience" risk={data.health.ux_risk}          icon={Users}     subtitle={riskSubtitle(data.health.ux_risk, uxCount)}           findingsCount={uxCount} />
          <HealthCard label="Architecture"    risk={data.health.architecture_risk} icon={Layers}    subtitle={riskSubtitle(data.health.architecture_risk, archCount)} findingsCount={archCount} />
          <HealthCard label="Organization"    risk={data.health.org_risk}          icon={Building2} subtitle={riskSubtitle(data.health.org_risk, orgCount)}           findingsCount={orgCount} />
        </div>

        <div className="grid gap-4" style={{ gridTemplateColumns: 'repeat(auto-fit, minmax(340px, 1fr))', alignItems: 'start' }}>
          <SectionCard title="User Experience Layer" risk={data.health.ux_risk} icon={Users} empty={uxCount === 0}
            aiArea="User Experience" modelId={modelId ?? ''} findingsSummary={buildUxSummary(ux)} aiEnabled={aiEnabled}
            sectionRef={uxRef} highlighted={highlightedSection === 'ux'}>
            {ux.needs_requiring_3plus_teams.length > 0 && <SubSection title="Needs served by 3+ teams" icon={Zap} count={ux.needs_requiring_3plus_teams.length}>{ux.needs_requiring_3plus_teams.map((item, i) => <NeedRow key={i} item={item} variant="cross-team" />)}</SubSection>}
            {ux.needs_with_no_capability_backing.length > 0 && <SubSection title="Needs with no capability" icon={Link2Off} count={ux.needs_with_no_capability_backing.length}>{ux.needs_with_no_capability_backing.map((item, i) => <NeedRow key={i} item={item} variant="unbacked" />)}</SubSection>}
            {ux.needs_at_risk.length > 0 && <SubSection title="Needs at risk" icon={AlertTriangle} count={ux.needs_at_risk.length}>{ux.needs_at_risk.map((item, i) => <NeedRow key={i} item={item} variant="at-risk" />)}</SubSection>}
          </SectionCard>

          <SectionCard title="Architecture Layer" risk={data.health.architecture_risk} icon={Layers} empty={archCount === 0}
            aiArea="Architecture" modelId={modelId ?? ''} findingsSummary={buildArchSummary(arch)} aiEnabled={aiEnabled}
            sectionRef={archRef} highlighted={highlightedSection === 'arch'}>
            {arch.user_facing_caps_with_cross_team_services.length > 0 && <SubSection title="User-facing caps with cross-team services" icon={GitMerge} count={arch.user_facing_caps_with_cross_team_services.length}>{arch.user_facing_caps_with_cross_team_services.map((item, i) => <CapRow key={i} item={item} variant="cross-team" />)}</SubSection>}
            {arch.capabilities_not_connected_to_any_need.length > 0 && <SubSection title="Caps not connected to any need" icon={Link2Off} count={arch.capabilities_not_connected_to_any_need.length}>{arch.capabilities_not_connected_to_any_need.map((item, i) => <CapRow key={i} item={item} variant="unconnected" />)}</SubSection>}
            {arch.capabilities_fragmented_across_teams.length > 0 && <SubSection title="Caps fragmented across teams" icon={GitMerge} count={arch.capabilities_fragmented_across_teams.length}>{arch.capabilities_fragmented_across_teams.map((item, i) => <CapRow key={i} item={item} variant="fragmented" />)}</SubSection>}
          </SectionCard>

          <SectionCard title="Organization Layer" risk={data.health.org_risk} icon={Building2} empty={orgCount === 0}
            aiArea="Organization" modelId={modelId ?? ''} findingsSummary={buildOrgSummary(org)} aiEnabled={aiEnabled}
            sectionRef={orgRef} highlighted={highlightedSection === 'org'}>
            {org.top_teams_by_structural_load.length > 0 && <SubSection title="Teams under high structural load" icon={Zap} count={org.top_teams_by_structural_load.length}>{org.top_teams_by_structural_load.map((item, i) => <TeamRow key={i} item={item} variant="load" />)}</SubSection>}
            {org.critical_bottleneck_services.length > 0 && <SubSection title="Critical bottleneck services" icon={Server} count={org.critical_bottleneck_services.length}>{org.critical_bottleneck_services.map((item, i) => <ServiceRow key={i} item={item} />)}</SubSection>}
            {org.low_coherence_teams.length > 0 && <SubSection title="Low value stream coherence" icon={TrendingDown} count={org.low_coherence_teams.length}>{org.low_coherence_teams.map((item, i) => <TeamRow key={i} item={item} variant="coherence" />)}</SubSection>}
            {org.critical_external_deps && org.critical_external_deps.length > 0 && <SubSection title="External dependency concentration" icon={Link2} count={org.critical_external_deps.length}>{org.critical_external_deps.map((item, i) => <ExtDepRow key={i} item={item} />)}</SubSection>}
          </SectionCard>
        </div>
      </div>
    </ModelRequired>
  )
}
