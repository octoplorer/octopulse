<script setup lang="ts">
import { Dialog as ArkDialog } from '@ark-ui/vue/dialog'
import { Button } from '../button'

defineOptions({ inheritAttrs: false })

withDefaults(defineProps<{ title: string, description?: string, wide?: boolean, closeLabel?: string }>(), { closeLabel: 'Close dialog' })

const open = defineModel<boolean>('open', { default: false })
</script>

<template>
  <ArkDialog.Root v-model:open="open" :lazy-mount="true" :unmount-on-exit="true">
    <Teleport to="body">
      <ArkDialog.Backdrop class="modal-backdrop" position="fixed" inset="0" bg="backdrop" z="100" /><ArkDialog.Positioner position="fixed" inset="0" flex="~ items-center justify-center" p="24px" z="101" overflow-y="auto" class="[@media(max-width:700px)]:p-15px">
        <ArkDialog.Content
          v-bind="$attrs"
          class="modal-content [&.wide]:max-w-900px [@media(max-width:700px)]:p-21px"
          bg="base"
          un-text="default"
          p="24px"
          border="~ solid line"
          rounded="12px"
          w="full"
          max-w="520px"
          max-h="90vh"
          overflow="auto"
          shadow="panel"
          :class="{ wide }"
        >
          <div flex="~ justify-between items-start gap-4">
            <div>
              <ArkDialog.Title class="block modal-title" un-text="20px" font="600" tracking="-0.4px" mb="6px">
                {{ title }}
              </ArkDialog.Title><ArkDialog.Description v-if="description" un-text="13px subtle">
                {{
                  description
                }}
              </ArkDialog.Description>
            </div>
            <ArkDialog.CloseTrigger as-child>
              <Button shape="square" size="sm" variant="ghost" :aria-label="closeLabel">
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
