<script setup lang="ts">
import { useLocalStorage } from '@vueuse/core'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { Button } from '@/components/ui/button'
import {
  Sidebar,
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
import { dark, theme } from '@/composables/preferences'

const sidebarOpen = useLocalStorage('octopulse.sidebar.open', true)
const { t } = useI18n()
const route = useRoute()

const title = computed(() => {
  const value = route.meta.title
  return typeof value === 'string' && value ? t(value) : 'Octopulse'
})
</script>

<template>
  <SidebarProvider v-model:open="sidebarOpen" peekable class="h-svh">
    <Sidebar>
      <SidebarHeader>
        Logo
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
          <SidebarMenu>
            <SidebarMenuButton icon="i-lucide-layout-dashboard" to="/app">
              {{ t('navigation.overview') }}
            </SidebarMenuButton>
            <SidebarMenuButton icon="i-lucide-activity" to="/app/monitors">
              {{ t('common.monitors') }}
            </SidebarMenuButton>
            <SidebarMenuButton icon="i-lucide-globe" to="/app/pages">
              {{ t('navigation.statusPages') }}
            </SidebarMenuButton>
            <SidebarMenuButton icon="i-lucide-message-square" to="/app/incidents">
              {{ t('navigation.incidents') }}
            </SidebarMenuButton>
            <SidebarMenuButton icon="i-lucide-calendar-clock" to="/app/maintenance">
              {{ t('navigation.maintenance') }}
            </SidebarMenuButton>
            <SidebarMenuButton icon="i-lucide-bell" to="/app/notifications">
              {{ t('navigation.notifications') }}
            </SidebarMenuButton>
            <SidebarMenuButton icon="i-lucide-server" to="/app/servers">
              {{ t('navigation.servers') }}
            </SidebarMenuButton>
          </SidebarMenu>
        </SidebarGroup>
        <SidebarGroup>
          <SidebarGroupLabel>管理</SidebarGroupLabel>
          <SidebarMenu>
            <SidebarMenuButton icon="i-lucide-users" to="/app/users">
              {{ t('navigation.members') }}
            </SidebarMenuButton>
            <SidebarMenuButton icon="i-lucide-key-round" to="/app/secrets">
              {{ t('navigation.secrets') }}
            </SidebarMenuButton>
            <SidebarMenuButton icon="i-lucide-users" to="/app/audit">
              {{ t('navigation.auditLog') }}
            </SidebarMenuButton>
          </SidebarMenu>
        </SidebarGroup>
      </SidebarContent>
      <SidebarFooter>
        <SidebarTrigger />
      </SidebarFooter>
    </Sidebar>
    <div class="min-h-0 min-w-0 flex-1 overflow-auto">
      <header
        flex="~ justify-between items-center gap-12px"
        min-h="58px" px="32px"
        border="b-1 solid line" bg="canvas"
        class="[@media(max-width:1200px)]:px-25px [@media(max-width:700px)]:h-62px [@media(max-width:380px)]:[&>div]:gap-4px [@media(max-width:700px)]:px-17px [@media(max-width:380px)]:px-12px [@container_workspace_(max-width:_700px)]:px-16px!"
      >
        <div flex="~ items-center gap-3" min-w="0">
          <Button shape="square" as-child>
            <SidebarTrigger
              aria-controls="workspace-navigation"
              class="hidden [@media(max-width:700px)]:inline-flex"
            >
              <span w="20px" h="20px" aria-hidden="true" class="i-lucide-menu" />
            </SidebarTrigger>
          </Button>
          <span
            flex="~ items-center" min-w="0"
            un-text="13px subtle"
            whitespace="nowrap" class="[@media(max-width:700px)]:text-12px"
          >
            <span
              class="[@media(max-width:700px)]:hidden [@container_workspace_(max-width:_700px)]:hidden"
            >
              {{ t('common.workspace') }}
            </span>
            <span
              mx="2"
              w="11px" h="11px" aria-hidden="true"
              class="i-lucide-chevron-right [@media(max-width:700px)]:hidden [@container_workspace_(max-width:_700px)]:hidden"
            />
            <strong un-text="default ellipsis" font="500" overflow="hidden">
              {{ title }}
            </strong>
          </span>
        </div>
        <div
          flex="~ items-center shrink-0 gap-6px" min-w="0"
        >
          <Button
            :aria-label="t('navigation.toggleColorScheme')" shape="square"
            @click="theme = dark ? 'light' : 'dark'"
          >
            <span v-if="dark" w="17px" h="17px" aria-hidden="true" class="i-lucide-sun" />
            <span v-else w="17px" h="17px" aria-hidden="true" class="i-lucide-moon" />
          </Button>
        </div>
      </header>
      <main
        p="t-32px x-32px b-48px"
        max-w="1600px"
        min-w="0" m="x-auto"
        class="[@media(max-width:1200px)]:p-25px [@media(max-width:700px)]:px-16px [@media(max-width:700px)]:pt-22px [@media(max-width:700px)]:pb-35px [@container_workspace_(max-width:_700px)]:px-16px! [@container_workspace_(max-width:_700px)]:pt-24px! [@container_workspace_(max-width:_700px)]:pb-35px!"
      >
        <RouterView />
      </main>
    </div>
  </SidebarProvider>
</template>
