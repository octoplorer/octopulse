<script setup lang="ts">
import { FieldDescription, FieldError, FieldLabel } from './ui/field'

defineProps<{ label: string, hint?: string, error?: string }>()
</script>

<template>
  <label class="field" flex="~ col gap-7px" min-w="0" :data-invalid="error ? '' : undefined">
    <FieldLabel>{{ label }}</FieldLabel>
    <slot />
    <FieldDescription v-if="hint">{{ hint }}</FieldDescription>
    <FieldError v-if="error">{{ error }}</FieldError>
  </label>
</template>
