<script setup lang="ts">
const { data } = useQuery({
  key: ['overview'],
  query: () => $fetch('/api/services/overviews'),
})
</script>

<template>
  <div w-full grid="~ cols-1 md:cols-2 lg:cols-3 gap-6">
    <div
      v-for="overview in data?.overviews"
      :key="overview.service.id"
    >
      <HomeServiceCard :overview="overview" />
    </div>
  </div>
</template>
