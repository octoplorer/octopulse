<script setup lang="ts">
withDefaults(defineProps<{
  as?: string
  variant?: 'default' | 'destructive' | 'validation'
}>(), { as: 'div', variant: 'default' })

const variants = {
  default: 'note py-13px px-16px border-line bg-$surface-soft text-$muted text-12px leading-[1.7]',
  destructive: 'error-banner flex items-center justify-between gap-14px bg-$danger-soft text-$danger py-14px px-16px [border-color:color-mix(in_srgb,var(--danger)_25%,var(--border))] text-13px mb-18px',
  validation: 'validation-error [border-color:color-mix(in_srgb,var(--danger)_25%,var(--border))] bg-$danger-soft text-$danger py-12px px-15px text-13px my-15px mx-0',
}
</script>

<template>
  <component
    :is="as"
    :class="variants[variant]"
    :role="variant === 'default' ? undefined : 'alert'"
    border="1 solid"
    rounded="8px"
  >
    <slot />
  </component>
</template>
