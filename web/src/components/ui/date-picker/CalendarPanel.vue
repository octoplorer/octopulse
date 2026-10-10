<script setup lang="ts">
import type { DatePickerPreset } from './types'
import { DatePicker as ArkDatePicker, parseDate } from '@ark-ui/vue/date-picker'
import { inject } from 'vue'
import CalendarDayCell from './CalendarDayCell.vue'
import { calendarAppearanceKey } from './context'

withDefaults(defineProps<{
  numberOfMonths?: number
  showOutsideDays?: boolean
  showWeekNumbers?: boolean
  presets?: DatePickerPreset[]
  clearable?: boolean
  clearLabel?: string
  timezone?: string
  size?: 'sm' | 'base' | 'lg'
}>(), { numberOfMonths: 1, showOutsideDays: true, clearLabel: 'Reset', size: 'base' })
const appearance = inject(calendarAppearanceKey, 'calendar')
const navigationClasses = 'flex size-7 items-center justify-center rounded-md text-default outline-none hover:bg-tint focus-visible:ring-2 focus-visible:ring-brand focus-visible:ring-inset disabled:opacity-35'
</script>

<template>
  <div class="calendar-panel rounded-xl bg-base p-3 text-default tabular-nums" :class="appearance === 'date-range' && 'calendar-range-panel'">
    <ArkDatePicker.Context v-slot="calendar">
      <div v-if="presets?.length" class="mb-3 flex flex-wrap gap-1.5 border-b border-hairline pb-3">
        <ArkDatePicker.PresetTrigger v-for="preset in presets" :key="preset.label" :value="Array.isArray(preset.value) ? preset.value.map(value => parseDate(value)) : preset.value" class="rounded-md bg-tint px-2 py-1 text-size-xs font-medium outline-none hover:bg-fill-hover focus-visible:ring-2 focus-visible:ring-brand focus-visible:ring-inset">
          {{ preset.label }}
        </ArkDatePicker.PresetTrigger>
      </div>
      <ArkDatePicker.View view="day">
        <ArkDatePicker.ViewControl class="mb-3 flex items-center justify-between gap-2">
          <ArkDatePicker.PrevTrigger :class="navigationClasses">
            <span class="i-lucide-chevron-left size-4" aria-hidden="true" />
          </ArkDatePicker.PrevTrigger>
          <ArkDatePicker.ViewTrigger class="rounded-md px-2 py-1 text-size-base font-medium outline-none hover:bg-tint focus-visible:ring-2 focus-visible:ring-brand focus-visible:ring-inset">
            <ArkDatePicker.RangeText />
          </ArkDatePicker.ViewTrigger>
          <ArkDatePicker.NextTrigger :class="navigationClasses">
            <span class="i-lucide-chevron-right size-4" aria-hidden="true" />
          </ArkDatePicker.NextTrigger>
        </ArkDatePicker.ViewControl>
        <div class="flex flex-wrap gap-4">
          <div v-for="(_, monthIndex) in numberOfMonths" :key="monthIndex">
            <p v-if="numberOfMonths > 1" class="mb-2 text-center text-size-sm font-medium">
              {{ calendar.format(calendar.getOffset({ months: monthIndex }).visibleRange.start, { month: 'long', year: 'numeric' }) }}
            </p>
            <ArkDatePicker.Table :id="`calendar-${monthIndex}`" class="border-collapse" view="day">
              <ArkDatePicker.TableHead>
                <ArkDatePicker.TableRow>
                  <ArkDatePicker.WeekNumberHeaderCell v-if="showWeekNumbers" class="px-1 text-size-xs text-subtle">
                    #
                  </ArkDatePicker.WeekNumberHeaderCell>
                  <ArkDatePicker.TableHeader v-for="day in calendar.weekDays" :key="day.short" class="h-7 text-center text-size-xs font-normal text-subtle">
                    <abbr :title="day.long" class="no-underline">{{ day.narrow }}</abbr>
                  </ArkDatePicker.TableHeader>
                </ArkDatePicker.TableRow>
              </ArkDatePicker.TableHead>
              <ArkDatePicker.TableBody>
                <ArkDatePicker.TableRow v-for="(week, weekIndex) in calendar.getMonthWeeks(calendar.getOffset({ months: monthIndex }).visibleRange.start)" :key="weekIndex">
                  <ArkDatePicker.WeekNumberCell v-if="showWeekNumbers" :week="week" :week-index="weekIndex" class="px-1 text-size-xs text-subtle">
                    {{ calendar.getWeekNumber(week) }}
                  </ArkDatePicker.WeekNumberCell>
                  <CalendarDayCell v-for="day in week" :key="day.toString()" :value="day" :visible-range="{ start: calendar.getOffset({ months: monthIndex }).visibleRange.start, end: calendar.getOffset({ months: monthIndex }).visibleRange.start.add({ months: 1 }).subtract({ days: 1 }) }" :show-outside-days="showOutsideDays" :size="size">
                    <template #default="dayProps">
                      <slot name="day" v-bind="dayProps">
                        {{ day.day }}
                      </slot>
                    </template>
                  </CalendarDayCell>
                </ArkDatePicker.TableRow>
              </ArkDatePicker.TableBody>
            </ArkDatePicker.Table>
          </div>
        </div>
      </ArkDatePicker.View>
      <ArkDatePicker.View v-for="view in ['month', 'year'] as const" :key="view" :view="view">
        <ArkDatePicker.ViewControl class="mb-3 flex items-center justify-between gap-2">
          <ArkDatePicker.PrevTrigger :class="navigationClasses">
            <span class="i-lucide-chevron-left size-4" aria-hidden="true" />
          </ArkDatePicker.PrevTrigger>
          <ArkDatePicker.ViewTrigger class="rounded-md px-2 py-1 text-size-base font-medium outline-none hover:bg-tint focus-visible:ring-2 focus-visible:ring-brand focus-visible:ring-inset">
            <ArkDatePicker.RangeText />
          </ArkDatePicker.ViewTrigger>
          <ArkDatePicker.NextTrigger :class="navigationClasses">
            <span class="i-lucide-chevron-right size-4" aria-hidden="true" />
          </ArkDatePicker.NextTrigger>
        </ArkDatePicker.ViewControl>
        <ArkDatePicker.Table :view="view" :columns="4" class="w-full border-collapse">
          <ArkDatePicker.TableBody>
            <ArkDatePicker.TableRow v-for="(row, rowIndex) in view === 'month' ? calendar.getMonthsGrid({ columns: 4, format: 'short' }) : calendar.getYearsGrid({ columns: 4 })" :key="rowIndex">
              <ArkDatePicker.TableCell v-for="cell in row" :key="cell.value" :value="cell.value" :columns="4" class="p-1">
                <ArkDatePicker.TableCellTrigger class="h-9 w-full rounded-md px-3 text-size-sm outline-none hover:bg-[var(--calendar-hover)] data-[selected]:bg-[var(--calendar-accent)] data-[selected]:text-[var(--calendar-accent-text)] data-[selected]:hover:bg-[var(--calendar-accent)] data-[disabled]:opacity-40 data-[disabled]:hover:bg-transparent focus-visible:ring-2 focus-visible:ring-brand focus-visible:ring-inset">
                  {{ cell.label }}
                </ArkDatePicker.TableCellTrigger>
              </ArkDatePicker.TableCell>
            </ArkDatePicker.TableRow>
          </ArkDatePicker.TableBody>
        </ArkDatePicker.Table>
      </ArkDatePicker.View>
      <div v-if="clearable || timezone || $slots.footer" class="mt-3 flex items-center justify-between gap-3 border-t border-hairline pt-3 text-size-xs text-subtle">
        <slot name="footer">
          <span>{{ timezone }}</span>
        </slot>
        <ArkDatePicker.ClearTrigger v-if="clearable" class="rounded-md px-2 py-1 text-default outline-none hover:bg-tint focus-visible:ring-2 focus-visible:ring-brand focus-visible:ring-inset">
          {{ clearLabel }}
        </ArkDatePicker.ClearTrigger>
      </div>
    </ArkDatePicker.Context>
  </div>
