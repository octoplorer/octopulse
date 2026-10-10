import { expect, it } from 'vitest'
import { createBubbleMapOption, createSankeyOption, createTimeseriesOption, projectGlobePoint } from './model'

it('normalizes time-series observations without mutating the source', () => {
  const data = [{ name: 'Latency', color: 'var(--color-brand)', data: [[2000, 4], [1000, 2], [3000, Number.NaN]] as [number, number][] }]
  const option = createTimeseriesOption({ data, thresholds: [{ value: 5, color: 'var(--color-danger)' }] })
  expect((option.series as { data: unknown[], markLine: { data: unknown[] } }[])[0]?.data).toEqual([[1000, 2], [2000, 4]])
  expect((option.series as { markLine: { data: { yAxis: number }[] } }[])[0]?.markLine.data[0]?.yAxis).toBe(5)
  expect(data[0]?.data[0]?.[0]).toBe(2000)
})

it('drops dangling Sankey links and keeps stable node identifiers', () => {
  const option = createSankeyOption({ nodes: [{ id: 'a', name: 'Same' }, { id: 'b', name: 'Same' }], links: [{ source: 0, target: 1, value: 5 }, { source: 0, target: 3, value: 9 }] })
  const series = (option.series as { data: { name: string }[], links: unknown[] }[])[0]!
  expect(series.data.map(node => node.name)).toEqual(['a', 'b'])
  expect(series.links).toEqual([expect.objectContaining({ source: 'a', target: 'b', value: 5 })])
})

it('rejects invalid map coordinates and scales bubbles by their area', () => {
  const option = createBubbleMapOption({ mapName: 'test', data: [{ lng: 0, lat: 0, n: 0 }, { lng: 180, lat: 90, n: 100 }, { lng: 181, lat: 0, n: 3 }], lng: 'lng', lat: 'lat', value: 'n', minRadius: 2, maxRadius: 10 })
  const series = (option.series as { data: { symbolSize: number }[] }[])[0]!
  expect(series.data).toHaveLength(2)
  expect(series.data.map(item => item.symbolSize)).toEqual([4, 20])
})

it('hides globe markers on the far side of the earth', () => {
  expect(projectGlobePoint(0, 0, [0, 0, 0])).toMatchObject({ x: 200, y: 200, visible: true })
  expect(projectGlobePoint(180, 0, [0, 0, 0]).visible).toBe(false)
})

it('rejects invalid geographic marker coordinates', () => {
  expect(projectGlobePoint(360, 0, [0, 0, 0]).visible).toBe(false)
  expect(projectGlobePoint(0, 91, [0, 0, 0]).visible).toBe(false)
})
