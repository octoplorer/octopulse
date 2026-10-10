<script setup lang="ts">
import { useFieldContext } from '@ark-ui/vue/field'
import { RadioGroup as ArkRadio } from '@ark-ui/vue/radio-group'
import { computed, provide } from 'vue'
import { radioAppearanceKey } from './context'

const props = withDefaults(defineProps<{
  label?: string
  legend?: string
  description?: string
  name?: string
  form?: string
  disabled?: boolean
  readOnly?: boolean
  invalid?: boolean
  orientation?: 'horizontal' | 'vertical'
  appearance?: 'default' | 'card' | 'segmented'
  controlPosition?: 'start' | 'end'
}>(), { orientation: 'vertical', appearance: 'default', controlPosition: 'start', disabled: undefined, readOnly: undefined, invalid: undefined })
const field = useFieldContext()
const isDisabled = computed(() => props.disabled ?? field?.value.disabled)
const isReadOnly = computed(() => props.readOnly ?? field?.value.readOnly)
const value = defineModel<string>()
provide(radioAppearanceKey, computed(() => ({ appearance: props.appearance, controlPosition: props.controlPosition })))
function onKeydown(event: KeyboardEvent) {
  if (event.defaultPrevented || isDisabled.value || isReadOnly.value || !['ArrowLeft', 'ArrowRight', 'ArrowUp', 'ArrowDown'].includes(event.key))
    return
  const target = event.target
  if (!(target instanceof HTMLInputElement) || target.type !== 'radio')
    return
  const root = event.currentTarget as HTMLElement
  const inputs = Array.from(root.querySelectorAll<HTMLInputElement>('input[type="radio"]')).filter(input => !input.disabled)
  const index = inputs.indexOf(target)
  if (index < 0 || inputs.length === 0)
    return
  const rtl = root.getAttribute('dir') === 'rtl'
  const forward = event.key === 'ArrowDown' || event.key === (rtl ? 'ArrowLeft' : 'ArrowRight')
  const next = inputs[(index + (forward ? 1 : -1) + inputs.length) % inputs.length]!
  event.preventDefault()
  next.focus()
  next.click()
}
</script>

<template>
  <ArkRadio.Root v-model="value" :name="name" :form="form" :disabled="isDisabled" :read-only="isReadOnly" :invalid="invalid" :orientation="appearance === 'segmented' ? 'horizontal' : orientation" class="flex gap-3" :class="[appearance === 'segmented' ? 'flex-row rounded-lg bg-recessed p-1' : orientation === 'horizontal' ? 'flex-row flex-wrap' : 'flex-col']" @keydown="onKeydown">
    <ArkRadio.Label v-if="label || legend || $slots.label" class="text-size-base font-medium text-strong" :class="appearance === 'segmented' && 'sr-only'">
      <slot name="label">
        {{ label || legend }}
      </slot>
    </ArkRadio.Label>
    <p v-if="description" class="text-size-sm text-subtle">
      {{ description }}
    </p>
    <slot />
  </ArkRadio.Root>
</template>
