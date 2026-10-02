<script setup lang="ts">
import { reactive, ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  Save,
  Send,
  Plus,
  X,
  ArrowUp,
  ArrowDown,
  ArrowLeft,
  Upload,
  Trash2,
  ExternalLink,
} from '@lucide/vue'
import { api, canEdit } from '../lib/api'
import type { Page, Monitor, Settings, PublicPage, PageConfig } from '../lib/types'
import { publishedEntry } from '../lib/pages'
import { t, formatDate } from '../lib/preferences'
import { notify, errorText } from '../lib/notices'
import { clone } from '../lib/form'
import PageHeader from './PageHeader.vue'
import Field from './Field.vue'
import StatusPage from './StatusPage.vue'
import AsyncState from './AsyncState.vue'
import Modal from './Modal.vue'
import Toggle from './Toggle.vue'
const origin = location.origin
const newID = () =>
  crypto.randomUUID?.() || `group-${Date.now().toString(36)}-${Math.random().toString(36).slice(2)}`
const route = useRoute<'/app/(admin)/pages/new' | '/app/(admin)/pages/[id]'>(),
  router = useRouter(),
  id = computed(() => ('id' in route.params ? route.params.id : undefined)),
  editing = computed(() => !!id.value || !!form.id),
  loading = ref(true),
  saving = ref(false),
  error = ref(''),
  monitors = ref<Monitor[]>([]),
  allowedDomains = ref<string[]>([]),
  savedPreview = ref<PublicPage | null>(null),
  deleteOpen = ref(false),
  newMonitorIds = reactive<Record<string, string>>({})
