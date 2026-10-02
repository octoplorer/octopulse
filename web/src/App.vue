<template>
  <RouterView />
  <div class="toast-stack" aria-live="polite">
    <div v-for="item in notices" :key="item.id" class="toast" :class="item.kind">
      <span>{{ item.message }}</span
      ><button @click="dismissNotice(item.id)" :aria-label="t('关闭通知', 'Close notification')">
        ×
      </button>
    </div>
  </div>
</template>
<script setup lang="ts">
import { watch } from 'vue'
import { t } from './lib/preferences'
import { useQueryCache } from '@pinia/colada'
import { notices, dismissNotice } from './lib/notices'
import { currentUser, observeMutations } from './lib/api'
const cache = useQueryCache()
observeMutations(() => {
  void cache.invalidateQueries().catch(() => {})
})
watch(
  () => currentUser.value?.id,
  (id, previous) => {
    if (previous && id !== previous) {
      cache.cancelQueries()
      for (const entry of cache.getEntries()) cache.remove(entry)
    }
  },
)
</script>
