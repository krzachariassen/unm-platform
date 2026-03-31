import { describe, it, expect } from 'vitest'
import { buildNodePanel, buildConnPanel } from './panels'
import type { PNode, Conn } from './types'

const slug = (s: string) => s.toLowerCase().replace(/\s+/g, '-')

function makeNodePos(nodes: PNode[]): Map<string, PNode> {
  return new Map(nodes.map(n => [n.id, n]))
}

describe('buildNodePanel', () => {
  it('builds panel for actor', () => {
    const actor: PNode = { id: 'a1', label: 'Actor', type: 'actor', x: 0, y: 0, w: 138, h: 46 }
    const need: PNode = { id: 'n1', label: 'Need 1', type: 'need', x: 0, y: 0, w: 150, h: 52, isMapped: true }
    const panel = buildNodePanel(actor, [{ source: 'a1', target: 'n1' }], [], makeNodePos([actor, need]), {}, slug)
    expect(panel.title).toBe('Actor')
    expect(panel.badge?.text).toBe('Actor')
    expect(panel.fields.some(f => f.label.includes('Need 1'))).toBe(true)
  })

  it('builds panel for need (unmapped)', () => {
    const need: PNode = { id: 'n1', label: 'My Need', type: 'need', x: 0, y: 0, w: 150, h: 52, isMapped: false }
    const panel = buildNodePanel(need, [], [], makeNodePos([need]), {}, slug)
    expect(panel.badge?.color).toBe('#ef4444')
    expect(panel.fields.some(f => f.value.includes('Unmapped'))).toBe(true)
  })

  it('builds panel for capability with AI insight', () => {
    const cap: PNode = { id: 'c1', label: 'Auth Service', type: 'capability', x: 0, y: 0, w: 172, h: 96, vis: 'domain' }
    const insights = { 'cap:auth-service': { explanation: 'AI says good', suggestion: 'Keep it up' } }
    const panel = buildNodePanel(cap, [], [], makeNodePos([cap]), insights, slug)
    expect(panel.fields.some(f => f.value === 'AI says good')).toBe(true)
    expect(panel.fields.some(f => f.value === 'Keep it up')).toBe(true)
  })

  it('builds panel for ext-dep', () => {
    const extDep = { id: 'dep1', name: 'Stripe', service_count: 2, services: ['payment-svc'], is_critical: true, is_warning: false }
    const node: PNode = { id: 'ext-dep:dep1', label: 'Stripe', type: 'ext-dep', x: 0, y: 0, w: 150, h: 40, extDep }
    const panel = buildNodePanel(node, [], [], makeNodePos([node]), {}, slug)
    expect(panel.badge?.text).toBe('External Dependency')
    expect(panel.badge?.color).toBe('#ef4444')
  })
})

describe('buildConnPanel', () => {
  const actor: PNode = { id: 'a1', label: 'Actor', type: 'actor', x: 0, y: 0, w: 138, h: 46 }
  const need: PNode = { id: 'n1', label: 'Need 1', type: 'need', x: 0, y: 0, w: 150, h: 52 }
  const cap: PNode = { id: 'c1', label: 'Cap', type: 'capability', x: 0, y: 0, w: 172, h: 96 }
  const nodePos = makeNodePos([actor, need, cap])

  it('builds actor-need conn panel', () => {
    const conn: Conn = { x1: 0, y1: 0, x2: 0, y2: 0, color: '#3b82f6', sourceId: 'a1', targetId: 'n1', edgeType: 'actor-need' }
    const panel = buildConnPanel(conn, nodePos)
    expect(panel.title).toBe('Actor → Need')
    expect(panel.fields.some(f => f.label === 'Actor')).toBe(true)
  })

  it('builds need-capability conn panel', () => {
    const conn: Conn = { x1: 0, y1: 0, x2: 0, y2: 0, color: '#6366f1', sourceId: 'n1', targetId: 'c1', edgeType: 'need-capability' }
    const panel = buildConnPanel(conn, nodePos)
    expect(panel.title).toBe('Need → Capability')
  })

  it('includes description if provided', () => {
    const conn: Conn = { x1: 0, y1: 0, x2: 0, y2: 0, color: '#7c3aed', sourceId: 'c1', targetId: 'n1', edgeType: 'cap-dep', description: 'depends on auth' }
    const panel = buildConnPanel(conn, nodePos)
    expect(panel.fields.some(f => f.value === 'depends on auth')).toBe(true)
  })
})
