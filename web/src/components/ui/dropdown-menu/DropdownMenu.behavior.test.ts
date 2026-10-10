import userEvent from '@testing-library/user-event'
// @vitest-environment happy-dom
import { DOMWrapper, mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'
import { defineComponent, h, nextTick, ref } from 'vue'
import { Button } from '../button'
import { Dialog } from '../dialog'
import { DropdownMenu, DropdownMenuCheckboxItem, DropdownMenuContent, DropdownMenuItem, DropdownMenuRadioGroup, DropdownMenuRadioItem, DropdownMenuSub, DropdownMenuSubTrigger, DropdownMenuTrigger } from './index'

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

describe('dropdown menu repeated option interactions', () => {
  it('keeps actions accessible and clickable inside a modal dialog', async () => {
    const selected = ref('')
    const wrapper = mount(Dialog, { props: { open: true, title: 'Monitor settings' }, attachTo: document.body, slots: { default: () => h(DropdownMenu, { onSelect: (value: string) => {
      selected.value = value
    } }, () => [
      h(DropdownMenuTrigger, { asChild: true }, () => h(Button, {}, () => 'Actions')),
      h(DropdownMenuContent, {}, () => h(DropdownMenuItem, { value: 'export' }, () => 'Export')),
    ]) } })
    wrappers.push(wrapper)
    await settle()
    const user = userEvent.setup()
    const trigger = Array.from(document.querySelectorAll<HTMLButtonElement>('button')).find(button => button.textContent === 'Actions')!
    await user.click(trigger)
    await settle()
    const item = document.querySelector<HTMLElement>('[role="menuitem"]')!
    expect(item.closest('[aria-hidden="true"]')).toBeNull()
    await user.click(item)
    await settle()
    expect(selected.value).toBe('export')
    expect(document.querySelector('[role="dialog"]')).not.toBeNull()
  })
  it('keeps controlled radio, checkbox and submenu state stable across reopening', async () => {
    const checked = ref(true)
    const interval = ref('30')
    const selected = ref('')
    const errors: unknown[] = []
    const wrapper = mount(defineComponent({
      setup: () => () => h(DropdownMenu, { onSelect: (value: string) => { selected.value = value } }, () => [
        h(DropdownMenuTrigger, { asChild: true }, () => h(Button, {}, () => 'Monitor actions')),
        h(DropdownMenuContent, { portalled: false }, () => [
          h(DropdownMenuCheckboxItem, { 'checked': checked.value, 'value': 'email', 'onUpdate:checked': (value: boolean) => { checked.value = value } }, () => 'Email'),
          h(DropdownMenuRadioGroup, { 'modelValue': interval.value, 'onUpdate:modelValue': (value: string) => { interval.value = value } }, () => [h(DropdownMenuRadioItem, { value: '30' }, () => '30 seconds'), h(DropdownMenuRadioItem, { value: '60' }, () => '60 seconds')]),
          h(DropdownMenuSub, { placement: 'right-start', onSelect: (value: string) => { selected.value = value } }, () => [h(DropdownMenuSubTrigger, {}, () => 'More actions'), h(DropdownMenuContent, { portalled: false }, () => h(DropdownMenuItem, { value: 'export' }, () => 'Export'))]),
        ]),
      ]),
    }), { attachTo: document.body, global: { config: { errorHandler: error => errors.push(error) } } })
    wrappers.push(wrapper)
    for (let index = 0; index < 8; index++) {
      await wrapper.get('button').trigger('click')
      await settle()
      await wrapper.get('[role="menuitemcheckbox"]').trigger('click')
      await settle()
      expect(checked.value).toBe(index % 2 !== 0)
      const next = index % 2 ? '30' : '60'
      await wrapper.get(`[role="menuitemradio"][data-value="${next}"]`).trigger('click')
      await settle()
      expect(interval.value).toBe(next)
      expect(wrapper.get('button').attributes('aria-expanded')).toBe('false')
      expect(errors).toEqual([])
    }
  })

  it.each([true, false])('opens a nested submenu without recursive updates (portalled: %s)', async (portalled) => {
    const selected = ref('')
    const errors: unknown[] = []
    const wrapper = mount(DropdownMenu, { attachTo: document.body, global: { config: { errorHandler: error => errors.push(error) } }, slots: { default: () => [
      h(DropdownMenuTrigger, { asChild: true }, () => h(Button, {}, () => 'Actions')),
      h(DropdownMenuContent, { portalled: false }, () => h(DropdownMenuSub, { placement: 'right-start', onSelect: (value: string) => { selected.value = value } }, () => [h(DropdownMenuSubTrigger, {}, () => 'More'), h(DropdownMenuContent, { portalled }, () => h(DropdownMenuItem, { value: 'export' }, () => 'Export'))])),
    ] } })
    wrappers.push(wrapper)
    ;(wrapper.get('button').element as HTMLButtonElement).focus()
    await wrapper.get('button').trigger('click')
    await settle()
    const content = wrapper.findAll('[role="menu"]')[0]!
    await content.trigger('keydown', { key: 'End' })
    await content.trigger('keydown', { key: 'ArrowRight' })
    await settle()
    const submenu = new DOMWrapper(document.querySelectorAll<HTMLElement>('[role="menu"]')[1]!)
    expect(document.activeElement).toBe(submenu.element)
    await submenu.trigger('keydown', { key: 'Enter' })
    await settle()
    expect(selected.value).toBe('export')
    expect(errors).toEqual([])
  })
})
