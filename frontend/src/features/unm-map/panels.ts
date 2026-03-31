import type { PNode, Conn, PanelItem, PanelField } from './types'

export function buildNodePanel(
  node: PNode,
  actorNeedEdges: Array<{ source: string; target: string }>,
  needCapEdges: Array<{ source: string; target: string }>,
  nodePos: Map<string, PNode>,
  insights: Record<string, { explanation: string; suggestion?: string }>,
  slug: (s: string) => string,
): PanelItem {
  if (node.type === 'ext-dep' && node.extDep) {
    const dep = node.extDep
    const severityColor = dep.is_critical ? '#ef4444' : dep.is_warning ? '#f59e0b' : '#6b7280'
    return {
      title: node.label,
      badge: { text: 'External Dependency', color: severityColor },
      fields: [
        { label: 'Description', value: dep.description ?? '' },
        ...(dep.is_critical ? [{ label: 'Severity', value: 'Critical — this dependency is flagged as critical' }] : []),
        ...(dep.is_warning && !dep.is_critical ? [{ label: 'Severity', value: 'Warning — this dependency has issues' }] : []),
        { label: 'Services using this', value: dep.services.join('\n') },
        { label: 'Service count', value: String(dep.service_count) },
      ],
    }
  }

  if (node.type === 'actor') {
    const myNeedEdges = actorNeedEdges.filter(e => e.source === node.id)
    const needFields: PanelField[] = []
    for (const ne of myNeedEdges) {
      const needNode = nodePos.get(ne.target)
      if (!needNode) continue
      const capEdges = needCapEdges.filter(e => e.source === ne.target)
      const capNames = capEdges.map(ce => nodePos.get(ce.target)?.label).filter(Boolean)
      const isMapped = needNode.isMapped !== false
      const statusBadge = isMapped ? '[Mapped]' : '[Unmapped]'
      const capText = capNames.length > 0 ? capNames.join(', ') : 'No capabilities linked'
      needFields.push({ label: `${statusBadge} ${needNode.label}`, value: capText })
    }
    return {
      title: node.label, badge: { text: 'Actor', color: '#3b82f6' },
      fields: [
        { label: 'Description', value: node.description ?? '' },
        ...(needFields.length > 0
          ? [{ label: 'Needs', value: '' }, ...needFields]
          : [{ label: 'Needs', value: 'No needs defined for this actor' }]),
      ],
    }
  }

  if (node.type === 'need') {
    return {
      title: node.label, badge: { text: 'Need', color: node.isMapped === false ? '#ef4444' : '#2563eb' },
      fields: [
        { label: 'Outcome', value: node.outcome ?? '' },
        { label: 'Status', value: node.isMapped === false ? 'Unmapped — no capability supports this need' : 'Mapped' },
      ],
    }
  }

  // Capability
  const nodeSlug = slug(node.label)
  const ai = insights[`cap:${nodeSlug}`] ?? insights[`cap-fragmented:${nodeSlug}`] ?? insights[`cap-disconnected:${nodeSlug}`]
  const aiFields: PanelField[] = ai
    ? [{ label: 'AI Insight', value: ai.explanation }, ...(ai.suggestion ? [{ label: 'Recommendation', value: ai.suggestion }] : [])]
    : []
  const uniqueTeamsForPanel = [...new Set((node.svcs ?? []).map(s => s.teamName).filter(Boolean))]
  const isPendingNode = node.id.startsWith('pending:')
  const svcsText = node.svcs?.map(s => `${s.label}${s.teamName ? ` (${s.teamName})` : ''}`).join('\n') ?? ''

  return {
    title: node.label,
    badge: {
      text: isPendingNode ? 'Pending' : (node.vis ?? 'capability'),
      color: isPendingNode ? '#f59e0b' : '#94a3b8',
    },
    fields: [
      ...(isPendingNode ? [{ label: '⏳ Status', value: 'Staged but not yet committed.' }] : []),
      { label: 'Description', value: node.description ?? '' },
      { label: 'Visibility', value: node.vis ?? '' },
      { label: 'Owning Team', value: node.team?.label ?? 'Unowned' },
      { label: 'Team Type', value: node.team?.type ?? '' },
      { label: 'Realized By', value: svcsText || 'No service linked yet' },
      ...(node.crossTeam ? [{ label: '⚠ Multi-team', value: `Services owned by ${uniqueTeamsForPanel.length} different teams: ${uniqueTeamsForPanel.join(', ')}` }] : []),
      ...(node.isFragmented ? [{ label: '⚠ Fragmented', value: 'Multiple teams own services for this capability' }] : []),
      ...aiFields,
    ],
  }
}

export function buildConnPanel(conn: Conn, nodePos: Map<string, PNode>): PanelItem {
  const src = nodePos.get(conn.sourceId)
  const tgt = nodePos.get(conn.targetId)
  const descField = conn.description ? [{ label: 'Description', value: conn.description }] : []
  if (conn.edgeType === 'actor-need') {
    return { title: 'Actor → Need', badge: { text: 'Demand', color: '#3b82f6' }, fields: [{ label: 'Actor', value: src?.label ?? '' }, { label: 'Need', value: tgt?.label ?? '' }, ...descField] }
  } else if (conn.edgeType === 'need-capability') {
    return { title: 'Need → Capability', badge: { text: 'Support', color: '#6366f1' }, fields: [{ label: 'Need', value: src?.label ?? '' }, { label: 'Capability', value: tgt?.label ?? '' }, ...descField] }
  } else if (conn.edgeType === 'ext-dep') {
    return { title: 'External Dependency', badge: { text: 'External', color: '#f59e0b' }, fields: [{ label: 'Capability', value: src?.label ?? '' }, { label: 'Depends on', value: tgt?.label ?? '' }, ...descField] }
  }
  return { title: 'Capability Dependency', badge: { text: 'Supply-side', color: '#7c3aed' }, fields: [{ label: 'From', value: src?.label ?? '' }, { label: 'Depends on', value: tgt?.label ?? '' }, ...descField] }
}
