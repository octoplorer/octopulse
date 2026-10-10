import userEvent from '@testing-library/user-event'
// @vitest-environment happy-dom
import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'
import { nextTick } from 'vue'
import { CommandPalette, CommandPalettePanel } from './index'

const items = [
  { value: 'disabled', label: 'Disabled monitor', disabled: true },
  { value: 'create', label: 'Create monitor' },
]
let cleanup: (() => void) | undefined
async function settle() {
  await nextTick()
  await new Promise(resolve => setTimeout(resolve, 30))
}
afterEach(async () => {
  cleanup?.()
  await settle()
  document.body.innerHTML = ''
})

describe('command palette lifecycle', () => {
  it('keeps commands inside the modal available to the pointer', async () => {
    const user = userEvent.setup()
    const wrapper = mount(CommandPalette, { props: { open: false, items }, attachTo: document.body })
    cleanup = () => wrapper.unmount()
    await wrapper.setProps({ open: true })
    await settle()
    const option = document.querySelector<HTMLElement>('[role="option"][data-value="create"]')!
    expect(getComputedStyle(option.closest('[role="listbox"]')!).pointerEvents).not.toBe('none')
    expect(getComputedStyle(option).pointerEvents).not.toBe('none')
    await user.pointer([{ target: option, keys: '[MouseLeft>]' }, { keys: '[/MouseLeft]' }])
    await settle()
    expect(wrapper.emitted('select')).toEqual([[items[1], { newTab: false }]])
    expect(wrapper.emitted('update:open')).toEqual([[false]])
  })

  it('honors Ctrl pointer activation while ignoring disabled commands', async () => {
    const user = userEvent.setup()
    const wrapper = mount(CommandPalette, { props: { open: false, items, closeOnSelect: false }, attachTo: document.body })
    cleanup = () => wrapper.unmount()
    await wrapper.setProps({ open: true })
    await settle()
    await user.pointer([{ target: document.querySelector<HTMLElement>('[data-value=disabled]')!, keys: '[MouseLeft>]' }, { keys: '[/MouseLeft]' }])
    expect(wrapper.emitted('select')).toBeUndefined()
    await user.keyboard('[ControlLeft>]')
    await user.pointer([{ target: document.querySelector<HTMLElement>('[data-value=create]')!, keys: '[MouseLeft>]' }, { keys: '[/MouseLeft]' }])
    await user.keyboard('[/ControlLeft]')
    expect(wrapper.emitted('select')).toEqual([[items[1], { newTab: true }]])
    expect(wrapper.emitted('update:open')).toBeUndefined()
  })

  it('closes from Escape without the permanently open list swallowing dismissal', async () => {
    const user = userEvent.setup()
    const wrapper = mount(CommandPalette, { props: { open: false, items }, attachTo: document.body })
    cleanup = () => wrapper.unmount()
    await wrapper.setProps({ open: true })
    await settle()
    await user.keyboard('{Escape}')
    await settle()
    expect(wrapper.emitted('update:open')).toEqual([[false]])
    expect(document.querySelector('[role=dialog]')).toBeNull()
  })

  it('allows repeating an action while the palette remains open', async () => {
    const user = userEvent.setup()
    const wrapper = mount(CommandPalette, { props: { open: false, items, closeOnSelect: false }, attachTo: document.body })
    cleanup = () => wrapper.unmount()
    await wrapper.setProps({ open: true })
    await settle()
    const option = document.querySelector<HTMLElement>('[data-value=create]')!
    await user.click(option)
    await user.click(option)
    expect(wrapper.emitted('select')).toEqual([[items[1], { newTab: false }], [items[1], { newTab: false }]])
  })

  it('focuses its own input when opened and closes after selecting a command', async () => {
    const wrapper = mount(CommandPalette, { props: { open: true, items }, attachTo: document.body })
    cleanup = () => wrapper.unmount()
    await settle()
    const input = document.querySelector<HTMLInputElement>('[role="combobox"]')!
    expect(document.activeElement).toBe(input)
    input.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowDown', bubbles: true }))
    await settle()
    input.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true }))
    await settle()
    expect(wrapper.emitted('select')?.[0]?.[0]).toEqual(items[1])
    expect(wrapper.emitted('update:open')).toEqual([[false]])
  })

  it('preserves new-tab intent for Ctrl Enter and skips disabled items', async () => {
    const wrapper = mount(CommandPalettePanel, { props: { items }, attachTo: document.body })
    cleanup = () => wrapper.unmount()
    const input = wrapper.get('[role="combobox"]')
    await input.trigger('keydown', { key: 'ArrowDown' })
    await input.trigger('keydown', { key: 'Enter', ctrlKey: true })
    await settle()
    expect(wrapper.emitted('select')).toEqual([[items[1], { newTab: true }]])
  })

  it('keeps controlled search text while updating filtered results', async () => {
    const wrapper = mount(CommandPalettePanel, { props: { items, query: 'missing' }, attachTo: document.body })
    cleanup = () => wrapper.unmount()
    expect(wrapper.findAll('[role="option"]')).toHaveLength(0)
    await wrapper.setProps({ query: 'create' })
    expect(wrapper.findAll('[role="option"]')).toHaveLength(1)
    expect((wrapper.get('[role="combobox"]').element as HTMLInputElement).value).toBe('create')
  })

  it('automatically highlights the first enabled result after the query changes', async () => {
    const options = [...items, { value: 'settings', label: 'Settings' }]
    const wrapper = mount(CommandPalettePanel, { props: { items: options }, attachTo: document.body })
    cleanup = () => wrapper.unmount()
    const input = wrapper.get('[role="combobox"]')
    ;(input.element as HTMLInputElement).focus()
    await settle()
    expect(input.attributes('aria-activedescendant')).toBe(wrapper.get('[data-value="create"]').attributes('id'))
    await wrapper.setProps({ query: 'settings' })
    await settle()
    await input.trigger('keydown', { key: 'Enter' })
    expect(wrapper.emitted('select')).toEqual([[options[2], { newTab: false }]])
    await wrapper.setProps({ query: 'missing' })
    await settle()
    expect(input.attributes('aria-activedescendant')).toBeUndefined()
  })

  it('highlights mouse-hovered commands and excludes disabled options', async () => {
    const user = userEvent.setup()
    const wrapper = mount(CommandPalettePanel, { props: { items }, attachTo: document.body })
    cleanup = () => wrapper.unmount()
    ;(wrapper.get('[role="combobox"]').element as HTMLInputElement).focus()
    await settle()
    const enabled = wrapper.get('[role="option"][data-value="create"]')
    await user.hover(enabled.element)
    await settle()
    expect(enabled.attributes('data-highlighted')).toBeDefined()
    expect(wrapper.get('[role="combobox"]').attributes('aria-activedescendant')).toBe(enabled.attributes('id'))
    const disabled = wrapper.get('[role="option"][data-value="disabled"]')
    await user.hover(disabled.element)
    await settle()
    expect(disabled.attributes('data-highlighted')).toBeUndefined()
    await user.click(disabled.element)
    expect(wrapper.emitted('select')).toBeUndefined()
  })

  it('moves highlighting from keyboard selection to the command under the pointer', async () => {
    const user = userEvent.setup()
    const options = [...items, { value: 'settings', label: 'Settings' }]
    const wrapper = mount(CommandPalettePanel, { props: { items: options }, attachTo: document.body })
    cleanup = () => wrapper.unmount()
    const input = wrapper.get('[role="combobox"]')
    ;(input.element as HTMLInputElement).focus()
    await input.trigger('keydown', { key: 'ArrowDown' })
    await input.trigger('keydown', { key: 'ArrowDown' })
    await settle()
    expect(wrapper.get('[data-value="create"]').attributes('data-highlighted')).toBeDefined()
    const hovered = wrapper.get('[data-value="settings"]')
    await user.pointer({ target: hovered.element, coords: { clientX: 30, clientY: 60 } })
    await settle()
    expect(hovered.attributes('data-highlighted')).toBeDefined()
    expect(wrapper.get('[data-value="create"]').attributes('data-highlighted')).toBeUndefined()
    await user.unhover(hovered.element)
    expect(hovered.attributes('data-highlighted')).toBeDefined()
    await input.trigger('keydown', { key: 'Enter' })
    await settle()
    expect(wrapper.emitted('select')?.[0]?.[0]).toEqual(options[2])
  })
})
