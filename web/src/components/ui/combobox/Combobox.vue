<script setup lang="ts">
import type { ComboboxItem, ComboboxOption } from './types'
import { Combobox as ArkCombobox, createListCollection } from '@ark-ui/vue/combobox'
import { useDialogContext } from '@ark-ui/vue/dialog'
import { useFieldContext } from '@ark-ui/vue/field'
import { cva } from 'cva'
import { computed, shallowRef, useId, watch } from 'vue'
import { Input } from '../input'

defineOptions({ inheritAttrs: false })
const props = withDefaults(defineProps<{
  items: ComboboxItem[]
  label?: string
  description?: string
  error?: string
  placeholder?: string
  name?: string
  disabled?: boolean
  readOnly?: boolean
  required?: boolean
  invalid?: boolean
  multiple?: boolean
  clearable?: boolean
  freeText?: boolean
  size?: 'xs' | 'sm' | 'base' | 'lg'
  variant?: 'default' | 'error'
  emptyLabel?: string
  clearLabel?: string
  removeLabel?: (item: ComboboxOption) => string
  filter?: (item: ComboboxOption, query: string) => boolean
}>(), { size: 'base', variant: 'default', emptyLabel: 'No results found', clearLabel: 'Clear selection', removeLabel: (item: ComboboxOption) => `Remove ${item.label}`, disabled: undefined, readOnly: undefined, required: undefined, invalid: undefined })
const emit = defineEmits<{ queryChange: [value: string], select: [items: ComboboxOption[]] }>()
const value = defineModel<string | string[]>()
const open = defineModel<boolean>('open', { default: false })
const inputModel = defineModel<string>('inputValue')
const field = useFieldContext()
const nestedDialog = !!useDialogContext()
const id = useId()
const descriptionId = `${id}-description`
const errorId = `${id}-error`
const positioning = { sameWidth: true, placement: 'bottom-start' as const, strategy: 'fixed' as const }
const isDisabled = computed(() => props.disabled ?? field?.value.disabled)
const isReadOnly = computed(() => props.readOnly ?? field?.value.readOnly)
const options = computed<ComboboxOption[]>(() => props.items.map(item => typeof item === 'string' ? { value: item, label: item } : item))
const selected = computed({
  get: () => Array.isArray(value.value) ? value.value : value.value ? [value.value] : [],
  set: (next: string[]) => value.value = props.multiple ? next : next[0] ?? '',
})
const selectedItems = computed(() => options.value.filter(item => selected.value.includes(item.value)))
const initialInput = props.multiple ? '' : selectedItems.value[0]?.label ?? ''
const internalInput = shallowRef(initialInput)
const filterText = shallowRef('')
const inputValue = computed({ get: () => inputModel.value ?? internalInput.value, set: (next: string) => {
  internalInput.value = next
  inputModel.value = next
} })
const filtered = computed(() => options.value.filter(item => props.filter ? props.filter(item, filterText.value) : item.label.toLocaleLowerCase().includes(filterText.value.toLocaleLowerCase())))
const collection = computed(() => createListCollection({ items: filtered.value, itemToString: item => item.label, itemToValue: item => item.value, isItemDisabled: item => !!item.disabled }))
const groups = computed(() => [...new Set(filtered.value.map(item => item.group ?? ''))].map(label => ({ label, items: filtered.value.filter(item => (item.group ?? '') === label) })))

function updateInput(next: string) {
  inputValue.value = next
  filterText.value = next
  emit('queryChange', next)
}
function updateSelection(next: string[]) {
  selected.value = next
  filterText.value = ''
  emit('select', options.value.filter(item => next.includes(item.value)))
}
function remove(item: ComboboxOption) {
  if (isDisabled.value || isReadOnly.value)
    return
  selected.value = selected.value.filter(current => current !== item.value)
}
watch(() => [props.items, value.value] as const, () => {
  if (!props.multiple && !open.value)
    inputValue.value = selectedItems.value[0]?.label ?? ''
})
watch(open, (next) => {
  if (!next)
    filterText.value = ''
})
const controlClasses = cva({
  base: 'ui-control-group flex min-w-0 flex-wrap items-center gap-1 rounded-lg border-0 bg-control ring-1 ring-control-border outline-none px-2 focus-within:ring-2 focus-within:ring-focus data-[invalid]:ring-action-danger data-[invalid]:focus-within:ring-action-danger data-[disabled]:opacity-50',
  variants: { size: { xs: 'min-h-5 max-sm:min-h-7', sm: 'min-h-6.5 max-sm:min-h-8', base: 'min-h-9', lg: 'min-h-10' } },
  defaultVariants: { size: 'base' },
})
</script>

