<script setup lang="ts">
import type { Round } from '../../../../../client/types.gen'
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
import EChart from '../../../../../components/EChart.vue'
import EmptyState from '../../../../../components/EmptyState.vue'
import Modal from '../../../../../components/Modal.vue'
import PageHeader from '../../../../../components/PageHeader.vue'
import StateBadge from '../../../../../components/StateBadge.vue'
import { Alert } from '../../../../../components/ui/alert'
import { Badge } from '../../../../../components/ui/badge'
import { Button } from '../../../../../components/ui/button'
import { Card, CardContent } from '../../../../../components/ui/card'
import { FieldDescription, FieldError } from '../../../../../components/ui/field'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '../../../../../components/ui/select'
import { Table, TableBody, TableCell, TableContainer, TableHead, TableHeader, TableRow } from '../../../../../components/ui/table'
import { TabsContent, TabsList, TabsRoot, TabsTrigger } from '../../../../../components/ui/tabs'
import { canEdit } from '../../../../../composables/api'
import { notify } from '../../../../../composables/notices'
import { usePollingEnabled } from '../../../../../composables/polling'
import { duration, formatDate, formatPercent, timezone } from '../../../../../composables/preferences'
import { normalizeTimeSeries, timeSeriesOption } from '../../../../../lib/chart'
import { errorText } from '../../../../../lib/errors'
import { targetOf } from '../../../../../lib/monitor'

