<script setup lang="ts">
import { Menu as ArkMenu } from '@ark-ui/vue/menu'

const model = defineModel<string>({ default: '' })
</script>

<template>
  <ArkMenu.RadioItemGroup v-model="model" class="flex flex-col">
    <slot />
  </ArkMenu.RadioItemGroup>
</template>
