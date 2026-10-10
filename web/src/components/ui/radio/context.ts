import type { ComputedRef, InjectionKey } from 'vue'

export interface RadioAppearance {
  appearance: 'default' | 'card' | 'segmented'
  controlPosition: 'start' | 'end'
}
export const radioAppearanceKey: InjectionKey<ComputedRef<RadioAppearance>> = Symbol('radio-appearance')
