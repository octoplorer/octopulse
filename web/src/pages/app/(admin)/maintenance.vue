<script setup lang="ts">
import type { Maintenance } from '../../../client/types.gen'
import { useMutation, useQuery } from '@pinia/colada'
import { useForm } from '@tanstack/vue-form'
import { useIntervalFn, useNow } from '@vueuse/core'
import { computed, ref, shallowRef } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  createMaintenanceMutation,
  deleteMaintenanceMutation,
  listMaintenanceQuery,
  listMonitorsQuery,
  listPagesQuery,
  updateMaintenanceMutation,
} from '../../../client/@pinia/colada.gen'
import { Badge } from '../../../components/ui/badge'
import { Banner } from '../../../components/ui/banner'
import { PageHeader } from '../../../components/ui/blocks/page-header'
import { Button } from '../../../components/ui/button'
import { Checkbox, CheckboxGroup } from '../../../components/ui/checkbox'
import { Dialog } from '../../../components/ui/dialog'
import { Empty } from '../../../components/ui/empty'
import { Field, FieldGroup } from '../../../components/ui/field'
import { Input, InputArea } from '../../../components/ui/input'
import { InputGroup, InputGroupAddon, InputGroupInput } from '../../../components/ui/input-group'
import { LayerCard } from '../../../components/ui/layer-card'
import { Loader } from '../../../components/ui/loader'
import { Table, TableBody, TableCell, TableContainer, TableHead, TableHeader, TableRow, TableToolbar } from '../../../components/ui/table'
import { TabsContent, TabsList, TabsRoot, TabsTrigger } from '../../../components/ui/tabs'
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
const deleting = ref(false)
const search = shallowRef('')
const filter = shallowRef('all')
const now = useNow({ scheduler: callback => useIntervalFn(callback, 30000) })
const allItems = computed(() => query.data.value?.items || [])
function maintenanceState(window: Maintenance) {
  return window.endsAt <= now.value.getTime()
    ? 'completed'
    : window.startsAt > now.value.getTime()
      ? 'scheduled'
      : 'inProgress'
}
const statusTabs = computed(() => [
  { value: 'all', label: t('maintenance.all'), count: allItems.value.length },
  ...(['inProgress', 'scheduled', 'completed'] as const).map(value => ({ value, label: t(`maintenance.${value}`), count: allItems.value.filter(window => maintenanceState(window) === value).length })),
])
const items = computed(() => {
  const text = search.value.trim().toLowerCase()
  const order = { inProgress: 0, scheduled: 1, completed: 2 }
  return allItems.value
    .filter(window => (filter.value === 'all' || maintenanceState(window) === filter.value)
      && (!text || `${window.name} ${window.description}`.toLowerCase().includes(text)))
    .sort((a, b) => order[maintenanceState(a)] - order[maintenanceState(b)]
      || (maintenanceState(a) === 'completed' ? b.startsAt - a.startsAt : a.startsAt - b.startsAt))
})
function clearFilters() {
  search.value = ''
  filter.value = 'all'
}
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
  if (!deleteTarget.value || deleting.value)
    return
  deleting.value = true
  try {
    await deleteMaintenance.mutateAsync({ path: { id: deleteTarget.value.id } })
    deleteOpen.value = false
    notify(t('maintenance.maintenanceDeleted'))
    await query.refresh()
  }
  catch (e) {
    notify(errorText(e), 'error')
  }
  finally {
    deleting.value = false
  }
}
function confirmDelete(value: Maintenance) {
  deleteTarget.value = value
  deleteOpen.value = true
}
</script>

