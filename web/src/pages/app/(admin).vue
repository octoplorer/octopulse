<script setup lang="ts">
import { useMutation } from '@pinia/colada'
import { onKeyStroke, useIntervalFn, useMediaQuery } from '@vueuse/core'
import { computed, nextTick, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { deleteSessionMutation } from '../../client/@pinia/colada.gen'
import Brand from '../../components/Brand.vue'
import { Avatar } from '../../components/ui/avatar'
import { Button } from '../../components/ui/button'
import { SidebarIcon, SidebarLabel, SidebarLink, SidebarLinkLabel } from '../../components/ui/sidebar'
import { applySession, currentUser, isAdmin } from '../../composables/api'
import { dark, formatDate, statusLabel, theme } from '../../composables/preferences'
import { SIDEBAR_COLLAPSED_WIDTH, SIDEBAR_MAX_WIDTH, useSidebar } from '../../composables/sidebar'

const { t, locale } = useI18n({ useScope: 'global' })
const deleteSession = useMutation({
  ...deleteSessionMutation(),
  onSuccess: () => applySession(null),
})

const route = useRoute()
const router = useRouter()
const sidebarElement = ref<HTMLElement>()
const mobileTrigger = ref<HTMLButtonElement>()
const mobileClose = ref<HTMLButtonElement>()
const desktopTrigger = ref<HTMLButtonElement>()
const now = ref(new Date())
useIntervalFn(() => {
  now.value = new Date()
}, 60000)
const mobile = useMediaQuery('(max-width: 700px)')
const {
  open,
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
} = useSidebar(mobile)
const userName = computed(() => currentUser.value?.name || currentUser.value?.username || '')
const toggleLabel = computed(() => t(open.value ? 'navigation.collapseSidebar' : 'navigation.expandSidebar'))

onKeyStroke('Escape', (event) => {
  if (mobileOpen.value || state.value === 'peeking') {
    event.preventDefault()
    closeMobile()
    dismissPeek()
  }
})
onKeyStroke('Tab', (event) => {
  if (!mobile.value || !mobileOpen.value)
    return
  const elements = sidebarElement.value?.querySelectorAll<HTMLElement>('a[href], button:not([disabled])')
  const focusable = Array.from(elements || []).filter(element => element.getClientRects().length)
  const first = focusable[0]
  const last = focusable.at(-1)
  if (event.shiftKey && document.activeElement === first) {
    event.preventDefault()
    last?.focus()
  }
  else if (!event.shiftKey && document.activeElement === last) {
    event.preventDefault()
    first?.focus()
  }
})
watch(mobileOpen, async (value, previous) => {
  await nextTick()
  if (value)
    mobileClose.value?.focus()
  else if (previous)
    (mobile.value ? mobileTrigger.value : desktopTrigger.value)?.focus()
})
watch(() => route.fullPath, () => {
  closeMobile()
  dismissPeek()
})

function onPeekFocusOut(event: FocusEvent) {
  if (!(event.relatedTarget instanceof Node) || !(event.currentTarget as HTMLElement).contains(event.relatedTarget))
    setFocused(false)
}
function onPeekPointerLeave(event: PointerEvent) {
  const target = event.relatedTarget
  if (target instanceof Element && sidebarElement.value?.contains(target)
    && target.closest('.sidebar-peek-zone, .sidebar-resize-handle')) {
    return
  }
  setHovered(false)
}
function onResizeStart(event: PointerEvent) {
  if (event.button !== 0 || mobile.value)
    return
  event.preventDefault()
  ;(event.currentTarget as HTMLElement).setPointerCapture(event.pointerId)
  beginResize(event.clientX)
}
function onResizeEnd(event: PointerEvent) {
  endResize()
  const target = event.currentTarget as HTMLElement
  if (target.hasPointerCapture(event.pointerId))
    target.releasePointerCapture(event.pointerId)
}
function onResizeKey(event: KeyboardEvent) {
  if (resizeWithKey(event.key))
    event.preventDefault()
}
const nav = [
  { path: '/app', icon: 'i-lucide-layout-dashboard', label: 'navigation.overview' },
  { path: '/app/monitors', icon: 'i-lucide-activity', label: 'common.monitors' },
  { path: '/app/pages', icon: 'i-lucide-globe', label: 'navigation.statusPages' },
  { path: '/app/incidents', icon: 'i-lucide-message-square', label: 'navigation.incidents' },
  { path: '/app/maintenance', icon: 'i-lucide-calendar-clock', label: 'navigation.maintenance' },
  { path: '/app/notifications', icon: 'i-lucide-bell', label: 'navigation.notifications' },
  { path: '/app/servers', icon: 'i-lucide-server', label: 'navigation.servers' },
]
const adminNav = [
  { path: '/app/users', icon: 'i-lucide-users', label: 'navigation.members' },
  { path: '/app/secrets', icon: 'i-lucide-key-round', label: 'navigation.secrets' },
  { path: '/app/audit', icon: 'i-lucide-scroll-text', label: 'navigation.auditLog' },
]
const title = computed(() => {
  const value = route.meta.title
  return typeof value === 'string' && value ? t(value) : 'Octopulse'
})
async function signout() {
  await deleteSession.mutateAsync({})
  router.push('/app/login')
}
</script>

<template>
  <div :data-theme="dark ? 'dark' : 'light'" :data-sidebar-state="state" :data-resizing="isResizing" :data-mobile-open="mobileOpen" :style="{ '--sidebar-rail-width': `${railWidth}px`, '--sidebar-panel-width': `${panelWidth}px` }" class="admin-app">
    <div v-if="mobile && mobileOpen" aria-hidden="true" class="fixed inset-0 z-29 bg-[#0004]" @click="closeMobile" />
    <aside id="workspace-navigation" ref="sidebarElement" position="fixed left-0 inset-y-0" z="30" :class="{ open: mobileOpen }" :inert="mobile && !mobileOpen" :role="mobile ? 'dialog' : undefined" :aria-modal="mobile && mobileOpen ? true : undefined" :aria-hidden="mobile && !mobileOpen ? true : undefined" :aria-label="t('common.workspace')" class="sidebar">
      <div h="full" flex="~ col" overflow="hidden" bg="$surface" border="r-1 solid $border" class="sidebar-panel">
        <div flex="~ col 1" min-h="0" class="sidebar-peek-zone" @pointerenter="($event.pointerType === 'mouse') && setHovered(true)" @pointerleave="onPeekPointerLeave" @focusin="setFocused(true)" @focusout="onPeekFocusOut">
          <header flex="~ items-center justify-between shrink-0" h="60px" px="12px" border="b-1 solid $border" class="sidebar-header">
            <RouterLink to="/app" aria-label="Octopulse" min-w="0" whitespace="nowrap" class="sidebar-brand [&_.brand]:text-18px" @click="closeMobile">
              <Brand />
            </RouterLink>
            <Button size="icon" as-child>
              <button ref="mobileClose" :aria-label="t('navigation.closeNavigation')" class="mobile-menu" @click="closeMobile">
                <span w="18px" h="18px" aria-hidden="true" class="i-lucide-x" />
              </button>
            </Button>
          </header>
          <div flex="1" min-h="0" p="12px" overflow="x-hidden y-auto" class="sidebar-content [scrollbar-width:thin]" @click="closeMobile">
            <SidebarLabel>
              {{ t('common.workspace') }}
            </SidebarLabel>
            <nav :aria-label="t('common.workspace')">
              <SidebarLink v-for="item in nav" :key="item.path" as-child>
                <RouterLink :to="item.path" :aria-label="t(item.label)" :class="{ 'router-link-active': item.path === '/app' && route.path === '/app' }" :exact-active-class="item.path === '/app' ? 'router-link-active' : ''" :active-class="item.path === '/app' ? '' : 'router-link-active'">
                  <SidebarIcon :class="item.icon" aria-hidden="true" />
                  <SidebarLinkLabel>{{ t(item.label) }}</SidebarLinkLabel>
                </RouterLink>
              </SidebarLink>
            </nav>
            <SidebarLabel mt="24px">
              {{ t('navigation.administration') }}
            </SidebarLabel>
            <nav :aria-label="t('navigation.administration')">
              <template v-if="isAdmin()">
                <SidebarLink v-for="item in adminNav" :key="item.path" as-child>
                  <RouterLink :to="item.path" :aria-label="t(item.label)">
                    <SidebarIcon :class="item.icon" aria-hidden="true" />
                    <SidebarLinkLabel>{{ t(item.label) }}</SidebarLinkLabel>
                  </RouterLink>
                </SidebarLink>
              </template>
              <SidebarLink as-child>
                <RouterLink to="/app/settings" :aria-label="t('common.settings')">
                  <SidebarIcon aria-hidden="true" class="i-lucide-settings" />
                  <SidebarLinkLabel>{{ t('common.settings') }}</SidebarLinkLabel>
                </RouterLink>
              </SidebarLink>
            </nav>
          </div>
        </div>
        <footer flex="shrink-0" h="49px" p="y-8px x-12px" border="t-1 solid $border" class="sidebar-footer">
          <button ref="desktopTrigger" flex="~ items-center gap-10px" h="32px" w="full" p="7px" border="0 rounded-8px" bg="transparent hover:$surface-soft" un-text="12px $muted hover:$text" whitespace="nowrap" :aria-label="toggleLabel" :aria-expanded="open" aria-controls="workspace-navigation" class="sidebar-toggle" @click="toggle">
            <SidebarIcon :class="open ? 'i-lucide-panel-left-close' : 'i-lucide-panel-left-open'" aria-hidden="true" />
            <SidebarLinkLabel>{{ toggleLabel }}</SidebarLinkLabel>
          </button>
        </footer>
      </div>
      <div v-if="!mobile" role="separator" tabindex="0" position="absolute inset-y-0 right--5px" w="11px" z="1" cursor="col-resize" touch="none" outline="none" aria-orientation="vertical" :aria-label="t('navigation.resizeSidebar')" :aria-valuemin="SIDEBAR_COLLAPSED_WIDTH" :aria-valuemax="SIDEBAR_MAX_WIDTH" :aria-valuenow="railWidth" :aria-valuetext="t('navigation.sidebarWidth', { width: railWidth })" aria-controls="workspace-navigation" aria-describedby="sidebar-resize-hint" class="sidebar-resize-handle" @pointerdown="onResizeStart" @pointerleave="onPeekPointerLeave" @pointermove="resizeTo($event.clientX)" @pointerup="onResizeEnd" @pointercancel="onResizeEnd" @lostpointercapture="endResize" @keydown="onResizeKey" />
      <span v-if="!mobile" id="sidebar-resize-hint" class="sr-only">{{ t('navigation.resizeSidebarHint') }}</span>
    </aside>
    <div min-w="0" min-h="screen" :inert="mobile && mobileOpen" class="workspace">
      <header flex="~ justify-between items-center gap-12px" min-h="60px" px="32px" border="b-1 solid $border" bg="$bg" class="topbar">
        <div flex="~ items-center gap-3" min-w="0">
          <Button size="icon" as-child>
            <button ref="mobileTrigger" :aria-label="t('navigation.openNavigation')" :aria-expanded="mobileOpen" aria-controls="workspace-navigation" class="mobile-menu" @click="toggle">
              <span w="20px" h="20px" aria-hidden="true" class="i-lucide-menu" />
            </button>
          </Button><span flex="~ items-center" min-w="0" un-text="13px $muted" whitespace="nowrap" class="topbar-crumb"><span class="crumb-workspace">{{ t('common.workspace') }}</span><span mx="2" w="11px" h="11px" aria-hidden="true" class="i-lucide-chevron-right" /><strong un-text="$text ellipsis" font="500" overflow="hidden">{{
            title
          }}</strong></span>
        </div>
        <div flex="~ items-center shrink-0 gap-6px" min-w="0" class="topbar-controls">
          <span flex="~ items-center gap-7px" un-text="12px $muted" mr="12px" whitespace="nowrap" class="topbar-time tabular-nums"><i w="6px" h="6px" rounded="full" bg="$success" />{{ formatDate(now.getTime()) }}</span><Button :aria-label="t('common.switchLanguage')" size="icon" @click="locale = locale === 'zh-CN' ? 'en' : 'zh-CN'">
            <span w="17px" h="17px" aria-hidden="true" class="i-lucide-languages" />
          </Button><Button :aria-label="t('navigation.toggleColorScheme')" size="icon" @click="theme = dark ? 'light' : 'dark'">
            <span v-if="dark" w="17px" h="17px" aria-hidden="true" class="i-lucide-sun" /><span v-else w="17px" h="17px" aria-hidden="true" class="i-lucide-moon" />
          </Button>
          <div flex="~ items-center gap-8px" border="l-1 solid $border" pl="12px" ml="6px" class="topbar-account">
            <RouterLink to="/app/settings" flex="~ items-center gap-8px" rounded="8px" p="4px" bg="hover:$surface-soft" :aria-label="`${userName} · ${statusLabel(currentUser?.role || '')}`" :title="userName" class="account-profile">
              <Avatar aria-hidden="true">
                {{ userName[0] }}
              </Avatar>
              <span flex="~ col" un-text="11px" leading="[1.4]" min-w="0" class="account-details">
                <strong un-text="12px ellipsis" font="500" max-w="120px" overflow="hidden" whitespace="nowrap">{{ userName }}</strong>
                <span class="muted" un-text="13px $muted">{{ statusLabel(currentUser?.role || '') }}</span>
              </span>
            </RouterLink>
            <Button :aria-label="t('navigation.signOut')" :title="t('navigation.signOut')" size="icon" @click="signout">
              <span w="16px" h="16px" aria-hidden="true" class="i-lucide-log-out" />
            </Button>
          </div>
        </div>
      </header>
      <main p="t-32px x-32px b-48px" max-w="1600px" min-w="0" m="x-auto" class="content">
        <RouterView />
      </main>
    </div>
  </div>
</template>
