<script setup lang="ts">
import { computed } from 'vue'
import { Button } from '../button'
import { useInputGroup } from './context'

defineProps<{ disabled?: boolean }>()
const group = useInputGroup()
const individual = computed(() => group?.focusMode.value === 'individual')
const size = computed(() => {
  const value = group?.size.value ?? 'base'
  return individual.value ? value : ({ xs: 'xs', sm: 'xs', base: 'sm', lg: 'base' } as const)[value]
})
</script>

<template>
  <Button variant="ghost" :size="size" :disabled="disabled || group?.disabled.value" class="shrink-0 shadow-none! focus:ring-0" :class="individual ? 'focus-visible:ring-0!' : 'mx-1 focus-visible:ring-2 focus-visible:ring-focus'">
    <slot />
  </Button>
</template>
