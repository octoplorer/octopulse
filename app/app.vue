<script setup lang="ts">
import 'temporal-polyfill/global'

useHead({
  titleTemplate(title) {
    if (!title)
      return `Octopulse`
    return `${title} - Octopulse`
  },
})
</script>

<template>
  <NuxtRouteAnnouncer />
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>
</template>
