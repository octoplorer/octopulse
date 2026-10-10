<script setup lang="ts">
import { useQueryCache } from '@pinia/colada'
import { useHead } from '@unhead/vue'
import { watch } from 'vue'

import { useI18n } from 'vue-i18n'
import { Toasty } from './components/ui/toast'
import { currentUser } from './composables/api'
import { dark } from './composables/preferences'

const { t, locale } = useI18n({ useScope: 'global' })
useHead({
  title: () => `Octopulse · ${t('app.service-monitoring')}`,
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
  <Toasty :close-label="t('app.close-notification')" />
</template>
