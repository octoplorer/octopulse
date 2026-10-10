import type { ComputedRef, InjectionKey } from 'vue'
import { inject } from 'vue'

export interface LayerDialogContext { dismissDisabled: ComputedRef<boolean>, close: () => void }
export const layerDialogKey: InjectionKey<LayerDialogContext> = Symbol('layer-dialog')
export const useLayerDialog = () => inject(layerDialogKey)
