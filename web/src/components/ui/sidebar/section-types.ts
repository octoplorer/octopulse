import type { VNode } from 'vue'

export interface SidebarCollapsibleProps {
  defaultOpen?: boolean
  open?: boolean
  autoScrollOnOpen?: boolean
}

export interface SidebarCollapsibleTriggerProps {
  /** Optional alternative to the default slot. */
  render?: VNode
}

export interface SidebarSlidingViewsProps {
  activeKey: string
  direction?: 'left' | 'right'
}

export interface SidebarSlidingViewProps {
  value: string
}
