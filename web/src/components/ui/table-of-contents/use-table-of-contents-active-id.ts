import type { MaybeRefOrGetter } from 'vue'
import { onBeforeUnmount, onMounted, shallowRef, toValue, watch } from 'vue'

export interface UseTableOfContentsActiveIdOptions {
  ids: MaybeRefOrGetter<string[]>
  offset?: MaybeRefOrGetter<number>
  root?: MaybeRefOrGetter<Element | null>
  trackHash?: boolean
}
export function useTableOfContentsActiveId(options: UseTableOfContentsActiveIdOptions) {
  const activeId = shallowRef<string | null>(null)
  let observer: IntersectionObserver | undefined
  let pinned = false
  let timer: ReturnType<typeof setTimeout> | undefined
  let scrollTarget: EventTarget | undefined
  function settle() {
    clearTimeout(timer)
    timer = setTimeout(() => {
      pinned = false
    }, 150)
  }
  function trackHash() {
    if (options.trackHash === false)
      return
    let id = window.location.hash.slice(1)
    try {
      id = decodeURIComponent(id)
    }
    catch {
      return
    }
    if (toValue(options.ids).includes(id))
      activeId.value = id
  }
  function observe() {
    observer?.disconnect()
    if (typeof IntersectionObserver === 'undefined')
      return
    const elements = toValue(options.ids).map(id => document.getElementById(id)).filter((element): element is HTMLElement => element !== null)
    const visible = new Set<Element>()
    observer = new IntersectionObserver((entries) => {
      entries.forEach(entry => entry.isIntersecting ? visible.add(entry.target) : visible.delete(entry.target))
      const first = elements.find(element => visible.has(element))
      if (first && !pinned)
        activeId.value = first.id
    }, { root: toValue(options.root) ?? null, rootMargin: `-${toValue(options.offset) ?? 0}px 0px 0px 0px` })
    elements.forEach(element => observer?.observe(element))
  }
  function selectSection(id: string) {
    activeId.value = id
    pinned = true
    settle()
  }
  onMounted(() => {
    observe()
    trackHash()
    scrollTarget = toValue(options.root) ?? window
    scrollTarget.addEventListener('scroll', settle, { passive: true })
    window.addEventListener('hashchange', trackHash)
    watch(() => [toValue(options.ids).join('\0'), toValue(options.offset), toValue(options.root)], observe)
  })
  onBeforeUnmount(() => {
    observer?.disconnect()
    clearTimeout(timer)
    scrollTarget?.removeEventListener('scroll', settle)
    window.removeEventListener('hashchange', trackHash)
  })
  return { activeId, selectSection }
}
