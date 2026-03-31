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
    <div
      role="button"
      tabIndex={0}
      onClick={onClick}
      onKeyDown={e => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); onClick() } }}
      title={cap.description || cap.label}
      className="w-full text-left cursor-pointer rounded-lg border border-border bg-card overflow-hidden transition-all hover:-translate-y-px hover:shadow-md group"
    >
      <div className="flex">
        <div
          className="w-[3px] shrink-0"
          style={{ background: cap.is_fragmented ? 'linear-gradient(180deg, #ef4444 0%, #f87171 100%)' : `linear-gradient(180deg, ${accent} 0%, ${accentSoft} 100%)` }}
        />
        <div className="flex-1 px-3 py-2.5">
          <div className="flex items-start justify-between gap-2">
            <span className="text-sm font-semibold leading-snug text-foreground">{cap.label}</span>
            <div className="flex items-center gap-1 shrink-0" onClick={e => e.stopPropagation()}>
              <span className="opacity-0 group-hover:opacity-100 transition-opacity">
                <QuickAction size={11} options={[
                  { label: 'Update visibility', action: { type: 'update_capability_visibility', capability_name: cap.label } },
                  { label: 'Reassign to another team', action: { type: 'reassign_capability', capability_name: cap.label } },
                ]} />
              </span>
              {cap.is_fragmented && (
                <span className="text-[10px] font-semibold px-1.5 py-0.5 rounded bg-red-100 text-red-700">split</span>
              )}
              {isBottleneck && !cap.is_fragmented && (
                <span className="text-[10px] font-semibold px-1.5 py-0.5 rounded bg-amber-50 text-amber-700" title={`${cap.depended_on_by_count} capabilities depend on this`}>↙{cap.depended_on_by_count}</span>
              )}
            </div>
          </div>
          {cap.services.length > 0 ? (
            <div className="flex flex-wrap gap-1 mt-1.5">
              {cap.services.map(svc => (
                <span key={svc.id} className="font-mono text-[10px] rounded px-1.5 py-0.5 bg-muted text-muted-foreground border border-border">{svc.label}</span>
              ))}
            </div>
          ) : (
            <span className="text-[10px] italic mt-1 block text-muted-foreground/60">no services</span>
          )}
        </div>
      </div>
    </div>
  )
}
