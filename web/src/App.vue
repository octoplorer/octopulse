<script setup lang="ts">
import { useQueryCache } from '@pinia/colada'
import { useHead } from '@unhead/vue'
import { watch } from 'vue'

import { useI18n } from 'vue-i18n'
import { Toast, ToastClose, ToastViewport } from './components/ui/toast'
import { currentUser } from './composables/api'
import { dismissNotice, notices } from './composables/notices'
import { dark } from './composables/preferences'

const { t, locale } = useI18n({ useScope: 'global' })
useHead({
  title: () => `Octopulse · ${t('app.serviceMonitoring')}`,
  htmlAttrs: {
    'lang': locale,
    'data-mode': () => dark.value ? 'dark' : 'light',
  },
  meta: [
    { name: 'color-scheme', content: 'light dark' },
    { name: 'theme-color', content: () => dark.value ? '#0a0a0a' : '#fafafa' },
  ],
})

const cache = useQueryCache()
watch(
  () => currentUser.value?.id,
  (id, previous) => {
    if (id !== previous) {
      cache.cancelQueries()
      for (const entry of cache.getEntries()) cache.remove(entry)
    }
  },
  { flush: 'sync' },
)
</script>

<template>
  <RouterView />
  <ToastViewport>
    <Toast v-for="item in notices" :key="item.id" :variant="item.kind">
      <span>{{ item.message }}</span><ToastClose :aria-label="t('app.closeNotification')" @click="dismissNotice(item.id)">
        ×
      </ToastClose>
    </Toast>
  </ToastViewport>
</template>
