// @vitest-environment happy-dom
import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'
import { defineComponent, h, nextTick, shallowRef } from 'vue'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '../collapsible'
import { CommandPalettePanel } from '../command-palette'
import { isDeleteConfirmed } from '../delete-resource/confirmation'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '../dropdown-menu'
import { MenuBar, MenuBarItem } from '../menubar'
import { Pagination } from '../pagination'
import { Toolbar, ToolbarButton } from '../toolbar'
import { Popover, PopoverContent, PopoverTitle, PopoverTrigger } from './index'

const wrappers: ReturnType<typeof mount>[] = []
function track<T extends ReturnType<typeof mount>>(wrapper: T): T {
  wrappers.push(wrapper)
  return wrapper
}
async function settle() {
  await nextTick()
  await new Promise(resolve => setTimeout(resolve, 30))
}
afterEach(() => {
  wrappers.forEach(wrapper => wrapper.unmount())
  wrappers.length = 0
  document.body.innerHTML = ''
})

describe('kumo overlay and navigation interactions', () => {
  it('opens the controlled popover from its trigger', async () => {
    const open = shallowRef(false)
    const wrapper = track(mount(defineComponent({
      setup: () => () => h(Popover, { 'open': open.value, 'onUpdate:open': (value: boolean | undefined) => { open.value = value ?? false } }, { default: () => [h(PopoverTrigger, {}, () => 'Settings'), h(PopoverContent, { portalled: false }, () => h(PopoverTitle, {}, () => 'Display settings'))] }),
    }), { attachTo: document.body }))
    await wrapper.get('[data-part="trigger"]').trigger('click')
    await settle()
    expect(open.value).toBe(true)
    expect(wrapper.get('[data-part="trigger"]').attributes('aria-expanded')).toBe('true')
    expect(wrapper.get('[role="dialog"]').attributes('aria-labelledby')).toBe(wrapper.get('[data-part="title"]').attributes('id'))
  })

  it('does not select disabled menu items and emits active actions', async () => {
    const wrapper = track(mount(DropdownMenu, { props: { open: true }, attachTo: document.body, slots: { default: () => [h(DropdownMenuTrigger, {}, () => 'Actions'), h(DropdownMenuContent, { portalled: false }, () => [h(DropdownMenuItem, { value: 'disabled', disabled: true }, () => 'Disabled'), h(DropdownMenuItem, { value: 'edit' }, () => 'Edit')])] } }))
    await settle()
    await wrapper.get('[data-value="disabled"]').trigger('click')
    expect(wrapper.emitted('select')).toBeUndefined()
    await wrapper.get('[data-value="edit"]').trigger('pointerdown', { pointerType: 'mouse', button: 0 })
    await wrapper.get('[data-value="edit"]').trigger('click', { button: 0 })
    await settle()
    expect(wrapper.emitted('select')).toEqual([['edit']])
    expect(wrapper.emitted('update:open')).toEqual([[false]])
  })

  it('toggles collapsible state and respects disabled triggers', async () => {
    const wrapper = track(mount(Collapsible, { props: { disabled: false }, slots: { default: () => [h(CollapsibleTrigger, {}, () => 'Advanced'), h(CollapsibleContent, {}, () => 'Settings')] } }))
    await wrapper.get('button').trigger('click')
    expect(wrapper.emitted('update:open')).toEqual([[true]])
    await wrapper.setProps({ disabled: true })
    await wrapper.get('button').trigger('click')
    expect(wrapper.emitted('update:open')).toHaveLength(1)
  })

  it('disables first-page controls and clamps typed pages to the last page', async () => {
    const wrapper = track(mount(Pagination, { props: { count: 95, pageSize: 10 } }))
    expect(wrapper.get('button[aria-label="Previous page"]').attributes('disabled')).toBeDefined()
    const input = wrapper.get('input[aria-label="Page number"]')
    await input.setValue('500')
    await input.trigger('keydown', { key: 'Enter' })
    expect(wrapper.emitted('update:page')?.at(-1)).toEqual([10])
    await wrapper.setProps({ page: 10 })
    expect(wrapper.get('button[aria-label="Next page"]').attributes('disabled')).toBeDefined()
  })

  it('uses sequential controls and hasNextPage for unknown totals', async () => {
    const wrapper = track(mount(Pagination, { props: { page: 2, hasNextPage: false } }))
    expect(wrapper.find('input[aria-label="Page number"]').exists()).toBe(false)
    expect(wrapper.get('button[aria-label="Next page"]').attributes('disabled')).toBeDefined()
    await wrapper.get('button[aria-label="Previous page"]').trigger('click')
    expect(wrapper.emitted('update:page')).toEqual([[1]])
  })

  it('filters commands and selects the highlighted result with Enter', async () => {
    const items = [{ value: 'create', label: 'Create monitor' }, { value: 'settings', label: 'Organization settings' }]
    const wrapper = track(mount(CommandPalettePanel, { props: { items, label: 'Search commands' }, attachTo: document.body }))
    const input = wrapper.get('[role="combobox"]')
    await input.setValue('monitor')
    await settle()
    expect(wrapper.findAll('[role="option"]')).toHaveLength(1)
    await input.trigger('keydown', { key: 'ArrowDown' })
    await input.trigger('keydown', { key: 'Enter' })
    await settle()
    expect(wrapper.emitted('select')?.[0]?.[0]).toEqual(items[0])
  })

  it('moves toolbar focus with arrows while skipping disabled controls', async () => {
    const wrapper = track(mount(Toolbar, { props: { label: 'Monitor actions' }, attachTo: document.body, slots: { default: () => [h(ToolbarButton, {}, () => 'Refresh'), h(ToolbarButton, { disabled: true }, () => 'Disabled'), h(ToolbarButton, {}, () => 'Create')] } }))
    await settle()
    const buttons = wrapper.findAll('button')
    ;(buttons[0]!.element as HTMLButtonElement).focus()
    await buttons[0]!.trigger('keydown', { key: 'ArrowRight' })
    expect(document.activeElement).toBe(buttons[2]!.element)
    await buttons[2]!.trigger('keydown', { key: 'Home' })
    expect(document.activeElement).toBe(buttons[0]!.element)
  })

  it('honors uncontrolled defaultOpen instead of Vue Boolean prop casting', async () => {
    const wrapper = track(mount(Popover, { props: { defaultOpen: true }, slots: { default: () => [h(PopoverTrigger, {}, () => 'Settings'), h(PopoverContent, { portalled: false }, () => 'Display settings')] } }))
    expect(wrapper.get('[data-part="trigger"]').attributes('aria-expanded')).toBe('true')
  })

  it('selects a menubar tab with keyboard navigation and skips disabled items', async () => {
    const wrapper = track(mount(MenuBar, { props: { modelValue: 'all', label: 'Monitor filter' }, attachTo: document.body, slots: { default: () => [h(MenuBarItem, { value: 'all' }, () => 'All'), h(MenuBarItem, { value: 'disabled', disabled: true }, () => 'Disabled'), h(MenuBarItem, { value: 'active' }, () => 'Active')] } }))
    const tabs = wrapper.findAll('[role="tab"]')
    ;(tabs[0]!.element as HTMLButtonElement).focus()
    await tabs[0]!.trigger('keydown', { key: 'ArrowRight' })
    await settle()
    expect(document.activeElement).toBe(tabs[2]!.element)
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual(['active'])
  })

  it('requires the exact non-empty resource name before confirming deletion', () => {
    expect(isDeleteConfirmed('Production', 'Production')).toBe(true)
    expect(isDeleteConfirmed('production', 'Production')).toBe(false)
    expect(isDeleteConfirmed('production', 'Production', false)).toBe(true)
    expect(isDeleteConfirmed('', '')).toBe(false)
  })
})
