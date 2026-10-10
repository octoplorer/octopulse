import type { InjectionKey, Ref } from 'vue'
import type { InputSize } from '../input/styles'
import { inject } from 'vue'

export interface InputGroupContext {
  size: Readonly<Ref<InputSize>>
  disabled: Readonly<Ref<boolean>>
  focusMode: Readonly<Ref<'container' | 'individual'>>
}

export const inputGroupKey: InjectionKey<InputGroupContext> = Symbol('InputGroup')
export const useInputGroup = () => inject(inputGroupKey, undefined)
