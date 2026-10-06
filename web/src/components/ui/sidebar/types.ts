import type { ComputedRef, Ref } from 'vue'

export type SidebarSide = 'left' | 'right'
export type SidebarState = 'expanded' | 'collapsed' | 'peeking'
export type SidebarScrollAlign = 'start' | 'center' | 'end' | 'auto'

export interface SidebarScrollToItemOptions {
  align?: SidebarScrollAlign
  behavior?: 'auto' | 'smooth'
}

export interface SidebarProviderProps {
  defaultOpen?: boolean
  open?: boolean
  side?: SidebarSide
  resizable?: boolean
  defaultWidth?: number
  minWidth?: number
  maxWidth?: number
  contained?: boolean
  peekable?: boolean
  animationDuration?: number
  mobileBreakpoint?: number
}

export interface SidebarRootProps {
  contentClass?: string
  fullScreenOnMobile?: boolean
}

type SidebarRef<T> = Ref<T> | ComputedRef<T>

export interface SidebarContextValue {
  state: SidebarRef<SidebarState>
  open: SidebarRef<boolean>
  openMobile: SidebarRef<boolean>
  isMobile: SidebarRef<boolean>
  side: SidebarRef<SidebarSide>
  width: SidebarRef<number>
  resizable: SidebarRef<boolean>
  minWidth: SidebarRef<number>
  maxWidth: SidebarRef<number>
  isResizing: SidebarRef<boolean>
  isPeeking: SidebarRef<boolean>
  peekable: SidebarRef<boolean>
  contained: SidebarRef<boolean>
  animationDuration: SidebarRef<number>
  setOpen: (open: boolean) => void
  setOpenMobile: (open: boolean) => void
  toggleSidebar: () => void
  setIsResizing: (resizing: boolean) => void
  setWidth: (width: number) => void
  startPeek: () => void
  stopPeek: () => void
  registerItem: (id: string, node: HTMLElement | null) => void
  scrollToItem: (id: string, options?: SidebarScrollToItemOptions) => void
  scrollItemIntoView: (id: string, options?: SidebarScrollToItemOptions) => void
}
