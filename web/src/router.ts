import { createRouter, createWebHistory } from 'vue-router'
import { currentUser, loadSession } from './lib/api'
export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/app/login', component: () => import('./views/AuthView.vue') },
    {
      path: '/app',
      component: () => import('./components/AppShell.vue'),
      children: [
        {
          path: '',
          component: () => import('./views/DashboardView.vue'),
          meta: { title: ['概览', 'Overview'] },
        },
        {
          path: 'monitors',
          component: () => import('./views/MonitorsView.vue'),
          meta: { title: ['监控项', 'Monitors'] },
        },
        {
          path: 'monitors/new',
          component: () => import('./views/MonitorEditor.vue'),
          meta: { title: ['创建监控项', 'New monitor'], roles: ['admin', 'operator'] },
        },
        {
          path: 'monitors/:id',
          component: () => import('./views/MonitorDetail.vue'),
          meta: { title: ['监控详情', 'Monitor details'] },
        },
        {
          path: 'monitors/:id/edit',
          component: () => import('./views/MonitorEditor.vue'),
          meta: { title: ['编辑监控项', 'Edit monitor'], roles: ['admin', 'operator'] },
        },
        {
          path: 'pages',
          component: () => import('./views/PagesView.vue'),
          meta: { title: ['状态页', 'Status pages'] },
        },
        {
          path: 'pages/new',
          component: () => import('./views/PageEditor.vue'),
          meta: { title: ['创建状态页', 'New status page'], roles: ['admin', 'operator'] },
        },
        {
          path: 'pages/:id',
          component: () => import('./views/PageEditor.vue'),
          meta: { title: ['自定义状态页', 'Customize status page'] },
        },
        {
          path: 'incidents',
          component: () => import('./views/IncidentsView.vue'),
          meta: { title: ['事件公告', 'Incidents'] },
        },
        {
          path: 'maintenance',
          component: () => import('./views/MaintenanceView.vue'),
          meta: { title: ['计划维护', 'Maintenance'] },
        },
        {
          path: 'notifications',
          component: () => import('./views/NotificationsView.vue'),
          meta: { title: ['通知与投递', 'Notifications'] },
        },
        {
          path: 'servers',
          component: () => import('./views/ServersView.vue'),
          meta: { title: ['服务器', 'Servers'] },
        },
        {
          path: 'settings',
          component: () => import('./views/SettingsView.vue'),
          meta: { title: ['设置', 'Settings'] },
        },
        {
          path: 'users',
          component: () => import('./views/UsersView.vue'),
          meta: { title: ['成员与权限', 'Members'], roles: ['admin'] },
        },
        {
          path: 'secrets',
          component: () => import('./views/SecretsView.vue'),
          meta: { title: ['秘密凭据', 'Secrets'], roles: ['admin'] },
        },
        {
          path: 'audit',
          component: () => import('./views/AuditView.vue'),
          meta: { title: ['审计日志', 'Audit log'], roles: ['admin'] },
        },
      ],
    },
    { path: '/:slug?/:rest(.*)*', component: () => import('./views/PublicView.vue') },
  ],
})
let sessionLoaded = false
router.beforeEach(async (to) => {
  if (!to.path.startsWith('/app') || to.path === '/app/login') return
  if (!sessionLoaded) {
    try {
      await loadSession()
    } catch {}
    sessionLoaded = true
  }
  if (!currentUser.value) return { path: '/app/login', query: { next: to.fullPath } }
  const roles = to.meta.roles as string[] | undefined
  if (roles && !roles.includes(currentUser.value.role)) return { path: '/app' }
})