<template>
  <ArkCombobox.Root
    v-model:open="open" :model-value="selected" :input-value="inputValue" :collection="collection" :multiple="multiple" :name="freeText ? name : undefined"
    :disabled="isDisabled" :read-only="isReadOnly" :required="required" :invalid="invalid || !!error || variant === 'error' || undefined"
    :allow-custom-value="freeText" :selection-behavior="multiple ? 'clear' : 'replace'" :positioning="positioning"
    class="flex min-w-0 flex-col gap-1.5" @update:model-value="updateSelection" @update:input-value="updateInput"
  >
    <ArkCombobox.Label v-if="label || $slots.label" class="text-size-base font-medium text-strong">
      <slot name="label">
        {{ label }}
      </slot>
    </ArkCombobox.Label>
    <ArkCombobox.Control :class="controlClasses({ size })">
      <slot name="prefix" />
      <span v-for="item in multiple ? selectedItems : []" :key="item.value" class="my-1 inline-flex max-w-full items-center gap-1 rounded-sm bg-overlay px-1.5 py-0.5 text-size-xs text-default ring-1 ring-hairline">
        <slot name="chip" :item="item"><span class="min-w-0 whitespace-normal [overflow-wrap:anywhere]">{{ item.label }}</span></slot>
        <button type="button" :disabled="isDisabled || isReadOnly" :aria-label="removeLabel(item)" class="inline-flex size-24px shrink-0 items-center justify-center rounded-sm text-subtle hover:bg-fill-hover outline-none! focus-visible:ring-2 focus-visible:ring-focus focus-visible:ring-inset disabled:cursor-not-allowed" @click="remove(item)"><span class="i-lucide-x size-3" aria-hidden="true" /></button>
      </span>
      <ArkCombobox.Input as-child>
        <Input v-bind="$attrs" :size="size" :placeholder="placeholder" :aria-describedby="[field?.ariaDescribedby, $attrs['aria-describedby'], description && descriptionId, error && errorId].filter(Boolean).join(' ') || undefined" class="min-w-0! basis-16 flex-1 border-0! bg-transparent! px-1! outline-none! ring-0!" />
      </ArkCombobox.Input>
      <ArkCombobox.ClearTrigger v-if="clearable && (selected.length || inputValue)" :aria-label="clearLabel" class="flex size-24px shrink-0 items-center justify-center rounded-md bg-transparent text-subtle hover:bg-tint outline-none! focus-visible:ring-2 focus-visible:ring-focus focus-visible:ring-inset">
        <span class="i-lucide-x size-3.5" aria-hidden="true" />
      </ArkCombobox.ClearTrigger>
      <ArkCombobox.Trigger class="flex size-24px shrink-0 items-center justify-center rounded-md bg-transparent text-subtle hover:bg-tint outline-none! focus-visible:ring-2 focus-visible:ring-focus focus-visible:ring-inset" :aria-label="label ? `Show ${label}` : 'Show options'">
        <span class="i-lucide-chevron-down size-4" aria-hidden="true" />
      </ArkCombobox.Trigger>
      <slot name="suffix" />
    </ArkCombobox.Control>
    <p v-if="description" :id="descriptionId" class="text-size-sm text-subtle">
      {{ description }}
    </p>
    <p v-if="error" :id="errorId" class="text-size-sm text-fg-danger" role="alert">
      {{ error }}
    </p>
    <template v-if="name && !freeText">
      <input v-for="selectedValue in selected" :key="selectedValue" type="hidden" :name="name" :value="selectedValue" :disabled="isDisabled">
    </template>
    <Teleport to="body" :disabled="nestedDialog">
      <ArkCombobox.Positioner class="z-220">
        <ArkCombobox.Content class="z-220 max-h-[min(18rem,var(--available-height))] min-w-0 max-w-[var(--available-width)] overflow-auto overscroll-contain rounded-lg border border-line bg-elevated p-1 text-size-base text-default shadow-panel outline-none">
          <ArkCombobox.ItemGroup v-for="group in groups" :key="group.label">
            <ArkCombobox.ItemGroupLabel v-if="group.label" class="px-2 py-1.5 text-size-xs font-medium text-subtle">
              {{ group.label }}
            </ArkCombobox.ItemGroupLabel>
            <ArkCombobox.Item v-for="item in group.items" :key="item.value" :item="item" class="flex cursor-pointer items-center justify-between gap-3 rounded-md px-2 py-1.5 data-[highlighted]:bg-tint data-[disabled]:pointer-events-none data-[disabled]:opacity-50">
              <ArkCombobox.ItemText class="min-w-0">
                <slot name="item" :item="item" :selected="selected.includes(item.value)">
                  <span class="block whitespace-normal [overflow-wrap:anywhere]">{{ item.label }}</span><span v-if="item.description" class="block text-size-xs text-subtle">{{ item.description }}</span>
                </slot>
              </ArkCombobox.ItemText>
              <ArkCombobox.ItemIndicator><span class="i-lucide-check size-4 text-default" aria-hidden="true" /></ArkCombobox.ItemIndicator>
            </ArkCombobox.Item>
          </ArkCombobox.ItemGroup>
          <ArkCombobox.Empty class="px-2 py-3 text-center text-size-sm text-subtle">
            <slot name="empty">
              {{ emptyLabel }}
            </slot>
          </ArkCombobox.Empty>
          <slot name="footer" />
        </ArkCombobox.Content>
      </ArkCombobox.Positioner>
    </Teleport>
  </ArkCombobox.Root>
</template>
