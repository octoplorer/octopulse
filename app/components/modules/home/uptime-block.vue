<script setup lang="ts">
import type { SerializeObject } from 'nitropack'
import { TooltipArrow, TooltipContent, TooltipPortal, TooltipProvider, TooltipRoot, TooltipTrigger } from 'reka-ui'

const props = defineProps<{
  history: SerializeObject<ServiceHistroy>
}>()

const uptimeBgColor = computed(() => {
  if (props.history.uptime === null) {
    return 'bg-gray-400'
  }
  if (props.history.uptime >= 99.9) {
    return 'bg-emerald-500'
  }
  else if (props.history.uptime > 98) {
    return 'bg-amber-400'
  }
  else {
    return 'bg-rose-500'
  }
})

const uptimeTextColor = computed(() => {
  if (props.history.uptime === null) {
    return 'text-gray-400'
  }
  if (props.history.uptime >= 99.9) {
    return 'text-emerald-500'
  }
  else if (props.history.uptime > 98) {
    return 'text-amber-400'
  }
  else {
    return 'text-rose-500'
  }
})

const latencyColor = computed(() => {
  if (props.history.latency === null) {
    return 'text-gray-400'
  }
  if (props.history.latency > 800) {
    return 'text-rose-500'
  }
  else if (props.history.latency > 500) {
    return 'text-amber-400'
  }
  else {
    return 'text-emerald-500'
  }
})

const outages = computed<string[]>(() => {
  return props.history.outages.map((outage) => {
    const from = Temporal.ZonedDateTime.from(outage.from)
      .withTimeZone(Intl.DateTimeFormat().resolvedOptions().timeZone)
    const duration = Temporal.Duration.from(outage.duration)
    const to = from.add(duration)

    const fromString = from.toLocaleString('zh-CN', {
      timeStyle: 'long',
      hourCycle: 'h24',
    })

    const endString = to.toLocaleString('zh-CN', {
      timeStyle: 'long',
      hourCycle: 'h24',
    })

    return `${fromString} ~ ${endString}`
  })
})

const datetime = computed(() => {
  const instant = Temporal.ZonedDateTime.from(props.history.date)
    .withTimeZone(Intl.DateTimeFormat().resolvedOptions().timeZone)
  return new Date(instant.epochMilliseconds)
})
</script>

<template>
  <TooltipProvider>
    <TooltipRoot>
      <TooltipTrigger as-child>
        <div class="group relative flex-1 h-8 min-w-[4px] mx-[1px] first:ml-0 last:mr-0">
          <div
            class="w-full h-full rounded-sm opacity-80 hover:opacity-100 transition-all duration-200 cursor-pointer hover:scale-110"
            :class="uptimeBgColor"
          />
        </div>
      </TooltipTrigger>
      <Teleport to="#teleports">
        <TooltipPortal>
          <TooltipContent
            relative
            class="bg-white dark:bg-zinc-700 text-white text-xs rounded-lg py-2 px-3 shadow-xl border border-zinc-200 dark:border-zinc-600"
          >
            <div flex="~ col">
              <p className="font-semibold pb-1 mb-1.5" border="b zinc-700 dark:zinc-600">
                <NuxtTime
                  :datetime="datetime"
                  year="numeric"
                  month="long"
                  day="numeric"
                />
              </p>

              <div flex="~ items-center gap-2" mb-1>
                <div v-if="history.uptime" flex="~ items-center gap-2">
                  <span class="text-zinc-900 dark:text-zinc-300" i-lucide:clock-arrow-up size-4 />
                  <span class="font-medium" :class="uptimeTextColor">{{ history.uptime.toFixed(2) }}%</span>
                </div>

                <div v-if="history.latency" flex="~ items-center gap-2">
                  <span class="text-zinc-900 dark:text-zinc-300" i-lucide:zap size-4 />
                  <span class="font-medium" :class="latencyColor">{{ history.latency }}ms</span>
                </div>

                <div v-else>
                  <span class="text-gray-400">无数据</span>
                </div>
              </div>

              <div v-if="outages.length > 0" mt-2 pt-2 border="t zinc-700 dark:zinc-600">
                <div class="flex justify-between items-center mb-1">
                  <span class="text-rose-600 dark:text-rose-300 font-medium">中断时段</span>
                </div>

                <div flex="~ col gap-1">
                  <span
                    v-for="(outage, idx) in outages"
                    :key="idx"
                    bg="rose-50/50 dark:rose-950/50"
                    border="~ rose-100/30 dark:rose-900/30 rounded"
                    un-text="rose-800 dark:rose-200"
                    class="px-1.5 py-0.5 text-10px font-mono"
                  >
                    {{ outage }}
                  </span>
                </div>
              </div>
            </div>

            <TooltipArrow class="fill-zinc-200! dark:fill-zinc-600!" />
          </TooltipContent>
        </TooltipPortal>
      </Teleport>
    </TooltipRoot>
  </TooltipProvider>
</template>