const form = reactive<Page>({
  id: '',
  name: '',
  slug: '',
  domain: '',
  draft: {
    title: '',
    description: '',
    logoUrl: '',
    brandColor: '#0c8b76',
    colorScheme: 'system',
    links: [],
    groups: [],
  },
  publishedAt: 0,
  version: 0,
  createdAt: 0,
  updatedAt: 0,
})
function replacePage(data: Page) {
  // Empty optional publication fields are omitted by Go. Clear an earlier
  // snapshot before applying the canonical response, including domain removal.
  Object.assign(
    form,
    { publishedSlug: undefined, publishedDomain: undefined, published: undefined },
    data,
  )
}
async function load() {
  loading.value = true
  try {
    const [m, s] = await Promise.all([
      api<{ items: Monitor[] }>('monitors'),
      api<Settings>('settings'),
    ])
    monitors.value = m.items
    allowedDomains.value = s.allowedDomains || []
    if (editing.value) {
      replacePage(await api<Page>(`pages/${id.value}`))
      savedPreview.value = await api<PublicPage>(`pages/${id.value}/preview`)
    }
  } catch (e) {
    error.value = errorText(e)
  } finally {
    loading.value = false
  }
}
onMounted(load)
const preview = computed<PublicPage>(() => {
  const existing = savedPreview.value?.groups.flatMap((g) => g.monitors) || []
  const groups = form.draft.groups.map((group) => ({
    id: group.id,
    name: group.name,
    monitors: group.monitors.map((pm) => {
      const live = existing.find((x) => x.id === pm.monitorId)
      if (live)
        return { ...live, name: pm.alias || live.name, latency: pm.showLatency ? live.latency : [] }
      const monitor = monitors.value.find((x) => x.id === pm.monitorId)
      return {
        id: pm.monitorId,
        name: pm.alias || monitor?.name || '',
        type: monitor?.type || 'http',
        state: monitor?.state || 'unknown',
        paused: !monitor?.enabled,
        maintenance: false,
        availability: {
          from: 0,
          to: 0,
          upMs: 0,
          downMs: 0,
          unknownMs: 0,
          excludedMs: 0,
          effectiveMs: 0,
          uptime: null,
          coverage: null,
        },
        latency: [],
        ...(monitor?.certificate
          ? {
              certificate: {
                state: monitor.certificate.state,
                expiresAt: monitor.certificate.expiresAt,
                daysRemaining: monitor.certificate.daysRemaining,
              },
            }
          : {}),
      }
    }),
  }))
  const states = groups
    .flatMap((g) => g.monitors)
    .filter((m) => m.type !== 'certificate' && !m.paused)
    .map((m) => (m.maintenance ? 'maintenance' : m.state))
  let state = !states.length
    ? 'unknown'
    : states.every((s) => s === 'down')
      ? 'outage'
      : states.includes('down')
        ? 'partial'
        : states.includes('unknown')
          ? 'unknown'
          : states.includes('maintenance')
            ? 'maintenance'
            : 'operational'
  const incidents = savedPreview.value?.incidents || []
  if (incidents.some((i) => i.status !== 'resolved' && i.impact === 'outage')) state = 'outage'
  else if (
    state !== 'outage' &&
    incidents.some((i) => i.status !== 'resolved' && i.impact === 'partial')
  )
    state = 'partial'
  return {
    id: form.id,
    slug: form.slug,
    config: clone(form.draft),
    state,
    groups,
    incidents,
    maintenance: savedPreview.value?.maintenance || [],
    updatedAt: savedPreview.value?.updatedAt || form.updatedAt,
  }
})
function move<T>(values: T[], index: number, delta: number) {
  const target = index + delta
  if (target < 0 || target >= values.length) return
  const [value] = values.splice(index, 1)
  values.splice(target, 0, value!)
}
function addMonitor(groupId: string) {
  const group = form.draft.groups.find((x) => x.id === groupId),
    id = newMonitorIds[groupId]
  if (!group || !id) return
  if (form.draft.groups.some((g) => g.monitors.some((m) => m.monitorId === id))) {
    notify(t('此监控项已被选入页面', 'This monitor is already on the page'), 'error')
    return
  }
  group.monitors.push({
    monitorId: id,
    alias: monitors.value.find((m) => m.id === id)?.name || '',
    showUptime: true,
    showLatency: true,
  })
  newMonitorIds[groupId] = ''
}
async function save(publish = false) {
  saving.value = true
  error.value = ''
  try {
    if (!form.name.trim() || !form.draft.title.trim())
      throw new Error(t('请填写页面名称与标题', 'Page name and title are required'))
    if (!/^[a-z0-9][a-z0-9-]*$/.test(form.slug))
      throw new Error(
        t(
          '路径标识只允许小写字母、数字与连字符',
          'Slug may contain lowercase letters, numbers, and hyphens',
        ),
      )
    for (const link of form.draft.links) {
      const parsed = new URL(link.url)
      if (!['https:', 'http:'].includes(parsed.protocol) || !link.label.trim())
        throw new Error(
          t('公开链接需要名称与 HTTP(S) 地址', 'Public links need labels and HTTP(S) URLs'),
        )
    }
    if (
      form.draft.logoUrl &&
      !/^https:\/\//.test(form.draft.logoUrl) &&
      !form.draft.logoUrl.startsWith('/assets/')
    )
      throw new Error(
        t('Logo 需使用 HTTPS 或上传资源地址', 'Logo must use HTTPS or an uploaded asset URL'),
      )
    const data = await api<Page>(editing.value ? `pages/${form.id}` : 'pages', {
      method: editing.value ? 'PATCH' : 'POST',
      body: clone(form),
    })
    replacePage(data)
    if (publish) {
      await api<Page>(`pages/${form.id}/publish`, { method: 'POST' })
      replacePage(await api<Page>(`pages/${form.id}`))
      notify(t('状态页已发布', 'Status page published'))
    } else notify(t('草稿已保存', 'Draft saved'))
    savedPreview.value = await api<PublicPage>(`pages/${form.id}/preview`)
    if (!id.value) await router.replace(`/app/pages/${form.id}`)
  } catch (e) {
    error.value = errorText(e)
  } finally {
    saving.value = false
  }
}
async function uploadLogo(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return
  if (file.size > 4 * 1024 * 1024) {
    notify(t('图片需小于 4 MiB', 'Image must be smaller than 4 MiB'), 'error')
    return
  }
  try {
    const base64 = await new Promise<string>((resolve, reject) => {
      const reader = new FileReader()
      reader.onload = () => resolve(String(reader.result).split(',')[1] || '')
      reader.onerror = reject
      reader.readAsDataURL(file)
    })
    form.draft.logoUrl = (
      await api<{ url: string }>('assets', {
        method: 'POST',
        body: { filename: file.name, contentType: file.type, base64 },
      })
    ).url
    notify(t('Logo 已上传', 'Logo uploaded'))
  } catch (e) {
    notify(errorText(e), 'error')
  }
}
async function remove() {
  try {
    await api(`pages/${form.id}`, { method: 'DELETE' })
    notify(t('状态页已删除', 'Status page deleted'))
    router.push('/app/pages')
  } catch (e) {
    notify(errorText(e), 'error')
  }
}
</script>
<template>
  <PageHeader
    :title="
      editing
        ? form.name || t('自定义状态页', 'Customize status page')
        : t('创建状态页', 'Create status page')
    "
    :description="
      t(
        '每页独立自定义，保存草稿后显式发布。',
        'Customize each page independently. Save a draft, then publish.',
      )
    "
    ><RouterLink to="/app/pages" class="button ghost"
      ><ArrowLeft :size="14" />{{ t('列表', 'All pages') }}</RouterLink
    ><template v-if="canEdit()"
      ><button class="button" :disabled="saving" @click="save()">
        <Save :size="14" />{{ t('保存草稿', 'Save draft') }}</button
      ><button class="button primary" :disabled="saving" @click="save(true)">
        <Send :size="14" />{{ t('发布页面', 'Publish page') }}
      </button></template
    ></PageHeader
  ><AsyncState :pending="loading"
    ><div v-if="error" class="validation-error" role="alert">{{ error }}</div>
    <div class="editor-layout">
      <div>
        <section class="card">
          <div class="form-section">
            <h2>{{ t('页面入口', 'Page access') }}</h2>
            <p class="muted">
              {{
                t(
                  '路径与独立域名访问相同的已发布内容；入口变更在发布时生效。',
                  'Path and custom domain serve the same published content. Entry changes take effect on publication.',
                )
              }}
            </p>
            <div class="form-grid">
              <Field :label="t('后台页面名称', 'Internal page name')"
                ><input v-model="form.name" :disabled="!canEdit()" required /></Field
              ><Field
                :label="t('路径标识', 'Page slug')"
                :hint="`${origin}/${form.slug || 'status1'}`"
                ><input
                  v-model="form.slug"
                  :disabled="!canEdit()"
                  placeholder="status1"
                  pattern="[a-z0-9][a-z0-9-]*"
                  required /></Field
              ><Field
                class="span-full"
                :label="t('独立域名', 'Custom domain')"
                :hint="
                  t(
                    '管理员预先配置域名入口，DNS 与 HTTPS 由部署反代提供。',
                    'Select an administrator-configured domain. DNS and HTTPS are handled by your reverse proxy.',
                  )
                "
                ><select v-model="form.domain" :disabled="!canEdit()">
                  <option value="">{{ t('仅使用路径访问', 'Path access only') }}</option>
                  <option v-for="domain in allowedDomains" :key="domain">{{ domain }}</option>
                </select></Field
              >
            </div>
            <p v-if="form.publishedAt" class="note" un-mt="5">
              {{ t('上次发布', 'Last published') }} {{ formatDate(form.publishedAt) }} · v{{
                form.version
              }}<a
                :href="publishedEntry(form).url"
                class="button small ghost"
                target="_blank"
                rel="noopener"
                ><ExternalLink :size="12" />{{ t('访问公开页面', 'Visit public page') }}</a
              >
            </p>
          </div>
          <div class="form-section">
            <h2>{{ t('品牌与外观', 'Brand & appearance') }}</h2>
            <p class="muted">
              {{ t('设置只应用于此状态页。', 'These settings apply only to this status page.') }}
            </p>
            <div class="form-grid">
              <Field class="span-full" :label="t('公开标题', 'Public title')"
                ><input v-model="form.draft.title" :disabled="!canEdit()" required /></Field
              ><Field class="span-full" :label="t('页面描述', 'Page description')">
                <textarea v-model="form.draft.description" :disabled="!canEdit()" rows="3" /></Field
              ><Field class="span-full" :label="t('Logo 地址', 'Logo URL')"
                ><div class="field-row">
                  <input
                    v-model="form.draft.logoUrl"
                    :disabled="!canEdit()"
                    placeholder="https://… or /assets/uploads/…"
                  /><label v-if="canEdit()" class="button"
                    ><Upload :size="14" />{{ t('上传', 'Upload')
                    }}<input
                      type="file"
                      accept="image/png,image/jpeg,image/gif"
                      un-hidden=""
                      @change="uploadLogo"
                  /></label></div></Field
              ><Field :label="t('品牌色', 'Brand color')"
                ><div class="field-row">
                  <input
                    v-model="form.draft.brandColor"
                    type="color"
                    :disabled="!canEdit()"
                  /><input
                    v-model="form.draft.brandColor"
                    :disabled="!canEdit()"
                    pattern="#[0-9a-fA-F]{6}"
                  /></div></Field
              ><Field :label="t('深浅色', 'Color scheme')"
                ><select v-model="form.draft.colorScheme" :disabled="!canEdit()">
                  <option value="system">{{ t('跟随系统', 'System') }}</option>
                  <option value="light">{{ t('浅色', 'Light') }}</option>
                  <option value="dark">{{ t('深色', 'Dark') }}</option>
                </select></Field
              >
              <div class="span-full">
                <label class="field-label">{{ t('公开链接', 'Public links') }}</label>
                <div
                  v-for="(link, index) in form.draft.links"
                  :key="index"
                  class="link-editor"
                  un-mt="3"
                >
                  <input
                    v-model="link.label"
                    :disabled="!canEdit()"
                    :placeholder="t('链接名称', 'Link label')"
                  /><input
                    v-model="link.url"
                    :disabled="!canEdit()"
                    type="url"
                    placeholder="https://…"
                  /><button
                    v-if="canEdit()"
                    class="icon-button"
                    @click="form.draft.links.splice(index, 1)"
                    :aria-label="t('移除链接', 'Remove link')"
                  >
                    <X :size="14" />
                  </button>
                </div>
                <button
                  v-if="canEdit()"
                  class="button small ghost"
                  un-mt="2"
                  @click="form.draft.links.push({ label: '', url: '' })"
                >
                  <Plus :size="13" />{{ t('添加链接', 'Add link') }}
                </button>
              </div>
            </div>
          </div>
          <div class="form-section">
            <h2>{{ t('服务与分组', 'Services & groups') }}</h2>
            <p class="muted">
              {{
                t(
                  '只发布明确选中的监控项。公开别名不会改变后台名称。',
                  'Publish only selected monitors. Public aliases leave internal names unchanged.',
                )
              }}
            </p>
            <div v-for="(group, index) in form.draft.groups" :key="group.id" class="group-editor">
              <div class="group-editor-header">
                <input
                  :aria-label="t('分组名称', 'Group name')"
                  v-model="group.name"
                  :disabled="!canEdit()"
                  :placeholder="t('分组名称', 'Group name')"
                /><template v-if="canEdit()"
                  ><button
                    class="icon-button"
                    :disabled="index === 0"
                    @click="move(form.draft.groups, index, -1)"
                    :aria-label="t('分组上移', 'Move group up')"
                  >
                    <ArrowUp :size="13" /></button
                  ><button
                    class="icon-button"
                    :disabled="index === form.draft.groups.length - 1"
                    @click="move(form.draft.groups, index, 1)"
                    :aria-label="t('分组下移', 'Move group down')"
                  >
                    <ArrowDown :size="13" /></button
                  ><button
                    class="icon-button"
                    @click="form.draft.groups.splice(index, 1)"
                    :aria-label="t('移除分组', 'Remove group')"
                  >
                    <X :size="14" /></button
                ></template>
              </div>
              <div
                v-for="(item, mIndex) in group.monitors"
                :key="item.monitorId"
                class="group-editor-row"
              >
                <div un-flex="1" un-min-w="0">
                  <span class="mini-label">{{
                    monitors.find((x) => x.id === item.monitorId)?.name
                  }}</span
                  ><input
                    :aria-label="t('公开别名', 'Public alias')"
                    v-model="item.alias"
                    :disabled="!canEdit()"
                    :placeholder="t('公开别名', 'Public alias')"
                    un-mt="1"
                  />
                  <div un-flex="~ gap-4" un-mt="2">
                    <label class="checkbox-label"
                      ><input v-model="item.showUptime" type="checkbox" :disabled="!canEdit()" />{{
                        t('可用率', 'Uptime')
                      }}</label
                    ><label class="checkbox-label"
                      ><input v-model="item.showLatency" type="checkbox" :disabled="!canEdit()" />{{
                        t('延迟', 'Latency')
                      }}</label
                    >
                  </div>
                </div>
                <template v-if="canEdit()"
                  ><button
                    class="icon-button"
                    :disabled="mIndex === 0"
                    @click="move(group.monitors, mIndex, -1)"
                    :aria-label="t('服务上移', 'Move service up')"
                  >
                    <ArrowUp :size="13" /></button
                  ><button
                    class="icon-button"
                    :disabled="mIndex === group.monitors.length - 1"
                    @click="move(group.monitors, mIndex, 1)"
                    :aria-label="t('服务下移', 'Move service down')"
                  >
                    <ArrowDown :size="13" /></button
                  ><button
                    class="icon-button"
                    @click="group.monitors.splice(mIndex, 1)"
                    :aria-label="t('移除服务', 'Remove service')"
                  >
                    <X :size="14" /></button
                ></template>
              </div>
              <div v-if="canEdit()" class="group-editor-row">
                <select
                  v-model="newMonitorIds[group.id]"
                  :aria-label="t('选择监控项', 'Select a monitor')"
                >
                  <option value="">{{ t('选择监控项', 'Select a monitor') }}</option>
                  <option v-for="monitor in monitors" :key="monitor.id" :value="monitor.id">
                    {{ monitor.name }} · {{ monitor.type }}
                  </option></select
                ><button class="button small" @click="addMonitor(group.id)">
                  <Plus :size="13" />{{ t('加入', 'Add') }}
                </button>
              </div>
            </div>
            <button
              v-if="canEdit()"
              class="button small"
              un-mt="4"
              @click="
                form.draft.groups.push({
                  id: newID(),
                  name: t('服务分组', 'Services'),
                  monitors: [],
                })
              "
            >
              <Plus :size="13" />{{ t('添加分组', 'Add group') }}
            </button>
          </div>
        </section>
        <div class="form-actions">
          <button
            v-if="editing && canEdit()"
            class="button danger"
            un-mr="auto"
            @click="deleteOpen = true"
          >
            <Trash2 :size="13" />{{ t('删除页面', 'Delete page') }}</button
          ><button v-if="canEdit()" class="button primary" :disabled="saving" @click="save()">
            <Save :size="14" />{{ t('保存草稿', 'Save draft') }}
          </button>
        </div>
      </div>
      <aside class="editor-preview">
        <div class="preview-label">
          <strong>{{ t('实时草稿预览', 'Live draft preview') }}</strong
          ><span>{{ t('桌面 / 移动自适应', 'Responsive preview') }}</span>
        </div>
        <StatusPage :page="preview" preview />
        <p class="field-hint" un-mt="3">
          {{
            t(
              '保存后重新读取真实统计；未保存的新选项显示已有服务状态，暂无历史统计。发布前不会改变公开页面。',
              'Saving refreshes real statistics. Newly selected services show current states without fabricated history. Draft edits do not alter the public page.',
            )
          }}
        </p>
      </aside>
    </div></AsyncState
  ><Modal
    v-model:open="deleteOpen"
    :title="t('删除状态页', 'Delete status page')"
    :description="
      t('该页面的路径与域名入口将停止发布。', 'The page path and domain will stop publishing.')
    "
    ><template #footer
      ><button class="button" @click="deleteOpen = false">{{ t('取消', 'Cancel') }}</button
      ><button class="button danger" @click="remove">{{ t('删除', 'Delete') }}</button></template
    ></Modal
  >
</template>
