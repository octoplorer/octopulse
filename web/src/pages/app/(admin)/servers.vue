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
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  getBeszelConfigQuery,
  getBeszelHistoryQuery,
  listBeszelContainersQuery,
  listBeszelSystemsQuery,
  listSecretsQuery,
  updateBeszelConfigMutation,
} from '../../../client/@pinia/colada.gen'
import AsyncState from '../../../components/AsyncState.vue'
import EmptyState from '../../../components/EmptyState.vue'
import Field from '../../../components/Field.vue'
import Modal from '../../../components/Modal.vue'
import PageHeader from '../../../components/PageHeader.vue'
import SecretSelect from '../../../components/SecretSelect.vue'
import Sparkline from '../../../components/Sparkline.vue'
import Toggle from '../../../components/Toggle.vue'
import { Alert } from '../../../components/ui/alert'
import { Badge } from '../../../components/ui/badge'
import { Button } from '../../../components/ui/button'
import { Card } from '../../../components/ui/card'
import { FieldError, FieldInput } from '../../../components/ui/field'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '../../../components/ui/select'
import { Separator } from '../../../components/ui/separator'
import { Table, TableBody, TableCell, TableContainer, TableHead, TableHeader, TableRow } from '../../../components/ui/table'
import { TabsContent, TabsList, TabsRoot, TabsTrigger } from '../../../components/ui/tabs'
import { isAdmin } from '../../../composables/api'
import { notify } from '../../../composables/notices'
import { usePollingEnabled } from '../../../composables/polling'
import { duration, formatDate } from '../../../composables/preferences'
import { errorText } from '../../../lib/errors'

