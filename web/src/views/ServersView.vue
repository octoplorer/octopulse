<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { Server, Settings, RefreshCw, ExternalLink, ArrowUpRight, Save, Clock } from '@lucide/vue'
import { Tabs } from '@ark-ui/vue/tabs'
import { useRecord, useCollection } from '../lib/data'
import type {
  BeszelConfig,
  BeszelSystem,
  BeszelSystems,
  BeszelHistory,
  BeszelHistoryPoint,
  BeszelContainer,
  BeszelContainers,
  Secret,
} from '../lib/types'
import { api, isAdmin } from '../lib/api'
import { t, formatDate, duration } from '../lib/preferences'
import { notify, errorText } from '../lib/notices'
import { useIntervalFn } from '@vueuse/core'
import PageHeader from '../components/PageHeader.vue'
import Field from '../components/Field.vue'
import Toggle from '../components/Toggle.vue'
import SecretSelect from '../components/SecretSelect.vue'
import Modal from '../components/Modal.vue'
import AsyncState from '../components/AsyncState.vue'
import EmptyState from '../components/EmptyState.vue'
import Sparkline from '../components/Sparkline.vue'
const query = useRecord<BeszelSystems>(() => 'beszel/systems'),
  secrets = useCollection<Secret>('secrets'),
  configOpen = ref(false),
  detailOpen = ref(false),
  selected = ref<BeszelSystem | null>(null),
  saving = ref(false),
  error = ref(''),
  history = ref<BeszelHistoryPoint[]>([]),
  containers = ref<BeszelContainer[]>([]),
  detailLoading = ref(false),
  detailError = ref(''),
  tab = ref('history'),
  historyRange = ref('24h'),
  historyMeta = ref<BeszelHistory | null>(null),
  containersMeta = ref<BeszelContainers | null>(null),
  config = reactive<BeszelConfig>({
    url: '',
    email: '',
    passwordSecretId: '',
    enabled: false,
    pollSeconds: 60,
  })
