<script setup lang="ts">
import { ark } from '@ark-ui/vue/factory'

const props = defineProps<{ label: string, disabled?: boolean, readOnly?: boolean, placeholder?: string }>()
const value = defineModel<string>({ default: '' })
</script>

<template>
  <ark.input :value="value" data-toolbar-control :aria-label="props.label" :disabled="props.disabled" :readonly="props.readOnly" :placeholder="props.placeholder" tabindex="-1" class="relative h-8 min-w-0 rounded-md bg-base px-2 text-size-sm text-default outline-none ring-1 ring-control-border placeholder:text-placeholder max-sm:text-16px disabled:opacity-50" :class="!props.disabled && !props.readOnly ? 'focus:z-2 focus:ring-2 focus:ring-focus' : undefined" @input="value = ($event.target as HTMLInputElement).value" />
</template>
