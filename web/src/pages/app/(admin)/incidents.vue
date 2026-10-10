<script setup lang="ts">
import type { Incident } from '../../../client/types.gen'
import { useMutation, useQuery } from '@pinia/colada'
import { useForm } from '@tanstack/vue-form'
import { computed, ref, shallowRef } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  createIncidentsMutation,
  createIncidentUpdateMutation,
  listIncidentsQuery,
  listMonitorsQuery,
  listPagesQuery,
  updateIncidentsMutation,
} from '../../../client/@pinia/colada.gen'
import { Separator } from '../../../components/common/separator'
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
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '../../../components/ui/select'
import { Table, TableBody, TableCell, TableContainer, TableHead, TableHeader, TableRow, TableToolbar } from '../../../components/ui/table'
import { TabsContent, TabsList, TabsRoot, TabsTrigger } from '../../../components/ui/tabs'
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
const filter = shallowRef('active')
const search = shallowRef('')
const allItems = computed(() => query.data.value?.items || [])
const statusTabs = computed(() => [
  { value: 'active', label: t('incidents.active'), count: allItems.value.filter(incident => incident.status !== 'resolved').length },
  { value: 'resolved', label: t('incidents.resolved'), count: allItems.value.filter(incident => incident.status === 'resolved').length },
  { value: 'all', label: t('incidents.all'), count: allItems.value.length },
])
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
      notify(t('incidents.incident-saved'))
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
      notify(t('incidents.update-published'))
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
    allItems.value
      .filter(
        x =>
          (filter.value === 'all'
            || (filter.value === 'active' ? x.status !== 'resolved' : x.status === 'resolved'))
          && (!search.value.trim() || `${x.title} ${x.body}`.toLowerCase().includes(search.value.trim().toLowerCase())),
      )
      .sort((a, b) => b.updatedAt - a.updatedAt),
)
function clearFilters() {
  search.value = ''
  filter.value = 'all'
}
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
  <PageHeader class="mb-6" :title="t('navigation.incidents')" :description="t('incidents.communicate-impact-and-progress-with-clear-consistent-updates')">
    <template #actions>
      <Button v-if="canEdit()" variant="primary" @click="create">
        <span w="15px" h="15px" aria-hidden="true" class="i-lucide-plus" />{{ t('incidents.create-incident') }}
      </Button>
    </template>
  </PageHeader>
  <LayerCard>
    <TabsRoot v-model="filter">
      <TabsList variant="line" class="[@container_workspace_(max-width:_700px)]:px-4!" :aria-label="t('incidents.incident-filter')">
        <TabsTrigger v-for="status in statusTabs" :key="status.value" variant="line" :value="status.value">
          {{ status.label }}<span class="ms-2 text-size-xs text-subtle">{{ status.count }}</span>
        </TabsTrigger>
      </TabsList>
      <TabsContent :value="filter">
        <TableToolbar>
          <InputGroup class="w-full max-w-sm [@container_workspace_(max-width:_700px)]:max-w-none">
            <InputGroupAddon><span class="i-lucide-search size-4" aria-hidden="true" /></InputGroupAddon>
            <InputGroupInput v-model="search" type="search" :placeholder="t('incidents.search-title-or-content')" :aria-label="t('incidents.search-incidents')" />
          </InputGroup>
          <span class="text-size-sm text-subtle">{{ t('incidents.showing-incidents', { shown: items.length, total: allItems.length }) }}</span>
        </TableToolbar>
        <div v-if="query.isPending.value" class="loading-state" flex="~ justify-center items-center gap-10px" p="60px" un-text="12px subtle" role="status">
          <Loader :label="t('async-state.loading-data')" />{{ t('async-state.loading-data') }}
        </div>
        <Banner v-else-if="query.error.value" variant="error">
          {{ errorText(query.error.value) }}
          <Button variant="ghost" @click="query.refetch()">
            {{ t('async-state.retry') }}
          </Button>
        </Banner>
        <template v-else>
          <Empty v-if="!items.length" :title="allItems.length ? t('incidents.no-matching-incidents') : t('incidents.no-incidents-here')" :description="allItems.length ? t('monitors.try-changing-your-search-or-filters') : t('incidents.manual-incidents-do-not-change-monitor-states-or')" size="sm" class="rounded-none border-none">
            <template #icon>
              <span class="i-lucide-message-square size-8 text-subtle" aria-hidden="true" />
            </template>
            <template #actions>
              <Button v-if="allItems.length" @click="clearFilters">
                {{ t('common.clear-filters') }}
              </Button>
              <Button v-else-if="canEdit()" variant="primary" @click="create">
                {{ t('incidents.create-incident') }}
              </Button>
            </template>
          </Empty>
          <TableContainer v-else :scroll-label="t('common.scroll-table')">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{{ t('incidents.incident') }}</TableHead>
                  <TableHead class="[@container_workspace_(max-width:_700px)]:hidden">
                    {{ t('incidents.progress') }}
                  </TableHead>
                  <TableHead class="[@container_workspace_(max-width:_700px)]:hidden">
                    {{ t('incidents.impact') }}
                  </TableHead>
                  <TableHead class="[@container_workspace_(max-width:_700px)]:hidden">
                    {{ t('common.updated') }}
                  </TableHead>
                  <TableHead class="w-28 [@container_workspace_(max-width:_700px)]:w-12">
                    <span class="sr-only">{{ t('incidents.view-updates') }}</span>
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="incident in items" :key="incident.id">
                  <TableCell>
                    <button class="block max-w-sm rounded text-start font-medium text-default outline-none hover:text-brand focus-visible:ring-2 focus-visible:ring-brand [overflow-wrap:anywhere]" @click="detail(incident)">
                      {{ incident.title }}
                    </button><span class="mt-1 block max-w-sm text-size-xs text-subtle [overflow-wrap:anywhere]">{{
                      t('counts.pages', { count: incident.pageIds.length }, incident.pageIds.length)
                    }}</span>
                    <div class="mt-2 hidden space-y-2 [@container_workspace_(max-width:_700px)]:block">
                      <div class="flex flex-wrap items-center gap-2">
                        <Badge :variant="incident.status === 'resolved' ? 'success' : 'warning'" dot>
                          {{ statusLabel(incident.status) }}
                        </Badge>
                        <span class="text-size-xs text-subtle">{{ statusLabel(incident.impact) }}</span>
                      </div>
                      <div class="text-size-xs text-subtle">
                        {{ formatDate(incident.updatedAt) }}
                      </div>
                    </div>
                  </TableCell>
                  <TableCell class="[@container_workspace_(max-width:_700px)]:hidden">
                    <Badge :variant="incident.status === 'resolved' ? 'success' : 'warning'" dot>
                      {{ statusLabel(incident.status) }}
                    </Badge>
                  </TableCell>
                  <TableCell class="[@container_workspace_(max-width:_700px)]:hidden">
                    {{ statusLabel(incident.impact) }}
                  </TableCell>
                  <TableCell class="whitespace-nowrap text-size-sm text-subtle [@container_workspace_(max-width:_700px)]:hidden">
                    {{ formatDate(incident.updatedAt) }}
                  </TableCell>
                  <TableCell class="text-end">
                    <Button size="sm" variant="ghost" :aria-label="t('incidents.view-updates')" @click="detail(incident)">
                      <span class="[@container_workspace_(max-width:_700px)]:hidden">{{ t('incidents.view-updates') }}</span>
                      <span class="i-lucide-chevron-right hidden size-4 [@container_workspace_(max-width:_700px)]:block" aria-hidden="true" />
                    </Button>
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </TableContainer>
        </template>
      </TabsContent>
    </TabsRoot>
  </LayerCard>
  <p class="mt-4 text-size-sm text-subtle">
    {{ t('incidents.active-incidents-may-raise-a-page-s-impact') }}
  </p>
  <Dialog v-model:open="open" :close-label="t('common.close')" :title="editingId ? t('incidents.edit-incident') : t('incidents.create-incident-2')" size="xl">
    <form id="incident-form" @submit.prevent="form.handleSubmit">
      <FieldGroup>
        <form.Field v-slot="{ field }" name="title">
          <Field :label="t('incidents.title')" class="span-full">
            <Input :model-value="field.state.value" required @update:model-value="field.handleChange(String($event ?? ''))" @blur="field.handleBlur" />
          </Field>
        </form.Field><form.Field v-slot="{ field }" name="body">
          <Field :label="t('incidents.description')" class="span-full">
            <InputArea :model-value="field.state.value" :min-rows="5" required @update:model-value="field.handleChange($event ?? '')" @blur="field.handleBlur" />
          </Field>
        </form.Field><form.Field v-slot="{ field }" name="status">
          <Field :label="t('incidents.progress-status')">
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
          <Field :label="t('incidents.impact-level')">
            <Select :model-value="field.state.value" @update:model-value="field.handleChange($event as Incident['impact'])" @focusout="field.handleBlur">
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="none">
                    {{ t('incidents.informational') }}
                  </SelectItem>
                  <SelectItem value="partial">
                    {{ t('incidents.partial-outage') }}
                  </SelectItem>
                  <SelectItem value="outage">
                    {{ t('incidents.major-outage') }}
                  </SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </Field>
        </form.Field>
        <form.Field v-slot="{ field }" name="pageIds">
          <div class="span-full">
            <div v-if="pages.isPending.value" class="loading-state" flex="~ justify-center items-center gap-10px" p="60px" un-text="12px subtle" role="status">
              <Loader :label="t('async-state.loading-data')" />{{ t('async-state.loading-data') }}
            </div>
            <Banner v-else-if="pages.error.value" variant="error">
              {{ errorText(pages.error.value) }}
              <Button variant="ghost" @click="pages.refetch()">
                {{ t('async-state.retry') }}
              </Button>
            </Banner>
            <template v-else>
              <CheckboxGroup :model-value="field.state.value" :label="t('incidents.publish-to-status-pages')" orientation="horizontal" @update:model-value="field.handleChange" @focusout="field.handleBlur">
                <Checkbox v-for="page in pages.data.value?.items" :key="page.id" :value="page.id" :label="page.name" />
              </CheckboxGroup>
              <p v-if="!pages.data.value?.items.length" class="mt-2 text-size-sm text-subtle">
                {{ t('counts.pages', { count: 0 }, 0) }}
              </p>
            </template>
          </div>
        </form.Field>
        <form.Field v-slot="{ field }" name="monitorIds">
          <div class="span-full">
            <div v-if="monitors.isPending.value" class="loading-state" flex="~ justify-center items-center gap-10px" p="60px" un-text="12px subtle" role="status">
              <Loader :label="t('async-state.loading-data')" />{{ t('async-state.loading-data') }}
            </div>
            <Banner v-else-if="monitors.error.value" variant="error">
              {{ errorText(monitors.error.value) }}
              <Button variant="ghost" @click="monitors.refetch()">
                {{ t('async-state.retry') }}
              </Button>
            </Banner>
            <template v-else>
              <CheckboxGroup :model-value="field.state.value" :label="t('incidents.related-monitors')" orientation="horizontal" @update:model-value="field.handleChange" @focusout="field.handleBlur">
                <Checkbox v-for="monitor in monitors.data.value?.items" :key="monitor.id" :value="monitor.id" :label="monitor.name" />
              </CheckboxGroup>
              <p v-if="!monitors.data.value?.items.length" class="mt-2 text-size-sm text-subtle">
                {{ t('counts.monitors', { count: 0 }, 0) }}
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
      </Button><Button type="submit" form="incident-form" :loading="saving" variant="primary">
        {{ t('incidents.save-incident') }}
      </Button>
    </template>
  </Dialog><Dialog v-model:open="detailOpen" :close-label="t('common.close')" :title="selected?.title || ''" size="xl">
    <template v-if="selected">
      <div flex="~ items-start justify-between gap-4" mb="5">
        <p class="muted" un-text="13px subtle">
          {{ selected.body }}
        </p>
        <Button v-if="canEdit()" size="sm" @click="edit(selected)">
          {{ t('incidents.edit-incident-2') }}
        </Button>
      </div>
      <div ml="5px" pl="21px" border="l-1 solid line" class="[&_.timeline-entry]:relative [&_.timeline-entry]:pb-24px [&_.timeline-entry]:before:content-empty [&_.timeline-entry]:before:absolute [&_.timeline-entry]:before:left-[-26px] [&_.timeline-entry]:before:top-5px [&_.timeline-entry]:before:size-9px [&_.timeline-entry]:before:rounded-full [&_.timeline-entry]:before:border-2 [&_.timeline-entry]:before:border-solid [&_.timeline-entry]:before:border-base [&_.timeline-entry]:before:bg-brand [&_.timeline-entry_h3]:text-12px [&_.timeline-entry_p]:mt-6px [&_.timeline-entry_p]:whitespace-pre-wrap [&_.timeline-entry_p]:text-12px [&_.timeline-entry_p]:text-subtle [&_.timeline-entry_small]:text-12px [&_.timeline-entry_small]:text-subtle">
        <div class="timeline-entry">
          <h3>{{ t('incidents.initial-announcement') }}</h3>
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
          {{ t('incidents.publish-an-update') }}
        </h3>
        <updateForm.Field v-slot="{ field }" name="status">
          <Field :label="t('incidents.progress-2')">
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
            <InputArea :model-value="field.state.value" required :min-rows="4" @update:model-value="field.handleChange($event ?? '')" @blur="field.handleBlur" />
          </Field>
        </updateForm.Field><Button type="submit" :loading="publishing" mt="4" variant="primary">
          {{ t('incidents.publish-update') }}
        </Button>
      </form>
    </template>
  </Dialog>
</template>
