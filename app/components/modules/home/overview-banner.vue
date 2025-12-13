<script setup lang="ts">
const { data } = useQuery({
  key: ['overview'],
  query: () => $fetch('/api/services/overviews'),
})

const isSystemHealthy = computed(() => data.value?.overviews.every(overview => overview.service.status === ServiceStatus.Operational))

const activeServices = computed(() => data.value?.overviews.filter(overview => overview.service.status === ServiceStatus.Operational).length)
const totalServices = computed(() => data.value?.overviews.length ?? 0)

const averageUptime = computed(() => {
  const uptimes = data.value?.overviews.reduce((acc, overview) => {
    const serviceUptime = overview.histories.reduce((acc, history) => acc + (history.uptime ?? 0), 0) ?? 0
    return acc + serviceUptime
  }, 0) ?? 0

  return uptimes / totalServices.value
})
</script>

<template>
  <div mb-12>
    <div
      relative of-hidden
      border="~ black/5 dark:zinc-800 rounded-2xl"
      bg="white/70 dark:[#18181b]/80" p="6 sm:8"
      class="backdrop-blur-xl shadow-[0_8px_30px_rgba(0,0,0,0.04)] dark:shadow-none"
    >
      <div
        class="absolute top-0 right-0 w-96 h-96 bg-gradient-to-br to-transparent dark:to-transparent rounded-bl-full pointer-events-none opacity-50"
        :class="isSystemHealthy ? 'from-emerald-500/5 dark:from-emerald-500/10' : 'from-rose-500/5 dark:from-rose-500/10'"
      />

      <div
        relative
        flex="~ col sm:row items-start sm:items-center justify-between gap-6"
      >
        <div flex="~ items-start gap-5">
          <div
            relative size-12
            flex="~ shrink-0 items-center justify-center"
            bg="white dark:zinc-900" border="~ zinc-100 dark:zinc-800 rounded-xl"
            class="shadow-[0_2px_10px_rgba(0,0,0,0.05)] dark:shadow-inner"
          >
            <div v-if="isSystemHealthy" size-7 stroke-width-3 i-lucide:check />
            <div v-else size-7 stroke-width-3 i-lucide:alert-circle />
            <span
              absolute inset-0 rounded-xl ring-1 ring-inset
              :class="isSystemHealthy ? 'ring-emerald-500/20' : 'ring-rose-500/20'"
            />
          </div>
          <div>
            <h2 un-text="xl zinc-900 dark:white" font-bold tracking-tight>
              {{ isSystemHealthy ? '系统运行正常' : '系统存在故障' }}
            </h2>
            <p un-text="sm zinc-500" mt-1 max-w-md leading-relaxed>
              {{ isSystemHealthy ? '所有服务运行正常。没有出现故障。' : '部分服务出现故障或性能下降。' }}
            </p>
          </div>
        </div>

        <div class="flex items-center gap-8 sm:border-l border-black/5 dark:border-zinc-800 sm:pl-8">
          <div space-y-0.5>
            <div un-text="2xl zinc-900 dark:white" font-mono font-bold tracking-tight>
              {{ averageUptime }}%
            </div>
            <div class="text-10px font-semibold text-zinc-400 uppercase tracking-wider">
              Uptime (30d)
            </div>
          </div>
          <div space-y-0.5>
            <div un-text="2xl zinc-900 dark:white" font-mono font-bold tracking-tight>
              {{ activeServices }}/{{ totalServices }}
            </div>
            <div class="text-10px font-semibold text-zinc-400 uppercase tracking-wider">
              Services Active
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
