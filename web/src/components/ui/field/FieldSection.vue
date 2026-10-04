<script setup lang="ts">
withDefaults(defineProps<{ as?: string }>(), { as: 'div' })
</script>

<template>
  <component :is="as" class="form-section last:border-b-0 [&_h2]:text-15px [&_h2]:mb-5px [&>p]:mb-20px" p="25px" border="b-1 solid $border">
    <slot />
  </component>
</template>
