<script setup lang="ts">
import { useMutation } from '@pinia/colada'
import { onKeyStroke, useIntervalFn, useMediaQuery } from '@vueuse/core'
import { computed, nextTick, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { deleteSessionMutation } from '../../client/@pinia/colada.gen'
import Brand from '../../components/Brand.vue'
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
  <div
    class="admin-app"
    :data-theme="dark ? 'dark' : 'light'"
    :data-sidebar-state="state"
    :data-resizing="isResizing"
    :data-mobile-open="mobileOpen"
    :style="{ '--sidebar-rail-width': `${railWidth}px`, '--sidebar-panel-width': `${panelWidth}px` }"
  >
    <div v-if="mobile && mobileOpen" class="sidebar-scrim" aria-hidden="true" @click="closeMobile" />
    <aside
      id="workspace-navigation"
      ref="sidebarElement"
      class="sidebar"
      :class="{ open: mobileOpen }"
      :inert="mobile && !mobileOpen"
      :role="mobile ? 'dialog' : undefined"
      :aria-modal="mobile && mobileOpen ? true : undefined"
      :aria-hidden="mobile && !mobileOpen ? true : undefined"
      :aria-label="t('common.workspace')"
    >
      <div class="sidebar-panel">
        <div
          class="sidebar-peek-zone"
          @pointerenter="($event.pointerType === 'mouse') && setHovered(true)"
          @pointerleave="onPeekPointerLeave"
          @focusin="setFocused(true)"
          @focusout="onPeekFocusOut"
        >
          <header class="sidebar-header">
            <RouterLink to="/app" aria-label="Octopulse" class="sidebar-brand" @click="closeMobile">
              <Brand />
            </RouterLink>
            <button ref="mobileClose" class="icon-button mobile-menu" :aria-label="t('navigation.closeNavigation')" @click="closeMobile">
              <span class="i-lucide-x" un-w="18px" un-h="18px" aria-hidden="true" />
            </button>
          </header>
          <div class="sidebar-content" @click="closeMobile">
            <div class="nav-label">
              {{ t('common.workspace') }}
            </div>
            <nav :aria-label="t('common.workspace')">
              <RouterLink
                v-for="item in nav"
                :key="item.path"
                :to="item.path"
                class="nav-link"
                :aria-label="t(item.label)"
                :class="{ 'router-link-active': item.path === '/app' && route.path === '/app' }"
                :exact-active-class="item.path === '/app' ? 'router-link-active' : ''"
                :active-class="item.path === '/app' ? '' : 'router-link-active'"
              >
                <span :class="item.icon" class="nav-icon" aria-hidden="true" />
                <span class="nav-link-label">{{ t(item.label) }}</span>
              </RouterLink>
            </nav>
            <div class="nav-label">
              {{ t('navigation.administration') }}
            </div>
            <nav :aria-label="t('navigation.administration')">
              <template v-if="isAdmin()">
                <RouterLink
                  v-for="item in adminNav"
                  :key="item.path"
                  :to="item.path"
                  class="nav-link"
                  :aria-label="t(item.label)"
                >
                  <span :class="item.icon" class="nav-icon" aria-hidden="true" />
                  <span class="nav-link-label">{{ t(item.label) }}</span>
                </RouterLink>
              </template>
              <RouterLink to="/app/settings" class="nav-link" :aria-label="t('common.settings')">
                <span class="i-lucide-settings nav-icon" aria-hidden="true" />
                <span class="nav-link-label">{{ t('common.settings') }}</span>
              </RouterLink>
            </nav>
          </div>
        </div>
        <footer class="sidebar-footer">
          <button
            ref="desktopTrigger"
            class="sidebar-toggle"
            :aria-label="toggleLabel"
            :aria-expanded="open"
            aria-controls="workspace-navigation"
            @click="toggle"
          >
            <span :class="open ? 'i-lucide-panel-left-close' : 'i-lucide-panel-left-open'" class="nav-icon" aria-hidden="true" />
            <span class="nav-link-label">{{ toggleLabel }}</span>
          </button>
        </footer>
      </div>
      <div
        v-if="!mobile"
        role="separator"
        tabindex="0"
        class="sidebar-resize-handle"
        aria-orientation="vertical"
        :aria-label="t('navigation.resizeSidebar')"
        :aria-valuemin="SIDEBAR_COLLAPSED_WIDTH"
        :aria-valuemax="SIDEBAR_MAX_WIDTH"
        :aria-valuenow="railWidth"
        :aria-valuetext="t('navigation.sidebarWidth', { width: railWidth })"
        aria-controls="workspace-navigation"
        aria-describedby="sidebar-resize-hint"
        @pointerdown="onResizeStart"
        @pointerleave="onPeekPointerLeave"
        @pointermove="resizeTo($event.clientX)"
        @pointerup="onResizeEnd"
        @pointercancel="onResizeEnd"
        @lostpointercapture="endResize"
        @keydown="onResizeKey"
      />
      <span v-if="!mobile" id="sidebar-resize-hint" class="visually-hidden">{{ t('navigation.resizeSidebarHint') }}</span>
    </aside>
    <div class="workspace" :inert="mobile && mobileOpen">
      <header class="topbar">
        <div un-flex="~ items-center gap-3">
          <button
            ref="mobileTrigger"
            class="icon-button mobile-menu"
            :aria-label="t('navigation.openNavigation')"
            :aria-expanded="mobileOpen"
            aria-controls="workspace-navigation"
            @click="toggle"
          >
            <span class="i-lucide-menu" un-w="20px" un-h="20px" aria-hidden="true" />
          </button><span class="topbar-crumb"><span class="crumb-workspace">{{ t('common.workspace') }}</span><span class="i-lucide-chevron-right" un-mx="2" un-w="11px" un-h="11px" aria-hidden="true" /><strong>{{
            title
          }}</strong></span>
        </div>
        <div class="topbar-controls">
          <span class="topbar-time"><i />{{ formatDate(now.getTime()) }}</span><button
            class="icon-button"
            :aria-label="t('common.switchLanguage')"
            @click="locale = locale === 'zh-CN' ? 'en' : 'zh-CN'"
          >
            <span class="i-lucide-languages" un-w="17px" un-h="17px" aria-hidden="true" />
          </button><button
            class="icon-button"
            :aria-label="t('navigation.toggleColorScheme')"
            @click="theme = dark ? 'light' : 'dark'"
          >
            <span v-if="dark" class="i-lucide-sun" un-w="17px" un-h="17px" aria-hidden="true" /><span v-else class="i-lucide-moon" un-w="17px" un-h="17px" aria-hidden="true" />
          </button>
          <div class="topbar-account">
            <RouterLink to="/app/settings" class="account-profile" :aria-label="`${userName} · ${statusLabel(currentUser?.role || '')}`" :title="userName">
              <span class="user-avatar" aria-hidden="true">{{ userName[0] }}</span>
              <span class="account-details">
                <strong>{{ userName }}</strong>
                <span class="muted">{{ statusLabel(currentUser?.role || '') }}</span>
              </span>
            </RouterLink>
            <button class="icon-button" :aria-label="t('navigation.signOut')" :title="t('navigation.signOut')" @click="signout">
              <span class="i-lucide-log-out" un-w="16px" un-h="16px" aria-hidden="true" />
            </button>
          </div>
        </div>
      </header>
      <main class="content">
        <RouterView />
      </main>
    </div>
  </div>
</template>
