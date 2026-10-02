import assert from 'node:assert/strict'
import { fileURLToPath } from 'node:url'
import test from 'node:test'
import { createServer } from 'vite'
import { createMemoryHistory, createRouter } from 'vue-router'

test('generated file routes preserve workspace and public URLs', async (t) => {
  const vite = await createServer({
    root: fileURLToPath(new URL('../', import.meta.url)),
    configFile: fileURLToPath(new URL('../vite.config.ts', import.meta.url)),
    optimizeDeps: { noDiscovery: true, include: [] },
    server: { middlewareMode: true, watch: null, hmr: false },
  })
  t.after(() => vite.close())
  const { routes } = await vite.ssrLoadModule('vue-router/auto-routes')
  const router = createRouter({ history: createMemoryHistory(), routes })
  const components = (route) => route.matched.filter((record) => record.components)
  const dashboard = router.resolve('/app')
  const [layout] = components(dashboard)
  assert.equal(components(dashboard).length, 2)

  await t.test('workspace pages retain their titles, roles, and shared layout', () => {
    const pages = [
      ['/app', ['概览', 'Overview']],
      ['/app/monitors', ['监控项', 'Monitors']],
      ['/app/monitors/new', ['创建监控项', 'New monitor'], ['admin', 'operator']],
      ['/app/monitors/monitor-id', ['监控详情', 'Monitor details']],
      ['/app/monitors/monitor-id/edit', ['编辑监控项', 'Edit monitor'], ['admin', 'operator']],
      ['/app/pages', ['状态页', 'Status pages']],
      ['/app/pages/new', ['创建状态页', 'New status page'], ['admin', 'operator']],
      ['/app/pages/page-id', ['自定义状态页', 'Customize status page']],
      ['/app/incidents', ['事件公告', 'Incidents']],
      ['/app/maintenance', ['计划维护', 'Maintenance']],
      ['/app/notifications', ['通知与投递', 'Notifications']],
      ['/app/servers', ['服务器', 'Servers']],
      ['/app/settings', ['设置', 'Settings']],
      ['/app/users', ['成员与权限', 'Members'], ['admin']],
      ['/app/secrets', ['秘密凭据', 'Secrets'], ['admin']],
      ['/app/audit', ['审计日志', 'Audit log'], ['admin']],
    ]
    for (const [path, title, roles] of pages) {
      const route = router.resolve(path)
      assert.deepEqual(route.meta.title, title, path)
      assert.deepEqual(route.meta.roles, roles, path)
      assert.equal(components(route).length, 2, path)
      assert.equal(components(route)[0], layout, path)
    }
  })

  await t.test('login stays outside the workspace and retains its return URL', () => {
    const next = '/app/monitors/monitor-id/edit?tab=history#check'
    const route = router.resolve(`/app/login?next=${encodeURIComponent(next)}`)
    assert.equal(components(route).length, 1)
    assert.ok(!route.matched.includes(layout))
    assert.equal(route.query.next, next)
    assert.equal(route.params.slug, undefined)
  })

  await t.test('static creation pages take precedence over dynamic IDs', () => {
    for (const section of ['monitors', 'pages']) {
      const create = router.resolve(`/app/${section}/new`)
      const detail = router.resolve(`/app/${section}/record-id`)
      assert.equal(create.params.id, undefined, section)
      assert.equal(detail.params.id, 'record-id', section)
      assert.notEqual(create.name, detail.name, section)
    }
    assert.equal(router.resolve('/app/monitors/monitor-id/edit').params.id, 'monitor-id')
  })

  await t.test('public pages retain optional slug and repeatable incident paths', () => {
    const publicRecord = components(router.resolve('/'))[0]
    const paths = [
      ['/', undefined, undefined],
      ['/status', 'status', undefined],
      ['/incidents/incident-id', 'incidents', ['incident-id']],
      ['/status/incidents/incident-id', 'status', ['incidents', 'incident-id']],
      ['/status/incidents/incident%20id', 'status', ['incidents', 'incident id']],
      ['/status/arbitrary/deep/path', 'status', ['arbitrary', 'deep', 'path']],
    ]
    for (const [path, slug, rest] of paths) {
      const route = router.resolve(path)
      assert.equal(components(route).length, 1, path)
      assert.equal(components(route)[0], publicRecord, path)
      assert.equal(route.params.slug, slug, path)
      assert.deepEqual(route.params.rest, rest, path)
      assert.ok(!route.matched.includes(layout), path)
    }
  })
})
