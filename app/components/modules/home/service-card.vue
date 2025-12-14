<script setup lang="ts">
import type { SerializeObject } from 'nitropack'

const props = defineProps<{
  overview: SerializeObject<ServiceOverview>
}>()

function getStatusColor(status: ServiceStatus): string {
  switch (status) {
    case ServiceStatus.Operational: return 'bg-emerald-500'
    case ServiceStatus.Degraded: return 'bg-amber-400'
    case ServiceStatus.Down: return 'bg-rose-500'
    case ServiceStatus.Maintenance: return 'bg-blue-400'
    default: return 'bg-gray-400'
  }
}

function getStatusText(status: ServiceStatus): string {
  switch (status) {
    case ServiceStatus.Operational: return '运行正常'
    case ServiceStatus.Degraded: return '性能降低'
    case ServiceStatus.Down: return '服务中断'
    case ServiceStatus.Maintenance: return '维护中'
    default: return '未知状态'
  }
}

// Calculate 30-day uptime percentage
const uptimePercentage = computed(() => {
  if (!props.overview.histories || props.overview.histories.length === 0) {
    return 0
  }

  const validHistories = props.overview.histories.filter(h => h.uptime !== null)
  if (validHistories.length === 0) {
    return 0
  }

  const totalUptime = validHistories.reduce((sum, h) => sum + (h.uptime ?? 0), 0)
  return totalUptime / validHistories.length
})

// Get latest latency from most recent history
const latestLatency = computed(() => {
  const histories = [...props.overview.histories].sort((a, b) =>
    Temporal.ZonedDateTime.from(b.date).epochMilliseconds - Temporal.ZonedDateTime.from(a.date).epochMilliseconds,
  )
  return histories.find(h => h.latency !== null)?.latency ?? null
})

// Get last check time from overview
const lastCheckTime = computed(() => {
  if (!props.overview.lastCheckTime) {
    return null
  }

  try {
    // Reconstruct Temporal.ZonedDateTime from serialized string
    const date = Temporal.ZonedDateTime.from(props.overview.lastCheckTime)
      .withTimeZone(Intl.DateTimeFormat().resolvedOptions().timeZone)

    return date.toLocaleString('zh-CN', {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      hourCycle: 'h24',
    })
  }
  catch (error) {
    console.error('Error parsing lastCheckTime:', error)
    return null
  }
})

const latencyColor = computed(() => {
  if (latestLatency.value === null) {
    return 'text-gray-400'
  }
  if (latestLatency.value > 800) {
    return 'text-rose-500'
  }
  else if (latestLatency.value > 500) {
    return 'text-amber-400'
  }
  else {
    return 'text-emerald-500'
  }
})
</script>

<template>
  <div
    bg="white dark:zinc-900"
    border="xl zinc-200 dark:zinc-800 rounded-lg"
    overflow-hidden shadow-sm hover:shadow-md
    transition-all duration-300
  >
    <div p-5 flex="~ grow col">
      <div flex="~ justify-between items-start" data-part="header">
        <div flex="~ 1 items-center gap-x-3">
          <div relative>
            <div
              size-3 rounded-full
              :class="[
                getStatusColor(overview.service.status),
                overview.service.status === ServiceStatus.Operational && 'animate-pulse',
              ]"
            />
            <div
              v-if="overview.service.status === ServiceStatus.Operational"
              absolute top-0 left-0 size-3
              rounded-full bg-emerald-500 opacity-50 animate-ping
            />
          </div>
          <div>
            <h3 un-text="lg zinc-800 dark:zinc-100" font-bold leading-tight transition-colors>
              {{ overview.service.name }}
            </h3>
            <NuxtLink
              v-if="overview.service.url"
              :href="overview.service.url"
              external
              rel="noopener noreferrer"
              un-text="sm zinc-500 dark:zinc-400 hover:indigo-600 dark:hover:indigo-400"
              flex="~ items-center gap-1"
              mt-0.5
              transition-colors
            >
              {{ overview.service.url }}
              <div size-4 i-lucide:external-link />
            </NuxtLink>
          </div>
        </div>
        <span
          p="x3 y1"
          rounded-full
          un-text="xs white"
          font-medium
          :class="getStatusColor(overview.service.status)"
        >
          {{ getStatusText(overview.service.status) }}
        </span>
      </div>

      <div
        mt-4 pt-3
        border="t zinc-100 dark:zinc-800 dashed"
        flex="~ items-center justify-between gap-4"
        data-part="information"
      >
        <div
          v-if="overview.service.tags && overview.service.tags.length > 0"
          flex="~ item-center 1 gap-1.5"
          min-w-0 overflow-x-auto
          class="[scrollbar-width:none] [-ms-overflow-style:none] [&::-webkit-scrollbar]:hidden mask-linear-fade"
        >
          <span
            v-for="tag in overview.service.tags"
            :key="tag"
            flex="inline items-center shrink-0"
            p="x2 y0.5"
            un-text="[10px] zinc-600 dark:zinc-400"
            font-medium
            bg="zinc-100 dark:zinc-800"
            border="rounded zinc-200 dark:zinc-700/50"
            whitespace-nowrap
          >
            {{ tag }}
          </span>
        </div>

        <div
          flex="~ items-center gap-3 shrink-0"
          un-text="xs zinc-500 dark:zinc-400"
          font-medium ml-auto pl-2
        >
          <div flex="~ items-center gap-1.5" title="实时延迟">
            <div size-13px i-lucide:zap :class="latencyColor" aria-hidden="true" />
            <span>{{ latestLatency }}ms</span>
          </div>
          <div w-px h-3 bg="zinc-200 dark:zinc-700" />
          <div flex="~ items-center gap-1.5" title="最后更新">
            <div size-13px i-lucide:clock un-text="zinc-600 dark:zinc-400" aria-hidden="true" />
            <span tabular-nums>{{ lastCheckTime }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Bottom Section: Uptime Percentage and History -->
    <div data-part="visualizer" p="x5 b5" mt-auto>
      <div pt-2 border="t zinc-100 dark:zinc-800/50">
        <div flex="~ justify-between items-center" mb-3>
          <div flex="~ items-baseline gap-2">
            <span un-text="2xl zinc-800 dark:zinc-100" font-bold>
              {{ uptimePercentage.toFixed(2) }}%
            </span>
            <span un-text="xs zinc-500 dark:zinc-400" font-medium>30天在线率</span>
          </div>
        </div>
        <div flex="~ items-center gap-2px" w-full h-8>
          <HomeUptimeBlock
            v-for="history in overview.histories"
            :key="history.date.toString()"
            :history="history"
          />
        </div>
        <div flex="~ justify-between" mt-2 un-text="[10px] zinc-400 dark:zinc-500">
          <span>30天前</span>
          <span>今天</span>
        </div>
      </div>
    </div>
  </div>
</template>
