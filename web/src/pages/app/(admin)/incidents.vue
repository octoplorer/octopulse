<script setup lang="ts">
import type { Incident } from '../../../client/types.gen'
import { useMutation, useQuery } from '@pinia/colada'
import { useForm } from '@tanstack/vue-form'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  createIncidentsMutation,
  createIncidentUpdateMutation,
  listIncidentsQuery,
  listMonitorsQuery,
  listPagesQuery,
  updateIncidentsMutation,
} from '../../../client/@pinia/colada.gen'
import AsyncState from '../../../components/AsyncState.vue'
import EmptyState from '../../../components/EmptyState.vue'
import Field from '../../../components/Field.vue'
import Modal from '../../../components/Modal.vue'
import PageHeader from '../../../components/PageHeader.vue'
import { Alert } from '../../../components/ui/alert'
import { Badge } from '../../../components/ui/badge'
import { Button } from '../../../components/ui/button'
import { Card } from '../../../components/ui/card'
import { FieldError, FieldGroup, FieldInput, FieldLabel, FieldTextarea } from '../../../components/ui/field'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '../../../components/ui/select'
import { Separator } from '../../../components/ui/separator'
import { Table, TableBody, TableCell, TableContainer, TableHead, TableHeader, TableRow, TableToolbar } from '../../../components/ui/table'
import { canEdit } from '../../../composables/api'
import { notify } from '../../../composables/notices'
import { formatDate, statusLabel } from '../../../composables/preferences'
import { errorText } from '../../../lib/errors'
import { clone } from '../../../lib/form'

const { t } = useI18n({ useScope: 'global' })

definePage({ meta: { title: 'navigation.incidents' } })

const createIncident = useMutation(createIncidentsMutation())
const updateIncident = useMutation(updateIncidentsMutation())
const publishIncidentUpdate = useMutation(createIncidentUpdateMutation())

const query = useQuery({ ...listIncidentsQuery(), staleTime: 10000 })
const pages = useQuery({ ...listPagesQuery(), staleTime: 10000 })
const monitors = useQuery({ ...listMonitorsQuery(), staleTime: 10000 })
const open = ref(false)
const detailOpen = ref(false)
const selected = ref<Incident | null>(null)
const error = ref('')
const filter = ref('active')
function empty(): Incident {
  return {
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
  }
}
const form = useForm({
  defaultValues: empty(),
  onSubmit: async ({ value }) => {
    error.value = ''
    try {
      if (value.id)
        await updateIncident.mutateAsync({ path: { id: value.id }, body: value })
      else await createIncident.mutateAsync({ body: value })
      open.value = false
      notify(t('incidents.incidentSaved'))
      await query.refresh()
    }
    catch (e) {
      error.value = errorText(e)
    }
  },
})
const editingId = form.useSelector(state => state.values.id)
const saving = form.useSelector(state => state.isSubmitting)
const updateForm = useForm({
  defaultValues: { body: '', status: 'investigating' as Incident['status'] },
  onSubmit: async ({ value }) => {
    if (!selected.value)
      return
    try {
      const result = await publishIncidentUpdate.mutateAsync({
        path: { id: selected.value.id },
        body: value,
      })
      selected.value = result
      notify(t('incidents.updatePublished'))
      await query.refresh()
      updateForm.reset({ body: '', status: result.status })
    }
    catch (e) {
      notify(errorText(e), 'error')
    }
  },
})
const publishing = updateForm.useSelector(state => state.isSubmitting)
const items = computed(
  () =>
    query.data.value?.items
      .filter(
        x =>
          filter.value === 'all'
          || (filter.value === 'active' ? x.status !== 'resolved' : x.status === 'resolved'),
      )
      .sort((a, b) => b.updatedAt - a.updatedAt) || [],
)
function edit(incident: Incident) {
  if (form.state.isSubmitting || updateForm.state.isSubmitting)
    return
  form.reset(clone(incident))
  error.value = ''
  detailOpen.value = false
  open.value = true
}
function create() {
  if (form.state.isSubmitting || updateForm.state.isSubmitting)
    return
  form.reset(empty())
  error.value = ''
  open.value = true
}
function detail(incident: Incident) {
  if (form.state.isSubmitting || updateForm.state.isSubmitting)
    return
  selected.value = clone(incident)
  updateForm.reset({ body: '', status: incident.status })
  detailOpen.value = true
}
</script>

