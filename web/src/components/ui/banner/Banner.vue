<script setup lang="ts">
import { cva } from 'cva'

withDefaults(defineProps<{ variant?: 'default' | 'alert' | 'error' | 'secondary', size?: 'sm' | 'base', title?: string, description?: string }>(), { variant: 'default', size: 'base' })
const bannerVariants = cva({
  base: 'flex min-w-0 w-full flex-wrap gap-3 rounded-lg',
  variants: { variant: { default: 'bg-info-tint text-fg-info', alert: 'bg-warning-tint text-fg-warning', error: 'bg-danger-tint text-fg-danger', secondary: 'bg-recessed text-subtle' }, size: { sm: 'items-center px-3 py-2 text-size-sm', base: 'items-start px-4 py-3 text-size-base' } },
})
const icons = { default: 'i-lucide-info', alert: 'i-lucide-triangle-alert', error: 'i-lucide-circle-alert', secondary: 'i-lucide-info' }
</script>

<template>
  <div :class="bannerVariants({ variant, size })" :role="variant === 'error' || variant === 'alert' ? 'alert' : 'status'">
    <slot name="icon">
      <span class="mt-0.5 size-4 shrink-0" :class="[icons[variant]]" aria-hidden="true" />
    </slot>
    <div class="min-w-[min(100%,12rem)] flex-1 [overflow-wrap:anywhere]">
      <strong v-if="title" class="block font-medium">{{ title }}</strong><p v-if="description" class="mt-1 text-size-sm">
        {{ description }}
      </p><slot />
    </div>
    <div v-if="$slots.actions" class="ms-auto flex max-w-full flex-wrap items-center gap-2">
      <slot name="actions" />
    </div>
  </div>
</template>
