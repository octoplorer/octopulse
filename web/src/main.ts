import { watch } from 'vue'
import { i18n, locale, t } from './lib/i18n'
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { PiniaColada } from '@pinia/colada'
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
createApp(App).use(i18n).use(createPinia()).use(PiniaColada).use(router).mount('#app')
