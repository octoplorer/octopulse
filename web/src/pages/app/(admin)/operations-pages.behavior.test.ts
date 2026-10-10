// @vitest-environment happy-dom
import type { Channel, DeliveryView, Incident, Maintenance as MaintenanceWindow, User } from '../../../client/types.gen'
import { PiniaColada } from '@pinia/colada'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { currentUser } from '../../../composables/api'
import { i18n } from '../../../composables/i18n'
import Incidents from './incidents.vue'
import Maintenance from './maintenance.vue'
import Notifications from './notifications.vue'

const api = vi.hoisted(() => ({
  listMaintenance: vi.fn(),
  listIncidents: vi.fn(),
  listChannels: vi.fn(),
  listDeliveries: vi.fn(),
  listSecrets: vi.fn(),
  listMonitors: vi.fn(),
  listPages: vi.fn(),
}))
vi.mock('../../../client/sdk.gen', async importOriginal => ({ ...await importOriginal<typeof import('../../../client/sdk.gen')>(), ...api }))
vi.mock('../../../composables/notices', () => ({ notify: vi.fn() }))

const wrappers: ReturnType<typeof mount>[] = []
const now = Date.UTC(2026, 9, 10, 12)
const admin: User = { id: 'admin', username: 'admin', name: 'Administrator', role: 'admin', enabled: true, locale: 'zh-CN', timezone: 'UTC', createdAt: 0, updatedAt: 0 }
const channels: Channel[] = [
  { id: 'channel-1', name: 'Pager', serviceUrlSecretId: 'secret-1', enabled: true, createdAt: 0, updatedAt: 0 },
  { id: 'channel-2', name: 'Team chat', serviceUrlSecretId: 'secret-2', enabled: true, createdAt: 0, updatedAt: 0 },
]
const deliveries: DeliveryView[] = [
  { id: 'old-failure', channelId: 'channel-1', monitorId: '', eventId: 'event-1', kind: 'test', status: 'failed', attempts: 3, lastError: 'Connection refused: notification gateway unavailable', createdAt: 1, dueAt: 1 },
  { id: 'recent-success', channelId: 'channel-2', monitorId: '', eventId: 'event-2', kind: 'up', status: 'sent', attempts: 1, lastError: '', createdAt: 3, dueAt: 3 },
  { id: 'recent-failure', channelId: 'channel-2', monitorId: '', eventId: 'event-3', kind: 'down', status: 'failed', attempts: 2, lastError: 'Rate limit exceeded', createdAt: 2, dueAt: 2 },
]
function maintenance(id: string, startsAt: number, endsAt: number): MaintenanceWindow {
  return { id, name: id, description: '', startsAt, endsAt, timezone: 'UTC', monitorIds: [], pageIds: ['page-1'], createdAt: 0, updatedAt: 0 }
}
function incident(id: string, status: Incident['status'], updatedAt: number, body = ''): Incident {
  return { id, title: id, body, status, impact: 'none', monitorIds: [], pageIds: [], updates: [], createdAt: 0, updatedAt, resolvedAt: status === 'resolved' ? updatedAt : 0 }
}

beforeEach(() => {
  vi.resetAllMocks()
  vi.useFakeTimers({ toFake: ['Date'] })
  vi.setSystemTime(now)
  currentUser.value = { ...admin }
  i18n.global.locale.value = 'zh-CN'
  api.listMaintenance.mockResolvedValue({ data: { items: [] } })
  api.listIncidents.mockResolvedValue({ data: { items: [] } })
  api.listChannels.mockResolvedValue({ data: { items: channels } })
  api.listDeliveries.mockResolvedValue({ data: { items: deliveries } })
  api.listSecrets.mockResolvedValue({ data: { items: [] } })
  api.listMonitors.mockResolvedValue({ data: { items: [] } })
  api.listPages.mockResolvedValue({ data: { items: [] } })
})
afterEach(() => {
  wrappers.splice(0).forEach(wrapper => wrapper.unmount())
  document.body.innerHTML = ''
  vi.useRealTimers()
})
async function render(component: Parameters<typeof mount>[0]) {
  const wrapper = mount(component, { attachTo: document.body, global: { plugins: [createPinia(), [PiniaColada, { queryOptions: { refetchOnWindowFocus: false } }], i18n], stubs: { RouterLink: { template: '<a><slot /></a>' } } } })
  wrappers.push(wrapper)
  await flushPromises()
  return wrapper
}
async function selectTab(wrapper: ReturnType<typeof mount>, value: string) {
  const tab = wrapper.find(`[role=tab][data-value=${value}]`)
  expect(tab.exists()).toBe(true)
  await tab.trigger('click')
  await flushPromises()
}
function rows(wrapper: ReturnType<typeof mount>) {
  return wrapper.findAll('[role=tabpanel]:not([hidden]) tbody tr').map(row => row.text())
}
async function selectOption(label: string, value: string) {
  const trigger = document.querySelector<HTMLButtonElement>(`button[aria-label="${label}"]`)
  expect(trigger).not.toBeNull()
  trigger!.click()
  await flushPromises()
  const option = document.querySelector<HTMLElement>(`[role=option][data-value="${value}"]`)
  expect(option).not.toBeNull()
  option!.click()
  await flushPromises()
}

