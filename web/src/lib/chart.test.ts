import type { BarSeriesOption, DefaultLabelFormatterCallbackParams, LineSeriesOption, TooltipComponentOption, XAxisComponentOption, YAxisComponentOption } from 'echarts'
import { describe, expect, it } from 'vitest'
import { dailyAvailabilityOption, normalizeTimeSeries, timeSeriesOption } from './chart.ts'

const defaults = {
  name: 'Latency',
  unit: 'ms',
  valueFormatter: (value: number) => value.toLocaleString('en-US'),
  timeFormatter: (at: number) => `Time ${at}`,
}

function firstSeries(option: ReturnType<typeof timeSeriesOption>) {
  const series = option.series
  if (!Array.isArray(series) || !series[0] || series[0].type !== 'line')
    throw new Error('Expected a line series')
  return series[0] as LineSeriesOption
}

function firstBarSeries(option: ReturnType<typeof timeSeriesOption>) {
  const series = option.series
  if (!Array.isArray(series) || !series[0] || series[0].type !== 'bar')
    throw new Error('Expected a bar series')
  return series[0] as BarSeriesOption
}

describe('normalizeTimeSeries', () => {
  it('filters invalid values and timestamps, sorts actual observation times, and preserves duplicates', () => {
    expect(normalizeTimeSeries([
      { at: 20, value: 2 },
      { at: Number.NaN, value: 3 },
      { at: 10, value: -1 },
      { at: 20, value: 4 },
      { at: 30, value: Number.POSITIVE_INFINITY },
      { at: Number.NEGATIVE_INFINITY, value: 5 },
      { at: 8640000000000001, value: 6 },
      { at: 0, value: 0 },
    ])).toStrictEqual([
      { at: 0, value: 0 },
      { at: 10, value: -1 },
      { at: 20, value: 2 },
      { at: 20, value: 4 },
    ])
  })

  it('returns independent observations without mutating incoming data', () => {
    const input = [{ at: 20, value: 2 }, { at: 10, value: 1 }]
    const result = normalizeTimeSeries(input)
    result[0]!.value = 99
    expect(input).toStrictEqual([{ at: 20, value: 2 }, { at: 10, value: 1 }])
  })
})

