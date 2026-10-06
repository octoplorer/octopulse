import type { InjectionKey, Ref } from 'vue'
import { inject, provide, ref } from 'vue'

interface SidebarCollapsibleContext {
  contentId: string
  isOpen: Readonly<Ref<boolean>>
  isCollapsible: boolean
  autoScrollOnOpen: Readonly<Ref<boolean>>
  toggle: () => void
  completeOpenChange: () => void
}

const sidebarCollapsibleContextKey: InjectionKey<SidebarCollapsibleContext> = Symbol('sidebar-collapsible')

const defaultContext: SidebarCollapsibleContext = {
  contentId: '',
  isOpen: ref(true),
  isCollapsible: false,
  autoScrollOnOpen: ref(false),
  toggle: () => {},
  completeOpenChange: () => {},
}

export function provideSidebarCollapsible(context: SidebarCollapsibleContext) {
  provide(sidebarCollapsibleContextKey, context)
}

export function useSidebarCollapsible() {
  return inject(sidebarCollapsibleContextKey, defaultContext)
}
