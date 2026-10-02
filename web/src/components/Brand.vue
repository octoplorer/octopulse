<script setup lang="ts">
defineProps<{ compact?: boolean }>()
</script>

<template>
  <span class="brand" un-flex="~ items-center gap-2.5"><span class="brand-icon"><span class="i-lucide-activity" un-w="21px" un-h="21px" aria-hidden="true" /></span><span v-if="!compact">octopulse<span class="brand-dot">.</span></span></span>
</template>
