<script setup lang="ts">
import { useFieldContext } from '@ark-ui/vue/field'
import { Switch as ArkSwitch } from '@ark-ui/vue/switch'
import { cva } from 'cva'
import { computed, useAttrs, useId, useSlots } from 'vue'
import { useSwitchGroup } from './context'

defineOptions({ inheritAttrs: false })
const props = withDefaults(defineProps<{
  label?: string
  description?: string
  name?: string
  value?: string
  form?: string
  disabled?: boolean
  readOnly?: boolean
  required?: boolean
  invalid?: boolean
  checked?: boolean
  controlFirst?: boolean
  size?: 'sm' | 'base' | 'lg'
  variant?: 'default' | 'neutral'
  transitioning?: boolean
}>(), { value: 'on', checked: undefined, disabled: undefined, readOnly: undefined, required: undefined, invalid: undefined, controlFirst: undefined, size: 'base', variant: 'default' })
const emit = defineEmits<{ 'checkedChange': [checked: boolean], 'update:checked': [checked: boolean] }>()
const value = defineModel<boolean>({ default: false })
const group = useSwitchGroup()
const field = useFieldContext()
const attrs = useAttrs()
const slots = useSlots()
const descriptionId = useId()
const disabled = computed(() => group?.disabled.value || (props.disabled ?? field?.value.disabled))
const readOnly = computed(() => props.readOnly ?? field?.value.readOnly)
const checked = computed(() => props.checked ?? value.value)
const ringColor = computed(() => checked.value ? props.variant === 'neutral' ? 'ring-contrast' : 'ring-brand' : 'ring-control-border')
const controlFirst = computed(() => props.controlFirst ?? group?.controlFirst.value ?? true)
const inputAttrs = computed(() => ({
  ...Object.fromEntries(Object.entries(attrs).filter(([key]) => ['aria-label', 'aria-labelledby'].includes(key))),
  'aria-describedby': [attrs['aria-describedby'], props.description ? descriptionId : undefined].filter(Boolean).join(' ') || undefined,
  'aria-readonly': readOnly.value || undefined,
  'role': 'switch',
  'aria-checked': props.checked ?? value.value,
  ...(!props.label && !slots.default ? { 'aria-label': typeof attrs['aria-label'] === 'string' ? attrs['aria-label'] : 'Switch', 'aria-labelledby': undefined } : {}),
}))
const controlClasses = cva({
  base: 'relative inline-flex shrink-0 items-center rounded-[5px] bg-interact p-0 ring-1 transition-colors motion-reduce:transition-none [corner-shape:squircle] supports-[corner-shape:squircle]:rounded-[10px]',
  variants: { size: { sm: 'h-4 w-8', base: 'h-4.5 w-9', lg: 'h-5 w-10' }, variant: { default: 'data-[state=checked]:bg-brand', neutral: 'data-[state=checked]:bg-contrast' } },
  defaultVariants: { size: 'base', variant: 'default' },
})
const thumbClasses = cva({
  base: 'absolute inset-y-0 left-0 rounded-[5px] bg-control-thumb shadow-control transition-transform duration-150 motion-reduce:transition-none [corner-shape:squircle] supports-[corner-shape:squircle]:rounded-[10px]',
  variants: { size: { sm: 'w-4 data-[state=checked]:translate-x-4 rtl:data-[state=checked]:-translate-x-4 rtl:left-auto rtl:right-0', base: 'w-4.5 data-[state=checked]:translate-x-4.5 rtl:data-[state=checked]:-translate-x-4.5 rtl:left-auto rtl:right-0', lg: 'w-5 data-[state=checked]:translate-x-5 rtl:data-[state=checked]:-translate-x-5 rtl:left-auto rtl:right-0' } },
  defaultVariants: { size: 'base' },
})
function setChecked(checked: boolean) {
  value.value = checked
  emit('checkedChange', checked)
  emit('update:checked', checked)
}
</script>

<template>
  <ArkSwitch.Root v-bind="$attrs" :checked="props.checked ?? value" :disabled="disabled" :read-only="readOnly" :required="props.required" :invalid="props.invalid" :name="props.name" :value="props.value" :form="props.form" :aria-busy="props.transitioning || undefined" class="group inline-flex min-h-24px min-w-0 max-w-full cursor-pointer items-center gap-2.5 text-size-base text-default data-[disabled]:cursor-not-allowed data-[disabled]:opacity-50" :class="controlFirst ? '' : 'flex-row-reverse justify-between'" @update:checked="setChecked">
    <ArkSwitch.Control :class="[controlClasses({ size: props.size, variant: props.variant }), ringColor, !disabled && !readOnly ? 'data-[focus-visible]:ring-2 data-[focus-visible]:!ring-brand' : undefined, group && !disabled && !readOnly ? 'data-[focus]:ring-focus' : undefined]">
      <ArkSwitch.Thumb :class="thumbClasses({ size: props.size })" />
    </ArkSwitch.Control>
    <span v-if="props.label || $slots.default || props.description" class="min-w-0 [overflow-wrap:anywhere]">
      <ArkSwitch.Label v-if="props.label || $slots.default" class="block font-medium"><slot>{{ props.label }}</slot></ArkSwitch.Label>
      <span v-if="props.description" :id="descriptionId" class="mt-1 block text-size-sm text-subtle">{{ props.description }}</span>
    </span>
    <ArkSwitch.HiddenInput v-bind="inputAttrs" class="!outline-none" />
  </ArkSwitch.Root>
</template>
