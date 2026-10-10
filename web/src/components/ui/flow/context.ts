import type { ComputedRef, InjectionKey, ShallowRef } from 'vue'
import type { FlowState, NodePositions, TreeNode } from './flow-layout'
import { inject, provide, shallowReactive } from 'vue'

export interface FlowContext {
  orderVersion: ShallowRef<number>
  orientation: ComputedRef<FlowState['orientation']>
  positions: ComputedRef<NodePositions>
  reportNode: (id: string, measurement: FlowState['nodes'][string]) => void
  removeNode: (id: string) => void
}
interface FlowEntry {
  tree: ComputedRef<TreeNode>
  element: () => HTMLElement | undefined
}
interface FlowGroupContext {
  entries: FlowEntry[]
  register: (entry: FlowEntry) => () => void
}
export interface FlowNodeContext {
  reportAnchor: (type: 'start' | 'end' | 'both', element: HTMLElement | undefined) => void
}
export const flowKey: InjectionKey<FlowContext> = Symbol('Flow')
export const flowGroupKey: InjectionKey<FlowGroupContext> = Symbol('FlowGroup')
export const flowNodeKey: InjectionKey<FlowNodeContext> = Symbol('FlowNode')

export function createFlowGroup(): FlowGroupContext {
  const entries = shallowReactive<FlowEntry[]>([])
  return {
    entries,
    register: (entry) => {
      entries.push(entry)
      return () => {
        const index = entries.indexOf(entry)
        if (index >= 0)
          entries.splice(index, 1)
      }
    },
  }
}

export function flowGroupChildren(group: FlowGroupContext): TreeNode[] {
  return [...group.entries].sort((a, b) => {
    const first = a.element()
    const second = b.element()
    if (!first || !second || first === second)
      return 0
    const relation = first.compareDocumentPosition(second)
    return relation & 4 ? -1 : relation & 2 ? 1 : 0
  }).map(entry => entry.tree.value)
}

export function useFlowContext() {
  const context = inject(flowKey)
  if (!context)
    throw new Error('Flow components must be rendered inside Flow.')
  return context
}

export function provideFlowGroup(group: FlowGroupContext) {
  provide(flowGroupKey, group)
}
