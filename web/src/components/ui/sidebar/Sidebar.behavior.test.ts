// @vitest-environment happy-dom
import { mount } from '@vue/test-utils'
import { expect, it } from 'vitest'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { SidebarMenu, SidebarMenuButton, SidebarProvider, SidebarTrigger } from './index'

it('keeps tooltip menu links interactive when the sidebar collapses', async () => {
  const router = createRouter({ history: createMemoryHistory(), routes: [{ path: '/', component: { template: '<div />' } }, { path: '/monitors', component: { template: '<div />' } }] })
  const wrapper = mount(defineComponent({
    setup: () => () => h(SidebarProvider, { defaultOpen: true }, { default: () => [
      h(SidebarMenu, {}, { default: () => h(SidebarMenuButton, { to: '/monitors', tooltip: 'Monitors' }, { default: () => 'Monitors' }) }),
      h(SidebarTrigger),
    ] }),
  }), { global: { plugins: [router] }, attachTo: document.body })
  try {
    expect(wrapper.get('a').text()).toBe('Monitors')
    await wrapper.get('button').trigger('click')
    expect(wrapper.get('button').attributes('aria-expanded')).toBe('false')
    await wrapper.get('a').trigger('click')
    await router.isReady()
    expect(router.currentRoute.value.path).toBe('/monitors')
  }
  finally {
    wrapper.unmount()
  }
})
