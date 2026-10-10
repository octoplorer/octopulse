<script setup lang="ts">
import { cva } from 'cva'
import { provide, toRefs } from 'vue'
import { inputGroupKey } from './context'

const props = withDefaults(defineProps<{ size?: 'xs' | 'sm' | 'base' | 'lg', disabled?: boolean, focusMode?: 'container' | 'individual' }>(), { size: 'base', disabled: false, focusMode: 'container' })
const { size, disabled, focusMode } = toRefs(props)
provide(inputGroupKey, { size, disabled, focusMode })
const inputGroupVariants = cva({
  base: 'ui-control-group relative flex min-w-0 items-center gap-0 rounded-lg bg-control [&_.ui-input]:shadow-none [&_.ui-input]:outline-none [&_.ui-input]:h-full! [&_.ui-input]:min-h-0 [&_.ui-input]:rounded-none',
  variants: {
    size: { xs: 'h-5 text-size-xs max-sm:h-7', sm: 'h-6.5 text-size-xs max-sm:h-8', base: 'h-9 text-size-base max-sm:h-10', lg: 'h-10 text-size-base' },
    focusMode: {
      container: 'overflow-hidden ring-1 ring-control-border focus-within:ring-2 focus-within:ring-focus [&_.ui-input]:ring-0! [&_.ui-input]:border-0 [&_.ui-input]:bg-transparent [&_button:focus-visible]:ring-2 [&_button:focus-visible]:ring-focus',
      individual: 'isolate overflow-visible shadow-none ring-0 [&>*]:h-full! [&>*]:relative [&>*]:rounded-none! [&>*]:border-1! [&>*]:border-solid [&>*]:border-control-border [&>*]:ring-0! [&>*:first-child]:rounded-l-[inherit]! [&>*:last-child]:rounded-r-[inherit]! [&>*:not(:first-child)]:-ml-px [&>*:hover]:z-1 [&_.ui-input:focus]:border-focus [&_.ui-input:focus]:z-2 [&_button:focus-visible]:border-focus [&_button:focus]:z-2',
    },
  },
})
</script>

<template>
  <div :data-focus-mode="focusMode" :class="[inputGroupVariants({ size, focusMode }), disabled && 'opacity-50']">
    <slot />
  </div>
</template>
