import type { BarSeriesOption, EChartsOption, LineSeriesOption } from 'echarts'

export interface TimeSeriesPoint {
  at: number
  value: number
}

export interface TimeSeriesOptions {
  points: readonly TimeSeriesPoint[]
  name: string
  unit?: string
  compact?: boolean
  color?: string
  type?: 'line' | 'bar'
  valueFormatter: (value: number) => string
  timeFormatter: (at: number) => string
}

export interface DailyAvailabilityPoint {
  from: number
  to: number
  uptime: number | null
  coverage: number | null
}

export interface DailyAvailabilityOptions {
  points: readonly DailyAvailabilityPoint[]
  name: string
  color?: string
  dayFormatter: (from: number) => string
  // Availability values are percentages in the range 0–100.
  valueFormatter: (percentage: number) => string
  noDataLabel: string
  coverageLabel: string
}

export function normalizeTimeSeries(points: readonly TimeSeriesPoint[]): TimeSeriesPoint[] {
  return points
    .filter(point => Number.isFinite(point.value) && Number.isFinite(point.at)
      && Math.abs(point.at) <= 8640000000000000)
    .map(point => ({ at: point.at, value: point.value }))
    .sort((a, b) => a.at - b.at)
}

function tooltipText(text: string) {
  // ECharts parses brace fragments as rich-text styles, even in canvas tooltips.
  return text.replaceAll('{', '｛').replaceAll('}', '｝')
}

export function timeSeriesOption({
  points,
  name,
  unit = '',
  compact = false,
  color = 'var(--color-brand)',
  type = 'line',
  valueFormatter,
  timeFormatter,
}: TimeSeriesOptions): EChartsOption {
  const observations = normalizeTimeSeries(points)
  const gaps = observations.slice(1)
    .map((point, index) => point.at - observations[index]!.at)
    .filter(gap => gap > 0)
    .sort((a, b) => a - b)
  const gapLimit = gaps.length ? gaps[Math.floor(gaps.length / 2)]! * 3 : Infinity
  const data: [number, number | null][] = []
  for (const [index, point] of observations.entries()) {
    const previous = observations[index - 1]
    if (type === 'line' && previous && point.at - previous.at > gapLimit)
      data.push([previous.at + (point.at - previous.at) / 2, null])
    data.push([point.at, point.value])
  }
  const isolated = new Set(data.flatMap((point, index) => {
    const previous = data[index - 1]
    const next = data[index + 1]
    return point[1] !== null && (!previous || previous[1] === null)
      && (!next || next[1] === null)
      ? [index]
      : []
  }))
  const range = observations.reduce((range, point) => ({
    min: Math.min(range.min, point.value),
    max: Math.max(range.max, point.value),
  }), { min: 0, max: 1 })
  const first = observations[0]?.at
  const last = observations.at(-1)?.at
  const singleTime = first !== undefined && first === last
  const padding = type === 'bar' ? (gaps[0] ?? 2) / 2 : singleTime ? 1 : 0
  const series: LineSeriesOption | BarSeriesOption = type === 'bar'
    ? {
        type: 'bar',
        name,
        data,
        barMaxWidth: 8,
        barMinHeight: 2,
        clip: false,
        itemStyle: { color, borderRadius: 2 },
      }
    : {
        type: 'line',
        name,
        data,
        connectNulls: false,
        smooth: false,
        clip: false,
        showSymbol: true,
        symbol: 'circle',
        symbolSize: (_value, params) => isolated.has(params.dataIndex) ? 4 : 0,
        lineStyle: { color, width: 2.5, cap: 'round', join: 'round' },
        itemStyle: { color },
        emphasis: { scale: false },
      }

  return {
    grid: {
      left: compact ? type === 'bar' ? 6 : 2 : 0,
      right: compact ? type === 'bar' ? 6 : 2 : 12,
      top: compact ? 4 : 8,
      bottom: compact ? 4 : 0,
      containLabel: !compact,
    },
    xAxis: {
      type: 'time',
      show: !compact,
      min: first === undefined ? undefined : first - padding,
      max: last === undefined ? undefined : last + padding,
      boundaryGap: [0, 0],
      splitNumber: 2,
      axisLine: { show: false },
      axisTick: { show: false },
      splitLine: { show: false },
      axisLabel: {
        hideOverlap: true,
        fontSize: 10,
        color: 'var(--text-color-subtle)',
        formatter: value => timeFormatter(Number(value)),
      },
    },
    yAxis: {
      type: 'value',
      show: !compact,
      min: range.min,
      max: range.max,
      splitNumber: 2,
      axisLine: { show: false },
      axisTick: { show: false },
      axisLabel: {
        fontSize: 10,
        color: 'var(--text-color-subtle)',
        formatter: value => valueFormatter(Number(value)),
      },
      splitLine: { show: !compact, lineStyle: { color: 'var(--color-hairline)' } },
    },
    tooltip: {
      show: !compact,
      trigger: 'axis',
      renderMode: 'richText',
      confine: true,
      axisPointer: { type: 'line', lineStyle: { color: 'var(--color-line)', type: 'dashed' } },
      formatter: (params) => {
        const point = Array.isArray(params) ? params[0] : params
        const value = point?.value
        if (!Array.isArray(value) || typeof value[0] !== 'number' || typeof value[1] !== 'number'
          || !Number.isFinite(value[0]) || !Number.isFinite(value[1])) {
          return ''
        }
        const metric = `${name}: ${valueFormatter(value[1])}${unit ? ` ${unit}` : ''}`
        return tooltipText(`${timeFormatter(value[0])}\n${metric}`)
      },
    },
    series: observations.length ? [series] : [],
  }
}

