<script setup lang="ts" generic="T">
import { onMounted, ref, toRaw, watch } from 'vue'
import { errorText } from '../lib/errors'
import { parseJSON } from '../lib/form'
import { FieldTextarea } from './ui/field'

defineOptions({ inheritAttrs: false })

const props = defineProps<{
  modelValue: T
  label: string
}>()
const emit = defineEmits<{
  'update:modelValue': [value: T]
  'validityChange': [message: string]
}>()

function format(value: T) {
  return JSON.stringify(value, null, 2) || ''
}

const text = ref(format(props.modelValue))
let emittedValue: T | undefined

watch(() => props.modelValue, (value) => {
  if (toRaw(value) === emittedValue) {
    emittedValue = undefined
    return
  }
  text.value = format(value)
  emit('validityChange', '')
})
onMounted(() => emit('validityChange', ''))

function update(textValue: string) {
  text.value = textValue
  try {
    const value = parseJSON<T>(text.value, props.label)
    emittedValue = toRaw(value)
    emit('validityChange', '')
    emit('update:modelValue', value)
  }
  catch (e) {
    emit('validityChange', errorText(e))
  }
}
</script>

<template>
  <FieldTextarea v-bind="$attrs" :model-value="text" @update:model-value="update" />
</template>
