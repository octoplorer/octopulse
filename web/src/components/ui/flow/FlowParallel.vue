<script setup lang="ts">
import FlowGroup from './FlowGroup.vue'

defineProps<{ align?: 'end' }>()
</script>

<template>
  <FlowGroup kind="parallel" :align="align">
    <slot />
  </FlowGroup>
</template>
