<script setup lang="ts">
import { computed } from 'vue'
import { Checkbox } from '../checkbox'
import TableCell from './TableCell.vue'

const props = withDefaults(defineProps<{ label?: string, disabled?: boolean, indeterminate?: boolean, sticky?: 'left' | 'right' }>(), { label: 'Select row' })
const emit = defineEmits<{ checkedChange: [checked: boolean] }>()
const checked = defineModel<boolean>({ default: false })
const state = computed(() => props.indeterminate ? 'indeterminate' : checked.value)
function select(value: boolean | 'indeterminate' | undefined) {
  checked.value = value === true
  emit('checkedChange', checked.value)
}
</script>

<template>
  <TableCell :sticky="props.sticky" class="w-10 leading-none">
    <Checkbox :model-value="state" :disabled="props.disabled" :aria-label="props.label" class="relative before:absolute before:-inset-3 before:content-empty" @update:model-value="select" />
  </TableCell>
</template>
