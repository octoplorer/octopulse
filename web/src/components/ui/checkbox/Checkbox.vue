<script setup lang="ts">
import { Checkbox as ArkCheckbox } from '@ark-ui/vue/checkbox'
import { cva } from 'cva'
import { computed, inject, useAttrs } from 'vue'
import { checkboxAppearanceKey } from './context'

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
  variant?: 'default' | 'error'
  appearance?: 'default' | 'card'
  controlFirst?: boolean
}>(), { disabled: undefined, readOnly: undefined, required: undefined, invalid: undefined, controlFirst: undefined })
const attrs = useAttrs()
const accessibleInputAttrs = computed(() => Object.fromEntries(Object.entries(attrs).filter(([key]) => ['aria-label', 'aria-labelledby', 'aria-describedby'].includes(key))))
const checked = defineModel<boolean | 'indeterminate'>()
const group = inject(checkboxAppearanceKey, undefined)
const appearance = computed(() => props.appearance ?? group?.value.appearance ?? 'default')
const controlFirst = computed(() => props.controlFirst ?? group?.value.controlFirst ?? appearance.value !== 'card')
const checkboxClasses = cva({
  base: 'group inline-flex min-h-24px min-w-24px cursor-pointer items-start gap-2.5 text-size-base text-default data-[disabled]:cursor-not-allowed data-[disabled]:opacity-50',
  variants: {
    appearance: { default: '', card: 'w-full rounded-lg border border-hairline bg-base p-3 hover:bg-elevated data-[state=checked]:border-interact data-[state=checked]:bg-tint' },
    controlFirst: { true: '', false: 'flex-row-reverse justify-between' },
  },
  defaultVariants: { appearance: 'default', controlFirst: true },
})
</script>

<template>
  <ArkCheckbox.Root
    v-bind="$attrs" :checked="checked" :name="name" :value="value" :form="form" :disabled="disabled" :read-only="readOnly"
    :required="required" :invalid="invalid || variant === 'error' || undefined"
    :class="checkboxClasses({ appearance, controlFirst })"
    @update:checked="checked = $event"
  >
    <ArkCheckbox.Control class="mt-0.5 flex size-4 shrink-0 items-center justify-center rounded-sm bg-base text-inverse ring-1 ring-control-border outline-none transition-colors data-[state=checked]:bg-contrast data-[state=checked]:ring-contrast data-[state=indeterminate]:bg-contrast data-[state=indeterminate]:ring-contrast data-[invalid]:ring-action-danger data-[focus]:ring-focus! data-[focus-visible]:ring-2 data-[focus]:data-[focus-visible]:ring-brand!" :class="appearance === 'default' && 'data-[focus]:ring-2'">
      <ArkCheckbox.Indicator><span class="i-lucide-check size-3" aria-hidden="true" /></ArkCheckbox.Indicator>
      <ArkCheckbox.Indicator indeterminate>
        <span class="i-lucide-minus size-3" aria-hidden="true" />
      </ArkCheckbox.Indicator>
    </ArkCheckbox.Control>
    <ArkCheckbox.Label v-if="label || $slots.default" class="min-w-0 [overflow-wrap:anywhere]">
      <span class="block font-medium"><slot>{{ label }}</slot></span>
      <span v-if="description || $slots.description" class="mt-1 block text-size-sm text-subtle"><slot name="description">{{ description }}</slot></span>
    </ArkCheckbox.Label>
    <ArkCheckbox.HiddenInput v-bind="accessibleInputAttrs" :indeterminate.prop="checked === 'indeterminate'" />
  </ArkCheckbox.Root>
</template>
