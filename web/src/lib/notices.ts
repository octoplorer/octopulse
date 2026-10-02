import { ref } from 'vue'
export const notices = ref<{ id: number; message: string; kind: 'success' | 'error' }[]>([])
let sequence = 0
export function notify(message: string, kind: 'success' | 'error' = 'success') {
  const id = ++sequence
  notices.value.push({ id, message, kind })
  setTimeout(() => dismissNotice(id), 7000)
}
export function dismissNotice(id: number) {
  notices.value = notices.value.filter((x) => x.id !== id)
}
export function errorText(error: unknown): string {
  if (error instanceof Error) return error.message
  if (typeof error === 'object' && error !== null && 'detail' in error) return String(error.detail)
  return String(error)
}
