// @vitest-environment happy-dom
import type { Settings as OrganizationSettings, User } from '../../client/types.gen'
import { PiniaColada } from '@pinia/colada'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { currentUser } from '../../composables/api'
import { i18n } from '../../composables/i18n'
import Settings from './(admin)/settings.vue'
import Login from './login.vue'

const api = vi.hoisted(() => ({ getSettings: vi.fn(), getSetup: vi.fn(), getSession: vi.fn() }))
vi.mock('../../client/sdk.gen', async importOriginal => ({ ...await importOriginal<typeof import('../../client/sdk.gen')>(), ...api }))
vi.mock('vue-router', async importOriginal => ({ ...await importOriginal<typeof import('vue-router')>(), useRoute: () => ({ query: {} }), useRouter: () => ({ replace: vi.fn() }) }))

const wrappers: ReturnType<typeof mount>[] = []
const admin: User = { id: 'admin', username: 'admin', name: 'Administrator', role: 'admin', enabled: true, locale: 'zh-CN', timezone: 'UTC', createdAt: 0, updatedAt: 0 }
const settings: OrganizationSettings = { organizationName: 'Octopulse', timezone: 'UTC', locale: 'zh-CN', allowedDomains: [], retention: { roundDays: 30, attemptDays: 7, fiveMinuteDays: 90, historyMonths: 12 } }

beforeEach(() => {
  vi.resetAllMocks()
  currentUser.value = { ...admin }
  i18n.global.locale.value = 'zh-CN'
  api.getSettings.mockResolvedValue({ data: settings })
  api.getSetup.mockResolvedValue({ data: { required: true } })
})
afterEach(() => {
  wrappers.splice(0).forEach(wrapper => wrapper.unmount())
  document.body.innerHTML = ''
})
async function render(component: Parameters<typeof mount>[0]) {
  const wrapper = mount(component, { attachTo: document.body, global: { plugins: [createPinia(), [PiniaColada, { queryOptions: { refetchOnWindowFocus: false } }], i18n] } })
  wrappers.push(wrapper)
  await flushPromises()
  return wrapper
}

describe('configuration load recovery', () => {
  it('does not expose default organization settings when loading failed and retries the request', async () => {
    api.getSettings.mockRejectedValueOnce(new Error('Settings unavailable'))
    const wrapper = await render(Settings)
    expect(wrapper.find('[role=alert]').text()).toContain('Settings unavailable')
    expect(wrapper.find('form').exists()).toBe(false)
    const retry = wrapper.findAll('button').find(button => button.text() === i18n.global.t('async-state.retry'))
    expect(retry).toBeDefined()
    await retry!.trigger('click')
    await flushPromises()
    expect(wrapper.find('[role=alert]').exists()).toBe(false)
    expect(wrapper.find('form').exists()).toBe(true)
  })

  it('retries initialization rather than presenting the wrong sign-in mode after failure', async () => {
    api.getSetup.mockRejectedValueOnce(new Error('Setup unavailable'))
    const wrapper = await render(Login)
    expect(wrapper.find('[role=alert]').text()).toContain('Setup unavailable')
    expect(wrapper.find('form').exists()).toBe(false)
    const retry = wrapper.findAll('button').find(button => button.text() === i18n.global.t('async-state.retry'))
    expect(retry).toBeDefined()
    await retry!.trigger('click')
    await flushPromises()
    expect(wrapper.find('[role=alert]').exists()).toBe(false)
    expect(wrapper.findAll('h1').map(heading => heading.text())).toContain(i18n.global.t('login.create-your-workspace'))
    expect(wrapper.find('form').exists()).toBe(true)
  })
})
