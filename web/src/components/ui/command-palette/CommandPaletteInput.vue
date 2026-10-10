<script setup lang="ts">
import { Combobox as ArkCombobox } from '@ark-ui/vue/combobox'

const props = withDefaults(defineProps<{ label?: string, placeholder?: string }>(), { label: 'Search commands', placeholder: 'Search…' })
</script>

<template>
  <ArkCombobox.Control class="flex items-center gap-3 bg-base px-4 py-3 focus-within:ring-2 focus-within:ring-brand">
    <span class="i-lucide-search size-4 text-subtle" aria-hidden="true" />
    <ArkCombobox.Input :aria-label="props.label" :placeholder="props.placeholder" class="h-8 min-w-0 flex-1 border-none bg-transparent text-size-base text-default outline-none placeholder:text-placeholder" />
  </ArkCombobox.Control>
</template>
