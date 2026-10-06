<script setup lang="ts">
import type { SidebarSlidingViewProps } from './section-types'
import { computed, ref } from 'vue'
import { useSidebarSlidingViewActive } from './sliding-view-context'

const props = defineProps<SidebarSlidingViewProps>()
const activeKey = useSidebarSlidingViewActive()
const element = ref<HTMLDivElement | null>(null)
const isActive = computed(() => activeKey.value === props.value)

defineExpose({ element })
</script>

<template>
  <div
    ref="element"
    data-sidebar="sliding-view"
    :data-value="value"
    :aria-hidden="!isActive"
    :inert="!isActive || undefined"
    class="flex w-full min-h-0 shrink-0 flex-col"
    :class="{ 'pointer-events-none': !isActive }"
  >
    <slot />
  </div>
</template>
