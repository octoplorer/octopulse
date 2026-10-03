<script setup lang="ts">
import { useMutation } from '@pinia/colada'
import { onKeyStroke, useIntervalFn, useMediaQuery } from '@vueuse/core'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { deleteSessionMutation } from '../../client/@pinia/colada.gen'
import Brand from '../../components/Brand.vue'
import { applySession, currentUser, isAdmin } from '../../composables/api'
import { dark, formatDate, statusLabel, theme } from '../../composables/preferences'

const { t, locale } = useI18n({ useScope: 'global' })
const deleteSession = useMutation({
  ...deleteSessionMutation(),
  onSuccess: () => applySession(null),
})

const route = useRoute()
const router = useRouter()
const menuOpen = ref(false)
const now = ref(new Date())
useIntervalFn(() => {
  now.value = new Date()
}, 60000)
const mobile = useMediaQuery('(max-width: 700px)')
onKeyStroke('Escape', () => {
  menuOpen.value = false
})
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
  return value ? t(value) : 'Octopulse'
})
async function signout() {
  await deleteSession.mutateAsync({})
  router.push('/app/login')
}
</script>

<template>
  <div class="admin-app" :data-theme="dark ? 'dark' : 'light'">
    <div v-if="menuOpen" class="sidebar-scrim" @click="menuOpen = false" />
    <aside
      id="workspace-navigation"
      class="sidebar"
      :class="{ open: menuOpen }"
      :inert="mobile && !menuOpen"
    >
      <RouterLink to="/app" @click="menuOpen = false">
        <Brand />
      </RouterLink>
      <div class="nav-label">
        {{ t('common.workspace') }}
      </div>
      <nav :aria-label="t('common.workspace')">
        <RouterLink
          v-for="item in nav"
          :key="item.path"
          :to="item.path"
          class="nav-link"
          :class="{ 'router-link-active': item.path === '/app' && route.path === '/app' }"
          :exact-active-class="item.path === '/app' ? 'router-link-active' : ''"
          :active-class="item.path === '/app' ? '' : 'router-link-active'"
          @click="menuOpen = false"
        >
          <span :class="item.icon" un-w="17px" un-h="17px" aria-hidden="true" />{{ t(item.label) }}
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
            @click="menuOpen = false"
          >
            <span :class="item.icon" un-w="17px" un-h="17px" aria-hidden="true" />{{ t(item.label) }}
          </RouterLink>
        </template><RouterLink to="/app/settings" class="nav-link" @click="menuOpen = false">
          <span class="i-lucide-settings" un-w="17px" un-h="17px" aria-hidden="true" />{{ t('common.settings') }}
        </RouterLink>
      </nav>
      <div class="sidebar-footer">
        <div un-flex="~ items-center gap-3">
          <span class="user-avatar">{{
            currentUser?.name?.[0] || currentUser?.username?.[0]
          }}</span>
          <div un-flex="1">
            <strong un-text="xs">{{ currentUser?.name || currentUser?.username }}</strong>
            <div class="muted" un-text="xs">
              {{ statusLabel(currentUser?.role || '') }}
            </div>
          </div>
          <button class="icon-button" :aria-label="t('navigation.signOut')" @click="signout">
            <span class="i-lucide-log-out" un-w="16px" un-h="16px" aria-hidden="true" />
          </button>
        </div>
      </div>
    </aside>
    <div class="workspace">
      <header class="topbar">
        <div un-flex="~ items-center gap-3">
          <button
            class="icon-button mobile-menu"
            :aria-label="
              menuOpen ? t('navigation.closeNavigation') : t('navigation.openNavigation')
            "
            :aria-expanded="menuOpen"
            aria-controls="workspace-navigation"
            @click="menuOpen = !menuOpen"
          >
            <span class="i-lucide-menu" un-w="20px" un-h="20px" aria-hidden="true" />
          </button><span class="topbar-crumb">{{ t('common.workspace') }}<span class="i-lucide-chevron-right" un-mx="2" un-w="11px" un-h="11px" aria-hidden="true" /><strong>{{
            title
          }}</strong></span>
        </div>
        <div un-flex="~ items-center gap-3">
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
        </div>
      </header>
      <main class="content">
        <RouterView />
      </main>
    </div>
  </div>
</template>
