import type { InjectionKey, VNode } from 'vue'
import { h, inject, provide, Text } from 'vue'

const menuItemContextKey: InjectionKey<boolean> = Symbol('sidebar-menu-item')
const menuSubItemContextKey: InjectionKey<boolean> = Symbol('sidebar-menu-sub-item')

export function provideMenuItemContext() {
  provide(menuItemContextKey, true)
}

export function useMenuItemContext() {
  return inject(menuItemContextKey, false)
}

export function provideMenuSubItemContext() {
  provide(menuSubItemContextKey, true)
}

export function useMenuSubItemContext() {
  return inject(menuSubItemContextKey, false)
}

export function wrapMenuText(children: VNode[] = []) {
  return children.map(child => child.type === Text
    ? h('span', { class: 'truncate' }, child.children as string)
    : child)
}
