<script setup lang="ts">
import type {
  BeszelConfig,
  Container as BeszelContainer,
  ContainersResponse as BeszelContainers,
  HistoryResponse as BeszelHistory,
  HistoryPoint as BeszelHistoryPoint,
  System as BeszelSystem,
  GetBeszelHistoryData,
} from '../../../client/types.gen'
import { useMutation, useQuery, useQueryCache } from '@pinia/colada'
import { useForm } from '@tanstack/vue-form'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  getBeszelConfigQuery,
  getBeszelHistoryQuery,
  listBeszelContainersQuery,
  listBeszelSystemsQuery,
  listSecretsQuery,
  updateBeszelConfigMutation,
} from '../../../client/@pinia/colada.gen'
import { Badge } from '../../../components/ui/badge'
import { Banner } from '../../../components/ui/banner'
import { PageHeader } from '../../../components/ui/blocks/page-header'
import { Button } from '../../../components/ui/button'
import { Chart } from '../../../components/ui/chart'
import { Dialog } from '../../../components/ui/dialog'
import { Empty } from '../../../components/ui/empty'
import { Field, FieldError } from '../../../components/ui/field'
import { Input } from '../../../components/ui/input'
import { LayerCard, LayerCardPrimary } from '../../../components/ui/layer-card'
import { Loader } from '../../../components/ui/loader'
import { Meter } from '../../../components/ui/meter'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '../../../components/ui/select'
import { Switch } from '../../../components/ui/switch'
import { Table, TableBody, TableCell, TableContainer, TableHead, TableHeader, TableRow } from '../../../components/ui/table'
import { TabsContent, TabsList, TabsRoot, TabsTrigger } from '../../../components/ui/tabs'
import { isAdmin } from '../../../composables/api'
import { notify } from '../../../composables/notices'
import { usePollingEnabled } from '../../../composables/polling'
import { duration, formatDate, timezone } from '../../../composables/preferences'
import { normalizeTimeSeries, timeSeriesOption } from '../../../lib/chart'
import { errorText } from '../../../lib/errors'

const { t, n, d } = useI18n({ useScope: 'global' })

definePage({ meta: { title: 'navigation.servers' } })

