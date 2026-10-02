<script setup lang="ts">
import type { DefineQueryOptions } from '@pinia/colada'
import type { ErrorModel } from '../../../client/types.gen'
import type { Maintenance, Monitor, Page } from '../../../lib/types'
import { useMutation, useQuery } from '@pinia/colada'
import { reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  createMaintenanceMutation,
  deleteMaintenanceMutation,
  listMaintenanceQuery,
  listMonitorsQuery,
  listPagesQuery,
  updateMaintenanceMutation,
} from '../../../client/@pinia/colada.gen'
import AsyncState from '../../../components/AsyncState.vue'
import EmptyState from '../../../components/EmptyState.vue'
import Field from '../../../components/Field.vue'
import Modal from '../../../components/Modal.vue'
import PageHeader from '../../../components/PageHeader.vue'
import { canEdit } from '../../../composables/api'
import { notify } from '../../../composables/notices'
import {
  datetimeInput,
  datetimeMilliseconds,
  formatDate,
  timezone,
} from '../../../composables/preferences'
import { errorText } from '../../../lib/errors'
import { clone } from '../../../lib/form'

const { t } = useI18n({ useScope: 'global' })

definePage({ meta: { title: 'navigation.maintenance' } })

const createMaintenance = useMutation(createMaintenanceMutation())
const updateMaintenance = useMutation(updateMaintenanceMutation())
const deleteMaintenance = useMutation(deleteMaintenanceMutation())

const query = useQuery({ ...listMaintenanceQuery(), staleTime: 10000 } as DefineQueryOptions<
  { items: Maintenance[] },
  ErrorModel
>)
const monitors = useQuery({ ...listMonitorsQuery(), staleTime: 10000 } as DefineQueryOptions<
  { items: Monitor[] },
  ErrorModel
>)
const pages = useQuery({ ...listPagesQuery(), staleTime: 10000 } as DefineQueryOptions<
  { items: Page[] },
  ErrorModel
>)
const open = ref(false)
const saving = ref(false)
const error = ref('')
const start = ref('')
const end = ref('')
const deleteTarget = ref<Maintenance | null>(null)
const deleteOpen = ref(false)
function empty(): Maintenance {
  return {
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
  }
}
const form = reactive(empty())
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
    if (form.id)
      await updateMaintenance.mutateAsync({ path: { id: form.id }, body })
    else await createMaintenance.mutateAsync({ body })
    open.value = false
    notify(t('maintenance.maintenanceSaved'))
    await query.refresh()
  }
  catch (e) {
    error.value = errorText(e)
  }
  finally {
    saving.value = false
  }
}
async function remove() {
  if (!deleteTarget.value)
    return
  try {
    await deleteMaintenance.mutateAsync({ path: { id: deleteTarget.value.id } })
    deleteOpen.value = false
    notify(t('maintenance.maintenanceDeleted'))
    await query.refresh()
  }
  catch (e) {
    notify(errorText(e), 'error')
  }
}
function maintenanceStatus(window: Maintenance) {
  return window.endsAt < Date.now()
    ? t('maintenance.completed')
    : window.startsAt > Date.now()
      ? t('maintenance.scheduled')
      : t('maintenance.inProgress')
}
function confirmDelete(value: Maintenance) {
  deleteTarget.value = value
  deleteOpen.value = true
}
</script>

