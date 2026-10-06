import type { EChartsOption } from 'echarts'
import { BarChart, LineChart } from 'echarts/charts'
import {
  AriaComponent,
  GridComponent,
  LegendComponent,
  TooltipComponent,
} from 'echarts/components'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'

use([CanvasRenderer, LineChart, BarChart, GridComponent, TooltipComponent, LegendComponent, AriaComponent])

export interface ChartAppearance {
  palette: string[]
  text: string
  subtle: string
  line: string
  surface: string
  fontFamily: string
  dark: boolean
  reducedMotion: boolean
}

export const defaultChartAppearance: ChartAppearance = {
  palette: ['#2563eb', '#059669', '#d97706', '#dc2626', '#0891b2'],
  text: '#171717',
  subtle: '#737373',
  line: 'rgba(0, 0, 0, 0.1)',
  surface: '#ffffff',
  fontFamily: 'sans-serif',
  dark: false,
  reducedMotion: false,
}

type ColorResolver = (color: string) => string
type ChartAxis = Exclude<NonNullable<EChartsOption['xAxis'] | EChartsOption['yAxis']>, unknown[]>

const colorKeys = new Set([
  'color',
  'backgroundColor',
  'borderColor',
  'shadowColor',
  'textBorderColor',
  'textShadowColor',
  'areaColor',
  'fill',
  'stroke',
])

function isPlainObject(value: unknown): value is Record<string, unknown> {
  if (!value || typeof value !== 'object')
    return false
  const prototype = Object.getPrototypeOf(value)
  return prototype === Object.prototype || prototype === null
}

function mapOption(value: unknown, resolveColor: ColorResolver, key = '', appearance?: ChartAppearance): unknown {
  if (typeof value === 'string')
    return colorKeys.has(key) ? resolveColor(value) : value
  if (Array.isArray(value))
    return value.map(entry => mapOption(entry, resolveColor, key, appearance))
  if (!isPlainObject(value))
    return value
  const mapped = Object.fromEntries(
    Object.entries(value).map(([name, entry]) => [name, mapOption(entry, resolveColor, name, appearance)]),
  )
  if (key === 'tooltip' && appearance) {
    return {
      backgroundColor: appearance.surface,
      borderColor: appearance.line,
      ...mapped,
      textStyle: { color: appearance.text, fontFamily: appearance.fontFamily, ...isPlainObject(mapped.textStyle) ? mapped.textStyle : {} },
      renderMode: 'richText',
    }
  }
  return mapped
}

export function resolveChartColors<T>(value: T, resolveColor: ColorResolver): T {
  return mapOption(value, resolveColor) as T
}

function mapComponents<T>(value: T | T[] | undefined, prepare: (component: T) => T): T | T[] | undefined {
  if (Array.isArray(value))
    return value.map(prepare)
  return value ? prepare(value) : value
}

export function prepareChartOption(
  option: EChartsOption,
  appearance: ChartAppearance,
  resolveColor: ColorResolver = color => color,
  ariaLabel?: string,
): EChartsOption {
  const prepared = mapOption(option, resolveColor, '', appearance) as EChartsOption
  const axis = <T extends ChartAxis>(component: T): T => ({
    ...component,
    axisLabel: { color: appearance.subtle, ...component.axisLabel },
    axisLine: { ...component.axisLine, lineStyle: { color: appearance.line, ...component.axisLine?.lineStyle } },
    splitLine: { ...component.splitLine, lineStyle: { color: appearance.line, ...component.splitLine?.lineStyle } },
  })
  return {
    ...prepared,
    backgroundColor: prepared.backgroundColor ?? 'transparent',
    color: prepared.color ?? appearance.palette,
    textStyle: { color: appearance.text, fontFamily: appearance.fontFamily, ...prepared.textStyle },
    animation: appearance.reducedMotion ? false : prepared.animation,
    animationDuration: prepared.animationDuration ?? 300,
    animationDurationUpdate: prepared.animationDurationUpdate ?? 200,
    aria: {
      ...prepared.aria,
      enabled: prepared.aria?.enabled ?? true,
      label: { ...prepared.aria?.label, ...ariaLabel ? { description: ariaLabel } : {} },
    },
    tooltip: prepared.tooltip ?? { renderMode: 'richText', backgroundColor: appearance.surface, borderColor: appearance.line, textStyle: { color: appearance.text, fontFamily: appearance.fontFamily } },
    xAxis: mapComponents(prepared.xAxis, axis),
    yAxis: mapComponents(prepared.yAxis, axis),
    legend: mapComponents(prepared.legend, legend => ({ ...legend, textStyle: { color: appearance.text, fontFamily: appearance.fontFamily, ...legend.textStyle } })),
    series: appearance.reducedMotion
      ? mapComponents(prepared.series, series => ({ ...series, animation: false }))
      : prepared.series,
  }
}

// Resolve CSS in the chart's own context, including local status-page brand overrides.
// Sampling a pixel converts modern CSS colors to the RGB colors Canvas/ECharts accepts.
export function createChartColorResolver(root: HTMLElement): ColorResolver {
  const probe = root.ownerDocument.createElement('span')
  probe.style.cssText = 'position:absolute;visibility:hidden;pointer-events:none'
  const canvas = root.ownerDocument.createElement('canvas')
  canvas.width = canvas.height = 1
  const context = canvas.getContext('2d', { willReadFrequently: true })
  const cache = new Map<string, string>()
  return (color) => {
    const cached = cache.get(color)
    if (cached)
      return cached
    probe.style.color = ''
    probe.style.color = color
    if (!probe.style.color || !context)
      return color
    root.append(probe)
    const computed = getComputedStyle(probe).color
    probe.remove()
    context.clearRect(0, 0, 1, 1)
    context.fillStyle = computed
    context.fillRect(0, 0, 1, 1)
    const [red, green, blue, alpha] = context.getImageData(0, 0, 1, 1).data
    const resolved = `rgba(${red}, ${green}, ${blue}, ${Number((alpha! / 255).toFixed(3))})`
    cache.set(color, resolved)
    return resolved
  }
}

export function readChartAppearance(root: HTMLElement, reducedMotion: boolean): ChartAppearance {
  const resolve = createChartColorResolver(root)
  const style = getComputedStyle(root)
  return {
    palette: ['--color-brand', '--color-success', '--color-warning', '--color-danger', '--color-info'].map(token => resolve(`var(${token})`)),
    text: resolve('var(--text-color-default)'),
    subtle: resolve('var(--text-color-subtle)'),
    line: resolve('var(--color-line)'),
    surface: resolve('var(--color-base)'),
    fontFamily: style.fontFamily,
    dark: style.colorScheme === 'dark',
    reducedMotion,
  }
}
