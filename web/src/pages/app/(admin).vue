<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  Activity,
  LayoutDashboard,
  Globe,
  Server,
  Settings,
  Users,
  KeyRound,
  ScrollText,
  Bell,
  CalendarClock,
  MessageSquare,
  Menu,
  Sun,
  Moon,
  Languages,
  LogOut,
  ChevronRight,
} from '@lucide/vue'
import { dark, theme, formatDate, statusLabel } from '../../composables/preferences'
import { currentUser, applySession, isAdmin } from '../../composables/api'
import { useMutation } from '@pinia/colada'
import { deleteSessionMutation } from '../../client/@pinia/colada.gen'
import { useIntervalFn, useMediaQuery, onKeyStroke } from '@vueuse/core'
import Brand from '../../components/Brand.vue'
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
  { path: '/app', icon: LayoutDashboard, label: 'navigation.overview' },
  { path: '/app/monitors', icon: Activity, label: 'common.monitors' },
  { path: '/app/pages', icon: Globe, label: 'navigation.statusPages' },
  { path: '/app/incidents', icon: MessageSquare, label: 'navigation.incidents' },
  { path: '/app/maintenance', icon: CalendarClock, label: 'navigation.maintenance' },
  { path: '/app/notifications', icon: Bell, label: 'navigation.notifications' },
  { path: '/app/servers', icon: Server, label: 'navigation.servers' },
]
const adminNav = [
  { path: '/app/users', icon: Users, label: 'navigation.members' },
  { path: '/app/secrets', icon: KeyRound, label: 'navigation.secrets' },
  { path: '/app/audit', icon: ScrollText, label: 'navigation.auditLog' },
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
      <RouterLink to="/app" @click="menuOpen = false"><Brand /></RouterLink>
      <div class="nav-label">{{ t('common.workspace') }}</div>
      <nav>
        <RouterLink
          v-for="item in nav"
          :key="item.path"
          :to="item.path"
          class="nav-link"
          :class="{ 'router-link-active': item.path === '/app' && route.path === '/app' }"
          :exact-active-class="item.path === '/app' ? 'router-link-active' : ''"
          :active-class="item.path === '/app' ? '' : 'router-link-active'"
          @click="menuOpen = false"
          ><component :is="item.icon" :size="17" />{{ t(item.label) }}</RouterLink
        >
      </nav>
      <div class="nav-label">{{ t('navigation.administration') }}</div>
      <nav>
        <template v-if="isAdmin()"
          ><RouterLink
            v-for="item in adminNav"
            :key="item.path"
            :to="item.path"
            class="nav-link"
            @click="menuOpen = false"
            ><component :is="item.icon" :size="17" />{{ t(item.label) }}</RouterLink
          ></template
        ><RouterLink to="/app/settings" class="nav-link" @click="menuOpen = false"
          ><Settings :size="17" />{{ t('common.settings') }}</RouterLink
        >
      </nav>
      <div class="sidebar-footer">
        <div un-flex="~ items-center gap-3">
          <span class="user-avatar">{{
            currentUser?.name?.[0] || currentUser?.username?.[0]
          }}</span>
          <div un-flex="1">
            <strong un-text="xs">{{ currentUser?.name || currentUser?.username }}</strong>
            <div class="muted" un-text="10px">{{ statusLabel(currentUser?.role || '') }}</div>
          </div>
          <button class="icon-button" :aria-label="t('navigation.signOut')" @click="signout">
            <LogOut :size="16" />
          </button>
        </div>
      </div>
    </aside>
    <div class="workspace">
      <header class="topbar">
        <div un-flex="~ items-center gap-3">
          <button
            class="icon-button mobile-menu"
            @click="menuOpen = !menuOpen"
            :aria-label="
              menuOpen ? t('navigation.closeNavigation') : t('navigation.openNavigation')
            "
            :aria-expanded="menuOpen"
            aria-controls="workspace-navigation"
          >
            <Menu :size="20" /></button
          ><span class="topbar-crumb"
            >{{ t('common.workspace') }}<ChevronRight un-inline="" un-mx="2" :size="11" /><strong>{{
              title
            }}</strong></span
          >
        </div>
        <div un-flex="~ items-center gap-3">
          <span class="topbar-time"><i />{{ formatDate(now.getTime()) }}</span
          ><button
            class="icon-button"
            :aria-label="t('common.switchLanguage')"
            @click="locale = locale === 'zh-CN' ? 'en' : 'zh-CN'"
          >
            <Languages :size="17" /></button
          ><button
            class="icon-button"
            :aria-label="t('navigation.toggleColorScheme')"
            @click="theme = dark ? 'light' : 'dark'"
          >
            <Sun v-if="dark" :size="17" /><Moon v-else :size="17" />
          </button>
        </div>
      </header>
      <main class="content"><RouterView /></main>
    </div>
  </div>
</template>
