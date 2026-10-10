// @vitest-environment happy-dom
import type { Page, PublicPage, Settings, User } from '../client/types.gen'
import { PiniaColada } from '@pinia/colada'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { currentUser } from '../composables/api'
import { i18n } from '../composables/i18n'
import PageEditor from './PageEditor.vue'

const api = vi.hoisted(() => ({ listMonitors: vi.fn(), getSettings: vi.fn(), createPages: vi.fn(), previewPage: vi.fn() }))
vi.mock('../client/sdk.gen', async importOriginal => ({ ...await importOriginal<typeof import('../client/sdk.gen')>(), ...api }))
vi.mock('vue-router', async importOriginal => ({ ...await importOriginal<typeof import('vue-router')>(), useRoute: () => ({ params: {} }), useRouter: () => ({ replace: vi.fn(), push: vi.fn() }) }))

const wrappers: ReturnType<typeof mount>[] = []
const admin: User = { id: 'admin', username: 'admin', name: 'Administrator', role: 'admin', enabled: true, locale: 'zh-CN', timezone: 'UTC', createdAt: 0, updatedAt: 0 }
const settings: Settings = { organizationName: 'Octopulse', timezone: 'UTC', locale: 'zh-CN', allowedDomains: [], retention: { roundDays: 30, attemptDays: 7, fiveMinuteDays: 90, historyMonths: 12 } }

beforeEach(() => {
  vi.resetAllMocks()
  currentUser.value = { ...admin }
  i18n.global.locale.value = 'zh-CN'
  api.listMonitors.mockResolvedValue({ data: { items: [] } })
  api.getSettings.mockResolvedValue({ data: settings })
})
afterEach(async () => {
  vi.restoreAllMocks()
  wrappers.splice(0).forEach(wrapper => wrapper.unmount())
  await flushPromises()
  document.body.innerHTML = ''
})

