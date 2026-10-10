import userEvent from '@testing-library/user-event'
// @vitest-environment happy-dom
import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h, nextTick, shallowRef } from 'vue'
import { Dialog } from '../dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './index'

const wrappers: ReturnType<typeof mount>[] = []
afterEach(async () => {
  wrappers.splice(0).forEach(wrapper => wrapper.unmount())
  await new Promise(resolve => setTimeout(resolve, 30))
  document.body.innerHTML = ''
})
async function render(props: Record<string, unknown> = {}) {
  const model = shallowRef(props.modelValue as string | string[] | undefined)
  const wrapper = mount(defineComponent({
    setup: () => () => h(Select, { ...props, ...('modelValue' in props ? { 'modelValue': model.value, 'onUpdate:modelValue': (next: string | string[]) => model.value = next } : {}) }, {
      default: () => [
        h(SelectTrigger, { 'aria-label': 'Database' }, { default: () => h(SelectValue, { placeholder: 'Pick a database' }) }),
        h(SelectContent, {}, { default: () => [h(SelectItem, { value: 'sqlite' }, () => 'SQLite'), h(SelectItem, { value: 'postgres' }, () => 'PostgreSQL'), h(SelectItem, { value: 'disabled', disabled: true }, () => 'Unavailable')] }),
      ],
    }),
  }), { attachTo: document.body })
  wrappers.push(wrapper)
  await nextTick()
  await flushPromises()
  return wrapper
}
async function open(wrapper: ReturnType<typeof mount>) {
  await wrapper.find('button').trigger('click')
  await flushPromises()
}

async function renderInsideDialog() {
  const dialogOpen = shallowRef(false)
  const database = shallowRef('sqlite')
  const saved = shallowRef('')
  const wrapper = mount(defineComponent({
    setup: () => () => [
      h('button', { onClick: () => dialogOpen.value = true }, 'Edit database'),
      h(Dialog, { 'open': dialogOpen.value, 'title': 'Database settings', 'onUpdate:open': next => dialogOpen.value = next }, {
        default: () => h(Select<string>, { 'modelValue': database.value, 'onUpdate:modelValue': next => database.value = next }, {
          default: () => [
            h(SelectTrigger, { 'aria-label': 'Database' }, () => h(SelectValue)),
            h(SelectContent, {}, () => [h(SelectItem, { value: 'sqlite' }, () => 'SQLite'), h(SelectItem, { value: 'postgres' }, () => 'PostgreSQL')]),
          ],
        }),
        footer: () => h('button', { onClick: () => saved.value = database.value }, 'Save database'),
      }),
    ],
  }), { attachTo: document.body })
  wrappers.push(wrapper)
  const user = userEvent.setup()
  await user.click(wrapper.get('button').element)
  await vi.waitFor(() => expect(document.querySelector('[role=dialog]')).not.toBeNull())
  await new Promise(resolve => setTimeout(resolve, 30))
  return { user, database, saved }
}

