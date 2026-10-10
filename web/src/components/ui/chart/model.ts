import type { EChartsOption, LineSeriesOption, SankeySeriesOption } from 'echarts'
import type { BubbleMapProps, ChoroplethMapProps, MapAccessor, MapProjection, MapStyle, SankeyChartProps, SankeyNodeData, TimeseriesChartProps } from './types'
import { ChartPalette } from './palette'

// Rich-text tooltips interpret braces as styling commands; labels remain plain text.
export function chartText(value: unknown) {
  return String(value).replaceAll('{', '｛').replaceAll('}', '｝')
}

export function createTimeseriesOption(props: TimeseriesChartProps): EChartsOption {
  const type = props.type ?? 'line'
  const valueFormat = props.tooltipValueFormat ?? ((value: number) => value.toLocaleString())
  const timestampFormat = props.tooltipTimestampFormat ?? ((value: number) => new Date(value).toLocaleString())
  const annotations = [
    ...(props.thresholds ?? []).filter(item => Number.isFinite(item.value)).map(item => ({
      name: item.label ?? '',
      yAxis: item.value,
      lineStyle: { color: item.color, type: 'dashed' as const },
      label: { show: !!item.label, color: item.color, formatter: chartText(item.label ?? '') },
    })),
    ...(props.markers ?? []).filter(item => Number.isFinite(item.timestamp)).map(item => ({
      name: item.label ?? item.description ?? '',
      xAxis: item.timestamp,
      lineStyle: { color: item.color ?? ChartPalette.text('secondary'), type: item.lineStyle ?? 'dashed' },
      label: { show: !!item.label, formatter: chartText(item.label ?? '') },
    })),
  ]
  const series = props.data.map((item, index) => {
    const color = item.color ?? ChartPalette.categorical(index)
    const data = item.data.filter(([at, value]) => Number.isFinite(at) && Math.abs(at) <= 8640000000000000 && (value === null || Number.isFinite(value))).map(([at, value]) => [at, value] as [number, number | null]).sort((a, b) => a[0] - b[0])
    const base = {
      name: item.name,
      data,
      itemStyle: { color },
      markLine: index === 0 ? { symbol: 'none' as const, silent: true, data: annotations } : undefined,
      markArea: index === 0 && props.incomplete
        ? {
            silent: true,
            itemStyle: { color: 'var(--color-fill)', opacity: 0.35 },
            data: [
              ...(props.incomplete.before !== undefined ? [[{ xAxis: 'min' }, { xAxis: props.incomplete.before }]] : []),
              ...(props.incomplete.after !== undefined ? [[{ xAxis: props.incomplete.after }, { xAxis: 'max' }]] : []),
            ],
          }
        : undefined,
    }
    return type === 'bar'
      ? { ...base, type: 'bar' as const, barMaxWidth: 24 }
      : {
          ...base,
          type: 'line' as const,
          connectNulls: false,
          showSymbol: data.length === 1,
          lineStyle: { color, width: 2 },
          areaStyle: props.gradient ? { color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops: [{ offset: 0, color }, { offset: 1, color: 'transparent' }] }, opacity: 0.2 } : undefined,
        }
  })
  return {
    grid: { left: 12, right: 20, top: props.yAxisName ? 32 : 16, bottom: props.xAxisName ? 36 : 12, containLabel: true },
    xAxis: { type: 'time', name: props.xAxisName, splitNumber: props.xAxisTickCount ?? 5, axisLabel: { formatter: props.xAxisTickFormat, hideOverlap: true }, axisTick: { show: false }, axisLine: { show: false }, splitLine: { show: false } },
    yAxis: { type: 'value', name: props.yAxisName, splitNumber: props.yAxisTickCount ?? 4, minInterval: props.yAxisMinInterval, axisLabel: { formatter: props.yAxisTickFormat }, axisLine: { show: false }, axisTick: { show: false } },
    legend: props.enableLegendSelection ? { show: false, data: props.data.map(item => item.name) } : undefined,
    brush: props.selectable ? { toolbox: ['lineX', 'clear'], xAxisIndex: 0, brushMode: 'single' } : undefined,
    toolbox: props.selectable ? { show: true, feature: { brush: { type: ['lineX', 'clear'] } } } : undefined,
    tooltip: {
      trigger: 'axis',
      renderMode: 'richText',
      confine: true,
      position: props.tooltipFollowCursor === 'x' ? (point: number[]) => [point[0], 0] : undefined,
      formatter: (parameters) => {
        const points = (Array.isArray(parameters) ? parameters : [parameters]).filter(item => Array.isArray(item.value) && typeof item.value[1] === 'number' && Number.isFinite(item.value[1]))
        const visible = points.slice(0, props.tooltipMode === 'single' ? 1 : props.tooltipMaxItems ?? 10)
        const first = visible[0]?.value as [number, number] | undefined
        if (!first)
          return ''
        return [
          timestampFormat(first[0]),
          ...visible.map(item => `${item.seriesName}: ${valueFormat((item.value as [number, number])[1])}`),
          ...(points.length > visible.length ? [`+${points.length - visible.length}`] : []),
          ...(props.tooltipFooter ? [props.tooltipFooter] : []),
        ].map(chartText).join('\n')
      },
    },
    series: series as LineSeriesOption[],
  }
}

