import { useRef, useEffect, useState, useCallback, useEffect as useEff } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useSearchParams } from 'react-router-dom'
import { Users, Layers, Building2, AlertTriangle, Zap, Link2Off, GitMerge, Server, TrendingDown, Link2 } from 'lucide-react'
import { ChevronDown, ChevronUp, Info, Lightbulb, CheckCircle, Sparkles, X } from 'lucide-react'
import { ModelRequired } from '@/components/ui/ModelRequired'
import { Prose } from '@/components/ui/prose'
import { ContentContainer } from '@/components/ui/content-container'
import { PageHeader } from '@/components/ui/page-header'
import { LoadingState, ErrorState } from '@/components/ViewState'
import { useModel } from '@/lib/model-context'
import { useAIEnabled } from '@/hooks/useAIEnabled'
import { viewsApi, advisorApi } from '@/services/api'
import {
  rs,
  buildUxSummary, buildArchSummary, buildOrgSummary,
} from '@/features/signals/risk-config'
import type { SignalsNeedRisk, SignalsCapItem, SignalsTeamItem, SignalsServiceItem, SignalsExtDepItem } from '@/types/views'

function RiskPill({ risk }: { risk: string }) {
  const s = rs(risk)
  return <span className="inline-flex items-center gap-1 text-[10px] font-semibold px-1.5 py-0.5 rounded" style={{ background: s.badgeBg, color: s.text }}><span className="w-1.5 h-1.5 rounded-full shrink-0" style={{ background: s.dot }} />{s.label}</span>
}

function Tag({ text, color }: { text: string; color?: 'amber' | 'red' }) {
  const style = color === 'red' ? { background: '#fee2e2', color: '#b91c1c' } : color === 'amber' ? { background: '#fef3c7', color: '#92400e' } : { background: '#f1f5f9', color: '#64748b' }
  return <span className="inline-flex items-center text-[10px] font-semibold px-1.5 py-0.5 rounded" style={style}>{text}</span>
}

function ExpandableRow({ summary, explanation, suggestion }: { summary: React.ReactNode; explanation: string; suggestion: string }) {
  const [open, setOpen] = useState(false)
  return (
    <div className="rounded-lg border border-border overflow-hidden">
      <button type="button" onClick={() => setOpen(o => !o)} className="w-full flex items-center justify-between gap-2 px-3 py-2 text-left hover:bg-muted/50 transition-colors">
        <div className="flex-1 min-w-0">{summary}</div>
        <span className="shrink-0 text-muted-foreground">{open ? <ChevronUp className="w-3.5 h-3.5" /> : <ChevronDown className="w-3.5 h-3.5" />}</span>
      </button>
      {open && (
        <div className="px-3 pb-3 pt-1.5 space-y-2 bg-muted/30 border-t border-border">
          <div className="flex gap-2"><Info size={12} className="shrink-0 mt-0.5 text-muted-foreground" /><p className="text-xs text-muted-foreground">{explanation || 'No detail available'}</p></div>
          <div className="flex gap-2"><Lightbulb size={12} className="shrink-0 mt-0.5 text-sky-600" /><p className="text-xs text-sky-700">{suggestion || 'No detail available'}</p></div>
        </div>
      )}
    </div>
  )
}

function NeedRow({ item, variant }: { item: SignalsNeedRisk; variant: 'cross-team' | 'unbacked' | 'at-risk' }) {
  const explanation = variant === 'unbacked'
    ? `"${item.need_name}" (actor: ${item.actor_names?.join(' & ')}) has no capability mapped — no implementation path.`
    : `"${item.need_name}" requires ${item.team_span ?? 'multiple'} teams to coordinate.`
  const suggestion = variant === 'unbacked'
    ? 'Add a capability addressing this need, or retire it if no longer relevant.'
    : 'Restructure ownership so this need is served by a single stream-aligned team.'
  return (
    <ExpandableRow explanation={explanation} suggestion={suggestion} summary={
      <div className="space-y-1">
        <div className="text-sm font-semibold text-foreground">{item.need_name}</div>
        <div className="flex flex-wrap gap-1">{variant === 'unbacked' && <Tag text="No backing" color="red" />}{(item.team_span ?? 0) > 0 && <Tag text={`${item.team_span} teams`} color="amber" />}{item.teams?.slice(0, 3).map(t => <Tag key={t} text={t} />)}</div>
      </div>
    } />
  )
}

