<script setup lang="ts">
import { Dialog as ArkDialog } from '@ark-ui/vue/dialog'
import { Button } from '../button'
import DialogContent from './DialogContent.vue'

defineOptions({ inheritAttrs: false })
withDefaults(defineProps<{
  title: string
  description?: string
  size?: 'sm' | 'base' | 'lg' | 'xl'
  role?: 'dialog' | 'alertdialog'
  wide?: boolean
  closeLabel?: string
}>(), { size: 'base', role: 'dialog', closeLabel: 'Close dialog' })
const open = defineModel<boolean>('open', { default: false })
</script>

<template>
  <ArkDialog.Root v-model:open="open" :role="role" lazy-mount unmount-on-exit>
    <DialogContent v-bind="$attrs" :size="wide ? 'xl' : size" class="overflow-hidden!">
      <div class="flex shrink-0 items-start justify-between gap-4 p-5 pb-0 sm:p-6 sm:pb-0">
        <div class="min-w-0 space-y-2">
          <ArkDialog.Title class="text-size-xl font-semibold tracking-tight [overflow-wrap:anywhere]">
            {{ title }}
          </ArkDialog.Title>
          <ArkDialog.Description v-if="description" class="text-size-sm text-subtle">
            {{ description }}
          </ArkDialog.Description>
        </div>
        <ArkDialog.CloseTrigger as-child>
          <Button shape="square" size="sm" variant="ghost" class="shrink-0" :aria-label="closeLabel">
            <span class="i-lucide-x size-4" aria-hidden="true" />
          </Button>
        </ArkDialog.CloseTrigger>
      </div>
      <div class="min-h-0 overflow-auto overscroll-contain p-5 sm:p-6">
        <slot />
      </div>
      <div v-if="$slots.footer" class="flex shrink-0 flex-wrap justify-end gap-2 border-t border-line bg-recessed p-4">
        <slot name="footer" />
      </div>
    </DialogContent>
  </ArkDialog.Root>
</template>
