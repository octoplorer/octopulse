import type { InjectionKey } from 'vue'
import { inject } from 'vue'

export interface TooltipDefaults { openDelay?: number, closeDelay?: number }
export const tooltipDefaultsKey: InjectionKey<TooltipDefaults> = Symbol('tooltip-defaults')
export const useTooltipDefaults = () => inject(tooltipDefaultsKey, {})
