<script setup lang="ts">
import type { InputSize } from './styles'
import { Field } from '@ark-ui/vue/field'
import { onBeforeUnmount, onMounted, useTemplateRef, watchPostEffect } from 'vue'
import { inputClasses, inputSizes } from './styles'

const props = withDefaults(defineProps<{ size?: InputSize, variant?: 'default' | 'error', autoResize?: boolean, minRows?: number, maxRows?: number }>(), { size: 'base', variant: 'default', autoResize: false, minRows: 3 })
const value = defineModel<string>()
const control = useTemplateRef('control')
let native: HTMLTextAreaElement | undefined
let originalHeight = ''
let originalOverflow = ''
function restore() {
  if (!native)
    return
  native.style.height = originalHeight
  native.style.overflowY = originalOverflow
}
function resize() {
  if (!native)
    return
  if (!props.autoResize) {
    restore()
    return
  }
  const style = getComputedStyle(native)
  const rawLine = Number.parseFloat(style.lineHeight)
  const line = Number.isFinite(rawLine) ? style.lineHeight.endsWith('px') ? rawLine : rawLine * Number.parseFloat(style.fontSize) : 24
  const padding = (Number.parseFloat(style.paddingTop) || 0) + (Number.parseFloat(style.paddingBottom) || 0)
  native.style.height = 'auto'
  const min = Math.max(1, props.minRows) * line + padding
  const max = props.maxRows ? Math.max(props.minRows, props.maxRows) * line + padding : Infinity
  const height = Math.max(min, Math.min(max, native.scrollHeight))
  native.style.height = `${height}px`
  native.style.overflowY = native.scrollHeight > max ? 'auto' : 'hidden'
}
onMounted(() => {
  native = control.value?.$el as HTMLTextAreaElement | undefined
  originalHeight = native?.style.height ?? ''
  originalOverflow = native?.style.overflowY ?? ''
  resize()
})
watchPostEffect(() => {
  void value.value
  void props.autoResize
  void props.minRows
  void props.maxRows
  resize()
})
onBeforeUnmount(restore)
</script>

<template>
  <Field.Textarea ref="control" v-model="value" :rows="minRows" class="h-auto py-2 leading-relaxed" :class="[inputClasses, inputSizes[size], autoResize ? 'min-h-0 resize-none' : 'min-h-24 resize-y', variant === 'error' && 'ring-action-danger focus:ring-action-danger']" v-bind="variant === 'error' ? { 'aria-invalid': true } : {}" @input="resize" />
</template>
