<script setup lang="ts">
import { useMutation } from '@pinia/colada'
import { useLocalStorage } from '@vueuse/core'
import { computed, nextTick, useTemplateRef, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { deleteSessionMutation } from '@/client/@pinia/colada.gen'
import Brand from '@/components/Brand.vue'
import { Breadcrumbs } from '@/components/ui/breadcrumbs'
import { Button } from '@/components/ui/button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import {
  Sidebar,
  SidebarClose,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarProvider,
  SidebarTrigger,
} from '@/components/ui/sidebar'
import { applySession, currentUser, isAdmin } from '@/composables/api'
import { notify } from '@/composables/notices'
import { dark, theme } from '@/composables/preferences'
import { errorText } from '@/lib/errors'

const sidebarOpen = useLocalStorage('octopulse.sidebar.open', true)
const { t, locale } = useI18n({ useScope: 'global' })
const route = useRoute()
const router = useRouter()
const sidebar = useTemplateRef<InstanceType<typeof SidebarProvider>>('sidebar')
const workspaceScroll = useTemplateRef<HTMLElement>('workspace-scroll')
const workspaceMain = useTemplateRef<HTMLElement>('workspace-main')
watch(() => route.fullPath, () => sidebar.value?.setOpenMobile(false))
function focusContent() {
  if (workspaceScroll.value)
    workspaceScroll.value.scrollTop = 0
  workspaceMain.value?.focus({ preventScroll: true })
}
watch(() => route.path, async () => {
  await nextTick()
  focusContent()
}, { flush: 'post' })
const title = computed(() => {
  const value = route.meta.title
  return typeof value === 'string' && value ? t(value) : 'Octopulse'
})
const contentWidth = computed(() => route.meta.contentWidth || 'wide')
const primaryGroups = [
  {
    label: 'navigation.monitoring',
    links: [
      { path: '/app', icon: 'i-lucide-layout-dashboard', label: 'navigation.overview' },
      { path: '/app/monitors', icon: 'i-lucide-activity', label: 'common.monitors' },
      { path: '/app/servers', icon: 'i-lucide-server', label: 'navigation.servers' },
    ],
  },
  {
    label: 'navigation.serviceOperations',
    links: [
      { path: '/app/pages', icon: 'i-lucide-globe', label: 'navigation.statusPages' },
      { path: '/app/incidents', icon: 'i-lucide-message-square', label: 'navigation.incidents' },
      { path: '/app/maintenance', icon: 'i-lucide-calendar-clock', label: 'navigation.maintenance' },
      { path: '/app/notifications', icon: 'i-lucide-bell', label: 'navigation.notifications' },
    ],
  },
]
const administrationLinks = [
  { path: '/app/users', icon: 'i-lucide-users', label: 'navigation.members' },
  { path: '/app/secrets', icon: 'i-lucide-key-round', label: 'navigation.secrets' },
  { path: '/app/audit', icon: 'i-lucide-scroll-text', label: 'navigation.auditLog' },
]
function active(path: string) {
  return path === '/app' ? route.path === path : route.path.startsWith(path)
}
const signOut = useMutation({
  ...deleteSessionMutation(),
  onSuccess: async () => {
    applySession(null)
    await router.replace('/app/login')
  },
  onError: error => notify(errorText(error), 'error'),
})
</script>

