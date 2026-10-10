<script setup lang="ts">
import type { CommandPaletteItemData } from './types'
import { Dialog as ArkDialog } from '@ark-ui/vue/dialog'
import { useId } from 'vue'
import CommandPalettePanel from './CommandPalettePanel.vue'

defineOptions({ inheritAttrs: false })

const props = withDefaults(defineProps<{ items: CommandPaletteItemData[], title?: string, label?: string, placeholder?: string, emptyLabel?: string, loading?: boolean, closeOnSelect?: boolean }>(), { title: 'Command palette', closeOnSelect: true })
const emit = defineEmits<{ select: [item: CommandPaletteItemData, options: { newTab: boolean }] }>()
const open = defineModel<boolean>('open', { default: false })
const query = defineModel<string>('query', { default: '' })
const inputId = useId()
const initialFocus = () => document.getElementById(inputId)
function select(item: CommandPaletteItemData, options: { newTab: boolean }) {
  emit('select', item, options)
  if (props.closeOnSelect)
    open.value = false
}
</script>

<template>
  <ArkDialog.Root v-model:open="open" lazy-mount unmount-on-exit :initial-focus-el="initialFocus">
    <Teleport to="body">
      <ArkDialog.Backdrop class="fixed inset-0 z-150 bg-backdrop" /><ArkDialog.Positioner class="fixed inset-0 z-151 flex items-start justify-center px-4 pb-4 pt-[min(20vh,8rem)]">
        <ArkDialog.Content v-bind="$attrs" class="z-151 w-full max-w-xl overflow-hidden rounded-xl bg-base shadow-panel outline-none ring-1 ring-line">
          <ArkDialog.Title class="sr-only">
            {{ props.title }}
          </ArkDialog.Title><CommandPalettePanel v-model:query="query" :input-id="inputId" :items="props.items" :label="props.label" :placeholder="props.placeholder" :empty-label="props.emptyLabel" :loading="props.loading" @select="select">
            <template v-if="$slots.default" #default="slotProps">
              <slot v-bind="slotProps" />
            </template><template v-if="$slots.item" #item="slotProps">
              <slot name="item" v-bind="slotProps" />
            </template><template v-if="$slots.footer" #footer>
              <slot name="footer" />
            </template><template #input-end>
              <ArkDialog.CloseTrigger class="rounded px-1.5 py-0.5 text-size-xs text-subtle outline-none ring-1 ring-line focus:ring-focus/50 focus-visible:ring-2 focus-visible:ring-brand" aria-label="Close command palette">
                Esc
              </ArkDialog.CloseTrigger>
            </template>
          </CommandPalettePanel>
        </ArkDialog.Content>
      </ArkDialog.Positioner>
    </Teleport>
  </ArkDialog.Root>
</template>
