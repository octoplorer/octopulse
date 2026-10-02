import { expect, it } from 'vitest'
import { publishedEntry } from './pages.ts'

it('public entry remains on the published route while a draft moves', () => {
  expect(
    publishedEntry({
      slug: 'draft',
      domain: 'draft.example.com',
      publishedSlug: 'live',
      publishedDomain: 'live.example.com',
    }),
  ).toStrictEqual({ slug: 'live', domain: 'live.example.com', url: 'https://live.example.com/' })
})
it('an omitted snapshot domain does not publish the draft domain', () => {
  expect(
    publishedEntry({ slug: 'draft', domain: 'draft.example.com', publishedSlug: 'live' }),
  ).toStrictEqual({ slug: 'live', domain: '', url: '/live' })
})
it('old page records use their existing entry until republished', () => {
  expect(publishedEntry({ slug: 'legacy', domain: 'status.example.com' })).toStrictEqual({
    slug: 'legacy',
    domain: 'status.example.com',
    url: 'https://status.example.com/',
  })
})