describe('status page editor views', () => {
  it('allows correcting validation errors and saves the draft with the existing request shape', async () => {
    const wrapper = mount(PageEditor, { attachTo: document.body, global: { plugins: [createPinia(), [PiniaColada, { queryOptions: { refetchOnWindowFocus: false } }], i18n], stubs: { RouterLink: { template: '<a><slot /></a>' } } } })
    wrappers.push(wrapper)
    await flushPromises()
    const save = wrapper.findAll('button').find(button => button.text() === i18n.global.t('common.saveDraft'))!
    await save.trigger('click')
    await flushPromises()
    await wrapper.find('input[name=name]').setValue('Service status')
    await wrapper.find('input[name=slug]').setValue('services')
    await wrapper.find('input[name="draft.title"]').setValue('Public service status')
    api.createPages.mockImplementationOnce(({ body }: { body: Page }) => Promise.resolve({ data: { ...body, id: 'new-page' } }))
    const preview: PublicPage = { id: 'new-page', slug: 'services', state: 'unknown', config: { title: 'Public service status', description: '', logoUrl: '', brandColor: '#2563eb', colorScheme: 'system', links: [], groups: [] }, groups: [], incidents: [], maintenance: [], updatedAt: 0 }
    api.previewPage.mockResolvedValueOnce({ data: preview })
    await save.trigger('click')
    await flushPromises()
    expect(api.createPages).toHaveBeenCalledOnce()
    expect(api.createPages).toHaveBeenCalledWith(expect.objectContaining({ body: expect.objectContaining({ name: 'Service status', slug: 'services', draft: expect.objectContaining({ title: 'Public service status' }) }) }))
    expect(wrapper.find('[role=alert]').exists()).toBe(false)
    expect(wrapper.find<HTMLInputElement>('input[name="draft.title"]').element.value).toBe('Public service status')
  })

  it('blocks editing after initial data fails and restores it after retry', async () => {
    api.getSettings.mockRejectedValueOnce(new Error('Settings unavailable'))
    const wrapper = mount(PageEditor, { attachTo: document.body, global: { plugins: [createPinia(), [PiniaColada, { queryOptions: { refetchOnWindowFocus: false } }], i18n], stubs: { RouterLink: { template: '<a><slot /></a>' } } } })
    wrappers.push(wrapper)
    await flushPromises()
    expect(wrapper.find('[role=alert]').text()).toContain('Settings unavailable')
    expect(wrapper.find('input').exists()).toBe(false)
    const retry = wrapper.findAll('button').find(button => button.text() === i18n.global.t('asyncState.retry'))
    expect(retry).toBeDefined()
    await retry!.trigger('click')
    await flushPromises()
    expect(wrapper.find('[role=alert]').exists()).toBe(false)
    expect(wrapper.find('input').exists()).toBe(true)
  })

  it('opens logo selection through a keyboard-accessible upload button', async () => {
    const wrapper = mount(PageEditor, { attachTo: document.body, global: { plugins: [createPinia(), [PiniaColada, { queryOptions: { refetchOnWindowFocus: false } }], i18n], stubs: { RouterLink: { template: '<a><slot /></a>' } } } })
    wrappers.push(wrapper)
    await flushPromises()
    const upload = wrapper.findAll('button').find(button => button.text() === i18n.global.t('pageEditor.upload'))
    expect(upload).toBeDefined()
    const chooseFile = vi.spyOn(wrapper.find<HTMLInputElement>('input[type=file]').element, 'click')
    await upload!.trigger('click')
    expect(chooseFile).toHaveBeenCalledOnce()
  })

  it('reveals the editor and focuses the first invalid field when saving from preview', async () => {
    const wrapper = mount(PageEditor, { attachTo: document.body, global: { plugins: [createPinia(), [PiniaColada, { queryOptions: { refetchOnWindowFocus: false } }], i18n], stubs: { RouterLink: { template: '<a><slot /></a>' } } } })
    wrappers.push(wrapper)
    await flushPromises()
    const label = wrapper.findAll('label').find(label => label.text().startsWith(i18n.global.t('pageEditor.internalPageName')))!
    const name = wrapper.find<HTMLInputElement>(`#${label.attributes('for')}`)
    await wrapper.findAll('button').find(button => button.text() === i18n.global.t('pageEditor.preview'))!.trigger('click')
    await wrapper.findAll('button').find(button => button.text() === i18n.global.t('common.saveDraft'))!.trigger('click')
    await flushPromises()
    expect(name.isVisible()).toBe(true)
    expect(document.activeElement).toBe(name.element)
    expect(name.attributes('aria-invalid')).toBe('true')
    const errorId = name.attributes('aria-errormessage') || name.attributes('aria-describedby')
    expect(errorId).toBeTruthy()
    expect(document.getElementById(errorId!)?.textContent).toContain(i18n.global.t('pageEditor.pageNameAndTitleAreRequired'))
    await name.setValue('Service status')
    await wrapper.findAll('button').find(button => button.text() === i18n.global.t('common.saveDraft'))!.trigger('click')
    await flushPromises()
    expect(document.activeElement === wrapper.find('input[name=slug]').element).toBe(true)
  })

  it('keeps unsaved inputs mounted while switching to the live preview and back', async () => {
    const wrapper = mount(PageEditor, { attachTo: document.body, global: { plugins: [createPinia(), [PiniaColada, { queryOptions: { refetchOnWindowFocus: false } }], i18n], stubs: { RouterLink: { template: '<a><slot /></a>' } } } })
    wrappers.push(wrapper)
    await flushPromises()
    const label = wrapper.findAll('label').find(label => label.text().startsWith(i18n.global.t('pageEditor.publicTitle')))!
    const title = wrapper.find<HTMLInputElement>(`#${label.attributes('for')}`)
    const inputElement = title.element
    await title.setValue('未保存的服务状态')
    const preview = wrapper.findAll('button').find(button => button.text() === i18n.global.t('pageEditor.preview'))
    expect(preview, 'the editor exposes a preview switch').toBeDefined()
    await preview!.trigger('click')
    expect(title.isVisible()).toBe(false)
    expect(wrapper.find('aside').text()).toContain('未保存的服务状态')
    const edit = wrapper.findAll('button').find(button => button.text() === i18n.global.t('common.edit'))!
    await edit.trigger('click')
    expect(title.isVisible()).toBe(true)
    expect(title.element).toBe(inputElement)
    expect(title.element.value).toBe('未保存的服务状态')
  })
})
