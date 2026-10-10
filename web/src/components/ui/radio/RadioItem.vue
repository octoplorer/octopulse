<script setup lang="ts">
import { RadioGroup as ArkRadio } from '@ark-ui/vue/radio-group'
import { cva } from 'cva'
import { computed, inject } from 'vue'
import { radioAppearanceKey } from './context'

const props = withDefaults(defineProps<{
  value: string
  label?: string
  description?: string
  disabled?: boolean
  appearance?: 'default' | 'card' | 'segmented'
  controlPosition?: 'start' | 'end'
  variant?: 'default' | 'error'
}>(), { disabled: undefined })
const group = inject(radioAppearanceKey, undefined)
const appearance = computed(() => props.appearance ?? group?.value.appearance ?? 'default')
const controlPosition = computed(() => props.controlPosition ?? group?.value.controlPosition ?? (appearance.value === 'card' ? 'end' : 'start'))
const itemClasses = cva({
  base: 'group inline-flex min-h-24px min-w-24px cursor-pointer items-start gap-2.5 text-size-base text-default data-[disabled]:cursor-not-allowed data-[disabled]:opacity-50',
  variants: {
    appearance: { default: '', card: 'w-full rounded-lg border border-hairline bg-base p-3 hover:bg-elevated data-[state=checked]:border-interact data-[state=checked]:bg-tint', segmented: 'flex-1 justify-center rounded-md px-3 py-1.5 font-medium transition-colors data-[state=checked]:bg-contrast data-[state=checked]:text-inverse data-[focus-visible]:outline-2 data-[focus-visible]:outline-solid data-[focus-visible]:outline-brand data-[focus-visible]:outline-offset-2' },
    controlPosition: { start: '', end: 'flex-row-reverse justify-between' },
  },
  defaultVariants: { appearance: 'default', controlPosition: 'start' },
})
</script>

<template>
  <ArkRadio.Item :value="value" :disabled="disabled" :class="itemClasses({ appearance, controlPosition })">
    <ArkRadio.ItemControl v-if="appearance !== 'segmented'" class="mt-0.5 flex size-4 shrink-0 items-center justify-center rounded-full bg-base ring-1 ring-control-border outline-none data-[state=checked]:bg-contrast data-[invalid]:ring-action-danger data-[focus]:ring-focus! data-[focus-visible]:ring-2 data-[focus]:data-[focus-visible]:ring-brand!" :class="[variant === 'error' && 'ring-action-danger', appearance === 'default' && 'data-[focus]:ring-2']">
      <span class="size-2 rounded-full bg-base opacity-0 group-data-[state=checked]:opacity-100" />
    </ArkRadio.ItemControl>
    <ArkRadio.ItemText class="min-w-0 [overflow-wrap:anywhere]">
      <span class="block font-medium"><slot>{{ label }}</slot></span><span v-if="description || $slots.description" class="mt-1 block text-size-sm text-subtle"><slot name="description">{{ description }}</slot></span>
    </ArkRadio.ItemText>
    <ArkRadio.ItemHiddenInput />
  </ArkRadio.Item>
</template>
