import type { TeamTopologyViewResponse } from '@/types/views'

export type TeamTopologyTeam = TeamTopologyViewResponse['teams'][number]
export type TeamTopologyInteraction = TeamTopologyViewResponse['interactions'][number]

export const TEAM_TYPES: Record<string, {
  label: string; accent: string; bg: string; border: string
  gradientFrom: string; gradientTo: string; zoneBg: string
}> = {
  'platform': {
    label: 'Platform', accent: '#7c3aed', bg: '#faf5ff', border: '#ddd6fe',
    gradientFrom: '#7c3aed', gradientTo: '#a78bfa', zoneBg: 'rgba(124,58,237,0.03)',
  },
  'stream-aligned': {
    label: 'Stream-aligned', accent: '#1d4ed8', bg: '#eff6ff', border: '#bfdbfe',
    gradientFrom: '#1d4ed8', gradientTo: '#60a5fa', zoneBg: 'rgba(29,78,216,0.03)',
  },
  'complicated-subsystem': {
    label: 'Complicated Subsystem', accent: '#b45309', bg: '#fffbeb', border: '#fde68a',
    gradientFrom: '#b45309', gradientTo: '#fbbf24', zoneBg: 'rgba(180,83,9,0.03)',
  },
  'enabling': {
    label: 'Enabling', accent: '#15803d', bg: '#f0fdf4', border: '#bbf7d0',
    gradientFrom: '#15803d', gradientTo: '#4ade80', zoneBg: 'rgba(21,128,61,0.03)',
  },
}

export const TEAM_TYPE_DESCRIPTIONS: Record<string, string> = {
  'stream-aligned': 'Aligned to a flow of work from a business domain segment',
  'platform': 'Provides internal services to reduce cognitive load of other teams',
  'enabling': 'Helps other teams adopt new practices or technologies',
  'complicated-subsystem': 'Owns a subsystem requiring deep specialist knowledge',
}

export const INTERACTION_STYLE: Record<string, { label: string; bg: string; text: string; border: string; color: string }> = {
  'collaboration':  { label: 'Collaboration',  bg: '#dbeafe', text: '#1e40af', border: '#bfdbfe', color: '#1d4ed8' },
  'x-as-a-service': { label: 'X-as-a-Service', bg: '#ede9fe', text: '#5b21b6', border: '#ddd6fe', color: '#7c3aed' },
  'facilitating':   { label: 'Facilitating',   bg: '#d1fae5', text: '#065f46', border: '#a7f3d0', color: '#15803d' },
}

export const NODE_W = 220
export const NODE_H = 108
export const COL_PAD = 24
export const COL_GAP = 180
export const ROW_GAP = 18

export const COLUMNS = [
  { types: ['platform'],              label: 'Platform',            x: COL_PAD },
  { types: ['stream-aligned'],        label: 'Stream-aligned',      x: COL_PAD + NODE_W + COL_GAP },
  { types: ['complicated-subsystem', 'enabling'], label: 'Subsystem / Enabling', x: COL_PAD + (NODE_W + COL_GAP) * 2 },
]

export const HEADER_H = 56

export function getType(t: string) { return TEAM_TYPES[t] ?? TEAM_TYPES['stream-aligned'] }
export function getIx(m: string)   { return INTERACTION_STYLE[m] ?? { label: m, bg: '#f1f5f9', text: '#475569', border: '#e2e8f0', color: '#64748b' } }
