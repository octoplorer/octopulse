import type { Ref } from 'vue'
import { onScopeDispose, ref, watch } from 'vue'

export function useOpenChangeComplete(
  open: Ref<boolean>,
  duration: Ref<number>,
  onComplete: (open: boolean) => void,
  active: Ref<boolean> = ref(true),
) {
  let pendingOpen: boolean | undefined
  let timeout: ReturnType<typeof setTimeout> | undefined

  function clearPending() {
    clearTimeout(timeout)
    pendingOpen = undefined
  }

  function complete() {
    if (pendingOpen === undefined)
      return
    const next = pendingOpen
    clearPending()
    onComplete(next)
  }

  watch([open, active], ([next, enabled], [previous, wasEnabled]) => {
    clearPending()
    if (!enabled || !wasEnabled || next === previous)
      return

    pendingOpen = next
    const reducedMotion = typeof window !== 'undefined'
      && window.matchMedia('(prefers-reduced-motion: reduce)').matches
    if (duration.value === 0 || reducedMotion) {
      complete()
      return
    }
    timeout = setTimeout(complete, duration.value + 50)
  }, { flush: 'post' })

  onScopeDispose(clearPending)
  return { complete }
}
