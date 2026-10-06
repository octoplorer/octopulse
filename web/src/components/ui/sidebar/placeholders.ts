import { defineComponent } from 'vue'

// Replace these placeholders when the corresponding Vue components are ready.
export const Button = defineComponent({
  name: 'SidebarButtonPlaceholder',
  inheritAttrs: false,
  setup: () => () => null,
})

export const Tooltip = defineComponent({
  name: 'SidebarTooltipPlaceholder',
  inheritAttrs: false,
  setup: () => () => null,
})

export const TooltipProvider = defineComponent({
  name: 'SidebarTooltipProviderPlaceholder',
  inheritAttrs: false,
  setup: (_, { slots }) => () => slots.default?.(),
})
