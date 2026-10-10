<script setup lang="ts">
import type { TreeNode } from './flow-layout'
import { computed, inject, onBeforeUnmount, useTemplateRef } from 'vue'
import { createFlowGroup, flowGroupChildren, flowGroupKey, provideFlowGroup, useFlowContext } from './context'

const props = defineProps<{ kind: 'parallel' | 'list', align?: 'end' }>()
const parent = inject(flowGroupKey)!
const group = createFlowGroup()
const context = useFlowContext()
const element = useTemplateRef<HTMLElement>('element')
const unregister = parent.register({ tree: computed<TreeNode>(() => {
  void context.orderVersion.value
  return { kind: props.kind, align: props.align, children: flowGroupChildren(group) }
}), element: () => element.value ?? undefined })
provideFlowGroup(group)
onBeforeUnmount(unregister)
</script>

<template>
  <div ref="element" class="contents" role="presentation">
    <slot />
  </div>
</template>
