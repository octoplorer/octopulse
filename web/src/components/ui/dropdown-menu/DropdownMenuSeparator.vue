<script setup lang="ts">
import { Menu as ArkMenu } from '@ark-ui/vue/menu'
</script>

<template>
  <ArkMenu.Separator class="my-1 h-px border-0 bg-line">
    <slot />
  </ArkMenu.Separator>
</template>
