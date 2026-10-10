<script lang="ts">
import type { ComponentPublicInstance, PropType } from 'vue'
import type { RouteLocationRaw } from 'vue-router'
import { defineComponent, h, ref, watchEffect } from 'vue'
import { RouterLink } from 'vue-router'
import { Tooltip } from '../tooltip'
import { useSidebar } from './context'
import { useMenuItemContext, wrapMenuText } from './menu-context'

export default defineComponent({
  name: 'SidebarMenuButton',
  inheritAttrs: false,
  props: {
    icon: String,
    active: { type: Boolean, default: false },
    href: String,
    to: { type: [String, Object] as PropType<RouteLocationRaw> },
    target: String,
    tooltip: String,
    itemId: String,
  },
  setup(props, { attrs, slots, expose }) {
    const { state, peekable, registerItem } = useSidebar()
    const isInsideMenuItem = useMenuItemContext()
    const itemElement = ref<HTMLLIElement | null>(null)
    const element = ref<HTMLElement | null>(null)

    function setInteractiveRef(node: Element | ComponentPublicInstance | null) {
      element.value = (node && '$el' in node ? node.$el : node) as HTMLElement | null
    }

    watchEffect((onCleanup) => {
      if (!props.itemId)
        return
      const id = props.itemId
      registerItem(id, isInsideMenuItem ? element.value : itemElement.value)
      onCleanup(() => registerItem(id, null))
    })

    expose({ element, itemElement })

    return () => {
      const destination = props.to ?? props.href
      const isLink = destination !== undefined
      const iconClass = ['shrink-0 opacity-40', 'size-4']
      const icon = slots.icon
        ? slots.icon({ class: iconClass })
        : props.icon
          ? h('span', { 'class': [...iconClass, props.icon], 'aria-hidden': true })
          : null
      const content = h('div', {
        class: [
          'flex min-w-0 flex-1 items-center gap-3',
          'translate-x-[-3px] group-not-data-[state=collapsed]/sidebar:translate-x-0',
          'transition-transform duration-$sidebar-animation-duration',
        ],
      }, [
        icon,
        h('span', {
          class: 'flex min-w-0 flex-1 items-center gap-2 overflow-hidden text-left',
        }, wrapMenuText(slots.default?.())),
      ])
      const buttonClasses = [
        'group/menu-button relative flex w-full min-w-0 cursor-pointer items-center gap-2.5 rounded-lg outline-none',
        'before:absolute before:inset-x-0 before:-inset-y-px',
        'min-h-8.5 px-3 py-0 text-sm font-medium',
        'text-default transition-[color,box-shadow,outline] duration-$sidebar-animation-duration',
        props.active ? 'bg-$sidebar-active-bg' : 'hover:bg-$sidebar-active-bg',
        'has-[[data-active]]:bg-transparent has-[[data-active]]:hover:bg-$sidebar-active-bg',
        'focus:outline-none focus-visible:bg-$sidebar-active-bg focus-visible:text-strong',
        isLink && 'no-underline!',
        attrs.class,
      ]
      const commonProps = {
        ...attrs,
        'ref': setInteractiveRef,
        'class': buttonClasses,
        'data-active': props.active || undefined,
        'data-sidebar': 'menu-button',
        'data-sidebar-item-id': isInsideMenuItem ? props.itemId : undefined,
        'data-component': 'Sidebar',
        'data-part': isLink ? 'menu-button-link' : 'menu-button',
      }
      const interactive = isLink
        ? h(RouterLink, { ...commonProps, to: destination!, target: props.target }, { default: () => content })
        : h('button', { type: 'button', ...commonProps }, content)
      const button = props.tooltip
        ? h(Tooltip, {
            content: props.tooltip,
            disabled: state.value !== 'collapsed' || peekable.value,
            placement: 'right',
          }, { default: () => interactive })
        : interactive

      return isInsideMenuItem
        ? button
        : h('li', {
            'ref': itemElement,
            'data-sidebar': 'menu-item',
            'data-sidebar-item-id': props.itemId,
            'class': 'relative group-data-[state=collapsed]/sidebar:overflow-hidden',
          }, button)
    }
  },
})
</script>
