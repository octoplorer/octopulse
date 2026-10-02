<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { Plus, MessageSquare, CheckCircle2 } from '@lucide/vue'
import { useCollection } from '../../../lib/data'
import type { Incident, Page, Monitor } from '../../../lib/types'
import { response, canEdit } from '../../../lib/api'
import * as sdk from '../../../client/sdk.gen'
import {
  listIncidentsQuery,
  listPagesQuery,
  listMonitorsQuery,
} from '../../../client/@pinia/colada.gen'
import { t, formatDate, statusLabel } from '../../../lib/preferences'
import { notify, errorText } from '../../../lib/notices'
import { clone } from '../../../lib/form'
import PageHeader from '../../../components/PageHeader.vue'
import Field from '../../../components/Field.vue'
import Modal from '../../../components/Modal.vue'
import AsyncState from '../../../components/AsyncState.vue'
import EmptyState from '../../../components/EmptyState.vue'

definePage({ meta: { title: ['事件公告', 'Incidents'] } })

const query = useCollection<Incident>('incidents', listIncidentsQuery()),
  pages = useCollection<Page>('pages', listPagesQuery()),
  monitors = useCollection<Monitor>('monitors', listMonitorsQuery()),
  open = ref(false),
  detailOpen = ref(false),
  selected = ref<Incident | null>(null),
  saving = ref(false),
  error = ref(''),
  updateBody = ref(''),
  updateStatus = ref<Incident['status']>('investigating'),
  filter = ref('active')
const empty = (): Incident => ({
    id: '',
    title: '',
    body: '',
    status: 'investigating',
    impact: 'none',
    pageIds: [],
    monitorIds: [],
    updates: [],
    createdAt: 0,
    updatedAt: 0,
    resolvedAt: 0,
  }),
  form = reactive(empty()),
  items = computed(
    () =>
      query.data.value?.items
        .filter(
          (x) =>
            filter.value === 'all' ||
            (filter.value === 'active' ? x.status !== 'resolved' : x.status === 'resolved'),
        )
        .sort((a, b) => b.updatedAt - a.updatedAt) || [],
  )
