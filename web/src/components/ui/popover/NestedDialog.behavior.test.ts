// @vitest-environment happy-dom
import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'
import { h, nextTick, shallowRef } from 'vue'
import { Combobox } from '../combobox'
import { DatePicker } from '../date-picker'
import { Dialog } from '../dialog'
import { Popover, PopoverContent, PopoverTrigger } from './index'

const wrappers: ReturnType<typeof mount>[] = []
async function settle() {
  await nextTick()
  await new Promise(resolve => setTimeout(resolve, 40))
}
async function renderDialog(content: () => ReturnType<typeof h>) {
  const wrapper = mount(Dialog, { props: { open: true, title: 'Edit monitor', description: 'Update monitor settings' }, attachTo: document.body, slots: { default: content } })
  wrappers.push(wrapper)
  await settle()
  return wrapper
}
function element<T extends HTMLElement>(selector: string): T {
  const found = document.querySelector<T>(selector)
  expect(found).not.toBeNull()
  return found!
}
afterEach(async () => {
  wrappers.splice(0).forEach(wrapper => wrapper.unmount())
  await settle()
  document.body.innerHTML = ''
})

describe('interactive popups inside a modal Dialog', () => {
  it('keeps combobox options accessible and selects an option without dismissing the dialog', async () => {
    const wrapper = await renderDialog(() => h(Combobox, { label: 'Database', items: ['SQLite', 'PostgreSQL'] }))
    element<HTMLButtonElement>('[data-scope=combobox][data-part=trigger]').click()
    await settle()
    const option = element<HTMLElement>('[role=option][data-value="PostgreSQL"]')
    expect(option.closest('[aria-hidden=true], [inert]')).toBeNull()
    option.click()
    await settle()
    expect(wrapper.findComponent(Combobox).emitted('update:modelValue')?.at(-1)).toEqual(['PostgreSQL'])
    expect(element('[data-scope=dialog][data-part=content]').getAttribute('data-state')).toBe('open')
  })

  it('keeps calendar days accessible and commits the selected date inside the dialog', async () => {
    const wrapper = await renderDialog(() => h(DatePicker, { label: 'Start date', inline: false, modelValue: ['2026-10-09'] }))
    element<HTMLButtonElement>('[data-scope=date-picker][data-part=trigger]').click()
    await settle()
    const day = element<HTMLButtonElement>('[data-part=table-cell-trigger][data-value="2026-10-12"]')
    expect(day.closest('[aria-hidden=true], [inert]')).toBeNull()
    day.click()
    await settle()
    expect(wrapper.findComponent(DatePicker).emitted('update:modelValue')?.at(-1)).toEqual([['2026-10-12']])
    expect(element('[data-scope=dialog][data-part=content]').getAttribute('data-state')).toBe('open')
  })

  it('allows editing a portalled popover form and restores focus to its trigger after Escape', async () => {
    const name = shallowRef('Original name')
    const wrapper = await renderDialog(() => h(Popover, {}, { default: () => [
      h(PopoverTrigger, {}, () => 'Advanced settings'),
      h(PopoverContent, { portalled: true }, () => h('input', { 'aria-label': 'Monitor name', 'value': name.value, 'onInput': (event: Event) => { name.value = (event.target as HTMLInputElement).value } })),
    ] }))
    const trigger = element<HTMLButtonElement>('[data-scope=popover][data-part=trigger]')
    trigger.click()
    await settle()
    const input = element<HTMLInputElement>('input[aria-label="Monitor name"]')
    expect(input.closest('[aria-hidden=true], [inert]')).toBeNull()
    input.focus()
    expect(document.activeElement).toBe(input)
    input.value = 'Updated name'
    input.dispatchEvent(new Event('input', { bubbles: true }))
    expect(name.value).toBe('Updated name')
    input.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
    await settle()
    expect(wrapper.findComponent(Popover).emitted('update:open')?.at(-1)).toEqual([false])
    expect(document.activeElement).toBe(trigger)
    expect(element('[data-scope=dialog][data-part=content]').getAttribute('data-state')).toBe('open')
  })
})