export function createSankeyOption(props: SankeyChartProps): EChartsOption {
  const format = props.formatValue ?? ((value: number) => value.toLocaleString())
  const identities = props.nodes.map((node, index) => node.id ?? `node-${index}`)
  const showValues = props.showNodeValues ?? props.nodes.some(node => node.value !== undefined)
  const data = props.nodes.map((node, index) => ({
    ...node,
    name: identities[index],
    datum: node,
    itemStyle: { color: node.color ?? props.defaultNodeColor ?? ChartPalette.categorical(index) },
    label: { formatter: chartText(showValues && node.value !== undefined ? `${format(node.value)}${props.nodeLabelLayout === 'inline' ? ' ' : '\n'}${node.name}` : node.name) },
  }))
  const links = props.links.filter(link => identities[link.source] !== undefined && identities[link.target] !== undefined && Number.isFinite(link.value) && link.value >= 0).map(link => ({ ...link, source: identities[link.source]!, target: identities[link.target]!, datum: link }))
  const series: SankeySeriesOption = {
    type: 'sankey',
    data,
    links,
    left: props.left ?? '5%',
    right: props.right ?? '5%',
    nodeWidth: props.nodeWidth ?? 16,
    nodeGap: props.nodePadding ?? 12,
    emphasis: { focus: 'adjacency' },
    lineStyle: { color: props.linkColor === 'gray' ? 'var(--color-fill)' : 'gradient', opacity: props.linkOpacity ?? 0.35, curveness: 0.5 },
  }
  return {
    tooltip: {
      show: props.showTooltip ?? true,
      trigger: 'item',
      renderMode: 'richText',
      confine: true,
      formatter: (parameter) => {
        const item = Array.isArray(parameter) ? parameter[0] : parameter
        const row = item?.data as { datum?: SankeyNodeData, source?: string, target?: string, value?: number } | undefined
        if (!row)
          return ''
        if (item?.dataType === 'edge') {
          const source = props.nodes[identities.indexOf(row.source!)]?.name ?? ''
          const target = props.nodes[identities.indexOf(row.target!)]?.name ?? ''
          const value = row.value ?? 0
          return chartText(props.tooltipFormatter?.({ type: 'link', name: `${source} → ${target}`, link: { source, target, value } }) ?? `${source} → ${target}\n${format(value)}`)
        }
        const node = row.datum!
        return chartText(props.tooltipFormatter?.({ type: 'node', name: node.name, node }) ?? [node.name, ...(node.value !== undefined ? [format(node.value)] : []), ...Object.entries(node.tooltipData ?? {}).map(([name, value]) => `${name}: ${typeof value === 'number' ? format(value) : value}`)].join('\n'))
      },
    },
    series: [series],
  }
}

export function readMapValue<T, V>(row: T, accessor: MapAccessor<T, V>): V {
  return typeof accessor === 'function' ? accessor(row) : row[accessor] as V
}

function readMapStyle<T, V>(row: T, style: MapStyle<T, V> | undefined, fallback: V): V {
  return style === undefined ? fallback : typeof style === 'function' ? (style as (row: T) => V)(row) : style
}

export const mercatorProjection: MapProjection = {
  project: ([longitude = 0, latitude = 0]) => [longitude, Math.log(Math.tan(Math.PI / 4 + Math.max(-85.0511, Math.min(85.0511, latitude)) * Math.PI / 360)) * 180 / Math.PI],
  unproject: ([x = 0, y = 0]) => [x, (2 * Math.atan(Math.exp(y * Math.PI / 180)) - Math.PI / 2) * 180 / Math.PI],
}

