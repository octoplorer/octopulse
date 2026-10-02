<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { ref, reactive } from 'vue'
import { Server, Settings, RefreshCw, ArrowUpRight, Save } from '@lucide/vue'
import { Tabs } from '@ark-ui/vue/tabs'
import { useMutation, useQuery, useQueryCache, type DefineQueryOptions } from '@pinia/colada'
import {
  getBeszelConfigQuery,
  getBeszelHistoryQuery,
  listBeszelContainersQuery,
  listBeszelSystemsQuery,
  listSecretsQuery,
  updateBeszelConfigMutation,
} from '../../../client/@pinia/colada.gen'
import type { ErrorModel, GetBeszelHistoryData } from '../../../client/types.gen'
import type {
  BeszelConfig,
  BeszelSystem,
  BeszelSystems,
  BeszelHistory,
  BeszelHistoryPoint,
  BeszelContainer,
  BeszelContainers,
  Secret,
} from '../../../lib/types'
import { isAdmin } from '../../../lib/api'
import { formatDate, duration } from '../../../lib/preferences'
import { notify, errorText } from '../../../lib/notices'
import { useIntervalFn } from '@vueuse/core'
import PageHeader from '../../../components/PageHeader.vue'
import Field from '../../../components/Field.vue'
import Toggle from '../../../components/Toggle.vue'
import SecretSelect from '../../../components/SecretSelect.vue'
import Modal from '../../../components/Modal.vue'
import AsyncState from '../../../components/AsyncState.vue'
import EmptyState from '../../../components/EmptyState.vue'
import Sparkline from '../../../components/Sparkline.vue'

const { t, n } = useI18n({ useScope: 'global' })

definePage({ meta: { title: 'navigation.servers' } })

