<script setup lang="ts">
import { Dialog as ArkDialog } from '@ark-ui/vue/dialog'
</script>

<template>
  <ArkDialog.Title class="text-size-lg font-semibold text-strong">
    <slot />
  </ArkDialog.Title>
</template>
