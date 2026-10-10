// @vitest-environment happy-dom
import { mount } from '@vue/test-utils'
import { expect, it } from 'vitest'
import { defineComponent, h, nextTick } from 'vue'
import { Badge } from '../components/ui/badge'
import { locale } from '../composables/i18n'
import { monitorStateDisplay } from './monitor-state'

function renderState(props: { state?: string, paused?: boolean, maintenance?: boolean }) {
  return mount(defineComponent({
    setup: () => () => {
      const display = monitorStateDisplay(props.state, props)
      return h(Badge, { 'variant': display.variant, 'data-state': display.state, 'dot': true }, () => display.label)
    },
  }))
}

it('unrecognized monitor states cannot resolve inherited object properties', () => {
  const wrapper = renderState({ state: 'constructor' })
  expect(wrapper.text()).toBe('constructor')
  expect(wrapper.attributes('data-state')).toBe('constructor')
  expect(wrapper.classes()).toContain('bg-recessed')
  wrapper.unmount()
})

it('paused monitors take precedence over maintenance and failed checks', () => {
  const originalLocale = locale.value
  locale.value = 'en'
  try {
    const wrapper = renderState({ state: 'down', paused: true, maintenance: true })
    expect(wrapper.text()).toBe('Paused')
    expect(wrapper.attributes('data-state')).toBe('paused')
    expect(wrapper.classes()).toContain('bg-recessed')
    wrapper.unmount()
  }
  finally {
    locale.value = originalLocale
  }
})

it('maintenance overrides the monitor check result without a paused flag', () => {
  const wrapper = renderState({ state: 'down', maintenance: true })
  expect(wrapper.attributes('data-state')).toBe('maintenance')
  expect(wrapper.classes()).toContain('bg-info-tint')
  expect(wrapper.find('[aria-hidden="true"]').classes()).toContain('bg-info')
  wrapper.unmount()
})

it.each([
  { state: 'UP', normalized: 'up', background: 'bg-success-tint', dot: 'bg-success' },
  { state: 'expired', normalized: 'expired', background: 'bg-danger-tint', dot: 'bg-danger' },
  { state: 'expiring', normalized: 'expiring', background: 'bg-warning-tint', dot: 'bg-warning' },
  { state: 'check_failed', normalized: 'check_failed', background: 'bg-recessed', dot: 'bg-subtle' },
])('monitor state $state keeps its severity and dot', ({ state, normalized, background, dot }) => {
  const wrapper = renderState({ state })
  expect(wrapper.attributes('data-state')).toBe(normalized)
  expect(wrapper.classes()).toContain(background)
  expect(wrapper.find('[aria-hidden="true"]').classes()).toContain(dot)
  wrapper.unmount()
})

it('missing states are localized while new API states keep their original text', async () => {
  const originalLocale = locale.value
  locale.value = 'en'
  const missing = renderState({})
  const empty = renderState({ state: '' })
  const unrecognized = renderState({ state: 'New_API_State' })
  try {
    expect(missing.text()).toBe('Unknown')
    expect(empty.text()).toBe('Unknown')
    expect(missing.attributes('data-state')).toBe('unknown')
    expect(unrecognized.text()).toBe('New_API_State')
    expect(unrecognized.attributes('data-state')).toBe('new_api_state')
    locale.value = 'zh-CN'
    await nextTick()
    expect(missing.text()).toBe('等待数据')
    expect(empty.text()).toBe('等待数据')
    expect(unrecognized.text()).toBe('New_API_State')
  }
  finally {
    missing.unmount()
    empty.unmount()
    unrecognized.unmount()
    locale.value = originalLocale
  }
})
