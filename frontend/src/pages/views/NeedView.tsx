import { useEffect, useState } from 'react'
import { api } from '@/lib/api'
import { useRequireModel } from '@/lib/model-context'
import { ModelRequired } from '@/components/ui/ModelRequired'
import { useSearch, matchesQuery } from '@/lib/search-context'
import { usePageInsights, type InsightStatus } from '@/hooks/usePageInsights'
import { Users, Layers, AlertTriangle, ChevronDown, ChevronUp, Info, Lightbulb, Sparkles } from 'lucide-react'

const slug = (s: string) => s.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '')

const VIS_BADGE: Record<string, { bg: string; text: string }> = {
  'user-facing':    { bg: '#dbeafe', text: '#1e40af' },
  'domain':         { bg: '#ede9fe', text: '#5b21b6' },
  'foundational':   { bg: '#d1fae5', text: '#065f46' },
  'infrastructure': { bg: '#f1f5f9', text: '#475569' },
}

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

const pill: React.CSSProperties = {
  borderRadius: 8,
  padding: '3px 9px',
  fontSize: 11,
  fontWeight: 600,
  display: 'inline-flex',
  alignItems: 'center',
}

const spinnerEl = (
  <span
    className="animate-spin flex-shrink-0"
    style={{ width: 20, height: 20, border: '2px solid #e2e8f0', borderTopColor: '#6366f1', borderRadius: '50%' }}
  />
)

interface NeedViewCapability {
  id: string
  label: string
  data: { visibility: string }
}

interface NeedViewNeed {
  need: {
    id: string
    label: string
    data: {
      is_mapped: boolean
      outcome?: string
      anti_patterns?: Array<{ code: string; message: string; severity: string }>
      team_span?: number
      teams?: string[]
      at_risk?: boolean
      unbacked?: boolean
    }
  }
  capabilities: NeedViewCapability[]
}

interface NeedViewGroup {
  actor: { id: string; label: string }
  needs: NeedViewNeed[]
}

interface NeedViewResponse {
  view_type: string
  total_needs: number
  unmapped_count: number
  groups: NeedViewGroup[]
}

// Derive the human-readable reason for at_risk.
// at_risk fires when team_span >= 3 OR a delivery team has high cognitive load.
function atRiskReason(teamSpan: number): string {
  if (teamSpan >= 3) {
    return `Spans ${teamSpan} teams — coordinating delivery across this many teams introduces handoff overhead and release coordination risk.`
  }
  if (teamSpan > 0) {
    return `A team in the delivery chain is under high cognitive load, making reliable delivery of this need more fragile.`
  }
  return `Flagged at risk — a team in the delivery chain has high cognitive load or the need spans too many team boundaries.`
}

// ── Stat card ────────────────────────────────────────────────────────────────

