import type { Page } from './types'

// Published routing is a snapshot. An omitted publishedDomain on a new-format
// page means path-only publication, even when its draft now has a domain.
type PublishedEntry = Pick<Page, 'slug' | 'domain' | 'publishedSlug' | 'publishedDomain'>
export function publishedEntry(page: PublishedEntry) {
  const slug = page.publishedSlug || page.slug
  const domain = page.publishedDomain ?? (page.publishedSlug ? '' : page.domain)
  return { slug, domain, url: domain ? `https://${domain}/` : `/${slug}` }
}