const updateConfig = useMutation(updateBeszelConfigMutation())
const queryCache = useQueryCache()
const query = useQuery({
  ...listBeszelSystemsQuery(),
  staleTime: 5000,
} as DefineQueryOptions<BeszelSystems, ErrorModel>)
const secrets = useQuery({
  ...listSecretsQuery(),
  staleTime: 10000,
} as DefineQueryOptions<{ items: Secret[] }, ErrorModel>)
const configOpen = ref(false)
const detailOpen = ref(false)
const selected = ref<BeszelSystem | null>(null)
const saving = ref(false)
const error = ref('')
const history = ref<BeszelHistoryPoint[]>([])
const containers = ref<BeszelContainer[]>([])
const detailLoading = ref(false)
const detailError = ref('')
const tab = ref('history')
const historyRange = ref<NonNullable<GetBeszelHistoryData['query']>['range']>('24h')
const historyMeta = ref<BeszelHistory | null>(null)
const containersMeta = ref<BeszelContainers | null>(null)
const config = reactive<BeszelConfig>({
  url: '',
  email: '',
  passwordSecretId: '',
  enabled: false,
  pollSeconds: 60,
})
let detailRequest = 0
async function configure() {
  try {
    const state = await queryCache.refresh(
      queryCache.ensure({ ...getBeszelConfigQuery(), staleTime: 0 }),
    )
    if (state.status !== 'success') throw state.error || new Error(t('errors.requestFailed'))
    Object.assign(config, structuredClone(state.data))
    error.value = ''
    configOpen.value = true
  } catch (e) {
    notify(errorText(e), 'error')
  }
}
async function save() {
  saving.value = true
  error.value = ''
  try {
    Object.assign(config, await updateConfig.mutateAsync({ body: config }))
    configOpen.value = false
    notify(t('servers.beszelConnectionSaved'))
    await query.refresh()
  } catch (e) {
    error.value = errorText(e)
  } finally {
    saving.value = false
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
    if (request !== detailRequest) return
    if (h.status !== 'success') throw h.error || new Error(t('errors.requestFailed'))
    if (c.status !== 'success') throw c.error || new Error(t('errors.requestFailed'))
    historyMeta.value = h.data as BeszelHistory
    containersMeta.value = c.data as BeszelContainers
    history.value = h.data.items || []
    containers.value = c.data.items || []
  } catch (e) {
    if (request === detailRequest) detailError.value = errorText(e)
  } finally {
    if (request === detailRequest) detailLoading.value = false
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
useIntervalFn(() => query.refetch(), 30000)
</script>
<template>
  <PageHeader
    :title="t('navigation.servers')"
    :description="t('servers.independentServerMetricsFromBeszelSeparateFromWebsite')"
    ><button class="button" @click="query.refetch()">
      <RefreshCw :size="14" />{{ t('common.refresh') }}</button
    ><button v-if="isAdmin()" class="button primary" @click="configure">
      <Settings :size="14" />{{ t('servers.beszelConnection') }}
    </button></PageHeader
  >
  <div class="alert-strip">
    <Server :size="16" /><span
      >{{ t('common.source') }}: {{ query.data.value?.source || 'Beszel' }} ·
      {{ t('servers.lastSync') }} {{ formatDate(query.data.value?.syncedAt)
      }}<span v-if="query.data.value?.stale"> · {{ t('servers.dataIsStale') }}</span></span
    >
  </div>
  <div v-if="query.data.value?.error" class="error-banner" role="alert">
    {{ query.data.value.error }}
  </div>
  <AsyncState :pending="query.isPending.value" :error="query.error.value" @retry="query.refetch()"
    ><EmptyState
      v-if="!query.data.value?.items.length"
      :title="t('servers.connectYourBeszelHub')"
      :description="t('servers.useADedicatedAccountToReadItsVisible')"
      ><button v-if="isAdmin()" class="button primary" @click="configure">
        {{ t('servers.configureConnection') }}
      </button></EmptyState
    >
    <div v-else class="servers-grid">
      <article v-for="server in query.data.value.items" :key="server.id" class="card server-card">
        <div un-flex="~ items-center justify-between gap-3">
          <div un-flex="~ items-center gap-3">
            <span class="monitor-type-icon"><Server :size="17" /></span>
            <div>
              <h2>{{ server.name }}</h2>
              <p class="muted" un-text="10px">{{ server.host || server.id }}</p>
            </div>
          </div>
          <span class="pill">{{ server.status }}</span>
        </div>
        <div class="server-metrics">
          <div>
            <span class="mini-label">CPU</span><strong>{{ percentage(server.cpu) }}</strong>
            <div class="metric-progress">
              <span :style="{ width: `${Math.min(server.cpu || 0, 100)}%` }" />
            </div>
          </div>
          <div>
            <span class="mini-label">{{ t('servers.memory') }}</span
            ><strong>{{ percentage(server.memory) }}</strong>
            <div class="metric-progress">
              <span :style="{ width: `${Math.min(server.memory || 0, 100)}%` }" />
            </div>
          </div>
          <div>
            <span class="mini-label">{{ t('servers.disk') }}</span
            ><strong>{{ percentage(server.disk) }}</strong>
            <div class="metric-progress">
              <span :style="{ width: `${Math.min(server.disk || 0, 100)}%` }" />
            </div>
          </div>
          <div>
            <span class="mini-label">{{ t('servers.updated') }}</span>
            <p class="muted" un-text="10px" un-mt="2">{{ formatDate(server.updatedAt) }}</p>
            <span v-if="server.stale" class="certificate-risk">{{ t('servers.staleData') }}</span>
          </div>
        </div>
        <div class="section-divider" />
        <button class="button ghost" un-p="0!" @click="detail(server)">
          {{ t('servers.historyContainers') }}<ArrowUpRight :size="14" />
        </button>
      </article></div
  ></AsyncState>
  <p class="note" un-mt="6">
    {{ t('servers.offlineServersStaleDataOrIncompatibleVersionsDo') }}
  </p>
  <Modal v-model:open="configOpen" :title="t('servers.beszelHubConnection')"
    ><form id="beszel-form" @submit.prevent="save">
      <p class="note" un-mb="5">
        {{ t('servers.supportsBeszel020XUseADedicated') }}
      </p>
      <Field :label="t('servers.hubUrl')"
        ><input
          v-model="config.url"
          type="url"
          placeholder="https://beszel.example.com"
          required /></Field
      ><Field :label="t('servers.dedicatedAccountEmail')" un-mt="4"
        ><input v-model="config.email" type="email" required /></Field
      ><Field :label="t('servers.passwordSecretReference')" un-mt="4"
        ><SecretSelect
          v-model="config.passwordSecretId"
          :secrets="secrets.data.value?.items || []" /></Field
      ><Field :label="t('servers.syncIntervalSeconds')" un-mt="4"
        ><input v-model.number="config.pollSeconds" type="number" min="30" required /></Field
      ><Toggle v-model="config.enabled" :label="t('servers.enableIntegration')" un-mt="5" />
      <p v-if="error" class="inline-error">{{ error }}</p>
    </form>
    <template #footer
      ><button class="button" @click="configOpen = false">{{ t('common.cancel') }}</button
      ><button class="button primary" form="beszel-form" :disabled="saving">
        <Save :size="14" />{{ t('servers.saveConnection') }}
      </button></template
    ></Modal
  ><Modal v-model:open="detailOpen" :title="selected?.name || ''" wide
    ><div v-if="selected?.info" class="hint-grid" un-mb="5">
      <div>
        <span class="mini-label">{{ t('servers.hostname') }}</span>
        <p>{{ selected.info.hostname || '—' }}</p>
      </div>
      <div>
        <span class="mini-label">CPU</span>
        <p>{{ selected.info.cpuModel || '—' }}</p>
        <p class="muted">
          {{ selected.info.cores }} {{ t('servers.cores') }} / {{ selected.info.threads }}
          {{ t('servers.threads') }}
        </p>
      </div>
      <div>
        <span class="mini-label">{{ t('servers.systemAgent') }}</span>
        <p>{{ selected.info.kernel || '—' }}</p>
        <p class="muted">
          {{ selected.info.agentVersion || '—' }} ·
          {{ duration(selected.info.uptimeSeconds * 1000) }}
        </p>
      </div>
    </div>
    <AsyncState :pending="detailLoading" :error="detailError" @retry="selected && detail(selected)"
      ><Tabs.Root v-model="tab"
        ><Tabs.List class="tabs-list"
          ><Tabs.Trigger class="tabs-trigger" value="history">{{
            t('servers.history')
          }}</Tabs.Trigger
          ><Tabs.Trigger class="tabs-trigger" value="containers">{{
            t('servers.containers')
          }}</Tabs.Trigger></Tabs.List
        ><Tabs.Content value="history">
          <div un-flex="~ items-center justify-between gap-3" un-mt="5">
            <p class="note">
              {{ t('common.source') }}: {{ historyMeta?.source || 'Beszel' }} ·
              {{ formatDate(historyMeta?.syncedAt)
              }}<span v-if="historyMeta?.stale"> · {{ t('common.staleData') }}</span>
            </p>
            <select
              v-model="historyRange"
              :aria-label="t('servers.historyRange')"
              @change="selected && detail(selected)"
            >
              <option v-for="range in ['1h', '12h', '24h', '1w', '30d']" :key="range">
                {{ range }}
              </option>
            </select>
          </div>
          <p v-if="historyMeta?.error" class="inline-error">{{ historyMeta.error }}</p>
          <EmptyState v-if="!history.length" :title="t('servers.noHistoryReturnedByTheHub')" />
          <div v-else un-py="6">
            <div class="hint-grid">
              <section>
                <h3>CPU (%)</h3>
                <Sparkline
                  show-scale
                  :values="history.map((x) => x.cpu)"
                  :timestamps="history.map((x) => x.at)"
                  :height="115"
                />
              </section>
              <section>
                <h3>{{ t('common.memory') }} (%)</h3>
                <Sparkline
                  show-scale
                  :values="history.map((x) => x.memory)"
                  :timestamps="history.map((x) => x.at)"
                  :height="115"
                />
              </section>
              <section>
                <h3>{{ t('servers.disk2') }} (%)</h3>
                <Sparkline
                  show-scale
                  :values="history.map((x) => x.disk)"
                  :timestamps="history.map((x) => x.at)"
                  :height="115"
                />
              </section>
              <section>
                <h3>{{ t('servers.networkReceived') }} (MiB/s)</h3>
                <Sparkline
                  show-scale
                  :values="history.map((x) => x.networkIn / 1048576)"
                  :timestamps="history.map((x) => x.at)"
                  :height="115"
                />
              </section>
              <section>
                <h3>{{ t('servers.networkSent') }} (MiB/s)</h3>
                <Sparkline
                  show-scale
                  :values="history.map((x) => x.networkOut / 1048576)"
                  :timestamps="history.map((x) => x.at)"
                  :height="115"
                />
              </section>
            </div>
            <p class="muted" un-text="10px" un-mt="5">
              {{ formatDate(history[0]?.at) }} — {{ formatDate(history.at(-1)?.at) }}
            </p>
          </div></Tabs.Content
        ><Tabs.Content value="containers"
          ><p class="note" un-mt="5">
            {{ t('common.source') }}: {{ containersMeta?.source || 'Beszel' }} ·
            {{ formatDate(containersMeta?.syncedAt)
            }}<span v-if="containersMeta?.stale"> · {{ t('common.staleData') }}</span>
          </p>
          <p v-if="containersMeta?.error" class="inline-error">{{ containersMeta.error }}</p>
          <EmptyState v-if="!containers.length" :title="t('servers.noVisibleContainerData')" />
          <div v-else class="table-wrap" un-mt="5">
            <table class="data-table">
              <thead>
                <tr>
                  <th>{{ t('servers.container') }}</th>
                  <th>{{ t('common.status') }}</th>
                  <th>CPU</th>
                  <th>{{ t('common.memory') }} (MiB)</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="container in containers" :key="container.id || container.name">
                  <td>
                    <span class="monitor-name">{{ container.name }}</span
                    ><span class="monitor-sub">{{ container.image }}</span>
                  </td>
                  <td>{{ container.status }}</td>
                  <td>{{ percentage(container.cpu) }}</td>
                  <td>
                    {{
                      n(container.memory, { minimumFractionDigits: 1, maximumFractionDigits: 1 })
                    }}
                    MiB
                  </td>
                </tr>
              </tbody>
            </table>
          </div></Tabs.Content
        ></Tabs.Root
      ></AsyncState
    ></Modal
  >
</template>
