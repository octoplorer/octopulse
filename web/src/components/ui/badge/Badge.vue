<script setup lang="ts">
withDefaults(defineProps<{
  variant?: 'outline' | 'success' | 'danger' | 'info' | 'warning' | 'muted'
  dot?: boolean
}>(), { variant: 'outline', dot: false })

const variants = {
  outline: 'pill inline-block border-line py-3px px-7px rounded-5px text-subtle',
  success: 'text-fg-success bg-success-tint',
  danger: 'text-fg-danger bg-danger-tint',
  info: 'text-fg-info bg-info-tint',
  warning: 'text-fg-warning bg-warning-tint',
  muted: 'text-subtle bg-recessed',
}
</script>

<template>
  <span
    :class="[variants[variant], variant !== 'outline' && 'inline-flex state-badge items-center gap-6px font-500 leading-[1.5] py-2px px-7px border-current/12 rounded-6px whitespace-nowrap']"
    un-text="12px"
    border="1 solid"
  ><i v-if="dot" size="5px" rounded="full" bg="current" /><slot /></span>
</template>
