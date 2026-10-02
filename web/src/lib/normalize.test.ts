import test from 'node:test'
import assert from 'node:assert/strict'
import { normalizeCollections } from './normalize.ts'

test('nullable Go slices become usable lists without fabricating statistics', () => {
  assert.deepEqual(
    normalizeCollections({
      items: null,
      groups: [{ monitors: null }],
      availability: { uptime: null, coverage: null, effectiveMs: 0 },
      http: { headers: null, assertions: { regex: null } },
    }),
    {
      items: [],
      groups: [{ monitors: [] }],
      availability: { uptime: null, coverage: null, effectiveMs: 0 },
      http: { headers: [], assertions: { regex: [] } },
    },
  )
})
