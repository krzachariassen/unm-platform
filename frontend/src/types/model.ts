export type { ParseResponse, ValidationItem, Capability, Team, Need, Service, Actor, ViewNode, ViewEdge, ViewResponse } from '@/lib/api'

export type TeamType = 'stream-aligned' | 'platform' | 'enabling' | 'complicated-subsystem'

export const TEAM_TYPE_COLORS: Record<TeamType, string> = {
  'stream-aligned': '#3b82f6',    // blue
  'platform': '#a855f7',          // purple
  'enabling': '#22c55e',          // green
  'complicated-subsystem': '#f59e0b', // amber
}

export const VIEW_TYPES = [
  { id: 'need', label: 'Need View', icon: 'Users' },
  { id: 'capability', label: 'Capability View', icon: 'Layers' },
  { id: 'realization', label: 'Realization View', icon: 'GitBranch' },
  { id: 'ownership', label: 'Ownership View', icon: 'Flag' },
  { id: 'team-topology', label: 'Team Topology', icon: 'Network' },
  { id: 'cognitive-load', label: 'Cognitive Load', icon: 'Activity' },
] as const
