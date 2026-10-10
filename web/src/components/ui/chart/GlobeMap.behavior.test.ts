// @vitest-environment happy-dom
import type { CustomSeriesOption } from 'echarts'
import { mount } from '@vue/test-utils'
import { afterEach, expect, it } from 'vitest'
import { defineComponent, h, nextTick } from 'vue'
import Chart from './Chart.vue'
import GlobeMap from './GlobeMap.vue'

// Canvas is unavailable in this DOM environment; test the component's chart event contract.
const ChartStub = defineComponent({
  name: 'Chart',
  props: ['option', 'height'],
  emits: ['click', 'mouseover', 'mouseout'],
  setup: () => () => h('div'),
})
const wrappers: ReturnType<typeof mount>[] = []
afterEach(() => wrappers.splice(0).forEach(wrapper => wrapper.unmount()))

it('routes Vue ECharts marker clicks to the original marker and ignores other chart layers', async () => {
  const marker = { longitude: 0, latitude: 0, name: 'London' }
  const wrapper = mount(GlobeMap, { props: { markers: [marker] }, global: { stubs: { Chart: ChartStub } } })
  wrappers.push(wrapper)
  const chart = wrapper.findComponent(Chart)
  expect(chart.exists()).toBe(true)
  chart.vm.$emit('click', { seriesId: 'globe-land', data: { datum: marker } })
  expect(wrapper.emitted('markerClick')).toBeUndefined()
  chart.vm.$emit('click', { seriesId: 'globe-markers', data: { datum: marker } })
  expect(wrapper.emitted('markerClick')).toEqual([[marker]])
})

it('updates Vue ECharts marker data after keyboard rotation', async () => {
  const wrapper = mount(GlobeMap, { props: { markers: [{ longitude: 88, latitude: 0, name: 'Horizon' }] }, global: { stubs: { Chart: ChartStub } } })
  wrappers.push(wrapper)
  const chart = wrapper.findComponent(Chart)
  expect(chart.exists()).toBe(true)
  expect((chart.props('option').series as CustomSeriesOption[])[3]?.data).toHaveLength(1)
  await wrapper.get('[role="group"]').trigger('keydown', { key: 'ArrowRight' })
  await nextTick()
  expect(wrapper.emitted('userRotationChange')).toEqual([[[5, 0, 0]]])
  expect((chart.props('option').series as CustomSeriesOption[])[3]?.data).toHaveLength(0)
})
