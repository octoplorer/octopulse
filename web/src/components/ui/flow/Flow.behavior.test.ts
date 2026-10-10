// @vitest-environment happy-dom
import { mount } from '@vue/test-utils'
import { afterEach, expect, it, vi } from 'vitest'
import { defineComponent, h, nextTick, shallowRef } from 'vue'
import FlowNode from './FlowNode.vue'
import FlowRoot from './FlowRoot.vue'

const mounted: ReturnType<typeof mount>[] = []
afterEach(() => {
  mounted.splice(0).forEach(wrapper => wrapper.unmount())
  vi.restoreAllMocks()
})

it('repositions keyed nodes when the consumer reorders its workflow', async () => {
  const order = shallowRef(['first', 'second'])
  const wrapper = mount(FlowRoot, { slots: { default: () => order.value.map(id => h(FlowNode, { id, key: id }, () => id)) } })
  mounted.push(wrapper)
  await nextTick()
  const first = wrapper.get('[data-node-id="first"]').element as HTMLElement
  const second = wrapper.get('[data-node-id="second"]').element as HTMLElement
  expect(Number.parseFloat(first.style.left)).toBeLessThan(Number.parseFloat(second.style.left))
  order.value = ['second', 'first']
  await nextTick()
  expect(Number.parseFloat(first.style.left)).toBeGreaterThan(Number.parseFloat(second.style.left))
})

it('measures the actual root of a custom Vue node when laying out the following node', async () => {
  const Inner = defineComponent({ setup: (_, { slots }) => () => h('article', slots.default?.()) })
  const CustomNode = defineComponent({ setup: (_, { slots }) => () => h(Inner, null, slots) })
  vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockImplementation(function (this: HTMLElement) {
    const width = this.dataset.nodeId === 'custom' ? 240 : 160
    return { x: 0, y: 0, left: 0, top: 0, width, height: 48, right: width, bottom: 48, toJSON: () => ({}) }
  })
  const wrapper = mount(FlowRoot, { props: { columnGap: 64 }, slots: { default: () => [
    h(FlowNode, { as: CustomNode, id: 'custom' }, () => 'Custom node'),
    h(FlowNode, { id: 'next' }, () => 'Next node'),
  ] } })
  mounted.push(wrapper)
  await nextTick()
  expect((wrapper.get('[data-node-id="next"]').element as HTMLElement).style.left).toBe('304px')
})