describe('timeSeriesOption', () => {
  it('keeps actual timestamps and breaks gaps beyond three median positive intervals', () => {
    const option = timeSeriesOption({
      ...defaults,
      points: [0, 10, 20, 60, 70, 80, 120].map(at => ({ at, value: at })),
    })
    const line = firstSeries(option)
    expect(line.data).toStrictEqual([
      [0, 0],
      [10, 10],
      [20, 20],
      [40, null],
      [60, 60],
      [70, 70],
      [80, 80],
      [100, null],
      [120, 120],
    ])
    expect(line.connectNulls).toBe(false)
    expect((option.xAxis as XAxisComponentOption).type).toBe('time')
    expect((option.xAxis as XAxisComponentOption).min).toBe(0)
    expect((option.xAxis as XAxisComponentOption).max).toBe(120)
    expect(typeof line.symbolSize).toBe('function')
    if (typeof line.symbolSize === 'function') {
      expect(line.symbolSize([120, 120], { dataIndex: 8 } as DefaultLabelFormatterCallbackParams)).toBe(4)
      expect(line.symbolSize([10, 10], { dataIndex: 1 } as DefaultLabelFormatterCallbackParams)).toBe(0)
    }
  })

  it('ignores duplicate timestamps in the median and keeps the threshold interval connected', () => {
    const line = firstSeries(timeSeriesOption({
      ...defaults,
      points: [0, 0, 10, 20, 50].map(at => ({ at, value: 1 })),
    }))
    expect(line.data).toStrictEqual([[0, 1], [0, 1], [10, 1], [20, 1], [50, 1]])
  })

  it('renders no series for empty or entirely invalid observations', () => {
    expect(timeSeriesOption({ ...defaults, points: [] }).series).toStrictEqual([])
    expect(timeSeriesOption({ ...defaults, points: [{ at: 1, value: Number.NaN }] }).series).toStrictEqual([])
  })

  it('keeps a single zero observation visible and centers it on the time axis', () => {
    const option = timeSeriesOption({ ...defaults, points: [{ at: 10, value: 0 }] })
    const line = firstSeries(option)
    expect(line.data).toStrictEqual([[10, 0]])
    expect((option.xAxis as XAxisComponentOption).min).toBe(9)
    expect((option.xAxis as XAxisComponentOption).max).toBe(11)
    expect((option.yAxis as YAxisComponentOption).min).toBe(0)
    expect((option.yAxis as YAxisComponentOption).max).toBe(1)
    if (typeof line.symbolSize === 'function')
      expect(line.symbolSize([10, 0], { dataIndex: 0 } as DefaultLabelFormatterCallbackParams)).toBe(4)
  })

  it('includes negative observations and preserves the zero-to-one baseline', () => {
    const option = timeSeriesOption({ ...defaults, points: [{ at: 1, value: -5 }, { at: 2, value: -2 }] })
    expect((option.yAxis as YAxisComponentOption).min).toBe(-5)
    expect((option.yAxis as YAxisComponentOption).max).toBe(1)
    expect(firstSeries(option).data).toStrictEqual([[1, -5], [2, -2]])
  })

  it('computes the range for long histories without spreading observation arguments', () => {
    const points = Array.from({ length: 150000 }, (_, at) => ({ at, value: at - 75000 }))
    const option = timeSeriesOption({ ...defaults, points })
    expect((option.yAxis as YAxisComponentOption).min).toBe(-75000)
    expect((option.yAxis as YAxisComponentOption).max).toBe(74999)
  })

  it('rebuilds replaced observations without changing the preceding option or source', () => {
    const points = [{ at: 20, value: 2 }, { at: 10, value: 1 }]
    const before = timeSeriesOption({ ...defaults, points })
    points[0]!.value = 3
    const after = timeSeriesOption({ ...defaults, points })
    expect(firstSeries(before).data).toStrictEqual([[10, 1], [20, 2]])
    expect(firstSeries(after).data).toStrictEqual([[10, 1], [20, 3]])
    expect(points).toStrictEqual([{ at: 20, value: 3 }, { at: 10, value: 1 }])
  })

  it('uses supplied number and time formatting in literal canvas tooltips', () => {
    const option = timeSeriesOption({
      ...defaults,
      name: '<b>{danger|Latency}</b>',
      points: [{ at: 100, value: 1234 }],
    })
    const tooltip = option.tooltip as TooltipComponentOption
    expect(tooltip.renderMode).toBe('richText')
    if (typeof tooltip.formatter !== 'function')
      throw new Error('Expected a tooltip formatter')
    expect(tooltip.formatter([{ value: [100, 1234] } as DefaultLabelFormatterCallbackParams], '')).toBe(
      'Time 100\n<b>｛danger|Latency｝</b>: 1,234 ms',
    )
    expect(tooltip.formatter([{ value: [100, null] } as DefaultLabelFormatterCallbackParams], '')).toBe('')
    expect(tooltip.formatter([], '')).toBe('')
  })

  it('hides axes for compact charts and forwards number formatting to full y-axis labels', () => {
    const compact = timeSeriesOption({ ...defaults, compact: true, points: [{ at: 1, value: 1 }] })
    expect((compact.xAxis as XAxisComponentOption).show).toBe(false)
    expect((compact.yAxis as YAxisComponentOption).show).toBe(false)
    expect((compact.tooltip as TooltipComponentOption).show).toBe(false)
    const full = timeSeriesOption({ ...defaults, points: [{ at: 1, value: 1 }] })
    const axis = full.yAxis
    if (!axis || Array.isArray(axis) || axis.type !== 'value')
      throw new Error('Expected a value axis')
    const formatter = axis.axisLabel?.formatter
    if (typeof formatter !== 'function')
      throw new Error('Expected a y-axis formatter')
    expect(formatter(1234, 0, undefined)).toBe('1,234')
  })

  it('renders rounded bars for observations and leaves missing periods empty', () => {
    const points = [0, 10, 20, 60, 70, 80, 120].map(at => ({ at, value: at }))
    const option = timeSeriesOption({ ...defaults, type: 'bar', compact: true, points })
    const bars = firstBarSeries(option)
    expect(bars.data).toStrictEqual(points.map(point => [point.at, point.value]))
    expect(bars.barMaxWidth).toBe(8)
    expect(bars.barMinHeight).toBe(2)
    expect(bars.itemStyle).toMatchObject({ color: 'var(--color-brand)', borderRadius: 2 })
    expect(bars.clip).toBe(false)
    expect((option.xAxis as XAxisComponentOption).min).toBe(-5)
    expect((option.xAxis as XAxisComponentOption).max).toBe(125)
    expect((option.xAxis as XAxisComponentOption).show).toBe(false)
    expect((option.yAxis as YAxisComponentOption).show).toBe(false)
    expect((option.tooltip as TooltipComponentOption).show).toBe(false)
  })

  it('centers a single zero bar with nonzero height and filters invalid observations', () => {
    const points = [{ at: 10, value: 0 }, { at: 20, value: Number.NaN }]
    const option = timeSeriesOption({ ...defaults, type: 'bar', points })
    const bars = firstBarSeries(option)
    expect(bars.data).toStrictEqual([[10, 0]])
    expect(bars.barMinHeight).toBe(2)
    expect((option.xAxis as XAxisComponentOption).min).toBe(9)
    expect((option.xAxis as XAxisComponentOption).max).toBe(11)
    expect(points).toStrictEqual([{ at: 10, value: 0 }, { at: 20, value: Number.NaN }])
    expect(timeSeriesOption({ ...defaults, type: 'bar', points: [] }).series).toStrictEqual([])
  })

  it('keeps dense bar histories independent without synthesizing extra observations', () => {
    const points = Array.from({ length: 10000 }, (_, at) => ({ at, value: at % 10 }))
    points[5000]!.at = 12000
    const option = timeSeriesOption({ ...defaults, type: 'bar', points })
    const bars = firstBarSeries(option)
    expect(bars.data).toHaveLength(points.length)
    expect(bars.data?.at(-1)).toStrictEqual([12000, 0])
    points[0]!.value = 99
    expect(bars.data?.[0]).toStrictEqual([0, 0])
    expect(points[5000]!.at).toBe(12000)
    expect(bars.barWidth).toBeUndefined()
  })
})

