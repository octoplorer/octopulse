<script setup lang="ts">
import type { VNode } from 'vue'
import type { SidebarRootProps } from './types'
import { useEventListener } from '@vueuse/core'
import { computed, defineComponent, Fragment, h, nextTick, ref, watch } from 'vue'
import { SIDEBAR_FOCUSABLE_SELECTOR } from './constants'
import { useSidebar } from './context'
import { TooltipProvider } from './placeholders'
import SidebarFooter from './SidebarFooter.vue'

defineOptions({ inheritAttrs: false })

const props = withDefaults(defineProps<SidebarRootProps>(), {
  fullScreenOnMobile: false,
})

const {
  state,
  open,
  isMobile,
  openMobile,
  setOpenMobile,
  side,
  isResizing,
  resizable,
  width,
  isPeeking,
  startPeek,
  stopPeek,
  contained,
} = useSidebar()

const element = ref<HTMLElement | null>(null)
let previousFocus: HTMLElement | null = null
let pointerInPeekZone = false

defineExpose({ element })

const expandedWidth = computed(() => resizable.value ? `${width.value}px` : 'var(--sidebar-width)')
const collapsedWidth = 'var(--sidebar-width-icon)'
const railWidth = computed(() => open.value ? expandedWidth.value : collapsedWidth)
const contentWidth = computed(() => open.value || isPeeking.value ? expandedWidth.value : collapsedWidth)
const borderClasses = computed(() => side.value === 'left' ? 'border-r border-line' : 'border-l border-line')

function closeMobile() {
  setOpenMobile(false)
}

useEventListener('keydown', (event: KeyboardEvent) => {
  if (isMobile.value && openMobile.value && event.key === 'Escape') {
    event.preventDefault()
    closeMobile()
  }
})

watch([isMobile, openMobile], async ([mobile, isOpen]) => {
  if (typeof document === 'undefined')
    return
  if (mobile && isOpen) {
    if (!previousFocus && document.activeElement instanceof HTMLElement)
      previousFocus = document.activeElement
    await nextTick()
    if (isMobile.value && openMobile.value) {
      const target = element.value?.querySelector<HTMLElement>(SIDEBAR_FOCUSABLE_SELECTOR)
      ;(target ?? element.value)?.focus()
    }
  }
  else if (previousFocus) {
    previousFocus.focus()
    previousFocus = null
  }
}, { immediate: true, flush: 'post' })

function onPeekFocusOut(event: FocusEvent) {
  if (!(event.relatedTarget instanceof Node)
    || !(event.currentTarget as HTMLElement).contains(event.relatedTarget)) {
    if (!pointerInPeekZone)
      stopPeek()
  }
}

function flattenChildren(nodes: VNode[]): VNode[] {
  return nodes.flatMap(node => node.type === Fragment && Array.isArray(node.children)
    ? flattenChildren(node.children as VNode[])
    : [node])
}

// Keep direct Footer children outside the area that opens the hover preview.
const DesktopContent = defineComponent({
  name: 'SidebarDesktopContent',
  setup(_, { slots }) {
    return () => {
      const children = flattenChildren(slots.default?.() ?? [])
      const footerChildren = children.filter(child => child.type === SidebarFooter)
      const contentChildren = children.filter(child => child.type !== SidebarFooter)

      return h('div', {
        'data-sidebar': 'content-container',
        'style': { width: contentWidth.value },
        'class': [
          'flex h-full min-w-0 flex-col overflow-hidden whitespace-nowrap bg-$sidebar-bg text-default',
          'transition-[width] duration-$sidebar-animation-duration ease-$sidebar-easing motion-reduce:transition-none',
          borderClasses.value,
          isResizing.value && '!transition-none',
          open.value ? 'relative' : [contained.value ? 'absolute' : 'fixed', 'inset-y-0 z-40', side.value === 'left' ? 'left-0' : 'right-0'],
          props.contentClass,
        ],
      }, [
        h('div', {
          'data-sidebar': 'peek-zone',
          'class': 'flex min-h-0 flex-1 flex-col',
          'onMouseenter': () => {
            pointerInPeekZone = true
            startPeek()
          },
          'onMouseleave': () => {
            pointerInPeekZone = false
            stopPeek()
          },
          'onFocusin': startPeek,
          'onFocusout': onPeekFocusOut,
        }, contentChildren),
        ...footerChildren,
        ...(slots.footer?.() ?? []),
      ])
    }
  },
})
</script>

<template>
  <template v-if="isMobile">
    <div
      data-sidebar-backdrop=""
      aria-hidden="true"
      class="inset-0 z-40 bg-recessed transition-opacity duration-$sidebar-animation-duration ease-$sidebar-easing motion-reduce:transition-none"
      :class="[
        contained ? 'absolute' : 'fixed',
        openMobile && !fullScreenOnMobile ? 'opacity-80' : 'pointer-events-none opacity-0',
      ]"
      @click="closeMobile"
    />
    <nav
      ref="element"
      tabindex="-1"
      aria-label="Navigation"
      :aria-hidden="!openMobile"
      :inert="!openMobile"
      :data-state="openMobile ? 'expanded' : 'collapsed'"
      :data-side="side"
      data-sidebar="sidebar"
      data-mobile="true"
      class="group/sidebar inset-y-0 z-50 flex flex-col overflow-hidden bg-$sidebar-bg text-default transition-transform duration-$sidebar-animation-duration ease-$sidebar-easing motion-reduce:transition-none"
      :class="[
        contained ? 'absolute' : 'fixed',
        fullScreenOnMobile ? 'w-full' : 'w-$sidebar-width',
        !fullScreenOnMobile && (side === 'left' ? 'border-r border-line' : 'border-l border-line'),
        side === 'left' ? 'left-0' : 'right-0',
        openMobile ? 'translate-x-0' : side === 'left' ? '-translate-x-full' : 'translate-x-full',
      ]"
      v-bind="$attrs"
    >
      <slot />
      <slot name="footer" />
    </nav>
  </template>

  <aside
    v-else
    ref="element"
    :data-state="state"
    :data-side="side"
    data-sidebar="sidebar"
    :style="{ width: railWidth }"
    class="group/sidebar relative h-full shrink-0 grow-0 overflow-visible transition-[width] duration-$sidebar-animation-duration ease-$sidebar-easing motion-reduce:transition-none"
    :class="{ '!transition-none': isResizing }"
    v-bind="$attrs"
  >
    <TooltipProvider>
      <DesktopContent>
        <slot />
        <template #footer>
          <slot name="footer" />
        </template>
      </DesktopContent>
    </TooltipProvider>
  </aside>
</template>
