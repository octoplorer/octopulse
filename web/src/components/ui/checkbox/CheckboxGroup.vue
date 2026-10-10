<script setup lang="ts">
import { Checkbox as ArkCheckbox } from '@ark-ui/vue/checkbox'
import { computed, provide, useId } from 'vue'
import { checkboxAppearanceKey } from './context'

const props = withDefaults(defineProps<{
  label?: string
  legend?: string
  description?: string
  error?: string
  name?: string
  disabled?: boolean
  readOnly?: boolean
  invalid?: boolean
  maxSelectedValues?: number
  orientation?: 'vertical' | 'horizontal'
  appearance?: 'default' | 'card'
  controlFirst?: boolean
}>(), { orientation: 'vertical', appearance: 'default', controlFirst: undefined, disabled: undefined, readOnly: undefined, invalid: undefined })
const value = defineModel<string[]>({ default: () => [] })
const id = useId()
provide(checkboxAppearanceKey, computed(() => ({ appearance: props.appearance, controlFirst: props.controlFirst })))
</script>

<template>
  <fieldset class="min-w-0 border-0 p-0" :disabled="disabled">
    <legend v-if="label || legend || $slots.label" :id="`${id}-label`" class="mb-2 text-size-base font-medium text-strong">
      <slot name="label">
        {{ label || legend }}
      </slot>
    </legend>
    <p v-if="description" :id="`${id}-description`" class="mb-3 text-size-sm text-subtle">
      {{ description }}
    </p>
    <ArkCheckbox.Group
      v-model="value" :name="name" :disabled="disabled" :read-only="readOnly" :invalid="invalid || !!error || undefined" :max-selected-values="maxSelectedValues"
      :aria-labelledby="label || legend || $slots.label ? `${id}-label` : undefined"
      :aria-describedby="[description && `${id}-description`, error && `${id}-error`].filter(Boolean).join(' ') || undefined"
      class="flex gap-3" :class="orientation === 'horizontal' ? 'flex-row flex-wrap' : 'flex-col'"
    >
      <slot />
    </ArkCheckbox.Group>
    <p v-if="error" :id="`${id}-error`" class="mt-2 text-size-sm text-fg-danger" role="alert">
      {{ error }}
    </p>
  </fieldset>
</template>
