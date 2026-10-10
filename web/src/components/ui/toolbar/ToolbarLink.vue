<script setup lang="ts">
import { ark } from '@ark-ui/vue/factory'

const props = defineProps<{ href: string, disabled?: boolean }>()
</script>

<template>
  <ark.a data-toolbar-control :href="props.disabled ? undefined : props.href" :aria-disabled="props.disabled || undefined" tabindex="-1" class="relative inline-flex h-8 items-center gap-2 rounded-md px-2 text-size-sm text-default outline-none hover:bg-tint aria-disabled:pointer-events-none aria-disabled:opacity-50" :class="!props.disabled ? 'focus:z-2 focus:ring-focus/50 focus-visible:z-2 focus-visible:ring-2 focus-visible:ring-brand' : undefined">
    <slot />
  </ark.a>
</template>
