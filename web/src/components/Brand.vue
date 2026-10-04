<script setup lang="ts">
defineProps<{ compact?: boolean }>()
</script>

<template>
  <span class="brand" flex="~ items-center gap-2.5" font="[var(--font-sans)] 600" un-text="20px" tracking="-0.6px"><span class="brand-icon" flex="~ items-center justify-center" size="32px" shrink="0" un-text="$text" bg="$surface" border="1 solid $control-border" rounded="8px" shadow="$control-shadow"><span class="i-lucide-activity" w="18px" h="18px" aria-hidden="true" /></span><span v-if="!compact" class="brand-name">octopulse</span></span>
</template>
