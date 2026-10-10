import type { InjectionKey } from 'vue'

export type CalendarAppearance = 'calendar' | 'date-range'
export const calendarAppearanceKey: InjectionKey<CalendarAppearance> = Symbol('calendar-appearance')
