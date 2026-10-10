<script setup lang="ts">
import { Combobox as ArkCombobox } from '@ark-ui/vue/combobox'

const props = defineProps<{ id?: string }>()
</script>

<template>
  <ArkCombobox.ItemGroup :id="props.id" class="py-1">
    <slot />
  </ArkCombobox.ItemGroup>
</template>
