<script setup lang="ts">
import { Combobox as ArkCombobox } from '@ark-ui/vue/combobox'

const props = defineProps<{ for?: string }>()
</script>

<template>
  <ArkCombobox.ItemGroupLabel :for="props.for" class="px-3 py-2 text-size-xs font-semibold text-subtle">
    <slot />
  </ArkCombobox.ItemGroupLabel>
</template>
