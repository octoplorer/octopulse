<script setup lang="ts">
import { Dialog as ArkDialog } from '@ark-ui/vue/dialog'
import { computed, provide } from 'vue'
import { layerDialogKey } from './context'
import LayerDialogContent from './LayerDialogContent.vue'

defineOptions({ inheritAttrs: false })

const props = withDefaults(defineProps<{ defaultOpen?: boolean, title?: string, description?: string, dismissDisabled?: boolean, alert?: boolean, size?: 'sm' | 'base' | 'lg' | 'xl', verticalAlign?: 'top' | 'center', closeLabel?: string }>(), { dismissDisabled: false, closeLabel: 'Close dialog' })
const open = defineModel<boolean | undefined>('open', { default: undefined })
function onOpenChange(details: { open: boolean }) {
  if (!details.open && props.dismissDisabled)
    return
  open.value = details.open
}
provide(layerDialogKey, { dismissDisabled: computed(() => props.dismissDisabled), close: () => {
  if (!props.dismissDisabled)
    open.value = false
} })
</script>

<template>
  <ArkDialog.Root :open="open" :default-open="props.defaultOpen" :role="props.alert ? 'alertdialog' : 'dialog'" :modal="true" :close-on-escape="!props.dismissDisabled" :close-on-interact-outside="!props.dismissDisabled && !props.alert" lazy-mount unmount-on-exit @open-change="onOpenChange">
    <LayerDialogContent v-if="props.title" v-bind="$attrs" :title="props.title" :description="props.description" :size="props.size" :vertical-align="props.verticalAlign" :close-label="props.closeLabel">
      <slot /><template v-if="$slots.footer" #footer>
        <slot name="footer" />
      </template>
    </LayerDialogContent>
    <slot v-else />
  </ArkDialog.Root>
</template>
