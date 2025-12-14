<script setup lang="ts">
import 'temporal-polyfill/global'

useHead({
  titleTemplate(title) {
    if (!title)
      return `NuxtPulse`
    return `${title} - NuxtPulse`
  },
})
</script>

<template>
  <NuxtRouteAnnouncer />
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>
</template>
