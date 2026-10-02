import { watch } from 'vue'
import { i18n, locale, t } from './lib/i18n'
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { PiniaColada, useQueryCache } from '@pinia/colada'
import App from './App.vue'
import { router } from './router'
import 'virtual:uno.css'
import './style.css'
watch(
  locale,
  (value) => {
    document.documentElement.lang = value
  },
  { immediate: true },
)
watch(
  [() => router.currentRoute.value.path, locale],
  () => {
    if (router.currentRoute.value.path.startsWith('/app'))
      document.title = `Octopulse · ${t('app.serviceMonitoring')}`
  },
  { immediate: true },
)
const app = createApp(App)
  .use(i18n)
  .use(createPinia())
  .use(PiniaColada, {
    mutationOptions: {
      onSuccess() {
        // Imperative form reads use staleTime: 0; let in-flight reads finish
        // while subscribed queries refresh after a successful mutation.
        void cache
          .invalidateQueries({ predicate: (entry) => entry.active || !entry.pending })
          .catch(() => {})
      },
    },
  })
// Initialize in the app's injection context before guards and event handlers use it.
const cache = app.runWithContext(() => useQueryCache())
app.use(router).mount('#app')
