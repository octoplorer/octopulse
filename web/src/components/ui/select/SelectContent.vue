<script setup lang="ts">
import { useDialogContext } from '@ark-ui/vue/dialog'
import { Select as ArkSelect } from '@ark-ui/vue/select'

defineOptions({ inheritAttrs: false })
const nestedDialog = !!useDialogContext()
</script>

<template>
  <Teleport to="body" :disabled="nestedDialog">
    <ArkSelect.Positioner class="z-220">
      <ArkSelect.Content v-bind="$attrs" class="group/select-content z-220 max-h-[var(--available-height)] max-w-[var(--available-width)] min-w-[var(--reference-width)] overflow-hidden rounded-lg border border-line bg-base text-default shadow-panel outline-none">
        <ArkSelect.List class="group/select-content max-h-inherit overflow-y-auto overscroll-contain scroll-py-2 py-1.5">
          <slot />
        </ArkSelect.List>
      </ArkSelect.Content>
    </ArkSelect.Positioner>
  </Teleport>
</template>
