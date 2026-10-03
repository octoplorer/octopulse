import type { EffectScope } from 'vue'
import { afterEach, expect, it, vi } from 'vitest'
import { effectScope, nextTick, ref } from 'vue'
import { useSidebar } from './sidebar.ts'

const storage = vi.hoisted(() => new Map<string, string>())

vi.mock('@vueuse/core', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@vueuse/core')>()
  return {
    ...actual,
    useLocalStorage: <T extends boolean | number>(key: string, initialValue: T) => actual.useStorage<T>(key, initialValue, {
      getItem: key => storage.get(key) ?? null,
      setItem: (key, value) => storage.set(key, value),
      removeItem: key => storage.delete(key),
    }),
  }
})

const scopes: EffectScope[] = []

function setup(isMobile = false) {
  const scope = effectScope()
  scopes.push(scope)
  const mobile = ref(isMobile)
  const sidebar = scope.run(() => useSidebar(mobile))!
  return { sidebar, mobile, scope }
}

afterEach(() => {
  scopes.splice(0).forEach(scope => scope.stop())
  storage.clear()
})

it('persists desktop collapse and the last expanded width across sidebar mounts', async () => {
  const first = setup()
  expect(first.sidebar.state.value).toBe('expanded')
  expect(first.sidebar.railWidth.value).toBe(256)
  first.sidebar.beginResize(256)
  first.sidebar.resizeTo(320)
  first.sidebar.endResize()
  first.sidebar.toggle()
  await nextTick()
  first.scope.stop()

  const { sidebar } = setup()
  expect(sidebar.state.value).toBe('collapsed')
  expect(sidebar.railWidth.value).toBe(57)
  expect(sidebar.panelWidth.value).toBe(57)
  sidebar.toggle()
  expect(sidebar.panelWidth.value).toBe(320)
  expect(sidebar.railWidth.value).toBe(320)
})

it('peeks without moving content and remains open while hovered or focused', () => {
  const { sidebar } = setup()
  sidebar.toggle()
  sidebar.setHovered(true)
  expect(sidebar.state.value).toBe('peeking')
  expect(sidebar.railWidth.value).toBe(57)
  expect(sidebar.panelWidth.value).toBe(256)
  expect(sidebar.open.value).toBe(false)
  sidebar.setFocused(true)
  sidebar.setHovered(false)
  expect(sidebar.state.value).toBe('peeking')
  sidebar.setFocused(false)
  expect(sidebar.state.value).toBe('collapsed')
})

it('explicit dismissal waits until both pointer and focus leave before peeking again', () => {
  const { sidebar } = setup()
  sidebar.toggle()
  sidebar.setHovered(true)
  sidebar.setFocused(true)
  sidebar.dismissPeek()
  expect(sidebar.state.value).toBe('collapsed')
  sidebar.setHovered(false)
  sidebar.setHovered(true)
  expect(sidebar.state.value).toBe('collapsed')
  sidebar.setHovered(false)
  sidebar.setFocused(false)
  sidebar.setFocused(true)
  expect(sidebar.state.value).toBe('peeking')
})

it('collapsing with the toggle does not immediately reopen under the pointer', () => {
  const { sidebar } = setup()
  sidebar.setHovered(true)
  sidebar.toggle()
  expect(sidebar.state.value).toBe('collapsed')
  sidebar.setHovered(false)
  sidebar.setHovered(true)
  expect(sidebar.state.value).toBe('peeking')
  sidebar.toggle()
  expect(sidebar.state.value).toBe('expanded')
  expect(sidebar.open.value).toBe(true)
})

it('dragging collapses below the minimum, preserves width, and expands from the rail edge', () => {
  const { sidebar } = setup()
  sidebar.setHovered(true)
  sidebar.beginResize(256)
  sidebar.resizeTo(340)
  expect(sidebar.width.value).toBe(340)
  sidebar.resizeTo(190)
  expect(sidebar.state.value).toBe('collapsed')
  expect(sidebar.width.value).toBe(340)
  sidebar.endResize()
  expect(sidebar.state.value).toBe('collapsed')

  sidebar.beginResize(57)
  sidebar.resizeTo(199)
  expect(sidebar.open.value).toBe(false)
  sidebar.resizeTo(200)
  expect(sidebar.open.value).toBe(true)
  expect(sidebar.width.value).toBe(200)
  sidebar.resizeTo(900)
  expect(sidebar.width.value).toBe(400)
  sidebar.endResize()
  sidebar.resizeTo(250)
  expect(sidebar.width.value).toBe(400)
})