describe('operations list pages', () => {
  it('sorts maintenance by current work then upcoming time and separates completed windows at the end boundary', async () => {
    api.listMaintenance.mockResolvedValue({ data: { items: [
      maintenance('Later upgrade', now + 4000, now + 5000),
      maintenance('Finished upgrade', now - 2000, now),
      maintenance('Earlier upgrade', now + 1000, now + 2000),
      maintenance('Current upgrade', now - 1000, now + 1000),
    ] } })
    const wrapper = await render(Maintenance)
    expect(rows(wrapper).map(row => row.split('upgrade')[0]?.trim())).toEqual(['Current', 'Earlier', 'Later', 'Finished'])
    expect(rows(wrapper)[0]).toContain('进行中')
    await selectTab(wrapper, 'in-progress')
    expect(rows(wrapper)).toHaveLength(1)
    expect(rows(wrapper)[0]).toContain('Current upgrade')
    i18n.global.locale.value = 'en'
    await flushPromises()
    expect(rows(wrapper)[0]).toContain('In progress')
    expect(wrapper.find('[role=tab][data-value=in-progress]').text()).toContain('In progress')
    await selectTab(wrapper, 'scheduled')
    expect(rows(wrapper)).toHaveLength(2)
    expect(rows(wrapper)[0]).toContain('Earlier upgrade')
    await selectTab(wrapper, 'completed')
    expect(rows(wrapper)).toHaveLength(1)
    expect(rows(wrapper)[0]).toContain('Finished upgrade')
  })

  it('filters incident status with tabs and searches announcement content', async () => {
    api.listIncidents.mockResolvedValue({ data: { items: [incident('Earlier incident', 'investigating', 1), incident('Resolved incident', 'resolved', 2, 'gateway restored'), incident('Current incident', 'monitoring', 3)] } })
    const wrapper = await render(Incidents)
    await selectTab(wrapper, 'resolved')
    expect(rows(wrapper)).toHaveLength(1)
    expect(rows(wrapper)[0]).toContain('Resolved incident')
    await selectTab(wrapper, 'all')
    await wrapper.find(`input[aria-label="${i18n.global.t('incidents.search-incidents')}"]`).setValue('gateway')
    expect(rows(wrapper)).toHaveLength(1)
    expect(rows(wrapper)[0]).toContain('Resolved incident')
  })

  it('shows the channel action only in its tab and combines delivery status and channel filters', async () => {
    const wrapper = await render(Notifications)
    expect(wrapper.find('header').text()).toContain(i18n.global.t('notifications.add-channel'))
    await selectTab(wrapper, 'deliveries')
    expect(wrapper.find('header').text()).not.toContain(i18n.global.t('notifications.add-channel'))
    await selectOption(i18n.global.t('notifications.filter-status'), 'failed')
    expect(rows(wrapper)).toHaveLength(2)
    await selectOption(i18n.global.t('notifications.filter-channel'), 'channel-2')
    expect(rows(wrapper)).toHaveLength(1)
    expect(rows(wrapper)[0]).toContain('Team chat')
    expect(rows(wrapper)[0]).toContain('Rate limit exceeded')
  })

  it('searches channels and delivery error text without changing the original data', async () => {
    const wrapper = await render(Notifications)
    const channelSearch = wrapper.find(`input[aria-label="${i18n.global.t('notifications.search-channels')}"]`)
    expect(channelSearch.exists()).toBe(true)
    await channelSearch.setValue('Pager')
    expect(rows(wrapper)).toHaveLength(1)
    expect(rows(wrapper)[0]).toContain('Pager')
    await channelSearch.setValue('')
    expect(rows(wrapper)).toHaveLength(2)
    await selectTab(wrapper, 'deliveries')
    await wrapper.find(`input[aria-label="${i18n.global.t('notifications.search-deliveries')}"]`).setValue('Connection refused')
    expect(rows(wrapper)).toHaveLength(1)
    expect(rows(wrapper)[0]).toContain('Pager')
  })

  it('opens complete delivery errors in a detail dialog from the compact row', async () => {
    const wrapper = await render(Notifications)
    await selectTab(wrapper, 'deliveries')
    const detail = wrapper.findAll('tbody button').find(button => button.attributes('aria-label') === i18n.global.t('notifications.view-failure-details'))
    expect(detail).toBeDefined()
    await detail!.trigger('click')
    await flushPromises()
    expect(document.querySelector('[role=dialog][data-state=open]')?.textContent).toContain('Rate limit exceeded')
  })
})
