<script setup lang="ts">
import { ark } from '@ark-ui/vue/factory'

const props = withDefaults(defineProps<{ label?: string }>(), { label: 'On this page' })
</script>

<template>
  <ark.nav :aria-label="props.label" class="min-w-0 text-size-sm">
    <slot />
  </ark.nav>
</template>
