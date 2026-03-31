import { describe, it, expect } from 'vitest'
import { computeChain, buildChainData } from './chain'
import type { ChainData } from './types'

function makeChainData(): ChainData {
  const actorToNeed = [
    { id: 'e1', source: 'actor1', target: 'need1', label: 'has need' },
    { id: 'e2', source: 'actor1', target: 'need2', label: 'has need' },
  ]
  const needToCap = [
    { id: 'e3', source: 'need1', target: 'cap1', label: 'supportedBy' },
    { id: 'e4', source: 'need2', target: 'cap2', label: 'supportedBy' },
    { id: 'e5', source: 'need1', target: 'cap2', label: 'supportedBy' },
  ]
  const capDepEdges = [{ from: 'cap2', to: 'cap3' }]
  return buildChainData(actorToNeed, needToCap, capDepEdges, [])
}

describe('computeChain', () => {
  it('includes actor and all reachable nodes when clicking actor', () => {
    const chain = computeChain('actor1', 'actor', makeChainData())
    expect(chain.has('actor1')).toBe(true)
    expect(chain.has('need1')).toBe(true)
    expect(chain.has('need2')).toBe(true)
    expect(chain.has('cap1')).toBe(true)
    expect(chain.has('cap2')).toBe(true)
    expect(chain.has('cap3')).toBe(true)
  })

  it('includes need, its actor, and downstream caps when clicking need', () => {
    const chain = computeChain('need1', 'need', makeChainData())
    expect(chain.has('need1')).toBe(true)
    expect(chain.has('actor1')).toBe(true)
    expect(chain.has('cap1')).toBe(true)
    expect(chain.has('cap2')).toBe(true)
  })

  it('includes capability and upstream/downstream when clicking capability', () => {
    const chain = computeChain('cap2', 'capability', makeChainData())
    expect(chain.has('cap2')).toBe(true)
    expect(chain.has('need1')).toBe(true)
    expect(chain.has('need2')).toBe(true)
    expect(chain.has('actor1')).toBe(true)
    expect(chain.has('cap3')).toBe(true) // downstream via cap dep
  })

  it('includes only ext-dep and its linked caps when clicking ext-dep', () => {
    const chainData = buildChainData([], [], [], [{ sourceId: 'cap1', targetId: 'ext-dep:d1' }])
    const chain = computeChain('ext-dep:d1', 'ext-dep', chainData)
    expect(chain.has('ext-dep:d1')).toBe(true)
    expect(chain.has('cap1')).toBe(true)
  })

  it('includes just the clicked id if no connections', () => {
    const emptyChain = buildChainData([], [], [], [])
    const chain = computeChain('solo', 'capability', emptyChain)
    expect(chain.size).toBe(1)
    expect(chain.has('solo')).toBe(true)
  })
})

describe('buildChainData', () => {
  it('builds bidirectional extDep maps', () => {
    const data = buildChainData([], [], [], [{ sourceId: 'cap1', targetId: 'ext-dep:d1' }])
    expect(data.extDepToCapIds.get('ext-dep:d1')).toContain('cap1')
    expect(data.capToExtDepIds.get('cap1')).toContain('ext-dep:d1')
  })
})
