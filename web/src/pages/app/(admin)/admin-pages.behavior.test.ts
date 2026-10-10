// @vitest-environment happy-dom
import type { User } from '../../../client/types.gen'
import { PiniaColada } from '@pinia/colada'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'
import { currentUser } from '../../../composables/api'
import { i18n } from '../../../composables/i18n'
import { timezone } from '../../../composables/preferences'
import Maintenance from './maintenance.vue'
import Secrets from './secrets.vue'

const api = vi.hoisted(() => ({
  listMaintenance: vi.fn(),
  listMonitors: vi.fn(),
  listPages: vi.fn(),
  listSecrets: vi.fn(),
  createMaintenance: vi.fn(),
  deleteSecret: vi.fn(),
}))
vi.mock('../../../client/sdk.gen', async importOriginal => ({ ...await importOriginal<typeof import('../../../client/sdk.gen')>(), ...api }))
vi.mock('../../../composables/notices', () => ({ notify: vi.fn() }))

const wrappers: ReturnType<typeof mount>[] = []
const admin: User = { id: 'admin', username: 'admin', name: 'Administrator', role: 'admin', enabled: true, locale: 'zh-CN', timezone: 'UTC', createdAt: 0, updatedAt: 0 }
const secret = { id: 'secret-fixture', name: 'API token', createdAt: 0, updatedAt: 0 }

beforeEach(() => {
  vi.resetAllMocks()
  currentUser.value = { ...admin }
  timezone.value = 'UTC'
  i18n.global.locale.value = 'zh-CN'
  api.listMaintenance.mockResolvedValue({ data: { items: [] } })
  api.listMonitors.mockResolvedValue({ data: { items: [{ id: 'monitor-1', name: 'API endpoint' }, { id: 'monitor-2', name: 'Dashboard' }] } })
  api.listPages.mockResolvedValue({ data: { items: [{ id: 'page-1', name: 'Public status' }] } })
  api.listSecrets.mockResolvedValue({ data: { items: [secret] } })
  api.createMaintenance.mockResolvedValue({ data: {} })
})
afterEach(() => {
  wrappers.splice(0).forEach(wrapper => wrapper.unmount())
  document.body.innerHTML = ''
})
async function render(component: Parameters<typeof mount>[0]) {
  const wrapper = mount(component, { attachTo: document.body, global: { plugins: [createPinia(), [PiniaColada, { queryOptions: { refetchOnWindowFocus: false } }], i18n], stubs: { RouterLink: { template: '<a><slot /></a>' } } } })
  wrappers.push(wrapper)
  await flushPromises()
  return wrapper
}
function button(label: string, container: ParentNode = document) {
  const result = Array.from(container.querySelectorAll<HTMLButtonElement>('button')).find(candidate => candidate.textContent?.trim() === label || candidate.getAttribute('aria-label') === label)
  if (!result)
    throw new Error(`Button not found: ${label}`)
  return result
}
async function click(label: string, container: ParentNode = document) {
  button(label, container).click()
  await flushPromises()
}
async function input(selector: string, value: string) {
  const element = document.querySelector<HTMLInputElement | HTMLTextAreaElement>(selector)
  if (!element)
    throw new Error(`Input not found: ${selector}`)
  element.value = value
  element.dispatchEvent(new Event('input', { bubbles: true }))
  await nextTick()
}
const t = i18n.global.t

describe('migrated administration pages', () => {
  it('clears a cancelled secret draft and prevents duplicate delete requests while allowing retry after failure', async () => {
    await render(Secrets)
    await click(t('common.add-secret'))
    await input('#secret-form textarea[name=value]', 'unsaved-value')
    await click(t('common.cancel'), document.querySelector('[role=dialog][data-state=open]')!)
    await click(t('common.add-secret'))
    expect(document.querySelector<HTMLTextAreaElement>('#secret-form textarea')?.value).toBe('')
    await click(t('common.cancel'), document.querySelector('[role=dialog][data-state=open]')!)
    await click(t('common.delete'))
    let rejectRequest!: (reason: Error) => void
    api.deleteSecret.mockImplementationOnce(() => new Promise((_resolve, reject) => rejectRequest = reject))
    const remove = button(t('common.delete'), document.querySelector('[role=dialog][data-state=open]')!)
    remove.click()
    remove.click()
    await flushPromises()
    expect(api.deleteSecret).toHaveBeenCalledTimes(1)
    expect(remove.disabled).toBe(true)
    rejectRequest(new Error('Still referenced'))
    await flushPromises()
    expect(document.querySelector('[role=dialog][data-state=open]')).not.toBeNull()
    api.deleteSecret.mockResolvedValueOnce({ data: {} })
    await click(t('common.delete'), document.querySelector('[role=dialog][data-state=open]')!)
    expect(api.deleteSecret).toHaveBeenCalledTimes(2)
    expect(document.querySelector('[role=dialog][data-state=open]')).toBeNull()
  })

  it('submits maintenance scope from checkbox groups and converts the preserved wall-clock inputs in the selected timezone', async () => {
    await render(Maintenance)
    await click(t('common.schedule-maintenance'))
    const form = document.querySelector<HTMLFormElement>('#maintenance-form')!
    await input('#maintenance-form input:not([type])', 'Database upgrade')
    const dates = form.querySelectorAll<HTMLInputElement>('input[type=datetime-local]')
    await input(`#${dates[0]!.id}`, '2026-10-12T10:00')
    await input(`#${dates[1]!.id}`, '2026-10-12T11:30')
    const textInputs = Array.from(form.querySelectorAll<HTMLInputElement>('input')).filter(element => element.type === 'text')
    await input(`#${textInputs.at(-1)!.id}`, 'Asia/Shanghai')
    form.querySelector<HTMLInputElement>('input[type=checkbox][value=monitor-1]')!.click()
    form.querySelector<HTMLInputElement>('input[type=checkbox][value=page-1]')!.click()
    await nextTick()
    await click(t('maintenance.save-schedule'))
    expect(api.createMaintenance).toHaveBeenCalledWith(expect.objectContaining({ body: expect.objectContaining({ name: 'Database upgrade', monitorIds: ['monitor-1'], pageIds: ['page-1'], timezone: 'Asia/Shanghai', startsAt: Date.UTC(2026, 9, 12, 2), endsAt: Date.UTC(2026, 9, 12, 3, 30) }) }))
    expect(document.querySelector('[role=dialog][data-state=open]')).toBeNull()
  })
})
