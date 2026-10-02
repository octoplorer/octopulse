import { ref } from 'vue'

export const notices = ref<{ id: number, message: string, kind: 'success' | 'error' }[]>([])
let sequence = 0
export function notify(message: string, kind: 'success' | 'error' = 'success') {
  const id = ++sequence
  notices.value.push({ id, message, kind })
  setTimeout(dismissNotice, 7000, id)
}
export function dismissNotice(id: number) {
  notices.value = notices.value.filter(x => x.id !== id)
}
