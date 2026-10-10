import type { ComputedRef, InjectionKey } from 'vue'
import { inject } from 'vue'

export const toolbarSizeKey: InjectionKey<ComputedRef<'sm' | 'base' | 'lg'>> = Symbol('toolbar-size')
export const useToolbarSize = () => inject(toolbarSizeKey)
