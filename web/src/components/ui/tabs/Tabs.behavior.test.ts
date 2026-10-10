// @vitest-environment happy-dom
import { flushPromises, mount } from '@vue/test-utils'
import { expect, it } from 'vitest'
import { Tabs } from './index'

// Based on Kumo's tabs.test.tsx: controlled updates and disabled tab activation.
it('emits the selected tab while leaving disabled items inactive', async () => {
  const wrapper = mount(Tabs, {
    props: { modelValue: 'general', items: [{ value: 'general', label: 'General' }, { value: 'alerts', label: 'Alerts' }, { value: 'locked', label: 'Locked', disabled: true }] },
    attachTo: document.body,
  })
  const tabs = wrapper.findAll('[role="tab"]')
  await tabs[1]!.trigger('click')
  await flushPromises()
  expect(wrapper.emitted('update:modelValue')).toEqual([['alerts']])
  expect(tabs[2]!.element).toHaveProperty('disabled', true)
  await tabs[2]!.trigger('click')
  expect(wrapper.emitted('update:modelValue')).toHaveLength(1)
  wrapper.unmount()
})