function CapRow({ item, variant }: { item: SignalsCapItem; variant: 'cross-team' | 'unconnected' | 'fragmented' }) {
  const explanation = variant === 'unconnected' ? `"${item.capability_name}" is not connected to any user need.`
    : variant === 'cross-team' ? `"${item.capability_name}" is user-facing but its services are owned by ${item.team_count ?? 'multiple'} teams.`
    : `"${item.capability_name}" is realised across ${item.team_count ?? 'multiple'} teams.`
  const suggestion = variant === 'unconnected' ? 'Identify the need it supports or remove it.'
    : variant === 'cross-team' ? 'Move services under a single team or extract shared logic.'
    : 'Consolidate ownership or split into sub-capabilities.'
  return (
    <ExpandableRow explanation={explanation} suggestion={suggestion} summary={
      <div className="space-y-1">
        <div className="text-sm font-semibold text-foreground">{item.capability_name}</div>
        <div className="flex flex-wrap gap-1">{item.visibility && <Tag text={item.visibility} />}{item.team_count != null && item.team_count > 0 && <Tag text={`${item.team_count} teams`} color="amber" />}</div>
      </div>
    } />
  )
}

function TeamRow({ item, variant }: { item: SignalsTeamItem; variant: 'load' | 'coherence' }) {
  const explanation = variant === 'load' ? `${item.team_name} owns ${item.capability_count ?? '?'} caps and ${item.service_count ?? '?'} svcs.`
    : `${item.team_name} has ${item.coherence_score != null ? Math.round(item.coherence_score * 100) : '?'}% coherence.`
  const suggestion = variant === 'load' ? 'Consider splitting this team or moving capabilities.'
    : 'Redistribute capabilities for coherent ownership.'
  return (
    <ExpandableRow explanation={explanation} suggestion={suggestion} summary={
      <div className="space-y-1">
        <div className="text-sm font-semibold text-foreground">{item.team_name}</div>
        <div className="flex flex-wrap gap-1">
          {variant === 'load' && item.capability_count != null && <Tag text={`${item.capability_count} caps`} />}
          {variant === 'load' && item.service_count != null && <Tag text={`${item.service_count} svcs`} />}
          {variant === 'load' && item.overall_level && <RiskPill risk={item.overall_level === 'high' ? 'red' : item.overall_level === 'medium' ? 'amber' : 'green'} />}
          {variant === 'coherence' && item.coherence_score != null && <Tag text={`${Math.round(item.coherence_score * 100)}% coherent`} color="red" />}
        </div>
      </div>
    } />
  )
}

function ServiceRow({ item }: { item: SignalsServiceItem }) {
  return (
    <ExpandableRow
      explanation={`${item.service_name} has ${item.fan_in} dependents.`}
      suggestion={`Consider splitting ${item.service_name} or introducing a facade.`}
      summary={<div className="space-y-1"><div className="text-sm font-semibold text-foreground">{item.service_name}</div><div className="flex gap-1"><Tag text={`${item.fan_in} dependents`} color="red" /></div></div>}
    />
  )
}

function ExtDepRow({ item }: { item: SignalsExtDepItem }) {
  return (
    <ExpandableRow
      explanation={`${item.dep_name} is used by ${item.service_count} services.`}
      suggestion={item.is_critical ? `Add an abstraction layer around ${item.dep_name}.` : `Monitor ${item.dep_name} closely.`}
      summary={<div className="space-y-1"><div className="text-sm font-semibold text-foreground">{item.dep_name}</div><div className="flex gap-1"><Tag text={`${item.service_count} services`} color={item.is_critical ? 'red' : 'amber'} /></div></div>}
    />
  )
}

const aiCache = new Map<string, string>()

