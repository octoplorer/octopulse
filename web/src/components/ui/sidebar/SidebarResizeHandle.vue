<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue'
import { useSidebar } from './context'

defineOptions({ inheritAttrs: false })

const emit = defineEmits<{
  pointerdown: [event: PointerEvent]
  keydown: [event: KeyboardEvent]
}>()
const {
  side,
  resizable,
  isMobile,
  setIsResizing,
  setWidth,
  setOpen,
  open,
  minWidth,
  width: currentWidth,
  maxWidth,
} = useSidebar()
const element = ref<HTMLButtonElement | null>(null)
const KEYBOARD_STEP = 10
let stopResize: (() => void) | undefined

function handlePointerDown(event: PointerEvent) {
  emit('pointerdown', event)
  if (event.defaultPrevented || event.button !== 0 || !resizable.value || isMobile.value || element.value?.disabled)
    return

  event.preventDefault()
  stopResize?.()
  setIsResizing(true)

  const pointerId = event.pointerId
  const startX = event.clientX
  let wasCollapsed = !open.value
  const wrapper = element.value?.closest('[data-sidebar-wrapper]')
  const sidebar = wrapper?.querySelector('[data-sidebar="sidebar"]')
  const startWidth = sidebar?.getBoundingClientRect().width ?? (open.value ? currentWidth.value : 0)

  function handlePointerMove(moveEvent: PointerEvent) {
    if (moveEvent.pointerId !== pointerId)
      return

    const delta = side.value === 'left'
      ? moveEvent.clientX - startX
      : startX - moveEvent.clientX
    const rawWidth = startWidth + delta

    if (wasCollapsed) {
      if (rawWidth >= minWidth.value) {
        wasCollapsed = false
        setOpen(true)
        setWidth(Math.min(rawWidth, maxWidth.value))
      }
      return
    }

    if (rawWidth < minWidth.value) {
      setOpen(false)
      wasCollapsed = true
      return
    }

    setWidth(Math.min(rawWidth, maxWidth.value))
  }

  function handlePointerEnd(endEvent: PointerEvent) {
    if (endEvent.pointerId === pointerId)
      stopResize?.()
  }

  stopResize = () => {
    setIsResizing(false)
    document.removeEventListener('pointermove', handlePointerMove)
    document.removeEventListener('pointerup', handlePointerEnd)
    document.removeEventListener('pointercancel', handlePointerEnd)
    stopResize = undefined
  }

  document.addEventListener('pointermove', handlePointerMove)
  document.addEventListener('pointerup', handlePointerEnd)
  document.addEventListener('pointercancel', handlePointerEnd)
}

function handleKeyDown(event: KeyboardEvent) {
  emit('keydown', event)
  if (event.defaultPrevented || !resizable.value || isMobile.value || element.value?.disabled)
    return

  const grow = side.value === 'left' ? 'ArrowRight' : 'ArrowLeft'
  const shrink = side.value === 'left' ? 'ArrowLeft' : 'ArrowRight'

  if (event.key === grow) {
    event.preventDefault()
    if (!open.value) {
      setOpen(true)
      setWidth(minWidth.value)
    }
    else {
      setWidth(Math.min(currentWidth.value + KEYBOARD_STEP, maxWidth.value))
    }
  }
  else if (event.key === shrink) {
    event.preventDefault()
    const next = currentWidth.value - KEYBOARD_STEP
    if (next < minWidth.value)
      setOpen(false)
    else
      setWidth(next)
  }
  else if (event.key === 'Home') {
    event.preventDefault()
    setOpen(false)
  }
  else if (event.key === 'End') {
    event.preventDefault()
    setOpen(true)
    setWidth(maxWidth.value)
  }
}

watch([resizable, isMobile], ([enabled, mobile]) => {
  if (!enabled || mobile)
    stopResize?.()
})
onBeforeUnmount(() => stopResize?.())

defineExpose({ element })
</script>

<template>
  <button
    v-if="resizable && !isMobile"
    ref="element"
    type="button"
    aria-label="Resize sidebar"
    :tabindex="0"
    data-sidebar="resize-handle"
    class="absolute inset-y-0 z-2 hidden w-3 cursor-col-resize sm:block after:absolute after:inset-y-0 after:w-0.5 after:bg-transparent after:transition-colors hover:after:bg-hairline focus-visible:after:bg-hairline active:after:bg-hairline focus:outline-none"
    :class="side === 'left' ? 'right-0 after:right-0' : 'left-0 after:left-0'"
    v-bind="$attrs"
    @pointerdown="handlePointerDown"
    @keydown="handleKeyDown"
  >
    <slot />
  </button>
</template>
