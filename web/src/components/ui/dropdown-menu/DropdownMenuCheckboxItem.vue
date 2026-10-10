<script setup lang="ts">
import { Menu as ArkMenu } from '@ark-ui/vue/menu'

const props = defineProps<{ value: string, disabled?: boolean, closeOnSelect?: boolean }>()
const checked = defineModel<boolean>('checked', { default: false })
</script>

<template>
  <ArkMenu.CheckboxItem v-model:checked="checked" :value="props.value" :disabled="props.disabled" :close-on-select="props.closeOnSelect ?? false" class="flex min-h-8 cursor-pointer select-none items-center gap-2 rounded-md px-2 py-1.5 text-size-sm outline-none data-[highlighted]:bg-tint data-[disabled]:pointer-events-none data-[disabled]:opacity-50 group-focus/menu:data-[highlighted]:ring-focus/50 group-focus-visible/menu:data-[highlighted]:ring-2 group-focus-visible/menu:data-[highlighted]:!ring-brand group-focus-visible/menu:data-[highlighted]:ring-inset">
    <span class="flex size-4 items-center justify-center"><ArkMenu.ItemIndicator><span class="i-lucide-check size-4" aria-hidden="true" /></ArkMenu.ItemIndicator></span>
    <ArkMenu.ItemText class="flex-1">
      <slot />
    </ArkMenu.ItemText>
  </ArkMenu.CheckboxItem>
</template>
