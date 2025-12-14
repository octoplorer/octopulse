<script setup lang="ts">
const { data, isLoading } = useOverview()
</script>

<template>
  <HomeOverviewBanner />
  <section>
    <div flex="~ items-center justify-between" mb-6>
      <h2 un-text="sm zinc-900 dark:white" flex="~ items-center gap-2" font-semibold uppercase tracking-wider>
        服务 <span bg="zinc-100 dark:zinc-800" px-2 rounded-full text-10px text-zinc-500>{{ data?.overviews.length ?? 0 }}</span>
      </h2>
    </div>
    <div w-full relative grid="~ cols-1 md:cols-2 lg:cols-3 gap-6">
      <template v-if="isLoading">
        <div class="absolute inset-0 z-20 bg-white/50 dark:bg-zinc-950/50 backdrop-blur-[2px] flex items-center justify-center rounded-xl animate-in fade-in duration-200">
          <div
            flex="~ items-center gap-3"
            p="x5 y2.5"
            bg="white dark:zinc-800"
            border="~ zinc-100 dark:zinc-700 rounded-full"
            ring="1 black/5 dark:white/5"
            shadow-lg
          >
            <div i-lucide:loader-circle size-5 un-text="indigo-600 dark:indigo-400" animate-spin />
            <span un-text="sm zinc-500 dark:zinc-300" font-medium>更新数据中...</span>
          </div>
        </div>
      </template>
      <div
        v-if="data?.overviews.length === 0"
        border="~ dashed zinc-300 dark:zinc-700 rounded-xl"
        bg="white dark:zinc-900"
        un-text="center zinc-500"
        col-span-full py-12
      >
        目前没有监控任何服务。请登录后台添加。
      </div>
      <div
        v-for="overview in data?.overviews"
        :key="overview.service.id"
      >
        <HomeServiceCard :overview="overview" />
      </div>
    </div>
  </section>
</template>
