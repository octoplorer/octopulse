<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { ref, reactive, computed } from 'vue'
import { Plus, MessageSquare } from '@lucide/vue'
import type { Incident, Page, Monitor } from '../../../lib/types'
import { canEdit } from '../../../lib/api'
import { useQuery, useMutation, type DefineQueryOptions } from '@pinia/colada'
import type { ErrorModel } from '../../../client/types.gen'
import {
  listIncidentsQuery,
  listPagesQuery,
  listMonitorsQuery,
  createIncidentsMutation,
  updateIncidentsMutation,
  createIncidentUpdateMutation,
} from '../../../client/@pinia/colada.gen'
import { formatDate, statusLabel } from '../../../lib/preferences'
import { notify, errorText } from '../../../lib/notices'
import { clone } from '../../../lib/form'
import PageHeader from '../../../components/PageHeader.vue'
import Field from '../../../components/Field.vue'
import Modal from '../../../components/Modal.vue'
import AsyncState from '../../../components/AsyncState.vue'
import EmptyState from '../../../components/EmptyState.vue'

const { t } = useI18n({ useScope: 'global' })

definePage({ meta: { title: 'navigation.incidents' } })

const createIncident = useMutation(createIncidentsMutation())
const updateIncident = useMutation(updateIncidentsMutation())
const publishIncidentUpdate = useMutation(createIncidentUpdateMutation())

