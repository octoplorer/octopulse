<script setup lang="ts">
import type { CommandPaletteItemData } from './types'
import { Combobox as ArkCombobox } from '@ark-ui/vue/combobox'

const props = defineProps<{ item: CommandPaletteItemData }>()
</script>

<template>
  <ArkCombobox.Item :item="props.item" persist-focus class="group flex w-full items-center gap-3 rounded-lg px-3 py-2 text-size-sm text-default text-left outline-none transition-colors" :class="props.item.disabled ? 'cursor-default opacity-50' : 'cursor-pointer hover:bg-overlay data-[highlighted]:bg-overlay'">
    <ArkCombobox.ItemText class="min-w-0 flex-1">
      <slot>{{ props.item.label }}</slot>
    </ArkCombobox.ItemText>
  </ArkCombobox.Item>
</template>
