<script setup lang="ts">
import { Menu as ArkMenu } from '@ark-ui/vue/menu'

const props = defineProps<{ for?: string }>()
</script>

<template>
  <ArkMenu.ItemGroupLabel class="px-2 py-1.5 text-size-xs font-medium text-subtle" :for="props.for">
    <slot />
  </ArkMenu.ItemGroupLabel>
</template>
