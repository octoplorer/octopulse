import type { EChartsOption } from 'echarts'
import { expect, it } from 'vitest'
import { defaultChartAppearance, prepareChartOption, resolveChartColors } from './echarts.ts'

it('prepares theme defaults without mutating the native option or its callbacks', () => {
  const formatter = () => '24 ms'
  const option: EChartsOption = {
    xAxis: { type: 'time', axisLabel: { formatter } },
    yAxis: { type: 'value', axisLine: { show: false } },
    series: [{ type: 'line', data: [[1000, 24], [2000, null]], connectNulls: false }],
  }
  const original = structuredClone({ ...option, xAxis: undefined })
  const prepared = prepareChartOption(option, defaultChartAppearance, undefined, 'Response time')
  expect(prepared.backgroundColor).toBe('transparent')
  expect(prepared.color).toEqual(defaultChartAppearance.palette)
  expect(prepared.xAxis).toMatchObject({ axisLabel: { formatter, color: defaultChartAppearance.subtle } })
  expect(prepared.yAxis).toMatchObject({ axisLine: { show: false, lineStyle: { color: defaultChartAppearance.line } } })
  expect(prepared.aria).toMatchObject({ enabled: true, label: { description: 'Response time' } })
  expect(option.xAxis).toEqual({ type: 'time', axisLabel: { formatter } })
  expect({ ...option, xAxis: undefined }).toEqual(original)
  expect(prepared.series).toEqual(option.series)
  expect(prepared.series).not.toBe(option.series)
})

it('renders tooltip formatters as canvas text at every level', () => {
  const formatter = () => '<img src=x onerror=alert(1)>'
  const seriesTooltip = { renderMode: 'html', formatter }
  const option: EChartsOption = {
    tooltip: [{ renderMode: 'html', formatter, textStyle: { fontSize: 12 } }],
    series: [{ type: 'line', tooltip: seriesTooltip, data: [24] }],
  }
  const prepared = prepareChartOption(option, defaultChartAppearance)
  expect(prepared.tooltip).toMatchObject([{ renderMode: 'richText', formatter, textStyle: { fontSize: 12, color: defaultChartAppearance.text } }])
  expect(prepared.series).toMatchObject([{ tooltip: { renderMode: 'richText', formatter } }])
  expect(option.tooltip).toMatchObject([{ renderMode: 'html' }])
  expect(option.series).toMatchObject([{ tooltip: { renderMode: 'html' } }])
})

it('disables even explicitly enabled series animations for reduced motion', () => {
  const option: EChartsOption = { animation: true, series: { type: 'bar', animation: true, data: [1] } }
  const prepared = prepareChartOption(option, { ...defaultChartAppearance, reducedMotion: true })
  expect(prepared.animation).toBe(false)
  expect(prepared.series).toMatchObject({ animation: false })
  expect(option.animation).toBe(true)
  expect(option.series).toMatchObject({ animation: true })
})

it('resolves nested CSS colors and gradient stops while preserving data and non-plain values', () => {
  const timestamps = new Float64Array([1000, 2000])
  const option = {
    color: ['var(--color-brand)', 'oklch(0.6 0.2 250)'],
    series: [{
      name: 'var(--name-is-data)',
      data: timestamps,
      areaStyle: { color: { type: 'linear', colorStops: [{ offset: 0, color: 'color-mix(in oklch, red, blue)' }] } },
    }],
  }
  const resolve = (color: string) => `resolved:${color}`
  const prepared = resolveChartColors(option, resolve)
  expect(prepared.color).toEqual(['resolved:var(--color-brand)', 'resolved:oklch(0.6 0.2 250)'])
  expect(prepared.series[0]?.areaStyle.color.colorStops[0]?.color).toBe('resolved:color-mix(in oklch, red, blue)')
  expect(prepared.series[0]?.name).toBe('var(--name-is-data)')
  expect(prepared.series[0]?.data).toBe(timestamps)
  expect(option.color).toEqual(['var(--color-brand)', 'oklch(0.6 0.2 250)'])
})

it('reprepares the same native option with changed theme and local brand colors', () => {
  const option: EChartsOption = {
    series: [{ type: 'line', lineStyle: { color: 'var(--color-brand)' }, data: [24] }],
  }
  const light = prepareChartOption(option, defaultChartAppearance, color => color === 'var(--color-brand)' ? '#2563eb' : color)
  const dark = prepareChartOption(
    option,
    { ...defaultChartAppearance, palette: ['#b91c1c'], text: '#fafafa', dark: true },
    color => color === 'var(--color-brand)' ? '#b91c1c' : color,
  )
  expect(light.series).toMatchObject([{ lineStyle: { color: '#2563eb' } }])
  expect(dark.series).toMatchObject([{ lineStyle: { color: '#b91c1c' } }])
  expect(dark.color).toEqual(['#b91c1c'])
  expect(dark.textStyle?.color).toBe('#fafafa')
  expect(option.series).toMatchObject([{ lineStyle: { color: 'var(--color-brand)' } }])
})
