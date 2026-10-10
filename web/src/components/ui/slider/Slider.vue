<script setup lang="ts">
import { useFieldContext } from '@ark-ui/vue/field'
import { Slider as ArkSlider } from '@ark-ui/vue/slider'
import { cva } from 'cva'
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  label?: string
  name?: string
  min?: number
  max?: number
  step?: number
  disabled?: boolean
  readOnly?: boolean
  invalid?: boolean
  size?: 'sm' | 'base' | 'lg'
  orientation?: 'horizontal' | 'vertical'
  showValue?: boolean
  showBounds?: boolean
  format?: Intl.NumberFormatOptions
  locale?: string
  marks?: (number | { value: number, label: string })[]
  minStepsBetweenThumbs?: number
  thumbAlignment?: 'contain' | 'center'
}>(), { min: 0, max: 100, step: 1, size: 'base', orientation: 'horizontal', showValue: true, showBounds: true, thumbAlignment: 'contain', disabled: undefined, readOnly: undefined, invalid: undefined })
const emit = defineEmits<{ valueChangeEnd: [value: number | number[]] }>()
const field = useFieldContext()
const isDisabled = computed(() => props.disabled ?? field?.value.disabled)
const isReadOnly = computed(() => props.readOnly ?? field?.value.readOnly)
const value = defineModel<number | number[]>({ default: 0 })
const values = computed({ get: () => Array.isArray(value.value) ? value.value : [value.value], set: (next: number[]) => value.value = Array.isArray(value.value) ? next : next[0] ?? props.min })
const formatter = computed(() => new Intl.NumberFormat(props.locale, props.format))
const markers = computed(() => props.marks?.map(mark => typeof mark === 'number' ? { value: mark, label: formatter.value.format(mark) } : mark))
const frameClasses = cva({ base: 'bg-recessed p-3px ring-1 ring-control-border', variants: { orientation: { horizontal: 'w-full', vertical: 'h-full' }, size: { sm: 'rounded-md data-[orientation=horizontal]:h-6 data-[orientation=vertical]:w-6', base: 'rounded-lg data-[orientation=horizontal]:h-8 data-[orientation=vertical]:w-8', lg: 'rounded-lg data-[orientation=horizontal]:h-10 data-[orientation=vertical]:w-10' } }, defaultVariants: { size: 'base', orientation: 'horizontal' } })
const thumbClasses = cva({ base: 'group flex cursor-grab items-center justify-center bg-overlay shadow-control ring-1 ring-control-border outline-none focus-visible:ring-2 focus-visible:ring-focus data-[dragging]:cursor-grabbing data-[disabled]:cursor-not-allowed', variants: { size: { sm: 'rounded-sm', base: 'rounded-md', lg: 'rounded-md' }, orientation: { horizontal: 'h-full w-4', vertical: 'h-4 w-full' } }, defaultVariants: { size: 'base', orientation: 'horizontal' } })
</script>

<template>
  <ArkSlider.Root v-model="values" :name="name" :min="min" :max="max" :step="step" :disabled="isDisabled" :read-only="isReadOnly" :invalid="invalid" :orientation="orientation" :min-steps-between-thumbs="minStepsBetweenThumbs" :thumb-alignment="thumbAlignment" :aria-label="values.map((_, index) => values.length > 1 ? `${label || 'Value'} ${index + 1}` : label || 'Value')" class="flex min-w-0 flex-col gap-2 data-[disabled]:opacity-50" :class="orientation === 'vertical' && 'h-48 items-center'" @value-change-end="emit('valueChangeEnd', Array.isArray(value) ? $event.value : $event.value[0] ?? min)">
    <div v-if="label || $slots.label" class="flex items-center justify-between gap-3">
      <ArkSlider.Label v-if="label || $slots.label" class="text-size-base font-medium text-strong">
        <slot name="label">
          {{ label }}
        </slot>
      </ArkSlider.Label>
    </div>
    <div :data-orientation="orientation" :class="frameClasses({ size, orientation })">
      <ArkSlider.Control class="relative flex h-full w-full items-center justify-center">
        <ArkSlider.Track class="relative h-full w-full">
          <ArkSlider.Range class="absolute h-full rounded-md bg-overlay shadow-control ring-1 ring-line data-[orientation=vertical]:w-full" />
        </ArkSlider.Track>
        <ArkSlider.Thumb v-for="(current, index) in values" :key="index" :index="index" :class="thumbClasses({ size, orientation })">
          <span class="rounded-full bg-contrast/25" :class="orientation === 'vertical' ? 'h-0.5 w-1/2' : 'h-1/2 w-0.5'" aria-hidden="true" />
          <span v-if="showValue" class="absolute rounded-sm bg-contrast px-1.5 text-size-xs font-medium text-inverse tabular-nums whitespace-nowrap group-data-[disabled]:bg-fill group-data-[disabled]:text-default" :class="orientation === 'vertical' ? 'left-full ml-2' : 'top-full left-1/2 mt-11px -translate-x-1/2'" aria-hidden="true"><slot name="value" :value="current" :index="index">{{ formatter.format(current) }}</slot></span>
          <ArkSlider.HiddenInput />
        </ArkSlider.Thumb>
      </ArkSlider.Control>
    </div>
    <ArkSlider.MarkerGroup v-if="markers?.length" class="relative text-size-xs text-subtle" :class="orientation === 'vertical' ? 'h-full' : 'h-4 w-full'">
      <ArkSlider.Marker v-for="mark in markers" :key="mark.value" :value="mark.value">
        <slot name="mark" :mark="mark">
          {{ mark.label }}
        </slot>
      </ArkSlider.Marker>
    </ArkSlider.MarkerGroup>
    <div v-else-if="showBounds && orientation === 'horizontal'" class="flex justify-between text-size-xs text-subtle tabular-nums">
      <span>{{ formatter.format(min) }}</span><span>{{ formatter.format(max) }}</span>
    </div>
  </ArkSlider.Root>
</template>
