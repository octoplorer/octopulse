<script setup lang="ts">
import { Switch } from './ui/switch'

defineProps<{ label: string, description?: string, disabled?: boolean }>()
const value = defineModel<boolean>({ default: false })
</script>

<template>
  <Switch v-model="value" :label="label" :description="description" :disabled="disabled" />
</template>
