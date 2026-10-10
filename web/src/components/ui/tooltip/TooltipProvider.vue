<script setup lang="ts">
import { provide, reactive } from 'vue'
import { tooltipDefaultsKey } from './context'

const props = withDefaults(defineProps<{ openDelay?: number, closeDelay?: number }>(), { openDelay: 400, closeDelay: 100 })
provide(tooltipDefaultsKey, reactive({
  get openDelay() {
    return props.openDelay
  },
  get closeDelay() {
    return props.closeDelay
  },
}))
</script>

<template>
  <slot />
</template>
