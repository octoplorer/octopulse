<script setup lang="ts">
defineProps<{ title: string, description?: string }>()
</script>

<template>
  <div class="empty-state">
    <span class="empty-icon"><span class="i-lucide-activity" un-w="26px" un-h="26px" aria-hidden="true" /></span>
    <h3>{{ title }}</h3>
    <p v-if="description">
      {{ description }}
    </p>
    <slot />
  </div>
</template>
