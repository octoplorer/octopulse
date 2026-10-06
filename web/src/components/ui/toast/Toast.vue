<script setup lang="ts">
withDefaults(defineProps<{ variant?: 'success' | 'error' }>(), { variant: 'success' })
</script>

<template>
  <div
    class="toast [&[data-variant=error]]:border-l-danger"
    :data-variant="variant"
    flex="~ items-center gap-20px"
    bg="base"
    un-text="14px default"
    border="1 solid line l-3 l-success"
    py="12px"
    px="16px"
    rounded="8px"
    shadow="panel"
  >
    <slot />
  </div>
</template>
