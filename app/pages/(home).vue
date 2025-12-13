<script setup lang="ts">
const { data } = useQuery({
  key: ['overview'],
  query: () => $fetch('/api/services/overviews'),
})
</script>

<template>
  <div>
    <div v-for="overview in data?.overviews" :key="overview.service.id">
      {{ overview }}
    </div>
  </div>
</template>
