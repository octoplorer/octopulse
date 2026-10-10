// @vitest-environment happy-dom
import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'
import { h, nextTick } from 'vue'
import { Button } from '../button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '../dropdown-menu'
import { Switch } from '../switch'
import { Popover, PopoverContent, PopoverTrigger } from './index'

const wrappers: ReturnType<typeof mount>[] = []
async function settle() {
  await nextTick()
  await new Promise(resolve => setTimeout(resolve, 40))
}
afterEach(() => {
  wrappers.forEach(wrapper => wrapper.unmount())
  wrappers.length = 0
  document.body.innerHTML = ''
})

describe('ark focus transfer', () => {
  it('preserves an as-child button as the single trigger and restores focus after Escape', async () => {
    const wrapper = mount(Popover, { attachTo: document.body, slots: { default: () => [h(PopoverTrigger, { asChild: true }, () => h(Button, {}, () => 'Settings')), h(PopoverContent, { portalled: false }, () => h('input', { 'aria-label': 'Monitor name' }))] } })
    wrappers.push(wrapper)
    const trigger = wrapper.get('button').element as HTMLButtonElement
    expect(wrapper.findAll('button')).toHaveLength(1)
    trigger.focus()
    trigger.click()
    await settle()
    const input = wrapper.get('input').element as HTMLInputElement
    expect(document.activeElement).toBe(input)
    input.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
    await settle()
    expect(wrapper.emitted('update:open')?.at(-1)).toEqual([false])
    expect(document.activeElement).toBe(trigger)
  })

  it('keeps menu keyboard focus on the content and highlights only enabled actions', async () => {
    const wrapper = mount(DropdownMenu, { attachTo: document.body, slots: { default: () => [h(DropdownMenuTrigger, { asChild: true }, () => h(Button, {}, () => 'Actions')), h(DropdownMenuContent, { portalled: false }, () => [h(DropdownMenuItem, { value: 'disabled', disabled: true }, () => 'Disabled'), h(DropdownMenuItem, { value: 'edit' }, () => 'Edit'), h(DropdownMenuItem, { value: 'remove' }, () => 'Remove')])] } })
    wrappers.push(wrapper)
    const trigger = wrapper.get('button')
    ;(trigger.element as HTMLButtonElement).focus()
    await trigger.trigger('keydown', { key: 'ArrowDown' })
    await settle()
    const content = wrapper.get('[role="menu"]')
    expect(document.activeElement).toBe(content.element)
    const activeId = content.attributes('aria-activedescendant')
    expect(activeId).toBeTruthy()
    const active = document.getElementById(activeId ?? '')
    expect(active?.getAttribute('data-value')).toBe('edit')
    expect(active?.getAttribute('aria-disabled')).not.toBe('true')
    await content.trigger('keydown', { key: 'Escape' })
    await settle()
    expect(document.activeElement).toBe(trigger.element)
  })

  it('transfers keyboard focus state to the switch control and clears it when disabled', async () => {
    const wrapper = mount(Switch, { props: { label: 'Email notifications' }, attachTo: document.body })
    wrappers.push(wrapper)
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab', bubbles: true }))
    const input = wrapper.get('input').element as HTMLInputElement
    input.focus()
    await settle()
    expect(document.activeElement).toBe(input)
    expect(wrapper.get('[data-part="control"]').attributes('data-focus-visible')).toBeDefined()
    await wrapper.setProps({ disabled: true })
    await settle()
    expect(input.disabled).toBe(true)
    expect(wrapper.get('[data-part="control"]').attributes('data-focus-visible')).toBeUndefined()
  })
})
