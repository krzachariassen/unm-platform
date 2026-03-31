import { describe, it, expect } from 'vitest'
import { capCardHeight, buildLayout } from './layout'

describe('capCardHeight', () => {
  it('returns minimum height for zero services', () => {
    const h = capCardHeight(0, 10)
    expect(h).toBeGreaterThan(20)
  })

  it('increases with more services', () => {
    const h0 = capCardHeight(0, 10)
    const h3 = capCardHeight(3, 10)
    expect(h3).toBeGreaterThan(h0)
  })

  it('caps visible services at 3 and adds "+N more" line', () => {
    const h3 = capCardHeight(3, 10)
    const h5 = capCardHeight(5, 10)
    expect(h5).toBeGreaterThan(h3)
  })
})

describe('buildLayout', () => {
  const actor = { id: 'a1', type: 'actor', label: 'Actor 1', data: { description: 'desc' } }
  const need = { id: 'n1', type: 'need', label: 'Need 1', data: { is_mapped: true, outcome: 'outcome' } }
  const cap = { id: 'c1', type: 'capability', label: 'Cap 1', data: { visibility: 'domain', team_label: 'Team A', team_type: 'stream-aligned', services: [], is_fragmented: false } }
  const actorToNeed = [{ id: 'e1', source: 'a1', target: 'n1', label: 'has need' }]
  const needToCap = [{ id: 'e2', source: 'n1', target: 'c1', label: 'supportedBy' }]

  it('produces pnodes for actors, needs, caps', () => {
    const result = buildLayout([actor], [need], [cap], actorToNeed, needToCap, [], [])
    const ids = result.pnodes.map(n => n.id)
    expect(ids).toContain('a1')
    expect(ids).toContain('n1')
    expect(ids).toContain('c1')
  })

  it('produces actor-need and need-cap conns', () => {
    const result = buildLayout([actor], [need], [cap], actorToNeed, needToCap, [], [])
    const actorNeedConns = result.conns.filter(c => c.edgeType === 'actor-need')
    const needCapConns = result.conns.filter(c => c.edgeType === 'need-capability')
    expect(actorNeedConns).toHaveLength(1)
    expect(needCapConns).toHaveLength(1)
  })

  it('assigns y positions based on layer', () => {
    const result = buildLayout([actor], [need], [cap], actorToNeed, needToCap, [], [])
    const actorNode = result.pnodes.find(n => n.id === 'a1')!
    const needNode = result.pnodes.find(n => n.id === 'n1')!
    const capNode = result.pnodes.find(n => n.id === 'c1')!
    expect(actorNode.y).toBeLessThan(needNode.y)
    expect(needNode.y).toBeLessThan(capNode.y)
  })

  it('places ext-dep nodes below capability bands', () => {
    const extDep = { id: 'dep1', name: 'ExtDep', service_count: 1, services: [], is_critical: false, is_warning: false }
    const result = buildLayout([actor], [need], [cap], actorToNeed, needToCap, [], [extDep])
    const extDepNode = result.pnodes.find(n => n.id === 'ext-dep:dep1')
    expect(extDepNode).toBeDefined()
    expect(extDepNode!.y).toBeGreaterThan(400)
  })

  it('returns nodePos map with all nodes', () => {
    const result = buildLayout([actor], [need], [cap], actorToNeed, needToCap, [], [])
    expect(result.nodePos.has('a1')).toBe(true)
    expect(result.nodePos.has('n1')).toBe(true)
    expect(result.nodePos.has('c1')).toBe(true)
  })
})