<template>
  <PageHeader class="mb-6" :title="t('navigation.maintenance')" :description="t('maintenance.plannedWorkKeepsCollectionRunningExcludesDurationAnd')">
    <template #actions>
      <Button v-if="canEdit()" variant="primary" @click="edit()">
        <span w="15px" h="15px" aria-hidden="true" class="i-lucide-plus" />{{ t('common.scheduleMaintenance') }}
      </Button>
    </template>
  </PageHeader>
  <LayerCard>
    <TabsRoot v-model="filter">
      <TabsList variant="line" class="[@container_workspace_(max-width:_700px)]:px-4!" :aria-label="t('common.status')">
        <TabsTrigger v-for="status in statusTabs" :key="status.value" variant="line" :value="status.value">
          {{ status.label }}<span class="ms-2 text-size-xs text-subtle">{{ status.count }}</span>
        </TabsTrigger>
      </TabsList>
      <TabsContent :value="filter">
        <TableToolbar>
          <InputGroup class="w-full max-w-sm [@container_workspace_(max-width:_700px)]:max-w-none">
            <InputGroupAddon><span class="i-lucide-search size-4" aria-hidden="true" /></InputGroupAddon>
            <InputGroupInput v-model="search" type="search" :placeholder="t('maintenance.searchNameOrDescription')" :aria-label="t('maintenance.searchMaintenance')" />
          </InputGroup>
          <span class="text-size-sm text-subtle">{{ t('maintenance.showingMaintenance', { shown: items.length, total: allItems.length }) }}</span>
        </TableToolbar>
        <div v-if="query.isPending.value" class="loading-state" flex="~ justify-center items-center gap-10px" p="60px" un-text="12px subtle" role="status">
          <Loader :label="t('asyncState.loadingData')" />{{ t('asyncState.loadingData') }}
        </div>
        <Banner v-else-if="query.error.value" variant="error">
          {{ errorText(query.error.value) }}
          <Button variant="ghost" @click="query.refetch()">
            {{ t('asyncState.retry') }}
          </Button>
        </Banner>
        <template v-else>
          <Empty v-if="!items.length" :title="allItems.length ? t('maintenance.noMatchingMaintenance') : t('maintenance.noScheduledMaintenance')" :description="allItems.length ? t('monitors.tryChangingYourSearchOrFilters') : t('maintenance.planAnUpgradeWindowAndInformStatusPage')" size="sm" class="rounded-none border-none">
            <template #icon>
              <span class="i-lucide-calendar-clock size-8 text-subtle" aria-hidden="true" />
            </template>
            <template #actions>
              <Button v-if="allItems.length" @click="clearFilters">
                {{ t('common.clearFilters') }}
              </Button>
              <Button v-else-if="canEdit()" variant="primary" @click="edit()">
                {{ t('common.scheduleMaintenance') }}
              </Button>
            </template>
          </Empty>
          <TableContainer v-else :scroll-label="t('common.scrollTable')">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{{ t('maintenance.maintenance') }}</TableHead>
                  <TableHead class="[@container_workspace_(max-width:_700px)]:hidden">
                    {{ t('common.status') }}
                  </TableHead>
                  <TableHead class="[@container_workspace_(max-width:_700px)]:hidden">
                    {{ t('maintenance.window') }}
                  </TableHead>
                  <TableHead class="[@container_workspace_(max-width:_700px)]:hidden">
                    {{ t('maintenance.scope') }}
                  </TableHead>
                  <TableHead class="w-24 [@container_workspace_(max-width:_700px)]:w-20">
                    <span class="sr-only">{{ t('common.edit') }}</span>
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="window in items" :key="window.id">
                  <TableCell>
                    <span class="block max-w-sm font-medium [overflow-wrap:anywhere]">{{ window.name }}</span><span v-if="window.description" class="mt-1 block max-w-sm text-size-xs text-subtle [overflow-wrap:anywhere]">{{ window.description }}</span>
                    <div class="mt-2 hidden space-y-2 [@container_workspace_(max-width:_700px)]:block">
                      <Badge :variant="maintenanceState(window) === 'inProgress' ? 'warning' : 'outline'" dot>
                        {{ t(`maintenance.${maintenanceState(window)}`) }}
                      </Badge>
                      <div class="text-size-xs text-subtle">
                        {{ formatDate(window.startsAt) }} → {{ formatDate(window.endsAt) }}
                      </div>
                      <div class="text-size-xs text-subtle">
                        {{ t('counts.monitors', { count: window.monitorIds.length }, window.monitorIds.length) }} · {{ t('counts.pages', { count: window.pageIds.length }, window.pageIds.length) }}
                      </div>
                    </div>
                  </TableCell>
                  <TableCell class="[@container_workspace_(max-width:_700px)]:hidden">
                    <Badge :variant="maintenanceState(window) === 'inProgress' ? 'warning' : 'outline'" dot>
                      {{ t(`maintenance.${maintenanceState(window)}`) }}
                    </Badge>
                  </TableCell>
                  <TableCell class="text-size-sm [@container_workspace_(max-width:_700px)]:hidden">
                    <div class="whitespace-nowrap">
                      {{ formatDate(window.startsAt) }} <span class="text-subtle">→</span>
                    </div>
                    <div class="mt-1 whitespace-nowrap text-subtle">
                      {{ formatDate(window.endsAt) }}
                    </div>
                  </TableCell>
                  <TableCell class="[@container_workspace_(max-width:_700px)]:hidden">
                    {{
                      t(
                        'counts.monitors',
                        { count: window.monitorIds.length },
                        window.monitorIds.length,
                      )
                    }}
                    <span class="mt-1 block text-size-xs text-subtle">{{ t('counts.pages', { count: window.pageIds.length }, window.pageIds.length) }}</span>
                  </TableCell>
                  <TableCell>
                    <div v-if="canEdit()" class="flex justify-end gap-2">
                      <Button :aria-label="t('common.edit')" shape="square" size="sm" @click="edit(window)">
                        <span w="14px" h="14px" aria-hidden="true" class="i-lucide-pencil" />
                      </Button><Button :aria-label="t('common.delete')" shape="square" size="sm" @click="confirmDelete(window)">
                        <span w="14px" h="14px" aria-hidden="true" class="i-lucide-trash-2" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </TableContainer>
        </template>
      </TabsContent>
    </TabsRoot>
  </LayerCard>
  <Dialog v-model:open="open" :close-label="t('common.close')" :title="editingId ? t('maintenance.editMaintenance') : t('common.scheduleMaintenance')" size="xl">
    <form id="maintenance-form" @submit.prevent="form.handleSubmit">
      <FieldGroup>
        <form.Field v-slot="{ field }" name="name">
          <Field :label="t('common.name')" class="span-full">
            <Input :model-value="field.state.value" required @update:model-value="field.handleChange(String($event ?? ''))" @blur="field.handleBlur" />
          </Field>
        </form.Field><form.Field v-slot="{ field }" name="description">
          <Field :label="t('maintenance.description')" class="span-full">
            <InputArea :model-value="field.state.value" @update:model-value="field.handleChange($event ?? '')" @blur="field.handleBlur" />
          </Field>
        </form.Field><form.Field v-slot="{ field }" name="start">
          <Field :label="t('maintenance.startsAt')">
            <Input :model-value="field.state.value" type="datetime-local" required @update:model-value="field.handleChange(String($event ?? ''))" @blur="field.handleBlur" />
          </Field>
        </form.Field><form.Field v-slot="{ field }" name="end">
          <Field :label="t('maintenance.endsAt')">
            <Input :model-value="field.state.value" type="datetime-local" required @update:model-value="field.handleChange(String($event ?? ''))" @blur="field.handleBlur" />
          </Field>
        </form.Field><form.Field v-slot="{ field }" name="timezone">
          <Field :label="t('maintenance.windowTimeZone')" :description="t('maintenance.theInputsAboveAreInterpretedAsWallClock')" class="span-full">
            <Input :model-value="field.state.value" required placeholder="Asia/Shanghai" @update:model-value="field.handleChange(String($event ?? ''))" @blur="field.handleBlur" />
          </Field>
        </form.Field>
        <form.Field v-slot="{ field }" name="monitorIds">
          <div class="span-full">
            <div v-if="monitors.isPending.value" class="loading-state" flex="~ justify-center items-center gap-10px" p="60px" un-text="12px subtle" role="status">
              <Loader :label="t('asyncState.loadingData')" />{{ t('asyncState.loadingData') }}
            </div>
            <Banner v-else-if="monitors.error.value" variant="error">
              {{ errorText(monitors.error.value) }}
              <Button variant="ghost" @click="monitors.refetch()">
                {{ t('asyncState.retry') }}
              </Button>
            </Banner>
            <template v-else>
              <CheckboxGroup :model-value="field.state.value" :label="t('maintenance.affectedMonitors')" orientation="horizontal" @update:model-value="field.handleChange" @focusout="field.handleBlur">
                <Checkbox v-for="monitor in monitors.data.value?.items" :key="monitor.id" :value="monitor.id" :label="monitor.name" />
              </CheckboxGroup>
              <p v-if="!monitors.data.value?.items.length" class="mt-2 text-size-sm text-subtle">
                {{ t('counts.monitors', { count: 0 }, 0) }}
              </p>
            </template>
          </div>
        </form.Field>
        <form.Field v-slot="{ field }" name="pageIds">
          <div class="span-full">
            <div v-if="pages.isPending.value" class="loading-state" flex="~ justify-center items-center gap-10px" p="60px" un-text="12px subtle" role="status">
              <Loader :label="t('asyncState.loadingData')" />{{ t('asyncState.loadingData') }}
            </div>
            <Banner v-else-if="pages.error.value" variant="error">
              {{ errorText(pages.error.value) }}
              <Button variant="ghost" @click="pages.refetch()">
                {{ t('asyncState.retry') }}
              </Button>
            </Banner>
            <template v-else>
              <CheckboxGroup :model-value="field.state.value" :label="t('maintenance.showOnStatusPages')" orientation="horizontal" @update:model-value="field.handleChange" @focusout="field.handleBlur">
                <Checkbox v-for="page in pages.data.value?.items" :key="page.id" :value="page.id" :label="page.name" />
              </CheckboxGroup>
              <p v-if="!pages.data.value?.items.length" class="mt-2 text-size-sm text-subtle">
                {{ t('counts.pages', { count: 0 }, 0) }}
              </p>
            </template>
          </div>
        </form.Field>
      </FieldGroup>
      <Banner v-if="error" variant="error" class="mt-4">
        {{ error }}
      </Banner>
    </form>
    <template #footer>
      <Button :disabled="saving" @click="open = false">
        {{ t('common.cancel') }}
      </Button><Button type="submit" form="maintenance-form" :loading="saving" variant="primary">
        {{ t('maintenance.saveSchedule') }}
      </Button>
    </template>
  </Dialog><Dialog v-model:open="deleteOpen" :close-label="t('common.close')" :title="t('maintenance.deleteMaintenance')">
    <p>{{ deleteTarget?.name }}</p>
    <template #footer>
      <Button :disabled="deleting" @click="deleteOpen = false">
        {{ t('common.cancel') }}
      </Button><Button variant="destructive" :loading="deleting" @click="remove">
        {{ t('common.delete') }}
      </Button>
    </template>
  </Dialog>
</template>