const updateConfig = useMutation(updateBeszelConfigMutation())
const queryCache = useQueryCache()
const pollingEnabled = usePollingEnabled()
const query = useQuery(
  () =>
    ({
      ...listBeszelSystemsQuery(),
      staleTime: 5000,
      enabled: pollingEnabled.value,
      autoRefetch: 30000,
    }),
)
const integrationDisabled = computed(() => query.data.value?.error === 'Beszel integration is disabled')
const secrets = useQuery({
  ...listSecretsQuery(),
  staleTime: 10000,
})
const configOpen = ref(false)
const detailOpen = ref(false)
const selected = ref<BeszelSystem | null>(null)
const error = ref('')
const history = ref<BeszelHistoryPoint[]>([])
const containers = ref<BeszelContainer[]>([])
const detailLoading = ref(false)
const detailError = ref('')
const tab = ref('history')
const historyRange = ref<NonNullable<GetBeszelHistoryData['query']>['range']>('24h')
const historyMeta = ref<BeszelHistory | null>(null)
const containersMeta = ref<BeszelContainers | null>(null)
const historyCharts = computed(() => {
  const timeZone = timezone.value
  return ([
    { key: 'cpu', name: 'CPU', unit: '%', divisor: 1 },
    { key: 'memory', name: t('common.memory'), unit: '%', divisor: 1 },
    { key: 'disk', name: t('servers.disk2'), unit: '%', divisor: 1 },
    { key: 'networkIn', name: t('servers.networkReceived'), unit: 'MiB/s', divisor: 1048576 },
    { key: 'networkOut', name: t('servers.networkSent'), unit: 'MiB/s', divisor: 1048576 },
  ] as const).map((metric) => {
    const points = normalizeTimeSeries(history.value.map(point => ({
      at: point.at,
      value: point[metric.key] / metric.divisor,
    })))
    return {
      ...metric,
      hasData: points.length > 0,
      option: timeSeriesOption({
        points,
        name: metric.name,
        unit: metric.unit,
        valueFormatter: value => n(value, { maximumFractionDigits: 2 }),
        timeFormatter: at => d(at, { key: 'short', timeZone }),
      }),
      label: t('chart.summary', {
        name: metric.name,
        count: points.length,
        value: points.length ? n(points.at(-1)!.value, { maximumFractionDigits: 2 }) : '—',
        unit: metric.unit,
      }),
    }
  })
})
const configForm = useForm({
  defaultValues: {
    url: '',
    email: '',
    passwordSecretId: '',
    enabled: false,
    pollSeconds: 60,
  } as BeszelConfig,
  onSubmit: async ({ value }) => {
    error.value = ''
    try {
      const result = await updateConfig.mutateAsync({ body: value })
      configOpen.value = false
      notify(t('servers.beszelConnectionSaved'))
      await query.refresh()
      configForm.reset(result)
    }
    catch (e) {
      error.value = errorText(e)
    }
  },
})
const saving = configForm.useSelector(state => state.isSubmitting)
const configDialogOpen = computed({
  get: () => configOpen.value,
  set: (open: boolean) => {
    if (!saving.value)
      configOpen.value = open
  },
})
let detailRequest = 0
async function configure() {
  if (configForm.state.isSubmitting)
    return
  try {
    const state = await queryCache.refresh(
      queryCache.ensure({ ...getBeszelConfigQuery(), staleTime: 0 }),
    )
    if (configForm.state.isSubmitting)
      return
    if (state.status !== 'success')
      throw state.error || new Error(t('errors.requestFailed'))
    configForm.reset(structuredClone(state.data))
    error.value = ''
    configOpen.value = true
  }
  catch (e) {
    notify(errorText(e), 'error')
  }
}
async function detail(server: BeszelSystem) {
  const request = ++detailRequest
  selected.value = server
  detailOpen.value = true
  detailLoading.value = true
  detailError.value = ''
  try {
    const [h, c] = await Promise.all([
      queryCache.refresh(
        queryCache.ensure({
          ...getBeszelHistoryQuery({
            path: { id: server.id },
            query: { range: historyRange.value },
          }),
          staleTime: 0,
        }),
      ),
      queryCache.refresh(
        queryCache.ensure({
          ...listBeszelContainersQuery({
            path: { id: server.id },
          }),
          staleTime: 0,
        }),
      ),
    ])
    if (request !== detailRequest)
      return
    if (h.status !== 'success')
      throw h.error || new Error(t('errors.requestFailed'))
    if (c.status !== 'success')
      throw c.error || new Error(t('errors.requestFailed'))
    historyMeta.value = h.data
    containersMeta.value = c.data
    history.value = h.data.items || []
    containers.value = c.data.items || []
  }
  catch (e) {
    if (request === detailRequest)
      detailError.value = errorText(e)
  }
  finally {
    if (request === detailRequest)
      detailLoading.value = false
  }
}
function percentage(value: number | undefined) {
  return value == null
    ? '—'
    : n(Math.max(0, value) / 100, {
        style: 'percent',
        minimumFractionDigits: 1,
        maximumFractionDigits: 1,
      })
}
</script>

