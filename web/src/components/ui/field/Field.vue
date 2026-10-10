<script setup lang="ts">
import { Field as ArkField } from '@ark-ui/vue/field'
import FieldDescription from './FieldDescription.vue'
import FieldError from './FieldError.vue'
import FieldLabel from './FieldLabel.vue'

defineProps<{ label?: string, description?: string, hint?: string, error?: string, required?: boolean, disabled?: boolean, readOnly?: boolean }>()
</script>

<template>
  <ArkField.Root class="field flex min-w-0 flex-col gap-1.5" :invalid="!!error" :required="required" :disabled="disabled" :read-only="readOnly">
    <ArkField.Label v-if="label || $slots.label" as-child>
      <FieldLabel as="label">
        <slot name="label">
          {{ label }}
        </slot>
        <span v-if="required" class="ml-1 text-fg-danger" aria-hidden="true">*</span>
      </FieldLabel>
    </ArkField.Label>
    <slot />
    <ArkField.HelperText v-if="description || hint || $slots.description" as-child>
      <FieldDescription>
        <slot name="description">
          {{ description || hint }}
        </slot>
      </FieldDescription>
    </ArkField.HelperText>
    <ArkField.ErrorText v-if="error" as-child>
      <FieldError>{{ error }}</FieldError>
    </ArkField.ErrorText>
  </ArkField.Root>
</template>
