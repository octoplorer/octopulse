<script setup lang="ts">
import { Button } from '../button'

const props = defineProps<{ variant?: 'primary' | 'secondary' | 'destructive', loading?: boolean, disabled?: boolean }>()
</script>

<template>
  <Button :variant="props.variant ?? 'primary'" :disabled="props.disabled || props.loading" :aria-busy="props.loading || undefined">
    <span v-if="props.loading" class="i-lucide-loader-circle size-4 animate-spin" aria-hidden="true" /><slot />
  </Button>
</template>
