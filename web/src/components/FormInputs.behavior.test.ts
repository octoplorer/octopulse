import type { NameValueForm } from '../lib/monitor-form'
// @vitest-environment happy-dom
import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'
import { defineComponent, h, nextTick, ref } from 'vue'
import { useJSONInput, useListInput } from '../composables/form-inputs'
import { i18n } from '../composables/i18n'
import { emptyHTTP } from '../lib/monitor'
import ConnectionFields from './ConnectionFields.vue'
import KeyValues from './KeyValues.vue'
import TLSFields from './TLSFields.vue'
import { Input, InputArea } from './ui/input'

const wrappers: ReturnType<typeof mount>[] = []
const secrets = [{ id: 'token', name: 'API token', createdAt: 0, updatedAt: 0 }]
afterEach(() => {
  wrappers.splice(0).forEach(wrapper => wrapper.unmount())
  document.body.innerHTML = ''
})
function render(component: Parameters<typeof mount>[0], props: Record<string, unknown>) {
  const wrapper = mount(component, { props, attachTo: document.body, global: { plugins: [i18n] } })
  wrappers.push(wrapper)
  return wrapper
}

describe('form input adapters', () => {
  it('parses list values while preserving unfinished typed text until an external reset', async () => {
    const values = ref([200])
    const wrapper = render(defineComponent({
      setup() {
        const text = useListInput(values, next => values.value = next, { parseItem: Number })
        return () => h(Input, { 'modelValue': text.value, 'onUpdate:modelValue': next => text.value = String(next ?? '') })
      },
    }), {})
    await wrapper.find('input').setValue('200, 404, ')
    expect(values.value).toEqual([200, 404])
    expect(wrapper.find('input').element.value).toBe('200, 404, ')
    values.value = [500]
    await nextTick()
    expect(wrapper.find('input').element.value).toBe('500')
  })

  it('preserves line whitespace unless trimming was requested', async () => {
    const trim = ref(false)
    const values = ref<string[]>([])
    const wrapper = render(defineComponent({
      setup() {
        const text = useListInput(values, next => values.value = next, { parseItem: String, separator: 'lines', trim })
        return () => h(InputArea, { 'modelValue': text.value, 'onUpdate:modelValue': next => text.value = String(next ?? '') })
      },
    }), {})
    await wrapper.find('textarea').setValue(' first \n\n second ')
    expect(values.value).toEqual([' first ', ' second '])
    trim.value = true
    await nextTick()
    await wrapper.find('textarea').setValue(' first \n\n third ')
    expect(values.value).toEqual(['first', 'third'])
  })

  it('keeps invalid JSON visible and leaves the previous model untouched until valid JSON is entered', async () => {
    const values = ref({ attempts: 2 })
    const wrapper = render(defineComponent({
      setup() {
        const { text, error } = useJSONInput(values, next => values.value = next, 'Headers')
        return () => h('div', [
          h(InputArea, { 'modelValue': text.value, 'onUpdate:modelValue': next => text.value = String(next ?? '') }),
          error.value ? h('p', { role: 'alert' }, error.value) : null,
        ])
      },
    }), {})
    await wrapper.find('textarea').setValue('{"attempts":')
    expect(values.value).toEqual({ attempts: 2 })
    expect(wrapper.find('textarea').element.value).toBe('{"attempts":')
    expect(wrapper.find('[role=alert]').text()).toContain('Headers')
    await wrapper.find('textarea').setValue('{"attempts":3}')
    expect(values.value).toEqual({ attempts: 3 })
    expect(wrapper.find('textarea').element.value).toBe('{"attempts":3}')
    expect(wrapper.find('[role=alert]').exists()).toBe(false)
    values.value = { attempts: 4 }
    await nextTick()
    expect(wrapper.find('textarea').element.value).toBe('{\n  "attempts": 4\n}')
  })

  it('selects a secret and clears an optional reference without changing the stored literal value', async () => {
    const values = ref<NameValueForm[]>([{ name: 'Authorization', value: 'draft', secretRef: '' }])
    const wrapper = render(defineComponent({
      setup: () => () => h(KeyValues, { 'modelValue': values.value, secrets, 'onUpdate:modelValue': next => values.value = next }),
    }), {})
    await wrapper.find('[data-part=trigger]').trigger('click')
    await flushPromises()
    document.querySelector<HTMLElement>('[data-part=item][data-value=token]')!.click()
    await nextTick()
    expect(values.value[0]).toEqual({ name: 'Authorization', value: 'draft', secretRef: 'token' })
    expect(wrapper.findAll('input').filter(input => input.attributes('aria-label') === i18n.global.t('common.value'))).toHaveLength(0)
    await wrapper.find('[data-part=trigger]').trigger('click')
    await flushPromises()
    document.querySelector<HTMLElement>('[data-part=item][data-value=""]')!.click()
    await nextTick()
    expect(values.value[0]).toEqual({ name: 'Authorization', value: 'draft', secretRef: '' })
    expect(wrapper.find<HTMLInputElement>(`input[aria-label="${i18n.global.t('common.value')}"]`).element.value).toBe('draft')
  })

  it('associates the proxy secret trigger with its field label', async () => {
    const wrapper = render(ConnectionFields, { secrets, modelValue: emptyHTTP().connection })
    await nextTick()
    const label = wrapper.findAll('label').find(label => label.text() === i18n.global.t('connectionFields.proxyPasswordSecret'))!
    const trigger = wrapper.find('[data-part=trigger]')
    expect(wrapper.find('select').attributes('id')).toBe(label.attributes('for'))
    expect(trigger.attributes('aria-labelledby')?.split(' ')).toContain(label.attributes('id'))
  })

  it('names key-value controls independently of their placeholders', () => {
    const wrapper = render(KeyValues, { secrets, nameLabel: 'Header name', modelValue: [{ name: '', value: '' }] })
    expect(wrapper.find('input').attributes('aria-label')).toBe('Header name')
    expect(wrapper.find('[data-part=trigger]').attributes('aria-label')).toBe(i18n.global.t('keyValues.secretReference'))
  })

  it('edits TLS switches and server names while preserving the other TLS settings', async () => {
    const values = ref({ ...emptyHTTP().tls, caSecretRef: 'token', maxVersion: '1.3' })
    const wrapper = render(defineComponent({
      setup: () => () => h(TLSFields, { 'modelValue': values.value, secrets, 'allowToggle': true, 'onUpdate:modelValue': next => values.value = { ...next, caSecretRef: next.caSecretRef ?? '' } }),
    }), {})
    wrapper.findAll<HTMLInputElement>('input[role=switch]')[0]!.element.click()
    await nextTick()
    await wrapper.find('input[placeholder]').setValue('service.example.com')
    expect(values.value).toMatchObject({ enabled: true, serverName: 'service.example.com', caSecretRef: 'token', maxVersion: '1.3' })
    wrapper.findAll<HTMLInputElement>('input[role=switch]')[1]!.element.click()
    await nextTick()
    expect(values.value.insecureSkipVerify).toBe(true)
  })

  it('edits proxy settings without dropping its existing secret reference', async () => {
    const values = ref({ ...emptyHTTP().connection, proxyPasswordSecretRef: 'token' })
    const wrapper = render(defineComponent({
      setup: () => () => h(ConnectionFields, { 'modelValue': values.value, secrets, 'onUpdate:modelValue': next => values.value = { ...next, proxyPasswordSecretRef: next.proxyPasswordSecretRef ?? '' } }),
    }), {})
    await wrapper.find('input[placeholder="socks5://127.0.0.1:1080"]').setValue('socks5://proxy.example.com:1080')
    expect(values.value).toMatchObject({ proxyUrl: 'socks5://proxy.example.com:1080', proxyPasswordSecretRef: 'token', dnsServer: '', fixedIp: '' })
  })
})
