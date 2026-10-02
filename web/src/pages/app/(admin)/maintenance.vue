<script setup lang="ts">
import { ref, reactive } from 'vue'
import { Plus, Pencil, Trash2, CalendarClock } from '@lucide/vue'
import { useCollection } from '../../../lib/data'
import type { Maintenance, Monitor, Page } from '../../../lib/types'
import { api, canEdit } from '../../../lib/api'
import {
  t,
  formatDate,
  timezone,
  datetimeInput,
  datetimeMilliseconds,
} from '../../../lib/preferences'
import { notify, errorText } from '../../../lib/notices'
import { clone } from '../../../lib/form'
import PageHeader from '../../../components/PageHeader.vue'
import Field from '../../../components/Field.vue'
import Modal from '../../../components/Modal.vue'
import AsyncState from '../../../components/AsyncState.vue'
import EmptyState from '../../../components/EmptyState.vue'

definePage({ meta: { title: ['计划维护', 'Maintenance'] } })

const query = useCollection<Maintenance>('maintenance'),
  monitors = useCollection<Monitor>('monitors'),
  pages = useCollection<Page>('pages'),
  open = ref(false),
  saving = ref(false),
  error = ref(''),
  start = ref(''),
  end = ref(''),
  deleteTarget = ref<Maintenance | null>(null),
  deleteOpen = ref(false)
const empty = (): Maintenance => ({
    id: '',
    name: '',
    description: '',
    monitorIds: [],
    pageIds: [],
    startsAt: Date.now() + 3600000,
    endsAt: Date.now() + 7200000,
    timezone: timezone.value,
    createdAt: 0,
    updatedAt: 0,
  }),
  form = reactive(empty())
