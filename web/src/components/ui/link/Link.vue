<script setup lang="ts">
import { ark } from '@ark-ui/vue/factory'
import { cva } from 'cva'

withDefaults(defineProps<{ variant?: 'default' | 'plain' | 'inline', asChild?: boolean, external?: boolean }>(), { variant: 'default', asChild: false, external: false })
const linkVariants = cva({ base: 'inline-flex items-center gap-1 rounded-sm ', variants: { variant: { default: 'text-link hover:underline underline-offset-3', inline: 'text-link underline underline-offset-3 hover:opacity-80', plain: 'text-default hover:text-link' } } })
</script>

<template>
  <ark.a :as-child="asChild" :class="linkVariants({ variant })" :target="external ? '_blank' : undefined" :rel="external ? 'noopener noreferrer' : undefined">
    <slot /><span v-if="external && !asChild" class="i-lucide-arrow-up-right size-3" aria-hidden="true" />
  </ark.a>
</template>
