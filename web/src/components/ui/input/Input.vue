<script setup lang="ts">
import type { InputSize } from './styles'
import { Field } from '@ark-ui/vue/field'
import { inputClasses, inputSizes } from './styles'

withDefaults(defineProps<{ size?: InputSize, variant?: 'default' | 'error' }>(), { size: 'base', variant: 'default' })
const value = defineModel<string | number>()
</script>

<template>
  <Field.Input v-model="value" :class="[inputClasses, inputSizes[size], variant === 'error' && 'ring-action-danger focus:ring-action-danger']" v-bind="variant === 'error' ? { 'aria-invalid': true } : {}" />
</template>
