<script setup lang="ts">
import type { Round } from '../../../../../client/types.gen'
import { Tabs } from '@ark-ui/vue/tabs'
import { useMutation, useQuery } from '@pinia/colada'
import { useClipboard } from '@vueuse/core'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import {
  checkMonitorMutation,
  deleteMonitorMutation,
  getMonitorHistoryQuery,
  getMonitorHistoryQueryKey,
  getMonitorQuery,
  rotateHeartbeatMutation,
  updateMonitorMutation,
} from '../../../../../client/@pinia/colada.gen'
import AsyncState from '../../../../../components/AsyncState.vue'
import EmptyState from '../../../../../components/EmptyState.vue'
import Modal from '../../../../../components/Modal.vue'
import PageHeader from '../../../../../components/PageHeader.vue'
import Sparkline from '../../../../../components/Sparkline.vue'
import StateBadge from '../../../../../components/StateBadge.vue'
import { canEdit } from '../../../../../composables/api'
import { notify } from '../../../../../composables/notices'
import { usePollingEnabled } from '../../../../../composables/polling'
import { duration, formatDate, formatPercent } from '../../../../../composables/preferences'
import { errorText } from '../../../../../lib/errors'
import { targetOf } from '../../../../../lib/monitor'

const { t, n } = useI18n({ useScope: 'global' })

definePage({ meta: { title: 'navigation.monitorDetails' } })

const checkMonitor = useMutation(checkMonitorMutation())
const updateMonitor = useMutation(updateMonitorMutation())
const rotateHeartbeat = useMutation(rotateHeartbeatMutation())
const deleteMonitor = useMutation(deleteMonitorMutation())

const route = useRoute('/app/(admin)/monitors/[id]/')
const router = useRouter()
const pollingEnabled = usePollingEnabled()
const query = useQuery(
  () =>
    ({
      ...getMonitorQuery({ path: { id: route.params.id } }),
      staleTime: 5000,
      enabled: pollingEnabled.value,
      autoRefetch: 30000,
    }),
)
const period = ref('24h')
const history = useQuery(
  () => {
    const id = route.params.id
    const days = period.value === '7d' ? 7 : period.value === '30d' ? 30 : 1
    return {
      // Cache the relative range; compute its absolute window for each request.
      key: [...getMonitorHistoryQueryKey({ path: { id } }), { period: period.value }],
      query: (context) => {
        const to = Date.now()
        return getMonitorHistoryQuery({
          path: { id },
          query: { from: to - days * 86400000, to },
        }).query(context)
      },
      staleTime: 5000,
      enabled: pollingEnabled.value,
      autoRefetch: 30000,
    }
  },
)
const busy = ref(false)
const confirmDelete = ref(false)
const selectedRound = ref<Round | null>(null)
const diagnosticsOpen = ref(false)
const heartbeatToken = ref('')
const heartbeatUrl = ref('')
const tab = ref('history')
const { copy, copied } = useClipboard()
const monitor = computed(() => query.data.value)
const availability = computed(() => history.data.value?.availability)
function refreshHistory() {
  return history.refetch()
}
async function act(action: 'check' | 'toggle' | 'rotate' | 'delete') {
  if (!monitor.value)
    return
  busy.value = true
  try {
    if (action === 'check') {
      await checkMonitor.mutateAsync({ path: { id: monitor.value.id } })
      notify(t('monitorDetails.checkRequestAcceptedResultsUpdateWhenTheRound'))
    }
    else if (action === 'toggle') {
      await updateMonitor.mutateAsync({
        path: { id: monitor.value.id },
        body: { ...monitor.value, enabled: !monitor.value.enabled },
      })
      notify(t('monitorDetails.monitorUpdated'))
    }
    else if (action === 'rotate') {
      const data = await rotateHeartbeat.mutateAsync({ path: { id: monitor.value.id } })
      heartbeatToken.value = data.token
      heartbeatUrl.value = new URL(
        data.url || `/api/heartbeat/${monitor.value.id}/${data.token}`,
        location.origin,
      ).href
      notify(t('monitorDetails.saveThisTokenItIsShownOnlyOnce'))
    }
    else {
      await deleteMonitor.mutateAsync({ path: { id: monitor.value.id } })
      notify(t('monitorDetails.monitorDeleted'))
      router.push('/app/monitors')
      return
    }
    await query.refresh()
    await refreshHistory()
  }
  catch (e) {
    notify(errorText(e), 'error')
  }
  finally {
    busy.value = false
  }
}
function viewRound(round: Round) {
  selectedRound.value = round
  diagnosticsOpen.value = true
}
</script>

