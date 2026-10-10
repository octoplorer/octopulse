<script setup lang="ts">
withDefaults(defineProps<{ as?: string }>(), { as: 'div' })
</script>

<template>
  <component :is="as" class="p-5 sm:p-6 last:border-b-0 [&_h2]:text-15px [&_h2]:mb-5px [&>p]:mb-20px" border="b-1 solid line">
    <slot />
  </component>
</template>
