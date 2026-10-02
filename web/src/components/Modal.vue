<script setup lang="ts">
import { Dialog } from '@ark-ui/vue/dialog'
import { X } from '@lucide/vue'
import { dark, t } from '../lib/preferences'
const open = defineModel<boolean>('open', { default: false })
defineProps<{ title: string; description?: string; wide?: boolean }>()
</script>
<template>
  <Dialog.Root v-model:open="open" :lazy-mount="true" :unmount-on-exit="true"
    ><Teleport to="body"
      ><Dialog.Backdrop class="modal-backdrop" /><Dialog.Positioner class="modal-positioner"
        ><Dialog.Content
          class="modal-content"
          :class="{ wide }"
          :data-theme="dark ? 'dark' : 'light'"
          ><div un-flex="~ justify-between items-start gap-4">
            <div>
              <Dialog.Title class="modal-title">{{ title }}</Dialog.Title
              ><Dialog.Description v-if="description" class="muted">{{
                description
              }}</Dialog.Description>
            </div>
            <Dialog.CloseTrigger class="icon-button" :aria-label="t('关闭对话框', 'Close dialog')"
              ><X :size="20"
            /></Dialog.CloseTrigger>
          </div>
          <div class="modal-body"><slot /></div>
          <div v-if="$slots.footer" class="modal-footer">
            <slot name="footer" /></div></Dialog.Content></Dialog.Positioner></Teleport
  ></Dialog.Root>
</template>
