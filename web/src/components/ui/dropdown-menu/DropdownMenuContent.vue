<script setup lang="ts">
import { useDialogContext } from '@ark-ui/vue/dialog'
import { Menu as ArkMenu } from '@ark-ui/vue/menu'

defineOptions({ inheritAttrs: false })
const props = withDefaults(defineProps<{ portalled?: boolean }>(), { portalled: true })
const dialog = useDialogContext()
</script>

<template>
  <Teleport to="body" :disabled="!props.portalled || !!dialog">
    <ArkMenu.Positioner class="z-220">
      <ArkMenu.Content v-bind="$attrs" class="group/menu z-220 min-w-44 max-w-[var(--available-width)] max-h-[var(--available-height)] overflow-y-auto rounded-lg bg-base p-1.5 text-default shadow-panel outline-none ring-1 ring-line">
        <slot />
      </ArkMenu.Content>
    </ArkMenu.Positioner>
  </Teleport>
</template>
