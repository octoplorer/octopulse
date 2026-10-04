<script setup lang="ts">
withDefaults(defineProps<{ as?: keyof HTMLElementTagNameMap }>(), { as: 'span' })
</script>

<template>
  <component :is="as" class="nav-icon" size="18px" shrink="0" un-text="$muted" aria-hidden="true">
    <slot />
  </component>
</template>
