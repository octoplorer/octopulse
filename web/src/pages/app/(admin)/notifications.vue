<script setup lang="ts">
import type { Channel, DeliveryView } from '../../../client/types.gen'
import { useMutation, useQuery } from '@pinia/colada'
import { useForm } from '@tanstack/vue-form'
import { computed, ref, shallowRef } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  createChannelsMutation,
  deleteChannelsMutation,
  listChannelsQuery,
  listDeliveriesQuery,
  listMonitorsQuery,
  listSecretsQuery,
  testChannelMutation,
  updateChannelsMutation,
} from '../../../client/@pinia/colada.gen'
import { Badge } from '../../../components/ui/badge'
import { Banner } from '../../../components/ui/banner'
import { PageHeader } from '../../../components/ui/blocks/page-header'
import { Button } from '../../../components/ui/button'
import { Dialog } from '../../../components/ui/dialog'
import { Empty } from '../../../components/ui/empty'
import { Field, FieldDescription } from '../../../components/ui/field'
import { Input } from '../../../components/ui/input'
import { InputGroup, InputGroupAddon, InputGroupInput } from '../../../components/ui/input-group'
import { LayerCard } from '../../../components/ui/layer-card'
import { Loader } from '../../../components/ui/loader'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '../../../components/ui/select'
import { Switch } from '../../../components/ui/switch'
import { Table, TableBody, TableCell, TableContainer, TableHead, TableHeader, TableRow, TableToolbar } from '../../../components/ui/table'
import { TabsContent, TabsList, TabsRoot, TabsTrigger } from '../../../components/ui/tabs'
import { isAdmin } from '../../../composables/api'
import { notify } from '../../../composables/notices'
import { formatDate, statusLabel } from '../../../composables/preferences'
import { errorText } from '../../../lib/errors'
import { clone } from '../../../lib/form'

const { t } = useI18n({ useScope: 'global' })

definePage({ meta: { title: 'navigation.notifications' } })

const createChannel = useMutation(createChannelsMutation())
const updateChannel = useMutation(updateChannelsMutation())
const deleteChannel = useMutation(deleteChannelsMutation())
const testChannel = useMutation(testChannelMutation())

