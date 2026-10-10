// @vitest-environment happy-dom
import { flushPromises, mount } from '@vue/test-utils'
import { expect, it, vi } from 'vitest'
import { createKumoToastManager, Toasty } from './index'

// Adapted from Kumo's external manager, duplicate id and dismissal scenarios.
it('updates one notification in place and dismisses it through the close control', async () => {
  const manager = createKumoToastManager({ duration: Infinity })
  const wrapper = mount(Toasty, { props: { manager, closeLabel: 'Dismiss' }, attachTo: document.body })
  const id = manager.create({ id: 'save', title: 'Saving', type: 'loading' })
  await flushPromises()
  manager.update(id, { title: 'Saved', type: 'success' })
  await flushPromises()
  const notifications = document.querySelectorAll('[data-scope="toast"][data-part="root"]')
  expect(notifications).toHaveLength(1)
  expect(notifications[0]?.textContent).toContain('Saved')
  const close = document.querySelector<HTMLButtonElement>('button[aria-label="Dismiss"]')!
  close.click()
  await flushPromises()
  await vi.waitFor(() => expect(manager.isVisible(id)).toBe(false))
  wrapper.unmount()
  manager.remove()
})