async function configure() {
  try {
    Object.assign(config, await api<BeszelConfig>('beszel/config'))
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
    Object.assign(
      config,
      await api<BeszelConfig>('beszel/config', { method: 'PATCH', body: config }),
    )
    configOpen.value = false
    notify(t('Beszel 连接已保存', 'Beszel connection saved'))
    await query.refresh()
  } catch (e) {
    error.value = errorText(e)
  } finally {
    saving.value = false
  }
}
async function detail(server: BeszelSystem) {
  selected.value = server
  detailOpen.value = true
  detailLoading.value = true
  detailError.value = ''
  try {
    const [h, c] = await Promise.all([
      api<BeszelHistory>(`beszel/systems/${server.id}/history?range=${historyRange.value}`),
      api<BeszelContainers>(`beszel/systems/${server.id}/containers`),
    ])
    historyMeta.value = h
    containersMeta.value = c
    history.value = h.items || []
    containers.value = c.items || []
  } catch (e) {
    detailError.value = errorText(e)
  } finally {
    detailLoading.value = false
  }
}
function percentage(value: number | undefined) {
  return value == null ? '—' : `${Math.max(0, value).toFixed(1)}%`
}
useIntervalFn(() => query.refresh(), 30000)
</script>
<template>
  <PageHeader
    :title="t('服务器', 'Servers')"
    :description="
      t(
        '来自 Beszel 的独立服务器指标，与网站可用性互不关联。',
        'Independent server metrics from Beszel, separate from website availability.',
      )
    "
    ><button class="button" @click="query.refresh()">
      <RefreshCw :size="14" />{{ t('刷新', 'Refresh') }}</button
    ><button v-if="isAdmin()" class="button primary" @click="configure">
      <Settings :size="14" />{{ t('Beszel 连接', 'Beszel connection') }}
    </button></PageHeader
  >
  <div class="alert-strip">
    <Server :size="16" /><span
      >{{ t('来源', 'Source') }}: {{ query.data.value?.source || 'Beszel' }} ·
      {{ t('最近同步', 'Last sync') }} {{ formatDate(query.data.value?.syncedAt)
      }}<span v-if="query.data.value?.stale"> · {{ t('数据已过期', 'Data is stale') }}</span></span
    >
  </div>
  <div v-if="query.data.value?.error" class="error-banner" role="alert">
    {{ query.data.value.error }}
  </div>
  <AsyncState :pending="query.isPending.value" :error="query.error.value" @retry="query.refresh()"
    ><EmptyState
      v-if="!query.data.value?.items.length"
      :title="t('接入你的 Beszel Hub', 'Connect your Beszel Hub')"
      :description="
        t(
          '使用专用账号读取可见服务器，凭据仅保留在 Go 服务端。',
          'Use a dedicated account to read its visible systems. Credentials stay on the Go server.',
        )
      "
      ><button v-if="isAdmin()" class="button primary" @click="configure">
        {{ t('配置连接', 'Configure connection') }}
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
            <span class="mini-label">{{ t('内存', 'MEMORY') }}</span
            ><strong>{{ percentage(server.memory) }}</strong>
            <div class="metric-progress">
              <span :style="{ width: `${Math.min(server.memory || 0, 100)}%` }" />
            </div>
          </div>
          <div>
            <span class="mini-label">{{ t('磁盘', 'DISK') }}</span
            ><strong>{{ percentage(server.disk) }}</strong>
            <div class="metric-progress">
              <span :style="{ width: `${Math.min(server.disk || 0, 100)}%` }" />
            </div>
          </div>
          <div>
            <span class="mini-label">{{ t('同步时间', 'UPDATED') }}</span>
            <p class="muted" un-text="10px" un-mt="2">{{ formatDate(server.updatedAt) }}</p>
            <span v-if="server.stale" class="certificate-risk">{{
              t('数据过期', 'Stale data')
            }}</span>
          </div>
        </div>
        <div class="section-divider" />
        <button class="button ghost" un-p="0!" @click="detail(server)">
          {{ t('查看历史与容器', 'History & containers') }}<ArrowUpRight :size="14" />
        </button>
      </article></div
  ></AsyncState>
  <p class="note" un-mt="6">
    {{
      t(
        '服务器失联、数据过期或版本不兼容不会改变网站状态。历史范围以 Hub 的实际保留数据为准。',
        'Offline servers, stale data, or incompatible versions do not change website states. History follows the Hub’s actual retention.',
      )
    }}
  </p>
  <Modal v-model:open="configOpen" :title="t('Beszel Hub 连接', 'Beszel Hub connection')"
    ><form id="beszel-form" @submit.prevent="save">
      <p class="note" un-mb="5">
        {{
          t(
            '支持 Beszel 0.20.x，使用可见服务器范围受限的专用只读账号。',
            'Supports Beszel 0.20.x. Use a dedicated read-only account scoped to visible systems.',
          )
        }}
      </p>
      <Field :label="t('Hub URL', 'Hub URL')"
        ><input
          v-model="config.url"
          type="url"
          placeholder="https://beszel.example.com"
          required /></Field
      ><Field :label="t('专用账号邮箱', 'Dedicated account email')" un-mt="4"
        ><input v-model="config.email" type="email" required /></Field
      ><Field :label="t('密码秘密引用', 'Password secret reference')" un-mt="4"
        ><SecretSelect
          v-model="config.passwordSecretId"
          :secrets="secrets.data.value?.items || []" /></Field
      ><Field :label="t('同步间隔（秒）', 'Sync interval (seconds)')" un-mt="4"
        ><input v-model.number="config.pollSeconds" type="number" min="30" required /></Field
      ><Toggle v-model="config.enabled" :label="t('启用接入', 'Enable integration')" un-mt="5" />
      <p v-if="error" class="inline-error">{{ error }}</p>
    </form>
    <template #footer
      ><button class="button" @click="configOpen = false">{{ t('取消', 'Cancel') }}</button
      ><button class="button primary" form="beszel-form" :disabled="saving">
        <Save :size="14" />{{ t('保存连接', 'Save connection') }}
      </button></template
    ></Modal
  ><Modal v-model:open="detailOpen" :title="selected?.name || ''" wide
    ><div v-if="selected?.info" class="hint-grid" un-mb="5">
      <div>
        <span class="mini-label">{{ t('主机名', 'Hostname') }}</span>
        <p>{{ selected.info.hostname || '—' }}</p>
      </div>
      <div>
        <span class="mini-label">CPU</span>
        <p>{{ selected.info.cpuModel || '—' }}</p>
        <p class="muted">
          {{ selected.info.cores }} {{ t('核心', 'cores') }} / {{ selected.info.threads }}
          {{ t('线程', 'threads') }}
        </p>
      </div>
      <div>
        <span class="mini-label">{{ t('系统与 Agent', 'System & agent') }}</span>
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
            t('历史指标', 'History')
          }}</Tabs.Trigger
          ><Tabs.Trigger class="tabs-trigger" value="containers">{{
            t('容器', 'Containers')
          }}</Tabs.Trigger></Tabs.List
        ><Tabs.Content value="history">
          <div un-flex="~ items-center justify-between gap-3" un-mt="5">
            <p class="note">
              {{ t('来源', 'Source') }}: {{ historyMeta?.source || 'Beszel' }} ·
              {{ formatDate(historyMeta?.syncedAt)
              }}<span v-if="historyMeta?.stale"> · {{ t('数据已过期', 'Stale data') }}</span>
            </p>
            <select
              v-model="historyRange"
              :aria-label="t('历史范围', 'History range')"
              @change="selected && detail(selected)"
            >
              <option v-for="range in ['1h', '12h', '24h', '1w', '30d']" :key="range">
                {{ range }}
              </option>
            </select>
          </div>
          <p v-if="historyMeta?.error" class="inline-error">{{ historyMeta.error }}</p>
          <EmptyState
            v-if="!history.length"
            :title="t('Hub 未返回历史数据', 'No history returned by the Hub')"
          />
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
                <h3>{{ t('内存', 'Memory') }} (%)</h3>
                <Sparkline
                  show-scale
                  :values="history.map((x) => x.memory)"
                  :timestamps="history.map((x) => x.at)"
                  :height="115"
                />
              </section>
              <section>
                <h3>{{ t('磁盘', 'Disk') }} (%)</h3>
                <Sparkline
                  show-scale
                  :values="history.map((x) => x.disk)"
                  :timestamps="history.map((x) => x.at)"
                  :height="115"
                />
              </section>
              <section>
                <h3>{{ t('网络接收', 'Network received') }} (MiB/s)</h3>
                <Sparkline
                  show-scale
                  :values="history.map((x) => x.networkIn / 1048576)"
                  :timestamps="history.map((x) => x.at)"
                  :height="115"
                />
              </section>
              <section>
                <h3>{{ t('网络发送', 'Network sent') }} (MiB/s)</h3>
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
            {{ t('来源', 'Source') }}: {{ containersMeta?.source || 'Beszel' }} ·
            {{ formatDate(containersMeta?.syncedAt)
            }}<span v-if="containersMeta?.stale"> · {{ t('数据已过期', 'Stale data') }}</span>
          </p>
          <p v-if="containersMeta?.error" class="inline-error">{{ containersMeta.error }}</p>
          <EmptyState
            v-if="!containers.length"
            :title="t('没有可见容器数据', 'No visible container data')"
          />
          <div v-else class="table-wrap" un-mt="5">
            <table class="data-table">
              <thead>
                <tr>
                  <th>{{ t('容器', 'Container') }}</th>
                  <th>{{ t('状态', 'Status') }}</th>
                  <th>CPU</th>
                  <th>{{ t('内存', 'Memory') }} (MiB)</th>
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
                  <td>{{ container.memory.toFixed(1) }} MiB</td>
                </tr>
              </tbody>
            </table>
          </div></Tabs.Content
        ></Tabs.Root
      ></AsyncState
    ></Modal
  >
</template>
