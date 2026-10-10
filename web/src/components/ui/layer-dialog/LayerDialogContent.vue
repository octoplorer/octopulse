<script setup lang="ts">
import { Dialog as ArkDialog } from '@ark-ui/vue/dialog'
import { cva } from 'cva'
import { Button } from '../button'
import { useLayerDialog } from './context'

defineOptions({ inheritAttrs: false })
const props = withDefaults(defineProps<{ title?: string, description?: string, size?: 'sm' | 'base' | 'lg' | 'xl', verticalAlign?: 'top' | 'center', closeLabel?: string, showCloseButton?: boolean, portalled?: boolean }>(), { size: 'base', verticalAlign: 'center', closeLabel: 'Close dialog', showCloseButton: true, portalled: true })
const context = useLayerDialog()
const contentStyles = cva({ base: 'relative z-101 flex max-h-[calc(100dvh-2rem)] w-full flex-col overflow-hidden rounded-t-2xl bg-base text-default shadow-panel ring-1 ring-line sm:rounded-2xl', variants: { size: { sm: 'sm:max-w-md', base: 'sm:max-w-xl', lg: 'sm:max-w-2xl', xl: 'sm:max-w-3xl' } }, defaultVariants: { size: 'base' } })
</script>

<template>
  <Teleport to="body" :disabled="!props.portalled">
    <ArkDialog.Backdrop class="fixed inset-0 z-100 bg-backdrop" />
    <ArkDialog.Positioner class="fixed inset-0 z-101 flex items-end justify-center p-0 pt-8 sm:p-6" :class="props.verticalAlign === 'top' ? 'sm:items-start sm:pt-16' : 'sm:items-center'">
      <ArkDialog.Content v-bind="$attrs" :class="contentStyles({ size: props.size })">
        <div v-if="props.title || $slots.header" class="flex items-start justify-between gap-4 border-b border-line px-6 py-4">
          <div class="min-w-0">
            <slot name="header">
              <ArkDialog.Title class="text-size-lg font-semibold text-strong">
                {{ props.title }}
              </ArkDialog.Title><ArkDialog.Description v-if="props.description" class="mt-1 text-size-sm text-subtle">
                {{ props.description }}
              </ArkDialog.Description>
            </slot>
          </div>
          <ArkDialog.CloseTrigger v-if="props.showCloseButton" as-child>
            <Button variant="ghost" shape="square" size="sm" :aria-label="props.closeLabel" :disabled="context?.dismissDisabled.value">
              <span class="i-lucide-x size-4" aria-hidden="true" />
            </Button>
          </ArkDialog.CloseTrigger>
        </div>
        <div class="min-h-0 overflow-y-auto px-6 py-5">
          <slot />
        </div>
        <div v-if="$slots.footer" class="flex shrink-0 flex-wrap items-center justify-end gap-2 border-t border-line bg-tint px-6 py-4">
          <slot name="footer" />
        </div>
      </ArkDialog.Content>
    </ArkDialog.Positioner>
  </Teleport>
</template>
