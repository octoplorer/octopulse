<script setup lang="ts">
withDefaults(defineProps<{ variant?: 'success' | 'error' }>(), { variant: 'success' })
</script>

<template>
  <div
    class="toast [&[data-variant=error]]:border-l-$danger"
    :data-variant="variant"
    flex="~ items-center gap-20px"
    bg="$surface"
    un-text="14px $text"
    border="1 solid $border l-3 l-$success"
    py="12px"
    px="16px"
    rounded="8px"
    shadow="[0_4px_16px_#00000014]"
  >
    <slot />
  </div>
</template>
