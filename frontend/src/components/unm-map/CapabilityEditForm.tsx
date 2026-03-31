import type { EditState, SvcInfo } from '@/features/unm-map/types'
import { teamColor } from '@/features/unm-map/constants'

interface CapabilityEditFormProps {
  editState: EditState
  teams: string[]
  services: string[]
  isEditMode: boolean
  onUpdateState: (updater: (s: EditState) => EditState) => void
  onSave: () => void
  onMoveService: (svc: SvcInfo, toTeam: string) => void
  onUnlinkService: (svcLabel: string) => void
  onLinkService: (svcName: string) => void
  onAddService: (svcName: string) => void
}

export function CapabilityEditForm({
  editState, teams, services, isEditMode,
  onUpdateState, onSave, onMoveService, onUnlinkService, onLinkService, onAddService,
}: CapabilityEditFormProps) {
  if (editState.isPendingNode) return null

  return (
    <div style={{ borderBottom: isEditMode ? '1px solid #e5e7eb' : 'none', marginBottom: isEditMode ? 16 : 0, paddingBottom: isEditMode ? 4 : 0 }}>
      <div style={{ fontSize: 12, fontWeight: 600, color: '#374151', marginBottom: 12 }}>Edit this capability</div>

      <div style={{ marginBottom: 10 }}>
        <div style={{ fontSize: 11, color: '#6b7280', marginBottom: 4 }}>Description</div>
        <textarea
          value={editState.description}
          onChange={e => onUpdateState(s => ({ ...s, description: e.target.value }))}
          rows={3}
          style={{ width: '100%', fontSize: 12, padding: '6px 8px', borderRadius: 6, border: '1px solid #d1d5db', resize: 'vertical', minHeight: 56, boxSizing: 'border-box', fontFamily: 'inherit', lineHeight: 1.4 }}
        />
      </div>

      <div style={{ marginBottom: 10 }}>
        <div style={{ fontSize: 11, color: '#6b7280', marginBottom: 4 }}>Visibility</div>
        <select
          value={editState.visibility}
          onChange={e => onUpdateState(s => ({ ...s, visibility: e.target.value }))}
          style={{ width: '100%', fontSize: 12, padding: '5px 8px', borderRadius: 6, border: '1px solid #d1d5db', background: '#fff' }}
        >
          <option value="user-facing">User-facing</option>
          <option value="domain">Domain</option>
          <option value="foundational">Foundational</option>
          <option value="infrastructure">Infrastructure</option>
        </select>
      </div>

      <div style={{ marginBottom: 14 }}>
        <div style={{ fontSize: 11, color: '#6b7280', marginBottom: 4 }}>Owning Team</div>
        <select
          value={editState.teamName}
          onChange={e => onUpdateState(s => ({ ...s, teamName: e.target.value }))}
          style={{ width: '100%', fontSize: 12, padding: '5px 8px', borderRadius: 6, border: '1px solid #d1d5db', background: '#fff' }}
        >
          <option value="">— Unowned —</option>
          {teams.map(t => <option key={t} value={t}>{t}</option>)}
        </select>
      </div>

      <button
        onClick={onSave}
        style={{ width: '100%', padding: '7px', borderRadius: 6, background: '#111827', color: '#fff', border: 'none', fontSize: 12, fontWeight: 500, cursor: 'pointer', marginBottom: 16 }}
      >
        Stage changes →
      </button>

      <div>
        <div style={{ fontSize: 11, fontWeight: 600, color: '#374151', marginBottom: 8, textTransform: 'uppercase', letterSpacing: '0.05em' }}>Services</div>
        {editState.svcs.length > 0 ? editState.svcs.map(svc => (
          <div key={svc.id} style={{ display: 'flex', alignItems: 'center', gap: 5, marginBottom: 6, background: '#f9fafb', borderRadius: 5, padding: '4px 6px', border: '1px solid #e5e7eb' }}>
            <div style={{ width: 5, height: 5, borderRadius: '50%', background: teamColor(svc.teamName), flexShrink: 0 }} />
            <span style={{ fontSize: 11, color: '#374151', flex: 1, minWidth: 0, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{svc.label}</span>
            <select defaultValue="" onChange={e => { if (e.target.value) { onMoveService(svc, e.target.value); e.target.value = '' } }}
              style={{ fontSize: 10, padding: '2px 4px', borderRadius: 4, border: '1px solid #d1d5db', background: '#fff', maxWidth: 90 }}>
              <option value="">Move…</option>
              {teams.filter(t => t !== svc.teamName).map(t => <option key={t} value={t}>{t}</option>)}
            </select>
            <button onClick={() => onUnlinkService(svc.label)}
              style={{ fontSize: 10, padding: '2px 5px', borderRadius: 4, border: '1px solid #fca5a5', background: '#fef2f2', color: '#b91c1c', cursor: 'pointer', flexShrink: 0 }}>
              Unlink
            </button>
          </div>
        )) : (
          <div style={{ fontSize: 11, color: '#9ca3af', marginBottom: 8, fontStyle: 'italic' }}>No services linked yet</div>
        )}

        <div style={{ display: 'flex', gap: 5, marginTop: 8 }}>
          <select value={editState.linkSvcName}
            onChange={e => onUpdateState(s => ({ ...s, linkSvcName: e.target.value }))}
            style={{ flex: 1, fontSize: 11, padding: '4px 6px', borderRadius: 5, border: '1px solid #d1d5db', background: '#fff' }}>
            <option value="">Link existing service…</option>
            {services.filter(s => !editState.svcs.some(sv => sv.label === s)).map(s => <option key={s} value={s}>{s}</option>)}
          </select>
          <button onClick={() => { if (editState.linkSvcName) { onLinkService(editState.linkSvcName); onUpdateState(s => ({ ...s, linkSvcName: '' })) } }}
            disabled={!editState.linkSvcName}
            style={{ fontSize: 11, padding: '4px 10px', borderRadius: 5, border: 'none', background: '#1d4ed8', color: '#fff', cursor: editState.linkSvcName ? 'pointer' : 'default', opacity: editState.linkSvcName ? 1 : 0.35, flexShrink: 0 }}>
            Link
          </button>
        </div>

        <div style={{ display: 'flex', gap: 5, marginTop: 6 }}>
          <input type="text" placeholder="New service name…" value={editState.newSvcName}
            onChange={e => onUpdateState(s => ({ ...s, newSvcName: e.target.value }))}
            onKeyDown={e => { if (e.key === 'Enter' && editState.newSvcName.trim()) { onAddService(editState.newSvcName.trim()); onUpdateState(s => ({ ...s, newSvcName: '' })) } }}
            style={{ flex: 1, fontSize: 11, padding: '4px 6px', borderRadius: 5, border: '1px solid #d1d5db', outline: 'none' }} />
          <button onClick={() => { if (editState.newSvcName.trim()) { onAddService(editState.newSvcName.trim()); onUpdateState(s => ({ ...s, newSvcName: '' })) } }}
            disabled={!editState.newSvcName.trim()}
            style={{ fontSize: 11, padding: '4px 10px', borderRadius: 5, border: 'none', background: '#059669', color: '#fff', cursor: editState.newSvcName.trim() ? 'pointer' : 'default', opacity: editState.newSvcName.trim() ? 1 : 0.35, flexShrink: 0, whiteSpace: 'nowrap' }}>
            + Add
          </button>
        </div>
      </div>
    </div>
  )
}
