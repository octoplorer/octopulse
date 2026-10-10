// @vitest-environment happy-dom
import { PiniaColada } from '@pinia/colada'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { afterEach, expect, it } from 'vitest'
import { h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import Workspace from '../(admin).vue'
import { currentUser } from '../../../composables/api'
import { i18n } from '../../../composables/i18n'

const wrappers: ReturnType<typeof mount>[] = []
afterEach(() => {
  wrappers.splice(0).forEach(wrapper => wrapper.unmount())
  document.body.innerHTML = ''
})

async function render() {
  currentUser.value = { id: 'admin', username: 'admin', name: 'Administrator', role: 'admin', enabled: true, locale: 'zh-CN', timezone: 'UTC', createdAt: 0, updatedAt: 0 }
  i18n.global.locale.value = 'zh-CN'
  const router = createRouter({ history: createMemoryHistory(), routes: [
    { path: '/app', component: { render: () => h('h1', 'Overview') } },
    { path: '/app/monitors', component: { render: () => h('h1', 'Monitors') } },
    { path: '/:pathMatch(.*)*', component: { render: () => h('div') } },
  ] })
  await router.push('/app')
  await router.isReady()
  const wrapper = mount(Workspace, { attachTo: document.body, global: { plugins: [router, createPinia(), [PiniaColada, { queryOptions: { refetchOnWindowFocus: false } }], i18n] } })
  wrappers.push(wrapper)
  await flushPromises()
  return { wrapper, router }
}

it('lets keyboard users skip repeated navigation and focus the main content', async () => {
  const { wrapper } = await render()
  const skip = wrapper.find('a[href="#workspace-main"]')
  expect(skip.exists()).toBe(true)
  await skip.trigger('click')
  expect(document.activeElement).toBe(wrapper.get('main').element)
})

it('starts a newly navigated view at its main content', async () => {
  const { wrapper, router } = await render()
  const scroller = wrapper.get('.workspace-scroll').element
  scroller.scrollTop = 500
  ;(wrapper.get('a[href="/app/monitors"]').element as HTMLAnchorElement).focus()
  await router.push('/app/monitors')
  await flushPromises()
  expect(document.activeElement).toBe(wrapper.get('main').element)
  expect(scroller.scrollTop).toBe(0)
  expect(wrapper.get('main').text()).toContain('Monitors')
})
