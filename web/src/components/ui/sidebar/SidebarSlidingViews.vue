<script lang="ts">
import type { PropType, VNode } from 'vue'
import { defineComponent, Fragment, h, isVNode, mergeProps, toRef } from 'vue'
import SidebarSlidingView from './SidebarSlidingView.vue'
import { provideSidebarSlidingViewActive } from './sliding-view-context'

function flattenChildren(children: VNode[]): VNode[] {
  return children.flatMap((child) => {
    if (isVNode(child) && child.type === Fragment && Array.isArray(child.children))
      return flattenChildren(child.children as VNode[])
    return [child]
  })
}

export default defineComponent({
  name: 'SidebarSlidingViews',
  inheritAttrs: false,
  props: {
    activeKey: { type: String, required: true },
    direction: { type: String as PropType<'left' | 'right'>, default: 'left' },
  },
  setup(props, { attrs, slots, expose }) {
    let element: HTMLDivElement | null = null
    provideSidebarSlidingViewActive(toRef(props, 'activeKey'))
    expose({ get element() { return element } })

    return () => {
      const children = flattenChildren(slots.default?.() ?? [])
      const views = children.filter(child => isVNode(child) && child.type === SidebarSlidingView)
      const activeIndex = views.findIndex(child => child.props?.value === props.activeKey)
      const translateX = activeIndex > 0 ? `-${activeIndex * 100}%` : '0%'

      return h('div', mergeProps(attrs, {
        'ref': (node: unknown) => { element = node as HTMLDivElement | null },
        'data-sidebar': 'sliding-views',
        'class': 'flex min-h-0 max-w-$sidebar-width flex-1 overflow-hidden',
      }), [
        h('div', {
          class: 'flex w-full min-h-0 shrink-0 transition-transform duration-$sidebar-animation-duration ease-$sidebar-easing motion-reduce:transition-none',
          style: { transform: `translateX(${translateX})` },
        }, children),
      ])
    }
  },
})
</script>
