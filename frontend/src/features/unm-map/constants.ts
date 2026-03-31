// Layout constants
export const PAD_X = 60
export const ACTOR_Y = 52
export const ACTOR_W = 138
export const ACTOR_H = 46
export const NEED_Y = 188
export const NEED_W = 150
export const NEED_H = 52
export const NEED_GAP = 16
export const ACTOR_SECTION_GAP = 48
export const CAP_W = 172
export const CAP_H = 96
export const CAP_GAP = 14
export const CAP_BAND_PAD = 20
export const VIS_FIRST_Y = 340
export const MIN_BAND_H = 130
export const EXT_DEP_W = 150
export const EXT_DEP_H = 40
export const EXT_DEP_GAP = 14

export const VIS_ORDER = ['user-facing', 'domain', 'foundational', 'infrastructure'] as const
export type VisKey = (typeof VIS_ORDER)[number]

export const VIS: Record<string, {
  nodeBg: string; border: string; text: string; bandBg: string; label: string; line: string
}> = {
  'user-facing':    { nodeBg: '#fffbeb', border: '#d97706', text: '#92400e', bandBg: 'rgba(217,119,6,0.04)',   label: 'User-facing',    line: '#fbbf24' },
  'domain':         { nodeBg: '#f5f3ff', border: '#7c3aed', text: '#5b21b6', bandBg: 'rgba(124,58,237,0.04)',  label: 'Domain',         line: '#a78bfa' },
  'foundational':   { nodeBg: '#f0fdf4', border: '#059669', text: '#065f46', bandBg: 'rgba(5,150,105,0.04)',   label: 'Foundational',   line: '#6ee7b7' },
  'infrastructure': { nodeBg: '#f8fafc', border: '#94a3b8', text: '#475569', bandBg: 'rgba(148,163,184,0.06)', label: 'Infrastructure', line: '#94a3b8' },
}

export function teamColor(name: string): string {
  let h = 0
  for (let i = 0; i < name.length; i++) h = (h * 31 + name.charCodeAt(i)) & 0x7fffffff
  const palette = ['#3b82f6', '#8b5cf6', '#06b6d4', '#10b981', '#f59e0b', '#ef4444', '#ec4899', '#14b8a6', '#f97316', '#6366f1']
  return palette[h % palette.length]
}
