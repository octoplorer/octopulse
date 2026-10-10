import { expect, it } from 'vitest'
import { computeEdges, computePositions } from './flow-layout'

it('connects parallel branch entry and exit nodes to adjacent steps', () => {
  const state = { tree: { kind: 'list' as const, children: [{ kind: 'node' as const, id: 'a' }, { kind: 'parallel' as const, children: [{ kind: 'node' as const, id: 'b' }, { kind: 'node' as const, id: 'c' }] }, { kind: 'node' as const, id: 'd' }] }, nodes: { a: { width: 100, height: 40 }, b: { width: 100, height: 40 }, c: { width: 100, height: 40 }, d: { width: 100, height: 40 } }, align: 'start' as const, orientation: 'horizontal' as const }
  expect(computeEdges(state)).toEqual([['a', 'b'], ['a', 'c'], ['b', 'd'], ['c', 'd']])
  const positions = computePositions(state)
  expect(positions.b?.x).toBe(positions.c?.x)
  expect(positions.c?.y).toBeGreaterThan(positions.b?.y ?? 0)
  expect(positions.d?.x).toBeGreaterThan(positions.b?.x ?? 0)
})
