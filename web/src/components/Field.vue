<script setup lang="ts">
import { Field as ArkField } from '@ark-ui/vue/field'
import { FieldDescription, FieldError, FieldLabel } from './ui/field'

defineProps<{ label: string, hint?: string, error?: string }>()
</script>

<template>
  <ArkField.Root class="field" flex="~ col gap-7px" min-w="0" :invalid="!!error">
    <ArkField.Label as-child>
      <FieldLabel as="label">
        {{ label }}
      </FieldLabel>
    </ArkField.Label>
    <slot />
    <ArkField.HelperText v-if="hint" as-child>
      <FieldDescription>{{ hint }}</FieldDescription>
    </ArkField.HelperText>
    <ArkField.ErrorText v-if="error" as-child>
      <FieldError>{{ error }}</FieldError>
    </ArkField.ErrorText>
  </ArkField.Root>
</template>
