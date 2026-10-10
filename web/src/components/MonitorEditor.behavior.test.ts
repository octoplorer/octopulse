// @vitest-environment happy-dom
import type { User } from '../client/types.gen'
import { PiniaColada } from '@pinia/colada'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { currentUser } from '../composables/api'
import { i18n } from '../composables/i18n'
import MonitorEditor from './MonitorEditor.vue'

const api = vi.hoisted(() => ({ listChannels: vi.fn(), listSecrets: vi.fn() }))
vi.mock('../client/sdk.gen', async importOriginal => ({ ...await importOriginal<typeof import('../client/sdk.gen')>(), ...api }))
vi.mock('vue-router', async importOriginal => ({ ...await importOriginal<typeof import('vue-router')>(), useRoute: () => ({ params: {} }), useRouter: () => ({ push: vi.fn() }) }))

const wrappers: ReturnType<typeof mount>[] = []
const admin: User = { id: 'admin', username: 'admin', name: 'Administrator', role: 'admin', enabled: true, locale: 'zh-CN', timezone: 'UTC', createdAt: 0, updatedAt: 0 }

beforeEach(() => {
  vi.resetAllMocks()
  currentUser.value = { ...admin }
  i18n.global.locale.value = 'zh-CN'
  api.listChannels.mockResolvedValue({ data: { items: [] } })
  api.listSecrets.mockResolvedValue({ data: { items: [] } })
})
afterEach(async () => {
  wrappers.splice(0).forEach(wrapper => wrapper.unmount())
  await flushPromises()
  document.body.innerHTML = ''
})
async function render() {
  const wrapper = mount(MonitorEditor, { attachTo: document.body, global: { plugins: [createPinia(), [PiniaColada, { queryOptions: { refetchOnWindowFocus: false } }], i18n], stubs: { RouterLink: { template: '<a><slot /></a>' } } } })
  wrappers.push(wrapper)
  await flushPromises()
  return wrapper
}

describe('monitor editor recovery', () => {
  it('reveals a hidden target field when native validation rejects it', async () => {
    const wrapper = await render()
    const url = wrapper.find<HTMLInputElement>('input[type=url]')
    await wrapper.findAll('[role=tab]').find(tab => tab.text() === i18n.global.t('monitorEditor.notifications'))!.trigger('click')
    await flushPromises()
    expect(url.element.closest('[role=tabpanel]')?.hasAttribute('hidden')).toBe(true)
    await url.trigger('invalid')
    await flushPromises()
    expect(url.element.closest('[role=tabpanel]')?.hasAttribute('hidden')).toBe(false)
    expect(document.activeElement === url.element).toBe(true)
  })

  it('allows the native validation message after revealing a rejected field and on later attempts', async () => {
    const wrapper = await render()
    const url = wrapper.find<HTMLInputElement>('input[type=url]')
    await url.setValue('')
    const invalidEvents: { prevented: boolean, hidden: boolean }[] = []
    url.element.addEventListener('invalid', (event) => {
      invalidEvents.push({ prevented: event.defaultPrevented, hidden: !!url.element.closest('[role=tabpanel]')?.hasAttribute('hidden') })
    })
    await wrapper.findAll('[role=tab]').find(tab => tab.text() === i18n.global.t('monitorEditor.notifications'))!.trigger('click')
    await flushPromises()
    await url.trigger('invalid')
    await flushPromises()
    expect(invalidEvents).toEqual([{ prevented: true, hidden: true }, { prevented: false, hidden: false }])
    await url.trigger('invalid')
    await flushPromises()
    expect(invalidEvents.slice(2)).toEqual([{ prevented: true, hidden: false }, { prevented: false, hidden: false }])
  })

  it('blocks editing until dependencies load successfully and offers retry', async () => {
    api.listSecrets.mockRejectedValueOnce(new Error('Secrets unavailable'))
    const wrapper = await render()
    expect(wrapper.find('[role=alert]').text()).toContain('Secrets unavailable')
    expect(wrapper.find('form').exists()).toBe(false)
    const retry = wrapper.findAll('button').find(button => button.text() === i18n.global.t('asyncState.retry'))
    expect(retry).toBeDefined()
    await retry!.trigger('click')
    await flushPromises()
    expect(wrapper.find('[role=alert]').exists()).toBe(false)
    expect(wrapper.find('form').exists()).toBe(true)
  })

  it('returns to target assertions and focuses invalid JSON after saving from another tab', async () => {
    const wrapper = await render()
    const label = wrapper.findAll('label').find(label => label.text() === i18n.global.t('monitorEditor.responseHeaderAssertions'))!
    const headers = wrapper.find<HTMLTextAreaElement>(`#${label.attributes('for')}`)
    await headers.setValue('{')
    await wrapper.findAll('[role=tab]').find(tab => tab.text() === i18n.global.t('monitorEditor.notifications'))!.trigger('click')
    await flushPromises()
    expect(headers.element.closest('[role=tabpanel]')?.hasAttribute('hidden')).toBe(true)
    await wrapper.find('form').trigger('submit')
    await flushPromises()
    expect(headers.element.closest('[role=tabpanel]')?.hasAttribute('hidden')).toBe(false)
    expect(document.activeElement).toBe(headers.element)
    expect(headers.attributes('aria-invalid')).toBe('true')
  })
})
