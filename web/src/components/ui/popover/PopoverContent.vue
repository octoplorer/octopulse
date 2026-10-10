<script setup lang="ts">
import { useDialogContext } from '@ark-ui/vue/dialog'
import { Popover as ArkPopover } from '@ark-ui/vue/popover'

defineOptions({ inheritAttrs: false })
const props = withDefaults(defineProps<{ portalled?: boolean, arrow?: boolean }>(), { portalled: true })
const nestedDialog = !!useDialogContext()
</script>

<template>
  <Teleport to="body" :disabled="nestedDialog || !props.portalled">
    <ArkPopover.Positioner class="z-200">
      <ArkPopover.Content v-bind="$attrs" class="relative z-200 max-w-[var(--available-width)] rounded-xl bg-base p-4 text-default shadow-panel outline-none ring-1 ring-line">
        <ArkPopover.Arrow v-if="props.arrow" class="[--arrow-size:8px] [--arrow-background:var(--color-base)]">
          <ArkPopover.ArrowTip class="border-l border-t border-line" />
        </ArkPopover.Arrow>
        <slot />
      </ArkPopover.Content>
    </ArkPopover.Positioner>
  </Teleport>
</template>
