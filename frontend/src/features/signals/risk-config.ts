export const RISK = {
  red: {
    bg: '#fef2f2', text: '#b91c1c', dot: '#ef4444', border: '#fca5a5', strip: '#ef4444',
    stripGradient: 'linear-gradient(90deg, #ef4444 0%, #f87171 100%)', label: 'High Risk',
    badgeBg: '#fee2e2', cardGradient: 'linear-gradient(135deg, #fecaca 0%, #fef2f2 55%, #ffffff 100%)',
    iconWrap: 'linear-gradient(135deg, #fee2e2 0%, #fecaca 100%)',
  },
  amber: {
    bg: '#fffbeb', text: '#b45309', dot: '#f59e0b', border: '#fcd34d', strip: '#f59e0b',
    stripGradient: 'linear-gradient(90deg, #f59e0b 0%, #fbbf24 100%)', label: 'Elevated',
    badgeBg: '#fef3c7', cardGradient: 'linear-gradient(135deg, #fde68a 0%, #fffbeb 50%, #ffffff 100%)',
    iconWrap: 'linear-gradient(135deg, #fef3c7 0%, #fde68a 100%)',
  },
  green: {
    bg: '#f0fdf4', text: '#15803d', dot: '#22c55e', border: '#86efac', strip: '#22c55e',
    stripGradient: 'linear-gradient(90deg, #22c55e 0%, #4ade80 100%)', label: 'Healthy',
    badgeBg: '#dcfce7', cardGradient: 'linear-gradient(135deg, #bbf7d0 0%, #f0fdf4 55%, #ffffff 100%)',
    iconWrap: 'linear-gradient(135deg, #dcfce7 0%, #bbf7d0 100%)',
  },
} as const

export type RiskKey = keyof typeof RISK
export const rs = (risk: string) => RISK[risk as RiskKey] ?? RISK.green

export const tagBase: React.CSSProperties = { borderRadius: 8, padding: '4px 10px', fontSize: 11, fontWeight: 600 }

export function riskSubtitle(risk: string, findings: number) {
  return risk === 'green' ? 'No issues detected' : `${findings} finding${findings !== 1 ? 's' : ''} require${findings === 1 ? 's' : ''} attention`
}

export function buildUxSummary(ux: { needs_requiring_3plus_teams: Array<{need_name:string;team_span:number;teams?:string[]}>, needs_with_no_capability_backing: Array<{need_name:string}>, needs_at_risk: Array<{need_name:string;team_span:number}> }): string {
  const lines: string[] = []
  if (ux.needs_requiring_3plus_teams.length > 0) lines.push(`Cross-team needs (${ux.needs_requiring_3plus_teams.length}): ${ux.needs_requiring_3plus_teams.map(n => `"${n.need_name}" (span ${n.team_span}, teams: ${n.teams?.join(', ') ?? '?'})`).join('; ')}`)
  if (ux.needs_with_no_capability_backing.length > 0) lines.push(`Unbacked needs (${ux.needs_with_no_capability_backing.length}): ${ux.needs_with_no_capability_backing.map(n => `"${n.need_name}"`).join(', ')}`)
  if (ux.needs_at_risk.length > 0) lines.push(`At-risk needs (${ux.needs_at_risk.length}): ${ux.needs_at_risk.map(n => `"${n.need_name}" (span ${n.team_span})`).join(', ')}`)
  return lines.join('\n')
}

export function buildArchSummary(arch: { capabilities_not_connected_to_any_need: Array<{capability_name:string}>, capabilities_fragmented_across_teams: Array<{capability_name:string;team_count?:number;teams?:string[]}>, user_facing_caps_with_cross_team_services: Array<{capability_name:string}> }): string {
  const lines: string[] = []
  if (arch.capabilities_not_connected_to_any_need.length > 0) lines.push(`Unlinked capabilities (${arch.capabilities_not_connected_to_any_need.length}): ${arch.capabilities_not_connected_to_any_need.map(c => `"${c.capability_name}"`).join(', ')}`)
  if (arch.capabilities_fragmented_across_teams.length > 0) lines.push(`Fragmented capabilities (${arch.capabilities_fragmented_across_teams.length}): ${arch.capabilities_fragmented_across_teams.map(c => `"${c.capability_name}" (${c.team_count} teams: ${c.teams?.join(', ') ?? '?'})`).join('; ')}`)
  if (arch.user_facing_caps_with_cross_team_services.length > 0) lines.push(`Cross-team user-facing caps (${arch.user_facing_caps_with_cross_team_services.length}): ${arch.user_facing_caps_with_cross_team_services.map(c => `"${c.capability_name}"`).join(', ')}`)
  return lines.join('\n')
}

export function buildOrgSummary(org: { top_teams_by_structural_load: Array<{team_name:string;overall_level?:string;service_count?:number;capability_count?:number}>, critical_bottleneck_services: Array<{service_name:string;fan_in:number}>, low_coherence_teams: Array<{team_name:string;coherence_score?:number}> }): string {
  const lines: string[] = []
  if (org.top_teams_by_structural_load.length > 0) lines.push(`High structural load teams (${org.top_teams_by_structural_load.length}): ${org.top_teams_by_structural_load.map(t => `"${t.team_name}" (${t.overall_level}, ${t.service_count} svcs, ${t.capability_count} caps)`).join('; ')}`)
  if (org.critical_bottleneck_services.length > 0) lines.push(`Critical bottleneck services (${org.critical_bottleneck_services.length}): ${org.critical_bottleneck_services.map(s => `"${s.service_name}" (fan-in: ${s.fan_in})`).join(', ')}`)
  if (org.low_coherence_teams.length > 0) lines.push(`Low coherence teams (${org.low_coherence_teams.length}): ${org.low_coherence_teams.map(t => `"${t.team_name}" (${t.coherence_score != null ? Math.round(t.coherence_score * 100) : '?'}%)`).join(', ')}`)
  return lines.join('\n')
}
