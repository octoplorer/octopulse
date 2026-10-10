<script setup lang="ts">
import { Menu as ArkMenu } from '@ark-ui/vue/menu'
</script>

<template>
  <ArkMenu.TriggerItem class="flex min-h-8 cursor-pointer select-none items-center gap-2 rounded-md px-2 py-1.5 text-size-sm outline-none data-[highlighted]:bg-tint data-[disabled]:pointer-events-none data-[disabled]:opacity-50 group-focus/menu:data-[highlighted]:ring-focus/50 group-focus-visible/menu:data-[highlighted]:ring-2 group-focus-visible/menu:data-[highlighted]:!ring-brand group-focus-visible/menu:data-[highlighted]:ring-inset justify-between">
    <slot /><span class="i-lucide-chevron-right size-4 text-subtle" aria-hidden="true" />
  </ArkMenu.TriggerItem>
</template>
