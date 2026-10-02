import { useQueryCache } from '@pinia/colada'
import { createRouter, createWebHistory } from 'vue-router'
import { handleHotUpdate, routes } from 'vue-router/auto-routes'
import { getSessionQuery } from './client/@pinia/colada.gen'
import { applySession, currentUser } from './composables/api'

export const router = createRouter({
  history: createWebHistory(),
  routes,
})
if (import.meta.hot)
  handleHotUpdate(router)

let sessionLoaded = false
router.beforeEach(async (to) => {
  if (!to.path.startsWith('/app') || to.path === '/app/login')
    return
  if (!sessionLoaded) {
    try {
      const cache = useQueryCache()
      const state = await cache.refresh(cache.ensure({ ...getSessionQuery(), staleTime: 0 }))
      if (state.status === 'success')
        applySession(state.data)
    }
    catch {}
    sessionLoaded = true
  }
  if (!currentUser.value)
    return { path: '/app/login', query: { next: to.fullPath } }
  const roles = to.meta.roles
  if (roles && !roles.includes(currentUser.value.role))
    return { path: '/app' }
})
