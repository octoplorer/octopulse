<script setup lang="ts">
import { useMutationObserver, useResizeObserver } from '@vueuse/core'
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

defineOptions({ name: 'SidebarContent' })

const element = ref<HTMLDivElement | null>(null)
const viewport = ref<HTMLDivElement | null>(null)
const content = ref<HTMLDivElement | null>(null)
const overflowStart = ref(0)
const overflowEnd = ref(0)
const isScrolling = ref(false)
let scrollTimeout: ReturnType<typeof setTimeout> | undefined

const viewportStyle = computed(() => ({
  '--scroll-area-overflow-y-start': `${overflowStart.value}px`,
  '--scroll-area-overflow-y-end': `${overflowEnd.value}px`,
}))

function updateOverflow() {
  const node = viewport.value
  if (!node)
    return
  overflowStart.value = Math.max(0, node.scrollTop)
  overflowEnd.value = Math.max(0, node.scrollHeight - node.clientHeight - node.scrollTop)
}

function handleScroll() {
  updateOverflow()
  isScrolling.value = true
  clearTimeout(scrollTimeout)
  scrollTimeout = setTimeout(() => {
    isScrolling.value = false
  }, 150)
}

useResizeObserver([viewport, content], updateOverflow)
useMutationObserver(content, updateOverflow, {
  childList: true,
  subtree: true,
  characterData: true,
})
onMounted(updateOverflow)
onBeforeUnmount(() => clearTimeout(scrollTimeout))

defineExpose({ element, viewport })
</script>

<template>
  <div
    ref="element"
    data-sidebar="content"
    :data-scrolling="isScrolling || undefined"
    class="group/sidebar-content relative min-w-0 flex-1 overflow-hidden"
  >
    <div
      ref="viewport"
      data-sidebar="viewport"
      :tabindex="-1"
      :style="viewportStyle"
      class="
        h-full overflow-x-hidden overflow-y-auto px-[11px] py-3 transition-[padding] duration-$sidebar-animation-duration group-not-data-[state=collapsed]/sidebar:px-3.5
        [mask-image:linear-gradient(to_bottom,transparent_0,black_min(24px,var(--scroll-area-overflow-y-start)),black_calc(100%_-_min(24px,var(--scroll-area-overflow-y-end))),transparent_100%)]
        [scrollbar-width:thin] [scrollbar-color:transparent_transparent]
        group-hover/sidebar-content:[scrollbar-color:var(--color-line)_transparent]
        group-data-[scrolling]/sidebar-content:[scrollbar-color:var(--color-line)_transparent]
        [&::-webkit-scrollbar]:w-1.5
        [&::-webkit-scrollbar-thumb]:[border:1px_solid_transparent]
        [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-transparent [&::-webkit-scrollbar-thumb]:bg-clip-padding
        group-hover/sidebar-content:[&::-webkit-scrollbar-thumb]:bg-line
        group-data-[scrolling]/sidebar-content:[&::-webkit-scrollbar-thumb]:bg-line
        group-data-[state=collapsed]/sidebar:[scrollbar-width:none]
        group-data-[state=collapsed]/sidebar:[&::-webkit-scrollbar]:hidden
      "
      @scroll="handleScroll"
    >
      <div ref="content" class="flex min-w-0 flex-col">
        <slot />
      </div>
    </div>
  </div>
</template>
