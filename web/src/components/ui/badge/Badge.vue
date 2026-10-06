<script setup lang="ts">
withDefaults(defineProps<{
  variant?: 'outline' | 'success' | 'danger' | 'info' | 'warning' | 'muted'
  dot?: boolean
}>(), { variant: 'outline', dot: false })

const variants = {
  outline: 'pill inline-block border-line py-3px px-7px rounded-5px text-$muted',
  success: 'text-$success bg-$success-soft',
  danger: 'text-$danger bg-$danger-soft',
  info: 'text-$info bg-$info-soft',
  warning: 'text-$warning bg-$warning-soft',
  muted: 'text-$muted bg-$surface-soft',
}
</script>

<template>
  <span
    :class="[variants[variant], variant !== 'outline' && 'inline-flex state-badge items-center gap-6px font-500 leading-[1.5] py-2px px-7px [border-color:color-mix(in_srgb,currentColor_12%,transparent)] rounded-6px whitespace-nowrap']"
    un-text="12px"
    border="1 solid"
  ><i v-if="dot" size="5px" rounded="full" bg="current" /><slot /></span>
</template>
