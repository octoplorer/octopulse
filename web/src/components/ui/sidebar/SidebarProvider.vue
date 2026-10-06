<script setup lang="ts">
import type { SidebarContextValue, SidebarProviderProps, SidebarScrollToItemOptions, SidebarState } from './types'
import { useMediaQuery } from '@vueuse/core'
import { computed, ref, toRefs } from 'vue'
import {
  SIDEBAR_ANIMATION_DURATION_MS,
  SIDEBAR_DEFAULT_WIDTH,
  SIDEBAR_EASING,
  SIDEBAR_MAX_WIDTH,
  SIDEBAR_MIN_WIDTH,
  SIDEBAR_STYLING,
} from './constants'
import { provideSidebarContext } from './context'
import { useOpenChangeComplete } from './use-open-change-complete'

const props = withDefaults(defineProps<SidebarProviderProps>(), {
  defaultOpen: true,
  open: undefined,
  side: 'left',
  resizable: false,
  defaultWidth: SIDEBAR_DEFAULT_WIDTH,
  minWidth: SIDEBAR_MIN_WIDTH,
  maxWidth: SIDEBAR_MAX_WIDTH,
  contained: false,
  peekable: false,
  animationDuration: SIDEBAR_ANIMATION_DURATION_MS,
  mobileBreakpoint: SIDEBAR_STYLING.mobile.breakpoint,
})

const emit = defineEmits<{
  'update:open': [open: boolean]
  'openChange': [open: boolean]
  'openChangeComplete': [open: boolean]
  'widthChange': [width: number]
}>()

const { side, resizable, minWidth, maxWidth, contained, peekable, animationDuration } = toRefs(props)
const isMobile = useMediaQuery(computed(() => `(max-width: ${props.mobileBreakpoint - 1}px)`))
const internalOpen = ref(props.defaultOpen)
const internalOpenMobile = ref(false)
const open = computed(() => props.open ?? internalOpen.value)
const openMobile = computed(() => isMobile.value && props.open !== undefined ? props.open : internalOpenMobile.value)
const width = ref(props.defaultWidth)
const isResizing = ref(false)
const isPeeking = ref(false)
const state = computed<SidebarState>(() => isPeeking.value ? 'peeking' : open.value ? 'expanded' : 'collapsed')
const items = new Map<string, HTMLElement>()

function setOpen(next: boolean) {
  internalOpen.value = next
  emit('update:open', next)
  emit('openChange', next)
}

function setOpenMobile(next: boolean) {
  internalOpenMobile.value = next
  if (isMobile.value && props.open !== undefined) {
    emit('update:open', next)
    emit('openChange', next)
  }
}

function setWidth(next: number) {
  width.value = Math.min(maxWidth.value, Math.max(minWidth.value, next))
  emit('widthChange', width.value)
}

function setIsResizing(next: boolean) {
  isResizing.value = next
}

function toggleSidebar() {
  if (isMobile.value) {
    setOpenMobile(!openMobile.value)
  }
  else {
    stopPeek()
    setOpen(!open.value)
  }
}

function startPeek() {
  if (peekable.value && !open.value && !isMobile.value)
    isPeeking.value = true
}

function stopPeek() {
  isPeeking.value = false
}

function registerItem(id: string, node: HTMLElement | null) {
  if (node)
    items.set(id, node)
  else
    items.delete(id)
}

function scrollToItem(id: string, options: SidebarScrollToItemOptions = {}) {
  const target = items.get(id)
  const viewport = target?.closest<HTMLElement>('[data-sidebar="viewport"]')
  if (!target || !viewport)
    return

  const { align = 'auto', behavior = 'auto' } = options
  const targetRect = target.getBoundingClientRect()
  const viewportRect = viewport.getBoundingClientRect()
  const scale = viewport.offsetHeight > 0 && viewportRect.height > 0
    ? viewportRect.height / viewport.offsetHeight
    : 1
  const offset = (targetRect.top - viewportRect.top) / scale + viewport.scrollTop
  let desired = offset

  if (align === 'center') {
    desired -= (viewport.clientHeight - target.offsetHeight) / 2
  }
  else if (align === 'end') {
    desired -= viewport.clientHeight - target.offsetHeight
  }
  else if (align === 'auto') {
    if (offset < viewport.scrollTop) {
      desired = offset
    }
    else if (offset + target.offsetHeight > viewport.scrollTop + viewport.clientHeight) {
      desired = offset + target.offsetHeight - viewport.clientHeight
    }
    else {
      return
    }
  }

  const reducedMotion = typeof window !== 'undefined'
    && window.matchMedia('(prefers-reduced-motion: reduce)').matches
  viewport.scrollTo({
    top: Math.max(0, Math.min(desired, viewport.scrollHeight - viewport.clientHeight)),
    behavior: reducedMotion ? 'auto' : behavior,
  })
}

function scrollItemIntoView(id: string, options?: SidebarScrollToItemOptions) {
  const target = items.get(id)
  const viewport = target?.closest<HTMLElement>('[data-sidebar="viewport"]')
  if (!target || !viewport)
    return

  const targetRect = target.getBoundingClientRect()
  const viewportRect = viewport.getBoundingClientRect()
  if (targetRect.top < viewportRect.top || targetRect.bottom > viewportRect.bottom)
    scrollToItem(id, options)
}

const context: SidebarContextValue = {
  state,
  open,
  setOpen,
  openMobile,
  setOpenMobile,
  isMobile,
  toggleSidebar,
  side,
  width,
  resizable,
  minWidth,
  maxWidth,
  isResizing,
  setIsResizing,
  setWidth,
  isPeeking,
  peekable,
  startPeek,
  stopPeek,
  contained,
  animationDuration,
  registerItem,
  scrollToItem,
  scrollItemIntoView,
}

provideSidebarContext(context)
defineExpose(context)

const { complete: completeDesktopOpenChange } = useOpenChangeComplete(
  open,
  animationDuration,
  next => emit('openChangeComplete', next),
  computed(() => !isMobile.value),
)
const { complete: completeMobileOpenChange } = useOpenChangeComplete(
  openMobile,
  animationDuration,
  next => emit('openChangeComplete', next),
  isMobile,
)

function handleTransitionEnd(event: TransitionEvent) {
  const target = event.target
  if (!(target instanceof HTMLElement)
    || target.closest('[data-sidebar-wrapper]') !== event.currentTarget
    || target.dataset.sidebar !== 'sidebar'
    || event.propertyName !== (isMobile.value ? 'transform' : 'width')) {
    return
  }
  if (isMobile.value)
    completeMobileOpenChange()
  else
    completeDesktopOpenChange()
}

const wrapperStyle = computed(() => ({
  '--sidebar-width': resizable.value ? `${width.value}px` : SIDEBAR_STYLING.width.expanded,
  '--sidebar-width-icon': SIDEBAR_STYLING.width.icon,
  '--sidebar-animation-duration': `${animationDuration.value}ms`,
  '--sidebar-easing': SIDEBAR_EASING,
}))
</script>

<template>
  <div
    data-sidebar-wrapper=""
    :data-state="state"
    :data-side="side"
    :style="wrapperStyle"
    class="group/sidebar-wrapper relative isolate flex w-full [--sidebar-active-bg:var(--color-tint)] [--sidebar-bg:var(--color-base)]"
    :class="{ 'min-h-svh': !contained && !isMobile, 'select-none': isResizing }"
    @transitionend="handleTransitionEnd"
  >
    <slot />
  </div>
</template>