function edit(incident: Incident) {
  Object.assign(form, clone(incident))
  error.value = ''
  detailOpen.value = false
  open.value = true
}
function create() {
  Object.assign(form, empty())
  error.value = ''
  open.value = true
}
function detail(incident: Incident) {
  selected.value = clone(incident)
  updateStatus.value = incident.status
  updateBody.value = ''
  detailOpen.value = true
}
async function save() {
  saving.value = true
  error.value = ''
  try {
    if (form.id)
      await sdk.updateIncidents({ path: { id: form.id }, body: form, throwOnError: true })
    else await sdk.createIncidents({ body: form, throwOnError: true })
    open.value = false
    notify(t('公告已保存', 'Incident saved'))
    await query.refresh()
  } catch (e) {
    error.value = errorText(e)
  } finally {
    saving.value = false
  }
}
async function update() {
  if (!selected.value) return
  saving.value = true
  try {
    const result = await response<Incident>(
      sdk.createIncidentUpdate({
        path: { id: selected.value.id },
        body: { body: updateBody.value, status: updateStatus.value },
        throwOnError: true,
      }),
    )
    selected.value = result
    updateBody.value = ''
    notify(t('进展已发布', 'Update published'))
    await query.refresh()
  } catch (e) {
    notify(errorText(e), 'error')
  } finally {
    saving.value = false
  }
}
</script>
<template>
  <PageHeader
    :title="t('事件公告', 'Incidents')"
    :description="
      t(
        '公开说明影响与修复进展，保持信息清晰一致。',
        'Communicate impact and progress with clear, consistent updates.',
      )
    "
    ><button v-if="canEdit()" class="button primary" @click="create">
      <Plus :size="15" />{{ t('创建公告', 'Create incident') }}
    </button></PageHeader
  >
  <section class="card">
    <div class="filter-bar">
      <h2 un-text="sm">{{ t('公告列表', 'Incident announcements') }}</h2>
      <select v-model="filter" un-w="auto!" :aria-label="t('公告状态', 'Incident filter')">
        <option value="active">{{ t('进行中', 'Active') }}</option>
        <option value="resolved">{{ t('已解决', 'Resolved') }}</option>
        <option value="all">{{ t('全部', 'All') }}</option>
      </select>
    </div>
    <AsyncState :pending="query.isPending.value" :error="query.error.value" @retry="query.refresh()"
      ><EmptyState
        v-if="!items.length"
        :title="t('暂无事件公告', 'No incidents here')"
        :description="
          t(
            '手动公告不会改变监控状态与可用率。',
            'Manual incidents do not change monitor states or uptime.',
          )
        "
      />
      <div v-else class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th>{{ t('公告', 'Incident') }}</th>
              <th>{{ t('进展', 'Progress') }}</th>
              <th>{{ t('影响', 'Impact') }}</th>
              <th>{{ t('更新时间', 'Updated') }}</th>
              <th />
            </tr>
          </thead>
          <tbody>
            <tr v-for="incident in items" :key="incident.id">
              <td>
                <button class="button ghost" un-p="0!" @click="detail(incident)">
                  <MessageSquare :size="15" />{{ incident.title }}</button
                ><span class="monitor-sub"
                  >{{ incident.pageIds.length }} {{ t('个公开页面', 'public pages') }}</span
                >
              </td>
              <td>
                <span class="pill">{{ statusLabel(incident.status) }}</span>
              </td>
              <td>{{ statusLabel(incident.impact) }}</td>
              <td class="muted" un-text="10px">{{ formatDate(incident.updatedAt) }}</td>
              <td>
                <button class="button small" @click="detail(incident)">
                  {{ t('查看进展', 'View updates') }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div></AsyncState
    >
  </section>
  <p class="note" un-mt="5">
    {{
      t(
        '活动公告可提升页面影响等级，不能降低自动检测到的故障。公告关闭后不再影响页面总体状态。',
        'Active incidents may raise a page’s impact level, but cannot reduce automatically detected outages. Resolved incidents no longer raise the overall status.',
      )
    }}
  </p>
  <Modal
    v-model:open="open"
    :title="form.id ? t('编辑事件公告', 'Edit incident') : t('创建事件公告', 'Create incident')"
    wide
    ><form id="incident-form" @submit.prevent="save">
      <div class="form-grid">
        <Field class="span-full" :label="t('标题', 'Title')"
          ><input v-model="form.title" required /></Field
        ><Field class="span-full" :label="t('公告内容', 'Description')">
          <textarea v-model="form.body" rows="5" required /></Field
        ><Field :label="t('进展状态', 'Progress status')"
          ><select v-model="form.status">
            <option
              v-for="state in ['investigating', 'identified', 'monitoring', 'resolved']"
              :key="state"
              :value="state"
            >
              {{ statusLabel(state) }}
            </option>
          </select></Field
        ><Field :label="t('影响等级', 'Impact level')"
          ><select v-model="form.impact">
            <option value="none">{{ t('仅信息', 'Informational') }}</option>
            <option value="partial">{{ t('部分故障', 'Partial outage') }}</option>
            <option value="outage">{{ t('全面故障', 'Major outage') }}</option>
          </select></Field
        >
        <div class="span-full">
          <label class="field-label">{{ t('发布到状态页', 'Publish to status pages') }}</label>
          <div class="checkbox-group" un-mt="3">
            <label v-for="page in pages.data.value?.items" :key="page.id" class="checkbox-label"
              ><input v-model="form.pageIds" type="checkbox" :value="page.id" />{{
                page.name
              }}</label
            >
          </div>
        </div>
        <div class="span-full">
          <label class="field-label">{{ t('关联监控项', 'Related monitors') }}</label>
          <div class="checkbox-group" un-mt="3">
            <label
              v-for="monitor in monitors.data.value?.items"
              :key="monitor.id"
              class="checkbox-label"
              ><input v-model="form.monitorIds" type="checkbox" :value="monitor.id" />{{
                monitor.name
              }}</label
            >
          </div>
        </div>
      </div>
      <p v-if="error" class="inline-error">{{ error }}</p>
    </form>
    <template #footer
      ><button class="button" @click="open = false">{{ t('取消', 'Cancel') }}</button
      ><button class="button primary" form="incident-form" :disabled="saving">
        {{ t('保存公告', 'Save incident') }}
      </button></template
    ></Modal
  ><Modal v-model:open="detailOpen" :title="selected?.title || ''" wide
    ><template v-if="selected"
      ><div un-flex="~ items-start justify-between gap-4" un-mb="5">
        <p class="muted">{{ selected.body }}</p>
        <button v-if="canEdit()" class="button small" @click="edit(selected)">
          {{ t('编辑公告', 'Edit incident') }}
        </button>
      </div>
      <div class="timeline">
        <div class="timeline-entry">
          <h3>{{ t('首次公告', 'Initial announcement') }}</h3>
          <small>{{ formatDate(selected.createdAt) }}</small>
        </div>
        <div v-for="update in selected.updates" :key="update.id" class="timeline-entry">
          <h3>{{ statusLabel(update.status) }}</h3>
          <p>{{ update.body }}</p>
          <small>{{ formatDate(update.createdAt) }}</small>
        </div>
      </div>
      <form v-if="canEdit()" @submit.prevent="update">
        <div class="section-divider" />
        <h3 un-mb="4">{{ t('发布进展', 'Publish an update') }}</h3>
        <Field :label="t('进展状态', 'Progress')"
          ><select v-model="updateStatus">
            <option
              v-for="state in ['investigating', 'identified', 'monitoring', 'resolved']"
              :key="state"
              :value="state"
            >
              {{ statusLabel(state) }}
            </option>
          </select></Field
        ><Field :label="t('说明', 'Update')" un-mt="4">
          <textarea v-model="updateBody" required rows="4" /></Field
        ><button class="button primary" :disabled="saving" un-mt="4">
          {{ t('发布进展', 'Publish update') }}
        </button>
      </form></template
    ></Modal
  >
</template>
