// @vitest-environment happy-dom
import type { System, SystemsResponse, User } from '../../../client/types.gen'
import { PiniaColada } from '@pinia/colada'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { currentUser } from '../../../composables/api'
import { i18n } from '../../../composables/i18n'
import Servers from './servers.vue'

const api = vi.hoisted(() => ({ listBeszelSystems: vi.fn(), listSecrets: vi.fn() }))
vi.mock('../../../client/sdk.gen', async importOriginal => ({ ...await importOriginal<typeof import('../../../client/sdk.gen')>(), ...api }))
vi.mock('../../../composables/notices', () => ({ notify: vi.fn() }))

const wrappers: ReturnType<typeof mount>[] = []
const admin: User = { id: 'admin', username: 'admin', name: 'Administrator', role: 'admin', enabled: true, locale: 'zh-CN', timezone: 'UTC', createdAt: 0, updatedAt: 0 }
const server: System = { id: 'server', name: 'API server', host: 'api.example.com', cpu: 1, memory: 2, disk: 3, stale: false, status: 'up', updatedAt: 1000, info: { agentVersion: '0.20', cores: 1, cpuModel: '', hostname: '', kernel: '', threads: 1, uptimeSeconds: 60 } }
const connected: SystemsResponse = { items: [server], source: 'https://beszel.example.com', stale: false, syncedAt: 1000, version: '0.20' }

beforeEach(() => {
  vi.resetAllMocks()
  currentUser.value = { ...admin }
  i18n.global.locale.value = 'zh-CN'
  api.listSecrets.mockResolvedValue({ data: { items: [] } })
})
afterEach(() => {
  wrappers.splice(0).forEach(wrapper => wrapper.unmount())
  document.body.innerHTML = ''
})
async function render() {
  const wrapper = mount(Servers, { attachTo: document.body, global: { plugins: [createPinia(), [PiniaColada, { queryOptions: { refetchOnWindowFocus: false } }], i18n] } })
  wrappers.push(wrapper)
  await flushPromises()
  return wrapper
}
async function retry(wrapper: ReturnType<typeof mount>) {
  const button = wrapper.findAll('button').find(button => button.text() === i18n.global.t('asyncState.retry'))
  expect(button).toBeDefined()
  await button!.trigger('click')
  await flushPromises()
}

describe('beszel connection state', () => {
  it('offers setup without failure or stale-data indicators when the integration is disabled', async () => {
    const disabled: SystemsResponse = { items: [], source: '', stale: true, syncedAt: 0, version: '', error: 'Beszel integration is disabled' }
    api.listBeszelSystems.mockResolvedValue({ data: disabled })
    const wrapper = await render()
    expect(wrapper.text()).toContain(i18n.global.t('servers.configureConnection'))
    expect(wrapper.find('[role=alert]').exists()).toBe(false)
    expect(wrapper.text()).not.toContain(disabled.error)
    expect(wrapper.text()).not.toContain(i18n.global.t('servers.dataIsStale'))
    expect(wrapper.text()).not.toContain(i18n.global.t('servers.lastSync'))
    expect(wrapper.text()).not.toContain(i18n.global.t('asyncState.retry'))
    expect(wrapper.text()).not.toContain(i18n.global.t('common.refresh'))
  })

  it('retains an upstream failure and lets retry recover the server list', async () => {
    api.listBeszelSystems.mockResolvedValueOnce({ data: { ...connected, items: [], stale: true, error: 'Beszel Hub is unavailable' } }).mockResolvedValue({ data: connected })
    const wrapper = await render()
    expect(wrapper.get('[role=alert]').text()).toContain('Beszel Hub is unavailable')
    expect(wrapper.text()).toContain(i18n.global.t('servers.dataIsStale'))
    expect(wrapper.text()).not.toContain(i18n.global.t('servers.configureConnection'))
    await retry(wrapper)
    expect(wrapper.find('[role=alert]').exists()).toBe(false)
    expect(wrapper.text()).toContain(server.name)
    expect(wrapper.text()).not.toContain(i18n.global.t('servers.dataIsStale'))
  })

  it('lets retry recover an unexpected request failure', async () => {
    api.listBeszelSystems.mockRejectedValueOnce(new Error('Connection reset')).mockResolvedValue({ data: connected })
    const wrapper = await render()
    expect(wrapper.get('[role=alert]').text()).toContain('Connection reset')
    await retry(wrapper)
    expect(wrapper.find('[role=alert]').exists()).toBe(false)
    expect(wrapper.text()).toContain(server.name)
  })
})
