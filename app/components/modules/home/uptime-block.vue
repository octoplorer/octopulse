<script setup lang="ts">
import type { SerializeObject } from 'nitropack'

import { Tooltip } from '@ark-ui/vue/tooltip'

const props = defineProps<{
  history: SerializeObject<ServiceHistroy>
}>()

const uptimeColor = computed(() => {
  if (props.history.uptime === null) {
    return 'bg-gray-400'
  }
  if (props.history.uptime === 100) {
    return 'bg-emerald-500'
  }
  else if (props.history.uptime > 98) {
    return 'bg-amber-400'
  }
  else {
    return 'bg-rose-500'
  }
})

const outages = computed<string[]>(() => {
  // TODO
  return []
})

const datetime = computed(() => {
  const instant = Temporal.Instant.from(props.history.date)
  return new Date(instant.epochMilliseconds)
})
</script>

<template>
  <Tooltip.Root lazy-mount unmount-on-exit>
    <Tooltip.Trigger class="group relative flex-1 h-8 min-w-[4px] mx-[1px] first:ml-0 last:mr-0">
      <div
        class="w-full h-full rounded-sm opacity-80 hover:opacity-100 transition-all duration-200 cursor-pointer hover:scale-110"
        :class="uptimeColor"
      />
    </Tooltip.Trigger>
    <Teleport to="#teleports">
      <Tooltip.Positioner>
        <Tooltip.Content
          space-y-1.5 divide="y zinc-200 dark:zinc-600"
          class="bg-white dark:bg-zinc-700 text-white text-xs rounded-lg py-2 px-3 shadow-xl border border-zinc-200 dark:border-zinc-600"
        >
          <Tooltip.Arrow>
            <Tooltip.ArrowTip />
          </Tooltip.Arrow>

          <p class="text-zinc-900 dark:text-zinc-300 font-semibold" pb-1.5>
            <NuxtTime
              :datetime="datetime"
              year="2-digit" month="long" day="numeric"
            />
          </p>

          <div v-if="history.uptime" flex="~ items-center gap-2" pb-1.5>
            <span class="text-zinc-900 dark:text-zinc-300">在线率:</span>
            <span class="font-medium text-emerald-400">{{ history.uptime }}%</span>
          </div>

          <div v-else pb-1.5>
            <span class="text-gray-400">无数据</span>
          </div>

          <div v-if="outages.length > 0" pb-1.5>
            <div class="flex justify-between items-center mb-1">
              <span class="text-rose-300 dark:text-rose-300 font-medium">中断时段</span>
            </div>

            <div flex="~ col gap-1">
              <span
                v-for="(outage, idx) in outages"
                :key="idx"
                class="bg-rose-950/40 border border-rose-900/30 text-rose-200 px-1.5 py-0.5 rounded text-[10px] font-mono"
              >
                {{ outage }}
              </span>
            </div>
          </div>
        </Tooltip.Content>
      </Tooltip.Positioner>
    </Teleport>
  </Tooltip.Root>
</template>
