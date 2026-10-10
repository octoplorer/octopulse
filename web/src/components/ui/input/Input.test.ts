// @vitest-environment happy-dom
import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import { h, nextTick } from 'vue'
import { Field } from '../field'
import { Input, InputArea } from './index'

describe('form controls', () => {
  it('associates the label, hint and validation message with the input', async () => {
    const wrapper = mount(Field, {
      attachTo: document.body,
      props: { label: 'Endpoint', description: 'Use HTTPS', error: 'A target is required' },
      slots: { default: () => h(Input, { modelValue: '' }) },
    })
    await nextTick()
    const input = wrapper.get('input')
    expect(wrapper.get('label').attributes('for')).toBe(input.attributes('id'))
    expect(input.attributes('aria-invalid')).toBe('true')
    const describedBy = input.attributes('aria-describedby')?.split(' ') ?? []
    expect(describedBy.map(id => wrapper.get(`[id="${id}"]`).text())).toEqual(['Use HTTPS'])
    expect(wrapper.get(`[id="${input.attributes('aria-errormessage')}"]`).text()).toBe('A target is required')
    wrapper.unmount()
  })

  it('emits text updates and forwards native form constraints to the input', async () => {
    const wrapper = mount(Input, { props: { modelValue: '', size: 'sm' }, attrs: { name: 'url', required: true, maxlength: 100 } })
    await wrapper.get('input').setValue('https://example.com')
    expect(wrapper.emitted('update:modelValue')).toEqual([['https://example.com']])
    expect(wrapper.get('input').attributes('name')).toBe('url')
    expect(wrapper.get('input').attributes('required')).toBeDefined()
    expect(wrapper.get('input').attributes('size')).toBeUndefined()
    wrapper.unmount()
  })

  it('resizes a textarea within its row limit and restores native sizing when disabled', async () => {
    const scrollHeight = vi.spyOn(HTMLTextAreaElement.prototype, 'scrollHeight', 'get').mockReturnValue(200)
    const wrapper = mount(InputArea, { attachTo: document.body, props: { modelValue: 'one\ntwo', autoResize: true, maxRows: 4 }, attrs: { style: { lineHeight: '20px', padding: '0px', border: '0px' } } })
    await nextTick()
    const textarea = wrapper.get('textarea').element as HTMLTextAreaElement
    expect(textarea.style.height).toBe('80px')
    expect(textarea.style.overflowY).toBe('auto')
    await wrapper.setProps({ autoResize: false })
    expect(textarea.style.height).toBe('')
    wrapper.unmount()
    scrollHeight.mockRestore()
  })

  it('keeps textarea updates and readonly state on the native control', async () => {
    const wrapper = mount(InputArea, { props: { modelValue: 'line one' }, attrs: { readonly: true, rows: 4 } })
    expect(wrapper.get('textarea').attributes('readonly')).toBeDefined()
    await wrapper.get('textarea').setValue('line one\nline two')
    expect(wrapper.emitted('update:modelValue')).toEqual([['line one\nline two']])
    wrapper.unmount()
  })
})