<template>
  <SidebarProvider ref="sidebar" :default-open="sidebarOpen" peekable class="workspace-shell h-svh overflow-hidden" @update:open="sidebarOpen = $event">
    <a href="#workspace-main" class="skip-link" @click.prevent="focusContent">{{ t('common.skipToContent') }}</a>
    <Sidebar id="workspace-navigation">
      <SidebarHeader class="flex-row items-center justify-between">
        <RouterLink to="/app" :aria-label="t('navigation.overview')" class="min-w-0">
          <Brand :compact="!sidebarOpen" />
        </RouterLink>
        <SidebarClose :aria-label="t('navigation.closeNavigation')" class="md:hidden" />
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup v-for="group in primaryGroups" :key="group.label">
          <SidebarGroupLabel>{{ t(group.label) }}</SidebarGroupLabel>
          <SidebarMenu>
            <SidebarMenuButton v-for="link in group.links" :key="link.path" :icon="link.icon" :to="link.path" :tooltip="t(link.label)" :active="active(link.path)">
              {{ t(link.label) }}
            </SidebarMenuButton>
          </SidebarMenu>
        </SidebarGroup>
        <SidebarGroup>
          <SidebarGroupLabel>{{ t('navigation.administration') }}</SidebarGroupLabel>
          <SidebarMenu>
            <template v-if="isAdmin()">
              <SidebarMenuButton v-for="link in administrationLinks" :key="link.path" :icon="link.icon" :to="link.path" :tooltip="t(link.label)" :active="active(link.path)">
                {{ t(link.label) }}
              </SidebarMenuButton>
            </template>
            <SidebarMenuButton icon="i-lucide-settings" to="/app/settings" :tooltip="t('common.settings')" :active="active('/app/settings')">
              {{ t('common.settings') }}
            </SidebarMenuButton>
          </SidebarMenu>
        </SidebarGroup>
      </SidebarContent>
      <SidebarFooter>
        <SidebarTrigger :aria-label="sidebarOpen ? t('navigation.collapseSidebar') : t('navigation.expandSidebar')" />
      </SidebarFooter>
    </Sidebar>
    <div ref="workspace-scroll" class="workspace-scroll min-h-0 min-w-0 flex-1 overflow-auto overscroll-contain">
      <header class="sticky top-0 z-30 flex h-$workspace-header-height shrink-0 items-center justify-between gap-3 border-b border-solid border-line bg-canvas px-4 md:px-8">
        <div class="flex min-w-0 items-center gap-3">
          <SidebarTrigger :aria-label="t('navigation.openNavigation')" aria-controls="workspace-navigation" class="md:hidden" />
          <Breadcrumbs class="workspace-breadcrumbs min-w-0" :items="[{ label: t('common.workspace'), href: '/app' }, { label: title }]" :label="t('common.workspace')" />
        </div>
        <div class="flex shrink-0 items-center gap-2">
          <DropdownMenu placement="bottom-end" @select="value => value === 'sign-out' && signOut.mutate({})">
            <DropdownMenuTrigger as-child>
              <Button variant="ghost" size="sm" :aria-label="currentUser?.name || currentUser?.username" :disabled="signOut.isLoading.value">
                <span class="i-lucide-user-round size-4 shrink-0" aria-hidden="true" />
                <span class="hidden max-w-24 truncate sm:block sm:max-w-32">{{ currentUser?.name || currentUser?.username }}</span>
                <span class="i-lucide-chevron-down size-3.5 shrink-0" aria-hidden="true" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent>
              <DropdownMenuItem value="settings" as-child>
                <RouterLink to="/app/settings">
                  <span class="i-lucide-settings size-4" aria-hidden="true" />
                  {{ t('settings.personalPreferences') }}
                </RouterLink>
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem value="sign-out" :disabled="signOut.isLoading.value" destructive>
                <span class="i-lucide-log-out size-4" aria-hidden="true" />
                {{ t('navigation.signOut') }}
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
          <Button variant="ghost" shape="square" :aria-label="t('common.switchLanguage')" @click="locale = locale === 'zh-CN' ? 'en' : 'zh-CN'">
            <span class="i-lucide-languages size-4" aria-hidden="true" />
          </Button>
          <Button variant="ghost" shape="square" :aria-label="t('navigation.toggleColorScheme')" @click="theme = dark ? 'light' : 'dark'">
            <span :class="dark ? 'i-lucide-sun' : 'i-lucide-moon'" class="size-4" aria-hidden="true" />
          </Button>
        </div>
      </header>
      <main id="workspace-main" ref="workspace-main" tabindex="-1" class="workspace-page mx-auto min-w-0 w-full" :data-content-width="contentWidth">
        <RouterView />
      </main>
    </div>
  </SidebarProvider>
</template>

<style scoped>
.skip-link {
  position: fixed;
  inset-inline-start: 1rem;
  inset-block-start: 1rem;
  z-index: 60;
  transform: translateY(calc(-100% - 2rem));
  padding: 0.75rem 1rem;
  border-radius: 0.5rem;
  background: var(--color-base);
  color: var(--text-color-default);
}

.skip-link:focus-visible {
  transform: none;
  outline: 2px solid var(--color-focus);
  outline-offset: 2px;
}

.workspace-shell {
  --workspace-header-height: 4rem;
  --sidebar-header-height: var(--workspace-header-height);
}

.workspace-scroll {
  container: workspace / inline-size;
  scroll-padding-block-start: calc(var(--workspace-header-height) + 1rem);
  scroll-padding-block-end: 6rem;
}

@media (max-width: 700px) {
  .workspace-breadcrumbs :deep(ol > li:nth-child(-n + 2)) {
    display: none;
  }
}

.workspace-page {
  max-width: 1440px;
  padding: 1.5rem 1rem;
}

.workspace-page[data-content-width='compact'] {
  max-width: 1120px;
}

.workspace-page[data-content-width='form'],
.workspace-page[data-content-width='settings'] {
  max-width: 960px;
}

.workspace-page[data-content-width='editor'] {
  max-width: 1520px;
}

@container workspace (min-width: 700px) {
  .workspace-page {
    padding: 2rem 1.5rem;
  }
}

@container workspace (min-width: 1200px) {
  .workspace-page {
    padding: 2rem;
  }
}
</style>
