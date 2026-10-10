// @vitest-environment happy-dom
import type { User } from '../../../client/types.gen'
import { PiniaColada } from '@pinia/colada'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { currentUser } from '../../../composables/api'
import { i18n } from '../../../composables/i18n'
import Secrets from './secrets.vue'
import Users from './users.vue'

const api = vi.hoisted(() => ({ listUsers: vi.fn(), listSecrets: vi.fn() }))
vi.mock('../../../client/sdk.gen', async importOriginal => ({ ...await importOriginal<typeof import('../../../client/sdk.gen')>(), ...api }))

const admin: User = { id: 'admin', username: 'admin', name: 'Administrator', role: 'admin', enabled: true, locale: 'zh-CN', timezone: 'UTC', createdAt: 0, updatedAt: 0 }
const wrappers: ReturnType<typeof mount>[] = []
beforeEach(() => {
  vi.resetAllMocks()
  currentUser.value = { ...admin }
  i18n.global.locale.value = 'zh-CN'
  api.listUsers.mockResolvedValue({ data: { items: [admin, { ...admin, id: 'operator', username: 'oncall', name: 'Operations', role: 'operator' }] } })
  api.listSecrets.mockResolvedValue({ data: { items: [{ id: 'key-prod', name: 'Production token', createdAt: 0, updatedAt: 0 }, { id: 'key-staging', name: 'Staging token', createdAt: 0, updatedAt: 0 }] } })
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

it('filters members by username or display name and lets the user recover from no matches', async () => {
  const wrapper = await render(Users)
  expect(wrapper.findAll('tbody tr')).toHaveLength(2)
  const search = wrapper.find('input[type=search]')
  expect(search.exists()).toBe(true)
  await search.setValue('ONCALL')
  expect(wrapper.findAll('tbody tr')).toHaveLength(1)
  expect(wrapper.get('tbody').text()).toContain('Operations')
  await search.setValue('missing-member')
  expect(wrapper.findAll('tbody tr')).toHaveLength(0)
  const clear = wrapper.findAll('button').find(button => button.text() === i18n.global.t('common.clearFilters'))!
  await clear.trigger('click')
  expect(wrapper.findAll('tbody tr')).toHaveLength(2)
})

it('filters secret references without requesting or exposing their values', async () => {
  const wrapper = await render(Secrets)
  expect(wrapper.findAll('tbody tr')).toHaveLength(2)
  const search = wrapper.find('input[type=search]')
  expect(search.exists()).toBe(true)
  await search.setValue('KEY-PROD')
  expect(wrapper.findAll('tbody tr')).toHaveLength(1)
  expect(wrapper.get('tbody').text()).toContain('Production token')
  expect(wrapper.get('tbody').text()).not.toContain('Staging token')
})
