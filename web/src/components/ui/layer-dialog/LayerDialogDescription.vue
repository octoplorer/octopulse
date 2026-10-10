<script setup lang="ts">
import { Dialog as ArkDialog } from '@ark-ui/vue/dialog'
</script>

<template>
  <ArkDialog.Description class="mt-1 text-size-sm text-subtle">
    <slot />
  </ArkDialog.Description>
</template>
