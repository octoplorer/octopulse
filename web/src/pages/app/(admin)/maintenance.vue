<script setup lang="ts">
import type { Maintenance } from '../../../client/types.gen'
import { useMutation, useQuery } from '@pinia/colada'
import { useForm } from '@tanstack/vue-form'
import { ref } from 'vue'
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
import { Badge } from '../../../components/ui/badge'
import { Button } from '../../../components/ui/button'
import { Card } from '../../../components/ui/card'
import { FieldError, FieldGroup, FieldInput, FieldLabel, FieldTextarea } from '../../../components/ui/field'
import { Table, TableBody, TableCell, TableContainer, TableHead, TableHeader, TableRow } from '../../../components/ui/table'
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

const query = useQuery({ ...listMaintenanceQuery(), staleTime: 10000 })
const monitors = useQuery({ ...listMonitorsQuery(), staleTime: 10000 })
const pages = useQuery({ ...listPagesQuery(), staleTime: 10000 })
const open = ref(false)
const error = ref('')
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
function formValues(value: Maintenance) {
  return {
    ...value,
    start: datetimeInput(value.startsAt, value.timezone),
    end: datetimeInput(value.endsAt, value.timezone),
  }
}
const form = useForm({
  defaultValues: formValues(empty()),
  onSubmit: async ({ value }) => {
    error.value = ''
    try {
      const { start, end, ...maintenance } = value
      const body = {
        ...maintenance,
        startsAt: datetimeMilliseconds(start, value.timezone),
        endsAt: datetimeMilliseconds(end, value.timezone),
      }
      if (value.id)
        await updateMaintenance.mutateAsync({ path: { id: value.id }, body })
      else await createMaintenance.mutateAsync({ body })
      open.value = false
      notify(t('maintenance.maintenanceSaved'))
      await query.refresh()
    }
    catch (e) {
      error.value = errorText(e)
    }
  },
})
const editingId = form.useSelector(state => state.values.id)
const saving = form.useSelector(state => state.isSubmitting)
function edit(value?: Maintenance) {
  if (form.state.isSubmitting)
    return
  form.reset(formValues(value ? clone(value) : empty()))
  error.value = ''
  open.value = true
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
  <PageHeader :title="t('navigation.maintenance')" :description="t('maintenance.plannedWorkKeepsCollectionRunningExcludesDurationAnd')">
    <Button v-if="canEdit()" variant="primary" @click="edit()">
      <span w="15px" h="15px" aria-hidden="true" class="i-lucide-plus" />{{ t('common.scheduleMaintenance') }}
    </Button>
  </PageHeader>
  <Card as="section">
    <AsyncState :pending="query.isPending.value" :error="query.error.value" @retry="query.refetch()">
      <EmptyState v-if="!query.data.value?.items.length" :title="t('maintenance.noScheduledMaintenance')" :description="t('maintenance.planAnUpgradeWindowAndInformStatusPage')" />
      <TableContainer v-else>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{{ t('maintenance.maintenance') }}</TableHead>
              <TableHead>{{ t('maintenance.window') }}</TableHead>
              <TableHead>{{ t('maintenance.scope') }}</TableHead>
              <TableHead>{{ t('common.status') }}</TableHead>
              <TableHead />
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="window in query.data.value.items" :key="window.id">
              <TableCell>
                <span flex="~ items-center gap-2" class="monitor-name block" font="600" un-text="13px"><span w="15px" h="15px" aria-hidden="true" class="i-lucide-calendar-clock" />{{ window.name }}</span><span class="monitor-sub block [overflow-wrap:anywhere]" un-text="12px subtle" mt="3px" max-w="300px">{{ window.description }}</span>
              </TableCell>
              <TableCell class="muted" un-text="13px subtle">
                {{ formatDate(window.startsAt) }}<br>{{ formatDate(window.endsAt) }}
              </TableCell>
              <TableCell>
                {{
                  t(
                    'counts.monitors',
                    { count: window.monitorIds.length },
                    window.monitorIds.length,
                  )
                }}
              </TableCell>
              <TableCell>
                <Badge>{{ maintenanceStatus(window) }}</Badge>
              </TableCell>
              <TableCell>
                <div v-if="canEdit()" flex="~ gap-1">
                  <Button :aria-label="t('common.edit')" shape="square" @click="edit(window)">
                    <span w="14px" h="14px" aria-hidden="true" class="i-lucide-pencil" />
                  </Button><Button :aria-label="t('common.delete')" shape="square" @click="confirmDelete(window)">
                    <span w="14px" h="14px" aria-hidden="true" class="i-lucide-trash-2" />
                  </Button>
                </div>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </TableContainer>
    </AsyncState>
  </Card>
  <Modal v-model:open="open" :title="editingId ? t('maintenance.editMaintenance') : t('common.scheduleMaintenance')" wide>
    <form id="maintenance-form" @submit.prevent="form.handleSubmit">
      <FieldGroup>
        <form.Field v-slot="{ field }" name="name">
          <Field :label="t('common.name')" class="span-full">
            <FieldInput :model-value="field.state.value" required @update:model-value="field.handleChange($event)" @blur="field.handleBlur" />
          </Field>
        </form.Field><form.Field v-slot="{ field }" name="description">
          <Field :label="t('maintenance.description')" class="span-full">
            <FieldTextarea :model-value="field.state.value" @update:model-value="field.handleChange($event)" @blur="field.handleBlur" />
          </Field>
        </form.Field><form.Field v-slot="{ field }" name="start">
          <Field :label="t('maintenance.startsAt')">
            <FieldInput :model-value="field.state.value" type="datetime-local" required @update:model-value="field.handleChange($event)" @blur="field.handleBlur" />
          </Field>
        </form.Field><form.Field v-slot="{ field }" name="end">
          <Field :label="t('maintenance.endsAt')">
            <FieldInput :model-value="field.state.value" type="datetime-local" required @update:model-value="field.handleChange($event)" @blur="field.handleBlur" />
          </Field>
        </form.Field><form.Field v-slot="{ field }" name="timezone">
          <Field :label="t('maintenance.windowTimeZone')" :hint="t('maintenance.theInputsAboveAreInterpretedAsWallClock')" class="span-full">
            <FieldInput :model-value="field.state.value" required placeholder="Asia/Shanghai" @update:model-value="field.handleChange($event)" @blur="field.handleBlur" />
          </Field>
        </form.Field>
        <form.Field v-slot="{ field }" name="monitorIds">
          <div class="span-full">
            <FieldLabel as="label">
              {{ t('maintenance.affectedMonitors') }}
            </FieldLabel>
            <div mt="3" class="checkbox-group" flex="~ wrap" gap="12px">
              <label v-for="monitor in monitors.data.value?.items" :key="monitor.id" class="checkbox-label" flex="~ items-center" gap="8px" un-text="12px default"><input :checked="field.state.value.includes(monitor.id)" type="checkbox" :value="monitor.id" @change="field.handleChange(($event.target as HTMLInputElement).checked ? [...field.state.value, monitor.id] : field.state.value.filter(id => id !== monitor.id))" @blur="field.handleBlur">{{
                monitor.name
              }}</label>
            </div>
          </div>
        </form.Field>
        <form.Field v-slot="{ field }" name="pageIds">
          <div class="span-full">
            <FieldLabel as="label">
              {{ t('maintenance.showOnStatusPages') }}
            </FieldLabel>
            <div mt="3" class="checkbox-group" flex="~ wrap" gap="12px">
              <label v-for="page in pages.data.value?.items" :key="page.id" class="checkbox-label" flex="~ items-center" gap="8px" un-text="12px default"><input :checked="field.state.value.includes(page.id)" type="checkbox" :value="page.id" @change="field.handleChange(($event.target as HTMLInputElement).checked ? [...field.state.value, page.id] : field.state.value.filter(id => id !== page.id))" @blur="field.handleBlur">{{
                page.name
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
      </Button><Button form="maintenance-form" :disabled="saving" variant="primary">
        {{ t('maintenance.saveSchedule') }}
      </Button>
    </template>
  </Modal><Modal v-model:open="deleteOpen" :title="t('maintenance.deleteMaintenance')">
    <p>{{ deleteTarget?.name }}</p>
    <template #footer>
      <Button @click="deleteOpen = false">
        {{ t('common.cancel') }}
      </Button><Button variant="destructive" @click="remove">
        {{ t('common.delete') }}
      </Button>
    </template>
  </Modal>
</template>
