import type { InjectionKey, Ref } from 'vue'
import { inject, provide, ref } from 'vue'

const sidebarSlidingViewActiveKey: InjectionKey<Readonly<Ref<string>>> = Symbol('sidebar-sliding-view-active')
const defaultActiveKey = ref('')

export function provideSidebarSlidingViewActive(activeKey: Readonly<Ref<string>>) {
  provide(sidebarSlidingViewActiveKey, activeKey)
}

export function useSidebarSlidingViewActive() {
  return inject(sidebarSlidingViewActiveKey, defaultActiveKey)
}
