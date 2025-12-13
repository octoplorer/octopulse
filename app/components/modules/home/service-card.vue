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
    class="bg-white dark:bg-zinc-900 rounded-xl shadow-sm border border-zinc-200 dark:border-zinc-800 overflow-hidden hover:shadow-md transition-all duration-300"
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
              class="absolute top-0 left-0 size-3 rounded-full bg-emerald-500 opacity-50 animate-ping"
            />
          </div>
          <div>
            <h3 class="text-lg font-bold text-zinc-800 dark:text-zinc-100 leading-tight transition-colors">
              {{ overview.service.name }}
            </h3>
            <NuxtLink
              v-if="overview.service.url"
              :href="overview.service.url"
              external
              rel="noopener noreferrer"
              class="text-sm text-zinc-500 dark:text-zinc-400 hover:text-indigo-600 dark:hover:text-indigo-400 flex items-center gap-1 mt-0.5 transition-colors"
            >
              {{ overview.service.url }}
              <div size-4 class="i-lucide:external-link" />
            </NuxtLink>
          </div>
        </div>
        <span
          class="px-3 py-1 rounded-full text-xs font-medium text-white"
          :class="getStatusColor(overview.service.status)"
        >
          {{ getStatusText(overview.service.status) }}
        </span>
      </div>

      <div
        class="mt-4 pt-3 border-t border-zinc-100 dark:border-zinc-800 border-dashed flex items-center justify-between gap-4"
        data-part="information"
      >
        <div
          v-if="overview.service.tags && overview.service.tags.length > 0"
          flex="~ item-center 1 gap-1.5"
          class="min-w-0 overflow-x-auto [scrollbar-width:none] [-ms-overflow-style:none] [&::-webkit-scrollbar]:hidden mask-linear-fade"
        >
          <span
            v-for="tag in overview.service.tags"
            :key="tag"
            flex="inline items-center"
            class="px-2 py-0.5 rounded text-[10px] font-medium bg-zinc-100 dark:bg-zinc-800 text-zinc-600 dark:text-zinc-400 border border-zinc-200 dark:border-zinc-700/50 whitespace-nowrap shrink-0"
          >
            {{ tag }}
          </span>
        </div>

        <div class="flex items-center gap-3 text-xs font-medium text-zinc-500 dark:text-zinc-400 shrink-0 ml-auto pl-2 bg-white dark:bg-zinc-900 shadow-[-8px_0_4px_-4px_rgba(255,255,255,0.8)] dark:shadow-[-8px_0_4px_-4px_rgba(24,24,27,0.8)]">
          <div class="flex items-center gap-1.5" title="实时延迟">
            <div size-13px i-lucide:zap :class="latencyColor" aria-hidden="true" />
            <span>{{ latestLatency }}ms</span>
          </div>
          <div class="w-px h-3 bg-zinc-200 dark:bg-zinc-700" />
          <div class="flex items-center gap-1.5" title="最后更新">
            <div size-13px i-lucide:clock class="text-zinc-600 dark:text-zinc-400" aria-hidden="true" />
            <span class="tabular-nums">{{ lastCheckTime }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Bottom Section: Uptime Percentage and History -->
    <div data-part="visualizer" px-5 pb-5 mt-auto>
      <div pt-2 border="t zinc-100 dark:zinc-800/50">
        <div class="flex justify-between items-center mb-3">
          <div class="flex items-baseline gap-2">
            <span class="text-2xl font-bold text-zinc-800 dark:text-zinc-100">
              {{ uptimePercentage.toFixed(2) }}%
            </span>
            <span class="text-xs font-medium text-zinc-500 dark:text-zinc-400">30天在线率</span>
          </div>
        </div>
        <div flex="~ items-center gap-2px" w-full h-8>
          <HomeUptimeBlock
            v-for="history in overview.histories"
            :key="history.date.toString()"
            :history="history"
          />
        </div>
        <div class="flex justify-between mt-2 text-[10px] text-zinc-400 dark:text-zinc-500">
          <span>30天前</span>
          <span>今天</span>
        </div>
      </div>
    </div>
  </div>
</template>
