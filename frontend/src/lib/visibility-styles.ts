/**
 * Badge style map for capability visibility levels.
 * Provides bg and text color for badge rendering.
 */
export const VIS_BADGE: Record<string, { bg: string; text: string }> = {
  'user-facing':    { bg: '#dbeafe', text: '#1e40af' },
  'domain':         { bg: '#ede9fe', text: '#5b21b6' },
  'foundational':   { bg: '#d1fae5', text: '#065f46' },
  'infrastructure': { bg: '#f1f5f9', text: '#475569' },
}
