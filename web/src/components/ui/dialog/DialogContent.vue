<script setup lang="ts">
import { Dialog as ArkDialog } from '@ark-ui/vue/dialog'
import { cva } from 'cva'

defineOptions({ inheritAttrs: false })
withDefaults(defineProps<{ size?: 'sm' | 'base' | 'lg' | 'xl' }>(), { size: 'base' })
const contentVariants = cva({
  base: 'z-101 flex w-full min-w-0 flex-col max-h-[calc(100dvh-2rem)] sm:max-h-[calc(100dvh-4rem)] overflow-auto rounded-xl bg-base text-default shadow-panel ring-1 ring-line outline-none',
  variants: { size: { sm: 'max-w-72', base: 'max-w-96', lg: 'max-w-[32rem]', xl: 'max-w-[48rem]' } },
})
</script>

<template>
  <Teleport to="body">
    <ArkDialog.Backdrop class="fixed inset-0 z-100 bg-backdrop" />
    <ArkDialog.Positioner class="fixed inset-0 z-101 flex items-center justify-center overflow-y-auto overscroll-contain p-4 sm:p-8">
      <ArkDialog.Content v-bind="$attrs" :class="contentVariants({ size })">
        <slot />
      </ArkDialog.Content>
    </ArkDialog.Positioner>
  </Teleport>
</template>
