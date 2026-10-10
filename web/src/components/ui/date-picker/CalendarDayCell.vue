<script setup lang="ts">
import type { DateValue } from '@ark-ui/vue/date-picker'
import { DatePicker as ArkDatePicker, useDatePickerContext } from '@ark-ui/vue/date-picker'
import { cva } from 'cva'
import { computed, inject } from 'vue'
import { calendarAppearanceKey } from './context'

const props = withDefaults(defineProps<{
  value: DateValue
  visibleRange: { start: DateValue, end: DateValue }
  showOutsideDays?: boolean
  size?: 'sm' | 'base' | 'lg'
}>(), { showOutsideDays: true, size: 'base' })
const calendar = useDatePickerContext()
const appearance = inject(calendarAppearanceKey, 'calendar')
const state = computed(() => calendar.value.getDayTableCellState({ value: props.value, visibleRange: props.visibleRange }))
// Zag derives the end from numOfMonths; each rendered panel needs its own bounds.
const outside = computed(() => props.value.compare(props.visibleRange.start) < 0 || props.value.compare(props.visibleRange.end) > 0)
const selection = computed(() => {
  const { firstInRange, lastInRange, inRange, selected } = state.value
  if (appearance === 'date-range' && outside.value)
    return selected || inRange ? 'outside-range' : 'none'
  if (firstInRange && lastInRange)
    return 'range-both'
  if (firstInRange)
    return 'range-start'
  if (lastInRange)
    return 'range-end'
  if (inRange)
    return 'range-middle'
  return selected ? 'selected' : 'none'
})
const unavailable = computed(() => calendar.value.disabled || !state.value.selectable)
const foreground = computed(() => {
  if (selection.value.startsWith('range-') && selection.value !== 'range-middle')
    return 'inverse'
  if (selection.value === 'range-middle' && appearance === 'calendar')
    return 'range-middle'
  if (outside.value)
    return 'subtle'
  if (selection.value === 'selected')
    return 'inverse'
  if (state.value.today && selection.value === 'none')
    return 'today'
  return 'default'
})
const hover = computed(() => {
  if (unavailable.value || (outside.value && appearance === 'date-range'))
    return 'transparent'
  if (selection.value === 'selected')
    return 'selected'
  if (selection.value.startsWith('range-') && selection.value !== 'range-middle')
    return 'transparent'
  return selection.value === 'range-middle' ? 'range-middle' : 'default'
})
const cellClasses = cva({
  base: 'text-center',
  variants: {
    appearance: { 'calendar': 'p-2px', 'date-range': 'p-0' },
    selection: {
      'none': '',
      'selected': '',
      'range-start': 'rounded-l-md bg-[var(--calendar-accent)]',
      'range-end': 'rounded-r-md bg-[var(--calendar-accent)]',
      'range-both': 'rounded-md bg-[var(--calendar-accent)]',
      'range-middle': 'bg-[var(--calendar-range-middle)]',
      'outside-range': 'bg-fill',
    },
    hidden: { true: 'invisible', false: '' },
  },
})
const dayClasses = cva({
  base: 'flex cursor-pointer items-center justify-center rounded-md p-0 text-size-sm outline-none transition-colors focus-visible:ring-2 focus-visible:ring-brand focus-visible:ring-inset',
  variants: {
    selected: { true: 'bg-[var(--calendar-accent)]', false: 'bg-transparent' },
    foreground: {
      'default': 'text-[var(--calendar-day)]',
      'inverse': 'text-[var(--calendar-accent-text)]',
      'range-middle': 'text-[var(--calendar-range-middle-text)]',
      'subtle': 'text-[var(--calendar-subtle)]',
      'today': 'font-semibold text-[var(--calendar-today)]',
    },
    hover: {
      'default': 'hover:bg-[var(--calendar-hover)]',
      'selected': 'hover:bg-[var(--calendar-accent)]',
      'range-middle': 'hover:bg-[var(--calendar-fill-hover)]',
      'transparent': 'hover:bg-transparent',
    },
    dimmed: { true: 'opacity-40', false: '' },
    unavailable: { true: 'cursor-not-allowed', false: '' },
    appearance: { 'calendar': '', 'date-range': '' },
    size: { sm: '', base: '', lg: '' },
  },
  compoundVariants: [
    { appearance: 'calendar', size: 'sm', class: 'size-7' },
    { appearance: 'calendar', size: 'base', class: 'size-8' },
    { appearance: 'calendar', size: 'lg', class: 'size-10' },
    { appearance: 'date-range', size: 'sm', class: 'h-24px w-28px' },
    { appearance: 'date-range', size: 'base', class: 'h-28px w-32px' },
    { appearance: 'date-range', size: 'lg', class: 'h-32px w-36px' },
  ],
})
</script>

<template>
  <ArkDatePicker.TableCell :value="value" :visible-range="visibleRange" :class="cellClasses({ appearance, selection, hidden: outside && !showOutsideDays })">
    <ArkDatePicker.TableCellTrigger
      v-if="showOutsideDays || !outside"
      :data-outside-range="outside ? '' : undefined"
      :class="dayClasses({ appearance, size, selected: selection === 'selected', foreground, hover, dimmed: appearance === 'calendar' && (outside || (unavailable && selection === 'none')), unavailable })"
    >
      <slot :date="value" :state="{ ...state, outsideRange: outside }">
        {{ value.day }}
      </slot>
    </ArkDatePicker.TableCellTrigger>
  </ArkDatePicker.TableCell>
</template>
