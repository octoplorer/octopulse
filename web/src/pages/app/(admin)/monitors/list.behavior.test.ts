// @vitest-environment happy-dom
import type { Monitor } from '../../../../client/types.gen'
import { PiniaColada } from '@pinia/colada'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { currentUser } from '../../../../composables/api'
import { i18n } from '../../../../composables/i18n'
import { newMonitor } from '../../../../lib/monitor'
import Monitors from './index.vue'

const api = vi.hoisted(() => ({ listMonitors: vi.fn() }))
vi.mock('../../../../client/sdk.gen', async importOriginal => ({
  ...await importOriginal<typeof import('../../../../client/sdk.gen')>(),
  ...api,
}))

const wrappers: ReturnType<typeof mount>[] = []

beforeEach(() => {
  vi.resetAllMocks()
  currentUser.value = null
  i18n.global.locale.value = 'zh-CN'
  const monitors: Monitor[] = [
    { ...newMonitor(), id: 'api', name: 'API endpoint', state: 'down', http: { ...newMonitor().http!, url: 'https://api.example.com' } },
    { ...newMonitor(), id: 'dashboard', name: 'Dashboard', state: 'up' },
  ]
  api.listMonitors.mockResolvedValue({ data: { items: monitors } })
})

afterEach(() => {
  wrappers.splice(0).forEach(wrapper => wrapper.unmount())
  document.body.innerHTML = ''
})

describe('monitor list filters', () => {
  it('clears search, status and type together to restore every monitor', async () => {
    const wrapper = mount(Monitors, {
      attachTo: document.body,
      global: {
        plugins: [createPinia(), [PiniaColada, { queryOptions: { refetchOnWindowFocus: false } }], i18n],
        stubs: { RouterLink: { template: '<a><slot /></a>' } },
      },
    })
    wrappers.push(wrapper)
    await flushPromises()
    await wrapper.get('input').setValue('API')
    await wrapper.findAll('select')[0]!.setValue('down')
    await wrapper.findAll('select')[1]!.setValue('tcp')
    await flushPromises()
    expect(wrapper.findAll('tbody tr')).toHaveLength(0)

    const clear = wrapper.findAll('button').find(button => button.text() === i18n.global.t('common.clear-filters'))
    expect(clear, 'the empty result exposes an action that clears every filter').toBeDefined()
    await clear!.trigger('click')
    await flushPromises()

    expect(wrapper.get('input').element.value).toBe('')
    expect(wrapper.findAll('select').map(select => select.element.value)).toEqual(['all', 'all'])
    expect(wrapper.findAll('tbody tr').map(row => row.text())).toEqual([
      expect.stringContaining('API endpoint'),
      expect.stringContaining('Dashboard'),
    ])
  })
})
