<template>
  <RouterView />
  <div class="toast-stack" aria-live="polite">
    <div v-for="item in notices" :key="item.id" class="toast" :class="item.kind">
      <span>{{ item.message }}</span
      ><button @click="dismissNotice(item.id)" :aria-label="t('app.closeNotification')">×</button>
    </div>
  </div>
</template>
<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { watch } from 'vue'

import { useQueryCache } from '@pinia/colada'
import { notices, dismissNotice } from './composables/notices'
import { currentUser } from './composables/api'
const { t } = useI18n({ useScope: 'global' })

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