<template>
  <PageHeader
    :title="t('navigation.maintenance')"
    :description="t('maintenance.plannedWorkKeepsCollectionRunningExcludesDurationAnd')"
  >
    <button v-if="canEdit()" class="button primary" @click="edit()">
      <span class="i-lucide-plus" un-w="15px" un-h="15px" aria-hidden="true" />{{ t('common.scheduleMaintenance') }}
    </button>
  </PageHeader>
  <section class="card">
    <AsyncState :pending="query.isPending.value" :error="query.error.value" @retry="query.refetch()">
      <EmptyState
        v-if="!query.data.value?.items.length"
        :title="t('maintenance.noScheduledMaintenance')"
        :description="t('maintenance.planAnUpgradeWindowAndInformStatusPage')"
      />
      <div v-else class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th>{{ t('maintenance.maintenance') }}</th>
              <th>{{ t('maintenance.window') }}</th>
              <th>{{ t('maintenance.scope') }}</th>
              <th>{{ t('common.status') }}</th>
              <th />
            </tr>
          </thead>
          <tbody>
            <tr v-for="window in query.data.value.items" :key="window.id">
              <td>
                <span class="monitor-name" un-flex="~ items-center gap-2"><span class="i-lucide-calendar-clock" un-w="15px" un-h="15px" aria-hidden="true" />{{ window.name }}</span><span class="monitor-sub">{{ window.description }}</span>
              </td>
              <td class="muted" un-text="10px">
                {{ formatDate(window.startsAt) }}<br>{{ formatDate(window.endsAt) }}
              </td>
              <td>
                {{
                  t(
                    'counts.monitors',
                    { count: window.monitorIds.length },
                    window.monitorIds.length,
                  )
                }}
              </td>
              <td>
                <span class="pill">{{ maintenanceStatus(window) }}</span>
              </td>
              <td>
                <div v-if="canEdit()" un-flex="~ gap-1">
                  <button class="icon-button" :aria-label="t('common.edit')" @click="edit(window)">
                    <span class="i-lucide-pencil" un-w="14px" un-h="14px" aria-hidden="true" />
                  </button><button
                    class="icon-button"
                    :aria-label="t('common.delete')"
                    @click="confirmDelete(window)"
                  >
                    <span class="i-lucide-trash-2" un-w="14px" un-h="14px" aria-hidden="true" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </AsyncState>
  </section>
  <Modal
    v-model:open="open"
    :title="form.id ? t('maintenance.editMaintenance') : t('common.scheduleMaintenance')"
    wide
  >
    <form id="maintenance-form" @submit.prevent="save">
      <div class="form-grid">
        <Field class="span-full" :label="t('common.name')">
          <input v-model="form.name" required>
        </Field><Field class="span-full" :label="t('maintenance.description')">
          <textarea v-model="form.description" />
        </Field><Field :label="t('maintenance.startsAt')">
          <input v-model="start" type="datetime-local" required>
        </Field><Field :label="t('maintenance.endsAt')">
          <input v-model="end" type="datetime-local" required>
        </Field><Field
          class="span-full"
          :label="t('maintenance.windowTimeZone')"
          :hint="t('maintenance.theInputsAboveAreInterpretedAsWallClock')"
        >
          <input v-model="form.timezone" required placeholder="Asia/Shanghai">
        </Field>
        <div class="span-full">
          <label class="field-label">{{ t('maintenance.affectedMonitors') }}</label>
          <div class="checkbox-group" un-mt="3">
            <label
              v-for="monitor in monitors.data.value?.items"
              :key="monitor.id"
              class="checkbox-label"
            ><input v-model="form.monitorIds" type="checkbox" :value="monitor.id">{{
              monitor.name
            }}</label>
          </div>
        </div>
        <div class="span-full">
          <label class="field-label">{{ t('maintenance.showOnStatusPages') }}</label>
          <div class="checkbox-group" un-mt="3">
            <label v-for="page in pages.data.value?.items" :key="page.id" class="checkbox-label"><input v-model="form.pageIds" type="checkbox" :value="page.id">{{
              page.name
            }}</label>
          </div>
        </div>
      </div>
      <p v-if="error" class="inline-error">
        {{ error }}
      </p>
    </form>
    <template #footer>
      <button class="button" @click="open = false">
        {{ t('common.cancel') }}
      </button><button class="button primary" form="maintenance-form" :disabled="saving">
        {{ t('maintenance.saveSchedule') }}
      </button>
    </template>
  </Modal><Modal v-model:open="deleteOpen" :title="t('maintenance.deleteMaintenance')">
    <p>{{ deleteTarget?.name }}</p>
    <template #footer>
      <button class="button" @click="deleteOpen = false">
        {{ t('common.cancel') }}
      </button><button class="button danger" @click="remove">
        {{ t('common.delete') }}
      </button>
    </template>
  </Modal>
</template>
