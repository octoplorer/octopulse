<script setup lang="ts" generic="T extends string | undefined">
import type { SelectValueChangeDetails } from '@ark-ui/vue/select'
import type { SelectCollectionItem } from './types'
import { Select as ArkSelect, createListCollection } from '@ark-ui/vue/select'
import { computed, shallowReactive } from 'vue'
import { provideSelectContext } from './context'

const props = defineProps<{
  disabled?: boolean
  name?: string
  required?: boolean
}>()
const emit = defineEmits<{ change: [value: Exclude<T, undefined>] }>()
const value = defineModel<T>({ required: true })
const registeredItems = shallowReactive(new Map<symbol, SelectCollectionItem>())

const collection = computed(() => createListCollection({
  items: Array.from(registeredItems.values()),
  itemToString: item => item.label,
  itemToValue: item => item.value,
  isItemDisabled: item => item.disabled === true,
}))
const selectedValue = computed(() => value.value ? [value.value] : [])

provideSelectContext({
  collection,
  registerItem: (id, item) => registeredItems.set(id, item),
  unregisterItem: id => registeredItems.delete(id),
})

function select(details: SelectValueChangeDetails<SelectCollectionItem>) {
  const nextValue = (details.value[0] ?? '') as Exclude<T, undefined>
  value.value = nextValue as T
  emit('change', nextValue)
}
</script>

<template>
  <ArkSelect.Root
    class="select-root"
    :collection="collection"
    :disabled="props.disabled"
    :model-value="selectedValue"
    :name="props.name"
    :positioning="{
      placement: 'bottom-start',
      gutter: 4,
      fitViewport: true,
      strategy: 'fixed',
    }"
    :required="props.required"
    @value-change="select"
  >
    <slot />
    <ArkSelect.HiddenSelect />
  </ArkSelect.Root>
</template>

<style scoped>
.select-root {
  display: contents;
}
</style>
