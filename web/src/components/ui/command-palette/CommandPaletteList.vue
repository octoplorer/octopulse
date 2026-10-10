<script setup lang="ts">
import { Combobox as ArkCombobox } from '@ark-ui/vue/combobox'
</script>

<template>
  <ArkCombobox.Content class="max-h-80 overflow-y-auto p-2 outline-none">
    <slot />
  </ArkCombobox.Content>
</template>
