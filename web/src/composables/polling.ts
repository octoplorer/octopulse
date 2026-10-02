import { createSharedComposable, useDocumentVisibility, useOnline } from '@vueuse/core'
import { computed } from 'vue'

export const usePollingEnabled = createSharedComposable(() => {
  const visibility = useDocumentVisibility()
  const online = useOnline()
  return computed(() => visibility.value === 'visible' && online.value)
})
