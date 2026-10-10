<script setup lang="ts">
import type { EChartsOption } from 'echarts'
import type { GlobeMapMarker, MapGeoJson, TimeseriesData } from '../ui/chart'
import { computed, shallowRef } from 'vue'
import { Badge } from '../ui/badge'
import { ResourceListPage } from '../ui/blocks/resource-list'
import { Button } from '../ui/button'
import { BubbleMap, Chart, ChartLegendLargeItem, ChartLegendSmallItem, ChartPalette, ChoroplethMap, GlobeMap, SankeyChart, TimeseriesChart } from '../ui/chart'
import { Checkbox } from '../ui/checkbox'
import { FlowAnchor, FlowList, FlowNode, FlowParallel, FlowRoot } from '../ui/flow'
import { Grid } from '../ui/grid'
import { Input } from '../ui/input'
import { LayerCard, LayerCardPrimary, LayerCardSecondary } from '../ui/layer-card'
import { Loader } from '../ui/loader'
import { Meter } from '../ui/meter'
import { Table, TableBody, TableCell, TableContainer, TableHead, TableHeader, TableRow, TableToolbar } from '../ui/table'
import { Text } from '../ui/text'

const loading = shallowRef(false)
const showGraticule = shallowRef(true)
const autoRotate = shallowRef(false)
const selectedRegion = shallowRef('选择地图上的区域或标记')
const selectedSankey = shallowRef('选择节点或连接，查看原始数据')
const selectedGlobe = shallowRef('选择伦敦或纽约标记')
const rotation = shallowRef<[number, number, number]>([0, 0, 0])
const selectedRange = shallowRef<[number, number]>()
const timeChart = shallowRef<InstanceType<typeof TimeseriesChart>>()
const globe = shallowRef<InstanceType<typeof GlobeMap>>()
const startTime = Date.UTC(2026, 9, 9, 0)
const timeFormat = new Intl.DateTimeFormat('zh-CN', { hour: '2-digit', minute: '2-digit', timeZone: 'UTC' })
const visibleSeries = shallowRef(['主区域', '备区域'])
const allTimeseries: TimeseriesData[] = [
  { name: '主区域', color: ChartPalette.categorical(0), data: [42, 38, 52, 47, 36, 41, 58, 46, 39, 44, 37, 43].map((value, index) => [startTime + index * 3600000, value]) },
  { name: '备区域', color: ChartPalette.categorical(1), data: [51, 46, 60, 52, 44, 47, 66, 51, 48, 53, 46, 50].map((value, index) => [startTime + index * 3600000, value]) },
]
const timeseries = computed(() => allTimeseries.filter(series => visibleSeries.value.includes(series.name)))
const rangeLabel = computed(() => selectedRange.value ? `${timeFormat.format(selectedRange.value[0])} – ${timeFormat.format(selectedRange.value[1])} UTC` : '尚未选择时间范围')
const rawChartOption = computed<EChartsOption>(() => ({
  grid: { top: 20, left: 12, right: 12, bottom: 12, containLabel: true },
  xAxis: { type: 'category', data: ['欧洲', '北美', '亚太'] },
  yAxis: { type: 'value', name: '请求数' },
  tooltip: { trigger: 'axis' },
  series: [{ name: '请求数', type: 'bar', data: [120, 86, 104], barMaxWidth: 40, itemStyle: { color: ChartPalette.semantic('Success') } }],
}))
const sankeyNodes = [
  { name: '进入探测', value: 1000, color: ChartPalette.categorical(0) },
  { name: '检查成功', value: 920, color: ChartPalette.semantic('Success') },
  { name: '重试检查', value: 80, color: ChartPalette.semantic('Warning') },
  { name: '恢复', value: 65, color: ChartPalette.semantic('Success') },
  { name: '发送告警', value: 15, color: ChartPalette.semantic('Attention') },
]
const sankeyLinks = [
  { source: 0, target: 1, value: 920 },
  { source: 0, target: 2, value: 80 },
  { source: 2, target: 3, value: 65 },
  { source: 2, target: 4, value: 15 },
]

