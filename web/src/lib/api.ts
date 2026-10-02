import { ref } from 'vue'
import { client } from '../client/client.gen'
import * as sdk from '../client/sdk.gen'
import { routes } from '../client/routes.gen'
import { normalizeCollections } from './normalize'
import { locale, timezone } from './preferences'
import type { User } from './types'
export const currentUser = ref<User | null>(null)
let csrfToken = ''
let mutationObserver: () => void = () => {}
export function observeMutations(observer: () => void) {
  mutationObserver = observer
}
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
  if (response.ok && !['GET', 'HEAD', 'OPTIONS'].includes(request.method)) mutationObserver()
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
  let message = problem.detail || problem.title || 'Request failed'
  if (problem.errors?.length)
    message += `: ${problem.errors.map((e) => `${e.location || ''} ${e.message || ''}`).join('; ')}`
  return new Error(message)
})
export function resolveOperation(path: string, method = 'GET') {
  const url = new URL(path.startsWith('/api/') ? path : `/api/v1/${path}`, location.origin)
  for (const route of routes) {
    if (route.method !== method.toUpperCase()) continue
    const match = new RegExp(route.pattern).exec(url.pathname)
    if (!match) continue
    const params: Record<string, string> = {}
    route.parameters.forEach((name, index) => {
      params[name] = decodeURIComponent(match[index + 1]!)
    })
    return { operation: route.operation, path: params, query: Object.fromEntries(url.searchParams) }
  }
  throw new Error(`API operation is missing from the generated contract: ${method} ${url.pathname}`)
}
type OperationOptions = {
  body?: unknown
  path?: Record<string, string>
  query?: Record<string, string>
  signal?: AbortSignal
  throwOnError: true
}
// This dynamic dispatch boundary is generated from Go route metadata. Every
// request uses the corresponding HeyAPI SDK operation; DTOs come from its types.
const operations = sdk as unknown as Record<
  string,
  (options: OperationOptions) => Promise<{ data: unknown }>
>
export async function api<T>(
  path: string,
  options: { method?: string; body?: unknown; signal?: AbortSignal } = {},
): Promise<T> {
  const resolved = resolveOperation(path, options.method || 'GET'),
    operation = operations[resolved.operation]
  if (!operation) throw new Error(`Generated SDK operation unavailable: ${resolved.operation}`)
  const { data } = await operation({
    ...resolved,
    body: options.body,
    signal: options.signal,
    throwOnError: true,
  })
  return data as T
}
export async function loadSession() {
  const { data } = await sdk.getSession({ throwOnError: true })
  currentUser.value = data.user || null
  csrfToken = data.csrfToken || ''
  if (data.user) {
    locale.value = data.user.locale
    timezone.value = data.user.timezone
  }
  return data
}
export async function login(username: string, password: string) {
  const { data } = await sdk.createSession({ body: { username, password }, throwOnError: true })
  currentUser.value = data.user || null
  csrfToken = data.csrfToken || ''
  return loadSession()
}
export async function logout() {
  await sdk.deleteSession({ throwOnError: true })
  csrfToken = ''
  currentUser.value = null
}
export const canEdit = () =>
  currentUser.value?.role === 'admin' || currentUser.value?.role === 'operator'
export const isAdmin = () => currentUser.value?.role === 'admin'
