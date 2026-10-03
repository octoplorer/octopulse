<script setup lang="ts">
defineProps<{ compact?: boolean }>()
</script>

<template>
  <span class="brand" un-flex="~ items-center gap-2.5"><span class="brand-icon"><span class="i-lucide-activity" un-w="18px" un-h="18px" aria-hidden="true" /></span><span v-if="!compact" class="brand-name">octopulse</span></span>
</template>
