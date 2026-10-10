// @vitest-environment happy-dom
import type { Audit } from '../../../client/types.gen'
import { PiniaColada } from '@pinia/colada'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { expect, it, vi } from 'vitest'
import { i18n } from '../../../composables/i18n'
import AuditPage from './audit.vue'

const api = vi.hoisted(() => ({ listAudit: vi.fn() }))
vi.mock('../../../client/sdk.gen', async importOriginal => ({ ...await importOriginal<typeof import('../../../client/sdk.gen')>(), ...api }))

it('restores audit records from an unmatched search through the empty state', async () => {
  const entry: Audit = { id: 'record', userId: 'admin', username: 'Administrator', action: 'create', resourceType: 'monitor', resourceId: 'api-monitor', createdAt: 1 }
  api.listAudit.mockResolvedValue({ data: { items: [entry] } })
  i18n.global.locale.value = 'zh-CN'
  const wrapper = mount(AuditPage, { global: { plugins: [createPinia(), [PiniaColada, { queryOptions: { refetchOnWindowFocus: false } }], i18n] } })
  try {
    await flushPromises()
    await wrapper.get('input').setValue('missing resource')
    expect(wrapper.findAll('tbody tr')).toHaveLength(0)
    const clear = wrapper.findAll('button').find(button => button.text() === i18n.global.t('common.clearFilters'))
    expect(clear).toBeDefined()
    await clear!.trigger('click')
    await flushPromises()
    expect(wrapper.get('input').element.value).toBe('')
    expect(wrapper.findAll('tbody tr')).toHaveLength(1)
    expect(wrapper.find('tbody').text()).toContain('Administrator')
  }
  finally {
    wrapper.unmount()
  }
})
