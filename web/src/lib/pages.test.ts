import test from 'node:test'
import assert from 'node:assert/strict'
import { publishedEntry } from './pages.ts'

test('public entry remains on the published route while a draft moves', () => {
  assert.deepEqual(
    publishedEntry({
      slug: 'draft',
      domain: 'draft.example.com',
      publishedSlug: 'live',
      publishedDomain: 'live.example.com',
    }),
    { slug: 'live', domain: 'live.example.com', url: 'https://live.example.com/' },
  )
})
test('an omitted snapshot domain does not publish the draft domain', () => {
  assert.deepEqual(
    publishedEntry({ slug: 'draft', domain: 'draft.example.com', publishedSlug: 'live' }),
    { slug: 'live', domain: '', url: '/live' },
  )
})
test('old page records use their existing entry until republished', () => {
  assert.deepEqual(publishedEntry({ slug: 'legacy', domain: 'status.example.com' }), {
    slug: 'legacy',
    domain: 'status.example.com',
    url: 'https://status.example.com/',
  })
})
