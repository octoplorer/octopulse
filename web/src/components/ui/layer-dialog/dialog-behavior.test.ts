// @vitest-environment happy-dom
import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'
import { nextTick } from 'vue'
import { DeleteResource } from '../delete-resource'
import { Tooltip } from '../tooltip'
import { LayerDialog } from './index'

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

describe('kumo dialog lifecycle', () => {
  it('blocks user dismissal while pending and accepts a programmatic close', async () => {
    const wrapper = track(mount(LayerDialog, { props: { open: true, title: 'Saving', dismissDisabled: true }, attachTo: document.body }))
    await settle()
    const dialog = document.querySelector<HTMLElement>('[role="dialog"]')!
    expect(document.querySelector<HTMLButtonElement>('button[aria-label="Close dialog"]')?.disabled).toBe(true)
    dialog.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
    await settle()
    expect(wrapper.emitted('update:open')).toBeUndefined()
    await wrapper.setProps({ open: false })
    await settle()
    expect(document.querySelector('[role="dialog"]')).toBeNull()
  })

  it('requires matching confirmation and retains a failed deletion for retry', async () => {
    const wrapper = track(mount(DeleteResource, { props: { open: true, resourceName: 'Production', resourceType: 'Monitor', onDelete: async () => {
      throw new Error('Delete failed')
    } }, attachTo: document.body }))
    await settle()
    const buttons = Array.from(document.querySelectorAll<HTMLButtonElement>('button'))
    const action = buttons.find(button => button.textContent?.includes('Delete Monitor'))!
    expect(action.disabled).toBe(true)
    const input = document.querySelector<HTMLInputElement>('input')!
    input.value = 'Production'
    input.dispatchEvent(new Event('input', { bubbles: true }))
    await settle()
    expect(action.disabled).toBe(false)
    action.click()
    await settle()
    expect(wrapper.emitted('confirm')).toEqual([[]])
    expect(document.querySelector('[role="alert"]')?.textContent).toBe('Delete failed')
    expect(action.disabled).toBe(false)
    expect(wrapper.emitted('update:open')).toBeUndefined()
  })

  it('prevents duplicate async deletion and closes after success', async () => {
    let finish: (() => void) | undefined
    let count = 0
    const wrapper = track(mount(DeleteResource, { props: { open: true, resourceName: 'Production', resourceType: 'Monitor', onDelete: () => {
      count++
      return new Promise<void>((resolve) => {
        finish = resolve
      })
    } }, attachTo: document.body }))
    await settle()
    const input = document.querySelector<HTMLInputElement>('input')!
    input.value = 'Production'
    input.dispatchEvent(new Event('input', { bubbles: true }))
    await settle()
    const action = Array.from(document.querySelectorAll<HTMLButtonElement>('button')).find(button => button.textContent?.includes('Delete Monitor'))!
    action.click()
    action.click()
    await settle()
    expect(count).toBe(1)
    expect(action.disabled).toBe(true)
    finish?.()
    await settle()
    expect(wrapper.emitted('update:open')).toEqual([[false]])
  })

  it('shows tooltip content on keyboard focus without pointer interaction', async () => {
    const wrapper = track(mount(Tooltip, { props: { content: 'Refresh monitors', portalled: false, openDelay: 0 }, attachTo: document.body, slots: { default: '<button>Refresh</button>' } }))
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab', bubbles: true }))
    ;(wrapper.get('button').element as HTMLButtonElement).focus()
    await wrapper.get('button').trigger('focus')
    await settle()
    expect(wrapper.get('[role="tooltip"]').isVisible()).toBe(true)
    expect(wrapper.get('button').attributes('aria-describedby')).toBe(wrapper.get('[role="tooltip"]').attributes('id'))
  })
})
