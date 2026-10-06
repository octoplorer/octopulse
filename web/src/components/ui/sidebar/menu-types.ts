import type { RouteLocationRaw } from 'vue-router'

export interface SidebarLoadingProps {
  label?: string
}

export interface SidebarMenuItemProps {
  /** Anchor id for useSidebar().scrollToItem(id). */
  itemId?: string
}

export interface SidebarMenuButtonProps extends SidebarMenuItemProps {
  /** UnoCSS icon preset class, for example i-lucide-house. */
  icon?: string
  active?: boolean
  href?: string
  to?: RouteLocationRaw
  target?: string
  tooltip?: string
}

export interface SidebarMenuSubButtonProps {
  active?: boolean
  href?: string
  to?: RouteLocationRaw
  target?: string
}
