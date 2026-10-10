<script setup lang="ts">
import type { ButtonProps } from './Button.vue'
import Button from './Button.vue'

defineProps<{ href: string, variant?: ButtonProps['variant'], size?: ButtonProps['size'], external?: boolean }>()
</script>

<template>
  <Button as-child :variant="variant" :size="size">
    <a :href="href" :target="external ? '_blank' : undefined" :rel="external ? 'noopener noreferrer' : undefined"><slot /></a>
  </Button>
</template>
