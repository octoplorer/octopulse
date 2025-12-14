<script setup lang="ts">
const { data, isLoading } = useQuery({
  key: ['overview'],
  query: async () => {
    await new Promise(r => setTimeout(r, 1000))
    return $fetch('/api/services/overviews')
  },
})
</script>

<template>
  <HomeOverviewBanner />
  <section>
    <div flex="~ items-center justify-between" mb-6>
      <h2 un-text="sm zinc-900 dark:white" flex="~ items-center gap-2" font-semibold uppercase tracking-wider>
        服务 <span bg="zinc-100 dark:zinc-800" px-2 rounded-full text-10px text-zinc-500>{{ data?.overviews.length ?? 0 }}</span>
      </h2>
    </div>
    <div w-full grid="~ cols-1 md:cols-2 lg:cols-3 gap-6">
      <template v-if="isLoading">
        //TODO
      </template>
      <template v-else>
        <div v-if="data?.overviews.length === 0" class="col-span-full py-12 text-center text-zinc-500 bg-white dark:bg-zinc-900 rounded-xl border border-dashed border-zinc-300 dark:border-zinc-700">
          目前没有监控任何服务。请登录后台添加。
        </div>
        <div
          v-for="overview in data?.overviews"
          :key="overview.service.id"
        >
          <HomeServiceCard :overview="overview" />
        </div>
      </template>
    </div>
  </section>
</template>
