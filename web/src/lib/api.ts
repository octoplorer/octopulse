import { ref } from 'vue'
import { client } from '../client/client.gen'
import type { GetSessionResponse } from '../client/types.gen'
import { normalizeCollections } from './normalize'
import { timezone } from './preferences'
import { locale, t } from './i18n'
import type { User } from './types'
export const currentUser = ref<User | null>(null)
let csrfToken = ''
client.setConfig({
  baseUrl: location.origin,
  credentials: 'same-origin',
  throwOnError: true,
  responseTransformer: async (data) => normalizeCollections(data),
})
client.interceptors.request.use((request) => {
  if (csrfToken) request.headers.set('X-CSRF-Token', csrfToken)
  return request
})
client.interceptors.response.use((response, request) => {
  if (response.status === 401 && !new URL(request.url).pathname.endsWith('/session')) {
    currentUser.value = null
    if (location.pathname.startsWith('/app') && location.pathname !== '/app/login')
      location.assign(`/app/login?next=${encodeURIComponent(location.pathname + location.search)}`)
  }
  return response
})
client.interceptors.error.use((error) => {
  if (error instanceof Error) return error
  if (typeof error !== 'object' || error === null) return new Error(String(error))
  const problem = error as {
    detail?: string
    title?: string
    errors?: { message?: string; location?: string }[]
  }
  let message = problem.detail || problem.title || t('errors.requestFailed')
  if (problem.errors?.length)
    message += `: ${problem.errors.map((e) => `${e.location || ''} ${e.message || ''}`).join('; ')}`
  return new Error(message)
})
export function applySession(data: GetSessionResponse | null) {
  csrfToken = data?.csrfToken || ''
  currentUser.value = data?.user || null
  if (data?.user) {
    locale.value = data.user.locale
    timezone.value = data.user.timezone
  }
}
export const canEdit = () =>
  currentUser.value?.role === 'admin' || currentUser.value?.role === 'operator'
export const isAdmin = () => currentUser.value?.role === 'admin'
