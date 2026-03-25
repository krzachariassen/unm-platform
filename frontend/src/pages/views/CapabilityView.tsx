import { useEffect, useMemo, useState, type ReactNode } from 'react'
import { api } from '@/lib/api'
import { useRequireModel } from '@/lib/model-context'
import { useSearch, matchesQuery } from '@/lib/search-context'
import { usePageInsights } from '@/hooks/usePageInsights'
import { QuickAction } from '@/components/changeset/QuickAction'
import type { InsightItem } from '@/lib/api'

const slug = (s: string) => s.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '')

const VIS_BANDS = [
  { key: 'user-facing',    label: 'User-facing',    accent: '#2563eb', border: '#bfdbfe', bg: '#eff6ff' },
  { key: 'domain',         label: 'Domain',         accent: '#7c3aed', border: '#ddd6fe', bg: '#f5f3ff' },
  { key: 'foundational',   label: 'Foundational',   accent: '#059669', border: '#a7f3d0', bg: '#f0fdf4' },
  { key: 'infrastructure', label: 'Infrastructure', accent: '#6b7280', border: '#e5e7eb', bg: '#f9fafb' },
]

const CARD_SHELL = {
  borderRadius: 20,
  background: 'linear-gradient(135deg, #ffffff 0%, #f8fafc 100%)',
  border: '1px solid #e2e8f0',
  boxShadow: '0 1px 3px rgba(0,0,0,0.04)',
} as const

const H1_GRADIENT = {
  fontSize: 30,
  fontWeight: 800,
  letterSpacing: '-0.025em',
  lineHeight: 1.15,
  background: 'linear-gradient(135deg, #1e293b 0%, #475569 100%)',
  WebkitBackgroundClip: 'text' as const,
  WebkitTextFillColor: 'transparent' as const,
}

const SUBTITLE = { fontSize: 14, color: '#64748b' } as const

const SECTION_LABEL = {
  fontSize: 11,
  fontWeight: 600,
  color: '#64748b',
  textTransform: 'uppercase' as const,
  letterSpacing: '0.05em',
} as const

interface CapabilityType {
  id: string
  label: string
  description: string
  visibility: string
  is_leaf: boolean
  is_fragmented: boolean
  depended_on_by_count: number
  services: Array<{ id: string; label: string; cap_count: number }>
  teams: Array<{ id: string; label: string; type: string }>
  depends_on: Array<{ id: string; label: string }>
  children: Array<{ id: string; label: string }>
  anti_patterns?: Array<{ code: string; message: string; severity: string }>
  external_deps?: Array<{ name: string; description?: string }>
}

interface CapabilityViewResponse {
  view_type: string
  leaf_capability_count: number
  high_span_services: Array<{ name: string; capability_count: number }>
  fragmented_capabilities: Array<{ id: string; label: string; team_count: number }>
  parent_groups: Array<{
    id: string
    label: string
    children: string[]
  }>
  capabilities: CapabilityType[]
}

