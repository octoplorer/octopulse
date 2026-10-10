<script setup lang="ts">
import { computed } from 'vue'
import { Button } from '../button'
import { useToolbarSize } from './context'

const props = defineProps<{ disabled?: boolean, loading?: boolean, pressed?: boolean, variant?: 'primary' | 'secondary' | 'ghost' | 'destructive', shape?: 'base' | 'square' | 'circle' }>()
const size = useToolbarSize()
const controlSize = computed(() => size?.value ?? 'base')
</script>

<template>
  <Button data-toolbar-control :size="controlSize" :variant="props.variant ?? 'ghost'" :shape="props.shape" :disabled="props.disabled || props.loading" :aria-busy="props.loading || undefined" :aria-pressed="props.pressed" tabindex="-1">
    <span v-if="props.loading" class="i-lucide-loader-circle size-4 animate-spin" aria-hidden="true" /><slot />
  </Button>
</template>
