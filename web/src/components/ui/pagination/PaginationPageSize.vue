<script setup lang="ts">
import { usePagination } from './context'

const props = withDefaults(defineProps<{ options?: number[], label?: string }>(), { options: () => [25, 50, 100, 250] })
const context = usePagination()
</script>

<template>
  <label class="flex items-center gap-2 text-size-sm text-subtle">
    <span>{{ props.label ?? context.labels.value.perPage }}</span>
    <select :value="context.pageSize.value" :aria-label="context.labels.value.pageSize" class="h-7 rounded-md bg-base px-2 text-default outline-none ring-1 ring-line focus:ring-focus/50 focus-visible:ring-2 focus-visible:ring-brand" @change="context.setPageSize(Number(($event.target as HTMLSelectElement).value))"><option v-for="size in props.options" :key="size" :value="size">{{ size }}</option></select>
  </label>
</template>
