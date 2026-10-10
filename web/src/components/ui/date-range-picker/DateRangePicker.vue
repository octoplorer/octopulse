<script setup lang="ts">
import { provide } from 'vue'
import { DatePicker } from '../date-picker'
import { calendarAppearanceKey } from '../date-picker/context'

defineOptions({ inheritAttrs: false })
withDefaults(defineProps<{ inline?: boolean, numberOfMonths?: number, clearable?: boolean, showOutsideDays?: boolean }>(), { inline: false, numberOfMonths: 2, clearable: true, showOutsideDays: true })
provide(calendarAppearanceKey, 'date-range')
const value = defineModel<string[]>({ default: () => [] })
const open = defineModel<boolean>('open', { default: false })
</script>

<template>
  <DatePicker v-bind="$attrs" v-model="value" v-model:open="open" mode="range" :inline="inline" :number-of-months="numberOfMonths" :clearable="clearable" :show-outside-days="showOutsideDays">
    <template v-for="(_, name) in $slots" #[name]="slotProps">
      <slot :name="name" v-bind="slotProps" />
    </template>
  </DatePicker>
</template>
