<script setup lang="ts">
import type { ComboboxItem, ComboboxOption } from '../combobox'
import { computed } from 'vue'
import { Combobox } from '../combobox'

defineOptions({ inheritAttrs: false })
defineProps<{ items: ComboboxItem[], label?: string, disabled?: boolean, readOnly?: boolean }>()
const value = defineModel<string>({ default: '' })
const open = defineModel<boolean>('open', { default: false })
const selected = computed(() => '')
function select(items: ComboboxOption[]) {
  if (items[0])
    value.value = items[0].label
}
</script>

<template>
  <Combobox v-bind="$attrs" v-model:open="open" v-model:input-value="value" :model-value="selected" :items="items" :label="label" :disabled="disabled" :read-only="readOnly" free-text @select="select">
    <template v-for="(_, name) in $slots" #[name]="slotProps">
      <slot :name="name" v-bind="slotProps" />
    </template>
  </Combobox>
</template>
