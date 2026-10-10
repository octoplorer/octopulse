<script setup lang="ts">
import type { UseTagsInputProps } from '@ark-ui/vue/tags-input'
import { useFieldContext } from '@ark-ui/vue/field'
import { TagsInput as ArkTagsInput, useTagsInput } from '@ark-ui/vue/tags-input'
import { cva } from 'cva'
import { computed, shallowRef, useId } from 'vue'
import { Button } from '../button'

export interface TagInputLabels {
  input?: string
  removeValue?: (value: string) => string
  editValue?: (value: string) => string
  invalidValue?: (value: string) => string
  maxValuesReached?: (max: number) => string
}

defineOptions({ inheritAttrs: false })
const props = withDefaults(defineProps<{
  label?: string
  description?: string
  error?: string
  placeholder?: string
  name?: string
  form?: string
  disabled?: boolean
  readOnly?: boolean
  required?: boolean
  allowDuplicates?: boolean
  editable?: boolean
  addOnBlur?: boolean
  maxValues?: number
  validateValue?: (value: string, values: string[]) => boolean
  delimiter?: string | RegExp
  labels?: TagInputLabels
  variant?: 'default' | 'error'
  size?: 'sm' | 'base' | 'lg'
}>(), { editable: true, addOnBlur: true, variant: 'default', size: 'base', disabled: undefined, readOnly: undefined, required: undefined })
const values = defineModel<string[]>({ default: () => [] })
const inputValue = defineModel<string>('inputValue', { default: '' })
const message = shallowRef<string>()
const field = useFieldContext()
const id = useId()
const errorId = `${id}-error`
const descriptionId = `${id}-description`
const invalid = computed(() => !!props.error || !!message.value || props.variant === 'error')
const tags = useTagsInput(computed<UseTagsInputProps>(() => ({
  modelValue: values.value,
  inputValue: inputValue.value,
  name: props.name,
  form: props.form,
  disabled: props.disabled,
  readOnly: props.readOnly,
  required: props.required,
  invalid: invalid.value,
  max: props.maxValues,
  editable: props.editable,
  allowDuplicates: props.allowDuplicates,
  delimiter: props.delimiter ?? /[,\n]/,
  ids: { input: field?.value.ids.control, label: field?.value.ids.label },
  addOnPaste: false,
  sanitizeValue: raw => raw.trim(),
  validate: ({ inputValue, value }) => !props.validateValue || props.validateValue(inputValue.trim(), value),
  translations: {
    inputLabel: () => props.labels?.input ?? props.label ?? 'Tags',
    deleteTagTriggerLabel: tag => props.labels?.removeValue?.(tag) ?? `Remove ${tag}`,
  },
  onInputValueChange: details => inputValue.value = details.inputValue,
  onValueChange: details => values.value = details.value,
  onValueInvalid: details => message.value = details.reason === 'rangeOverflow'
    ? props.labels?.maxValuesReached?.(props.maxValues ?? 0) ?? `Limit of ${props.maxValues} tags reached.`
    : props.labels?.invalidValue?.(inputValue.value) ?? `"${inputValue.value}" is not valid.`,
})))
const controlClasses = cva({ base: 'ui-control-group flex min-w-0 flex-wrap items-center gap-1.5 rounded-lg border-0 bg-control ring-1 ring-control-border outline-none px-2 focus-within:ring-2 focus-within:ring-focus data-[invalid]:ring-action-danger data-[invalid]:focus-within:ring-action-danger data-[disabled]:opacity-50', variants: { size: { sm: 'min-h-7 py-0.5 text-size-xs', base: 'min-h-9 py-1 text-size-base', lg: 'min-h-10 py-1.5 text-size-base' } }, defaultVariants: { size: 'base' } })