<template>
  <PageHeader :title="t('navigation.incidents')" :description="t('incidents.communicateImpactAndProgressWithClearConsistentUpdates')">
    <Button v-if="canEdit()" variant="primary" @click="create">
      <span w="15px" h="15px" aria-hidden="true" class="i-lucide-plus" />{{ t('incidents.createIncident') }}
    </Button>
  </PageHeader>
  <Card as="section">
    <TableToolbar>
      <h2 un-text="sm">
        {{ t('incidents.incidentAnnouncements') }}
      </h2>
      <Select v-model="filter">
        <SelectTrigger w="auto!" :aria-label="t('incidents.incidentFilter')">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectGroup>
            <SelectItem value="active">
              {{ t('incidents.active') }}
            </SelectItem>
            <SelectItem value="resolved">
              {{ t('incidents.resolved') }}
            </SelectItem>
            <SelectItem value="all">
              {{ t('incidents.all') }}
            </SelectItem>
          </SelectGroup>
        </SelectContent>
      </Select>
    </TableToolbar>
    <AsyncState :pending="query.isPending.value" :error="query.error.value" @retry="query.refetch()">
      <EmptyState v-if="!items.length" :title="t('incidents.noIncidentsHere')" :description="t('incidents.manualIncidentsDoNotChangeMonitorStatesOr')" />
      <TableContainer v-else>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{{ t('incidents.incident') }}</TableHead>
              <TableHead>{{ t('incidents.progress') }}</TableHead>
              <TableHead>{{ t('incidents.impact') }}</TableHead>
              <TableHead>{{ t('common.updated') }}</TableHead>
              <TableHead />
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="incident in items" :key="incident.id">
              <TableCell>
                <Button p="0!" variant="ghost" @click="detail(incident)">
                  <span w="15px" h="15px" aria-hidden="true" class="i-lucide-message-square" />{{ incident.title }}
                </Button><span class="monitor-sub block [overflow-wrap:anywhere]" un-text="12px subtle" mt="3px" max-w="300px">{{
                  t('counts.pages', { count: incident.pageIds.length }, incident.pageIds.length)
                }}</span>
              </TableCell>
              <TableCell>
                <Badge>{{ statusLabel(incident.status) }}</Badge>
              </TableCell>
              <TableCell>{{ statusLabel(incident.impact) }}</TableCell>
              <TableCell class="muted" un-text="13px subtle">
                {{ formatDate(incident.updatedAt) }}
              </TableCell>
              <TableCell>
                <Button size="sm" @click="detail(incident)">
                  {{ t('incidents.viewUpdates') }}
                </Button>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </TableContainer>
    </AsyncState>
  </Card>
  <Alert mt="5" as="p" variant="default">
    {{ t('incidents.activeIncidentsMayRaiseAPageSImpact') }}
  </Alert>
  <Modal v-model:open="open" :title="editingId ? t('incidents.editIncident') : t('incidents.createIncident2')" wide>
    <form id="incident-form" @submit.prevent="form.handleSubmit">
      <FieldGroup>
        <form.Field v-slot="{ field }" name="title">
          <Field :label="t('incidents.title')" class="span-full">
            <FieldInput :model-value="field.state.value" required @update:model-value="field.handleChange($event)" @blur="field.handleBlur" />
          </Field>
        </form.Field><form.Field v-slot="{ field }" name="body">
          <Field :label="t('incidents.description')" class="span-full">
            <FieldTextarea :model-value="field.state.value" rows="5" required @update:model-value="field.handleChange($event)" @blur="field.handleBlur" />
          </Field>
        </form.Field><form.Field v-slot="{ field }" name="status">
          <Field :label="t('incidents.progressStatus')">
            <Select :model-value="field.state.value" @update:model-value="field.handleChange($event as Incident['status'])" @focusout="field.handleBlur">
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem
                    v-for="state in ['investigating', 'identified', 'monitoring', 'resolved']"
                    :key="state"
                    :value="state"
                  >
                    {{ statusLabel(state) }}
                  </SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </Field>
        </form.Field><form.Field v-slot="{ field }" name="impact">
          <Field :label="t('incidents.impactLevel')">
            <Select :model-value="field.state.value" @update:model-value="field.handleChange($event as Incident['impact'])" @focusout="field.handleBlur">
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="none">
                    {{ t('incidents.informational') }}
                  </SelectItem>
                  <SelectItem value="partial">
                    {{ t('incidents.partialOutage') }}
                  </SelectItem>
                  <SelectItem value="outage">
                    {{ t('incidents.majorOutage') }}
                  </SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </Field>
        </form.Field>
        <form.Field v-slot="{ field }" name="pageIds">
          <div class="span-full">
            <FieldLabel as="label">
              {{ t('incidents.publishToStatusPages') }}
            </FieldLabel>
            <div mt="3" class="checkbox-group" flex="~ wrap" gap="12px">
              <label v-for="page in pages.data.value?.items" :key="page.id" class="checkbox-label" flex="~ items-center" gap="8px" un-text="12px default"><input :checked="field.state.value.includes(page.id)" type="checkbox" :value="page.id" @change="field.handleChange(($event.target as HTMLInputElement).checked ? [...field.state.value, page.id] : field.state.value.filter(id => id !== page.id))" @blur="field.handleBlur">{{
                page.name
              }}</label>
            </div>
          </div>
        </form.Field>
        <form.Field v-slot="{ field }" name="monitorIds">
          <div class="span-full">
            <FieldLabel as="label">
              {{ t('incidents.relatedMonitors') }}
            </FieldLabel>
            <div mt="3" class="checkbox-group" flex="~ wrap" gap="12px">
              <label v-for="monitor in monitors.data.value?.items" :key="monitor.id" class="checkbox-label" flex="~ items-center" gap="8px" un-text="12px default"><input :checked="field.state.value.includes(monitor.id)" type="checkbox" :value="monitor.id" @change="field.handleChange(($event.target as HTMLInputElement).checked ? [...field.state.value, monitor.id] : field.state.value.filter(id => id !== monitor.id))" @blur="field.handleBlur">{{
                monitor.name
              }}</label>
            </div>
          </div>
        </form.Field>
      </FieldGroup>
      <FieldError v-if="error" as="p" py="10px" px="0">
        {{ error }}
      </FieldError>
    </form>
    <template #footer>
      <Button @click="open = false">
        {{ t('common.cancel') }}
      </Button><Button form="incident-form" :disabled="saving" variant="primary">
        {{ t('incidents.saveIncident') }}
      </Button>
    </template>
  </Modal><Modal v-model:open="detailOpen" :title="selected?.title || ''" wide>
    <template v-if="selected">
      <div flex="~ items-start justify-between gap-4" mb="5">
        <p class="muted" un-text="13px subtle">
          {{ selected.body }}
        </p>
        <Button v-if="canEdit()" size="sm" @click="edit(selected)">
          {{ t('incidents.editIncident2') }}
        </Button>
      </div>
      <div ml="5px" pl="21px" border="l-1 solid line" class="[&_.timeline-entry]:relative [&_.timeline-entry]:pb-24px [&_.timeline-entry]:before:content-empty [&_.timeline-entry]:before:absolute [&_.timeline-entry]:before:left-[-26px] [&_.timeline-entry]:before:top-5px [&_.timeline-entry]:before:size-9px [&_.timeline-entry]:before:rounded-full [&_.timeline-entry]:before:border-2 [&_.timeline-entry]:before:border-solid [&_.timeline-entry]:before:border-base [&_.timeline-entry]:before:bg-brand [&_.timeline-entry_h3]:text-12px [&_.timeline-entry_p]:mt-6px [&_.timeline-entry_p]:whitespace-pre-wrap [&_.timeline-entry_p]:text-12px [&_.timeline-entry_p]:text-subtle [&_.timeline-entry_small]:text-12px [&_.timeline-entry_small]:text-subtle">
        <div class="timeline-entry">
          <h3>{{ t('incidents.initialAnnouncement') }}</h3>
          <small>{{ formatDate(selected.createdAt) }}</small>
        </div>
        <div v-for="entry in selected.updates" :key="entry.id" class="timeline-entry">
          <h3>{{ statusLabel(entry.status) }}</h3>
          <p>{{ entry.body }}</p>
          <small>{{ formatDate(entry.createdAt) }}</small>
        </div>
      </div>
      <form v-if="canEdit()" @submit.prevent="updateForm.handleSubmit">
        <Separator />
        <h3 mb="4">
          {{ t('incidents.publishAnUpdate') }}
        </h3>
        <updateForm.Field v-slot="{ field }" name="status">
          <Field :label="t('incidents.progress2')">
            <Select :model-value="field.state.value" @update:model-value="field.handleChange($event as Incident['status'])" @focusout="field.handleBlur">
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem
                    v-for="state in ['investigating', 'identified', 'monitoring', 'resolved']"
                    :key="state"
                    :value="state"
                  >
                    {{ statusLabel(state) }}
                  </SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </Field>
        </updateForm.Field><updateForm.Field v-slot="{ field }" name="body">
          <Field :label="t('incidents.update')" mt="4">
            <FieldTextarea :model-value="field.state.value" required rows="4" @update:model-value="field.handleChange($event)" @blur="field.handleBlur" />
          </Field>
        </updateForm.Field><Button :disabled="publishing" mt="4" variant="primary">
          {{ t('incidents.publishUpdate') }}
        </Button>
      </form>
    </template>
  </Modal>
</template>