it('dragging a peeking panel pins it and starts from its visible width', () => {
  const { sidebar } = setup()
  sidebar.toggle()
  sidebar.setHovered(true)
  expect(sidebar.state.value).toBe('peeking')
  sidebar.beginResize(256)
  expect(sidebar.isResizing.value).toBe(true)
  expect(sidebar.open.value).toBe(true)
  expect(sidebar.state.value).toBe('expanded')
  expect(sidebar.railWidth.value).toBe(256)
  sidebar.resizeTo(280)
  expect(sidebar.width.value).toBe(280)
  expect(sidebar.panelWidth.value).toBe(280)
  sidebar.endResize()
  sidebar.setHovered(false)
  expect(sidebar.state.value).toBe('expanded')
})

it('keyboard resizing supports rail expansion, width bounds, and collapse', () => {
  const { sidebar } = setup()
  expect(sidebar.resizeWithKey('Tab')).toBe(false)
  expect(sidebar.resizeWithKey('Home')).toBe(true)
  expect(sidebar.open.value).toBe(false)
  sidebar.resizeWithKey('ArrowLeft')
  expect(sidebar.width.value).toBe(256)
  sidebar.resizeWithKey('ArrowRight')
  expect(sidebar.open.value).toBe(true)
  expect(sidebar.width.value).toBe(200)
  sidebar.resizeWithKey('ArrowRight')
  expect(sidebar.width.value).toBe(210)
  sidebar.resizeWithKey('ArrowLeft')
  expect(sidebar.width.value).toBe(200)
  expect(sidebar.open.value).toBe(true)
  sidebar.resizeWithKey('ArrowLeft')
  expect(sidebar.open.value).toBe(false)
  sidebar.resizeWithKey('End')
  expect(sidebar.open.value).toBe(true)
  expect(sidebar.width.value).toBe(400)
  sidebar.resizeWithKey('ArrowRight')
  expect(sidebar.width.value).toBe(400)
})

it('mobile navigation and breakpoint changes preserve desktop preferences', () => {
  const { sidebar, mobile } = setup()
  sidebar.resizeWithKey('End')
  sidebar.toggle()
  sidebar.setHovered(true)
  expect(sidebar.state.value).toBe('peeking')
  mobile.value = true
  expect(sidebar.state.value).toBe('collapsed')
  expect(sidebar.railWidth.value).toBe(0)
  expect(sidebar.panelWidth.value).toBe(400)
  sidebar.setHovered(true)
  sidebar.setFocused(true)
  expect(sidebar.state.value).toBe('collapsed')
  sidebar.toggle()
  expect(sidebar.mobileOpen.value).toBe(true)
  expect(sidebar.open.value).toBe(false)
  sidebar.beginResize(0)
  sidebar.resizeTo(250)
  expect(sidebar.isResizing.value).toBe(false)
  expect(sidebar.resizeWithKey('ArrowRight')).toBe(false)
  expect(sidebar.width.value).toBe(400)
  sidebar.closeMobile()
  expect(sidebar.mobileOpen.value).toBe(false)
  sidebar.toggle()
  mobile.value = false
  expect(sidebar.mobileOpen.value).toBe(false)
  expect(sidebar.state.value).toBe('collapsed')
  sidebar.toggle()
  expect(sidebar.railWidth.value).toBe(400)
})

it('breakpoint changes cancel an active drag', () => {
  const { sidebar, mobile } = setup()
  sidebar.beginResize(256)
  sidebar.resizeTo(300)
  mobile.value = true
  expect(sidebar.isResizing.value).toBe(false)
  sidebar.resizeTo(350)
  expect(sidebar.width.value).toBe(300)
  mobile.value = false
  expect(sidebar.state.value).toBe('expanded')
})

it('disposing the component cancels an unfinished drag', () => {
  const { sidebar, scope } = setup()
  sidebar.beginResize(256)
  expect(sidebar.isResizing.value).toBe(true)
  scope.stop()
  expect(sidebar.isResizing.value).toBe(false)
  sidebar.resizeTo(350)
  expect(sidebar.width.value).toBe(256)
})

it.each([
  ['NaN', 256],
  ['Infinity', 256],
  ['-50', 200],
  ['1000', 400],
  ['280.6', 281],
])('repairs persisted width %s to %i', async (storedWidth, expectedWidth) => {
  storage.set('octopulse.sidebar.width', storedWidth)
  const { sidebar } = setup()
  expect(sidebar.width.value).toBe(expectedWidth)
  expect(sidebar.panelWidth.value).toBe(expectedWidth)
  await nextTick()
  expect(storage.get('octopulse.sidebar.width')).toBe(String(expectedWidth))
})