function StatCard({ value, label, icon: Icon, gradient, iconTint }: {
  value: React.ReactNode; label: string; icon: React.ElementType; gradient: string; iconTint: string
}) {
  return (
    <div className="flex-1 min-w-[140px] relative overflow-hidden p-5 transition-all duration-200 ease-out hover:-translate-y-px"
      style={{ borderRadius: 20, background: gradient, border: '1px solid #e2e8f0', boxShadow: '0 4px 14px rgba(15,23,42,0.06)' }}>
      <div className="flex items-center gap-3 mb-3">
        <div className="rounded-xl p-2.5 flex items-center justify-center"
          style={{ background: 'rgba(255,255,255,0.75)', border: '1px solid rgba(226,232,240,0.9)', boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
          <Icon size={20} style={{ color: iconTint }} strokeWidth={2.25} />
        </div>
      </div>
      <div className="tabular-nums" style={{ fontSize: 26, fontWeight: 800, color: '#1e293b' }}>{value}</div>
      <div className="mt-1 uppercase font-semibold" style={{ fontSize: 11, fontWeight: 600, color: '#64748b', letterSpacing: '0.05em' }}>{label}</div>
    </div>
  )
}

// ── Expandable need row ──────────────────────────────────────────────────────

function NeedRow({ nr, isLast, insights, aiStatus }: {
  nr: NeedViewNeed
  isLast: boolean
  insights: Record<string, { explanation: string; suggestion: string }>
  aiStatus: InsightStatus
}) {
  const [open, setOpen] = useState(false)
  const { need, capabilities } = nr
  const d = need.data
  const isMapped = d.is_mapped !== false
  const teamSpan = d.team_span ?? 0
  const aiInsight = insights[`need:${slug(need.label)}`]

  return (
    <>
      {/* Summary row — always visible, clickable */}
      <tr
        className="transition-colors duration-150 cursor-pointer select-none"
        style={{ background: open ? '#f0f7ff' : undefined }}
        onClick={() => setOpen(o => !o)}
        onMouseEnter={e => { if (!open) (e.currentTarget as HTMLElement).style.background = '#f8fafc' }}
        onMouseLeave={e => { if (!open) (e.currentTarget as HTMLElement).style.background = '' }}
      >
        {/* Need / outcome */}
        <td className="px-4 sm:px-5 py-3.5 align-top" style={{ borderBottom: open || !isLast ? '1px solid #f1f5f9' : 'none', borderRight: '1px solid #f1f5f9' }}>
          <div className="flex items-start gap-2">
            <div className="flex-1 min-w-0">
              <span className="font-semibold text-sm" style={{ color: '#1e293b' }}>{need.label}</span>
              {d.outcome && (
                <div style={{fontSize:12, color:'#6b7280', marginTop:2,
                  overflow:'hidden', textOverflow:'ellipsis', whiteSpace:'nowrap', maxWidth:400}}
                  title={d.outcome}>
                  {d.outcome}
                </div>
              )}
              <div className="flex flex-wrap gap-1.5 mt-1.5">
                {d.unbacked && (
                  <span style={{ ...pill, background: '#fef3c7', color: '#92400e', cursor: 'help' }}
                    title="Unbacked: no capability has services backing this need"
                    aria-label="Need is unbacked">Unbacked</span>
                )}
                {d.at_risk && !d.unbacked && (
                  <span style={{ ...pill, background: '#fee2e2', color: '#b91c1c', cursor: 'help' }}
                    title={atRiskReason(teamSpan)}
                    aria-label="Need is at risk">
                    At risk · {teamSpan} team{teamSpan !== 1 ? 's' : ''}
                  </span>
                )}
                {!d.unbacked && !d.at_risk && teamSpan > 1 && (
                  <span style={{ ...pill, background: '#f1f5f9', color: '#475569' }}>
                    {teamSpan} teams
                  </span>
                )}
              </div>
            </div>
            <div className="flex-shrink-0 mt-0.5" style={{ color: '#94a3b8' }}>
              {open ? <ChevronUp size={13} /> : <ChevronDown size={13} />}
            </div>
          </div>
        </td>

        {/* Capabilities */}
        <td className="px-4 sm:px-5 py-3.5 align-top" style={{ borderBottom: open || !isLast ? '1px solid #f1f5f9' : 'none', borderRight: '1px solid #f1f5f9' }}>
          {capabilities.length === 0 ? (
            <span className="text-xs italic" style={{ color: '#94a3b8' }}>no capabilities linked</span>
          ) : (
            <div className="flex flex-wrap gap-1.5">
              {capabilities.map(cap => {
                const badge = VIS_BADGE[cap.data.visibility ?? ''] ?? { bg: '#f1f5f9', text: '#475569' }
                return (
                  <span key={cap.id} className="font-semibold" style={{ ...pill, background: badge.bg, color: badge.text }}>
                    {cap.label}
                  </span>
                )
              })}
            </div>
          )}
        </td>

        {/* Status */}
        <td className="px-4 sm:px-5 py-3.5 align-top" style={{ borderBottom: open || !isLast ? '1px solid #f1f5f9' : 'none' }}>
          {isMapped ? (
            <span className="font-semibold" style={{ ...pill, background: '#dcfce7', color: '#15803d' }}>Mapped</span>
          ) : (
            <span className="font-semibold" style={{ ...pill, background: '#fee2e2', color: '#b91c1c' }}>Unmapped</span>
          )}
        </td>
      </tr>

      {/* Expanded detail row */}
      {open && (
        <tr>
          <td colSpan={3} style={{ borderBottom: isLast ? 'none' : '1px solid #f1f5f9', padding: 0 }}>
            <div className="px-5 py-4 space-y-4" style={{ background: '#f8fafc', borderTop: '1px solid #e2e8f0' }}>

              {/* Outcome */}
              {d.outcome && (
                <p className="text-xs leading-relaxed" style={{ color: '#475569' }}>{d.outcome}</p>
              )}

              {/* Why at risk */}
              {d.at_risk && (
                <div className="flex gap-2 rounded-lg px-3 py-2.5" style={{ background: '#fff1f2', border: '1px solid #fecdd3' }}>
                  <span title={`At risk: ${atRiskReason(teamSpan)}`} aria-label="At risk warning" style={{display:'inline-flex'}}>
                    <AlertTriangle size={13} style={{ color: '#e11d48', flexShrink: 0, marginTop: 1 }} />
                  </span>
                  <p className="text-xs leading-relaxed" style={{ color: '#9f1239' }}>
                    <strong>Why at risk: </strong>{atRiskReason(teamSpan)}
                  </p>
                </div>
              )}

              {/* Delivery teams */}
              {(d.teams?.length ?? 0) > 0 && (
                <div>
                  <p className="text-xs font-semibold mb-1.5" style={{ color: '#64748b' }}>Delivery teams</p>
                  <div className="flex flex-wrap gap-1.5">
                    {d.teams!.map(t => (
                      <span key={t} style={{ ...pill, background: '#f1f5f9', color: '#475569', fontSize: 11 }}>{t}</span>
                    ))}
                  </div>
                </div>
              )}

              {/* AI insight */}
              {aiStatus === 'loading' ? (
                <div className="flex items-center gap-2 text-xs" style={{ color: '#94a3b8' }}>
                  <span className="w-3 h-3 border-2 border-current border-t-transparent rounded-full animate-spin flex-shrink-0" />
                  Loading AI insight…
                </div>
              ) : aiStatus === 'error' ? (
                <div className="text-xs" style={{ color: '#94a3b8' }}>
                  AI insight unavailable
                </div>
              ) : aiInsight ? (
                <div className="space-y-2 rounded-lg px-3 py-3" style={{ background: '#f0f9ff', border: '1px solid #bae6fd' }}>
                  <div className="flex items-center gap-1.5 mb-2">
                    <Sparkles size={12} style={{ color: '#0284c7' }} />
                    <span className="text-xs font-semibold" style={{ color: '#0284c7' }}>AI Insight</span>
                  </div>
                  <div className="flex gap-2">
                    <Info size={12} className="flex-shrink-0 mt-0.5" style={{ color: '#0369a1' }} />
                    <p className="text-xs leading-relaxed" style={{ color: '#0c4a6e' }}>{aiInsight.explanation}</p>
                  </div>
                  <div className="flex gap-2">
                    <Lightbulb size={12} className="flex-shrink-0 mt-0.5" style={{ color: '#0369a1' }} />
                    <p className="text-xs leading-relaxed" style={{ color: '#0c4a6e' }}>{aiInsight.suggestion}</p>
                  </div>
                </div>
              ) : null}
            </div>
          </td>
        </tr>
      )}
    </>
  )
}

// ── Main view ────────────────────────────────────────────────────────────────

export function NeedView() {
  const { modelId, isHydrating } = useRequireModel()
  const { query } = useSearch()
  const [viewData, setViewData] = useState<NeedViewResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const { insights, status: aiStatus } = usePageInsights('needs')

  useEffect(() => {
    if (isHydrating || !modelId) return
    api.getView(modelId, 'need')
      .then(data => setViewData(data as unknown as NeedViewResponse))
      .catch(e => setError((e as Error).message))
      .finally(() => setLoading(false))
  }, [isHydrating, modelId])

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center gap-3 h-full min-h-[240px]">
        {spinnerEl}
        <span style={{ fontSize: 14, color: '#94a3b8', fontWeight: 500 }}>Loading need view…</span>
      </div>
    )
  }
  if (error) {
    return (
      <div className="flex flex-col items-center justify-center gap-3 h-full min-h-[240px] px-4">
        <div className="flex items-center gap-3 rounded-2xl px-5 py-4 max-w-md w-full" style={{ ...cardShell, borderColor: '#fecaca' }}>
          <div className="rounded-xl p-2 flex-shrink-0" style={{ background: '#fee2e2' }}>
            <span title="Error loading need view" aria-label="Error">
              <AlertTriangle size={20} style={{ color: '#dc2626' }} />
            </span>
          </div>
          <span className="text-sm font-medium" style={{ color: '#b91c1c' }}>{error}</span>
        </div>
      </div>
    )
  }
  if (!viewData) return null

  // Compute stats for UI-23
  const totalNeeds = viewData.total_needs
  const unmappedCount = viewData.unmapped_count
  const unmappedPct = totalNeeds > 0 ? Math.round((unmappedCount / totalNeeds) * 100) : 0
  const atRiskCount = viewData.groups.flatMap(g => g.needs).filter(nr => nr.need.data.at_risk).length
  const atRiskPct = totalNeeds > 0 ? Math.round((atRiskCount / totalNeeds) * 100) : 0

  // Compute visibility counts for UI-24 — based on capabilities linked to needs
  const visCounts: Record<string, number> = { 'user-facing': 0, 'domain': 0, 'foundational': 0, 'infrastructure': 0 }
  const allNeeds = viewData.groups.flatMap(g => g.needs)
  for (const nr of allNeeds) {
    const visLevels = new Set(nr.capabilities.map(c => c.data.visibility))
    for (const v of visLevels) {
      if (v in visCounts) visCounts[v]++
    }
  }

  const filtered = viewData.groups
    .map(g => ({
      ...g,
      needs: g.needs.filter(nr =>
        !query ||
        matchesQuery(g.actor.label, query) ||
        matchesQuery(nr.need.label, query) ||
        nr.capabilities.some(c => matchesQuery(c.label, query))
      ),
    }))
    .filter(g => !query || g.needs.length > 0 || matchesQuery(g.actor.label, query))

  return (
    <ModelRequired>
      <div className="space-y-8">
      {/* Header */}
      <div className="flex flex-col gap-6 lg:flex-row lg:items-start lg:justify-between">
        <div>
          <h1 style={gradientTitle}>Need View</h1>
          <p style={{ fontSize: 14, color: '#64748b', marginTop: 6 }}>
            {viewData.total_needs} needs across {viewData.groups.length} actors
            {viewData.unmapped_count > 0 && (
              <span className="ml-2 inline-flex items-center" style={{ ...pill, background: '#fee2e2', color: '#b91c1c' }}>
                {viewData.unmapped_count} unmapped
              </span>
            )}
          </p>
        </div>
        <div className="flex flex-wrap gap-2 lg:justify-end lg:max-w-xl">
          {Object.entries(VIS_BADGE).map(([k, s]) => (
            <span key={k} className="inline-flex items-center capitalize" style={{ ...pill, background: s.bg, color: s.text }}>
              {k.replace(/-/g, ' ')} ({visCounts[k] ?? 0})
            </span>
          ))}
        </div>
      </div>

      {/* Stats */}
      <div className="grid gap-4" style={{ gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))' }}>
        <StatCard value={viewData.total_needs} label="Total needs" icon={Layers}
          gradient="linear-gradient(135deg, #ede9fe 0%, #e0e7ff 100%)" iconTint="#5b21b6" />
        <StatCard value={viewData.groups.length} label="Actors" icon={Users}
          gradient="linear-gradient(135deg, #dbeafe 0%, #e0f2fe 100%)" iconTint="#1d4ed8" />
        <StatCard
          value={<>{unmappedCount} <span style={{color:'#9ca3af', fontWeight:400, fontSize:16}}>({unmappedPct}%)</span></>}
          label="Unmapped" icon={AlertTriangle}
          gradient="linear-gradient(135deg, #ffedd5 0%, #fef3c7 100%)" iconTint="#c2410c" />
        <StatCard
          value={<>{atRiskCount} <span style={{color:'#9ca3af', fontWeight:400, fontSize:16}}>({atRiskPct}%)</span></>}
          label="At risk" icon={AlertTriangle}
          gradient="linear-gradient(135deg, #fee2e2 0%, #fecdd3 100%)" iconTint="#dc2626" />
      </div>

      {/* Actor groups */}
      <div className="space-y-6">
        {filtered.map(({ actor, needs }) => (
          <div key={actor.id} className="overflow-hidden transition-all duration-200 ease-out hover:shadow-md" style={cardShell}>
            <div style={{ height: 3, background: 'linear-gradient(90deg, #6366f1 0%, #a855f7 50%, #c084fc 100%)' }} />

            {/* Actor header */}
            <div className="px-6 py-4 flex items-center justify-between gap-3 flex-wrap"
              style={{ background: 'linear-gradient(135deg, #faf5ff 0%, #f5f3ff 45%, #ffffff 100%)', borderBottom: '1px solid #f1f5f9' }}>
              <div className="min-w-0">
                <div className="font-bold truncate" style={{ fontSize: 16, fontWeight: 700, color: '#1e293b' }}>{actor.label}</div>
                <div className="mt-0.5" style={{ fontSize: 12, color: '#94a3b8' }}>
                  Actor · {needs.length} need{needs.length === 1 ? '' : 's'} in view
                </div>
              </div>
              <span className="inline-flex items-center flex-shrink-0" style={{ ...pill, background: '#ede9fe', color: '#5b21b6' }}>
                {needs.length} needs
              </span>
            </div>

            {/* Needs table — click any row to expand */}
            <div className="p-4 sm:p-5" style={{ background: 'linear-gradient(180deg, #ffffff 0%, #fafafa 100%)' }}>
              <div className="overflow-hidden" style={{ borderRadius: 16, border: '1px solid #e2e8f0' }}>
                <table className="w-full text-sm" style={{ background: '#ffffff' }}>
                  <thead>
                    <tr style={{ background: '#f8fafc' }}>
                      {[['Need / Outcome', '35%'], ['Supported by capabilities', ''], ['Status', '100px']].map(([label, width]) => (
                        <th key={label} className="px-4 sm:px-5 py-3 text-left"
                          style={{ fontSize: 11, fontWeight: 600, color: '#64748b', textTransform: 'uppercase',
                            letterSpacing: '0.05em', width: width || undefined, borderBottom: '1px solid #f1f5f9',
                            borderRight: label !== 'Status' ? '1px solid #f1f5f9' : undefined }}>
                          {label}
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    {needs.length === 0 ? (
                      <tr>
                        <td colSpan={3} className="px-5 py-5 text-sm italic" style={{ color: '#94a3b8' }}>
                          No needs match this filter
                        </td>
                      </tr>
                    ) : needs.map((nr, idx) => (
                      <NeedRow
                        key={nr.need.id}
                        nr={nr}
                        isLast={idx === needs.length - 1}
                        insights={insights}
                        aiStatus={aiStatus}
                      />
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
    </ModelRequired>
  )
}
