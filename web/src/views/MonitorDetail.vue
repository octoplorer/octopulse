<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  Play,
  Pause,
  Pencil,
  Trash2,
  RefreshCw,
  KeyRound,
  Copy,
  ArrowLeft,
  Clock,
  Activity,
  ShieldCheck,
  CheckCircle2,
} from '@lucide/vue'
import { Tabs } from '@ark-ui/vue/tabs'
import { useRecord } from '../lib/data'
import { api, canEdit } from '../lib/api'
import type { Monitor, MonitorHistory, Round } from '../lib/types'
import { targetOf } from '../lib/monitor'
import { t, formatDate, formatPercent, duration } from '../lib/preferences'
import { notify, errorText } from '../lib/notices'
import { useIntervalFn, useClipboard } from '@vueuse/core'
import PageHeader from '../components/PageHeader.vue'
import StateBadge from '../components/StateBadge.vue'
import AsyncState from '../components/AsyncState.vue'
import Sparkline from '../components/Sparkline.vue'
import EmptyState from '../components/EmptyState.vue'
import Modal from '../components/Modal.vue'
const route = useRoute(),
  router = useRouter(),
  query = useRecord<Monitor>(() => `monitors/${route.params.id}`),
  period = ref('24h'),
  from = () =>
    Date.now() - (period.value === '7d' ? 7 : period.value === '30d' ? 30 : 1) * 86400000,
  history = useRecord<MonitorHistory>(
    () => `monitors/${route.params.id}/history?from=${from()}&to=${Date.now()}`,
  ),
  busy = ref(false),
  confirmDelete = ref(false),
  selectedRound = ref<Round | null>(null),
  diagnosticsOpen = ref(false),
  heartbeatToken = ref(''),
  heartbeatUrl = ref(''),
  tab = ref('history')
const { copy, copied } = useClipboard(),
  monitor = computed(() => query.data.value),
  availability = computed(() => history.data.value?.availability)
