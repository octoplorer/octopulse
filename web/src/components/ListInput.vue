<script setup lang="ts" generic="T">
import { ref, toRaw, watch } from 'vue'
import { splitValues } from '../lib/form'
import { FieldInput, FieldTextarea } from './ui/field'

defineOptions({ inheritAttrs: false })

const props = defineProps<{
  modelValue: T[]
  parseItem: (value: string) => T
  formatItem?: (value: T) => string
  separator?: 'values' | 'lines'
  trim?: boolean
  multiline?: boolean
}>()
const emit = defineEmits<{
  'update:modelValue': [value: T[]]
}>()

function format(value: T[]) {
  return value.map(props.formatItem || String).join(props.separator === 'lines' ? '\n' : ', ')
}

const text = ref(format(props.modelValue))
let emittedValue: T[] | undefined

watch(() => props.modelValue, (value) => {
  if (toRaw(value) === emittedValue) {
    emittedValue = undefined
    return
  }
  text.value = format(value)
})

function update(value: string) {
  text.value = value
  const items = props.separator === 'lines'
    ? text.value.split('\n').map(value => props.trim ? value.trim() : value).filter(Boolean)
    : splitValues(text.value)
  emittedValue = items.map(props.parseItem)
  emit('update:modelValue', emittedValue)
}
</script>

<template>
  <FieldTextarea v-if="multiline" v-bind="$attrs" :model-value="text" @update:model-value="update" />
  <FieldInput v-else v-bind="$attrs" :model-value="text" @update:model-value="update" />
</template>
