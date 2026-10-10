import type { ComputedRef, InjectionKey } from 'vue'

export interface CheckboxAppearance {
  appearance: 'default' | 'card'
  controlFirst?: boolean
}
export const checkboxAppearanceKey: InjectionKey<ComputedRef<CheckboxAppearance>> = Symbol('checkbox-appearance')