function availabilityPercentage(value: number | null) {
  return value !== null && Number.isFinite(value) && value >= 0 && value <= 100 ? value : null
}

export function dailyAvailabilityOption({
  points,
  name,
  color = 'var(--color-brand)',
  dayFormatter,
  valueFormatter,
  noDataLabel,
  coverageLabel,
}: DailyAvailabilityOptions): EChartsOption {
  const days = points
    .filter(point => Number.isFinite(point.from) && Number.isFinite(point.to)
      && Math.abs(point.from) <= 8640000000000000 && Math.abs(point.to) <= 8640000000000000
      && point.to >= point.from)
    .map(point => ({
      from: point.from,
      to: point.to,
      uptime: availabilityPercentage(point.uptime),
      coverage: availabilityPercentage(point.coverage),
    }))
    .sort((a, b) => a.from - b.from)
  const series: BarSeriesOption = {
    type: 'bar',
    name,
    data: days.map(point => ({
      value: point.uptime,
      itemStyle: {
        color: point.uptime === null
          ? 'var(--color-fill)'
          : point.uptime === 0
            ? 'var(--color-danger)'
            : point.uptime < 100 ? 'var(--color-warning)' : color,
      },
    })),
    barMaxWidth: 8,
    barMinHeight: 2,
    showBackground: true,
    backgroundStyle: { color: 'var(--color-fill)', borderRadius: 2 },
    itemStyle: { borderRadius: 2 },
  }

  return {
    grid: { left: 0, right: 0, top: 2, bottom: 2 },
    xAxis: {
      type: 'category',
      data: days.map(point => String(point.from)),
      boundaryGap: true,
      show: false,
    },
    yAxis: { type: 'value', min: 0, max: 100, show: false },
    tooltip: {
      trigger: 'axis',
      renderMode: 'richText',
      confine: true,
      padding: [4, 6],
      textStyle: { fontSize: 11, lineHeight: 14 },
      position: (point, _params, _dom, _rect, size) => [
        Math.max(0, Math.min(point[0] + 12, size.viewSize[0] - size.contentSize[0])),
        0,
      ],
      axisPointer: { type: 'none' },
      formatter: (params) => {
        const parameter = Array.isArray(params) ? params[0] : params
        const point = parameter ? days[parameter.dataIndex] : undefined
        if (!point)
          return ''
        const uptime = point.uptime === null ? noDataLabel : valueFormatter(point.uptime)
        const coverage = point.coverage === null ? '' : ` · ${coverageLabel}: ${valueFormatter(point.coverage)}`
        return tooltipText(`${dayFormatter(point.from)}\n${name}: ${uptime}${coverage}`)
      },
    },
    series: days.length ? [series] : [],
  }
}
