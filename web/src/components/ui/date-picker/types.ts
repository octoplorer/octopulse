import type { DatePickerDateRangePreset } from '@ark-ui/vue/date-picker'

export interface DatePickerPreset {
  label: string
  value: DatePickerDateRangePreset | string[]
}
