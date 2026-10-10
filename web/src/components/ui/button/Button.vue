<script setup lang="ts">
import type { ButtonHTMLAttributes } from 'vue'
import { ark } from '@ark-ui/vue/factory'
import { computed } from 'vue'
import { buttonVariants } from './variants'

export interface ButtonProps extends /* @vue-ignore */ ButtonHTMLAttributes {
  shape?: 'base' | 'square' | 'circle'
  variant?: 'primary' | 'secondary' | 'ghost' | 'destructive' | 'secondary-destructive' | 'outline'
  size?: 'xs' | 'sm' | 'base' | 'lg'
  asChild?: boolean
  type?: 'button' | 'submit' | 'reset'
  disabled?: boolean
  loading?: boolean
}
const props = withDefaults(defineProps<ButtonProps>(), { shape: 'base', variant: 'secondary', size: 'base', asChild: false, type: 'button' })
const emphasisStyle = computed(() => {
  if (props.variant !== 'primary' && props.variant !== 'destructive')
    return undefined
  const token = props.variant === 'primary' ? 'var(--color-brand)' : 'var(--color-action-danger)'
  const hoverToken = props.variant === 'primary' ? 'var(--color-brand-hover)' : `color-mix(in oklch, ${token}, black 10%)`
  return {
    '--button-emphasis-ring': `color-mix(in oklch, ${token}, black 10%)`,
    '--button-emphasis-bg': token,
    '--button-emphasis-gradient-start': token,
    '--button-emphasis-gradient-end': `color-mix(in oklch, ${token}, black 10%)`,
    '--button-emphasis-hover-start': hoverToken,
    '--button-emphasis-hover-end': `color-mix(in oklch, ${hoverToken}, black 10%)`,
  }
})
function guardActivation(event: MouseEvent) {
  if (props.disabled || props.loading) {
    event.preventDefault()
    event.stopImmediatePropagation()
  }
}
</script>

<template>
  <ark.button :as-child="asChild" :type="asChild ? undefined : type" :disabled="disabled || loading" v-bind="asChild && (disabled || loading) ? { 'aria-disabled': true, tabindex: -1 } : {}" :aria-busy="loading || undefined" :class="buttonVariants({ variant, size, shape })" :style="emphasisStyle" @click="guardActivation">
    <slot v-if="asChild" />
    <template v-else>
      <span v-if="loading" class="i-lucide-loader-circle size-4 animate-spin" aria-hidden="true" /><slot v-else name="icon" /><slot />
    </template>
  </ark.button>
</template>

<style scoped>
.button-emphasis {
  --un-ring-color: var(--button-emphasis-ring);
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

@media (hover: hover) {
  .button-emphasis:not(:disabled):not([aria-disabled='true']):hover::before {
    background: linear-gradient(
      to bottom,
      var(--button-emphasis-hover-start),
      var(--button-emphasis-hover-end)
    );
  }
}
</style>
