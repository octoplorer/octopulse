import { watch } from 'vue'
import { locale, t } from './lib/preferences'
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
      document.title = `Octopulse · ${t('服务监控', 'Service monitoring')}`
  },
  { immediate: true },
)
createApp(App).use(createPinia()).use(PiniaColada).use(router).mount('#app')
