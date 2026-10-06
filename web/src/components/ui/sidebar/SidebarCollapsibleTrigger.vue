<script lang="ts">
import type { PropType, SlotsType, VNode } from 'vue'
import { cloneVNode, Comment, defineComponent, Fragment, isVNode, mergeProps, Text } from 'vue'
import { useSidebarCollapsible } from './collapsible-context'

function findTrigger(children: VNode[]): VNode | undefined {
  for (const child of children) {
    if (!isVNode(child) || child.type === Comment || child.type === Text)
      continue
    if (child.type === Fragment && Array.isArray(child.children)) {
      const trigger = findTrigger(child.children as VNode[])
      if (trigger)
        return trigger
      continue
    }
    return child
  }
}

export default defineComponent({
  name: 'SidebarCollapsibleTrigger',
  inheritAttrs: false,
  props: {
    render: Object as PropType<VNode>,
  },
  slots: { } as SlotsType<{ default: () => VNode[] }>,
  setup(props, { attrs, slots }) {
    const { contentId, isOpen, toggle } = useSidebarCollapsible()

    return () => {
      const trigger = findTrigger(slots.default?.() ?? []) ?? props.render
      if (!trigger)
        return null

      return cloneVNode(trigger, mergeProps(attrs, {
        'aria-expanded': isOpen.value,
        'aria-controls': contentId,
        'data-open': isOpen.value || undefined,
        'onClick': toggle,
      }), true)
    }
  },
})
</script>