const query = useQuery({ ...listChannelsQuery(), staleTime: 10000 })
const secrets = useQuery({ ...listSecretsQuery(), staleTime: 10000 })
const deliveries = useQuery({ ...listDeliveriesQuery(), staleTime: 10000 })
const monitors = useQuery({ ...listMonitorsQuery(), staleTime: 10000 })
const open = ref(false)
const error = ref('')
const tab = ref('channels')
const testing = ref('')
const deleteTarget = ref<Channel | null>(null)
const deleteOpen = ref(false)
const deleting = ref(false)
const channelSearch = shallowRef('')
const deliverySearch = shallowRef('')
const channelFilter = shallowRef('all')
const statusFilter = shallowRef('all')
const failureOpen = shallowRef(false)
const selectedFailure = shallowRef<DeliveryView | null>(null)
const allChannels = computed(() => query.data.value?.items || [])
const allDeliveries = computed(() => deliveries.data.value?.items || [])
const channelNames = computed(() => new Map(allChannels.value.map(channel => [channel.id, channel.name])))
const monitorNames = computed(() => new Map(monitors.data.value?.items.map(monitor => [monitor.id, monitor.name]) || []))
const secretNames = computed(() => new Map(secrets.data.value?.items.map(secret => [secret.id, secret.name]) || []))
const filteredChannels = computed(() => {
  const text = channelSearch.value.trim().toLowerCase()
  return allChannels.value.filter(channel => !text || channel.name.toLowerCase().includes(text))
})
const deliveryStatuses = computed(() => [...new Set(['pending', 'leased', 'sent', 'failed', 'discarded', ...allDeliveries.value.map(delivery => delivery.status)])])
const deliveryChannels = computed(() => {
  const names = new Map(channelNames.value)
  allDeliveries.value.forEach((delivery) => {
    if (!names.has(delivery.channelId))
      names.set(delivery.channelId, delivery.channelId)
  })
  return [...names].map(([id, name]) => ({ id, name }))
})
const filteredDeliveries = computed(() => {
  const text = deliverySearch.value.trim().toLowerCase()
  return allDeliveries.value
    .map(delivery => ({
      ...delivery,
      channelName: channelNames.value.get(delivery.channelId) || delivery.channelId,
      monitorName: monitorNames.value.get(delivery.monitorId) || delivery.monitorId,
    }))
    .filter(delivery => (statusFilter.value === 'all' || delivery.status === statusFilter.value)
      && (channelFilter.value === 'all' || delivery.channelId === channelFilter.value)
      && (!text || `${delivery.monitorName} ${delivery.channelName} ${statusLabel(delivery.kind)} ${delivery.lastError}`.toLowerCase().includes(text)))
    .sort((a, b) => b.createdAt - a.createdAt)
})
function clearDeliveryFilters() {
  deliverySearch.value = ''
  channelFilter.value = 'all'
  statusFilter.value = 'all'
}
function showFailure(delivery: DeliveryView) {
  selectedFailure.value = delivery
  failureOpen.value = true
}
function empty(): Channel {
  return {
    id: '',
    name: '',
    serviceUrlSecretId: '',
    enabled: true,
    createdAt: 0,
    updatedAt: 0,
  }
}
const formApi = useForm({
  defaultValues: empty(),
  onSubmit: async ({ value }) => {
    error.value = ''
    try {
      if (value.id)
        await updateChannel.mutateAsync({ path: { id: value.id }, body: value })
      else await createChannel.mutateAsync({ body: value })
      open.value = false
      await query.refresh()
      notify(t('notifications.channel-saved'))
    }
    catch (e) {
      error.value = errorText(e)
    }
  },
})
const form = formApi.useSelector(state => state.values)
const saving = formApi.useSelector(state => state.isSubmitting)
function edit(channel?: Channel) {
  if (formApi.state.isSubmitting)
    return
  formApi.reset(channel ? clone(channel) : empty())
  error.value = ''
  open.value = true
}
async function test(channel: Channel) {
  if (testing.value || !channel.enabled)
    return
  testing.value = channel.id
  try {
    await testChannel.mutateAsync({ path: { id: channel.id } })
    notify(t('notifications.test-submitted-check-delivery-history'))
  }
  catch (e) {
    notify(errorText(e), 'error')
  }
  finally {
    await deliveries.refresh()
    testing.value = ''
  }
}
async function remove() {
  if (!deleteTarget.value || deleting.value)
    return
  deleting.value = true
  try {
    await deleteChannel.mutateAsync({ path: { id: deleteTarget.value.id } })
    deleteOpen.value = false
    await query.refresh()
    notify(t('notifications.channel-deleted'))
  }
  catch (e) {
    notify(errorText(e), 'error')
  }
  finally {
    deleting.value = false
  }
}
function confirmDelete(value: Channel) {
  deleteTarget.value = value
  deleteOpen.value = true
}
function refresh() {
  query.refetch()
  deliveries.refetch()
}
</script>