function AIModal({ area, answer, onClose }: { area: string; answer: string; onClose: () => void }) {
  useEff(() => {
    const handler = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose() }
    document.addEventListener('keydown', handler)
    document.body.style.overflow = 'hidden'
    return () => { document.removeEventListener('keydown', handler); document.body.style.overflow = '' }
  }, [onClose])
  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center bg-black/40 backdrop-blur-sm" onClick={onClose}>
      <div className="relative w-full max-w-3xl my-8 mx-4 bg-card rounded-lg border border-border shadow-lg max-h-[calc(100vh-4rem)] flex flex-col" onClick={e => e.stopPropagation()}>
        <div className="flex items-center justify-between px-5 py-3 border-b border-border">
          <div className="flex items-center gap-2">
            <Sparkles size={14} className="text-sky-600" />
            <span className="text-sm font-semibold text-foreground">AI Recommendation — {area}</span>
          </div>
          <button type="button" onClick={onClose} className="p-1 rounded hover:bg-muted text-muted-foreground"><X size={16} /></button>
        </div>
        <div className="overflow-y-auto flex-1 px-5 py-4">
          <Prose>{answer}</Prose>
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
      <div className="px-3 pb-3">
        {state === 'idle' && <button type="button" onClick={generate} className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded text-xs font-medium bg-foreground text-background hover:opacity-90 transition-opacity"><Sparkles size={12} />AI Recommendation</button>}
        {state === 'loading' && <div className="rounded-md p-2.5 flex items-center gap-2 bg-sky-50 border border-sky-200"><span className="w-3 h-3 border-2 border-sky-200 border-t-sky-600 rounded-full animate-spin shrink-0" /><span className="text-xs text-sky-700">Generating…</span></div>}
        {state === 'error' && <div className="rounded-md p-2.5 flex items-center justify-between bg-red-50 border border-red-200"><span className="text-xs text-red-700">Failed.</span><button type="button" onClick={generate} className="text-xs font-medium text-sky-600 underline">Retry</button></div>}
        {state === 'done' && <button type="button" onClick={() => setModalOpen(true)} className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded text-xs font-medium bg-sky-50 text-sky-600 border border-sky-200 hover:bg-sky-100 transition-colors"><Sparkles size={12} />View Recommendation</button>}
      </div>
      {modalOpen && state === 'done' && <AIModal area={area} answer={answer} onClose={() => setModalOpen(false)} />}
    </>
  )
}

function HealthStat({ label, risk, icon: Icon, findingsCount }: { label: string; risk: string; icon: React.ElementType; findingsCount: number }) {
  const s = rs(risk)
  return (
    <div className="flex-1 rounded-lg border bg-card p-3">
      <div className="flex items-center justify-between mb-1.5">
        <Icon size={14} className="text-muted-foreground" />
        <RiskPill risk={risk} />
      </div>
      <div className="text-xs font-semibold text-foreground">{label}</div>
      <div className="flex items-baseline gap-1 mt-0.5">
        <span className="text-lg font-bold tabular-nums" style={{ color: s.dot }}>{findingsCount}</span>
        <span className="text-[10px] text-muted-foreground">findings</span>
      </div>
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
    <div ref={sectionRef} className="rounded-lg border border-border overflow-hidden" style={highlighted ? { boxShadow: '0 0 0 2px #6366f1', borderColor: '#6366f1' } : undefined}>
      <div className="flex items-center justify-between px-3 py-2.5 border-b border-border">
        <div className="flex items-center gap-2 min-w-0">
          <div className="rounded p-1.5 shrink-0" style={{ background: s.bg, border: `1px solid ${s.border}` }}><Icon size={12} style={{ color: s.dot }} /></div>
          <span className="text-sm font-semibold text-foreground truncate">{title}</span>
        </div>
        <RiskPill risk={risk} />
      </div>
      <div className="px-3 py-3">
        {empty
          ? <div className="flex items-center gap-2"><CheckCircle size={14} className="text-green-500" /><span className="text-xs font-medium text-green-700">No risks detected</span></div>
          : <div className="space-y-3">{children}</div>}
      </div>
      {aiArea && modelId && findingsSummary && !empty && <AIRecommendationPanel area={aiArea} modelId={modelId} findingsSummary={findingsSummary} aiEnabled={!!aiEnabled} />}
    </div>
  )
}