describe('select', () => {
  it('exposes nested dialog options to accessibility and saves a pointer selection', async () => {
    const { user, database, saved } = await renderInsideDialog()
    await user.click(document.querySelector<HTMLElement>('[data-scope=select][data-part=trigger]')!)
    await vi.waitFor(() => expect(document.querySelector('[data-scope=select][data-part=content]')?.getAttribute('data-state')).toBe('open'))
    const option = document.querySelector<HTMLElement>('[data-scope=select][data-part=item][data-value=postgres]')!
    expect(option.closest('[aria-hidden=true]')).toBeNull()
    expect(option.closest('[role=dialog]')).toBe(document.querySelector('[data-scope=dialog][data-part=content]'))
    expect(option.textContent).toContain('PostgreSQL')
    await user.pointer([{ target: option, keys: '[MouseLeft>]' }, { keys: '[/MouseLeft]' }])
    await vi.waitFor(() => expect(database.value).toBe('postgres'))
    const save = Array.from(document.querySelectorAll('button')).find(button => button.textContent === 'Save database')!
    await user.click(save)
    expect(saved.value).toBe('postgres')
  })

  it('selects nested dialog options from the keyboard and restores trigger focus', async () => {
    const { user, database, saved } = await renderInsideDialog()
    const trigger = document.querySelector<HTMLElement>('[data-scope=select][data-part=trigger]')!
    trigger.focus()
    await user.keyboard('{Enter}')
    await vi.waitFor(() => expect(document.activeElement?.getAttribute('data-part')).toBe('list'))
    const option = document.querySelector<HTMLElement>('[data-scope=select][data-part=item][data-value=postgres]')!
    expect(option.closest('[aria-hidden=true]')).toBeNull()
    await user.keyboard('{ArrowDown}{Enter}')
    await vi.waitFor(() => expect(database.value).toBe('postgres'))
    await vi.waitFor(() => expect(document.activeElement).toBe(trigger))
    const save = Array.from(document.querySelectorAll('button')).find(button => button.textContent === 'Save database')!
    await user.click(save)
    expect(saved.value).toBe('postgres')
  })

  it('keeps the selected label and emits a single string when choosing another item', async () => {
    const wrapper = await render({ modelValue: 'sqlite' })
    expect(wrapper.find('button').text()).toBe('SQLite')
    await open(wrapper)
    const option = document.querySelector<HTMLElement>('[data-part=item][data-value=postgres]')!
    expect(option.closest('[data-scope=select][data-part=positioner]')?.parentElement).toBe(document.body)
    option.click()
    await nextTick()
    expect(wrapper.findComponent(Select).emitted('update:modelValue')?.at(-1)).toEqual(['postgres'])
  })

  it('opens from the keyboard, selects the next enabled item, and returns focus to its trigger', async () => {
    const wrapper = await render({ modelValue: 'sqlite' })
    const trigger = wrapper.find('button')
    trigger.element.focus()
    await trigger.trigger('keydown', { key: 'Enter' })
    await flushPromises()
    const content = document.querySelector<HTMLElement>('[data-part=content]')!
    const list = content.querySelector<HTMLElement>('[data-part=list]')!
    expect(document.activeElement).toBe(list)
    list.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowDown', bubbles: true }))
    await nextTick()
    expect(content.getAttribute('aria-activedescendant')).toBe(document.querySelector('[data-part=item][data-value=postgres]')?.id)
    list.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true }))
    await flushPromises()
    expect(wrapper.findComponent(Select).emitted('update:modelValue')?.at(-1)).toEqual(['postgres'])
    await vi.waitFor(() => expect(document.activeElement).toBe(trigger.element))
  })

  it('adds and removes independent multiple selections', async () => {
    const wrapper = await render({ modelValue: ['sqlite'], multiple: true })
    await open(wrapper)
    document.querySelector<HTMLElement>('[data-part=item][data-value=postgres]')!.click()
    await nextTick()
    expect(wrapper.findComponent(Select).emitted('update:modelValue')?.at(-1)).toEqual([['sqlite', 'postgres']])
    document.querySelector<HTMLElement>('[data-part=item][data-value=sqlite]')!.click()
    await nextTick()
    expect(wrapper.findComponent(Select).emitted('update:modelValue')?.at(-1)).toEqual([['postgres']])
  })

  it('initializes an uncontrolled value and submits it through the hidden select', async () => {
    const wrapper = await render({ defaultValue: 'postgres', name: 'database' })
    expect(wrapper.find('button').text()).toBe('PostgreSQL')
    expect(wrapper.find('select').element.value).toBe('postgres')
  })

  it('ignores a disabled option and preserves the selected value', async () => {
    const wrapper = await render({ modelValue: 'sqlite' })
    await open(wrapper)
    document.querySelector<HTMLElement>('[data-part=item][data-value=disabled]')!.click()
    await nextTick()
    expect(wrapper.findComponent(Select).emitted('update:modelValue')).toBeUndefined()
    expect(wrapper.find('button').text()).toBe('SQLite')
  })

  it('prevents opening a read-only select', async () => {
    const wrapper = await render({ modelValue: 'sqlite', readOnly: true })
    await open(wrapper)
    expect(wrapper.find('button').attributes('aria-expanded')).toBe('false')
    expect(wrapper.findComponent(Select).emitted('update:modelValue')).toBeUndefined()
  })
})