// Small fictional regions keep map previews deterministic and entirely offline.
const mapRegions = [
  { name: '西区', longitude: -10, latitude: 0, value: 86, color: ChartPalette.categorical(0) },
  { name: '中区', longitude: 0, latitude: 0, value: 52, color: ChartPalette.categorical(1) },
  { name: '东区', longitude: 10, latitude: 0, value: 24, color: ChartPalette.categorical(2) },
]
const geoJson: MapGeoJson = {
  type: 'FeatureCollection',
  features: mapRegions.map(region => ({
    type: 'Feature',
    properties: { name: region.name },
    geometry: { type: 'Polygon', coordinates: [[[region.longitude - 5, -5], [region.longitude + 5, -5], [region.longitude + 5, 5], [region.longitude - 5, 5], [region.longitude - 5, -5]]] },
  })),
}
const globeMarkers: GlobeMapMarker[] = [
  { longitude: -0.12, latitude: 51.5, name: '伦敦', description: '欧洲探测区域' },
  { longitude: -74, latitude: 40.7, name: '纽约', description: '北美探测区域', color: ChartPalette.categorical(1) },
]
const resourceSearch = shallowRef('')
const resources = shallowRef([
  { id: 'eu', name: '欧洲探测', region: '伦敦', enabled: true },
  { id: 'us', name: '北美探测', region: '纽约', enabled: true },
  { id: 'ap', name: '亚太探测', region: '东京', enabled: false },
])
const filteredResources = computed(() => resources.value.filter(resource => `${resource.name} ${resource.region}`.includes(resourceSearch.value.trim())))

function toggleSeries(name: string) {
  visibleSeries.value = visibleSeries.value.includes(name) ? visibleSeries.value.filter(series => series !== name) : [...visibleSeries.value, name]
}
function selectTimeRange() {
  timeChart.value?.dispatchAction({ type: 'takeGlobalCursor', key: 'brush', brushOption: { brushType: 'lineX', brushMode: 'single' } })
}
function clearTimeRange() {
  timeChart.value?.dispatchAction({ type: 'brush', areas: [] })
  selectedRange.value = undefined
}
function showRange(from: number, to: number) {
  selectedRange.value = [from, to]
}
function toggleResource(id: string) {
  resources.value = resources.value.map(resource => resource.id === id ? { ...resource, enabled: !resource.enabled } : resource)
}
</script>

