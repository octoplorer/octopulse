<script setup lang="ts">
import type { PopoverRootProps } from '@ark-ui/vue/popover'
import { Popover as ArkPopover } from '@ark-ui/vue/popover'

defineOptions({ inheritAttrs: false })
const props = withDefaults(defineProps<{
  defaultOpen?: boolean
  modal?: boolean
  placement?: NonNullable<PopoverRootProps['positioning']>['placement']
  closeOnEscape?: boolean
  closeOnInteractOutside?: boolean
}>(), { placement: 'bottom-start', closeOnEscape: true, closeOnInteractOutside: true })
const open = defineModel<boolean | undefined>('open', { default: undefined })
</script>

<template>
  <ArkPopover.Root v-bind="$attrs" v-model:open="open" :default-open="props.defaultOpen" :modal="props.modal" :close-on-escape="props.closeOnEscape" :close-on-interact-outside="props.closeOnInteractOutside" :positioning="{ placement: props.placement, gutter: 8, strategy: 'fixed' }">
    <slot />
  </ArkPopover.Root>
</template>
