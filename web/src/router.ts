import { createRouter, createWebHistory } from 'vue-router'
import { routes, handleHotUpdate } from 'vue-router/auto-routes'
import { currentUser, loadSession } from './lib/api'
export const router = createRouter({
  history: createWebHistory(),
  routes,
})
if (import.meta.hot) handleHotUpdate(router)

let sessionLoaded = false
router.beforeEach(async (to) => {
  if (!to.path.startsWith('/app') || to.path === '/app/login') return
  if (!sessionLoaded) {
    try {
      await loadSession()
    } catch {}
    sessionLoaded = true
  }
  if (!currentUser.value) return { path: '/app/login', query: { next: to.fullPath } }
  const roles = to.meta.roles
  if (roles && !roles.includes(currentUser.value.role)) return { path: '/app' }
})