</template>

<style scoped>
.calendar-panel {
  --calendar-accent: oklch(20.5% 0 0);
  --calendar-accent-text: oklch(97% 0 0);
  --calendar-range-middle: oklch(92.2% 0 0);
  --calendar-range-middle-text: oklch(21% 0.006 285.885);
  --calendar-day: oklch(21% 0.006 285.885);
  --calendar-subtle: oklch(55.6% 0 0);
  --calendar-today: oklch(54.6% 0.215 262.881);
  --calendar-hover: oklch(90% 0 0);
  --calendar-fill-hover: oklch(87% 0 0);
}

:global([data-mode='dark']) .calendar-panel {
  --calendar-accent: oklch(97% 0 0);
  --calendar-accent-text: oklch(20.5% 0 0);
  --calendar-range-middle: oklch(28% 0 0);
  --calendar-range-middle-text: oklch(97% 0 0);
  --calendar-day: oklch(97% 0 0);
  --calendar-subtle: oklch(70% 0 0);
  --calendar-today: oklch(62.3% 0.214 259.815);
  --calendar-hover: oklch(20% 0 0);
  --calendar-fill-hover: oklch(25% 0 0);
}

.calendar-panel.calendar-range-panel {
  --calendar-accent: var(--color-contrast);
  --calendar-accent-text: var(--text-color-inverse);
  --calendar-range-middle: var(--color-interact);
  --calendar-range-middle-text: var(--text-color-default);
  --calendar-day: var(--text-color-default);
  --calendar-subtle: var(--text-color-subtle);
  --calendar-hover: var(--color-interact);
  --calendar-fill-hover: var(--color-interact);
}
</style>
