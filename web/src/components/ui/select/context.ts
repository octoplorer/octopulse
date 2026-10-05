import type { ListCollection } from '@ark-ui/vue'
import type { ComputedRef, InjectionKey } from 'vue'
import type { SelectCollectionItem } from './types'
import { inject, provide } from 'vue'

interface SelectContext {
  collection: ComputedRef<ListCollection<SelectCollectionItem>>
  registerItem: (id: symbol, item: SelectCollectionItem) => void
  unregisterItem: (id: symbol) => void
}

const selectContextKey: InjectionKey<SelectContext> = Symbol('select-context')

export function provideSelectContext(context: SelectContext) {
  provide(selectContextKey, context)
}

export function useSelectContext() {
  const context = inject(selectContextKey)
  if (!context)
    throw new Error('Select components must be used inside Select')
  return context
}
