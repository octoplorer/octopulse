<script setup lang="ts">
import { ark } from '@ark-ui/vue/factory'

withDefaults(defineProps<{
  variant?: 'default' | 'primary' | 'danger' | 'ghost'
  size?: 'default' | 'sm' | 'icon'
  asChild?: boolean
}>(), { variant: 'default', size: 'default', asChild: false })

const variants = {
  default: '',
  primary: 'primary [&[data-variant=primary]]:[background:linear-gradient(#ffffff12,#ffffff00),var(--accent)] [&[data-variant=primary]]:[border-color:color-mix(in_srgb,var(--accent),#000_12%)] [&[data-variant=primary]]:text-white [&[data-variant=primary]]:[box-shadow:inset_0_1px_0_#ffffff26,0_1px_2px_#0000001a] [&[data-variant=primary]:hover]:[background:var(--accent-hover)]',
  danger: 'danger [&[data-variant=danger]]:text-$danger [&[data-variant=danger]]:[border-color:color-mix(in_srgb,var(--danger)_25%,var(--border))]',
  ghost: 'ghost [&[data-variant=ghost]]:bg-transparent [&[data-variant=ghost]]:border-transparent [&[data-variant=ghost]]:shadow-none [&[data-variant=ghost]:hover]:bg-transparent',
}
const sizes = {
  default: '',
  sm: 'small [&[data-size=sm]]:min-h-28px [&[data-size=sm]]:py-4px [&[data-size=sm]]:px-8px [&[data-size=sm]]:rounded-6px [&[data-size=sm]]:text-12px',
  icon: 'icon-button [&[data-size=icon]]:gap-0 [&[data-size=icon]]:[font-size:inherit] [&[data-size=icon]]:[font-weight:inherit] [&[data-size=icon]]:[line-height:inherit] [&[data-size=icon]]:[white-space:inherit] [&[data-size=icon]]:shrink-0 [&[data-size=icon]]:min-w-32px [&[data-size=icon]]:min-h-32px [&[data-size=icon]]:p-7px [&[data-size=icon]]:rounded-6px [&[data-size=icon]]:border-transparent [&[data-size=icon]]:bg-transparent [&[data-size=icon]]:text-$muted [&[data-size=icon]]:shadow-none [&[data-size=icon]:hover]:bg-$surface-soft [&[data-size=icon]:hover]:text-$text',
}
</script>

<template>
  <ark.button
    class="inline-flex button hover:bg-$surface-soft [&:active:not(:disabled)]:shadow-none"
    :class="[variants[variant], sizes[size]]"
    :as-child="asChild"
    :data-variant="variant"
    :data-size="size"
    items="center"
    justify="center"
    gap="6px"
    min-h="36px"
    p="x-12px y-7px"
    border="1 solid $control-border"
    rounded="8px"
    bg="$surface"
    un-text="14px $text"
    font="500"
    leading="[1.4]"
    whitespace="nowrap"
    shadow="$control-shadow"
  >
    <slot />
  </ark.button>
</template>