function SubSection({ title, icon: Icon, count, children }: { title: string; icon: React.ElementType; count: number; children: React.ReactNode }) {
  return (
    <div>
      <div className="flex items-center gap-1.5 mb-1.5">
        <Icon size={11} className="text-muted-foreground" />
        <span className="text-[10px] font-semibold text-muted-foreground uppercase tracking-wide">{title}</span>
        <span className="text-[10px] font-semibold text-muted-foreground bg-muted rounded px-1.5 py-0.5">{count}</span>
      </div>
      <div className="space-y-1.5">{children}</div>
    </div>
  )
}

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
      <ContentContainer className="space-y-4">
        <PageHeader title="Architecture Signals" description="Health signals across UX, architecture and organisational layers" />

        <div className="flex gap-3">
          <HealthStat label="User Experience" risk={data.health.ux_risk}          icon={Users}     findingsCount={uxCount} />
          <HealthStat label="Architecture"    risk={data.health.architecture_risk} icon={Layers}    findingsCount={archCount} />
          <HealthStat label="Organization"    risk={data.health.org_risk}          icon={Building2} findingsCount={orgCount} />
        </div>

        <div className="grid gap-3 grid-cols-1 lg:grid-cols-3" style={{ alignItems: 'start' }}>
          <SectionCard title="User Experience" risk={data.health.ux_risk} icon={Users} empty={uxCount === 0}
            aiArea="User Experience" modelId={modelId ?? ''} findingsSummary={buildUxSummary(ux)} aiEnabled={aiEnabled}
            sectionRef={uxRef} highlighted={highlightedSection === 'ux'}>
            {ux.needs_requiring_3plus_teams.length > 0 && <SubSection title="3+ teams" icon={Zap} count={ux.needs_requiring_3plus_teams.length}>{ux.needs_requiring_3plus_teams.map((item, i) => <NeedRow key={i} item={item} variant="cross-team" />)}</SubSection>}
            {ux.needs_with_no_capability_backing.length > 0 && <SubSection title="No capability" icon={Link2Off} count={ux.needs_with_no_capability_backing.length}>{ux.needs_with_no_capability_backing.map((item, i) => <NeedRow key={i} item={item} variant="unbacked" />)}</SubSection>}
            {ux.needs_at_risk.length > 0 && <SubSection title="At risk" icon={AlertTriangle} count={ux.needs_at_risk.length}>{ux.needs_at_risk.map((item, i) => <NeedRow key={i} item={item} variant="at-risk" />)}</SubSection>}
          </SectionCard>

          <SectionCard title="Architecture" risk={data.health.architecture_risk} icon={Layers} empty={archCount === 0}
            aiArea="Architecture" modelId={modelId ?? ''} findingsSummary={buildArchSummary(arch)} aiEnabled={aiEnabled}
            sectionRef={archRef} highlighted={highlightedSection === 'arch'}>
            {arch.user_facing_caps_with_cross_team_services.length > 0 && <SubSection title="Cross-team user-facing" icon={GitMerge} count={arch.user_facing_caps_with_cross_team_services.length}>{arch.user_facing_caps_with_cross_team_services.map((item, i) => <CapRow key={i} item={item} variant="cross-team" />)}</SubSection>}
            {arch.capabilities_not_connected_to_any_need.length > 0 && <SubSection title="No need connected" icon={Link2Off} count={arch.capabilities_not_connected_to_any_need.length}>{arch.capabilities_not_connected_to_any_need.map((item, i) => <CapRow key={i} item={item} variant="unconnected" />)}</SubSection>}
            {arch.capabilities_fragmented_across_teams.length > 0 && <SubSection title="Fragmented" icon={GitMerge} count={arch.capabilities_fragmented_across_teams.length}>{arch.capabilities_fragmented_across_teams.map((item, i) => <CapRow key={i} item={item} variant="fragmented" />)}</SubSection>}
          </SectionCard>

          <SectionCard title="Organization" risk={data.health.org_risk} icon={Building2} empty={orgCount === 0}
            aiArea="Organization" modelId={modelId ?? ''} findingsSummary={buildOrgSummary(org)} aiEnabled={aiEnabled}
            sectionRef={orgRef} highlighted={highlightedSection === 'org'}>
            {org.top_teams_by_structural_load.length > 0 && <SubSection title="High structural load" icon={Zap} count={org.top_teams_by_structural_load.length}>{org.top_teams_by_structural_load.map((item, i) => <TeamRow key={i} item={item} variant="load" />)}</SubSection>}
            {org.critical_bottleneck_services.length > 0 && <SubSection title="Bottleneck services" icon={Server} count={org.critical_bottleneck_services.length}>{org.critical_bottleneck_services.map((item, i) => <ServiceRow key={i} item={item} />)}</SubSection>}
            {org.low_coherence_teams.length > 0 && <SubSection title="Low coherence" icon={TrendingDown} count={org.low_coherence_teams.length}>{org.low_coherence_teams.map((item, i) => <TeamRow key={i} item={item} variant="coherence" />)}</SubSection>}
            {org.critical_external_deps && org.critical_external_deps.length > 0 && <SubSection title="Ext dep concentration" icon={Link2} count={org.critical_external_deps.length}>{org.critical_external_deps.map((item, i) => <ExtDepRow key={i} item={item} />)}</SubSection>}
          </SectionCard>
        </div>
      </ContentContainer>
    </ModelRequired>
  )
}
