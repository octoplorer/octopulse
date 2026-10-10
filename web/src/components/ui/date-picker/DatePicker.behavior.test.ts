// @vitest-environment happy-dom
import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'
import { nextTick } from 'vue'
import { DateRangePicker } from '../date-range-picker'
import { DatePicker } from './index'

const wrappers: ReturnType<typeof mount>[] = []
afterEach(() => {
  wrappers.splice(0).forEach(wrapper => wrapper.unmount())
  document.body.innerHTML = ''
})
async function render(component: Parameters<typeof mount>[0], props: Record<string, unknown>) {
  const wrapper = mount(component, { props, attachTo: document.body })
  wrappers.push(wrapper)
  await nextTick()
  await flushPromises()
  return wrapper
}

describe('calendar state interactions', () => {
  it('renders one selectable copy of each date in a multi-month DatePicker', async () => {
    const wrapper = await render(DatePicker, { inline: true, numberOfMonths: 2, mode: 'range', modelValue: ['2026-09-30', '2026-10-03'] })
    const octoberFirst = wrapper.findAll('[data-part=table-cell-trigger][data-value="2026-10-01"]')
    expect(octoberFirst).toHaveLength(1)
    expect(octoberFirst[0]!.attributes('aria-disabled')).toBeUndefined()
    expect(octoberFirst[0]!.element.closest('[role=gridcell]')?.getAttribute('aria-selected')).toBe('true')
  })

  it('allows explicitly displayed outside days to be selected', async () => {
    const wrapper = await render(DatePicker, { inline: true, showOutsideDays: true, modelValue: ['2026-10-09'] })
    const outside = wrapper.find('[data-part=table-cell-trigger][data-value="2026-09-30"]')
    expect(outside.attributes('aria-disabled')).toBeUndefined()
    await outside.trigger('click')
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual([['2026-09-30']])
  })

  it('moves keyboard focus across month panels and selects the range endpoint', async () => {
    const wrapper = await render(DatePicker, { inline: true, numberOfMonths: 2, mode: 'range', modelValue: ['2026-09-30'] })
    const start = wrapper.find<HTMLElement>('[data-part=table-cell-trigger][data-value="2026-09-30"]')
    start.element.focus()
    await nextTick()
    await start.trigger('keydown', { key: 'ArrowRight' })
    await flushPromises()
    const end = wrapper.find<HTMLElement>('[data-part=table-cell-trigger][data-value="2026-10-01"]')
    expect(document.activeElement).toBe(end.element)
    expect(end.element.closest('[role=grid]')).toBe(wrapper.findAll('[role=grid]')[1]!.element)
    await end.trigger('keydown', { key: 'Enter' })
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual([['2026-09-30', '2026-10-01']])
  })

  it('previews a hovered range without committing its endpoint', async () => {
    const wrapper = await render(DateRangePicker, { inline: true, numberOfMonths: 1, modelValue: ['2026-10-09'] })
    await wrapper.find('[data-part=table-cell-trigger][data-value="2026-10-12"]').trigger('pointermove', { pointerType: 'mouse' })
    expect(wrapper.find('[data-part=table-cell-trigger][data-value="2026-10-11"]').element.hasAttribute('data-in-hover-range')).toBe(true)
    expect(wrapper.find('[data-part=table-cell-trigger][data-value="2026-10-12"]').element.hasAttribute('data-selected')).toBe(false)
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
  })

  it('keeps an unavailable date unselected when clicked', async () => {
    const wrapper = await render(DatePicker, { inline: true, modelValue: ['2026-10-09'], isDateUnavailable: (date: { toString: () => string }) => date.toString() === '2026-10-10' })
    const unavailable = wrapper.find('[data-part=table-cell-trigger][data-value="2026-10-10"]')
    expect(unavailable.attributes('aria-disabled')).toBe('true')
    await unavailable.trigger('click')
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
    expect(wrapper.find('[data-selected][data-part=table-cell-trigger]').text()).toBe('9')
  })
})
