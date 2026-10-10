<script setup lang="ts">
import type { CreateToasterReturn } from '@ark-ui/vue/toast'
import { Toast as ArkToast, Toaster } from '@ark-ui/vue/toast'
import { provide } from 'vue'
import { Button } from '../button'
import { toastManagerKey, useKumoToastManager } from './manager'

const props = withDefaults(defineProps<{ manager?: CreateToasterReturn, closeLabel?: string }>(), { closeLabel: 'Close notification' })
const manager = props.manager ?? useKumoToastManager()
provide(toastManagerKey, manager)
</script>

<template>
  <slot />
  <Teleport to="body">
    <Toaster v-slot="toast" :toaster="manager" class="z-500">
      <ArkToast.Root class="kumo-toast relative flex w-90 max-w-[calc(100vw-2rem)] items-start gap-3 rounded-xl border border-solid border-line bg-base p-4 shadow-panel text-default">
        <span class="mt-0.5 size-4 shrink-0" :class="{ 'i-lucide-circle-check text-fg-success': toast.type === 'success', 'i-lucide-circle-alert text-fg-danger': toast.type === 'error', 'i-lucide-loader-circle animate-spin text-subtle': toast.type === 'loading', 'i-lucide-info text-fg-info': toast.type === 'info', 'i-lucide-triangle-alert text-fg-warning': toast.type === 'warning' }" aria-hidden="true" />
        <div class="min-w-0 flex-1">
          <ArkToast.Title class="text-size-sm font-medium">
            {{ toast.title }}
          </ArkToast.Title><ArkToast.Description v-if="toast.description" class="mt-1 text-size-sm text-subtle">
            {{ toast.description }}
          </ArkToast.Description><ArkToast.ActionTrigger v-if="toast.action" as-child>
            <Button size="sm" class="mt-2">
              {{ toast.action.label }}
            </Button>
          </ArkToast.ActionTrigger>
        </div>
        <ArkToast.CloseTrigger as-child>
          <Button variant="ghost" size="xs" shape="square" :aria-label="closeLabel">
            <span class="i-lucide-x size-3.5" aria-hidden="true" />
          </Button>
        </ArkToast.CloseTrigger>
      </ArkToast.Root>
    </Toaster>
  </Teleport>
</template>

<style scoped>
.kumo-toast {
  translate: var(--x) var(--y);
  scale: var(--scale);
  opacity: var(--opacity);
  z-index: var(--z-index);
  height: var(--height);
  transition:
    translate 200ms,
    opacity 200ms,
    scale 200ms,
    height 200ms;
}
</style>
