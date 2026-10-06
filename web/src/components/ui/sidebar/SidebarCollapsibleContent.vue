<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useSidebarCollapsible } from './collapsible-context'
import { useSidebar } from './context'

const { contentId, isOpen, autoScrollOnOpen, completeOpenChange } = useSidebarCollapsible()
const sidebar = useSidebar()
const element = ref<HTMLDivElement | null>(null)
const isContentShown = computed(() => isOpen.value && (
  sidebar.isMobile.value ? sidebar.openMobile.value : sidebar.state.value !== 'collapsed'
))

watch(
  [isContentShown, autoScrollOnOpen, sidebar.animationDuration],
  ([shown, shouldScroll, duration], _previous, onCleanup) => {
    if (!shown || !shouldScroll || typeof window === 'undefined')
      return

    const timeout = window.setTimeout(() => {
      element.value?.scrollIntoView({
        block: 'nearest',
        behavior: window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth',
      })
    }, duration)
    onCleanup(() => window.clearTimeout(timeout))
  },
  { immediate: true, flush: 'post' },
)

function handleTransitionEnd(event: TransitionEvent) {
  if (event.target === event.currentTarget && event.propertyName === 'grid-template-rows')
    completeOpenChange()
}

defineExpose({ element })
</script>

<template>
  <div
    :id="contentId"
    ref="element"
    role="region"
    :aria-hidden="!isContentShown"
    :inert="!isContentShown || undefined"
    class="grid transition-[grid-template-rows] duration-$sidebar-animation-duration ease-$sidebar-easing motion-reduce:transition-none"
    :class="isContentShown ? 'grid-rows-[1fr]' : 'grid-rows-[0fr]'"
    @transitionend="handleTransitionEnd"
  >
    <div class="overflow-hidden">
      <slot />
    </div>
  </div>
</template>