<template>
  <section class="space-y-4" aria-label="全部图表与组合布局预览">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <Text as="h2" variant="heading2">
          图表与组合布局
        </Text>
        <Text as="p" size="sm" variant="secondary">
          所有图表通过 VueECharts 渲染；区域地图采用离线样本，主题随页面切换。
        </Text>
      </div>
      <Checkbox v-model="loading" label="模拟图表加载" />
    </div>
    <Grid variant="2up" gap="sm">
      <LayerCard>
        <LayerCardSecondary>Chart · 通用图表与调色板</LayerCardSecondary>
        <LayerCardPrimary class="space-y-3">
          <Chart :option="rawChartOption" :height="220" :loading="loading" aria-label="各探测区域请求数条形图" />
          <div class="flex flex-wrap gap-4">
            <ChartLegendSmallItem name="欧洲" value="120" :color="ChartPalette.semantic('Success')" :loading="loading" />
            <ChartLegendSmallItem name="北美" value="86" :color="ChartPalette.categorical(1)" :loading="loading" />
            <ChartLegendSmallItem name="亚太" value="104" :color="ChartPalette.categorical(2)" :loading="loading" />
          </div>
        </LayerCardPrimary>
      </LayerCard>
      <LayerCard>
        <LayerCardSecondary>TimeseriesChart · 图例与时间范围</LayerCardSecondary>
        <LayerCardPrimary class="space-y-3">
          <div class="flex flex-wrap gap-4">
            <ChartLegendLargeItem v-for="series in allTimeseries" :key="series.name" :name="series.name" :value="`${series.data.at(-1)?.[1]}`" unit="ms" :color="series.color" :inactive="!visibleSeries.includes(series.name)" :loading="loading" @click="toggleSeries(series.name)" />
          </div>
          <TimeseriesChart
            ref="timeChart" :data="timeseries" :height="180" :loading="loading" selectable
            :thresholds="[{ value: 60, label: '延迟阈值', color: ChartPalette.semantic('Warning') }]"
            :markers="[{ timestamp: startTime + 6 * 3600000, label: '发布', description: '服务发布完成' }]"
            :x-axis-tick-format="value => timeFormat.format(value)" :tooltip-value-format="value => `${value} ms`"
            :option-update-behavior="{ replaceMerge: ['series'] }" aria-description="两个探测区域响应延迟，支持拖动选择时间范围"
            @time-range-change="showRange"
          />
          <div class="flex flex-wrap items-center gap-2">
            <Button size="sm" :disabled="loading" @click="selectTimeRange">
              选择时间范围
            </Button>
            <Button size="sm" variant="ghost" :disabled="loading" @click="clearTimeRange">
              清除范围
            </Button>
            <Text size="xs" variant="secondary" aria-live="polite">
              {{ rangeLabel }}
            </Text>
          </div>
        </LayerCardPrimary>
      </LayerCard>
      <LayerCard>
        <LayerCardSecondary>SankeyChart · 检查流量</LayerCardSecondary>
        <LayerCardPrimary class="space-y-3">
          <SankeyChart :nodes="sankeyNodes" :links="sankeyLinks" :height="220" :loading="loading" node-label-layout="inline" aria-label="检查成功、重试、恢复与告警之间的流量" @node-click="selectedSankey = `${$event.name}：${$event.value ?? 0} 次`" @link-click="selectedSankey = `${sankeyNodes[$event.source]?.name} → ${sankeyNodes[$event.target]?.name}：${$event.value} 次`" />
          <Text as="p" size="sm" variant="secondary" aria-live="polite">
            {{ selectedSankey }}
          </Text>
        </LayerCardPrimary>
      </LayerCard>
      <LayerCard>
        <LayerCardSecondary>GlobeMap · 旋转、标记与键盘操作</LayerCardSecondary>
        <LayerCardPrimary class="space-y-3">
          <div class="flex flex-wrap gap-4">
            <Checkbox v-model="autoRotate" label="自动旋转" />
            <Checkbox v-model="showGraticule" label="显示经纬线" />
            <Button size="xs" @click="globe?.setRotation([0, 0, 0])">
              重置视角
            </Button>
          </div>
          <div class="relative">
            <GlobeMap ref="globe" :markers="globeMarkers" :height="220" :show-graticule="showGraticule" :auto-rotate="autoRotate" :auto-rotate-speed="8" aria-label="全球探测区域，拖动或用方向键旋转，聚焦标记后按回车选择" @marker-click="selectedGlobe = `${$event.name}：${$event.description}`" @user-rotation-change="rotation = $event" />
            <div v-if="loading" class="absolute inset-0 flex items-center justify-center bg-base/80" aria-busy="true">
              <Loader label="正在加载地球图" />
            </div>
          </div>
          <Text as="p" size="xs" variant="secondary" aria-live="polite">
            {{ selectedGlobe }} · 经度 {{ rotation[0].toFixed(0) }}°，纬度 {{ rotation[1].toFixed(0) }}°
          </Text>
        </LayerCardPrimary>
      </LayerCard>
      <LayerCard>
        <LayerCardSecondary>BubbleMap · 区域请求量</LayerCardSecondary>
        <LayerCardPrimary class="space-y-3">
          <BubbleMap :geo-json="geoJson" :data="mapRegions" lng="longitude" lat="latitude" value="value" name="name" :bubble-color="row => row.color" :height="220" :loading="loading" :projection="null" aria-label="离线西区、中区、东区请求量气泡图" @bubble-click="selectedRegion = `${$event.name}：${$event.value} 个请求`" />
          <Text as="p" size="sm" variant="secondary" aria-live="polite">
            {{ selectedRegion }}
          </Text>
        </LayerCardPrimary>
      </LayerCard>
      <LayerCard>
        <LayerCardSecondary>ChoroplethMap · 区域负载</LayerCardSecondary>
        <LayerCardPrimary class="space-y-3">
          <ChoroplethMap :geo-json="geoJson" :data="mapRegions" name="name" value="value" :height="220" :loading="loading" :projection="null" :min="0" :max="100" :value-format="value => `${value}%`" aria-label="离线西区、中区、东区负载分级设色地图" @region-click="selectedRegion = `${$event.name}：负载 ${$event.value}%`" />
          <Text as="p" size="sm" variant="secondary" aria-live="polite">
            {{ selectedRegion }}
          </Text>
        </LayerCardPrimary>
      </LayerCard>
    </Grid>
    <LayerCard>
      <LayerCardSecondary>Flow · 并行分支、锚点与可拖动画布</LayerCardSecondary>
      <LayerCardPrimary class="space-y-3">
        <Text as="p" size="sm" variant="secondary">
          在节点之外拖动画布，或聚焦画布后用方向键滚动。
        </Text>
        <FlowRoot class="h-64" :padding="{ x: 24, y: 24 }" :column-gap="64" aria-label="探测任务并行处理流程，可拖动或用方向键滚动画布">
          <FlowNode id="showcase-schedule" :style="{ minWidth: '180px' }">
            <FlowAnchor type="start">
              计划探测
            </FlowAnchor>
            <Text as="p" size="xs" variant="secondary">
              每分钟执行一次
            </Text>
          </FlowNode>
          <FlowParallel align="end">
            <FlowList>
              <FlowNode id="showcase-http">
                <FlowAnchor>HTTP 检查</FlowAnchor>
              </FlowNode>
              <FlowNode id="showcase-assert">
                <FlowAnchor>响应断言</FlowAnchor>
              </FlowNode>
            </FlowList>
            <FlowNode id="showcase-tcp">
              <FlowAnchor>TCP 检查</FlowAnchor>
            </FlowNode>
          </FlowParallel>
          <FlowNode id="showcase-aggregate" :style="{ minWidth: '180px' }">
            <FlowAnchor type="end">
              合并检查结果
            </FlowAnchor>
            <Text as="p" size="xs" variant="secondary">
              更新状态与发送通知
            </Text>
          </FlowNode>
        </FlowRoot>
      </LayerCardPrimary>
    </LayerCard>
    <div class="overflow-hidden rounded-xl border border-line">
      <ResourceListPage title="探测区域资源" description="ResourceListPage · 列表、筛选和用量栏组合示例">
        <template #header>
          <Text as="h3" variant="heading3">
            ResourceListPage · 探测区域资源
          </Text>
          <Text as="p" size="sm" variant="secondary">
            筛选区域，并切换演示资源的启用状态。
          </Text>
        </template>
        <TableContainer>
          <TableToolbar>
            <Input v-model="resourceSearch" aria-label="筛选演示探测区域" placeholder="筛选区域，例如：欧洲" class="max-w-64" />
            <Badge>{{ filteredResources.length }} 个区域</Badge>
          </TableToolbar>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>探测区域</TableHead><TableHead>城市</TableHead><TableHead>状态</TableHead><TableHead class="text-right">
                  操作
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="resource in filteredResources" :key="resource.id">
                <TableCell>{{ resource.name }}</TableCell><TableCell>{{ resource.region }}</TableCell>
                <TableCell>
                  <Badge :variant="resource.enabled ? 'success' : 'outline'">
                    {{ resource.enabled ? '已启用' : '已暂停' }}
                  </Badge>
                </TableCell>
                <TableCell class="text-right">
                  <Button size="sm" @click="toggleResource(resource.id)">
                    {{ resource.enabled ? '暂停' : '启用' }}
                  </Button>
                </TableCell>
              </TableRow>
              <TableRow v-if="!filteredResources.length">
                <TableCell :colspan="4" class="py-6 text-center text-subtle">
                  没有匹配的区域
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </TableContainer>
        <template #usage>
          <LayerCard>
            <LayerCardSecondary>资源用量</LayerCardSecondary><LayerCardPrimary>
              <Meter :value="resources.filter(resource => resource.enabled).length / 5 * 100" label="启用探测区域" /><Text as="p" size="sm" variant="secondary" class="mt-3">
                {{ resources.filter(resource => resource.enabled).length }} / 5 个区域已启用
              </Text>
            </LayerCardPrimary>
          </LayerCard>
        </template>
        <template #additionalContent>
          <Text as="p" size="sm" variant="secondary">
            演示数据只保存在当前预览页面，刷新后恢复。
          </Text>
        </template>
      </ResourceListPage>
    </div>
  </section>
</template>
