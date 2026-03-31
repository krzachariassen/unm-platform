export const VIS_BANDS = [
  { key: 'user-facing',    label: 'User-facing',    accent: '#2563eb', border: '#bfdbfe', bg: '#eff6ff' },
  { key: 'domain',         label: 'Domain',         accent: '#7c3aed', border: '#ddd6fe', bg: '#f5f3ff' },
  { key: 'foundational',   label: 'Foundational',   accent: '#059669', border: '#a7f3d0', bg: '#f0fdf4' },
  { key: 'infrastructure', label: 'Infrastructure', accent: '#6b7280', border: '#e5e7eb', bg: '#f9fafb' },
]

export const TEAM_TYPE_CAP_BADGE: Record<string, { bg: string; text: string; accent: string }> = {
  'stream-aligned':        { bg: '#dbeafe', text: '#1e40af', accent: '#2563eb' },
  'platform':              { bg: '#ede9fe', text: '#5b21b6', accent: '#7c3aed' },
  'enabling':              { bg: '#d1fae5', text: '#065f46', accent: '#059669' },
  'complicated-subsystem': { bg: '#fef3c7', text: '#92400e', accent: '#d97706' },
}
