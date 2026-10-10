// @vitest-environment happy-dom
import { mount } from '@vue/test-utils'
import { expect, it, vi } from 'vitest'
import Button from './Button.vue'

// Adapted from Kumo's button.test.tsx: disabled polymorphic controls cannot activate.
it('blocks link activation while disabled and restores it when enabled', async () => {
  const onClick = vi.fn()
  const wrapper = mount(Button, { props: { asChild: true, disabled: true }, attrs: { onClick }, slots: { default: '<a href="/app/monitors">Monitors</a>' } })
  const event = new MouseEvent('click', { bubbles: true, cancelable: true })
  wrapper.get('a').element.dispatchEvent(event)
  expect(event.defaultPrevented).toBe(true)
  expect(onClick).not.toHaveBeenCalled()
  await wrapper.setProps({ disabled: false })
  await wrapper.get('a').trigger('click')
  expect(onClick).toHaveBeenCalledOnce()
  wrapper.unmount()
})
