// @vitest-environment happy-dom
import type { PublicPage } from '../client/types.gen'
import { mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { i18n } from '../composables/i18n'
import StatusPage from './StatusPage.vue'

const now = Date.UTC(2026, 9, 10, 12)
const page: PublicPage = {
  id: 'status',
  slug: 'service-status',
  state: 'operational',
  updatedAt: now,
  config: { brandColor: '#2563eb', colorScheme: 'light', title: 'Service status', description: '', logoUrl: '', links: [], groups: [] },
  groups: [],
  incidents: [],
  maintenance: [
    { id: 'current', name: 'Current upgrade', description: '', startsAt: now - 1000, endsAt: now + 1000 },
    { id: 'future', name: 'Upcoming upgrade', description: '', startsAt: now + 2000, endsAt: now + 3000 },
  ],
}
const wrappers: ReturnType<typeof mount>[] = []
beforeEach(() => {
  vi.useFakeTimers({ toFake: ['Date'] })
  vi.setSystemTime(now)
  i18n.global.locale.value = 'zh-CN'
})
afterEach(() => {
  wrappers.splice(0).forEach(wrapper => wrapper.unmount())
  vi.useRealTimers()
})
function render(props: { incidentId?: string, preview?: boolean } = {}) {
  const wrapper = mount(StatusPage, { props: { page, ...props }, global: { plugins: [i18n], stubs: { RouterLink: { props: ['to'], template: '<a :href="to"><slot /></a>' } } } })
  wrappers.push(wrapper)
  return wrapper
}

describe('public status page', () => {
  it('distinguishes upcoming work from maintenance that is currently in progress', () => {
    const wrapper = render()
    const windows = wrapper.findAll('.public-incident')
    expect(windows[0]!.text()).toContain(i18n.global.t('states.maintenance'))
    expect(windows[1]!.text()).toContain(i18n.global.t('maintenance.scheduled'))
    expect(windows[1]!.text()).not.toContain(i18n.global.t('states.maintenance'))
  })

  it('explains a missing incident and provides a route back to service health', () => {
    const wrapper = render({ incidentId: 'missing' })
    const heading = wrapper.find('h1')
    expect(heading.exists()).toBe(true)
    expect(heading.text()).toBe(i18n.global.t('statusPage.incidentUnavailable'))
    expect(wrapper.find('a[href="/service-status"]').text()).toBe(i18n.global.t('statusPage.backToStatusPage'))
  })

  it('exposes the public content landmark while previews remain inside the editor landmark', () => {
    expect(render().find('main').exists()).toBe(true)
    expect(render({ preview: true }).find('main').exists()).toBe(false)
  })
})
