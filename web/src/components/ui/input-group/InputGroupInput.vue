<script setup lang="ts">
import { Input } from '../input'
import { useInputGroup } from './context'

defineProps<{ disabled?: boolean }>()
const value = defineModel<string | number>()
const group = useInputGroup()
</script>

<template>
  <Input v-model="value" :size="group?.size.value" :disabled="disabled || group?.disabled.value" class="min-w-0 flex-1 text-ellipsis" />
</template>