function edit(value?: Maintenance) {
  Object.assign(form, value ? clone(value) : empty())
  start.value = datetimeInput(form.startsAt, form.timezone)
  end.value = datetimeInput(form.endsAt, form.timezone)
  error.value = ''
  open.value = true
}
async function save() {
  saving.value = true
  error.value = ''
  try {
    const body = {
      ...form,
      startsAt: datetimeMilliseconds(start.value, form.timezone),
      endsAt: datetimeMilliseconds(end.value, form.timezone),
    }
    await api(form.id ? `maintenance/${form.id}` : 'maintenance', {
      method: form.id ? 'PATCH' : 'POST',
      body,
    })
    open.value = false
    notify(t('维护计划已保存', 'Maintenance saved'))
    await query.refresh()
  } catch (e) {
    error.value = errorText(e)
  } finally {
    saving.value = false
  }
}
async function remove() {
  if (!deleteTarget.value) return
  try {
    await api(`maintenance/${deleteTarget.value.id}`, { method: 'DELETE' })
    deleteOpen.value = false
    notify(t('维护计划已删除', 'Maintenance deleted'))
    await query.refresh()
  } catch (e) {
    notify(errorText(e), 'error')
  }
}
function maintenanceStatus(window: Maintenance) {
  return window.endsAt < Date.now()
    ? t('已结束', 'Completed')
    : window.startsAt > Date.now()
      ? t('计划中', 'Scheduled')
      : t('进行中', 'In progress')
}
function confirmDelete(value: Maintenance) {
  deleteTarget.value = value
  deleteOpen.value = true
}
</script>
<template>
  <PageHeader
    :title="t('计划维护', 'Maintenance')"
    :description="
      t(
        '计划内工作继续采集，排除统计时长并抑制故障与恢复通知。',
        'Planned work keeps collection running, excludes duration, and suppresses outage/recovery notifications.',
      )
    "
    ><button v-if="canEdit()" class="button primary" @click="edit()">
      <Plus :size="15" />{{ t('安排维护', 'Schedule maintenance') }}
    </button></PageHeader
  >
  <section class="card">
    <AsyncState :pending="query.isPending.value" :error="query.error.value" @retry="query.refresh()"
      ><EmptyState
        v-if="!query.data.value?.items.length"
        :title="t('暂无维护计划', 'No scheduled maintenance')"
        :description="
          t(
            '安排升级窗口，并提前通知状态页访问者。',
            'Plan an upgrade window and inform status-page visitors in advance.',
          )
        " />
      <div v-else class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th>{{ t('维护计划', 'Maintenance') }}</th>
              <th>{{ t('时间窗口', 'Window') }}</th>
              <th>{{ t('范围', 'Scope') }}</th>
              <th>{{ t('状态', 'Status') }}</th>
              <th />
            </tr>
          </thead>
          <tbody>
            <tr v-for="window in query.data.value.items" :key="window.id">
              <td>
                <span class="monitor-name" un-flex="~ items-center gap-2"
                  ><CalendarClock :size="15" />{{ window.name }}</span
                ><span class="monitor-sub">{{ window.description }}</span>
              </td>
              <td class="muted" un-text="10px">
                {{ formatDate(window.startsAt) }}<br />{{ formatDate(window.endsAt) }}
              </td>
              <td>{{ window.monitorIds.length }} {{ t('项监控', 'monitors') }}</td>
              <td>
                <span class="pill">{{ maintenanceStatus(window) }}</span>
              </td>
              <td>
                <div v-if="canEdit()" un-flex="~ gap-1">
                  <button class="icon-button" @click="edit(window)" :aria-label="t('编辑', 'Edit')">
                    <Pencil :size="14" /></button
                  ><button
                    class="icon-button"
                    @click="confirmDelete(window)"
                    :aria-label="t('删除', 'Delete')"
                  >
                    <Trash2 :size="14" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table></div
    ></AsyncState>
  </section>
  <Modal
    v-model:open="open"
    :title="form.id ? t('编辑维护计划', 'Edit maintenance') : t('安排维护', 'Schedule maintenance')"
    wide
    ><form id="maintenance-form" @submit.prevent="save">
      <div class="form-grid">
        <Field class="span-full" :label="t('名称', 'Name')"
          ><input v-model="form.name" required /></Field
        ><Field class="span-full" :label="t('描述', 'Description')">
          <textarea v-model="form.description" /></Field
        ><Field :label="t('开始时间', 'Starts at')"
          ><input v-model="start" type="datetime-local" required /></Field
        ><Field :label="t('结束时间', 'Ends at')"
          ><input v-model="end" type="datetime-local" required /></Field
        ><Field
          class="span-full"
          :label="t('时间窗口时区', 'Window time zone')"
          :hint="
            t(
              '上方输入以此时区的当地时间解释。',
              'The inputs above are interpreted as wall-clock time in this time zone.',
            )
          "
          ><input v-model="form.timezone" required placeholder="Asia/Shanghai"
        /></Field>
        <div class="span-full">
          <label class="field-label">{{ t('影响的监控项', 'Affected monitors') }}</label>
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
        <div class="span-full">
          <label class="field-label">{{ t('公开展示到状态页', 'Show on status pages') }}</label>
          <div class="checkbox-group" un-mt="3">
            <label v-for="page in pages.data.value?.items" :key="page.id" class="checkbox-label"
              ><input v-model="form.pageIds" type="checkbox" :value="page.id" />{{
                page.name
              }}</label
            >
          </div>
        </div>
      </div>
      <p v-if="error" class="inline-error">{{ error }}</p>
    </form>
    <template #footer
      ><button class="button" @click="open = false">{{ t('取消', 'Cancel') }}</button
      ><button class="button primary" form="maintenance-form" :disabled="saving">
        {{ t('保存计划', 'Save schedule') }}
      </button></template
    ></Modal
  ><Modal v-model:open="deleteOpen" :title="t('删除维护计划', 'Delete maintenance')"
    ><p>{{ deleteTarget?.name }}</p>
    <template #footer
      ><button class="button" @click="deleteOpen = false">{{ t('取消', 'Cancel') }}</button
      ><button class="button danger" @click="remove">{{ t('删除', 'Delete') }}</button></template
    ></Modal
  >
</template>
