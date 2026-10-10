import type { CustomSeriesOption } from 'echarts'
import { expect, it } from 'vitest'
import { createGlobeOption, globeViewport } from './globe-model'

it('renders the sphere, geographic layers and markers through ECharts custom series', () => {
  const markers = [{ longitude: 0, latitude: 0, name: 'Front' }, { longitude: 180, latitude: 0, name: 'Back' }]
  const option = createGlobeOption({ markers, showGraticule: true }, [0, 0, 0])
  const layers = option.series as CustomSeriesOption[]
  expect(layers.map(layer => layer.type)).toEqual(['custom', 'custom', 'custom', 'custom'])
  expect(layers.map(layer => layer.id)).toEqual(['globe-sphere', 'globe-land', 'globe-graticule', 'globe-markers'])
  expect(layers[3]?.data).toEqual([expect.objectContaining({ name: 'Front', datum: markers[0], value: [200, 200, 1, 7] })])
  const rotated = createGlobeOption({ markers }, [180, 0, 0]).series as CustomSeriesOption[]
  expect(rotated[3]?.data).toEqual([expect.objectContaining({ name: 'Back', datum: markers[1] })])
})

it('fits the globe inside the smaller viewport dimension without stretching', () => {
  expect(globeViewport(600, 300)).toEqual({ scale: 0.75, x: 150, y: 0, centerX: 300, centerY: 150, radius: 138 })
  expect(globeViewport(300, 600)).toEqual({ scale: 0.75, x: 0, y: 150, centerX: 150, centerY: 300, radius: 138 })
})
