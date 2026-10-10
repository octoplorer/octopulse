// @vitest-environment happy-dom
import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'
import { h, nextTick } from 'vue'
import { Field } from '../field'
import { Switch, SwitchGroup } from './index'

const wrappers: ReturnType<typeof mount>[] = []
function render(props: Record<string, unknown>) {
  const wrapper = mount(Switch, { props, attachTo: document.body })
  wrappers.push(wrapper)
  return wrapper
}
afterEach(() => {
  wrappers.forEach(wrapper => wrapper.unmount())
  wrappers.length = 0
  document.body.innerHTML = ''
})

describe('switch form and accessibility behavior', () => {
  it('associates the switch label and its description with the form control', () => {
    const wrapper = render({ label: 'Email notifications', description: 'Send incident reports' })
    const input = wrapper.get('input')
    const descriptionId = input.attributes('aria-describedby')
    expect(descriptionId).toBeTruthy()
    expect(wrapper.get(`#${descriptionId}`).text()).toBe('Send incident reports')
    expect(input.attributes('aria-labelledby')).toBe(wrapper.get('[data-part="label"]').attributes('id'))
  })

  it('cannot toggle a read-only or disabled switch', async () => {
    const wrapper = render({ label: 'Email notifications', readOnly: true })
    ;(wrapper.get('input').element as HTMLInputElement).click()
    await nextTick()
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
    await wrapper.setProps({ readOnly: false, disabled: true })
    ;(wrapper.get('input').element as HTMLInputElement).click()
    await nextTick()
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
  })

  it('submits its named value to an external form only while checked', async () => {
    const form = document.createElement('form')
    form.id = 'settings'
    document.body.append(form)
    const wrapper = render({ label: 'Email notifications', name: 'notification', value: 'email', form: 'settings' })
    expect(new FormData(form).get('notification')).toBeNull()
    ;(wrapper.get('input').element as HTMLInputElement).click()
    await nextTick()
    expect(new FormData(form).get('notification')).toBe('email')
    expect(wrapper.emitted('update:modelValue')).toEqual([[true]])
  })

  it('inherits group disabling and preserves a meaningful fieldset legend', async () => {
    const wrapper = mount(SwitchGroup, { props: { legend: 'Notifications', disabled: true }, attachTo: document.body, slots: { default: () => h(Switch, { label: 'Email' }) } })
    wrappers.push(wrapper)
    expect(wrapper.get('legend').text()).toBe('Notifications')
    expect((wrapper.get('input').element as HTMLInputElement).disabled).toBe(true)
    await wrapper.setProps({ disabled: false })
    await nextTick()
    ;(wrapper.get('input').element as HTMLInputElement).click()
    await nextTick()
    expect(wrapper.findComponent(Switch).emitted('update:modelValue')).toEqual([[true]])
  })

  it('inherits a disabled Field until the Field is enabled', async () => {
    const wrapper = mount(Field, { props: { disabled: true }, attachTo: document.body, slots: { default: () => h(Switch, { label: 'Email' }) } })
    wrappers.push(wrapper)
    const input = wrapper.get('input').element as HTMLInputElement
    expect(input.disabled).toBe(true)
    input.click()
    await nextTick()
    expect(wrapper.findComponent(Switch).emitted('update:modelValue')).toBeUndefined()
    await wrapper.setProps({ disabled: false })
    input.click()
    await nextTick()
    expect(wrapper.findComponent(Switch).emitted('update:modelValue')).toEqual([[true]])
  })

  it('inherits a read-only Field until the Field becomes editable', async () => {
    const wrapper = mount(Field, { props: { readOnly: true }, attachTo: document.body, slots: { default: () => h(Switch, { label: 'Email' }) } })
    wrappers.push(wrapper)
    const input = wrapper.get('input').element as HTMLInputElement
    input.click()
    await nextTick()
    expect(wrapper.findComponent(Switch).emitted('update:modelValue')).toBeUndefined()
    await wrapper.setProps({ readOnly: false })
    input.click()
    await nextTick()
    expect(wrapper.findComponent(Switch).emitted('update:modelValue')).toEqual([[true]])
  })

  it('exposes Field validation on its native switch control', async () => {
    const wrapper = mount(Field, { props: { error: 'Consent is required' }, attachTo: document.body, slots: { default: () => h(Switch, { label: 'Consent' }) } })
    wrappers.push(wrapper)
    expect(wrapper.get('input').attributes('aria-invalid')).toBe('true')
    await wrapper.setProps({ error: undefined })
    expect(wrapper.get('input').attributes('aria-invalid')).toBe('false')
  })
})
