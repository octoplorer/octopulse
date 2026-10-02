import { PiniaColada, useQueryCache } from '@pinia/colada'
import { createHead } from '@unhead/vue/client'
import { createPinia } from 'pinia'
import { createApp } from 'vue'
import App from './App.vue'
import { i18n } from './composables/i18n'
import { router } from './router'
import 'virtual:uno.css'
import './style.css'

const app = createApp(App)
  .use(createHead())
  .use(i18n)
  .use(createPinia())
  .use(PiniaColada, {
    mutationOptions: {
      onSuccess: invalidateActiveQueries,
    },
  })
// Initialize in the app's injection context before guards and event handlers use it.
const cache = app.runWithContext(() => useQueryCache())

function invalidateActiveQueries() {
  // Imperative form reads use staleTime: 0; let in-flight reads finish
  // while subscribed queries refresh after a successful mutation.
  void cache
    .invalidateQueries({ predicate: entry => entry.active || !entry.pending })
    .catch(() => {})
}

app.use(router).mount('#app')
