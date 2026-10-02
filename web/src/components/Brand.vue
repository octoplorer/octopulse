<script setup lang="ts">
import { Activity } from '@lucide/vue'

defineProps<{ compact?: boolean }>()
</script>

<template>
  <span class="brand" un-flex="~ items-center gap-2.5"><span class="brand-icon"><Activity :size="21" :stroke-width="2.4" /></span><span v-if="!compact">octopulse<span class="brand-dot">.</span></span></span>
</template>
