import type { InjectionKey } from 'vue'
import type { SidebarContextValue } from './types'
import { inject, provide } from 'vue'

const sidebarContextKey: InjectionKey<SidebarContextValue> = Symbol('sidebar-context')

export function provideSidebarContext(context: SidebarContextValue) {
  provide(sidebarContextKey, context)
}

export function useSidebar() {
  const context = inject(sidebarContextKey)
  if (!context)
    throw new Error('useSidebar must be used within SidebarProvider')
  return context
}