const { t, n } = useI18n({ useScope: 'global' })

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
  <PageHeader :title="t('navigation.servers')" :description="t('servers.independentServerMetricsFromBeszelSeparateFromWebsite')">
    <Button @click="query.refetch()">
      <span w="14px" h="14px" aria-hidden="true" class="i-lucide-refresh-cw" />{{ t('common.refresh') }}
    </Button><Button v-if="isAdmin()" variant="primary" @click="configure">
      <span w="14px" h="14px" aria-hidden="true" class="i-lucide-settings" />{{ t('servers.beszelConnection') }}
    </Button>
  </PageHeader>
  <div flex="~ items-center" gap="9px" mb="22px" p="y-13px x-16px" border="1 solid line" rounded="8px" bg="base" un-text="12px subtle">
    <span w="16px" h="16px" aria-hidden="true" class="i-lucide-server" /><span>{{ t('common.source') }}: {{ query.data.value?.source || 'Beszel' }} ·
      {{ t('servers.lastSync') }} {{ formatDate(query.data.value?.syncedAt)
      }}<span v-if="query.data.value?.stale"> · {{ t('servers.dataIsStale') }}</span></span>
  </div>
  <Alert v-if="query.data.value?.error" role="alert" variant="destructive">
    {{ query.data.value.error }}
  </Alert>
  <AsyncState :pending="query.isPending.value" :error="query.error.value" @retry="query.refetch()">
    <EmptyState v-if="!query.data.value?.items.length" :title="t('servers.connectYourBeszelHub')" :description="t('servers.useADedicatedAccountToReadItsVisible')">
      <Button v-if="isAdmin()" variant="primary" @click="configure">
        {{ t('servers.configureConnection') }}
      </Button>
    </EmptyState>
    <div v-else grid="~ cols-3" gap="20px" class="[@media(max-width:1200px)]:grid-cols-2 [@media(max-width:900px)]:grid-cols-2 [@media(max-width:700px)]:grid-cols-1 [@container_workspace_(max-width:_700px)]:grid-cols-1!">
      <Card v-for="server in query.data.value.items" :key="server.id" as="article" p="22px" class="[&_h2]:text-16px">
        <div flex="~ items-center justify-between gap-3">
          <div flex="~ items-center gap-3">
            <span class="monitor-type-icon" flex="~ items-center justify-center shrink-0" size="32px" border="1 solid line" rounded="8px" un-text="subtle" bg="base"><span w="17px" h="17px" aria-hidden="true" class="i-lucide-server" /></span>
            <div>
              <h2>{{ server.name }}</h2>
              <p class="muted" un-text="13px subtle">
                {{ server.host || server.id }}
              </p>
            </div>
          </div>
          <Badge>{{ server.status }}</Badge>
        </div>
        <div grid="~ cols-2" gap="18px" mt="25px" class="tabular-nums [&_strong]:block [&_strong]:mt-5px [&_strong]:text-19px [&_strong]:font-[var(--font-sans)] [&_.metric-progress]:mt-8px [&_.metric-progress]:h-5px [&_.metric-progress]:overflow-hidden [&_.metric-progress]:rounded-5px [&_.metric-progress]:bg-line [&_.metric-progress_span]:block [&_.metric-progress_span]:h-full [&_.metric-progress_span]:rounded-5px [&_.metric-progress_span]:bg-brand">
          <div>
            <span class="mini-label" un-text="12px subtle" tracking="0.5px">CPU</span><strong>{{ percentage(server.cpu) }}</strong>
            <div class="metric-progress">
              <span :style="{ width: `${Math.min(server.cpu || 0, 100)}%` }" />
            </div>
          </div>
          <div>
            <span class="mini-label" un-text="12px subtle" tracking="0.5px">{{ t('servers.memory') }}</span><strong>{{ percentage(server.memory) }}</strong>
            <div class="metric-progress">
              <span :style="{ width: `${Math.min(server.memory || 0, 100)}%` }" />
            </div>
          </div>
          <div>
            <span class="mini-label" un-text="12px subtle" tracking="0.5px">{{ t('servers.disk') }}</span><strong>{{ percentage(server.disk) }}</strong>
            <div class="metric-progress">
              <span :style="{ width: `${Math.min(server.disk || 0, 100)}%` }" />
            </div>
          </div>
          <div>
            <span class="mini-label" un-text="12px subtle" tracking="0.5px">{{ t('servers.updated') }}</span>
            <p mt="2" class="muted" un-text="13px subtle">
              {{ formatDate(server.updatedAt) }}
            </p>
            <span v-if="server.stale" class="certificate-risk" un-text="12px fg-warning">{{ t('servers.staleData') }}</span>
          </div>
        </div>
        <Separator />
        <Button p="0!" variant="ghost" @click="detail(server)">
          {{ t('servers.historyContainers') }}<span w="14px" h="14px" aria-hidden="true" class="i-lucide-arrow-up-right" />
        </Button>
      </Card>
    </div>
  </AsyncState>
  <Alert mt="6" as="p" variant="default">
    {{ t('servers.offlineServersStaleDataOrIncompatibleVersionsDo') }}
  </Alert>
  <Modal v-model:open="configOpen" :title="t('servers.beszelHubConnection')">
    <form id="beszel-form" @submit.prevent="configForm.handleSubmit">
      <Alert mb="5" as="p" variant="default">
        {{ t('servers.supportsBeszel020XUseADedicated') }}
      </Alert>
      <configForm.Field v-slot="{ field }" name="url">
        <Field :label="t('servers.hubUrl')">
          <FieldInput :model-value="field.state.value" type="url" placeholder="https://beszel.example.com" required @update:model-value="field.handleChange($event)" @blur="field.handleBlur" />
        </Field>
      </configForm.Field><configForm.Field v-slot="{ field }" name="email">
        <Field :label="t('servers.dedicatedAccountEmail')" mt="4">
          <FieldInput :model-value="field.state.value" type="email" required @update:model-value="field.handleChange($event)" @blur="field.handleBlur" />
        </Field>
      </configForm.Field><configForm.Field v-slot="{ field }" name="passwordSecretId">
        <Field :label="t('servers.passwordSecretReference')" mt="4">
          <SecretSelect :model-value="field.state.value" :secrets="secrets.data.value?.items || []" @update:model-value="field.handleChange($event || '')" @focusout="field.handleBlur" />
        </Field>
      </configForm.Field><configForm.Field v-slot="{ field }" name="pollSeconds">
        <Field :label="t('servers.syncIntervalSeconds')" mt="4">
          <FieldInput :model-value="field.state.value" type="number" min="30" required @update:model-value="field.handleChange(Number($event))" @blur="field.handleBlur" />
        </Field>
      </configForm.Field><configForm.Field v-slot="{ field }" name="enabled">
        <Toggle :model-value="field.state.value" :label="t('servers.enableIntegration')" mt="5" @update:model-value="field.handleChange" @focusout="field.handleBlur" />
      </configForm.Field>
      <FieldError v-if="error" as="p" py="10px" px="0">
        {{ error }}
      </FieldError>
    </form>
    <template #footer>
      <Button @click="configOpen = false">
        {{ t('common.cancel') }}
      </Button><Button form="beszel-form" :disabled="saving" variant="primary">
        <span w="14px" h="14px" aria-hidden="true" class="i-lucide-save" />{{ t('servers.saveConnection') }}
      </Button>
    </template>
  </Modal><Modal v-model:open="detailOpen" :title="selected?.name || ''" wide>
    <div v-if="selected?.info" mb="5" grid="~ cols-2" gap="15px" class="[@media(max-width:700px)]:grid-cols-1">
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
    <AsyncState :pending="detailLoading" :error="detailError" @retry="selected && detail(selected)">
      <TabsRoot v-model="tab">
        <TabsList>
          <TabsTrigger value="history">
            {{
              t('servers.history')
            }}
          </TabsTrigger><TabsTrigger value="containers">
            {{
              t('servers.containers')
            }}
          </TabsTrigger>
        </TabsList><TabsContent value="history">
          <div flex="~ items-center justify-between gap-3" mt="5">
            <Alert as="p" variant="default">
              {{ t('common.source') }}: {{ historyMeta?.source || 'Beszel' }} ·
              {{ formatDate(historyMeta?.syncedAt)
              }}<span v-if="historyMeta?.stale"> · {{ t('common.staleData') }}</span>
            </Alert>
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
          <FieldError v-if="historyMeta?.error" as="p" py="10px" px="0">
            {{ historyMeta.error }}
          </FieldError>
          <EmptyState v-if="!history.length" :title="t('servers.noHistoryReturnedByTheHub')" />
          <div v-else py="6">
            <div grid="~ cols-2" gap="15px" class="[@media(max-width:700px)]:grid-cols-1">
              <section>
                <h3>CPU (%)</h3>
                <Sparkline show-scale :values="history.map((x) => x.cpu)" :timestamps="history.map((x) => x.at)" :height="115" />
              </section>
              <section>
                <h3>{{ t('common.memory') }} (%)</h3>
                <Sparkline show-scale :values="history.map((x) => x.memory)" :timestamps="history.map((x) => x.at)" :height="115" />
              </section>
              <section>
                <h3>{{ t('servers.disk2') }} (%)</h3>
                <Sparkline show-scale :values="history.map((x) => x.disk)" :timestamps="history.map((x) => x.at)" :height="115" />
              </section>
              <section>
                <h3>{{ t('servers.networkReceived') }} (MiB/s)</h3>
                <Sparkline show-scale :values="history.map((x) => x.networkIn / 1048576)" :timestamps="history.map((x) => x.at)" :height="115" />
              </section>
              <section>
                <h3>{{ t('servers.networkSent') }} (MiB/s)</h3>
                <Sparkline show-scale :values="history.map((x) => x.networkOut / 1048576)" :timestamps="history.map((x) => x.at)" :height="115" />
              </section>
            </div>
            <p mt="5" class="muted" un-text="13px subtle">
              {{ formatDate(history[0]?.at) }} — {{ formatDate(history.at(-1)?.at) }}
            </p>
          </div>
        </TabsContent><TabsContent value="containers">
          <Alert mt="5" as="p" variant="default">
            {{ t('common.source') }}: {{ containersMeta?.source || 'Beszel' }} ·
            {{ formatDate(containersMeta?.syncedAt)
            }}<span v-if="containersMeta?.stale"> · {{ t('common.staleData') }}</span>
          </Alert>
          <FieldError v-if="containersMeta?.error" as="p" py="10px" px="0">
            {{ containersMeta.error }}
          </FieldError>
          <EmptyState v-if="!containers.length" :title="t('servers.noVisibleContainerData')" />
          <TableContainer v-else mt="5">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{{ t('servers.container') }}</TableHead>
                  <TableHead>{{ t('common.status') }}</TableHead>
                  <TableHead>CPU</TableHead>
                  <TableHead>{{ t('common.memory') }} (MiB)</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="container in containers" :key="container.id || container.name">
                  <TableCell>
                    <span class="monitor-name block" font="600" un-text="13px">{{ container.name }}</span><span class="monitor-sub block [overflow-wrap:anywhere]" un-text="12px subtle" mt="3px" max-w="300px">{{ container.image }}</span>
                  </TableCell>
                  <TableCell>{{ container.status }}</TableCell>
                  <TableCell>{{ percentage(container.cpu) }}</TableCell>
                  <TableCell>
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
    </AsyncState>
  </Modal>
</template>
