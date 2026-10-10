// @vitest-environment happy-dom
import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'
import { defineComponent, h, nextTick } from 'vue'
import { Autocomplete } from '../autocomplete'
import { Combobox } from '../combobox'
import { DatePicker } from '../date-picker'
import { DateRangePicker } from '../date-range-picker'
import { Field } from '../field'
import { RadioGroup, RadioItem } from '../radio'
import { SensitiveInput } from '../sensitive-input'
import { Slider } from '../slider'
import { TagInput } from '../tag-input'
import { Checkbox, CheckboxGroup, CheckboxItem } from './index'

const wrappers: ReturnType<typeof mount>[] = []
afterEach(() => {
  wrappers.splice(0).forEach(wrapper => wrapper.unmount())
  document.body.innerHTML = ''
})
async function render(component: Parameters<typeof mount>[0], options?: Parameters<typeof mount>[1]) {
  const wrapper = mount(component, { attachTo: document.body, ...options })
  wrappers.push(wrapper)
  await nextTick()
  await flushPromises()
  return wrapper
}

describe('input components', () => {
  it('emits a checkbox boolean and exposes mixed selection to assistive technology', async () => {
    const wrapper = await render(Checkbox, { props: { label: 'Notifications', modelValue: 'indeterminate' } })
    expect(wrapper.find('input').element.indeterminate).toBe(true)
    await wrapper.find('input').trigger('click')
    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual([true])
  })

  it.each(['default', 'card'] as const)('tracks keyboard focus separately from pointer focus on a %s checkbox', async (appearance) => {
    const wrapper = await render(Checkbox, { props: { label: 'Notifications', appearance } })
    const input = wrapper.find('input').element
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab', bubbles: true }))
    input.focus()
    await nextTick()
    expect(document.activeElement).toBe(input)
    expect(wrapper.find('[data-part=root]').element.hasAttribute('data-focus-visible')).toBe(true)

    input.blur()
    document.dispatchEvent(new PointerEvent('pointerdown', { bubbles: true }))
    input.focus()
    await nextTick()
    expect(document.activeElement).toBe(input)
    expect(wrapper.find('[data-part=root]').element.hasAttribute('data-focus-visible')).toBe(false)
  })

  it('adds a checkbox choice to the group model', async () => {
    const wrapper = await render(defineComponent({
      setup: () => () => h(CheckboxGroup, { modelValue: ['email'], label: 'Channels' }, {
        default: () => [h(CheckboxItem, { value: 'email', label: 'Email' }), h(CheckboxItem, { value: 'sms', label: 'SMS' })],
      }),
    }))
    await wrapper.findAll('input')[1]!.trigger('click')
    expect(wrapper.findComponent(CheckboxGroup).emitted('update:modelValue')?.[0]).toEqual([['email', 'sms']])
  })

  it('selects a radio choice using its declared value', async () => {
    const wrapper = await render(defineComponent({
      setup: () => () => h(RadioGroup, { label: 'Interval', modelValue: 'short' }, {
        default: () => [h(RadioItem, { value: 'short', label: 'Short' }), h(RadioItem, { value: 'long', label: 'Long' })],
      }),
    }))
    await wrapper.findAll('input')[1]!.trigger('click')
    expect(wrapper.findComponent(RadioGroup).emitted('update:modelValue')?.[0]).toEqual(['long'])
  })

  it('filters combobox items without losing the declared selected label', async () => {
    const wrapper = await render(Combobox, { props: { modelValue: 'postgres', items: [{ value: 'sqlite', label: 'SQLite' }, { value: 'postgres', label: 'PostgreSQL' }], label: 'Database' } })
    expect(wrapper.find('input').element.value).toBe('PostgreSQL')
    wrapper.find('input').element.focus()
    await nextTick()
    await wrapper.find('input').setValue('sql')
    await nextTick()
    expect(wrapper.find('[role=combobox]').attributes('aria-expanded')).toBe('true')
  })

  it('allows autocomplete free text without selecting a suggestion', async () => {
    const wrapper = await render(Autocomplete, { props: { label: 'Host', items: ['example.com'] } })
    wrapper.find('input').element.focus()
    await nextTick()
    await wrapper.find('input').setValue('my-host.test')
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual(['my-host.test'])
  })

  it('renders all slider thumbs with separate submitted values', async () => {
    const wrapper = await render(Slider, { props: { label: 'Latency', name: 'latency', modelValue: [10, 90], min: 0, max: 100 } })
    expect(wrapper.findAll('[role=slider]').map(thumb => thumb.attributes('aria-valuenow'))).toEqual(['10', '90'])
    expect(wrapper.findAll('input').map(input => input.element.value)).toEqual(['10', '90'])
  })

  it('reveals a sensitive value without changing its content', async () => {
    const wrapper = await render(SensitiveInput, { props: { label: 'API key', modelValue: 'secret', copyable: false } })
    expect(wrapper.find('input').attributes('type')).toBe('password')
    await wrapper.find('button').trigger('click')
    expect(wrapper.find('input').attributes('type')).toBe('text')
    expect(wrapper.find('input').element.value).toBe('secret')
  })

  it('commits unique trimmed tags and retains rejected input', async () => {
    const wrapper = await render(TagInput, { props: { label: 'Hosts', modelValue: ['alpha'], maxValues: 2 } })
    wrapper.find<HTMLInputElement>('input[data-part=input]').element.focus()
    await nextTick()
    await wrapper.find<HTMLInputElement>('input[data-part=input]').setValue(' alpha, beta, gamma ')
    await wrapper.find<HTMLInputElement>('input[data-part=input]').trigger('keydown', { key: 'Enter' })
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual([['alpha', 'beta']])
    expect(wrapper.find<HTMLInputElement>('input[data-part=input]').element.value).toBe('gamma')
    expect(wrapper.find('[role=alert]').exists()).toBe(true)
  })

  it('shows a selected ISO day in an inline calendar', async () => {
    const wrapper = await render(DatePicker, { props: { modelValue: ['2026-10-09'], inline: true, locale: 'en-US' } })
    expect(wrapper.find('[data-selected][data-part=table-cell-trigger]').text()).toBe('9')
    expect(wrapper.findAll('[role=grid]').length).toBeGreaterThan(0)
  })

  it('moves calendar focus with an arrow key without changing its selected date', async () => {
    const wrapper = await render(DatePicker, { props: { modelValue: ['2026-10-09'], inline: true, locale: 'en-US' } })
    const day = wrapper.find<HTMLElement>('[data-part=table-cell-trigger][data-value="2026-10-09"]')
    day.element.focus()
    await nextTick()
    await day.trigger('keydown', { key: 'ArrowRight' })
    await flushPromises()
    expect(document.activeElement?.getAttribute('data-value')).toBe('2026-10-10')
    expect(wrapper.find('[data-selected][data-part=table-cell-trigger]').text()).toBe('9')
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
  })

  it('shows both range endpoints in accessible inputs', async () => {
    const wrapper = await render(DateRangePicker, { props: { modelValue: ['2026-10-09', '2026-10-12'], locale: 'en-US', label: 'Time range' } })
    expect(wrapper.findAll('input').length).toBe(2)
    expect(wrapper.findAll('input').every(input => input.element.value.length > 0)).toBe(true)
  })

  it('forwards a label to the native checkbox when there is no visible label', async () => {
    const wrapper = await render(Checkbox, { attrs: { 'aria-label': 'Select monitor' } })
    expect(wrapper.find('input').attributes('aria-label')).toBe('Select monitor')
  })

  it('selects a segmented radio choice with the arrow keys', async () => {
    const wrapper = await render(defineComponent({
      setup: () => () => h(RadioGroup, { label: 'Interval', modelValue: 'short', appearance: 'segmented' }, {
        default: () => [h(RadioItem, { value: 'short', label: 'Short' }), h(RadioItem, { value: 'long', label: 'Long' })],
      }),
    }))
    wrapper.findAll('input')[0]!.element.focus()
    await nextTick()
    await wrapper.findAll('input')[0]!.trigger('keydown', { key: 'ArrowRight' })
    await flushPromises()
    expect(wrapper.findComponent(RadioGroup).emitted('update:modelValue')?.at(-1)).toEqual(['long'])
  })

  it('selects a filtered combobox option from the keyboard', async () => {
    const wrapper = await render(Combobox, { props: { items: [{ value: 'sqlite', label: 'SQLite' }, { value: 'postgres', label: 'PostgreSQL' }], label: 'Database' } })
    wrapper.find('input').element.focus()
    await nextTick()
    await wrapper.find('input').setValue('post')
    await wrapper.find('input').trigger('keydown', { key: 'ArrowDown' })
    await wrapper.find('input').trigger('keydown', { key: 'Enter' })
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual(['postgres'])
  })

  it('removes one selected combobox chip without dropping the other selection', async () => {
    const wrapper = await render(Combobox, { props: { items: ['alpha', 'beta'], multiple: true, modelValue: ['alpha', 'beta'] } })
    await wrapper.find('button[aria-label="Remove alpha"]').trigger('click')
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual([['beta']])
  })

  it('uses the suggestion label when autocomplete selects an object', async () => {
    const wrapper = await render(Autocomplete, { props: { label: 'Database', items: [{ value: 'postgres', label: 'PostgreSQL' }] } })
    wrapper.find('input').element.focus()
    await nextTick()
    await wrapper.find('input').setValue('post')
    await wrapper.find('input').trigger('keydown', { key: 'ArrowDown' })
    await wrapper.find('input').trigger('keydown', { key: 'Enter' })
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual(['PostgreSQL'])
  })

  it('advances a slider by its step and clamps at its maximum', async () => {
    const wrapper = await render(Slider, { props: { modelValue: 8, step: 2, min: 0, max: 10 } })
    wrapper.find<HTMLElement>('[role=slider]').element.focus()
    await nextTick()
    await wrapper.find('[role=slider]').trigger('keydown', { key: 'ArrowRight' })
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual([10])
    await wrapper.find('[role=slider]').trigger('keydown', { key: 'ArrowRight' })
    expect(wrapper.find('[role=slider]').attributes('aria-valuenow')).toBe('10')
  })

  it('retains invalid pasted tags after accepting the preceding valid tags', async () => {
    const wrapper = await render(TagInput, { props: { label: 'Emails', validateValue: (value: string) => value.includes('@') } })
    const clipboardData = new DataTransfer()
    clipboardData.setData('text', 'valid@example.com, invalid')
    wrapper.find('input[data-part=input]').element.dispatchEvent(new ClipboardEvent('paste', { bubbles: true, cancelable: true, clipboardData }))
    await nextTick()
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual([['valid@example.com']])
    expect(wrapper.find<HTMLInputElement>('input[data-part=input]').element.value).toBe('invalid')
    expect(wrapper.find('[role=alert]').text()).toBe('"invalid" is not valid.')
  })

  it('removes a tag using its labelled delete action', async () => {
    const wrapper = await render(TagInput, { props: { modelValue: ['alpha', 'beta'] } })
    await wrapper.find('button[aria-label="Remove alpha"]').trigger('click')
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual([['beta']])
  })

  it('protects disabled tag removal', async () => {
    const wrapper = await render(TagInput, { props: { modelValue: ['beta'], disabled: true } })
    await wrapper.find('button[aria-label="Remove beta"]').trigger('click')
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
  })

  it('selects two date range endpoints and resets them', async () => {
    const wrapper = await render(DateRangePicker, { props: { modelValue: ['2026-10-09'], inline: true, numberOfMonths: 1, clearable: true, locale: 'en-US' } })
    await wrapper.find('[data-part=table-cell-trigger][data-value="2026-10-12"]').trigger('click')
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual([['2026-10-09', '2026-10-12']])
    await wrapper.find('[data-part=clear-trigger]').trigger('click')
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual([[]])
  })

  it('submits inline dates as ISO form values', async () => {
    const wrapper = await render(defineComponent({ setup: () => () => h('form', [h(DatePicker, { name: 'dates', mode: 'multiple', modelValue: ['2026-10-09', '2026-10-12'], inline: true })]) }))
    expect(new FormData(wrapper.find('form').element).getAll('dates')).toEqual(['2026-10-09', '2026-10-12'])
  })

  it('gives separate calendar instances distinct accessible grid identities', async () => {
    const wrapper = await render(defineComponent({ setup: () => () => h('div', [h(DatePicker, { inline: true }), h(DatePicker, { inline: true })]) }))
    const ids = wrapper.findAll('[role=grid]').map(grid => grid.attributes('id'))
    expect(new Set(ids).size).toBe(ids.length)
  })

  it('protects combobox chip removal when the containing field is disabled', async () => {
    const wrapper = await render(defineComponent({ setup: () => () => h(Field, { disabled: true }, { default: () => h(Combobox, { items: ['alpha'], modelValue: ['alpha'], multiple: true }) }) }))
    const remove = wrapper.find('button[aria-label="Remove alpha"]')
    expect(remove.attributes('disabled')).toBeDefined()
    await remove.trigger('click')
    expect(wrapper.findComponent(Combobox).emitted('update:modelValue')).toBeUndefined()
  })

  it('disables every action on a sensitive input inside a disabled field', async () => {
    const wrapper = await render(defineComponent({ setup: () => () => h(Field, { disabled: true }, { default: () => h(SensitiveInput, { modelValue: 'secret' }) }) }))
    expect(wrapper.findAll('button').every(button => button.attributes('disabled') !== undefined)).toBe(true)
  })

  it('preserves the caller accessible label on an unlabelled sensitive input', async () => {
    const wrapper = await render(SensitiveInput, { attrs: { 'aria-label': 'API token' } })
    expect(wrapper.find('input').attributes('aria-label')).toBe('API token')
  })

  it('omits inline date form values inside a disabled field', async () => {
    const wrapper = await render(defineComponent({ setup: () => () => h('form', [h(Field, { disabled: true }, { default: () => h(DatePicker, { name: 'date', modelValue: ['2026-10-09'], inline: true }) })]) }))
    expect(new FormData(wrapper.find('form').element).getAll('date')).toEqual([])
  })

  it('submits selected combobox values instead of display labels', async () => {
    const wrapper = await render(defineComponent({ setup: () => () => h('form', [h(Combobox, { name: 'database', items: [{ value: 'postgres', label: 'PostgreSQL' }, { value: 'sqlite', label: 'SQLite' }], modelValue: ['postgres', 'sqlite'], multiple: true })]) }))
    expect(new FormData(wrapper.find('form').element).getAll('database')).toEqual(['postgres', 'sqlite'])
  })
})
