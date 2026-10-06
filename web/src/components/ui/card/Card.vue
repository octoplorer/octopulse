<script setup lang="ts">
withDefaults(defineProps<{ as?: string }>(), { as: 'div' })
</script>

<template>
  <component :is="as" class="card border-1 border-solid border-line [&_details]:border-t-1 [&_details]:border-solid [&_details]:border-line" min-w="0" overflow="hidden" bg="$surface" rounded="10px" shadow="$shadow">
    <slot />
  </component>
</template>
