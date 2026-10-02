import { expect, test } from 'vitest'
import { publishedEntry } from './pages.ts'

test('public entry remains on the published route while a draft moves', () => {
  expect(
    publishedEntry({
      slug: 'draft',
      domain: 'draft.example.com',
      publishedSlug: 'live',
      publishedDomain: 'live.example.com',
    }),
  ).toStrictEqual({ slug: 'live', domain: 'live.example.com', url: 'https://live.example.com/' })
})
test('an omitted snapshot domain does not publish the draft domain', () => {
  expect(
    publishedEntry({ slug: 'draft', domain: 'draft.example.com', publishedSlug: 'live' }),
  ).toStrictEqual({ slug: 'live', domain: '', url: '/live' })
})
test('old page records use their existing entry until republished', () => {
  expect(publishedEntry({ slug: 'legacy', domain: 'status.example.com' })).toStrictEqual({
    slug: 'legacy',
    domain: 'status.example.com',
    url: 'https://status.example.com/',
  })
})
