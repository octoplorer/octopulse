<script setup lang="ts">
import type { DateValue } from '@ark-ui/vue/date-picker'
import type { DatePickerPreset } from './types'
import { DatePicker as ArkDatePicker, parseDate } from '@ark-ui/vue/date-picker'
import { useDialogContext } from '@ark-ui/vue/dialog'
import { useFieldContext } from '@ark-ui/vue/field'
import { computed, shallowRef } from 'vue'
import { Input } from '../input'
import CalendarPanel from './CalendarPanel.vue'

defineOptions({ inheritAttrs: false })
const props = withDefaults(defineProps<{
  label?: string
  name?: string
  disabled?: boolean
  readOnly?: boolean
  required?: boolean
  invalid?: boolean
  mode?: 'single' | 'multiple' | 'range'
  inline?: boolean
  numberOfMonths?: number
  showOutsideDays?: boolean
  showWeekNumbers?: boolean
  startOfWeek?: number
  min?: string
  max?: string
  maxSelectedDates?: number
  isDateUnavailable?: (date: DateValue, locale: string) => boolean
  locale?: string
  timeZone?: string
  timezone?: string
  placeholder?: string
  clearable?: boolean
  clearLabel?: string
  presets?: DatePickerPreset[]
  size?: 'sm' | 'base' | 'lg'
}>(), { mode: 'single', inline: true, numberOfMonths: 1, locale: 'en-US', size: 'base', disabled: undefined, readOnly: undefined, required: undefined, invalid: undefined, showOutsideDays: undefined })
const value = defineModel<string[]>({ default: () => [] })
const open = defineModel<boolean>('open', { default: false })
const field = useFieldContext()
const nestedDialog = !!useDialogContext()
const positioning = { placement: 'bottom-start' as const, strategy: 'fixed' as const }
const isDisabled = computed(() => props.disabled ?? field?.value.disabled)
const isReadOnly = computed(() => props.readOnly ?? field?.value.readOnly)
const parsed = computed(() => value.value.map(value => parseDate(value)))
const focused = shallowRef<DateValue | undefined>(parsed.value[0])
const minimum = computed(() => props.min ? parseDate(props.min) : undefined)
const maximum = computed(() => props.max ? parseDate(props.max) : undefined)
const panelProps = computed(() => ({ numberOfMonths: props.numberOfMonths, showOutsideDays: props.showOutsideDays ?? props.numberOfMonths === 1, showWeekNumbers: props.showWeekNumbers, clearable: props.clearable, clearLabel: props.clearLabel, presets: props.presets, timezone: props.timezone, size: props.size }))
function update(next: DateValue[]) {
  value.value = next.map(date => date.toString())
}
</script>

<template>
  <ArkDatePicker.Root
    v-bind="$attrs" v-model:open="open" v-model:focused-value="focused" :model-value="parsed" :selection-mode="mode" :inline="inline"
    :name="name" :disabled="isDisabled" :read-only="isReadOnly" :required="required" :invalid="invalid" :min="minimum" :max="maximum"
    :max-selected-dates="maxSelectedDates" :is-date-unavailable="isDateUnavailable" :num-of-months="numberOfMonths" :start-of-week="startOfWeek"
    :show-week-numbers="showWeekNumbers" :locale="locale" :time-zone="timeZone" :outside-day-selectable="true" :close-on-select="mode !== 'multiple'" :positioning="positioning"
    class="flex min-w-0 flex-col gap-1.5" @update:model-value="update"
  >
    <ArkDatePicker.Label v-if="label || $slots.label" class="text-size-base font-medium text-strong">
      <slot name="label">
        {{ label }}
      </slot>
    </ArkDatePicker.Label>
    <ArkDatePicker.Control v-if="!inline" class="ui-control-group flex min-w-0 items-center gap-2 rounded-lg border-0 bg-control ring-1 ring-control-border outline-none px-2 focus-within:ring-2 focus-within:ring-focus data-[invalid]:ring-action-danger data-[invalid]:focus-within:ring-action-danger data-[disabled]:opacity-50">
      <ArkDatePicker.Input as-child :index="0">
        <Input :name="mode === 'range' && name ? `${name}-start` : name" :size="size" :placeholder="placeholder" :aria-label="mode === 'range' ? `${label || 'Date range'} start` : label" class="min-w-0 flex-1 border-0! bg-transparent! px-0! shadow-none! outline-none! focus-visible:outline-none! ring-0!" />
      </ArkDatePicker.Input>
      <template v-if="mode === 'range'">
        <span class="text-subtle" aria-hidden="true">–</span><ArkDatePicker.Input as-child :index="1">
          <Input :name="name ? `${name}-end` : undefined" :size="size" :placeholder="placeholder" :aria-label="`${label || 'Date range'} end`" class="min-w-0 flex-1 border-0! bg-transparent! px-0! shadow-none! outline-none! focus-visible:outline-none! ring-0!" />
        </ArkDatePicker.Input>
      </template>
      <ArkDatePicker.Trigger class="flex size-9 shrink-0 items-center justify-center rounded-md bg-transparent text-default outline-none! hover:bg-tint focus-visible:ring-2 focus-visible:ring-focus focus-visible:ring-inset disabled:opacity-50">
        <span class="i-lucide-calendar-days size-4" aria-hidden="true" />
      </ArkDatePicker.Trigger>
    </ArkDatePicker.Control>
    <template v-if="inline && name">
      <input v-for="(date, index) in value" :key="index" type="hidden" :name="name" :value="date" :disabled="isDisabled">
    </template>
    <ArkDatePicker.Content v-if="inline" class="outline-none">
      <CalendarPanel v-bind="panelProps">
        <template v-for="(_, slotName) in $slots" #[slotName]="slotProps">
          <slot :name="slotName" v-bind="slotProps" />
        </template>
      </CalendarPanel>
    </ArkDatePicker.Content>
    <Teleport v-else to="body" :disabled="nestedDialog">
      <ArkDatePicker.Positioner class="z-220">
        <ArkDatePicker.Content class="z-220 rounded-xl border border-line bg-elevated shadow-panel outline-none">
          <CalendarPanel v-bind="panelProps">
            <template v-for="(_, slotName) in $slots" #[slotName]="slotProps">
              <slot :name="slotName" v-bind="slotProps" />
            </template>
          </CalendarPanel>
        </ArkDatePicker.Content>
      </ArkDatePicker.Positioner>
    </Teleport>
  </ArkDatePicker.Root>
</template>
