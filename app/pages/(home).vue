<script setup lang="ts">
const { data } = useQuery({
  key: ['overview'],
  query: () => $fetch('/api/services/overviews'),
})
</script>

<template>
  <HomeOverviewBanner />
  <section>
    <div flex="~ items-center justify-between" mb-6>
      <h2 un-text="sm zinc-900 white" flex="~ items-center gap-2" font-semibold uppercase tracking-wider>
        服务 <span bg="zinc-100 dark:zinc-800" px-2 rounded-full text-10px text-zinc-500>{{ data?.overviews.length ?? 0 }}</span>
      </h2>
    </div>
    <div w-full grid="~ cols-1 md:cols-2 lg:cols-3 gap-6">
      <div
        v-for="overview in data?.overviews"
        :key="overview.service.id"
      >
        <HomeServiceCard :overview="overview" />
      </div>
    </div>
  </section>
</template>
