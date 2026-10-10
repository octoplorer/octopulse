<script setup lang="ts">
withDefaults(defineProps<{ variant?: 'default' | 'bordered', mobileDivider?: boolean }>(), { variant: 'default' })
</script>

<template>
  <div class="min-w-0" :class="[variant === 'bordered' && 'border border-solid border-line rounded-lg p-4', mobileDivider && 'max-md:border-b max-md:border-line max-md:pb-4']">
    <slot />
  </div>
</template>