useIntervalFn(() => {
  query.refresh()
  history.refresh()
}, 30000)
async function act(action: 'check' | 'toggle' | 'rotate' | 'delete') {
  if (!monitor.value) return
  busy.value = true
  try {
    if (action === 'check') {
      await api(`monitors/${monitor.value.id}/check`, { method: 'POST' })
      notify(
        t(
          '检查请求已接受，结果将在完成后更新。',
          'Check request accepted. Results update when the round finishes.',
        ),
      )
    } else if (action === 'toggle') {
      await api(`monitors/${monitor.value.id}`, {
        method: 'PATCH',
        body: { ...monitor.value, enabled: !monitor.value.enabled },
      })
      notify(t('监控状态已更新', 'Monitor updated'))
    } else if (action === 'rotate') {
      const data = await api<{ token: string; url: string }>(
        `monitors/${monitor.value.id}/heartbeat/rotate`,
        { method: 'POST' },
      )
      heartbeatToken.value = data.token
      heartbeatUrl.value = new URL(
        data.url || `/api/heartbeat/${monitor.value.id}/${data.token}`,
        location.origin,
      ).href
      notify(t('新密钥只在此显示一次，请保存。', 'Save this token. It is shown only once.'))
    } else {
      await api(`monitors/${monitor.value.id}`, { method: 'DELETE' })
      notify(t('监控项已删除', 'Monitor deleted'))
      router.push('/app/monitors')
      return
    }
    await query.refresh()
    await history.refresh()
  } catch (e) {
    notify(errorText(e), 'error')
  } finally {
    busy.value = false
  }
}
function viewRound(round: Round) {
  selectedRound.value = round
  diagnosticsOpen.value = true
}
</script>
<template>
  <AsyncState :pending="query.isPending.value" :error="query.error.value" @retry="query.refresh()"
    ><template v-if="monitor"
      ><PageHeader :title="monitor.name" :description="targetOf(monitor)"
        ><RouterLink to="/app/monitors" class="button ghost"
          ><ArrowLeft :size="14" />{{ t('列表', 'All monitors') }}</RouterLink
        ><template v-if="canEdit()"
          ><button class="button" :disabled="busy" @click="act('toggle')">
            <Pause v-if="monitor.enabled" :size="14" /><Play v-else :size="14" />{{
              monitor.enabled ? t('暂停', 'Pause') : t('启用', 'Enable')
            }}</button
          ><RouterLink :to="`/app/monitors/${monitor.id}/edit`" class="button"
            ><Pencil :size="14" />{{ t('编辑', 'Edit') }}</RouterLink
          ><button
            v-if="monitor.type !== 'heartbeat'"
            class="button primary"
            :disabled="busy || !monitor.enabled"
            @click="act('check')"
          >
            <RefreshCw :size="14" />{{ t('立即检查', 'Check now') }}
          </button></template
        ></PageHeader
      >
      <div class="stats-grid">
        <div class="card stat-card">
          <div class="stat-label">{{ t('当前状态', 'Current state') }}<Activity :size="15" /></div>
          <div un-mt="5" un-mb="4">
            <StateBadge
              :state="monitor.type === 'certificate' ? monitor.certificate?.state : monitor.state"
              :paused="!monitor.enabled"
            />
          </div>
          <p class="stat-meta">{{ formatDate(monitor.lastCheckedAt) }}</p>
        </div>
        <div class="card stat-card">
          <div class="stat-label">
            {{
              monitor.certificate
                ? t('证书到期时间', 'Certificate expires')
                : t('有效时长可用率', 'Duration uptime')
            }}<CheckCircle2 :size="15" />
          </div>
          <div class="stat-value" :style="monitor.certificate ? { fontSize: '16px' } : undefined">
            {{
              monitor.certificate
                ? formatDate(monitor.certificate.expiresAt)
                : formatPercent(availability?.uptime)
            }}
          </div>
          <p class="stat-meta">
            <template v-if="monitor.certificate">{{
              t('证书风险不计入可用率', 'Certificate risk is excluded from uptime')
            }}</template
            ><template v-else
              >{{ t('有效统计时长', 'Effective duration') }}
              {{ duration(availability?.effectiveMs) }}</template
            >
          </p>
        </div>
        <div class="card stat-card">
          <div class="stat-label">
            {{
              monitor.certificate
                ? t('剩余有效期', 'Days remaining')
                : t('观测覆盖率', 'Observation coverage')
            }}<ShieldCheck :size="15" />
          </div>
          <div class="stat-value">
            {{
              monitor.certificate
                ? monitor.certificate.expiresAt
                  ? monitor.certificate.daysRemaining.toFixed(1)
                  : '—'
                : formatPercent(availability?.coverage)
            }}
          </div>
          <p class="stat-meta">
            {{
              monitor.certificate
                ? t('提醒阈值（天）', 'Warning thresholds (days)') +
                  ': ' +
                  (monitor.certificate.warningDays || []).join(', ')
                : t('不将缺失数据计作正常', 'Missing data never counts as Up')
            }}
          </p>
        </div>
        <div class="card stat-card">
          <div class="stat-label">{{ t('最近延迟', 'Latest latency') }}<Clock :size="15" /></div>
          <div class="stat-value">
            {{ duration(history.data.value?.latency?.at(-1)?.latencyMs) }}
          </div>
          <p class="stat-meta">
            {{
              monitor.type === 'heartbeat'
                ? monitor.heartbeat?.periodSeconds
                : monitor.intervalSeconds
            }}
            s
            {{
              monitor.type === 'heartbeat'
                ? t('预期周期', 'expected period')
                : t('检查间隔', 'check interval')
            }}
          </p>
        </div>
      </div>
      <div v-if="monitor.type === 'heartbeat'" class="card card-body" un-mb="6">
        <div un-flex="~ items-center justify-between gap-4">
          <div>
            <h2>{{ t('心跳上报', 'Heartbeat reporting') }}</h2>
            <p class="muted" un-mt="2">
              {{
                t(
                  '密钥不会回读。轮换后旧密钥立即失效。',
                  'Tokens cannot be read back. Rotation invalidates the previous token.',
                )
              }}
            </p>
          </div>
          <button v-if="canEdit()" class="button" :disabled="busy" @click="act('rotate')">
            <KeyRound :size="14" />{{ t('生成 / 轮换密钥', 'Generate / rotate token') }}
          </button>
        </div>
        <div v-if="monitor.heartbeat?.lastReceivedAt" class="note" un-mt="4">
          {{ t('最近上报', 'Last report') }} {{ formatDate(monitor.heartbeat.lastReceivedAt) }} ·
          {{ monitor.heartbeat.lastSuccess ? t('成功', 'Up') : t('失败', 'Down') }}
          <p v-if="monitor.heartbeat.description" un-mt="2">{{ monitor.heartbeat.description }}</p>
        </div>
        <div v-if="heartbeatToken" class="heartbeat-url">
          <code>{{ heartbeatUrl }}</code
          ><button class="button small ghost" @click="copy(heartbeatUrl)">
            <Copy :size="13" />{{ copied ? t('已复制', 'Copied') : t('复制', 'Copy') }}
          </button>
          <p class="muted" un-mt="3">
            {{
              t(
                'POST 上报：成功 {"status":"up"}；失败 {"status":"down"}，可附 description。',
                'POST reports: {"status":"up"} or {"status":"down"}, with optional description.',
              )
            }}
          </p>
        </div>
      </div>
      <div v-if="monitor.certificate" class="alert-strip">
        <ShieldCheck :size="16" /><span
          >{{ t('证书到期时间', 'Certificate expires') }}
          {{ formatDate(monitor.certificate.expiresAt) }} · {{ t('剩余', 'Remaining') }}
          {{ monitor.certificate.expiresAt ? monitor.certificate.daysRemaining.toFixed(1) : '—' }}
          {{ t('天', 'days') }} ·
          {{ t('证书风险不计入可用率', 'Certificate risk is excluded from uptime') }}</span
        >
      </div>
      <div class="detail-layout">
        <section class="card">
          <Tabs.Root v-model="tab"
            ><Tabs.List class="tabs-list"
              ><Tabs.Trigger class="tabs-trigger" value="history">{{
                t('历史与趋势', 'History & trends')
              }}</Tabs.Trigger
              ><Tabs.Trigger class="tabs-trigger" value="configuration">{{
                t('配置摘要', 'Configuration')
              }}</Tabs.Trigger>
              <div un-ml="auto" class="historical-period">
                <select
                  v-model="period"
                  un-w="auto!"
                  :aria-label="t('统计窗口', 'Statistics window')"
                >
                  <option value="24h">24 {{ t('小时', 'hours') }}</option>
                  <option value="7d">7 {{ t('天', 'days') }}</option>
                  <option value="30d">30 {{ t('天', 'days') }}</option>
                </select>
              </div></Tabs.List
            ><Tabs.Content value="history"
              ><div class="card-body">
                <div un-flex="~ justify-between items-center" un-mb="4">
                  <h3>{{ t('响应延迟', 'Response latency') }}</h3>
                  <span class="mini-label">ms</span>
                </div>
                <Sparkline
                  show-scale
                  :values="history.data.value?.latency?.map((p) => p.latencyMs) || []"
                  :timestamps="history.data.value?.latency?.map((p) => p.at) || []"
                  :height="125"
                />
                <p class="field-hint" un-mt="2">
                  {{
                    t(
                      '来源：实际检查轮次；缺失观测不插值。',
                      'Source: actual check rounds. Missing observations are not interpolated.',
                    )
                  }}
                </p>
              </div>
              <AsyncState
                :pending="history.isPending.value"
                :error="history.error.value"
                @retry="history.refresh()"
                ><EmptyState
                  v-if="!history.data.value?.rounds.length"
                  :title="t('暂无检查记录', 'No check records yet')"
                  :description="
                    t(
                      '首次检查完成后显示轮次与诊断。',
                      'Rounds and diagnostics appear after the first check.',
                    )
                  "
                />
                <div v-else class="table-wrap">
                  <table class="data-table">
                    <thead>
                      <tr>
                        <th>{{ t('检查时间', 'Checked at') }}</th>
                        <th>{{ t('结果', 'Result') }}</th>
                        <th>{{ t('尝试', 'Attempts') }}</th>
                        <th>{{ t('耗时', 'Duration') }}</th>
                        <th />
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="round in history.data.value.rounds" :key="round.id">
                        <td class="muted" un-text="10px">{{ formatDate(round.startedAt) }}</td>
                        <td><StateBadge :state="round.success ? 'up' : 'down'" /></td>
                        <td>{{ round.attempts?.length || 0 }}</td>
                        <td>{{ duration(round.latencyMs) }}</td>
                        <td>
                          <button class="button small ghost" @click="viewRound(round)">
                            {{ t('诊断', 'Diagnostics') }}
                          </button>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div></AsyncState
              ></Tabs.Content
            ><Tabs.Content value="configuration"
              ><div class="card-body">
                <pre class="json-output">{{ JSON.stringify(monitor, null, 2) }}</pre>
              </div></Tabs.Content
            ></Tabs.Root
          >
        </section>
        <aside>
          <section class="card card-body">
            <h2 un-mb="6">{{ t('监控信息', 'Monitor information') }}</h2>
            <dl class="definition-list">
              <div>
                <dt>{{ t('类型', 'Type') }}</dt>
                <dd>{{ monitor.type.toUpperCase() }}</dd>
              </div>
              <div>
                <dt>{{ t('分组', 'Group') }}</dt>
                <dd>{{ monitor.group || '—' }}</dd>
              </div>
              <div v-if="monitor.type !== 'heartbeat'">
                <dt>{{ t('检查间隔', 'Check interval') }}</dt>
                <dd>{{ monitor.intervalSeconds }} s</dd>
              </div>
              <div v-if="monitor.type !== 'heartbeat'">
                <dt>{{ t('探测超时', 'Attempt timeout') }}</dt>
                <dd>{{ monitor.timeoutSeconds }} s</dd>
              </div>
              <div v-if="['http', 'tcp', 'dns'].includes(monitor.type)">
                <dt>{{ t('追加重试上限', 'Additional retry limit') }}</dt>
                <dd>{{ monitor.retries }}</dd>
              </div>
              <div v-if="['http', 'tcp', 'dns'].includes(monitor.type)">
                <dt>{{ t('失败 / 恢复阈值', 'Failure / recovery threshold') }}</dt>
                <dd>{{ monitor.failureThreshold }} / {{ monitor.recoveryThreshold }}</dd>
              </div>
              <template v-if="monitor.heartbeat"
                ><div>
                  <dt>{{ t('预期周期', 'Expected period') }}</dt>
                  <dd>{{ monitor.heartbeat.periodSeconds }} s</dd>
                </div>
                <div>
                  <dt>{{ t('宽限时间', 'Grace period') }}</dt>
                  <dd>{{ monitor.heartbeat.graceSeconds }} s</dd>
                </div></template
              >
              <div>
                <dt>{{ t('创建时间', 'Created') }}</dt>
                <dd>{{ formatDate(monitor.createdAt) }}</dd>
              </div>
            </dl>
            <p v-if="monitor.description" class="note" un-mt="6">{{ monitor.description }}</p>
            <div un-flex="~ wrap gap-2" un-mt="4">
              <span v-for="tag in monitor.tags" :key="tag" class="pill">{{ tag }}</span>
            </div>
          </section>
          <p class="note" un-mt="5">
            {{
              t(
                '可用率基于确认后的状态时长。Unknown、暂停和维护时间被排除，零有效数据时不会显示 100%。',
                'Uptime uses confirmed state duration. Unknown, paused, and maintenance time are excluded; zero observations never show 100%.',
              )
            }}
          </p>
          <button v-if="canEdit()" class="button danger" un-mt="5" @click="confirmDelete = true">
            <Trash2 :size="14" />{{ t('删除监控项', 'Delete monitor') }}
          </button>
        </aside>
      </div></template
    ></AsyncState
  ><Modal
    v-model:open="confirmDelete"
    :title="t('删除监控项', 'Delete monitor')"
    :description="
      t(
        '此操作会删除监控配置。请先确认状态页和通知中不再需要它。',
        'This deletes the monitor configuration. Confirm it is no longer needed on status pages or notifications.',
      )
    "
    ><template #footer
      ><button class="button" @click="confirmDelete = false">{{ t('取消', 'Cancel') }}</button
      ><button class="button danger" :disabled="busy" @click="act('delete')">
        {{ t('确认删除', 'Delete') }}
      </button></template
    ></Modal
  ><Modal v-model:open="diagnosticsOpen" :title="t('检查轮次诊断', 'Round diagnostics')" wide
    ><template v-if="selectedRound"
      ><p class="muted" un-mb="4">
        {{ formatDate(selectedRound.startedAt) }} · {{ duration(selectedRound.latencyMs) }}
      </p>
      <div
        v-for="attempt in selectedRound.attempts"
        :key="attempt.number"
        class="card card-body"
        un-mb="4"
      >
        <div un-flex="~ items-center justify-between">
          <h3>{{ t('尝试', 'Attempt') }} {{ attempt.number }}</h3>
          <StateBadge :state="attempt.success ? 'up' : 'down'" />
        </div>
        <p v-if="attempt.error" class="inline-error">{{ attempt.error }}</p>
        <pre class="json-output" un-mt="4">{{ JSON.stringify(attempt.detail, null, 2) }}</pre>
      </div></template
    ></Modal
  >
</template>
