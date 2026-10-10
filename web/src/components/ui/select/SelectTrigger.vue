<script setup lang="ts">
import type { SelectTriggerBaseProps } from '@ark-ui/vue/select'
import { Select as ArkSelect } from '@ark-ui/vue/select'
import { useForwardProps } from '@ark-ui/vue/utils'
import { cva } from 'cva'
import { computed, mergeProps, useAttrs } from 'vue'
import { useSelectContext } from './context'

defineOptions({ inheritAttrs: false })
const props = defineProps<SelectTriggerBaseProps & { size?: 'xs' | 'sm' | 'base' | 'lg' }>()
const forwarded = useForwardProps(props)
const attrs = useAttrs()
const context = useSelectContext()
const triggerProps = computed(() => mergeProps({ 'aria-describedby': context.describedBy.value }, forwarded.value, attrs))
const triggerClasses = cva({
  base: 'flex w-full min-w-0 shrink-0 cursor-pointer items-center justify-between gap-1.5 border-0 ring-1 ring-control-border bg-control text-left text-default shadow-control outline-none transition-colors hover:bg-fill-hover data-[state=open]:bg-control focus:ring-focus focus-visible:ring-2 focus-visible:ring-brand! focus-visible:ring-inset data-[invalid]:ring-action-danger data-[invalid]:focus:ring-2 data-[invalid]:focus:ring-action-danger! disabled:cursor-not-allowed disabled:text-inactive disabled:opacity-50',
  variants: { size: { xs: 'h-5 rounded-sm px-1.5 text-size-xs', sm: 'h-6.5 rounded-md px-2 text-size-xs', base: 'h-9 rounded-lg px-3 text-size-base max-sm:h-10 max-sm:text-16px', lg: 'h-10 rounded-lg px-4 text-size-base' } },
  defaultVariants: { size: 'base' },
})
</script>

<template>
  <ArkSelect.Control class="contents">
    <ArkSelect.Trigger v-bind="triggerProps" :class="triggerClasses({ size: size ?? context.size.value })">
      <slot />
      <ArkSelect.Indicator class="flex shrink-0 items-center text-subtle" aria-hidden="true">
        <span class="i-lucide-chevrons-up-down size-4" />
      </ArkSelect.Indicator>
    </ArkSelect.Trigger>
  </ArkSelect.Control>
</template>
