import { toastManager } from '../components/ui/toast'

export function notify(message: string, kind: 'success' | 'error' = 'success') {
  return toastManager.create({ title: message, type: kind, duration: 7000 })
}