<template>
  <AsyncState :pending="query.isPending.value" :error="query.error.value" @retry="query.refetch()">
    <template v-if="monitor">
      <PageHeader :title="monitor.name" :description="targetOf(monitor)">
        <RouterLink to="/app/monitors" class="button ghost">
          <span class="i-lucide-arrow-left" un-w="14px" un-h="14px" aria-hidden="true" />{{ t('monitorDetails.allMonitors') }}
        </RouterLink><template v-if="canEdit()">
          <button class="button" :disabled="busy" @click="act('toggle')">
            <span v-if="monitor.enabled" class="i-lucide-pause" un-w="14px" un-h="14px" aria-hidden="true" /><span v-else class="i-lucide-play" un-w="14px" un-h="14px" aria-hidden="true" />{{
              monitor.enabled ? t('monitorDetails.pause') : t('monitorDetails.enable')
            }}
          </button><RouterLink :to="`/app/monitors/${monitor.id}/edit`" class="button">
            <span class="i-lucide-pencil" un-w="14px" un-h="14px" aria-hidden="true" />{{ t('common.edit') }}
          </RouterLink><button
            v-if="monitor.type !== 'heartbeat'"
            class="button primary"
            :disabled="busy || !monitor.enabled"
            @click="act('check')"
          >
            <span class="i-lucide-refresh-cw" un-w="14px" un-h="14px" aria-hidden="true" />{{ t('monitorDetails.checkNow') }}
          </button>
        </template>
      </PageHeader>
      <div class="stats-grid">
        <div class="card stat-card">
          <div class="stat-label">
            {{ t('monitorDetails.currentState') }}<span class="i-lucide-activity" un-w="15px" un-h="15px" aria-hidden="true" />
          </div>
          <div un-mt="5" un-mb="4">
            <StateBadge
              :state="monitor.type === 'certificate' ? monitor.certificate?.state : monitor.state"
              :paused="!monitor.enabled"
            />
          </div>
          <p class="stat-meta">
            {{ formatDate(monitor.lastCheckedAt) }}
          </p>
        </div>
        <div class="card stat-card">
          <div class="stat-label">
            {{
              monitor.certificate
                ? t('common.certificateExpires')
                : t('monitorDetails.durationUptime')
            }}<span class="i-lucide-circle-check" un-w="15px" un-h="15px" aria-hidden="true" />
          </div>
          <div class="stat-value" :style="monitor.certificate ? { fontSize: '16px' } : undefined">
            {{
              monitor.certificate
                ? formatDate(monitor.certificate.expiresAt)
                : formatPercent(availability?.uptime)
            }}
          </div>
          <p class="stat-meta">
            <template v-if="monitor.certificate">
              {{
                t('common.certificateRiskIsExcludedFromUptime')
              }}
            </template><template v-else>
              {{ t('monitorDetails.effectiveDuration') }}
              {{ duration(availability?.effectiveMs) }}
            </template>
          </p>
        </div>
        <div class="card stat-card">
          <div class="stat-label">
            {{
              monitor.certificate
                ? t('monitorDetails.daysRemaining')
                : t('monitorDetails.observationCoverage')
            }}<span class="i-lucide-shield-check" un-w="15px" un-h="15px" aria-hidden="true" />
          </div>
          <div class="stat-value">
            {{
              monitor.certificate
                ? monitor.certificate.expiresAt
                  ? n(monitor.certificate.daysRemaining, 'decimal')
                  : '—'
                : formatPercent(availability?.coverage)
            }}
          </div>
          <p class="stat-meta">
            {{
              monitor.certificate
                ? t('monitorDetails.warningThresholdsSummary', {
                  thresholds: (monitor.certificate.warningDays || [])
                    .map((days) => n(days))
                    .join(', '),
                })
                : t('monitorDetails.missingDataNeverCountsAsUp')
            }}
          </p>
        </div>
        <div class="card stat-card">
          <div class="stat-label">
            {{ t('monitorDetails.latestLatency') }}<span class="i-lucide-clock" un-w="15px" un-h="15px" aria-hidden="true" />
          </div>
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
                ? t('monitorDetails.expectedPeriod')
                : t('monitorDetails.checkInterval')
            }}
          </p>
        </div>
      </div>
      <div v-if="monitor.type === 'heartbeat'" class="card card-body" un-mb="6">
        <div un-flex="~ items-center justify-between gap-4">
          <div>
            <h2>{{ t('monitorDetails.heartbeatReporting') }}</h2>
            <p class="muted" un-mt="2">
              {{ t('monitorDetails.tokensCannotBeReadBackRotationInvalidatesThe') }}
            </p>
          </div>
          <button v-if="canEdit()" class="button" :disabled="busy" @click="act('rotate')">
            <span class="i-lucide-key-round" un-w="14px" un-h="14px" aria-hidden="true" />{{ t('monitorDetails.generateRotateToken') }}
          </button>
        </div>
        <div v-if="monitor.heartbeat?.lastReceivedAt" class="note" un-mt="4">
          {{ t('monitorDetails.lastReport') }} {{ formatDate(monitor.heartbeat.lastReceivedAt) }} ·
          {{ monitor.heartbeat.lastSuccess ? t('monitorDetails.up') : t('monitorDetails.down') }}
          <p v-if="monitor.heartbeat.description" un-mt="2">
            {{ monitor.heartbeat.description }}
          </p>
        </div>
        <div v-if="heartbeatToken" class="heartbeat-url">
          <code>{{ heartbeatUrl }}</code><button class="button small ghost" @click="copy(heartbeatUrl)">
            <span class="i-lucide-copy" un-w="13px" un-h="13px" aria-hidden="true" />{{ copied ? t('monitorDetails.copied') : t('monitorDetails.copy') }}
          </button>
          <p class="muted" un-mt="3">
            {{ t('monitorDetails.postReportsStatusUpOrStatusDownWith') }}
          </p>
        </div>
      </div>
      <div v-if="monitor.certificate" class="alert-strip">
        <span class="i-lucide-shield-check" un-w="16px" un-h="16px" aria-hidden="true" /><span>{{ t('common.certificateExpires') }} {{ formatDate(monitor.certificate.expiresAt) }} ·
          {{
            t('monitorDetails.remainingDays', {
              days: monitor.certificate.expiresAt
                ? n(monitor.certificate.daysRemaining, 'decimal')
                : '—',
            })
          }}
          · {{ t('common.certificateRiskIsExcludedFromUptime') }}</span>
      </div>
      <div class="detail-layout">
        <section class="card">
          <Tabs.Root v-model="tab">
            <Tabs.List class="tabs-list">
              <Tabs.Trigger class="tabs-trigger" value="history">
                {{
                  t('monitorDetails.historyTrends')
                }}
              </Tabs.Trigger><Tabs.Trigger class="tabs-trigger" value="configuration">
                {{
                  t('monitorDetails.configuration')
                }}
              </Tabs.Trigger>
              <div un-ml="auto" class="historical-period">
                <select
                  v-model="period"
                  un-w="auto!"
                  :aria-label="t('monitorDetails.statisticsWindow')"
                >
                  <option value="24h">
                    {{ t('counts.hours', { count: 24 }, 24) }}
                  </option>
                  <option value="7d">
                    {{ t('counts.days', { count: 7 }, 7) }}
                  </option>
                  <option value="30d">
                    {{ t('counts.days', { count: 30 }, 30) }}
                  </option>
                </select>
              </div>
            </Tabs.List><Tabs.Content value="history">
              <div class="card-body">
                <div un-flex="~ justify-between items-center" un-mb="4">
                  <h3>{{ t('monitorDetails.responseLatency') }}</h3>
                  <span class="mini-label">ms</span>
                </div>
                <Sparkline
                  show-scale
                  :values="history.data.value?.latency?.map((p) => p.latencyMs) || []"
                  :timestamps="history.data.value?.latency?.map((p) => p.at) || []"
                  :height="125"
                />
                <p class="field-hint" un-mt="2">
                  {{ t('monitorDetails.sourceActualCheckRoundsMissingObservationsAreNot') }}
                </p>
              </div>
              <AsyncState
                :pending="history.isPending.value"
                :error="history.error.value"
                @retry="refreshHistory()"
              >
                <EmptyState
                  v-if="!history.data.value?.rounds.length"
                  :title="t('monitorDetails.noCheckRecordsYet')"
                  :description="t('monitorDetails.roundsAndDiagnosticsAppearAfterTheFirstCheck')"
                />
                <div v-else class="table-wrap">
                  <table class="data-table">
                    <thead>
                      <tr>
                        <th>{{ t('monitorDetails.checkedAt') }}</th>
                        <th>{{ t('monitorDetails.result') }}</th>
                        <th>{{ t('monitorDetails.attempts') }}</th>
                        <th>{{ t('monitorDetails.duration') }}</th>
                        <th />
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="round in history.data.value.rounds" :key="round.id">
                        <td class="muted" un-text="10px">
                          {{ formatDate(round.startedAt) }}
                        </td>
                        <td><StateBadge :state="round.success ? 'up' : 'down'" /></td>
                        <td>{{ round.attempts?.length || 0 }}</td>
                        <td>{{ duration(round.latencyMs) }}</td>
                        <td>
                          <button class="button small ghost" @click="viewRound(round)">
                            {{ t('monitorDetails.diagnostics') }}
                          </button>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </AsyncState>
            </Tabs.Content><Tabs.Content value="configuration">
              <div class="card-body">
                <pre class="json-output">{{ JSON.stringify(monitor, null, 2) }}</pre>
              </div>
            </Tabs.Content>
          </Tabs.Root>
        </section>
        <aside>
          <section class="card card-body">
            <h2 un-mb="6">
              {{ t('monitorDetails.monitorInformation') }}
            </h2>
            <dl class="definition-list">
              <div>
                <dt>{{ t('monitorDetails.type') }}</dt>
                <dd>{{ monitor.type.toUpperCase() }}</dd>
              </div>
              <div>
                <dt>{{ t('common.group') }}</dt>
                <dd>{{ monitor.group || '—' }}</dd>
              </div>
              <div v-if="monitor.type !== 'heartbeat'">
                <dt>{{ t('monitorDetails.checkInterval2') }}</dt>
                <dd>{{ monitor.intervalSeconds }} s</dd>
              </div>
              <div v-if="monitor.type !== 'heartbeat'">
                <dt>{{ t('monitorDetails.attemptTimeout') }}</dt>
                <dd>{{ monitor.timeoutSeconds }} s</dd>
              </div>
              <div v-if="['http', 'tcp', 'dns'].includes(monitor.type)">
                <dt>{{ t('monitorDetails.additionalRetryLimit') }}</dt>
                <dd>{{ monitor.retries }}</dd>
              </div>
              <div v-if="['http', 'tcp', 'dns'].includes(monitor.type)">
                <dt>{{ t('monitorDetails.failureRecoveryThreshold') }}</dt>
                <dd>{{ monitor.failureThreshold }} / {{ monitor.recoveryThreshold }}</dd>
              </div>
              <template v-if="monitor.heartbeat">
                <div>
                  <dt>{{ t('monitorDetails.expectedPeriod2') }}</dt>
                  <dd>{{ monitor.heartbeat.periodSeconds }} s</dd>
                </div>
                <div>
                  <dt>{{ t('monitorDetails.gracePeriod') }}</dt>
                  <dd>{{ monitor.heartbeat.graceSeconds }} s</dd>
                </div>
              </template>
              <div>
                <dt>{{ t('common.created') }}</dt>
                <dd>{{ formatDate(monitor.createdAt) }}</dd>
              </div>
            </dl>
            <p v-if="monitor.description" class="note" un-mt="6">
              {{ monitor.description }}
            </p>
            <div un-flex="~ wrap gap-2" un-mt="4">
              <span v-for="tag in monitor.tags" :key="tag" class="pill">{{ tag }}</span>
            </div>
          </section>
          <p class="note" un-mt="5">
            {{ t('monitorDetails.uptimeUsesConfirmedStateDurationUnknownPausedAnd') }}
          </p>
          <button v-if="canEdit()" class="button danger" un-mt="5" @click="confirmDelete = true">
            <span class="i-lucide-trash-2" un-w="14px" un-h="14px" aria-hidden="true" />{{ t('common.deleteMonitor') }}
          </button>
        </aside>
      </div>
    </template>
  </AsyncState><Modal
    v-model:open="confirmDelete"
    :title="t('common.deleteMonitor')"
    :description="t('monitorDetails.thisDeletesTheMonitorConfigurationConfirmItIs')"
  >
    <template #footer>
      <button class="button" @click="confirmDelete = false">
        {{ t('common.cancel') }}
      </button><button class="button danger" :disabled="busy" @click="act('delete')">
        {{ t('monitorDetails.delete') }}
      </button>
    </template>
  </Modal><Modal v-model:open="diagnosticsOpen" :title="t('monitorDetails.roundDiagnostics')" wide>
    <template v-if="selectedRound">
      <p class="muted" un-mb="4">
        {{ formatDate(selectedRound.startedAt) }} · {{ duration(selectedRound.latencyMs) }}
      </p>
      <div
        v-for="attempt in selectedRound.attempts"
        :key="attempt.number"
        class="card card-body"
        un-mb="4"
      >
        <div un-flex="~ items-center justify-between">
          <h3>{{ t('monitorDetails.attemptNumber', { number: n(attempt.number) }) }}</h3>
          <StateBadge :state="attempt.success ? 'up' : 'down'" />
        </div>
        <p v-if="attempt.error" class="inline-error">
          {{ attempt.error }}
        </p>
        <pre class="json-output" un-mt="4">{{ JSON.stringify(attempt.detail, null, 2) }}</pre>
      </div>
    </template>
  </Modal>
</template>
