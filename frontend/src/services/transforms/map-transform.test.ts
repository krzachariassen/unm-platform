import { describe, it, expect } from 'vitest'
import { transformMapResponse, extractRawData } from './map-transform'
import type { UNMMapViewResponse } from '@/types/views'

const mockData: UNMMapViewResponse = {
  view_type: 'unm-map',
  nodes: [
    { id: 'a1', type: 'actor', label: 'Actor 1', data: { description: '' } },
    { id: 'n1', type: 'need', label: 'Need 1', data: { is_mapped: true, outcome: '' } },
    { id: 'c1', type: 'capability', label: 'Cap 1', data: { visibility: 'domain', team_label: '', services: [], is_fragmented: false } },
  ],
  edges: [
    { id: 'e1', source: 'a1', target: 'n1', label: 'has need' },
    { id: 'e2', source: 'n1', target: 'c1', label: 'supportedBy' },
  ],
  external_deps: [],
}

describe('transformMapResponse', () => {
  it('produces rfNodes for each node type', () => {
    const result = transformMapResponse(mockData)
    const rfIds = result.rfNodes.map(n => n.id)
    expect(rfIds).toContain('a1')
    expect(rfIds).toContain('n1')
    expect(rfIds).toContain('c1')
  })

  it('produces rfEdges for demand and supply connections', () => {
    const result = transformMapResponse(mockData)
    expect(result.rfEdges.length).toBeGreaterThan(0)
    expect(result.rfEdges.every(e => e.source && e.target)).toBe(true)
  })

  it('assigns correct RF node types', () => {
    const result = transformMapResponse(mockData)
    const actorNode = result.rfNodes.find(n => n.id === 'a1')
    const needNode = result.rfNodes.find(n => n.id === 'n1')
    const capNode = result.rfNodes.find(n => n.id === 'c1')
    expect(actorNode?.type).toBe('actor')
    expect(needNode?.type).toBe('need')
    expect(capNode?.type).toBe('capability')
  })

  it('includes nodePos with all computed positions', () => {
    const result = transformMapResponse(mockData)
    expect(result.nodePos.has('a1')).toBe(true)
    expect(result.nodePos.has('c1')).toBe(true)
  })

  it('includes pending caps in layout', () => {
    const pendingCap = { id: 'pending:NewCap', type: 'capability', label: 'NewCap', data: { visibility: 'domain', services: [], isPending: true } }
    const result = transformMapResponse(mockData, [pendingCap])
    expect(result.nodePos.has('pending:NewCap')).toBe(true)
  })
})

describe('extractRawData', () => {
  it('splits nodes by type', () => {
    const raw = extractRawData(mockData)
    expect(raw.actors).toHaveLength(1)
    expect(raw.needs).toHaveLength(1)
    expect(raw.caps).toHaveLength(1)
  })

  it('splits edges by label', () => {
    const raw = extractRawData(mockData)
    expect(raw.actorToNeed).toHaveLength(1)
    expect(raw.needToCap).toHaveLength(1)
    expect(raw.capDepEdges).toHaveLength(0)
  })
})
