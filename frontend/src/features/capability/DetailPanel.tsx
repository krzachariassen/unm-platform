import type { CapabilityType } from './CapabilityCard'
import type { InsightItem } from '@/types/insights'
import { VIS_BANDS } from './constants'

export function DetailPanel({ cap, allCaps, onClose, insight }: {
  cap: CapabilityType
  allCaps: CapabilityType[]
  onClose: () => void
  insight?: InsightItem
}) {
  const visBand = VIS_BANDS.find(b => b.key === cap.visibility)
  const dependedOnBy = allCaps.filter(c => c.depends_on.some(d => d.id === cap.id))
  const vb = visBand ?? { label: cap.visibility, accent: '#475569', border: '#e2e8f0', bg: '#f1f5f9' }

  const Section = ({ label, children }: { label: string; children: React.ReactNode }) => (
    <div>
      <p className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider mb-1.5">{label}</p>
      {children}
    </div>
  )

  const pillCls = 'text-[11px] font-semibold px-2.5 py-1 rounded-lg'

  return (
    <>
      <div onClick={onClose} className="fixed inset-0 bg-black/10 backdrop-blur-sm z-40" />
      <div className="fixed top-0 right-0 h-full w-[360px] bg-white z-50 overflow-y-auto shadow-2xl">
        <div className="sticky top-0 z-10 bg-white border-b border-slate-200 shadow-sm">
          <div className="h-0.5" style={{ background: 'linear-gradient(90deg, #6366f1 0%, #8b5cf6 50%, #ec4899 100%)' }} />
          <div className="flex items-center justify-between px-5 py-4">
            <span className={pillCls} style={{ background: vb.bg, color: vb.accent, border: `1px solid ${vb.border}` }}>{vb.label}</span>
            <button type="button" onClick={onClose} className="text-2xl leading-none text-slate-400 rounded-lg px-2 py-1 hover:bg-slate-100 transition-colors" aria-label="Close">×</button>
          </div>
          <h3 className="px-5 pb-4 font-bold text-base text-slate-900 leading-snug">{cap.label}</h3>
        </div>

        <div className="px-5 py-4 space-y-4">
          <Section label="Description">
            {cap.description
              ? <p className="text-sm leading-relaxed text-slate-600">{cap.description}</p>
              : <p className="text-xs italic text-slate-400">None</p>}
          </Section>

          {insight && (
            <div className="rounded-2xl p-4 border border-slate-200 bg-gradient-to-br from-white to-slate-50">
              <p className="text-[11px] font-semibold text-indigo-600 uppercase tracking-wider mb-1.5">AI Insight</p>
              <p className="text-sm leading-relaxed text-slate-700">{insight.explanation}</p>
              {insight.suggestion && <p className="text-sm leading-relaxed text-indigo-700 mt-1 font-medium">{insight.suggestion}</p>}
            </div>
          )}

          <Section label="Teams">
            {cap.teams.length > 0
              ? <div className="flex flex-wrap gap-1.5">{cap.teams.map(t => <span key={t.id} className="text-xs rounded-lg px-2.5 py-1 bg-slate-100 text-slate-700 border border-slate-200">{t.label} <span className="text-slate-400">({t.type})</span></span>)}</div>
              : <p className="text-xs italic text-red-500">No team assigned</p>}
          </Section>

          <Section label="Services">
            {cap.services.length > 0
              ? <div className="flex flex-wrap gap-1.5">{cap.services.map(s => <span key={s.id} className="font-mono text-xs rounded-lg px-2.5 py-1 bg-slate-100 text-slate-700 border border-slate-200">{s.label}</span>)}</div>
              : <p className="text-xs italic text-slate-400">None</p>}
          </Section>

          <Section label="Depends on">
            {cap.depends_on.length > 0
              ? <div className="flex flex-wrap gap-1.5">{cap.depends_on.map(d => <span key={d.id} className="text-xs rounded-lg px-2 py-1 bg-slate-100 text-slate-700 border border-slate-200">{d.label}</span>)}</div>
              : <p className="text-xs italic text-slate-400">None</p>}
          </Section>

          <Section label="Depended on by">
            {dependedOnBy.length > 0
              ? <ul className="text-xs space-y-1 text-slate-700">{dependedOnBy.map(c => <li key={c.id}>• {c.label}</li>)}</ul>
              : <p className="text-xs italic text-slate-400">None</p>}
          </Section>

          {cap.external_deps && cap.external_deps.length > 0 && (
            <Section label="External Dependencies">
              <div className="space-y-1.5">
                {cap.external_deps.map(dep => (
                  <div key={dep.name} className="flex items-center gap-2 px-3 py-2 rounded-xl bg-slate-50 border border-slate-200">
                    <span className="text-xs font-medium text-slate-700">{dep.name}</span>
                    {dep.description && <span className="text-xs text-slate-400">{dep.description}</span>}
                  </div>
                ))}
              </div>
            </Section>
          )}

          {(cap.is_fragmented || cap.depended_on_by_count >= 3 || (cap.anti_patterns?.length ?? 0) > 0) && (
            <Section label="Anti-patterns">
              <div className="space-y-2">
                {cap.is_fragmented && <div className="text-xs rounded-xl px-3 py-2 bg-red-50 text-red-700 border border-red-200 cursor-help" title="Owned by multiple teams">Fragmented — owned by multiple teams</div>}
                {cap.depended_on_by_count >= 3 && <div className="text-xs rounded-xl px-3 py-2 bg-amber-50 text-amber-800 border border-amber-200">High fan-in — {cap.depended_on_by_count} capabilities depend on this</div>}
                {cap.anti_patterns?.map((ap, i) => <div key={i} className="text-xs rounded-xl px-3 py-2 bg-orange-50 text-orange-700 border border-orange-200" title={ap.message}>{ap.message}</div>)}
              </div>
            </Section>
          )}

          {!cap.is_leaf && cap.children.length > 0 && (
            <Section label="Sub-capabilities">
              <ul className="text-xs space-y-1 text-slate-700">{cap.children.map(ch => <li key={ch.id}>• {ch.label}</li>)}</ul>
            </Section>
          )}
        </div>
      </div>
    </>
  )
}
