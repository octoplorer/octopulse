<script setup lang="ts">
import type { ButtonHTMLAttributes, StyleValue, VNode } from 'vue'
import { ark } from '@ark-ui/vue/factory'
import { computed, normalizeClass as normalize } from 'vue'

export interface ButtonProps extends /* @vue-ignore */ ButtonHTMLAttributes {
  shape?: 'base' | 'square' | 'circle'
  variant?: 'primary' | 'secondary' | 'ghost' | 'destructive' | 'secondary-destructive' | 'outline'
  size?: 'xs' | 'sm' | 'base' | 'lg'
  asChild?: boolean
}

const {
  shape = 'base',
  variant = 'secondary',
  size = 'base',
  asChild = false,
  ...restProps
} = defineProps<ButtonProps>()

defineSlots<{
  default: () => VNode[]
  icon: () => VNode[]
}>()

const isCompactShape = computed(() => shape === 'square' || shape === 'circle')

const isEmphasis = computed(() => variant === 'primary' || variant === 'destructive')

const emphasisStyle = computed<StyleValue>(() => {
  if (!isEmphasis.value)
    return undefined
  const token = (variant === 'primary') ? 'var(--color-brand)' : 'var(--color-danger)'
  return {
    '--button-emphasis-ring': `color-mix(in oklch, ${token}, black 10%)`,
    '--button-emphasis-bg': `color-mix(in oklch, ${token}, white 30%)`,
    '--button-emphasis-gradient-start': `color-mix(in oklch, ${token}, white 15%)`,
    '--button-emphasis-gradient-end': token,
  }
})
</script>

<template>
  <ark.button
    class="group"
    cursor="pointer disabled:not-allowed"
    border="0"
    outline="focus:none!"
    font="medium"
    select-none
    :opacity="normalize({
      'disabled:50': isEmphasis,
    })"
    :un-text="normalize([
      'disabled:subtle',
      {
        '!on-brand': variant === 'primary',
        '!on-danger': variant === 'destructive',
        '!default disabled:!default/70': variant === 'secondary',
        'default': ['ghost', 'outline'].includes(variant),
        '!fg-danger not-disabled:hover:!fg-danger disabled:!fg-danger/70': variant === 'secondary-destructive',
        'not-disabled:hover:strong': variant === 'outline',
      },
    ])"
    :flex="normalize([
      '~ shrink-0 items-center',
      isCompactShape && 'justify-center',
    ])"
    :p="isCompactShape ? '0' : normalize({
      'x-1.5': size === 'xs',
      'x-2': size === 'sm',
      'x-3': size === 'base',
      'x-4': size === 'lg',
    })"
    :rounded="normalize([
      shape !== 'circle' && size === 'xs' && 'sm',
      shape !== 'circle' && size === 'sm' && 'md',
      shape !== 'circle' && ['base', 'lg'].includes(size) && 'lg',
      shape === 'circle' && 'full',
    ])"
    :ring="normalize([
      'focus:focus/50 focus-visible:2 focus-visible:brand',
      {
        '~': isEmphasis,
        '~ line': ['secondary', 'secondary-destructive', 'outline'].includes(variant),
        'not-disabled:hover:danger/30': variant === 'secondary-destructive',
        'not-disabled:hover:focus/25': variant === 'outline',
      },
    ])"
    :gap="asChild && isEmphasis ? '1.5' : normalize({
      '1': ['xs', 'sm'].includes(size),
      '1.5': size === 'base',
      '2': size === 'lg',
    })"
    :h="normalize({
      '3.5': isCompactShape && size === 'xs',
      '5': !isCompactShape && size === 'xs',
      '6.5': size === 'sm',
      '9': size === 'base',
      '10': size === 'lg',
    })"
    :w="isCompactShape ? normalize({
      '3.5': size === 'xs',
      '6.5': size === 'sm',
      '9': size === 'base',
      '10': size === 'lg',
    }) : 'max'"
    :bg="normalize({
      'base disabled:base/50 data-[state=open]:base': ['secondary', 'secondary-destructive'].includes(variant),
      'not-disabled:hover:tint': variant === 'secondary',
      'inherit hover:tint': variant === 'ghost',
      'transparent': variant === 'outline',
    })"
    :class="normalize([
      ['xs', 'sm'].includes(size) ? 'text-size-xs' : 'text-size-base',
      isEmphasis && 'button-emphasis',
      variant === 'outline' && 'transition-colors',
    ])"
    :style="emphasisStyle"
    :shadow="variant === 'ghost' ? 'none' : 'control'"
    :as-child
    v-bind="restProps"
  >
    <template v-if="asChild">
      <slot />
    </template>
    <template v-else-if="isEmphasis">
      <span relative flex="~ items-center gap-1.5">
        <slot name="icon" />
        <span contents><slot /></span>
      </span>
    </template>
    <template v-else>
      <slot name="icon" />
      <span contents><slot /></span>
    </template>
  </ark.button>
</template>

<style scoped>
.button-emphasis,
.button-emphasis:focus,
.button-emphasis:focus-visible,
.button-emphasis:active {
  --un-ring-color: var(--button-emphasis-ring);
}

.button-emphasis {
  position: relative;
  isolation: isolate;
  overflow: hidden;
  background: var(--button-emphasis-bg);
}

.button-emphasis::before {
  position: absolute;
  z-index: -1;
  inset: 0;
  pointer-events: none;
  border-radius: inherit;
  background: linear-gradient(
    to bottom,
    var(--button-emphasis-gradient-start),
    var(--button-emphasis-gradient-end)
  );
  box-shadow: inset 0 1px 0 0 var(--button-emphasis-bg);
  content: '';
}

.button-emphasis:hover::before {
  background: linear-gradient(
    to bottom,
    var(--button-emphasis-bg),
    var(--button-emphasis-gradient-end)
  );
}
</style>
