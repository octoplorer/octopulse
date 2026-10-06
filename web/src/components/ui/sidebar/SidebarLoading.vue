<script setup lang="ts">
import type { SidebarLoadingProps } from './menu-types'
import { SkeletonLine } from '../loader/skeleton-line'

defineOptions({ name: 'SidebarLoading' })

withDefaults(defineProps<SidebarLoadingProps>(), { label: 'Loading' })

const groups = [
  ['w-28', 'w-40', 'w-24'],
  ['w-24', 'w-36', 'w-32'],
]
</script>

<template>
  <div
    data-sidebar="loading"
    role="status"
    :aria-label="label"
    class="flex min-h-0 w-full flex-1 flex-col gap-4 px-2 py-3"
  >
    <div v-for="(widths, groupIndex) in groups" :key="groupIndex" class="flex flex-col gap-0.5">
      <SkeletonLine
        class="mb-1 ml-2 h-2 w-16 rounded-full group-data-[state=collapsed]/sidebar:hidden"
      />
      <div
        v-for="(width, itemIndex) in widths"
        :key="itemIndex"
        class="flex min-h-8.5 items-center gap-3 rounded-lg px-3 group-data-[state=collapsed]/sidebar:justify-center group-data-[state=collapsed]/sidebar:px-0"
      >
        <SkeletonLine class="size-4.5 shrink-0 rounded-md" />
        <SkeletonLine
          class="h-2.5 rounded-full group-data-[state=collapsed]/sidebar:hidden"
          :class="width"
        />
      </div>
    </div>
  </div>
</template>
