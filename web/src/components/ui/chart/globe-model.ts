import type { CustomSeriesOption, CustomSeriesRenderItemAPI, EChartsOption } from 'echarts'
import type { GlobeMapProps } from './types'
import { isCanonicalLand } from './globe-land-mask'
import { projectGlobePoint } from './model'

const landCache = new Map<number, [number, number][]>()

function globeLandPoints(hatchSpacing: number) {
  const spacing = Math.max(2, Math.min(12, Math.round(hatchSpacing / 3)))
  const cached = landCache.get(spacing)
  if (cached)
    return cached
  const points: [number, number][] = []
  for (let latitude = -89; latitude < 90; latitude += spacing) {
    for (let longitude = -180; longitude < 180; longitude += spacing) {
      if (isCanonicalLand(longitude, latitude))
        points.push([longitude, latitude])
    }
  }
  landCache.set(spacing, points)
  return points
}

export function globeViewport(width: number, height: number) {
  const scale = Math.max(0, Math.min(width, height)) / 400
  return { scale, x: (width - 400 * scale) / 2, y: (height - 400 * scale) / 2, centerX: width / 2, centerY: height / 2, radius: 184 * scale }
}

function globeLandPath(rotation: [number, number, number], hatchSpacing: number) {
  return globeLandPoints(hatchSpacing).flatMap(([longitude, latitude]) => {
    const point = projectGlobePoint(longitude, latitude, rotation)
    return point.visible ? [`M${point.x.toFixed(2)},${point.y.toFixed(2)}l2,-2`] : []
  }).join('')
}

const graticuleLines: [number, number][][] = []
for (let latitude = -60; latitude <= 60; latitude += 30)
  graticuleLines.push(Array.from({ length: 181 }, (_, index) => [-180 + index * 2, latitude]))
for (let longitude = -180; longitude < 180; longitude += 30)
  graticuleLines.push(Array.from({ length: 91 }, (_, index) => [longitude, -90 + index * 2]))

function globeGraticulePath(rotation: [number, number, number]) {
  return graticuleLines.map((line) => {
    let previousVisible = false
    return line.map(([longitude, latitude]) => {
      const point = projectGlobePoint(longitude, latitude, rotation)
      const path = point.visible ? `${previousVisible ? 'L' : 'M'}${point.x.toFixed(2)},${point.y.toFixed(2)}` : ''
      previousVisible = point.visible
      return path
    }).join('')
  }).join('')
}

function geographicPathLayer(id: string, path: string, color: string, lineWidth: number, z: number): CustomSeriesOption {
  return {
    id,
    type: 'custom',
    coordinateSystem: 'none',
    silent: true,
    z,
    data: path ? [{ value: [0], itemStyle: { color: 'transparent', borderColor: color } }] : [],
    renderItem: (_params, api) => {
      const viewport = globeViewport(api.getWidth(), api.getHeight())
      return {
        type: 'group',
        x: viewport.x,
        y: viewport.y,
        scaleX: viewport.scale,
        scaleY: viewport.scale,
        clipPath: { type: 'circle' as const, shape: { cx: 200, cy: 200, r: 184 } },
        children: [{ type: 'path', shape: { pathData: path }, style: { fill: 'none', stroke: api.visual('borderColor'), lineWidth } }],
      }
    },
  }
}

function markerShape(api: CustomSeriesRenderItemAPI) {
  const viewport = globeViewport(api.getWidth(), api.getHeight())
  return {
    type: 'circle' as const,
    emphasisDisabled: true,
    shape: { cx: viewport.x + Number(api.value(0)) * viewport.scale, cy: viewport.y + Number(api.value(1)) * viewport.scale, r: Number(api.value(3)) * viewport.scale },
    style: { fill: api.visual('color'), stroke: api.visual('borderColor'), lineWidth: 1.5 * viewport.scale, opacity: Math.min(1, Number(api.value(2)) * 4) },
  }
}

export function globeVisibleMarkers(props: GlobeMapProps, rotation: [number, number, number]) {
  return (props.markers ?? []).flatMap((marker, index) => {
    const point = projectGlobePoint(marker.longitude, marker.latitude, rotation)
    return point.visible ? [{ marker, index, ...point }] : []
  })
}

export function createGlobeOption(props: GlobeMapProps, rotation: [number, number, number]): EChartsOption {
  const sphere: CustomSeriesOption = {
    id: 'globe-sphere',
    type: 'custom',
    coordinateSystem: 'none',
    silent: true,
    z: 0,
    data: [{ value: [0], itemStyle: { color: props.oceanColor ?? 'var(--color-base)', borderColor: 'var(--color-line)' } }],
    renderItem: (_params, api) => {
      const viewport = globeViewport(api.getWidth(), api.getHeight())
      return { type: 'circle', shape: { cx: viewport.centerX, cy: viewport.centerY, r: viewport.radius }, style: { fill: api.visual('color'), stroke: api.visual('borderColor'), lineWidth: 1 } }
    },
  }
  const markers: CustomSeriesOption = {
    id: 'globe-markers',
    type: 'custom',
    coordinateSystem: 'none',
    z: 3,
    renderItem: (_params, api) => markerShape(api),
    data: globeVisibleMarkers(props, rotation).map(point => ({
      id: `marker-${point.index}`,
      name: point.marker.name,
      datum: point.marker,
      value: [point.x, point.y, point.depth, point.marker.radius ?? props.markerRadius ?? 7],
      itemStyle: { color: point.marker.color ?? props.markerColor ?? 'var(--color-brand)', borderColor: 'var(--color-base)' },
    })),
  }
  return {
    animation: false,
    tooltip: { show: false },
    aria: { enabled: false },
    series: [
      sphere,
      geographicPathLayer('globe-land', globeLandPath(rotation, props.landHatchSpacing ?? 10), props.landColor ?? 'var(--color-interact)', 1.5, 1),
      geographicPathLayer('globe-graticule', props.showGraticule ? globeGraticulePath(rotation) : '', 'var(--color-line)', 0.7, 2),
      markers,
    ],
  }
}
