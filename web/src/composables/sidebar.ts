import type { Ref } from 'vue'
import { useLocalStorage } from '@vueuse/core'
import { computed, onScopeDispose, ref, watch } from 'vue'

export const SIDEBAR_DEFAULT_WIDTH = 256
export const SIDEBAR_MIN_WIDTH = 200
export const SIDEBAR_MAX_WIDTH = 400
export const SIDEBAR_COLLAPSED_WIDTH = 57
export const SIDEBAR_RESIZE_STEP = 10

export function useSidebar(mobile: Ref<boolean>) {
  const open = useLocalStorage('octopulse.sidebar.open', true)
  const width = useLocalStorage('octopulse.sidebar.width', SIDEBAR_DEFAULT_WIDTH)
  const mobileOpen = ref(false)
  const isResizing = ref(false)
  const hovered = ref(false)
  const focused = ref(false)
  const peekDismissed = ref(false)
  let resizeOrigin = 0
  let resizeWidth = 0

  watch(width, (value) => {
    const validWidth = Number.isFinite(value)
      ? Math.min(SIDEBAR_MAX_WIDTH, Math.max(SIDEBAR_MIN_WIDTH, Math.round(value)))
      : SIDEBAR_DEFAULT_WIDTH
    if (value !== validWidth)
      width.value = validWidth
  }, { immediate: true, flush: 'sync' })

  const state = computed<'expanded' | 'collapsed' | 'peeking'>(() => {
    if (mobile.value)
      return mobileOpen.value ? 'expanded' : 'collapsed'
    if (open.value)
      return 'expanded'
    return !isResizing.value && !peekDismissed.value && (hovered.value || focused.value)
      ? 'peeking'
      : 'collapsed'
  })
  const railWidth = computed(() => mobile.value ? 0 : open.value ? width.value : SIDEBAR_COLLAPSED_WIDTH)
  const panelWidth = computed(() => mobile.value || state.value !== 'collapsed' ? width.value : SIDEBAR_COLLAPSED_WIDTH)

  function resetTransientState() {
    mobileOpen.value = false
    isResizing.value = false
    hovered.value = false
    focused.value = false
    peekDismissed.value = false
  }

  watch(mobile, resetTransientState, { flush: 'sync' })
  onScopeDispose(resetTransientState)

  function releasePeekDismissal() {
    if (!hovered.value && !focused.value)
      peekDismissed.value = false
  }

  function setHovered(value: boolean) {
    hovered.value = value
    releasePeekDismissal()
  }

  function setFocused(value: boolean) {
    focused.value = value
    releasePeekDismissal()
  }

  function dismissPeek() {
    peekDismissed.value = hovered.value || focused.value
  }

  function toggle() {
    if (mobile.value) {
      mobileOpen.value = !mobileOpen.value
      return
    }
    dismissPeek()
    open.value = !open.value
  }

  function closeMobile() {
    mobileOpen.value = false
  }

  function beginResize(clientX: number) {
    if (mobile.value || !Number.isFinite(clientX))
      return
    resizeOrigin = clientX
    resizeWidth = panelWidth.value
    if (state.value === 'peeking')
      open.value = true
    isResizing.value = true
    dismissPeek()
  }

  function resizeTo(clientX: number) {
    if (!isResizing.value || !Number.isFinite(clientX))
      return
    const nextWidth = resizeWidth + clientX - resizeOrigin
    if (nextWidth < SIDEBAR_MIN_WIDTH) {
      open.value = false
      return
    }
    width.value = Math.min(SIDEBAR_MAX_WIDTH, nextWidth)
    open.value = true
  }

  function endResize() {
    if (!isResizing.value)
      return
    dismissPeek()
    isResizing.value = false
  }

  function resizeWithKey(key: string) {
    if (mobile.value || !['ArrowLeft', 'ArrowRight', 'Home', 'End'].includes(key))
      return false
    dismissPeek()
    if (key === 'Home' || (key === 'ArrowLeft' && width.value <= SIDEBAR_MIN_WIDTH)) {
      open.value = false
    }
    else if (key === 'End') {
      width.value = SIDEBAR_MAX_WIDTH
      open.value = true
    }
    else if (key === 'ArrowRight') {
      width.value = open.value ? Math.min(SIDEBAR_MAX_WIDTH, width.value + SIDEBAR_RESIZE_STEP) : SIDEBAR_MIN_WIDTH
      open.value = true
    }
    else if (open.value) {
      width.value = Math.max(SIDEBAR_MIN_WIDTH, width.value - SIDEBAR_RESIZE_STEP)
    }
    return true
  }

  return {
    open,
    width,
    mobileOpen,
    isResizing,
    state,
    railWidth,
    panelWidth,
    toggle,
    closeMobile,
    setHovered,
    setFocused,
    dismissPeek,
    beginResize,
    resizeTo,
    endResize,
    resizeWithKey,
  }
}