describe('dailyAvailabilityOption', () => {
  const day = 86400000
  const availabilityDefaults = {
    name: 'Uptime',
    dayFormatter: (from: number) => new Date(from).toISOString().slice(0, 10),
    valueFormatter: (percentage: number) => `${percentage.toLocaleString('en-US')}%`,
    noDataLabel: 'No observations',
    coverageLabel: 'Coverage',
  }

  function tooltipAt(option: ReturnType<typeof dailyAvailabilityOption>, dataIndex: number) {
    const tooltip = option.tooltip as TooltipComponentOption
    if (typeof tooltip.formatter !== 'function')
      throw new Error('Expected a tooltip formatter')
    return tooltip.formatter([{ dataIndex } as DefaultLabelFormatterCallbackParams], '')
  }

  it('uses 90 equally spaced daily categories with a fixed percentage scale and rounded bars', () => {
    const points = Array.from({ length: 90 }, (_, index) => ({
      from: (89 - index) * day,
      to: (90 - index) * day,
      uptime: 100,
      coverage: 100,
    }))
    const option = dailyAvailabilityOption({ ...availabilityDefaults, points })
    const bars = firstBarSeries(option)
    const axis = option.xAxis as XAxisComponentOption
    expect(axis.type).toBe('category')
    expect(axis.boundaryGap).toBe(true)
    if (axis.type !== 'category')
      throw new Error('Expected a category axis')
    expect(axis.data).toStrictEqual(Array.from({ length: 90 }, (_, index) => String(index * day)))
    expect(bars.data).toHaveLength(90)
    expect(bars.barMaxWidth).toBe(8)
    expect(bars.barMinHeight).toBe(2)
    expect(bars.itemStyle?.borderRadius).toBe(2)
    expect(option.yAxis).toMatchObject({ type: 'value', min: 0, max: 100, show: false })
    expect(points[0]!.from).toBe(89 * day)
  })

  it('keeps unobserved days null while distinguishing real zero and partial downtime by color', () => {
    const points = [null, 0, 1, 99.5, 100].map((uptime, index) => ({
      from: index * day,
      to: (index + 1) * day,
      uptime,
      coverage: uptime === null ? 0 : 100,
    }))
    const option = dailyAvailabilityOption({ ...availabilityDefaults, points, color: '#008877' })
    const bars = firstBarSeries(option)
    expect(bars.data).toStrictEqual([
      { value: null, itemStyle: { color: 'var(--color-fill)' } },
      { value: 0, itemStyle: { color: 'var(--color-danger)' } },
      { value: 1, itemStyle: { color: 'var(--color-warning)' } },
      { value: 99.5, itemStyle: { color: 'var(--color-warning)' } },
      { value: 100, itemStyle: { color: '#008877' } },
    ])
    expect(bars.showBackground).toBe(true)
    expect(bars.backgroundStyle).toMatchObject({ color: 'var(--color-fill)', borderRadius: 2 })
    expect(tooltipAt(option, 0)).toBe('1970-01-01\nUptime: No observations · Coverage: 0%')
    expect(tooltipAt(option, 1)).toBe('1970-01-02\nUptime: 0% · Coverage: 100%')
    expect(tooltipAt(option, 2)).toBe('1970-01-03\nUptime: 1% · Coverage: 100%')
  })

  it('keeps the unobserved today slot at exactly UTC midnight', () => {
    const points = Array.from({ length: 90 }, (_, index) => ({
      from: index * day,
      to: index === 89 ? index * day : (index + 1) * day,
      uptime: index === 89 ? null : 100,
      coverage: index === 89 ? null : 100,
    }))
    const option = dailyAvailabilityOption({ ...availabilityDefaults, points })
    expect(firstBarSeries(option).data).toHaveLength(90)
    expect(tooltipAt(option, 89)).toBe('1970-03-31\nUptime: No observations')
  })

  it('does not invent missing daily buckets or reinterpret invalid percentages as availability', () => {
    const points = [
      { from: 0, to: day, uptime: -1, coverage: Number.NaN },
      { from: day * 3, to: day * 4, uptime: 101, coverage: 101 },
      { from: Number.NaN, to: day * 5, uptime: 100, coverage: 100 },
    ]
    const option = dailyAvailabilityOption({ ...availabilityDefaults, points })
    expect(firstBarSeries(option).data).toStrictEqual([
      { value: null, itemStyle: { color: 'var(--color-fill)' } },
      { value: null, itemStyle: { color: 'var(--color-fill)' } },
    ])
    expect(tooltipAt(option, 0)).toBe('1970-01-01\nUptime: No observations')
    expect(tooltipAt(option, 1)).toBe('1970-01-04\nUptime: No observations')
    expect(tooltipAt(option, 99)).toBe('')
    expect(dailyAvailabilityOption({ ...availabilityDefaults, points: [] }).series).toStrictEqual([])
  })

  it('keeps null tooltip text literal without calling percentage formatting and preserves source data', () => {
    const formatted: number[] = []
    const points = [{ from: 0, to: day, uptime: null, coverage: null }]
    const option = dailyAvailabilityOption({
      ...availabilityDefaults,
      points,
      name: '<b>{style|Uptime}</b>',
      noDataLabel: '{danger|No data}',
      valueFormatter: (percentage) => {
        formatted.push(percentage)
        return `${percentage}%`
      },
    })
    expect(tooltipAt(option, 0)).toBe('1970-01-01\n<b>｛style|Uptime｝</b>: ｛danger|No data｝')
    expect(formatted).toStrictEqual([])
    expect(points).toStrictEqual([{ from: 0, to: day, uptime: null, coverage: null }])
    const tooltip = option.tooltip as TooltipComponentOption
    expect(tooltip.renderMode).toBe('richText')
    expect(tooltip.confine).toBe(true)
    expect(tooltip.textStyle).toMatchObject({ fontSize: 11, lineHeight: 14 })
  })
})