function commit(raw = inputValue.value) {
  if (props.disabled || props.readOnly || field?.value.disabled || field?.value.readOnly)
    return false
  const entries = raw.split(props.delimiter ?? /[,\n]/).map(entry => entry.trim()).filter(Boolean)
  const next = [...values.value]
  let error: string | undefined
  let remainder = ''
  for (const [index, entry] of entries.entries()) {
    if (!props.allowDuplicates && next.includes(entry))
      continue
    if (props.maxValues !== undefined && next.length >= props.maxValues) {
      error = props.labels?.maxValuesReached?.(props.maxValues) ?? `Limit of ${props.maxValues} tags reached.`
    }
    else if (props.validateValue && !props.validateValue(entry, next)) {
      error = props.labels?.invalidValue?.(entry) ?? `"${entry}" is not valid.`
    }
    else {
      next.push(entry)
      continue
    }
    remainder = entries.slice(index).join(', ')
    break
  }
  tags.value.setValue(next)
  tags.value.setInputValue(remainder)
  message.value = error
  return !error
}
function onKeydown(event: KeyboardEvent) {
  if (event.isComposing || !inputValue.value || !['Enter', ',', 'Tab'].includes(event.key))
    return
  const added = commit()
  if (event.key !== 'Tab' || added)
    event.preventDefault()
}
function onPaste(event: ClipboardEvent) {
  const text = event.clipboardData?.getData('text')
  if (!text || !/[,\n]/.test(text))
    return
  event.preventDefault()
  commit(text)
}
function onBlur() {
  if (props.addOnBlur && inputValue.value.trim())
    commit()
}
defineExpose({ commit, focus: () => tags.value.focus() })
</script>

<template>
  <ArkTagsInput.RootProvider :value="tags" class="flex min-w-0 flex-col gap-1.5">
    <ArkTagsInput.Label v-if="label || $slots.label" class="text-size-base font-medium text-strong">
      <slot name="label">
        {{ label }}
      </slot>
    </ArkTagsInput.Label>
    <ArkTagsInput.Control :class="controlClasses({ size })">
      <ArkTagsInput.Item v-for="(tag, index) in values" :key="`${index}-${tag}`" :index="index" :value="tag" class="inline-flex max-w-full items-center rounded-sm bg-overlay px-1.5 py-0.5 text-size-xs text-default ring-1 ring-hairline data-[highlighted]:bg-tint">
        <ArkTagsInput.ItemPreview class="flex min-w-0 items-center gap-1.5">
          <ArkTagsInput.ItemText class="min-w-0 whitespace-normal [overflow-wrap:anywhere]">
            <slot name="tag" :tag="tag" :index="index">
              {{ tag }}
            </slot>
          </ArkTagsInput.ItemText><ArkTagsInput.ItemDeleteTrigger as-child :disabled="(disabled ?? field?.disabled) || (readOnly ?? field?.readOnly)">
            <Button variant="ghost" shape="square" size="xs" :disabled="(disabled ?? field?.disabled) || (readOnly ?? field?.readOnly)" @mousedown.prevent>
              <span class="i-lucide-x size-10px" aria-hidden="true" />
            </Button>
          </ArkTagsInput.ItemDeleteTrigger>
        </ArkTagsInput.ItemPreview>
        <ArkTagsInput.ItemInput class="min-w-0 max-w-full border-0 bg-transparent p-0 outline-none ring-0 max-sm:text-16px" :aria-label="labels?.editValue?.(tag) ?? `Edit ${tag}`" />
      </ArkTagsInput.Item>
      <ArkTagsInput.Input v-bind="$attrs" :placeholder="placeholder" :aria-describedby="[field?.ariaDescribedby, description && descriptionId, (error || message) && errorId].filter(Boolean).join(' ') || undefined" class="min-w-0 basis-32 flex-1 max-sm:text-16px border-0 bg-transparent px-1 py-0.5 text-default outline-none ring-0 placeholder:text-placeholder disabled:cursor-not-allowed" @keydown.capture="onKeydown" @paste.capture="onPaste" @blur="onBlur" />
    </ArkTagsInput.Control>
    <ArkTagsInput.HiddenInput />
    <p v-if="description" :id="descriptionId" class="text-size-sm text-subtle">
      {{ description }}
    </p>
    <p v-if="error || message" :id="errorId" class="text-size-sm text-fg-danger" role="alert">
      {{ error || message }}
    </p>
  </ArkTagsInput.RootProvider>
</template>
