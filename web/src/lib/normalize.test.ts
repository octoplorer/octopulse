import { expect, it } from 'vitest'
import { normalizeCollections } from './normalize.ts'

it('nullable Go slices become usable lists without fabricating statistics', () => {
  expect(
    normalizeCollections({
      items: null,
      groups: [{ monitors: null }],
      availability: { uptime: null, coverage: null, effectiveMs: 0 },
      http: { headers: null, assertions: { regex: null } },
    }),
  ).toStrictEqual({
    items: [],
    groups: [{ monitors: [] }],
    availability: { uptime: null, coverage: null, effectiveMs: 0 },
    http: { headers: [], assertions: { regex: [] } },
  })
})