const { t, n, d } = useI18n({ useScope: 'global' })

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
const latencyPoints = computed(() => normalizeTimeSeries(
  (history.data.value?.latency ?? []).map(point => ({ at: point.at, value: point.latencyMs })),
))
const latencyOption = computed(() => {
  const timeZone = timezone.value
  return timeSeriesOption({
    points: latencyPoints.value,
    name: t('monitorDetails.responseLatency'),
    unit: 'ms',
    valueFormatter: value => n(value, { maximumFractionDigits: 2 }),
    timeFormatter: at => d(at, { key: 'short', timeZone }),
  })
})
const latencyLabel = computed(() => t('chart.summary', {
  name: t('monitorDetails.responseLatency'),
  count: latencyPoints.value.length,
  value: latencyPoints.value.length ? n(latencyPoints.value.at(-1)!.value, { maximumFractionDigits: 2 }) : '—',
  unit: 'ms',
}))
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
        <Button variant="ghost" as-child>
          <RouterLink to="/app/monitors">
            <span w="14px" h="14px" aria-hidden="true" class="i-lucide-arrow-left" />{{ t('monitorDetails.allMonitors') }}
          </RouterLink>
        </Button><template v-if="canEdit()">
          <Button :disabled="busy" @click="act('toggle')">
            <span v-if="monitor.enabled" w="14px" h="14px" aria-hidden="true" class="i-lucide-pause" /><span v-else w="14px" h="14px" aria-hidden="true" class="i-lucide-play" />{{
              monitor.enabled ? t('monitorDetails.pause') : t('monitorDetails.enable')
            }}
          </Button><Button as-child>
            <RouterLink :to="`/app/monitors/${monitor.id}/edit`">
              <span w="14px" h="14px" aria-hidden="true" class="i-lucide-pencil" />{{ t('common.edit') }}
            </RouterLink>
          </Button><Button v-if="monitor.type !== 'heartbeat'" :disabled="busy || !monitor.enabled" variant="primary" @click="act('check')">
            <span w="14px" h="14px" aria-hidden="true" class="i-lucide-refresh-cw" />{{ t('monitorDetails.checkNow') }}
          </Button>
        </template>
      </PageHeader>
      <div class="grid grid-cols-4 gap-16px mb-24px [&_.stat-card]:p-20px [&_.stat-label]:flex [&_.stat-label]:items-center [&_.stat-label]:justify-between [&_.stat-label]:gap-8px [&_.stat-label]:text-13px [&_.stat-label]:text-subtle [&_.stat-label]:font-400 [&_.stat-icon]:flex [&_.stat-icon]:text-subtle [&_.stat-value]:mt-20px [&_.stat-value]:mb-8px [&_.stat-value]:text-32px [&_.stat-value]:font-600 [&_.stat-value]:tracking-[-1px] [&_.stat-value]:leading-[1.2] [&_.stat-value]:tabular-nums [&_.stat-meta]:text-12px [&_.stat-meta]:text-subtle [&_.positive]:text-fg-success [@media(max-width:1200px)]:gap-12px [@media(max-width:1200px)]:[&_.stat-card]:p-18px [@media(max-width:1200px)]:[&_.stat-value]:text-27px [@media(max-width:900px)]:grid-cols-2 [@media(max-width:700px)]:gap-10px [@media(max-width:700px)]:[&_.stat-card]:p-17px [@media(max-width:700px)]:[&_.stat-value]:text-25px [@media(max-width:380px)]:grid-cols-1 [@container_workspace_(max-width:_700px)]:grid-cols-2! [@container_workspace_(max-width:_380px)]:[&&]:grid-cols-1!">
        <Card class="stat-card">
          <div class="stat-label">
            {{ t('monitorDetails.currentState') }}<span w="15px" h="15px" aria-hidden="true" class="i-lucide-activity" />
          </div>
          <div mt="5" mb="4">
            <StateBadge :state="monitor.type === 'certificate' ? monitor.certificate?.state : monitor.state" :paused="!monitor.enabled" />
          </div>
          <p class="stat-meta">
            {{ formatDate(monitor.lastCheckedAt) }}
          </p>
        </Card>
        <Card class="stat-card">
          <div class="stat-label">
            {{
              monitor.certificate
                ? t('common.certificateExpires')
                : t('monitorDetails.durationUptime')
            }}<span w="15px" h="15px" aria-hidden="true" class="i-lucide-circle-check" />
          </div>
          <div :style="monitor.certificate ? { fontSize: '16px' } : undefined" class="stat-value">
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
        </Card>
        <Card class="stat-card">
          <div class="stat-label">
            {{
              monitor.certificate
                ? t('monitorDetails.daysRemaining')
                : t('monitorDetails.observationCoverage')
            }}<span w="15px" h="15px" aria-hidden="true" class="i-lucide-shield-check" />
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
        </Card>
        <Card class="stat-card">
          <div class="stat-label">
            {{ t('monitorDetails.latestLatency') }}<span w="15px" h="15px" aria-hidden="true" class="i-lucide-clock" />
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
        </Card>
      </div>
      <Card v-if="monitor.type === 'heartbeat'" mb="6">
        <CardContent>
          <div flex="~ items-center justify-between gap-4">
            <div>
              <h2>{{ t('monitorDetails.heartbeatReporting') }}</h2>
              <p mt="2" class="muted" un-text="13px subtle">
                {{ t('monitorDetails.tokensCannotBeReadBackRotationInvalidatesThe') }}
              </p>
            </div>
            <Button v-if="canEdit()" :disabled="busy" @click="act('rotate')">
              <span w="14px" h="14px" aria-hidden="true" class="i-lucide-key-round" />{{ t('monitorDetails.generateRotateToken') }}
            </Button>
          </div>
          <Alert v-if="monitor.heartbeat?.lastReceivedAt" mt="4" variant="default">
            {{ t('monitorDetails.lastReport') }} {{ formatDate(monitor.heartbeat.lastReceivedAt) }} ·
            {{ monitor.heartbeat.lastSuccess ? t('monitorDetails.up') : t('monitorDetails.down') }}
            <p v-if="monitor.heartbeat.description" mt="2">
              {{ monitor.heartbeat.description }}
            </p>
          </Alert>
          <div v-if="heartbeatToken" class="heartbeat-url [overflow-wrap:anywhere]" p="14px" border="1 solid line" bg="tint" rounded="8px" un-text="12px" mt="15px">
            <code>{{ heartbeatUrl }}</code><Button variant="ghost" size="sm" @click="copy(heartbeatUrl)">
              <span w="13px" h="13px" aria-hidden="true" class="i-lucide-copy" />{{ copied ? t('monitorDetails.copied') : t('monitorDetails.copy') }}
            </Button>
            <p mt="3" class="muted" un-text="13px subtle">
              {{ t('monitorDetails.postReportsStatusUpOrStatusDownWith') }}
            </p>
          </div>
        </CardContent>
      </Card>
      <div v-if="monitor.certificate" flex="~ items-center" gap="9px" mb="22px" p="y-13px x-16px" border="1 solid line" rounded="8px" bg="base" un-text="12px subtle">
        <span w="16px" h="16px" aria-hidden="true" class="i-lucide-shield-check" /><span>{{ t('common.certificateExpires') }} {{ formatDate(monitor.certificate.expiresAt) }} ·
          {{
            t('monitorDetails.remainingDays', {
              days: monitor.certificate.expiresAt
                ? n(monitor.certificate.daysRemaining, 'decimal')
                : '—',
            })
          }}
          · {{ t('common.certificateRiskIsExcludedFromUptime') }}</span>
      </div>
      <div grid="~ cols-[minmax(0,1fr)_280px]" gap="22px" class="[&>*]:min-w-0 [@media(max-width:1200px)]:grid-cols-1 [@container_workspace_(max-width:_700px)]:grid-cols-1!">
        <Card as="section">
          <TabsRoot v-model="tab">
            <TabsList>
              <TabsTrigger value="history">
                {{
                  t('monitorDetails.historyTrends')
                }}
              </TabsTrigger><TabsTrigger value="configuration">
                {{
                  t('monitorDetails.configuration')
                }}
              </TabsTrigger>
              <div ml="auto" class="historical-period" flex="~ items-center" gap="9px" un-text="12px subtle">
                <Select v-model="period">
                  <SelectTrigger w="auto!" :aria-label="t('monitorDetails.statisticsWindow')">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectGroup>
                      <SelectItem value="24h">
                        {{ t('counts.hours', { count: 24 }, 24) }}
                      </SelectItem>
                      <SelectItem value="7d">
                        {{ t('counts.days', { count: 7 }, 7) }}
                      </SelectItem>
                      <SelectItem value="30d">
                        {{ t('counts.days', { count: 30 }, 30) }}
                      </SelectItem>
                    </SelectGroup>
                  </SelectContent>
                </Select>
              </div>
            </TabsList><TabsContent value="history">
              <CardContent>
                <div flex="~ justify-between items-center" mb="4">
                  <h3>{{ t('monitorDetails.responseLatency') }}</h3>
                  <span class="mini-label" un-text="12px subtle" tracking="0.5px">ms</span>
                </div>
                <EChart
                  v-if="latencyPoints.length"
                  :option="latencyOption"
                  :height="180"
                  :aria-label="latencyLabel"
                  :loading="history.isPending.value"
                />
                <div v-else h="180px" flex="~ items-center justify-center" un-text="subtle" role="status">
                  {{ t('chart.noObservations') }}
                </div>
                <FieldDescription as="p" mt="2">
                  {{ t('monitorDetails.sourceActualCheckRoundsMissingObservationsAreNot') }}
                </FieldDescription>
              </CardContent>
              <AsyncState :pending="history.isPending.value" :error="history.error.value" @retry="refreshHistory()">
                <EmptyState v-if="!history.data.value?.rounds.length" :title="t('monitorDetails.noCheckRecordsYet')" :description="t('monitorDetails.roundsAndDiagnosticsAppearAfterTheFirstCheck')" />
                <TableContainer v-else>
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>{{ t('monitorDetails.checkedAt') }}</TableHead>
                        <TableHead>{{ t('monitorDetails.result') }}</TableHead>
                        <TableHead>{{ t('monitorDetails.attempts') }}</TableHead>
                        <TableHead>{{ t('monitorDetails.duration') }}</TableHead>
                        <TableHead />
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      <TableRow v-for="round in history.data.value.rounds" :key="round.id">
                        <TableCell class="muted" un-text="13px subtle">
                          {{ formatDate(round.startedAt) }}
                        </TableCell>
                        <TableCell><StateBadge :state="round.success ? 'up' : 'down'" /></TableCell>
                        <TableCell>{{ round.attempts?.length || 0 }}</TableCell>
                        <TableCell>{{ duration(round.latencyMs) }}</TableCell>
                        <TableCell>
                          <Button variant="ghost" size="sm" @click="viewRound(round)">
                            {{ t('monitorDetails.diagnostics') }}
                          </Button>
                        </TableCell>
                      </TableRow>
                    </TableBody>
                  </Table>
                </TableContainer>
              </AsyncState>
            </TabsContent><TabsContent value="configuration">
              <CardContent>
                <pre class="json-output [overflow-wrap:anywhere]" un-text="12px" whitespace="pre-wrap" bg="tint" p="15px" border="1 solid line" rounded="8px" max-h="400px" overflow="auto">{{ JSON.stringify(monitor, null, 2) }}</pre>
              </CardContent>
            </TabsContent>
          </TabsRoot>
        </Card>
        <aside>
          <Card as="section">
            <CardContent>
              <h2 mb="6">
                {{ t('monitorDetails.monitorInformation') }}
              </h2>
              <dl class="definition-list tabular-nums [&_div]:flex [&_div]:justify-between [&_div]:gap-15px [&_div]:text-12px [&_dt]:text-subtle [&_dd]:m-0 [&_dd]:text-right [&_dd]:[overflow-wrap:anywhere]" grid="~" gap="16px">
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
              <Alert v-if="monitor.description" mt="6" as="p" variant="default">
                {{ monitor.description }}
              </Alert>
              <div flex="~ wrap gap-2" mt="4">
                <Badge v-for="tag in monitor.tags" :key="tag">
                  {{ tag }}
                </Badge>
              </div>
            </CardContent>
          </Card>
          <Alert mt="5" as="p" variant="default">
            {{ t('monitorDetails.uptimeUsesConfirmedStateDurationUnknownPausedAnd') }}
          </Alert>
          <Button v-if="canEdit()" mt="5" variant="destructive" @click="confirmDelete = true">
            <span w="14px" h="14px" aria-hidden="true" class="i-lucide-trash-2" />{{ t('common.deleteMonitor') }}
          </Button>
        </aside>
      </div>
    </template>
  </AsyncState><Modal v-model:open="confirmDelete" :title="t('common.deleteMonitor')" :description="t('monitorDetails.thisDeletesTheMonitorConfigurationConfirmItIs')">
    <template #footer>
      <Button @click="confirmDelete = false">
        {{ t('common.cancel') }}
      </Button><Button :disabled="busy" variant="destructive" @click="act('delete')">
        {{ t('monitorDetails.delete') }}
      </Button>
    </template>
  </Modal><Modal v-model:open="diagnosticsOpen" :title="t('monitorDetails.roundDiagnostics')" wide>
    <template v-if="selectedRound">
      <p mb="4" class="muted" un-text="13px subtle">
        {{ formatDate(selectedRound.startedAt) }} · {{ duration(selectedRound.latencyMs) }}
      </p>
      <Card v-for="attempt in selectedRound.attempts" :key="attempt.number" mb="4">
        <CardContent>
          <div flex="~ items-center justify-between">
            <h3>{{ t('monitorDetails.attemptNumber', { number: n(attempt.number) }) }}</h3>
            <StateBadge :state="attempt.success ? 'up' : 'down'" />
          </div>
          <FieldError v-if="attempt.error" as="p" py="10px" px="0">
            {{ attempt.error }}
          </FieldError>
          <pre mt="4" class="json-output [overflow-wrap:anywhere]" un-text="12px" whitespace="pre-wrap" bg="tint" p="15px" border="1 solid line" rounded="8px" max-h="400px" overflow="auto">{{ JSON.stringify(attempt.detail, null, 2) }}</pre>
        </CardContent>
      </Card>
    </template>
  </Modal>
</template>
