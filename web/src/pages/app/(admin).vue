<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  Activity,
  LayoutDashboard,
  Globe,
  Radio,
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
import { dark, locale, theme, t, formatDate, statusLabel } from '../../lib/preferences'
import { currentUser, logout, isAdmin } from '../../lib/api'
import { useIntervalFn, useMediaQuery, onKeyStroke } from '@vueuse/core'
import Brand from '../../components/Brand.vue'
const route = useRoute(),
  router = useRouter(),
  menuOpen = ref(false),
  now = ref(new Date())
useIntervalFn(() => {
  now.value = new Date()
}, 60000)
const mobile = useMediaQuery('(max-width: 700px)')
onKeyStroke('Escape', () => {
  menuOpen.value = false
})
const nav = [
  { path: '/app', icon: LayoutDashboard, zh: '概览', en: 'Overview' },
  { path: '/app/monitors', icon: Activity, zh: '监控项', en: 'Monitors' },
  { path: '/app/pages', icon: Globe, zh: '状态页', en: 'Status pages' },
  { path: '/app/incidents', icon: MessageSquare, zh: '事件公告', en: 'Incidents' },
  { path: '/app/maintenance', icon: CalendarClock, zh: '计划维护', en: 'Maintenance' },
  { path: '/app/notifications', icon: Bell, zh: '通知与投递', en: 'Notifications' },
  { path: '/app/servers', icon: Server, zh: '服务器', en: 'Servers' },
]
const adminNav = [
  { path: '/app/users', icon: Users, zh: '成员与权限', en: 'Members' },
  { path: '/app/secrets', icon: KeyRound, zh: '秘密凭据', en: 'Secrets' },
  { path: '/app/audit', icon: ScrollText, zh: '审计日志', en: 'Audit log' },
]
const title = computed(() => {
  const value = route.meta.title
  return value ? t(...value) : 'Octopulse'
})
async function signout() {
  await logout()
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
      <div class="nav-label">{{ t('工作空间', 'Workspace') }}</div>
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
          ><component :is="item.icon" :size="17" />{{ t(item.zh, item.en) }}</RouterLink
        >
      </nav>
      <div class="nav-label">{{ t('管理', 'Administration') }}</div>
      <nav>
        <template v-if="isAdmin()"
          ><RouterLink
            v-for="item in adminNav"
            :key="item.path"
            :to="item.path"
            class="nav-link"
            @click="menuOpen = false"
            ><component :is="item.icon" :size="17" />{{ t(item.zh, item.en) }}</RouterLink
          ></template
        ><RouterLink to="/app/settings" class="nav-link" @click="menuOpen = false"
          ><Settings :size="17" />{{ t('设置', 'Settings') }}</RouterLink
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
          <button class="icon-button" :aria-label="t('退出登录', 'Sign out')" @click="signout">
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
              menuOpen ? t('关闭导航', 'Close navigation') : t('打开导航', 'Open navigation')
            "
            :aria-expanded="menuOpen"
            aria-controls="workspace-navigation"
          >
            <Menu :size="20" /></button
          ><span class="topbar-crumb"
            >{{ t('工作空间', 'Workspace')
            }}<ChevronRight un-inline="" un-mx="2" :size="11" /><strong>{{ title }}</strong></span
          >
        </div>
        <div un-flex="~ items-center gap-3">
          <span class="topbar-time"><i />{{ formatDate(now.getTime()) }}</span
          ><button
            class="icon-button"
            :aria-label="t('切换语言', 'Switch language')"
            @click="locale = locale === 'zh-CN' ? 'en' : 'zh-CN'"
          >
            <Languages :size="17" /></button
          ><button
            class="icon-button"
            :aria-label="t('切换深浅色', 'Toggle color scheme')"
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
