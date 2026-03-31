import { QuickAction } from '@/components/changeset/QuickAction'
import type { CapabilityViewResponse } from '@/types/views'
import { VIS_BANDS } from './constants'

export type CapabilityType = CapabilityViewResponse['capabilities'][number]

export function CapabilityCard({ cap, onClick }: { cap: CapabilityType; onClick: () => void }) {
  const band = VIS_BANDS.find(b => b.key === cap.visibility)
  const isBottleneck = cap.depended_on_by_count >= 5
  const accent = cap.is_fragmented ? '#ef4444' : (band?.accent ?? '#94a3b8')
  const accentSoft = cap.is_fragmented ? '#fca5a5' : (band?.border ?? '#cbd5e1')

  return (
    <div onClick={onClick} title={cap.description || cap.label} className="cursor-pointer rounded-2xl border border-slate-200 bg-gradient-to-br from-white to-slate-50 overflow-hidden transition-all"
      style={{ boxShadow: '0 1px 3px rgba(0,0,0,0.04)' }}
      onMouseEnter={e => {
        (e.currentTarget as HTMLDivElement).style.transform = 'translateY(-1px)'
        ;(e.currentTarget as HTMLDivElement).style.boxShadow = '0 8px 24px rgba(15,23,42,0.08)'
        const btn = e.currentTarget.querySelector('.qa-cap') as HTMLElement; if (btn) btn.style.opacity = '1'
      }}
      onMouseLeave={e => {
        (e.currentTarget as HTMLDivElement).style.transform = ''
        ;(e.currentTarget as HTMLDivElement).style.boxShadow = '0 1px 3px rgba(0,0,0,0.04)'
        const btn = e.currentTarget.querySelector('.qa-cap') as HTMLElement; if (btn) btn.style.opacity = '0.35'
      }}>
      <div className="flex">
        <div style={{ width: 3, flexShrink: 0, background: cap.is_fragmented ? 'linear-gradient(180deg, #ef4444 0%, #f87171 100%)' : `linear-gradient(180deg, ${accent} 0%, ${accentSoft} 100%)` }} />
        <div className="flex-1 px-3 py-2.5">
          <div className="flex items-start justify-between gap-2">
            <span className="text-sm font-semibold leading-snug text-slate-900">{cap.label}</span>
            <div className="flex items-center gap-1 shrink-0">
              <span className="qa-cap" style={{ opacity: 0.35, transition: 'opacity 0.15s' }}>
                <QuickAction size={11} options={[
                  { label: 'Update visibility', action: { type: 'update_capability_visibility', capability_name: cap.label } },
                  { label: 'Reassign to another team', action: { type: 'reassign_capability', capability_name: cap.label } },
                ]} />
              </span>
              {cap.is_fragmented && (
                <span className="text-[11px] font-semibold px-2 py-0.5 rounded-md bg-red-100 text-red-700 cursor-help" title="Fragmented: owned by multiple teams">split</span>
              )}
              {isBottleneck && !cap.is_fragmented && (
                <span className="text-[11px] font-semibold px-2 py-0.5 rounded-md bg-amber-50 text-amber-700 cursor-help" title={`${cap.depended_on_by_count} capabilities depend on this`}>↙{cap.depended_on_by_count}</span>
              )}
            </div>
          </div>
          {cap.services.length > 0 ? (
            <div className="flex flex-wrap gap-1 mt-2">
              {cap.services.map(svc => (
                <span key={svc.id} className="font-mono text-[11px] rounded-md px-2 py-0.5 bg-slate-100 text-slate-600 border border-slate-200">{svc.label}</span>
              ))}
            </div>
          ) : (
            <span className="text-xs italic mt-1 block text-slate-400">no services</span>
          )}
        </div>
      </div>
    </div>
  )
}
