// @vitest-environment happy-dom
import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, expect, it, vi } from 'vitest'
import { h } from 'vue'
import TableContainer from './TableContainer.vue'

const callbacks: ResizeObserverCallback[] = []
const wrappers: ReturnType<typeof mount>[] = []
afterEach(() => {
  wrappers.splice(0).forEach(wrapper => wrapper.unmount())
  callbacks.length = 0
  vi.unstubAllGlobals()
})

it('keeps height limits and explicit keyboard access on the scrolling viewport', () => {
  const wrapper = mount(TableContainer, { attrs: { 'max-h': '400px', 'tabindex': 0, 'role': 'region', 'aria-label': 'Check history' }, slots: { default: () => h('table') } })
  wrappers.push(wrapper)
  const viewport = wrapper.get('.table-wrap')
  expect(viewport.attributes('max-h')).toBe('400px')
  expect(viewport.attributes('tabindex')).toBe('0')
  expect(viewport.attributes('role')).toBe('region')
  expect(viewport.attributes('aria-label')).toBe('Check history')
})

it('exposes overflowing columns through a named keyboard scroll region and visible hint', async () => {
  vi.stubGlobal('ResizeObserver', class {
    constructor(callback: ResizeObserverCallback) { callbacks.push(callback) }
    observe() {}
    unobserve() {}
    disconnect() {}
  })
  const wrapper = mount(TableContainer, { props: { scrollLabel: 'Scroll to see more columns' }, slots: { default: () => h('table', {}, [h('tr', [h('td', 'A long result')])]) } })
  wrappers.push(wrapper)
  await flushPromises()
  const viewport = wrapper.get('.table-wrap')
  Object.defineProperties(viewport.element, { clientWidth: { value: 300, configurable: true }, scrollWidth: { value: 600, configurable: true } })
  callbacks.forEach(callback => callback([], {} as ResizeObserver))
  await flushPromises()
  expect(viewport.attributes('tabindex')).toBe('0')
  expect(viewport.attributes('aria-label')).toBe('Scroll to see more columns')
  expect(wrapper.text()).toContain('Scroll to see more columns')
  Object.defineProperty(viewport.element, 'clientWidth', { value: 700, configurable: true })
  callbacks.forEach(callback => callback([], {} as ResizeObserver))
  await flushPromises()
  expect(viewport.attributes('tabindex')).toBeUndefined()
  expect(wrapper.text()).not.toContain('Scroll to see more columns')
})
