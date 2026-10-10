import type { ComputedRef, InjectionKey } from 'vue'
import { inject } from 'vue'

export interface SwitchGroupContext { disabled: ComputedRef<boolean>, controlFirst: ComputedRef<boolean> }
export const switchGroupKey: InjectionKey<SwitchGroupContext> = Symbol('switch-group')
export const useSwitchGroup = () => inject(switchGroupKey, undefined)