<template>
  <PageHeader :title="t('navigation.servers')" :description="t('servers.independentServerMetricsFromBeszelSeparateFromWebsite')" class="mb-6">
    <template #actions>
      <Button v-if="!integrationDisabled" @click="query.refetch()">
        <span w="14px" h="14px" aria-hidden="true" class="i-lucide-refresh-cw" />{{ t('common.refresh') }}
      </Button><Button v-if="isAdmin() && !integrationDisabled" variant="primary" @click="configure">
        <span w="14px" h="14px" aria-hidden="true" class="i-lucide-settings" />{{ t('servers.beszelConnection') }}
      </Button>
    </template>
  </PageHeader>
  <div v-if="!integrationDisabled" flex="~ items-center" gap="9px" mb="22px" p="y-13px x-16px" border="1 solid line" rounded="8px" bg="base" un-text="12px subtle">
    <span w="16px" h="16px" aria-hidden="true" class="i-lucide-server shrink-0" /><span class="min-w-0 [overflow-wrap:anywhere]">{{ t('common.source') }}: {{ query.data.value?.source || 'Beszel' }} ·
      {{ t('servers.lastSync') }} {{ formatDate(query.data.value?.syncedAt)
      }}<span v-if="query.data.value?.stale"> · {{ t('servers.dataIsStale') }}</span></span>
  </div>
  <Banner v-if="query.data.value?.error && !integrationDisabled" role="alert" variant="error" class="mb-6">
    {{ query.data.value.error }}
    <Button variant="ghost" @click="query.refetch()">
      {{ t('asyncState.retry') }}
    </Button>
  </Banner>
  <Loader v-if="query.isPending.value" :label="t('asyncState.loadingData')" class="flex! w-full justify-center p-15 text-size-xs">
    {{ t('asyncState.loadingData') }}
  </Loader>
  <Banner v-else-if="query.error.value" variant="error">
    {{ errorText(query.error.value) }}
    <Button variant="ghost" @click="query.refetch()">
      {{ t('asyncState.retry') }}
    </Button>
  </Banner>
  <template v-else>
    <Empty v-if="integrationDisabled || (!query.data.value?.error && !query.data.value?.items.length)" size="sm" :title="t('servers.connectYourBeszelHub')" :description="t('servers.useADedicatedAccountToReadItsVisible')">
      <Button v-if="isAdmin()" variant="primary" @click="configure">
        {{ t('servers.configureConnection') }}
      </Button>
    </Empty>
    <div v-else class="grid grid-cols-[repeat(auto-fill,minmax(min(100%,320px),1fr))] gap-6">
      <LayerCard v-for="server in query.data.value.items" :key="server.id" class="flex flex-col [&_h2]:text-16px">
        <LayerCardPrimary class="flex flex-1 flex-col gap-5 tabular-nums">
          <div flex="~ items-start justify-between gap-3">
            <div flex="~ items-start gap-3" class="min-w-0">
              <span class="monitor-type-icon" flex="~ items-center justify-center shrink-0" size="32px" border="1 solid line" rounded="8px" un-text="subtle" bg="base"><span w="17px" h="17px" aria-hidden="true" class="i-lucide-server" /></span>
              <div class="min-w-0 [overflow-wrap:anywhere]">
                <h2>{{ server.name }}</h2>
                <p class="muted" un-text="13px subtle">
                  {{ server.host || server.id }}
                </p>
              </div>
            </div>
            <Badge class="shrink-0">
              {{ server.status }}
            </Badge>
          </div>
          <div class="grid grid-cols-2 gap-5">
            <Meter label="CPU" :value="server.cpu ?? 0" :custom-value="percentage(server.cpu)" />
            <Meter :label="t('servers.memory')" :value="server.memory ?? 0" :custom-value="percentage(server.memory)" />
            <Meter :label="t('servers.disk')" :value="server.disk ?? 0" :custom-value="percentage(server.disk)" />
            <div class="space-y-2 text-size-xs text-subtle">
              <span>{{ t('servers.updated') }}</span><p>{{ formatDate(server.updatedAt) }}</p>
              <Badge v-if="server.stale" variant="warning">
                {{ t('servers.staleData') }}
              </Badge>
            </div>
          </div>
          <div class="mt-auto border-t border-line pt-4">
            <Button variant="ghost" size="sm" @click="detail(server)">
              {{ t('servers.historyContainers') }}<span class="i-lucide-arrow-up-right size-4" aria-hidden="true" />
            </Button>
          </div>
        </LayerCardPrimary>
      </LayerCard>
    </div>
  </template>
  <Banner v-if="!integrationDisabled" mt="6" variant="secondary">
    {{ t('servers.offlineServersStaleDataOrIncompatibleVersionsDo') }}
  </Banner>
  <Dialog v-model:open="configDialogOpen" :close-label="t('common.close')" :title="t('servers.beszelHubConnection')">
    <form id="beszel-form" @submit.prevent="configForm.handleSubmit()">
      <Banner mb="5" variant="secondary">
        {{ t('servers.supportsBeszel020XUseADedicated') }}
      </Banner>
      <configForm.Field v-slot="{ field }" name="url">
        <Field :label="t('servers.hubUrl')">
          <Input :model-value="field.state.value" type="url" placeholder="https://beszel.example.com" required @update:model-value="field.handleChange(String($event ?? ''))" @blur="field.handleBlur" />
        </Field>
      </configForm.Field><configForm.Field v-slot="{ field }" name="email">
        <Field :label="t('servers.dedicatedAccountEmail')" mt="4">
          <Input :model-value="field.state.value" type="email" required @update:model-value="field.handleChange(String($event ?? ''))" @blur="field.handleBlur" />
        </Field>
      </configForm.Field><configForm.Field v-slot="{ field }" name="passwordSecretId">
        <Field :label="t('servers.passwordSecretReference')" mt="4">
          <Select :model-value="field.state.value" @update:model-value="field.handleChange($event || '')">
            <SelectTrigger @focusout="field.handleBlur">
              <SelectValue :placeholder="t('secretSelect.chooseASecret')" />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                <SelectItem value="">
                  {{ t('secretSelect.chooseASecret') }}
                </SelectItem>
                <SelectItem v-for="secret in secrets.data.value?.items || []" :key="secret.id" :value="secret.id">
                  {{ secret.name }}
                </SelectItem>
              </SelectGroup>
            </SelectContent>
          </Select>
        </Field>
      </configForm.Field><configForm.Field v-slot="{ field }" name="pollSeconds">
        <Field :label="t('servers.syncIntervalSeconds')" mt="4">
          <Input :model-value="field.state.value" type="number" min="30" required @update:model-value="field.handleChange(Number($event))" @blur="field.handleBlur" />
        </Field>
      </configForm.Field><configForm.Field v-slot="{ field }" name="enabled">
        <Switch :model-value="field.state.value" :label="t('servers.enableIntegration')" mt="5" @update:model-value="field.handleChange" @focusout="field.handleBlur" />
      </configForm.Field>
      <FieldError v-if="error" as="p" py="10px" px="0">
        {{ error }}
      </FieldError>
    </form>
    <template #footer>
      <Button :disabled="saving" @click="configOpen = false">
        {{ t('common.cancel') }}
      </Button><Button type="submit" form="beszel-form" :loading="saving" variant="primary">
        <span w="14px" h="14px" aria-hidden="true" class="i-lucide-save" />{{ t('servers.saveConnection') }}
      </Button>
    </template>
  </Dialog><Dialog v-model:open="detailOpen" :close-label="t('common.close')" :title="selected?.name || ''" wide>
    <div v-if="selected?.info" mb="5" grid="~ cols-2" gap="15px" class="tabular-nums [overflow-wrap:anywhere] [@media(max-width:700px)]:grid-cols-1">
      <div>
        <span class="mini-label" un-text="12px subtle" tracking="0.5px">{{ t('servers.hostname') }}</span>
        <p>{{ selected.info.hostname || '—' }}</p>
      </div>
      <div>
        <span class="mini-label" un-text="12px subtle" tracking="0.5px">CPU</span>
        <p>{{ selected.info.cpuModel || '—' }}</p>
        <p class="muted" un-text="13px subtle">
          {{ selected.info.cores }} {{ t('servers.cores') }} / {{ selected.info.threads }}
          {{ t('servers.threads') }}
        </p>
      </div>
      <div>
        <span class="mini-label" un-text="12px subtle" tracking="0.5px">{{ t('servers.systemAgent') }}</span>
        <p>{{ selected.info.kernel || '—' }}</p>
        <p class="muted" un-text="13px subtle">
          {{ selected.info.agentVersion || '—' }} ·
          {{ duration(selected.info.uptimeSeconds * 1000) }}
        </p>
      </div>
    </div>
    <Loader v-if="detailLoading" :label="t('asyncState.loadingData')" class="flex! w-full justify-center p-15 text-size-xs">
      {{ t('asyncState.loadingData') }}
    </Loader>
    <Banner v-else-if="detailError" variant="error">
      {{ errorText(detailError) }}
      <Button variant="ghost" @click="selected && detail(selected)">
        {{ t('asyncState.retry') }}
      </Button>
    </Banner>
    <template v-else>
      <TabsRoot v-model="tab">
        <TabsList variant="line">
          <TabsTrigger variant="line" value="history">
            {{
              t('servers.history')
            }}
          </TabsTrigger><TabsTrigger variant="line" value="containers">
            {{
              t('servers.containers')
            }}
          </TabsTrigger>
        </TabsList><TabsContent value="history">
          <div flex="~ wrap items-center justify-between gap-3" mt="5">
            <Banner variant="secondary">
              {{ t('common.source') }}: {{ historyMeta?.source || 'Beszel' }} ·
              {{ formatDate(historyMeta?.syncedAt)
              }}<span v-if="historyMeta?.stale"> · {{ t('common.staleData') }}</span>
            </Banner>
            <Select v-model="historyRange" @change="selected && detail(selected)">
              <SelectTrigger :aria-label="t('servers.historyRange')">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem v-for="range in ['1h', '12h', '24h', '1w', '30d']" :key="range" :value="range">
                    {{ range }}
                  </SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </div>
          <Banner v-if="historyMeta?.error" variant="error" class="mt-4">
            {{ historyMeta.error }}
            <Button variant="ghost" @click="selected && detail(selected)">
              {{ t('asyncState.retry') }}
            </Button>
          </Banner>
          <Empty v-if="!history.length" size="sm" :title="t('servers.noHistoryReturnedByTheHub')" />
          <div v-else py="6">
            <div grid="~ cols-2" gap="15px" class="[@media(max-width:700px)]:grid-cols-1">
              <section v-for="chart in historyCharts" :key="chart.key" class="min-w-0">
                <h3>{{ chart.name }} ({{ chart.unit }})</h3>
                <Chart v-if="chart.hasData" :option="chart.option" :height="160" :aria-label="chart.label" />
                <div v-else h="160px" flex="~ items-center justify-center" un-text="subtle" role="status">
                  {{ t('chart.noObservations') }}
                </div>
              </section>
            </div>
            <p mt="5" class="muted" un-text="13px subtle">
              {{ formatDate(history[0]?.at) }} — {{ formatDate(history.at(-1)?.at) }}
            </p>
          </div>
        </TabsContent><TabsContent value="containers">
          <Banner mt="5" variant="secondary">
            {{ t('common.source') }}: {{ containersMeta?.source || 'Beszel' }} ·
            {{ formatDate(containersMeta?.syncedAt)
            }}<span v-if="containersMeta?.stale"> · {{ t('common.staleData') }}</span>
          </Banner>
          <Banner v-if="containersMeta?.error" variant="error" class="mt-4">
            {{ containersMeta.error }}
            <Button variant="ghost" @click="selected && detail(selected)">
              {{ t('asyncState.retry') }}
            </Button>
          </Banner>
          <Empty v-if="!containers.length" size="sm" :title="t('servers.noVisibleContainerData')" />
          <TableContainer v-else :scroll-label="t('common.scrollTable')" mt="5">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{{ t('servers.container') }}</TableHead>
                  <TableHead>{{ t('common.status') }}</TableHead>
                  <TableHead class="text-end">
                    CPU
                  </TableHead>
                  <TableHead class="text-end">
                    {{ t('common.memory') }} (MiB)
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="container in containers" :key="container.id || container.name">
                  <TableCell>
                    <span class="monitor-name block [overflow-wrap:anywhere]" font="600" un-text="13px">{{ container.name }}</span><span class="monitor-sub block [overflow-wrap:anywhere]" un-text="12px subtle" mt="3px" max-w="300px">{{ container.image }}</span>
                  </TableCell>
                  <TableCell>{{ container.status }}</TableCell>
                  <TableCell class="whitespace-nowrap text-end">
                    {{ percentage(container.cpu) }}
                  </TableCell>
                  <TableCell class="whitespace-nowrap text-end">
                    {{
                      n(container.memory, { minimumFractionDigits: 1, maximumFractionDigits: 1 })
                    }}
                    MiB
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </TableContainer>
        </TabsContent>
      </TabsRoot>
    </template>
  </Dialog>
</template>