export function createBubbleMapOption<T>(props: Omit<BubbleMapProps<T>, 'geoJson'> & { mapName: string }): EChartsOption {
  const rows = props.data.map(row => ({ row, lng: readMapValue(row, props.lng), lat: readMapValue(row, props.lat), value: readMapValue(row, props.value) })).filter(item => Number.isFinite(item.lng) && Math.abs(item.lng) <= 180 && Number.isFinite(item.lat) && Math.abs(item.lat) <= 90 && Number.isFinite(item.value))
  const maximum = Math.max(0, ...rows.map(item => item.value))
  const min = Math.max(0, props.minRadius ?? 4)
  const max = Math.max(min, props.maxRadius ?? 24)
  const data = rows.map(item => ({
    name: props.name ? readMapValue(item.row, props.name) : '',
    value: [item.lng, item.lat, item.value],
    datum: item.row,
    symbolSize: Math.max(0, props.bubbleSize?.(item.value) ?? (min + Math.sqrt(Math.max(0, item.value) / (maximum || 1)) * (max - min)) * 2),
    itemStyle: { color: readMapStyle(item.row, props.bubbleColor, ChartPalette.categorical(0)), borderColor: readMapStyle(item.row, props.bubbleBorderColor, 'var(--color-base)'), borderWidth: readMapStyle(item.row, props.bubbleBorderWidth, 1) },
  }))
  return {
    geo: { map: props.mapName, roam: props.roam ?? true, center: props.center, zoom: props.zoom ?? 1, projection: props.projection === null ? undefined : props.projection ?? mercatorProjection, itemStyle: { areaColor: 'var(--color-fill)', borderColor: 'var(--color-line)' }, emphasis: { disabled: true } },
    tooltip: { show: props.showTooltip ?? true, trigger: 'item', renderMode: 'richText', confine: true, formatter: (parameter) => {
      const point = (Array.isArray(parameter) ? parameter[0] : parameter)?.data as typeof data[number] | undefined
      return point ? chartText(props.tooltipFormatter?.(point.datum) ?? `${point.name}\n${(props.valueFormat ?? String)(point.value[2]!)}`) : ''
    } },
    series: [{ type: 'scatter', coordinateSystem: 'geo', data }],
  }
}

export function createChoroplethMapOption<T>(props: Omit<ChoroplethMapProps<T>, 'geoJson'> & { mapName: string }): EChartsOption {
  const data = props.data.map(row => ({ name: readMapValue(row, props.name), value: readMapValue(row, props.value), datum: row })).filter(item => Number.isFinite(item.value))
  const values = data.map(item => item.value)
  const min = props.min ?? Math.min(0, ...values)
  const max = props.max ?? Math.max(1, ...values)
  return {
    visualMap: { type: 'continuous', min, max, show: props.showLegend ?? true, left: 12, bottom: 12, calculable: true, inRange: { color: props.colorRange ?? ChartPalette.sequential() } },
    tooltip: { show: props.showTooltip ?? true, trigger: 'item', renderMode: 'richText', confine: true, formatter: (parameter) => {
      const point = (Array.isArray(parameter) ? parameter[0] : parameter)?.data as typeof data[number] | undefined
      return point?.datum ? chartText(props.tooltipFormatter?.(point.datum) ?? `${point.name}\n${(props.valueFormat ?? String)(point.value)}`) : ''
    } },
    series: [{ type: 'map', map: props.mapName, nameProperty: props.nameProperty ?? 'name', data, roam: props.roam ?? true, center: props.center, zoom: props.zoom ?? 1, projection: props.projection === null ? undefined : props.projection ?? mercatorProjection, itemStyle: { areaColor: props.noDataColor ?? 'var(--color-fill)', borderColor: 'var(--color-line)' }, emphasis: { itemStyle: { borderColor: 'var(--color-brand)', borderWidth: 2 } } }],
  }
}

export function projectGlobePoint(longitude: number, latitude: number, rotation: [number, number, number], radius = 184) {
  const longitudeAngle = (longitude + rotation[0]) * Math.PI / 180
  const latitudeAngle = latitude * Math.PI / 180
  const tilt = rotation[1] * Math.PI / 180
  const roll = rotation[2] * Math.PI / 180
  const x = Math.cos(latitudeAngle) * Math.sin(longitudeAngle)
  const y = Math.cos(tilt) * Math.sin(latitudeAngle) + Math.sin(tilt) * Math.cos(latitudeAngle) * Math.cos(longitudeAngle)
  const depth = Math.cos(tilt) * Math.cos(latitudeAngle) * Math.cos(longitudeAngle) - Math.sin(tilt) * Math.sin(latitudeAngle)
  return { x: 200 + radius * (x * Math.cos(roll) - y * Math.sin(roll)), y: 200 - radius * (x * Math.sin(roll) + y * Math.cos(roll)), depth, visible: depth > 0 && Number.isFinite(x) && Math.abs(longitude) <= 180 && Math.abs(latitude) <= 90 }
}
