// @vitest-environment happy-dom
import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'
import { h, nextTick } from 'vue'
import { Table, TableBody, TableCheckCell, TableCheckHead, TableHeader, TableResizeHandle, TableRow } from './index'

const wrappers: ReturnType<typeof mount>[] = []
afterEach(() => {
  wrappers.forEach(wrapper => wrapper.unmount())
  wrappers.length = 0
})

describe('table controls', () => {
  it('selects a row through its accessible checkbox', async () => {
    const wrapper = mount(Table, { slots: { default: () => h(TableBody, {}, () => h(TableRow, {}, () => h(TableCheckCell, { label: 'Select Production' }))) } })
    wrappers.push(wrapper)
    const cell = wrapper.findComponent(TableCheckCell)
    expect(cell.get('input').attributes('aria-label')).toBe('Select Production')
    ;(cell.get('input').element as HTMLInputElement).click()
    await nextTick()
    expect(cell.emitted('update:modelValue')).toEqual([[true]])
    expect(cell.emitted('checkedChange')).toEqual([[true]])
  })

  it('renders an indeterminate select-all control and prevents disabled selection', async () => {
    const wrapper = mount(Table, { slots: { default: () => h(TableHeader, {}, () => h(TableRow, {}, () => h(TableCheckHead, { indeterminate: true, disabled: true }))) } })
    wrappers.push(wrapper)
    const head = wrapper.findComponent(TableCheckHead)
    const input = head.get('input').element as HTMLInputElement
    expect(input.indeterminate).toBe(true)
    input.click()
    await nextTick()
    expect(head.emitted('update:modelValue')).toBeUndefined()
  })

  it('exposes resize adjustments to keyboard users', async () => {
    const wrapper = mount(TableResizeHandle, { props: { step: 20 } })
    wrappers.push(wrapper)
    await wrapper.get('button').trigger('keydown', { key: 'ArrowRight' })
    await wrapper.get('button').trigger('keydown', { key: 'ArrowLeft' })
    expect(wrapper.emitted('resize')).toEqual([[20], [-20]])
  })
})
