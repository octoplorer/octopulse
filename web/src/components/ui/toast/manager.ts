import type { CreateToasterProps, CreateToasterReturn } from '@ark-ui/vue/toast'
import type { InjectionKey } from 'vue'
import { createToaster } from '@ark-ui/vue/toast'
import { inject } from 'vue'

export function createKumoToastManager(options: Partial<CreateToasterProps> = {}) {
  return createToaster({ placement: 'bottom-end', overlap: false, gap: 12, duration: 7000, pauseOnPageIdle: true, ...options })
}
export const toastManager = createKumoToastManager()
export const toastManagerKey: InjectionKey<CreateToasterReturn> = Symbol('kumo-toast-manager')
export function useKumoToastManager() {
  return inject(toastManagerKey, toastManager)
}