function StatCard({
  value,
  label,
  gradient,
  iconBg,
  icon,
}: {
  value: number
  label: string
  gradient: string
  iconBg: string
  icon: ReactNode
}) {
  return (
    <div
      className="relative overflow-hidden"
      style={{
        borderRadius: 20,
        background: gradient,
        border: '1px solid rgba(255,255,255,0.35)',
        boxShadow: '0 1px 3px rgba(0,0,0,0.06), 0 8px 24px rgba(15,23,42,0.06)',
        padding: '16px 18px',
      }}
    >
      <div className="flex items-start justify-between gap-3">
        <div>
          <div className="tabular-nums" style={{ fontSize: 26, fontWeight: 800, color: '#0f172a' }}>
            {value}
          </div>
          <div style={{ fontSize: 11, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.05em', color: 'rgba(15,23,42,0.55)', marginTop: 4 }}>
            {label}
          </div>
        </div>
        <div
          className="flex items-center justify-center flex-shrink-0"
          style={{
            width: 40,
            height: 40,
            borderRadius: 12,
            background: iconBg,
            boxShadow: '0 1px 2px rgba(0,0,0,0.06)',
          }}
        >
          {icon}
        </div>
      </div>
    </div>
  )
}

function PillToggle({
  value,
  onChange,
  options,
}: {
  value: 'visibility' | 'domain' | 'team'
  onChange: (v: 'visibility' | 'domain' | 'team') => void
  options: { key: 'visibility' | 'domain' | 'team'; label: string }[]
}) {
  return (
    <div
      className="inline-flex gap-0.5 p-1"
      style={{
        borderRadius: 12,
        padding: 4,
        background: '#f1f5f9',
        border: '1px solid #e2e8f0',
      }}
    >
      {options.map(opt => {
        const active = value === opt.key
        return (
          <button
            key={opt.key}
            type="button"
            onClick={() => onChange(opt.key)}
            className="px-3 py-2 text-xs font-semibold transition-all"
            style={{
              borderRadius: 8,
              background: active ? 'linear-gradient(135deg, #6366f1 0%, #4f46e5 100%)' : 'transparent',
              color: active ? '#ffffff' : '#64748b',
              boxShadow: active ? '0 2px 8px rgba(99,102,241,0.35)' : 'none',
            }}
          >
            {opt.label}
          </button>
        )
      })}
    </div>
  )
}

// A lean card: name + services + one signal flag only.
// Everything else lives in the detail panel.
function CapabilityCard({
  cap,
  band,
  capById,
  onClick,
}: {
  cap: CapabilityType
  band?: typeof VIS_BANDS[0]
  capById: Map<string, CapabilityType>
  onClick: () => void
}) {
  void capById
  const isBottleneck = cap.depended_on_by_count >= 5
  const isFragmented = cap.is_fragmented
  const accent = isFragmented ? '#ef4444' : (band?.accent ?? '#94a3b8')
  const accentSoft = isFragmented ? '#fca5a5' : (band?.border ?? '#cbd5e1')

  return (
    <div
      onClick={onClick}
      title={cap.description || cap.label}
      className="cursor-pointer transition-all"
      style={{
        borderRadius: 14,
        border: '1px solid #e2e8f0',
        background: 'linear-gradient(135deg, #ffffff 0%, #f8fafc 100%)',
        boxShadow: '0 1px 3px rgba(0,0,0,0.04)',
        overflow: 'hidden',
      }}
      onMouseEnter={e => {
        e.currentTarget.style.transform = 'translateY(-1px)'
        e.currentTarget.style.boxShadow = '0 8px 24px rgba(15,23,42,0.08)'
        const btn = e.currentTarget.querySelector('.qa-cap') as HTMLElement; if (btn) btn.style.opacity = '1'
      }}
      onMouseLeave={e => {
        e.currentTarget.style.transform = 'translateY(0)'
        e.currentTarget.style.boxShadow = '0 1px 3px rgba(0,0,0,0.04)'
        const btn = e.currentTarget.querySelector('.qa-cap') as HTMLElement; if (btn) btn.style.opacity = '0.35'
      }}
    >
      <div className="flex" style={{ minHeight: '100%' }}>
        <div
          style={{
            width: 3,
            flexShrink: 0,
            background: isFragmented
              ? 'linear-gradient(180deg, #ef4444 0%, #f87171 100%)'
              : `linear-gradient(180deg, ${accent} 0%, ${accentSoft} 100%)`,
          }}
        />
        <div className="flex-1 px-3 py-2.5">
          <div className="flex items-start justify-between gap-2">
            <span className="text-sm font-semibold leading-snug" style={{ color: '#0f172a' }}>
              {cap.label}
            </span>
            <span className="qa-cap" style={{ opacity: 0.35, transition: 'opacity 0.15s' }}>
              <QuickAction size={11} options={[
                { label: 'Update visibility', action: { type: 'update_capability_visibility', capability_name: cap.label } },
                { label: 'Reassign to another team', action: { type: 'reassign_capability', capability_name: cap.label } },
              ]} />
            </span>
            <div className="flex items-center gap-1 flex-shrink-0">
              {isFragmented && (
                <span
                  className="font-semibold"
                  title={`Fragmented: this capability is owned by multiple teams`}
                  aria-label="Capability is fragmented"
                  style={{
                    fontSize: 11,
                    fontWeight: 600,
                    padding: '2px 8px',
                    borderRadius: 6,
                    background: '#fee2e2',
                    color: '#b91c1c',
                    cursor: 'help',
                  }}
                >
                  split
                </span>
              )}
              {isBottleneck && !isFragmented && (
                <span
                  className="font-semibold"
                  title={`${cap.depended_on_by_count} other capabilities depend on this capability`}
                  aria-label={`${cap.depended_on_by_count} dependents`}
                  style={{
                    fontSize: 11,
                    fontWeight: 600,
                    padding: '2px 8px',
                    borderRadius: 6,
                    background: '#fef3c7',
                    color: '#92400e',
                    cursor: 'help',
                  }}
                >
                  ↙{cap.depended_on_by_count}
                </span>
              )}
            </div>
          </div>

          {cap.services.length > 0 && (
            <div className="flex flex-wrap gap-1 mt-2">
              {cap.services.map(svc => (
                <span
                  key={svc.id}
                  className="font-mono"
                  style={{
                    fontSize: 11,
                    fontWeight: 500,
                    borderRadius: 6,
                    padding: '3px 8px',
                    background: '#f1f5f9',
                    color: '#334155',
                    border: '1px solid #e2e8f0',
                  }}
                >
                  {svc.label}
                </span>
              ))}
            </div>
          )}

          {cap.services.length === 0 && (
            <span className="text-xs italic mt-1 block" style={{ color: '#94a3b8' }}>no services</span>
          )}
        </div>
      </div>
    </div>
  )
}

function DetailPanel({
  cap,
  allCaps,
  capById,
  onClose,
  insight,
}: {
  cap: CapabilityType
  allCaps: CapabilityType[]
  capById: Map<string, CapabilityType>
  onClose: () => void
  insight?: InsightItem
}) {
  void capById
  const visBand = VIS_BANDS.find(b => b.key === cap.visibility)
  const dependedOnBy = allCaps.filter(c => c.depends_on.some(d => d.id === cap.id))

  return (
    <>
      <div
        onClick={onClose}
        style={{
          position: 'fixed',
          inset: 0,
          background: 'rgba(0,0,0,0.1)',
          backdropFilter: 'blur(4px)',
          WebkitBackdropFilter: 'blur(4px)',
          zIndex: 40,
        }}
      />
      <div
        style={{
          position: 'fixed',
          top: 0,
          right: 0,
          height: '100vh',
          width: 360,
          background: '#ffffff',
          zIndex: 50,
          overflowY: 'auto',
          boxShadow: '-8px 0 30px rgba(0,0,0,0.08)',
        }}
        className="px-5 pt-0 pb-5"
      >
        <div
          className="sticky top-0 z-10 -mx-5 px-5 pt-5 pb-4 mb-2"
          style={{
            background: 'linear-gradient(180deg, #ffffff 0%, #ffffff 85%, rgba(255,255,255,0.92) 100%)',
            borderBottom: '1px solid #e2e8f0',
            boxShadow: '0 1px 0 rgba(248,250,252,0.8)',
          }}
        >
          <div
            style={{
              height: 3,
              margin: '-20px -20px 16px -20px',
              borderRadius: '0',
              background: 'linear-gradient(90deg, #6366f1 0%, #8b5cf6 50%, #ec4899 100%)',
            }}
          />
          <div className="flex items-center justify-between gap-2">
            {(() => {
              const vb = visBand ?? { label: cap.visibility, accent: '#475569', border: '#e2e8f0', bg: '#f1f5f9' }
              return (
                <span
                  className="font-semibold"
                  style={{
                    fontSize: 11,
                    fontWeight: 600,
                    textTransform: 'uppercase',
                    letterSpacing: '0.05em',
                    padding: '4px 10px',
                    borderRadius: 8,
                    background: vb.bg,
                    color: vb.accent,
                    border: `1px solid ${vb.border}`,
                  }}
                >
                  {vb.label}
                </span>
              )
            })()}
            <button
              type="button"
              onClick={onClose}
              className="leading-none ml-auto rounded-lg px-2 py-1 transition-colors hover:bg-slate-100"
              style={{ fontSize: 22, color: '#94a3b8' }}
              aria-label="Close"
            >
              ×
            </button>
          </div>

          <h3 className="font-bold text-base leading-snug mt-3" style={{ color: '#0f172a' }}>{cap.label}</h3>
        </div>

        <div className="space-y-4">
          {/* 1. Description */}
          <Section label="Description">
            {cap.description ? (
              <p className="text-sm leading-relaxed" style={{ color: '#475569' }}>{cap.description}</p>
            ) : (
              <p className="text-xs italic" style={{ color: '#94a3b8' }}>None</p>
            )}
          </Section>

          {/* 2. Visibility */}
          <Section label="Visibility">
            {(() => {
              const vb = visBand ?? { label: cap.visibility, accent: '#475569', border: '#e2e8f0', bg: '#f1f5f9' }
              return (
                <span
                  className="font-semibold inline-flex"
                  style={{
                    fontSize: 11,
                    fontWeight: 600,
                    padding: '4px 10px',
                    borderRadius: 8,
                    background: vb.bg,
                    color: vb.accent,
                    border: `1px solid ${vb.border}`,
                  }}
                >
                  {vb.label}
                </span>
              )
            })()}
          </Section>

          {insight && (
            <div
              className="rounded-2xl p-4 space-y-2"
              style={{ ...CARD_SHELL }}
            >
              <p style={{ ...SECTION_LABEL, color: '#4f46e5' }}>AI Insight</p>
              <p className="text-sm leading-relaxed" style={{ color: '#334155' }}>{insight.explanation}</p>
              {insight.suggestion && (
                <p className="text-sm leading-relaxed" style={{ color: '#4338ca' }}>{insight.suggestion}</p>
              )}
            </div>
          )}

          {/* 3. Teams */}
          <Section label="Teams">
            {cap.teams.length > 0 ? (
              <div className="flex flex-wrap gap-1.5">
                {cap.teams.map(team => (
                  <span
                    key={team.id}
                    className="text-xs rounded-lg px-2.5 py-1 font-medium"
                    style={{ background: '#f1f5f9', color: '#334155', border: '1px solid #e2e8f0' }}
                  >
                    {team.label} <span style={{ color: '#94a3b8' }}>({team.type})</span>
                  </span>
                ))}
              </div>
            ) : (
              <p className="text-xs italic" style={{ color: '#ef4444' }}>No team assigned</p>
            )}
          </Section>

          {/* 4. Services */}
          <Section label="Services">
            {cap.services.length > 0 ? (
              <div className="flex flex-wrap gap-1.5">
                {cap.services.map(svc => (
                  <span
                    key={svc.id}
                    className="font-mono text-xs rounded-lg px-2.5 py-1"
                    style={{ background: '#f1f5f9', color: '#334155', border: '1px solid #e2e8f0' }}
                  >
                    {svc.label}
                  </span>
                ))}
              </div>
            ) : (
              <p className="text-xs italic" style={{ color: '#94a3b8' }}>None</p>
            )}
          </Section>

          {/* 5. Depends On */}
          <Section label="Depends on">
            {cap.depends_on.length > 0 ? (
              <div className="flex flex-wrap gap-1.5">
                {cap.depends_on.map(dep => (
                  <span
                    key={dep.id}
                    className="text-xs rounded-lg px-2 py-1"
                    style={{ background: '#f1f5f9', color: '#334155', border: '1px solid #e2e8f0' }}
                  >
                    {dep.label}
                  </span>
                ))}
              </div>
            ) : (
              <p className="text-xs italic" style={{ color: '#94a3b8' }}>None</p>
            )}
          </Section>

          {/* 6. Depended On By */}
          <Section label="Depended on by">
            {dependedOnBy.length > 0 ? (
              <ul className="text-xs space-y-1" style={{ color: '#334155' }}>
                {dependedOnBy.map(c => <li key={c.id}>• {c.label}</li>)}
              </ul>
            ) : (
              <p className="text-xs italic" style={{ color: '#94a3b8' }}>None</p>
            )}
          </Section>

          {/* 7. External Dependencies */}
          <Section label="External Dependencies">
            {cap.external_deps && cap.external_deps.length > 0 ? (
              <div className="space-y-1.5">
                {cap.external_deps.map(dep => (
                  <div
                    key={dep.name}
                    className="flex items-center gap-2 px-3 py-2 rounded-xl"
                    style={{ background: '#f8fafc', border: '1px solid #e2e8f0' }}
                  >
                    <span className="text-xs font-medium" style={{ color: '#334155' }}>{dep.name}</span>
                    {dep.description && <span className="text-xs" style={{ color: '#94a3b8' }}>{dep.description}</span>}
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-xs italic" style={{ color: '#94a3b8' }}>None</p>
            )}
          </Section>

          {/* 8. Anti-patterns / Signals */}
          <Section label="Anti-patterns">
            {(cap.is_fragmented || cap.depended_on_by_count >= 3 || (cap.anti_patterns && cap.anti_patterns.length > 0)) ? (
              <div className="space-y-2">
                {cap.is_fragmented && (
                  <div className="text-xs rounded-xl px-3 py-2" style={{ background: '#fef2f2', color: '#b91c1c', border: '1px solid #fecaca', cursor: 'help' }}
                    title="This capability is owned by multiple teams, causing coordination overhead"
                    aria-label="Fragmented capability warning">
                    Fragmented — owned by multiple teams
                  </div>
                )}
                {cap.depended_on_by_count >= 3 && (
                  <div className="text-xs rounded-xl px-3 py-2" style={{ background: '#fffbeb', color: '#92400e', border: '1px solid #fde68a', cursor: 'help' }}
                    title={`${cap.depended_on_by_count} other capabilities depend on this one, making it a potential bottleneck`}
                    aria-label="High fan-in warning">
                    High fan-in — {cap.depended_on_by_count} capabilities depend on this
                  </div>
                )}
                {cap.anti_patterns?.map((ap, i) => (
                  <div
                    key={i}
                    className="text-xs rounded-xl px-3 py-2"
                    style={{ background: '#fff7ed', color: '#c2410c', border: '1px solid #fed7aa', cursor: 'help' }}
                    title={ap.message}
                    aria-label={`Anti-pattern: ${ap.code}`}
                  >
                    {ap.message}
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-xs italic" style={{ color: '#94a3b8' }}>None</p>
            )}
          </Section>

          {!cap.is_leaf && cap.children.length > 0 && (
            <Section label="Sub-capabilities">
              <ul className="text-xs space-y-1" style={{ color: '#334155' }}>
                {cap.children.map(ch => <li key={ch.id}>• {ch.label}</li>)}
              </ul>
            </Section>
          )}
        </div>
      </div>
    </>
  )
}

function Section({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div>
      <p className="mb-1.5" style={SECTION_LABEL}>{label}</p>
      {children}
    </div>
  )
}

function LoadingBlock() {
  return (
    <div className="flex flex-col items-center justify-center gap-3 h-full min-h-[200px]">
      <div
        className="rounded-full animate-spin"
        style={{
          width: 36,
          height: 36,
          border: '2px solid #e2e8f0',
          borderTopColor: '#6366f1',
        }}
      />
      <span style={{ fontSize: 14, color: '#94a3b8' }}>Loading…</span>
    </div>
  )
}

export function CapabilityView() {
  const { modelId, isHydrating } = useRequireModel()
  const { query } = useSearch()
  const [viewData, setViewData] = useState<CapabilityViewResponse | null>(null)
  const [loading, setLoading]   = useState(true)
  const [error, setError]       = useState<string | null>(null)
  const [viewMode, setViewMode] = useState<'visibility' | 'domain' | 'team'>('visibility')
  const [expandedGroups, setExpandedGroups] = useState<Set<string>>(new Set())
  const [selectedCap, setSelectedCap] = useState<CapabilityType | null>(null)
  const { insights } = usePageInsights('capabilities')

  useEffect(() => {
    if (isHydrating || !modelId) return
    api.getView(modelId, 'capability')
      .then(data => {
        const d = data as unknown as CapabilityViewResponse
        setViewData(d)
        setExpandedGroups(new Set(d.parent_groups.map(g => g.id)))
      })
      .catch(e => setError((e as Error).message))
      .finally(() => setLoading(false))
  }, [isHydrating, modelId])

  const capById = useMemo(
    () => new Map(viewData?.capabilities.map(c => [c.id, c]) ?? []),
    [viewData]
  )

  if (loading) return <LoadingBlock />
  if (error)   return <div className="flex items-center justify-center h-full" style={{ color: '#ef4444' }}>{error}</div>
  if (!viewData) return null

  const matchesCap = (cap: CapabilityType) =>
    !query || matchesQuery(cap.label, query) || cap.services.some(s => matchesQuery(s.label, query))

  const toggleGroup = (id: string) => {
    setExpandedGroups(prev => {
      const next = new Set(prev)
      if (next.has(id)) { next.delete(id) } else { next.add(id) }
      return next
    })
  }

  const capToParent = new Map<string, string>()
  viewData.parent_groups.forEach(pg => {
    pg.children.forEach(childId => capToParent.set(childId, pg.id))
  })

  const fragmentedCount = viewData.fragmented_capabilities.length
  const highSpanCount = viewData.high_span_services.length
  const disconnectedCount = viewData.capabilities.filter(c => c.is_leaf && c.teams.length === 0).length

  const atRiskUserFacing = viewData.capabilities.filter(
    cap => cap.visibility === 'user-facing' && cap.teams.length > 1
  )

  const dashInsight = insights['summary'] ?? insights['dashboard']

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4">
        <div>
          <h1 style={H1_GRADIENT}>Capability View</h1>
          <p className="mt-1" style={SUBTITLE}>
            {viewData.parent_groups.length} domain groups · {viewData.leaf_capability_count} capabilities
          </p>
        </div>
        <PillToggle
          value={viewMode}
          onChange={setViewMode}
          options={[
            { key: 'visibility', label: 'By Visibility' },
            { key: 'domain', label: 'By Domain' },
            { key: 'team', label: 'By Team' },
          ]}
        />
      </div>

      {fragmentedCount === 0 && disconnectedCount === 0 && highSpanCount === 0 && atRiskUserFacing.length === 0 ? (
        <div style={{
          background: '#f0fdf4', border: '1px solid #bbf7d0',
          borderRadius: 8, padding: '12px 16px', display: 'flex',
          alignItems: 'center', gap: 8, marginBottom: 16
        }}>
          <span style={{color:'#16a34a', fontSize:16}}>✓</span>
          <span style={{fontSize:13, color:'#15803d', fontWeight:500}}>
            No architecture issues detected — capabilities are well-structured
          </span>
        </div>
      ) : (
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
        <StatCard
          value={fragmentedCount}
          label="Fragmented"
          gradient={fragmentedCount > 0
            ? 'linear-gradient(135deg, #ffe4e6 0%, #fecdd3 45%, #fda4af 100%)'
            : 'linear-gradient(135deg, #ecfdf5 0%, #d1fae5 50%, #a7f3d0 100%)'}
          iconBg={fragmentedCount > 0 ? 'rgba(225,29,72,0.2)' : 'rgba(16,185,129,0.2)'}
          icon={
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke={fragmentedCount > 0 ? '#be123c' : '#047857'} strokeWidth="2">
              <path d="M12 9v4M12 17h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          }
        />
        <StatCard
          value={disconnectedCount}
          label="Unowned"
          gradient={disconnectedCount > 0
            ? 'linear-gradient(135deg, #fff7ed 0%, #ffedd5 50%, #fed7aa 100%)'
            : 'linear-gradient(135deg, #ecfdf5 0%, #d1fae5 50%, #a7f3d0 100%)'}
          iconBg={disconnectedCount > 0 ? 'rgba(234,88,12,0.2)' : 'rgba(16,185,129,0.2)'}
          icon={
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke={disconnectedCount > 0 ? '#c2410c' : '#047857'} strokeWidth="2">
              <path d="M16 7a4 4 0 10-8 0M4 21v-2a4 4 0 014-4h8a4 4 0 014 4v2" />
            </svg>
          }
        />
        <StatCard
          value={highSpanCount}
          label="High-span"
          gradient={highSpanCount > 0
            ? 'linear-gradient(135deg, #fef9c3 0%, #fef08a 50%, #fde047 100%)'
            : 'linear-gradient(135deg, #ecfdf5 0%, #d1fae5 50%, #a7f3d0 100%)'}
          iconBg={highSpanCount > 0 ? 'rgba(161,98,7,0.2)' : 'rgba(16,185,129,0.2)'}
          icon={
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke={highSpanCount > 0 ? '#a16207' : '#047857'} strokeWidth="2">
              <path d="M4 6h16M4 12h10M4 18h14" />
            </svg>
          }
        />
        <StatCard
          value={atRiskUserFacing.length}
          label="User-facing at risk"
          gradient={atRiskUserFacing.length > 0
            ? 'linear-gradient(135deg, #fce7f3 0%, #fbcfe8 50%, #f9a8d4 100%)'
            : 'linear-gradient(135deg, #ecfdf5 0%, #d1fae5 50%, #a7f3d0 100%)'}
          iconBg={atRiskUserFacing.length > 0 ? 'rgba(190,24,93,0.2)' : 'rgba(16,185,129,0.2)'}
          icon={
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke={atRiskUserFacing.length > 0 ? '#be185d' : '#047857'} strokeWidth="2">
              <path d="M12 3l8 4v6c0 5-3.5 9-8 10-4.5-1-8-5-8-10V7l8-4z" />
            </svg>
          }
        />
      </div>
      )}

      {dashInsight && (
        <div className="rounded-[20px] p-5" style={{ ...CARD_SHELL, background: 'linear-gradient(135deg, #eef2ff 0%, #e0e7ff 40%, #f8fafc 100%)' }}>
          <p className="mb-2" style={{ ...SECTION_LABEL, color: '#4338ca' }}>AI Capability Analysis</p>
          <p className="text-sm leading-relaxed" style={{ color: '#334155' }}>{dashInsight.explanation}</p>
          {dashInsight.suggestion && (
            <p className="text-sm leading-relaxed mt-2" style={{ color: '#3730a3' }}>{dashInsight.suggestion}</p>
          )}
        </div>
      )}

      {atRiskUserFacing.length > 0 && (
        <div
          className="rounded-[20px] p-5 overflow-hidden relative"
          style={{
            ...CARD_SHELL,
            border: '1px solid #fecaca',
            background: 'linear-gradient(135deg, #fff1f2 0%, #ffe4e6 35%, #ffffff 100%)',
          }}
        >
          <div
            className="absolute left-0 top-0 bottom-0 w-1"
            style={{ background: 'linear-gradient(180deg, #ef4444 0%, #f97316 100%)' }}
          />
          <div className="pl-2">
            <h3 className="text-sm font-bold mb-1" style={{ color: '#9f1239' }}>
              User-Facing Capabilities Served by Multiple Teams
            </h3>
            <div className="space-y-3 mt-3">
              {atRiskUserFacing.map(cap => (
                <div key={cap.id} className="flex items-center gap-2 flex-wrap">
                  <span className="text-sm font-semibold" style={{ color: '#0f172a' }}>{cap.label}</span>
                  {cap.teams.map(team => (
                    <span
                      key={team.id}
                      style={{
                        fontSize: 11,
                        fontWeight: 600,
                        borderRadius: 20,
                        padding: '4px 12px',
                        background: '#6366f1',
                        color: '#ffffff',
                      }}
                    >
                      {team.label}
                    </span>
                  ))}
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      {(fragmentedCount > 0 || highSpanCount > 0) && (
        <div className="flex flex-wrap gap-2">
          {viewData.fragmented_capabilities.map(fc => (
            <div
              key={fc.id}
              className="flex items-center gap-1.5 font-semibold"
              style={{
                fontSize: 11,
                fontWeight: 600,
                borderRadius: 20,
                padding: '6px 14px',
                background: '#fee2e2',
                color: '#b91c1c',
                border: '1px solid #fecaca',
              }}
            >
              <span>{fc.label}</span>
              <span style={{ color: '#ef4444' }}>· {fc.team_count} teams</span>
            </div>
          ))}
          {viewData.high_span_services.map(hs => (
            <div
              key={hs.name}
              className="flex items-center gap-1.5 font-semibold font-mono"
              style={{
                fontSize: 11,
                fontWeight: 600,
                borderRadius: 20,
                padding: '6px 14px',
                background: '#fef3c7',
                color: '#92400e',
                border: '1px solid #fde68a',
              }}
            >
              <span>{hs.name}</span>
              <span style={{ color: '#b45309' }}>· {hs.capability_count} capabilities</span>
            </div>
          ))}
        </div>
      )}

      {viewMode === 'visibility' && (
        <div className="space-y-10">
          {VIS_BANDS.map(band => {
            const bandCaps = viewData.capabilities.filter(cap =>
              cap.visibility === band.key && cap.is_leaf && matchesCap(cap)
            )
            if (bandCaps.length === 0) return null

            return (
              <div key={band.key}>
                <div className="flex items-center gap-3 mb-4">
                  <div className="h-px flex-1" style={{ background: `linear-gradient(90deg, transparent, ${band.border}, transparent)` }} />
                  <span
                    className="font-bold whitespace-nowrap"
                    style={{
                      fontSize: 11,
                      fontWeight: 600,
                      textTransform: 'uppercase',
                      letterSpacing: '0.05em',
                      padding: '6px 14px',
                      borderRadius: 20,
                      background: `linear-gradient(135deg, ${band.bg} 0%, #ffffff 100%)`,
                      color: band.accent,
                      border: `1px solid ${band.border}`,
                      boxShadow: '0 1px 3px rgba(0,0,0,0.04)',
                    }}
                  >
                    {band.label} · {bandCaps.length}
                  </span>
                  <div className="h-px flex-1" style={{ background: `linear-gradient(90deg, transparent, ${band.border}, transparent)` }} />
                </div>
                <div className="grid gap-3" style={{ gridTemplateColumns: 'repeat(auto-fill, minmax(260px, 1fr))' }}>
                  {bandCaps.map(cap => (
                    <CapabilityCard key={cap.id} cap={cap} band={band} capById={capById} onClick={() => setSelectedCap(cap)} />
                  ))}
                </div>
              </div>
            )
          })}
        </div>
      )}

      {viewMode === 'domain' && (
        <div className="space-y-3">
          {viewData.parent_groups.map(pg => {
            const groupCaps = pg.children
              .map(childId => capById.get(childId))
              .filter((c): c is CapabilityType => c != null && c.is_leaf && matchesCap(c))
            if (groupCaps.length === 0) return null
            const isExpanded = expandedGroups.has(pg.id)
            const fragmentedInGroup = groupCaps.filter(c => c.is_fragmented).length

            return (
              <div
                key={pg.id}
                className="overflow-hidden transition-shadow"
                style={{ ...CARD_SHELL }}
              >
                <button
                  type="button"
                  onClick={() => toggleGroup(pg.id)}
                  className="flex items-center gap-2 w-full text-left px-4 py-3.5 transition-colors hover:bg-slate-50/80"
                  style={{
                    background: 'linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%)',
                    borderBottom: isExpanded ? '1px solid #e2e8f0' : 'none',
                  }}
                >
                  <div
                    className="w-1 self-stretch rounded-full flex-shrink-0"
                    style={{ background: 'linear-gradient(180deg, #6366f1 0%, #8b5cf6 100%)', minHeight: 24 }}
                  />
                  <span className="text-sm font-bold" style={{ color: '#0f172a' }}>{pg.label}</span>
                  <span className="text-xs font-medium" style={{ color: '#94a3b8' }}>{groupCaps.length}</span>
                  {fragmentedInGroup > 0 && (
                    <span
                      className="font-semibold"
                      style={{
                        fontSize: 11,
                        fontWeight: 600,
                        borderRadius: 20,
                        padding: '2px 10px',
                        background: '#fee2e2',
                        color: '#b91c1c',
                      }}
                    >
                      {fragmentedInGroup} fragmented
                    </span>
                  )}
                  <span className="ml-auto text-xs font-medium" style={{ color: '#94a3b8' }}>{isExpanded ? '▾' : '▸'}</span>
                </button>
                {isExpanded && (
                  <div className="grid gap-3 p-4" style={{ gridTemplateColumns: 'repeat(auto-fill, minmax(260px, 1fr))' }}>
                    {groupCaps.map(cap => {
                      const band = VIS_BANDS.find(b => b.key === cap.visibility)
                      return <CapabilityCard key={cap.id} cap={cap} band={band} capById={capById} onClick={() => setSelectedCap(cap)} />
                    })}
                  </div>
                )}
              </div>
            )
          })}

          {(() => {
            const uncategorized = viewData.capabilities.filter(
              c => c.is_leaf && !capToParent.has(c.id) && matchesCap(c)
            )
            if (uncategorized.length === 0) return null
            const isExpanded = expandedGroups.has('__uncategorized__')
            return (
              <div
                className="overflow-hidden"
                style={{ ...CARD_SHELL }}
              >
                <button
                  type="button"
                  onClick={() => toggleGroup('__uncategorized__')}
                  className="flex items-center gap-2 w-full text-left px-4 py-3.5 transition-colors hover:bg-slate-50/80"
                  style={{
                    background: 'linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%)',
                    borderBottom: isExpanded ? '1px solid #e2e8f0' : 'none',
                  }}
                >
                  <div
                    className="w-1 self-stretch rounded-full flex-shrink-0"
                    style={{ background: 'linear-gradient(180deg, #94a3b8 0%, #cbd5e1 100%)', minHeight: 24 }}
                  />
                  <span className="text-sm font-bold" style={{ color: '#64748b' }}>Uncategorized</span>
                  <span className="text-xs font-medium" style={{ color: '#94a3b8' }}>{uncategorized.length}</span>
                  <span className="ml-auto text-xs font-medium" style={{ color: '#94a3b8' }}>{isExpanded ? '▾' : '▸'}</span>
                </button>
                {isExpanded && (
                  <div className="grid gap-3 p-4" style={{ gridTemplateColumns: 'repeat(auto-fill, minmax(260px, 1fr))' }}>
                    {uncategorized.map(cap => {
                      const band = VIS_BANDS.find(b => b.key === cap.visibility)
                      return <CapabilityCard key={cap.id} cap={cap} band={band} capById={capById} onClick={() => setSelectedCap(cap)} />
                    })}
                  </div>
                )}
              </div>
            )
          })()}
        </div>
      )}

      {viewMode === 'team' && (() => {
        // Build a map: team label → { type, caps[] }
        const teamMap = new Map<string, { type: string; caps: CapabilityType[] }>()
        const unowned: CapabilityType[] = []

        for (const cap of viewData.capabilities) {
          if (!cap.is_leaf || !matchesCap(cap)) continue
          if (cap.teams.length === 0) {
            unowned.push(cap)
          } else {
            for (const t of cap.teams) {
              const entry = teamMap.get(t.label) ?? { type: t.type, caps: [] }
              entry.caps.push(cap)
              teamMap.set(t.label, entry)
            }
          }
        }

        const TEAM_TYPE_BADGE: Record<string, { bg: string; text: string; accent: string }> = {
          'stream-aligned':        { bg: '#dbeafe', text: '#1e40af', accent: '#2563eb' },
          'platform':              { bg: '#ede9fe', text: '#5b21b6', accent: '#7c3aed' },
          'enabling':              { bg: '#d1fae5', text: '#065f46', accent: '#059669' },
          'complicated-subsystem': { bg: '#fef3c7', text: '#92400e', accent: '#d97706' },
        }

        const sortedTeams = Array.from(teamMap.entries()).sort((a, b) => b[1].caps.length - a[1].caps.length)

        return (
          <div className="space-y-3">
            {sortedTeams.map(([teamName, { type, caps }]) => {
              const badge = TEAM_TYPE_BADGE[type] ?? { bg: '#f1f5f9', text: '#475569', accent: '#6b7280' }
              const isExpanded = expandedGroups.has(`team:${teamName}`)
              const fragmented = caps.filter(c => c.is_fragmented).length

              return (
                <div key={teamName} className="overflow-hidden transition-shadow" style={{ ...CARD_SHELL }}>
                  <button
                    type="button"
                    onClick={() => toggleGroup(`team:${teamName}`)}
                    className="flex items-center gap-2 w-full text-left px-4 py-3.5 transition-colors hover:bg-slate-50/80"
                    style={{
                      background: 'linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%)',
                      borderBottom: isExpanded ? '1px solid #e2e8f0' : 'none',
                    }}
                  >
                    <div className="w-1 self-stretch rounded-full flex-shrink-0" style={{ background: `linear-gradient(180deg, ${badge.accent} 0%, ${badge.accent}88 100%)`, minHeight: 24 }} />
                    <span className="text-sm font-bold" style={{ color: '#0f172a' }}>{teamName}</span>
                    <span className="text-xs font-semibold px-2 py-0.5 rounded-full" style={{ background: badge.bg, color: badge.text }}>{type}</span>
                    <span className="text-xs font-medium" style={{ color: '#94a3b8' }}>{caps.length} cap{caps.length !== 1 ? 's' : ''}</span>
                    {fragmented > 0 && (
                      <span className="font-semibold" style={{ fontSize: 11, fontWeight: 600, borderRadius: 20, padding: '2px 10px', background: '#fee2e2', color: '#b91c1c' }}>
                        {fragmented} fragmented
                      </span>
                    )}
                    <span className="ml-auto text-xs font-medium" style={{ color: '#94a3b8' }}>{isExpanded ? '▾' : '▸'}</span>
                  </button>
                  {isExpanded && (
                    <div className="grid gap-3 p-4" style={{ gridTemplateColumns: 'repeat(auto-fill, minmax(260px, 1fr))' }}>
                      {caps.map(cap => {
                        const band = VIS_BANDS.find(b => b.key === cap.visibility)
                        return <CapabilityCard key={cap.id} cap={cap} band={band} capById={capById} onClick={() => setSelectedCap(cap)} />
                      })}
                    </div>
                  )}
                </div>
              )
            })}

            {unowned.length > 0 && (() => {
              const isExpanded = expandedGroups.has('team:__unowned__')
              return (
                <div className="overflow-hidden" style={{ ...CARD_SHELL }}>
                  <button
                    type="button"
                    onClick={() => toggleGroup('team:__unowned__')}
                    className="flex items-center gap-2 w-full text-left px-4 py-3.5 transition-colors hover:bg-slate-50/80"
                    style={{ background: 'linear-gradient(135deg, #fff1f2 0%, #ffe4e6 100%)', borderBottom: isExpanded ? '1px solid #fecaca' : 'none' }}
                  >
                    <div className="w-1 self-stretch rounded-full flex-shrink-0" style={{ background: 'linear-gradient(180deg, #ef4444, #f87171)', minHeight: 24 }} />
                    <span className="text-sm font-bold" style={{ color: '#9f1239' }}>Unowned</span>
                    <span className="text-xs font-medium" style={{ color: '#94a3b8' }}>{unowned.length} cap{unowned.length !== 1 ? 's' : ''}</span>
                    <span className="ml-auto text-xs font-medium" style={{ color: '#94a3b8' }}>{isExpanded ? '▾' : '▸'}</span>
                  </button>
                  {isExpanded && (
                    <div className="grid gap-3 p-4" style={{ gridTemplateColumns: 'repeat(auto-fill, minmax(260px, 1fr))' }}>
                      {unowned.map(cap => {
                        const band = VIS_BANDS.find(b => b.key === cap.visibility)
                        return <CapabilityCard key={cap.id} cap={cap} band={band} capById={capById} onClick={() => setSelectedCap(cap)} />
                      })}
                    </div>
                  )}
                </div>
              )
            })()}
          </div>
        )
      })()}

      {selectedCap && (
        <DetailPanel
          cap={selectedCap}
          allCaps={viewData.capabilities}
          capById={capById}
          onClose={() => setSelectedCap(null)}
          insight={insights[`cap:${slug(selectedCap.label)}`]}
        />
      )}
    </div>
  )
}
