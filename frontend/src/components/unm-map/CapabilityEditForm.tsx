import { AlertTriangle } from 'lucide-react'
import type { EditState, SvcInfo } from '@/features/unm-map/types'
import { teamColor } from '@/features/unm-map/constants'

interface CapabilityEditFormProps {
  editState: EditState
  teams: string[]
  services: string[]
  onUpdateState: (updater: (s: EditState) => EditState) => void
  onSave: () => void
  onMoveService: (svc: SvcInfo, toTeam: string) => void
  onUnlinkService: (svcLabel: string) => void
  onLinkService: (svcName: string) => void
  onAddService: (svcName: string) => void
}

const inputClass = 'w-full rounded-md border px-2.5 py-1.5 text-xs outline-none focus:ring-1 focus:ring-blue-200'
const labelClass = 'text-[10px] font-medium mb-0.5 block'

export function CapabilityEditForm({
  editState, teams, services,
  onUpdateState, onSave, onMoveService, onUnlinkService, onLinkService, onAddService,
}: CapabilityEditFormProps) {
  const hasChanges =
    editState.description !== editState.origDescription ||
    editState.visibility !== editState.origVisibility ||
    editState.teamName !== editState.origTeam

  const noServices = editState.svcs.length === 0

  return (
    <div className="space-y-4">
      {/* Pending badge */}
      {editState.isPendingNode && (
        <div className="rounded-md px-3 py-2" style={{ background: '#fef3c7', border: '1px solid #fcd34d' }}>
          <p className="text-[11px] font-medium" style={{ color: '#92400e' }}>Pending — not yet committed</p>
        </div>
      )}

      {/* Validation warning */}
      {noServices && (
        <div className="flex items-start gap-2 rounded-md px-3 py-2" style={{ background: '#fef2f2', border: '1px solid #fca5a5' }}>
          <AlertTriangle size={12} className="shrink-0 mt-0.5" style={{ color: '#dc2626' }} />
          <p className="text-[11px]" style={{ color: '#991b1b' }}>
            No service linked — this will fail validation. Link or create a service below.
          </p>
        </div>
      )}

      {/* Properties */}
      <div>
        <p className="text-[10px] font-semibold uppercase tracking-wider mb-2" style={{ color: '#9ca3af' }}>Properties</p>
        <div className="space-y-2.5">
          <div>
            <label className={labelClass} style={{ color: '#6b7280' }}>Description</label>
            <textarea
              value={editState.description}
              onChange={e => onUpdateState(s => ({ ...s, description: e.target.value }))}
              rows={2}
              className={inputClass}
              style={{ background: '#ffffff', borderColor: '#e5e7eb', color: '#374151', minHeight: 44, resize: 'vertical' }}
            />
          </div>
          <div>
            <label className={labelClass} style={{ color: '#6b7280' }}>Visibility</label>
            <select
              value={editState.visibility}
              onChange={e => onUpdateState(s => ({ ...s, visibility: e.target.value }))}
              className={inputClass}
              style={{ background: '#ffffff', borderColor: '#e5e7eb', color: '#374151' }}
            >
              <option value="user-facing">User-facing</option>
              <option value="domain">Domain</option>
              <option value="foundational">Foundational</option>
              <option value="infrastructure">Infrastructure</option>
            </select>
          </div>
          <div>
            <label className={labelClass} style={{ color: '#6b7280' }}>Owning Team</label>
            <select
              value={editState.teamName}
              onChange={e => onUpdateState(s => ({ ...s, teamName: e.target.value }))}
              className={inputClass}
              style={{ background: '#ffffff', borderColor: '#e5e7eb', color: '#374151' }}
            >
              <option value="">— Unowned —</option>
              {teams.map(t => <option key={t} value={t}>{t}</option>)}
            </select>
          </div>
        </div>
      </div>

      {/* Services */}
      <div>
        <p className="text-[10px] font-semibold uppercase tracking-wider mb-2" style={{ color: noServices ? '#dc2626' : '#9ca3af' }}>
          Services {editState.svcs.length > 0 ? `(${editState.svcs.length})` : '— required'}
        </p>

        {editState.svcs.length > 0 ? (
          <div className="space-y-1 mb-2">
            {editState.svcs.map(svc => (
              <div key={svc.id} className="flex items-center gap-1.5 rounded-md border px-2 py-1.5" style={{ borderColor: '#e5e7eb', background: '#f9fafb' }}>
                <div className="h-2 w-2 shrink-0 rounded-full" style={{ background: teamColor(svc.teamName) }} />
                <span className="min-w-0 flex-1 truncate text-[11px]" style={{ color: '#374151' }}>{svc.label}</span>
                <select
                  defaultValue=""
                  onChange={e => { if (e.target.value) { onMoveService(svc, e.target.value); e.target.value = '' } }}
                  className="max-w-[72px] shrink-0 rounded border px-1 py-0.5 text-[9px]"
                  style={{ borderColor: '#e5e7eb', background: '#ffffff', color: '#6b7280' }}
                >
                  <option value="">Move…</option>
                  {teams.filter(t => t !== svc.teamName).map(t => <option key={t} value={t}>{t}</option>)}
                </select>
                <button
                  type="button"
                  onClick={() => onUnlinkService(svc.label)}
                  className="shrink-0 rounded border px-1.5 py-0.5 text-[9px] transition-colors hover:bg-red-50"
                  style={{ borderColor: '#fca5a5', color: '#dc2626', background: '#fef2f2' }}
                >
                  Unlink
                </button>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-[11px] italic mb-2" style={{ color: '#dc2626' }}>No services linked</p>
        )}

        <div className="flex gap-1 mb-1">
          <select
            value={editState.linkSvcName}
            onChange={e => onUpdateState(s => ({ ...s, linkSvcName: e.target.value }))}
            className="min-w-0 flex-1 rounded-md border px-2 py-1 text-[11px]"
            style={{ borderColor: '#e5e7eb', background: '#ffffff', color: '#374151' }}
          >
            <option value="">Link existing…</option>
            {services.filter(s => !editState.svcs.some(sv => sv.label === s)).map(s => <option key={s} value={s}>{s}</option>)}
          </select>
          <button
            type="button"
            onClick={() => { if (editState.linkSvcName) { onLinkService(editState.linkSvcName); onUpdateState(s => ({ ...s, linkSvcName: '' })) } }}
            disabled={!editState.linkSvcName}
            className="shrink-0 rounded-md px-2.5 py-1 text-[11px] font-medium text-white disabled:opacity-30"
            style={{ background: '#2563eb' }}
          >
            Link
          </button>
        </div>

        <div className="flex gap-1">
          <input
            type="text"
            placeholder="New service…"
            value={editState.newSvcName}
            onChange={e => onUpdateState(s => ({ ...s, newSvcName: e.target.value }))}
            onKeyDown={e => { if (e.key === 'Enter' && editState.newSvcName.trim()) { onAddService(editState.newSvcName.trim()); onUpdateState(s => ({ ...s, newSvcName: '' })) } }}
            className="min-w-0 flex-1 rounded-md border px-2 py-1 text-[11px] outline-none"
            style={{ borderColor: '#e5e7eb', background: '#ffffff', color: '#374151' }}
          />
          <button
            type="button"
            onClick={() => { if (editState.newSvcName.trim()) { onAddService(editState.newSvcName.trim()); onUpdateState(s => ({ ...s, newSvcName: '' })) } }}
            disabled={!editState.newSvcName.trim()}
            className="shrink-0 rounded-md px-2.5 py-1 text-[11px] font-medium text-white disabled:opacity-30"
            style={{ background: '#059669' }}
          >
            + Add
          </button>
        </div>
      </div>

      {/* Stage button — only for property changes on non-pending nodes */}
      {hasChanges && !editState.isPendingNode && (
        <button
          type="button"
          onClick={onSave}
          className="w-full rounded-md py-2 text-xs font-medium text-white transition-colors"
          style={{ background: '#111827' }}
        >
          Stage changes
        </button>
      )}
    </div>
  )
}
