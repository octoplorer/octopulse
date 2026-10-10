// @vitest-environment happy-dom
import { mount } from '@vue/test-utils'
import { expect, it } from 'vitest'
import { Meter } from './index'

it('normalizes a measured range and clamps values at its boundaries', async () => {
  const wrapper = mount(Meter, { props: { label: 'Capacity', value: 60, min: 20, max: 100 } })
  expect(wrapper.get('[role="meter"]').attributes('aria-valuenow')).toBe('60')
  expect(wrapper.text()).toContain('50%')
  await wrapper.setProps({ value: 200 })
  expect(wrapper.get('[role="meter"]').attributes('aria-valuenow')).toBe('100')
  expect(wrapper.text()).toContain('100%')
  await wrapper.setProps({ value: -10 })
  expect(wrapper.text()).toContain('0%')
  wrapper.unmount()
})
