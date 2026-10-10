<script setup lang="ts">
import { Menu as ArkMenu } from '@ark-ui/vue/menu'

const props = withDefaults(defineProps<{ value: string, disabled?: boolean, destructive?: boolean, asChild?: boolean, closeOnSelect?: boolean }>(), { closeOnSelect: undefined })
</script>

<template>
  <ArkMenu.Item class="flex min-h-8 cursor-pointer select-none items-center gap-2 rounded-md px-2 py-1.5 text-size-sm data-[highlighted]:bg-tint data-[disabled]:pointer-events-none data-[disabled]:opacity-50 group-focus/menu:data-[highlighted]:ring-focus/50 group-focus-visible/menu:data-[highlighted]:ring-2 group-focus-visible/menu:data-[highlighted]:!ring-brand group-focus-visible/menu:data-[highlighted]:ring-inset" :value="props.value" :disabled="props.disabled" :as-child="props.asChild" :tabindex="-1" :close-on-select="props.closeOnSelect" :class="[props.asChild ? undefined : 'outline-none', props.destructive ? 'text-fg-danger data-[highlighted]:bg-danger-tint' : 'text-default']">
    <slot />
  </ArkMenu.Item>
</template>
