<script setup lang="ts">
import { computed } from 'vue'
import { usePagination } from './context'

const context = usePagination()
const start = computed(() => context.count.value === 0 ? 0 : (context.page.value - 1) * context.pageSize.value + 1)
const end = computed(() => Math.min(context.count.value ?? Number.POSITIVE_INFINITY, context.page.value * context.pageSize.value))
</script>

<template>
  <div class="text-size-sm text-subtle" aria-live="polite">
    <slot :page="context.page.value" :page-size="context.pageSize.value" :count="context.count.value" :start="start" :end="end">
      <template v-if="context.count.value != null">
        Showing <span class="tabular-nums">{{ start }}–{{ end }}</span> of <span class="tabular-nums">{{ context.count.value }}</span>
      </template>
    </slot>
  </div>
</template>
