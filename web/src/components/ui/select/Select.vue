<script setup lang="ts" generic="T extends string | string[] | undefined = string">
import type { SelectValueChangeDetails } from '@ark-ui/vue/select'
import type { SelectCollectionItem } from './types'
import { Select as ArkSelect, createListCollection } from '@ark-ui/vue/select'
import { computed, shallowReactive, shallowRef, useId } from 'vue'
import { provideSelectContext } from './context'

const props = withDefaults(defineProps<{
  disabled?: boolean
  readOnly?: boolean
  invalid?: boolean
  name?: string
  form?: string
  required?: boolean
  multiple?: boolean
  modelValue?: T
  defaultValue?: T
  label?: string
  description?: string
  error?: string
  size?: 'xs' | 'sm' | 'base' | 'lg'
}>(), { disabled: undefined, readOnly: undefined, invalid: undefined, required: undefined, size: 'base' })
const emit = defineEmits<{ 'change': [value: Exclude<T, undefined>], 'update:modelValue': [value: Exclude<T, undefined>] }>()
const open = defineModel<boolean>('open', { default: false })
const initial = shallowRef(props.defaultValue)
const id = useId()
const registeredItems = shallowReactive(new Map<symbol, SelectCollectionItem>())
const collection = computed(() => createListCollection({
  items: Array.from(registeredItems.values()),
  itemToString: item => item.label,
  itemToValue: item => item.value,
  isItemDisabled: item => item.disabled === true,
}))
const selectedValue = computed(() => {
  const current = props.modelValue ?? initial.value
  return Array.isArray(current) ? current : current ? [current] : []
})
provideSelectContext({
  collection,
  size: computed(() => props.size),
  describedBy: computed(() => [props.description && `${id}-description`, props.error && `${id}-error`].filter(Boolean).join(' ') || undefined),
  registerItem: (id, item) => registeredItems.set(id, item),
  unregisterItem: id => registeredItems.delete(id),
})
function select(details: SelectValueChangeDetails<SelectCollectionItem>) {
  const next = (props.multiple ? details.value : details.value[0] ?? '') as Exclude<T, undefined>
  initial.value = next as T
  emit('update:modelValue', next)
  emit('change', next)
}
</script>

<template>
  <ArkSelect.Root
    v-model:open="open" :class="label || description || error || $slots.label ? 'flex min-w-0 flex-col gap-1.5' : 'contents'"
    :collection="collection" :disabled="disabled" :read-only="readOnly" :invalid="invalid || !!error || undefined"
    :model-value="selectedValue" :name="name" :form="form" :multiple="multiple" :close-on-select="!multiple"
    :positioning="{ placement: 'bottom-start', gutter: 4, fitViewport: true, strategy: 'fixed', sameWidth: true }"
    :required="required" @value-change="select"
  >
    <ArkSelect.Label v-if="label || $slots.label" class="text-size-base font-medium text-strong">
      <slot name="label">
        {{ label }}
      </slot>
    </ArkSelect.Label>
    <slot />
    <ArkSelect.HiddenSelect />
    <p v-if="description" :id="`${id}-description`" class="text-size-sm text-subtle">
      {{ description }}
    </p>
    <p v-if="error" :id="`${id}-error`" class="text-size-sm text-fg-danger" role="alert">
      {{ error }}
    </p>
  </ArkSelect.Root>
</template>