<template>
  <PageHeader class="mb-6" :title="t('navigation.notifications')" :description="t('notifications.connect-channels-through-shoutrrr-and-keep-every-delivery')">
    <template #actions>
      <Button :loading="query.isLoading.value || deliveries.isLoading.value" @click="refresh">
        <span w="14px" h="14px" aria-hidden="true" class="i-lucide-refresh-cw" />{{ t('common.refresh') }}
      </Button><Button v-if="isAdmin() && tab === 'channels'" variant="primary" @click="edit()">
        <span w="15px" h="15px" aria-hidden="true" class="i-lucide-plus" />{{ t('notifications.add-channel') }}
      </Button>
    </template>
  </PageHeader>
  <LayerCard>
    <TabsRoot v-model="tab">
      <TabsList variant="line" class="[@container_workspace_(max-width:_700px)]:px-4!" :aria-label="t('navigation.notifications')">
        <TabsTrigger variant="line" value="channels">
          {{
            t('notifications.channels')
          }}<span class="ms-2 text-size-xs text-subtle">{{ allChannels.length }}</span>
        </TabsTrigger><TabsTrigger variant="line" value="deliveries">
          {{
            t('notifications.deliveries')
          }}<span class="ms-2 text-size-xs text-subtle">{{ allDeliveries.length }}</span>
        </TabsTrigger>
      </TabsList><TabsContent value="channels">
        <TableToolbar>
          <InputGroup class="w-full max-w-sm [@container_workspace_(max-width:_700px)]:max-w-none">
            <InputGroupAddon><span class="i-lucide-search size-4" aria-hidden="true" /></InputGroupAddon>
            <InputGroupInput v-model="channelSearch" type="search" :placeholder="t('notifications.search-channels-placeholder')" :aria-label="t('notifications.search-channels')" />
          </InputGroup>
          <span class="text-size-sm text-subtle">{{ t('notifications.showing-channels', { shown: filteredChannels.length, total: allChannels.length }) }}</span>
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
          <Empty v-if="!filteredChannels.length" :title="allChannels.length ? t('notifications.no-matching-channels') : t('notifications.connect-a-notification-channel')" :description="allChannels.length ? t('monitors.try-changing-your-search-or-filters') : t('notifications.store-a-shoutrrr-service-url-as-a-secret')" size="sm" class="rounded-none border-none">
            <template #icon>
              <span class="i-lucide-bell size-8 text-subtle" aria-hidden="true" />
            </template>
            <template #actions>
              <Button v-if="allChannels.length" @click="channelSearch = ''">
                {{ t('common.clear-filters') }}
              </Button>
              <div v-else-if="isAdmin()" class="flex flex-wrap justify-center gap-2">
                <Button variant="primary" @click="edit()">
                  {{ t('notifications.add-channel') }}
                </Button>
                <Button as-child>
                  <RouterLink to="/app/secrets">
                    {{
                      t('notifications.manage-secrets')
                    }}
                  </RouterLink>
                </Button>
              </div>
            </template>
          </Empty>
          <TableContainer v-else :scroll-label="t('common.scroll-table')">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{{ t('common.channel') }}</TableHead>
                  <TableHead class="[@container_workspace_(max-width:_700px)]:hidden">
                    {{ t('notifications.secret-reference') }}
                  </TableHead>
                  <TableHead class="[@container_workspace_(max-width:_700px)]:hidden">
                    {{ t('common.status') }}
                  </TableHead>
                  <TableHead class="[@container_workspace_(max-width:_700px)]:hidden">
                    {{ t('common.updated') }}
                  </TableHead>
                  <TableHead class="w-36 [@container_workspace_(max-width:_700px)]:w-28">
                    <span class="sr-only">{{ t('common.edit') }}</span>
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="channel in filteredChannels" :key="channel.id">
                  <TableCell>
                    <span class="block max-w-sm font-medium [overflow-wrap:anywhere]">{{ channel.name }}</span>
                    <div class="mt-2 hidden space-y-2 [@container_workspace_(max-width:_700px)]:block">
                      <Badge :variant="channel.enabled ? 'success' : 'outline'" dot>
                        {{ channel.enabled ? t('common.enabled') : t('common.disabled') }}
                      </Badge>
                      <div class="text-size-xs text-subtle [overflow-wrap:anywhere]">
                        {{ secretNames.get(channel.serviceUrlSecretId) || '—' }}
                      </div>
                      <div class="text-size-xs text-subtle">
                        {{ formatDate(channel.updatedAt) }}
                      </div>
                    </div>
                  </TableCell>
                  <TableCell class="text-size-sm text-subtle [@container_workspace_(max-width:_700px)]:hidden">
                    {{
                      secretNames.get(channel.serviceUrlSecretId) || '—'
                    }}
                  </TableCell>
                  <TableCell class="[@container_workspace_(max-width:_700px)]:hidden">
                    <Badge :variant="channel.enabled ? 'success' : 'outline'" dot>
                      {{
                        channel.enabled ? t('common.enabled') : t('common.disabled')
                      }}
                    </Badge>
                  </TableCell>
                  <TableCell class="whitespace-nowrap text-size-sm text-subtle [@container_workspace_(max-width:_700px)]:hidden">
                    {{ formatDate(channel.updatedAt) }}
                  </TableCell>
                  <TableCell>
                    <div v-if="isAdmin()" class="flex items-center justify-end gap-2">
                      <Button :loading="testing === channel.id" :disabled="!!testing || !channel.enabled" :aria-label="t('notifications.test')" size="sm" @click="test(channel)">
                        <span w="12px" h="12px" aria-hidden="true" class="i-lucide-send" /><span class="[@container_workspace_(max-width:_700px)]:hidden">{{ t('notifications.test') }}</span>
                      </Button><Button :aria-label="t('common.edit')" shape="square" size="sm" @click="edit(channel)">
                        <span w="14px" h="14px" aria-hidden="true" class="i-lucide-pencil" />
                      </Button><Button :aria-label="t('common.delete')" shape="square" size="sm" @click="confirmDelete(channel)">
                        <span w="14px" h="14px" aria-hidden="true" class="i-lucide-trash-2" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </TableContainer>
        </template>
      </TabsContent><TabsContent value="deliveries">
        <TableToolbar>
          <InputGroup class="w-full max-w-sm [@container_workspace_(max-width:_700px)]:max-w-none">
            <InputGroupAddon><span class="i-lucide-search size-4" aria-hidden="true" /></InputGroupAddon>
            <InputGroupInput v-model="deliverySearch" type="search" :placeholder="t('notifications.search-deliveries-placeholder')" :aria-label="t('notifications.search-deliveries')" />
          </InputGroup>
          <div class="flex flex-wrap items-center gap-2">
            <Select v-model="statusFilter">
              <SelectTrigger class="w-auto!" :aria-label="t('notifications.filter-status')">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="all">
                    {{ t('notifications.all-statuses') }}
                  </SelectItem>
                  <SelectItem v-for="status in deliveryStatuses" :key="status" :value="status">
                    {{ statusLabel(status) }}
                  </SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
            <Select v-model="channelFilter">
              <SelectTrigger class="w-auto! max-w-64" :aria-label="t('notifications.filter-channel')">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="all">
                    {{ t('notifications.all-channels') }}
                  </SelectItem>
                  <SelectItem v-for="channel in deliveryChannels" :key="channel.id" :value="channel.id">
                    {{ channel.name }}
                  </SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
            <span class="text-size-sm text-subtle">{{ t('notifications.showing-deliveries', { shown: filteredDeliveries.length, total: allDeliveries.length }) }}</span>
          </div>
        </TableToolbar>
        <div v-if="deliveries.isPending.value" class="loading-state" flex="~ justify-center items-center gap-10px" p="60px" un-text="12px subtle" role="status">
          <Loader :label="t('async-state.loading-data')" />{{ t('async-state.loading-data') }}
        </div>
        <Banner v-else-if="deliveries.error.value" variant="error">
          {{ errorText(deliveries.error.value) }}
          <Button variant="ghost" @click="deliveries.refetch()">
            {{ t('async-state.retry') }}
          </Button>
        </Banner>
        <template v-else>
          <Empty v-if="!filteredDeliveries.length" :title="allDeliveries.length ? t('notifications.no-matching-deliveries') : t('notifications.no-deliveries-yet')" :description="allDeliveries.length ? t('monitors.try-changing-your-search-or-filters') : t('notifications.outages-recoveries-and-channel-tests-create-durable-delivery')" size="sm" class="rounded-none border-none">
            <template #icon>
              <span class="i-lucide-send size-8 text-subtle" aria-hidden="true" />
            </template>
            <template #actions>
              <Button v-if="allDeliveries.length" @click="clearDeliveryFilters">
                {{ t('common.clear-filters') }}
              </Button>
              <Button v-else @click="tab = 'channels'">
                {{ t('notifications.channels') }}
              </Button>
            </template>
          </Empty>
          <TableContainer v-else :scroll-label="t('common.scroll-table')">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{{ t('notifications.notification') }}</TableHead>
                  <TableHead class="[@container_workspace_(max-width:_700px)]:hidden">
                    {{ t('common.channel') }}
                  </TableHead>
                  <TableHead class="[@container_workspace_(max-width:_700px)]:hidden">
                    {{ t('notifications.status-attempts') }}
                  </TableHead>
                  <TableHead class="[@container_workspace_(max-width:_700px)]:hidden">
                    {{ t('common.created') }}
                  </TableHead>
                  <TableHead class="[@container_workspace_(max-width:_700px)]:w-12">
                    {{ t('notifications.failure-reason') }}
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="delivery in filteredDeliveries" :key="delivery.id">
                  <TableCell>
                    <span class="monitor-name block" font="600" un-text="13px">{{ statusLabel(delivery.kind) }}</span><RouterLink v-if="delivery.monitorId" :to="`/app/monitors/${delivery.monitorId}`" class="monitor-sub block [overflow-wrap:anywhere]" un-text="12px subtle" mt="3px" max-w="300px">
                      {{ delivery.monitorName }}
                    </RouterLink>
                    <div class="mt-2 hidden space-y-2 [@container_workspace_(max-width:_700px)]:block">
                      <div class="text-size-xs text-subtle [overflow-wrap:anywhere]">
                        {{ delivery.channelName }}
                      </div>
                      <div class="flex flex-wrap items-center gap-2">
                        <Badge :variant="delivery.status === 'failed' ? 'danger' : delivery.status === 'sent' ? 'success' : 'outline'" dot>
                          {{ statusLabel(delivery.status) }}
                        </Badge>
                        <span class="text-size-xs text-subtle">{{ delivery.attempts }}</span>
                      </div>
                      <div class="text-size-xs text-subtle">
                        {{ formatDate(delivery.createdAt) }}
                      </div>
                    </div>
                  </TableCell>
                  <TableCell class="[@container_workspace_(max-width:_700px)]:hidden">
                    <span class="block max-w-56 [overflow-wrap:anywhere]">{{ delivery.channelName }}</span>
                  </TableCell>
                  <TableCell class="[@container_workspace_(max-width:_700px)]:hidden">
                    <Badge :variant="delivery.status === 'failed' ? 'danger' : delivery.status === 'sent' ? 'success' : 'outline'" dot>
                      {{ statusLabel(delivery.status) }}
                    </Badge><span class="ms-2 text-size-sm text-subtle">{{ delivery.attempts }}</span>
                  </TableCell>
                  <TableCell class="whitespace-nowrap text-size-sm text-subtle [@container_workspace_(max-width:_700px)]:hidden">
                    {{ formatDate(delivery.createdAt) }}
                  </TableCell>
                  <TableCell class="text-size-sm text-subtle">
                    <button v-if="delivery.lastError" class="flex max-w-64 items-center gap-2 rounded text-start outline-none hover:text-default focus-visible:ring-2 focus-visible:ring-brand" :aria-label="t('notifications.view-failure-details')" @click="showFailure(delivery)">
                      <span class="line-clamp-2 [overflow-wrap:anywhere] [@container_workspace_(max-width:_700px)]:hidden">{{ delivery.lastError }}</span>
                      <span class="i-lucide-circle-alert hidden size-4 text-danger [@container_workspace_(max-width:_700px)]:block" aria-hidden="true" />
                      <span class="i-lucide-chevron-right size-4 shrink-0" aria-hidden="true" />
                    </button>
                    <span v-else>—</span>
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
    {{ t('notifications.failed-jobs-use-bounded-backoff-stale-outage-messages') }}
  </p>
  <Dialog v-model:open="failureOpen" :close-label="t('common.close')" :title="t('notifications.failure-reason')" size="lg">
    <template v-if="selectedFailure">
      <p class="mb-4 text-size-sm text-subtle">
        {{ channelNames.get(selectedFailure.channelId) || selectedFailure.channelId }} · {{ formatDate(selectedFailure.createdAt) }}
      </p>
      <pre class="max-h-80 overflow-auto overscroll-contain whitespace-pre-wrap rounded-lg bg-recessed p-4 text-size-sm [overflow-wrap:anywhere]">{{ selectedFailure.lastError }}</pre>
    </template>
  </Dialog>
  <Dialog v-model:open="open" :close-label="t('common.close')" :title="form.id ? t('notifications.edit-channel') : t('notifications.add-channel-2')">
    <form id="channel-form" @submit.prevent="formApi.handleSubmit()">
      <formApi.Field v-slot="{ field }" name="name">
        <Field :label="t('common.display-name')">
          <Input :name="field.name" :model-value="field.state.value" required @update:model-value="field.handleChange(String($event ?? ''))" @blur="field.handleBlur" />
        </Field>
      </formApi.Field>
      <formApi.Field v-slot="{ field }" name="serviceUrlSecretId">
        <Field :label="t('notifications.shoutrrr-service-url-secret')" mt="5">
          <Select :model-value="field.state.value" @update:model-value="field.handleChange($event || '')">
            <SelectTrigger @focusout="field.handleBlur">
              <SelectValue :placeholder="t('secret-select.choose-a-secret')" />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                <SelectItem value="">
                  {{ t('secret-select.choose-a-secret') }}
                </SelectItem>
                <SelectItem v-for="secret in secrets.data.value?.items" :key="secret.id" :value="secret.id">
                  {{ secret.name }}
                </SelectItem>
              </SelectGroup>
            </SelectContent>
          </Select>
        </Field>
      </formApi.Field>
      <FieldDescription as="p" mt="2">
        {{ t('notifications.for-smtp-telegram-and-other-service-urls-the') }}
      </FieldDescription>
      <formApi.Field v-slot="{ field }" name="enabled">
        <Switch :model-value="field.state.value" :label="t('notifications.enable-channel')" mt="5" @update:model-value="field.handleChange" @focusout="field.handleBlur" />
      </formApi.Field>
      <Banner v-if="error" variant="error" class="mt-4">
        {{ error }}
      </Banner>
    </form>
    <template #footer>
      <Button :disabled="saving" @click="open = false">
        {{ t('common.cancel') }}
      </Button><Button type="submit" form="channel-form" :loading="saving" variant="primary">
        {{ t('notifications.save') }}
      </Button>
    </template>
  </Dialog><Dialog v-model:open="deleteOpen" :close-label="t('common.close')" :title="t('notifications.delete-channel')">
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
