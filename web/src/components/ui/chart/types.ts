import type { EChartsOption } from 'echarts'
import type { init, SetOptionOpts } from 'echarts/core'

export type KumoChartOption = EChartsOption
export interface ChartProps {
  option: KumoChartOption
  height?: number
  ariaLabel?: string
  loading?: boolean
  theme?: Parameters<typeof init>[1]
  initOptions?: Omit<NonNullable<Parameters<typeof init>[2]>, 'renderer'> & { renderer?: 'canvas' }
  updateOptions?: SetOptionOpts
}
export interface ChartEvents {
  click?: (params: unknown) => void
  mouseover?: (params: unknown) => void
  mouseout?: (params: unknown) => void
  brushEnd?: (params: unknown) => void
}
export interface TimeseriesData {
  name: string
  data: [number, number | null][]
  color?: string
}
export interface TimeseriesMarker {
  timestamp: number
  label?: string
  description?: string
  color?: string
  lineStyle?: 'solid' | 'dashed' | 'dotted'
}
export interface TimeseriesThreshold {
  value: number
  label?: string
  color: string
}
export interface TimeseriesChartProps {
  data: TimeseriesData[]
  type?: 'line' | 'bar'
  markers?: TimeseriesMarker[]
  thresholds?: TimeseriesThreshold[]
  xAxisName?: string
  xAxisTickCount?: number
  xAxisTickFormat?: (value: number) => string
  yAxisName?: string
  yAxisTickCount?: number
  yAxisTickFormat?: (value: number) => string
  yAxisMinInterval?: number
  tooltipValueFormat?: (value: number) => string
  tooltipTimestampFormat?: (timestamp: number) => string
  tooltipFooter?: string
  tooltipMode?: 'all' | 'single'
  tooltipMaxItems?: number
  tooltipFollowCursor?: 'both' | 'x'
  incomplete?: { before?: number, after?: number }
  enableLegendSelection?: boolean
  height?: number
  gradient?: boolean
  loading?: boolean
  ariaDescription?: string
  optionUpdateBehavior?: SetOptionOpts
  selectable?: boolean
}
export interface SankeyNodeData {
  id?: string
  name: string
  color?: string
  value?: number
  tooltipData?: Record<string, number | string>
  isDrillable?: boolean
  childCount?: number
}
export interface SankeyLinkData {
  id?: string
  source: number
  target: number
  value: number
  isDrillable?: boolean
}
export interface SankeyTooltipParams {
  type: 'node' | 'link'
  name: string
  node?: SankeyNodeData
  link?: { source: string, target: string, value: number }
  color?: string
}
export interface SankeyChartProps {
  nodes: SankeyNodeData[]
  links: SankeyLinkData[]
  height?: number
  showNodeValues?: boolean
  nodeLabelLayout?: 'stacked' | 'inline'
  formatValue?: (value: number) => string
  tooltipFormatter?: (params: SankeyTooltipParams) => string
  nodeWidth?: number
  nodePadding?: number
  showTooltip?: boolean
  defaultNodeColor?: string
  left?: number | string
  right?: number | string
  linkColor?: 'gradient' | 'gray'
  linkOpacity?: number
  ariaLabel?: string
  loading?: boolean
}
export interface MapGeoJson {
  type: 'FeatureCollection'
  features: { type: 'Feature', id?: string | number, properties?: Record<string, unknown> | null, geometry: unknown }[]
}
export type MapAccessor<T, V> = { [K in keyof T]-?: T[K] extends V ? K : never }[keyof T] | ((row: T) => V)
export type MapStyle<T, V> = V | ((row: T) => V)
export interface MapProjection {
  project: (point: number[]) => number[]
  unproject: (point: number[]) => number[]
}
export interface MapBaseProps {
  geoJson: MapGeoJson
  mapName?: string
  center?: [number, number]
  zoom?: number
  roam?: boolean
  projection?: MapProjection | null
  showTooltip?: boolean
  valueFormat?: (value: number) => string
  height?: number
  aspectRatio?: number | string
  ariaLabel?: string
  loading?: boolean
}
export interface BubbleMapProps<T> extends MapBaseProps {
  data: T[]
  lng: MapAccessor<T, number>
  lat: MapAccessor<T, number>
  value: MapAccessor<T, number>
  name?: MapAccessor<T, string>
  minRadius?: number
  maxRadius?: number
  bubbleSize?: (value: number) => number
  bubbleColor?: MapStyle<T, string>
  bubbleBorderColor?: MapStyle<T, string>
  bubbleBorderWidth?: MapStyle<T, number>
  tooltipFormatter?: (row: T) => string
}
export interface ChoroplethMapProps<T> extends MapBaseProps {
  data: T[]
  name: MapAccessor<T, string>
  value: MapAccessor<T, number>
  nameProperty?: string
  colorRange?: string[]
  min?: number
  max?: number
  noDataColor?: string
  showLegend?: boolean
  tooltipFormatter?: (row: T) => string
}
export interface GlobeMapMarker {
  longitude: number
  latitude: number
  name: string
  description?: string
  color?: string
  radius?: number
}
export interface GlobeMapProps {
  landColor?: string
  landHatchSpacing?: number
  oceanColor?: string
  markers?: GlobeMapMarker[]
  markerColor?: string
  markerRadius?: number
  defaultRotation?: [number, number, number]
  draggable?: boolean
  autoRotate?: boolean
  autoRotateSpeed?: number
  showGraticule?: boolean
  showTooltip?: boolean
  ariaLabel?: string
  height?: number
}
