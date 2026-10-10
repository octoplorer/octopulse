<script setup lang="ts">
import { Menu as ArkMenu } from '@ark-ui/vue/menu'

const props = defineProps<{ id?: string }>()
</script>

<template>
  <ArkMenu.ItemGroup :id="props.id" class="flex flex-col">
    <slot />
  </ArkMenu.ItemGroup>
</template>
