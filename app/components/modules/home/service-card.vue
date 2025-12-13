<script setup lang="ts">
import type { SerializeObject } from 'nitropack'

defineProps<{
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
</script>

<template>
  <div
    p-5
    class="bg-white dark:bg-zinc-900 rounded-xl shadow-sm border border-zinc-200 dark:border-zinc-800 overflow-hidden hover:shadow-md transition-all duration-300"
  >
    <div flex="~ justify-between items-start" mb-6 data-part="header">
      <div class="flex items-center space-x-3">
        <div class="relative">
          <div
            class="size-3 rounded-full"
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
        :class=" getStatusColor(overview.service.status)"
      >
        {{ getStatusText(overview.service.status) }}
      </span>
    </div>
    <div data-part="visualizer">
      <div class="flex justify-between items-center mb-3">
        <div class="flex items-baseline gap-2">
          <span class="text-2xl font-bold text-zinc-800 dark:text-zinc-100">100%</span>
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
</template>
