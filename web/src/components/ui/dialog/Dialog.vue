<script setup lang="ts">
import { Dialog as ArkDialog } from '@ark-ui/vue/dialog'
import { Button } from '../button'

defineOptions({ inheritAttrs: false })

withDefaults(defineProps<{ title: string, description?: string, wide?: boolean, closeLabel?: string, theme?: 'light' | 'dark' }>(), { closeLabel: 'Close dialog' })

const open = defineModel<boolean>('open', { default: false })
</script>

<template>
  <ArkDialog.Root v-model:open="open" :lazy-mount="true" :unmount-on-exit="true">
    <Teleport to="body">
      <ArkDialog.Backdrop class="modal-backdrop" position="fixed" inset="0" bg="[#00000060]" z="100" /><ArkDialog.Positioner position="fixed" inset="0" flex="~ items-center justify-center" p="24px" z="101" overflow-y="auto" class="[@media(max-width:700px)]:p-15px">
        <ArkDialog.Content
          v-bind="$attrs"
          data-theme-boundary
          class="modal-content [&.wide]:max-w-900px [@media(max-width:700px)]:p-21px"
          bg="$surface"
          un-text="$text"
          p="24px"
          border="~ solid $control-border"
          rounded="12px"
          w="full"
          max-w="520px"
          max-h="90vh"
          overflow="auto"
          shadow="[0_16px_48px_#00000033,0_4px_12px_#00000014]"
          :class="{ wide }"
          :data-theme="theme"
        >
          <div flex="~ justify-between items-start gap-4">
            <div>
              <ArkDialog.Title class="block modal-title" un-text="20px" font="600" tracking="-0.4px" mb="6px">
                {{ title }}
              </ArkDialog.Title><ArkDialog.Description v-if="description" un-text="13px $muted">
                {{
                  description
                }}
              </ArkDialog.Description>
            </div>
            <ArkDialog.CloseTrigger as-child>
              <Button size="icon" :aria-label="closeLabel">
                <span class="i-lucide-x" w="20px" h="20px" aria-hidden="true" />
              </Button>
            </ArkDialog.CloseTrigger>
          </div>
          <div class="modal-body" mt="24px">
            <slot />
          </div>
          <div v-if="$slots.footer" class="modal-footer" border="t-1px t-solid t-line" pt="20px" mt="24px" flex="~ justify-end gap-9px">
            <slot name="footer" />
          </div>
        </ArkDialog.Content>
      </ArkDialog.Positioner>
    </Teleport>
  </ArkDialog.Root>
</template>