const query = useQuery({ ...listIncidentsQuery(), staleTime: 10000 } as DefineQueryOptions<
  { items: Incident[] },
  ErrorModel
>)
const pages = useQuery({ ...listPagesQuery(), staleTime: 10000 } as DefineQueryOptions<
  { items: Page[] },
  ErrorModel
>)
const monitors = useQuery({ ...listMonitorsQuery(), staleTime: 10000 } as DefineQueryOptions<
  { items: Monitor[] },
  ErrorModel
>)
const open = ref(false)
const detailOpen = ref(false)
const selected = ref<Incident | null>(null)
const saving = ref(false)
const error = ref('')
const updateBody = ref('')
const updateStatus = ref<Incident['status']>('investigating')
const filter = ref('active')
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
})
const form = reactive(empty())
const items = computed(
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
    if (form.id) await updateIncident.mutateAsync({ path: { id: form.id }, body: form })
    else await createIncident.mutateAsync({ body: form })
    open.value = false
    notify(t('incidents.incidentSaved'))
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
    const result = await publishIncidentUpdate.mutateAsync({
      path: { id: selected.value.id },
      body: { body: updateBody.value, status: updateStatus.value },
    })
    selected.value = result as Incident
    updateBody.value = ''
    notify(t('incidents.updatePublished'))
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
    :title="t('navigation.incidents')"
    :description="t('incidents.communicateImpactAndProgressWithClearConsistentUpdates')"
    ><button v-if="canEdit()" class="button primary" @click="create">
      <Plus :size="15" />{{ t('incidents.createIncident') }}
    </button></PageHeader
  >
  <section class="card">
    <div class="filter-bar">
      <h2 un-text="sm">{{ t('incidents.incidentAnnouncements') }}</h2>
      <select v-model="filter" un-w="auto!" :aria-label="t('incidents.incidentFilter')">
        <option value="active">{{ t('incidents.active') }}</option>
        <option value="resolved">{{ t('incidents.resolved') }}</option>
        <option value="all">{{ t('incidents.all') }}</option>
      </select>
    </div>
    <AsyncState :pending="query.isPending.value" :error="query.error.value" @retry="query.refetch()"
      ><EmptyState
        v-if="!items.length"
        :title="t('incidents.noIncidentsHere')"
        :description="t('incidents.manualIncidentsDoNotChangeMonitorStatesOr')"
      />
      <div v-else class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th>{{ t('incidents.incident') }}</th>
              <th>{{ t('incidents.progress') }}</th>
              <th>{{ t('incidents.impact') }}</th>
              <th>{{ t('common.updated') }}</th>
              <th />
            </tr>
          </thead>
          <tbody>
            <tr v-for="incident in items" :key="incident.id">
              <td>
                <button class="button ghost" un-p="0!" @click="detail(incident)">
                  <MessageSquare :size="15" />{{ incident.title }}</button
                ><span class="monitor-sub">{{
                  t('counts.pages', { count: incident.pageIds.length }, incident.pageIds.length)
                }}</span>
              </td>
              <td>
                <span class="pill">{{ statusLabel(incident.status) }}</span>
              </td>
              <td>{{ statusLabel(incident.impact) }}</td>
              <td class="muted" un-text="10px">{{ formatDate(incident.updatedAt) }}</td>
              <td>
                <button class="button small" @click="detail(incident)">
                  {{ t('incidents.viewUpdates') }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div></AsyncState
    >
  </section>
  <p class="note" un-mt="5">
    {{ t('incidents.activeIncidentsMayRaiseAPageSImpact') }}
  </p>
  <Modal
    v-model:open="open"
    :title="form.id ? t('incidents.editIncident') : t('incidents.createIncident2')"
    wide
    ><form id="incident-form" @submit.prevent="save">
      <div class="form-grid">
        <Field class="span-full" :label="t('incidents.title')"
          ><input v-model="form.title" required /></Field
        ><Field class="span-full" :label="t('incidents.description')">
          <textarea v-model="form.body" rows="5" required /></Field
        ><Field :label="t('incidents.progressStatus')"
          ><select v-model="form.status">
            <option
              v-for="state in ['investigating', 'identified', 'monitoring', 'resolved']"
              :key="state"
              :value="state"
            >
              {{ statusLabel(state) }}
            </option>
          </select></Field
        ><Field :label="t('incidents.impactLevel')"
          ><select v-model="form.impact">
            <option value="none">{{ t('incidents.informational') }}</option>
            <option value="partial">{{ t('incidents.partialOutage') }}</option>
            <option value="outage">{{ t('incidents.majorOutage') }}</option>
          </select></Field
        >
        <div class="span-full">
          <label class="field-label">{{ t('incidents.publishToStatusPages') }}</label>
          <div class="checkbox-group" un-mt="3">
            <label v-for="page in pages.data.value?.items" :key="page.id" class="checkbox-label"
              ><input v-model="form.pageIds" type="checkbox" :value="page.id" />{{
                page.name
              }}</label
            >
          </div>
        </div>
        <div class="span-full">
          <label class="field-label">{{ t('incidents.relatedMonitors') }}</label>
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
      ><button class="button" @click="open = false">{{ t('common.cancel') }}</button
      ><button class="button primary" form="incident-form" :disabled="saving">
        {{ t('incidents.saveIncident') }}
      </button></template
    ></Modal
  ><Modal v-model:open="detailOpen" :title="selected?.title || ''" wide
    ><template v-if="selected"
      ><div un-flex="~ items-start justify-between gap-4" un-mb="5">
        <p class="muted">{{ selected.body }}</p>
        <button v-if="canEdit()" class="button small" @click="edit(selected)">
          {{ t('incidents.editIncident2') }}
        </button>
      </div>
      <div class="timeline">
        <div class="timeline-entry">
          <h3>{{ t('incidents.initialAnnouncement') }}</h3>
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
        <h3 un-mb="4">{{ t('incidents.publishAnUpdate') }}</h3>
        <Field :label="t('incidents.progress2')"
          ><select v-model="updateStatus">
            <option
              v-for="state in ['investigating', 'identified', 'monitoring', 'resolved']"
              :key="state"
              :value="state"
            >
              {{ statusLabel(state) }}
            </option>
          </select></Field
        ><Field :label="t('incidents.update')" un-mt="4">
          <textarea v-model="updateBody" required rows="4" /></Field
        ><button class="button primary" :disabled="saving" un-mt="4">
          {{ t('incidents.publishUpdate') }}
        </button>
      </form></template
    ></Modal
  >
</template>
