<script lang="ts">
import type { ComponentPublicInstance, PropType } from 'vue'
import type { RouteLocationRaw } from 'vue-router'
import { defineComponent, h, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { useMenuSubItemContext, wrapMenuText } from './menu-context'

export default defineComponent({
  name: 'SidebarMenuSubButton',
  inheritAttrs: false,
  props: {
    active: { type: Boolean, default: false },
    href: String,
    to: { type: [String, Object] as PropType<RouteLocationRaw> },
    target: String,
  },
  setup(props, { attrs, slots, expose }) {
    const isInsideMenuSubItem = useMenuSubItemContext()
    const element = ref<HTMLElement | null>(null)

    function setInteractiveRef(node: Element | ComponentPublicInstance | null) {
      element.value = (node && '$el' in node ? node.$el : node) as HTMLElement | null
    }

    expose({ element })

    return () => {
      const destination = props.to ?? props.href
      const isLink = destination !== undefined
      const content = h('span', {
        class: 'flex min-w-0 flex-1 items-center gap-2 overflow-hidden text-left',
      }, wrapMenuText(slots.default?.()))
      const buttonClasses = [
        'group/menu-button relative flex min-h-8.5 w-full min-w-0 cursor-pointer items-center gap-2 rounded-lg px-3 py-0 text-sm font-medium outline-none',
        'before:absolute before:inset-x-0 before:-inset-y-px',
        'text-default transition-[color] duration-150',
        props.active ? 'bg-$sidebar-active-bg' : 'hover:bg-$sidebar-active-bg',
        'focus:outline-none focus-visible:bg-$sidebar-active-bg focus-visible:text-strong',
        isLink && 'no-underline!',
        attrs.class,
      ]
      const commonProps = {
        ...attrs,
        ref: setInteractiveRef,
        class: buttonClasses,
        'data-active': props.active || undefined,
        'data-sidebar': 'menu-sub-button',
        'data-component': 'Sidebar',
        'data-part': isLink ? 'menu-sub-button-link' : 'menu-sub-button',
      }
      const button = isLink
        ? h(RouterLink, { ...commonProps, to: destination!, target: props.target }, { default: () => content })
        : h('button', { type: 'button', ...commonProps }, content)

      return isInsideMenuSubItem
        ? button
        : h('li', { 'data-sidebar': 'menu-sub-item', class: 'relative' }, button)
    }
  },
})
</script>
