<script setup lang="ts">
import type { Round } from '../../../../../client/types.gen'
import { useMutation, useQuery } from '@pinia/colada'
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
import { Badge } from '../../../../../components/ui/badge'
import { Banner } from '../../../../../components/ui/banner'
import { PageHeader } from '../../../../../components/ui/blocks/page-header'
import { Button } from '../../../../../components/ui/button'
import { Chart } from '../../../../../components/ui/chart'
import { ClipboardText } from '../../../../../components/ui/clipboard-text'
import { Code } from '../../../../../components/ui/code'
import { Dialog } from '../../../../../components/ui/dialog'
import { Empty } from '../../../../../components/ui/empty'
import { FieldDescription, FieldError } from '../../../../../components/ui/field'
import { LayerCard, LayerCardPrimary, LayerCardSecondary } from '../../../../../components/ui/layer-card'
import { Loader } from '../../../../../components/ui/loader'
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
import { monitorStateDisplay } from '../../../../../lib/monitor-state'

const { t, n, d } = useI18n({ useScope: 'global' })

definePage({ meta: { title: 'navigation.monitorDetails', contentWidth: 'detail' } })

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
const deleteDialogOpen = computed({
  get: () => confirmDelete.value,
  set: (open: boolean) => {
    if (!busy.value)
      confirmDelete.value = open
  },
})
const selectedRound = ref<Round | null>(null)
const diagnosticsOpen = ref(false)
const heartbeatToken = ref('')
const heartbeatUrl = ref('')
const tab = ref('history')
const monitor = computed(() => query.data.value)
const currentState = computed(() => monitorStateDisplay(
  monitor.value?.type === 'certificate' ? monitor.value.certificate?.state : monitor.value?.state,
  { paused: monitor.value ? !monitor.value.enabled : false },
))
const rounds = computed(() => (history.data.value?.rounds ?? []).map(round => ({
  ...round,
  stateDisplay: monitorStateDisplay(round.success ? 'up' : 'down'),
})))
const attempts = computed(() => (selectedRound.value?.attempts ?? []).map(attempt => ({
  ...attempt,
  stateDisplay: monitorStateDisplay(attempt.success ? 'up' : 'down'),
})))
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
    <template v-if="monitor">
      <PageHeader :title="monitor.name" :description="targetOf(monitor)" class="mb-6">
        <template #actions>
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
        </template>
      </PageHeader>
      <div class="mb-6 grid grid-cols-4 gap-3 [@container_workspace_(max-width:_880px)]:grid-cols-2 [@container_workspace_(max-width:_440px)]:grid-cols-1 [&_.stat-label]:flex [&_.stat-label]:items-center [&_.stat-label]:justify-between [&_.stat-label]:gap-2 [&_.stat-label]:text-size-sm [&_.stat-label]:text-subtle [&_.stat-value]:my-4 [&_.stat-value]:text-size-3xl [&_.stat-value]:font-semibold [&_.stat-value]:tracking-tight [&_.stat-value]:tabular-nums [&_.stat-meta]:text-size-xs [&_.stat-meta]:text-subtle">
        <LayerCard>
          <LayerCardPrimary class="stat-card">
            <div class="stat-label">
              {{ t('monitorDetails.currentState') }}<span w="15px" h="15px" aria-hidden="true" class="i-lucide-activity" />
            </div>
            <div mt="5" mb="4">
              <Badge :variant="currentState.variant" :data-state="currentState.state" dot>
                {{ currentState.label }}
              </Badge>
            </div>
            <p class="stat-meta">
              {{ formatDate(monitor.lastCheckedAt) }}
            </p>
          </LayerCardPrimary>
        </LayerCard>
        <LayerCard>
          <LayerCardPrimary class="stat-card">
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
          </LayerCardPrimary>
        </LayerCard>
        <LayerCard>
          <LayerCardPrimary class="stat-card">
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
          </LayerCardPrimary>
        </LayerCard>
        <LayerCard>
          <LayerCardPrimary class="stat-card">
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
          </LayerCardPrimary>
        </LayerCard>
      </div>
      <LayerCard v-if="monitor.type === 'heartbeat'" mb="6">
        <LayerCardPrimary>
          <div flex="~ wrap items-center justify-between gap-4">
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
          <Banner v-if="monitor.heartbeat?.lastReceivedAt" mt="4" variant="secondary">
            {{ t('monitorDetails.lastReport') }} {{ formatDate(monitor.heartbeat.lastReceivedAt) }} ·
            {{ monitor.heartbeat.lastSuccess ? t('monitorDetails.up') : t('monitorDetails.down') }}
            <p v-if="monitor.heartbeat.description" mt="2">
              {{ monitor.heartbeat.description }}
            </p>
          </Banner>
          <div v-if="heartbeatToken" class="heartbeat-url [overflow-wrap:anywhere]" p="14px" border="1 solid line" bg="tint" rounded="8px" un-text="12px" mt="15px">
            <ClipboardText :text="heartbeatUrl" :copy-label="t('monitorDetails.copy')" :copied-label="t('monitorDetails.copied')" class="max-w-full" />
            <p mt="3" class="muted" un-text="13px subtle">
              {{ t('monitorDetails.postReportsStatusUpOrStatusDownWith') }}
            </p>
          </div>
        </LayerCardPrimary>
      </LayerCard>
      <div v-if="monitor.certificate" flex="~ items-center" gap="9px" mb="22px" p="y-13px x-16px" border="1 solid line" rounded="8px" bg="base" un-text="12px subtle">
        <span w="16px" h="16px" aria-hidden="true" class="i-lucide-shield-check shrink-0" /><span class="min-w-0 [overflow-wrap:anywhere]">{{ t('common.certificateExpires') }} {{ formatDate(monitor.certificate.expiresAt) }} ·
          {{
            t('monitorDetails.remainingDays', {
              days: monitor.certificate.expiresAt
                ? n(monitor.certificate.daysRemaining, 'decimal')
                : '—',
            })
          }}
          · {{ t('common.certificateRiskIsExcludedFromUptime') }}</span>
      </div>
      <div class="grid grid-cols-[minmax(0,1fr)_300px] items-start gap-6 [@container_workspace_(max-width:_960px)]:grid-cols-1">
        <LayerCard>
          <TabsRoot v-model="tab">
            <TabsList variant="line">
              <TabsTrigger variant="line" value="history">
                {{
                  t('monitorDetails.historyTrends')
                }}
              </TabsTrigger><TabsTrigger variant="line" value="configuration">
                {{
                  t('monitorDetails.configuration')
                }}
              </TabsTrigger>
            </TabsList><TabsContent value="history">
              <LayerCardPrimary>
                <div flex="~ wrap justify-between items-center gap-3" mb="4">
                  <h3>{{ t('monitorDetails.responseLatency') }}</h3>
                  <div class="historical-period" flex="~ items-center" gap="9px" un-text="12px subtle">
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
                </div>
                <Chart
                  v-if="latencyPoints.length"
                  :option="latencyOption"
                  :height="240"
                  :aria-label="latencyLabel"
                  :loading="history.isPending.value"
                />
                <div v-else h="240px" flex="~ items-center justify-center" un-text="subtle" role="status">
                  {{ t('chart.noObservations') }}
                </div>
                <FieldDescription as="p" mt="2">
                  {{ t('monitorDetails.sourceActualCheckRoundsMissingObservationsAreNot') }}
                </FieldDescription>
              </LayerCardPrimary>
              <div v-if="history.isPending.value" class="loading-state" flex="~ justify-center items-center gap-10px" p="60px" un-text="12px subtle" role="status">
                <Loader :label="t('asyncState.loadingData')" />{{ t('asyncState.loadingData') }}
              </div>
              <Banner v-else-if="history.error.value" variant="error">
                {{ errorText(history.error.value) }}
                <Button variant="ghost" @click="refreshHistory()">
                  {{ t('asyncState.retry') }}
                </Button>
              </Banner>
              <template v-else>
                <Empty v-if="!history.data.value?.rounds.length" size="sm" class="rounded-none border-none" :title="t('monitorDetails.noCheckRecordsYet')" :description="t('monitorDetails.roundsAndDiagnosticsAppearAfterTheFirstCheck')" />
                <TableContainer
                  v-else
                  :scroll-label="t('common.scrollTable')"
                  max-h="400px"
                  overscroll="contain"
                  tabindex="0"
                  role="region"
                  :aria-label="t('monitorDetails.historyTrends')"
                >
                  <Table>
                    <TableHeader sticky top="0" z="1">
                      <TableRow>
                        <TableHead>{{ t('monitorDetails.checkedAt') }}</TableHead>
                        <TableHead>{{ t('monitorDetails.result') }}</TableHead>
                        <TableHead class="text-end [@container_workspace_(max-width:_700px)]:hidden">
                          {{ t('monitorDetails.attempts') }}
                        </TableHead>
                        <TableHead class="text-end [@container_workspace_(max-width:_700px)]:hidden">
                          {{ t('monitorDetails.duration') }}
                        </TableHead>
                        <TableHead class="w-24 [@container_workspace_(max-width:_700px)]:w-12">
                          <span class="sr-only">{{ t('monitorDetails.diagnostics') }}</span>
                        </TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      <TableRow v-for="round in rounds" :key="round.id">
                        <TableCell class="muted" un-text="13px subtle">
                          {{ formatDate(round.startedAt) }}
                          <span class="mt-1 hidden text-size-xs [@container_workspace_(max-width:_700px)]:block">{{ t('monitorDetails.attempts') }} · {{ n(round.attempts?.length || 0) }}<br>{{ t('monitorDetails.duration') }} · {{ duration(round.latencyMs) }}</span>
                        </TableCell>
                        <TableCell>
                          <Badge :variant="round.stateDisplay.variant" :data-state="round.stateDisplay.state" dot>
                            {{ round.stateDisplay.label }}
                          </Badge>
                        </TableCell>
                        <TableCell class="text-end [@container_workspace_(max-width:_700px)]:hidden">
                          {{ n(round.attempts?.length || 0) }}
                        </TableCell>
                        <TableCell class="whitespace-nowrap text-end [@container_workspace_(max-width:_700px)]:hidden">
                          {{ duration(round.latencyMs) }}
                        </TableCell>
                        <TableCell class="text-end">
                          <Button variant="ghost" size="sm" :aria-label="t('monitorDetails.diagnostics')" @click="viewRound(round)">
                            <span class="[@container_workspace_(max-width:_700px)]:hidden">{{ t('monitorDetails.diagnostics') }}</span>
                            <span class="i-lucide-chevron-right hidden size-4 [@container_workspace_(max-width:_700px)]:block" aria-hidden="true" />
                          </Button>
                        </TableCell>
                      </TableRow>
                    </TableBody>
                  </Table>
                </TableContainer>
              </template>
            </TabsContent><TabsContent value="configuration">
              <LayerCardPrimary>
                <Code :code="JSON.stringify(monitor, null, 2)" lang="json" class="max-h-100" />
              </LayerCardPrimary>
            </TabsContent>
          </TabsRoot>
        </LayerCard>
        <aside class="min-w-0">
          <LayerCard>
            <LayerCardSecondary>
              <h2 class="text-size-lg font-semibold text-default">
                {{ t('monitorDetails.monitorInformation') }}
              </h2>
            </LayerCardSecondary>
            <LayerCardPrimary>
              <dl class="definition-list tabular-nums [&_div]:grid [&_div]:grid-cols-[minmax(0,1fr)_minmax(0,1fr)] [&_div]:gap-3 [&_div]:text-size-xs [&_dt]:text-subtle [&_dt]:[overflow-wrap:anywhere] [&_dd]:m-0 [&_dd]:text-end [&_dd]:[overflow-wrap:anywhere]" grid="~" gap="16px">
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
              <Banner v-if="monitor.description" mt="6" variant="secondary">
                {{ monitor.description }}
              </Banner>
              <div flex="~ wrap gap-2" mt="4">
                <Badge v-for="tag in monitor.tags" :key="tag" class="max-w-full whitespace-normal! [overflow-wrap:anywhere]">
                  {{ tag }}
                </Badge>
              </div>
            </LayerCardPrimary>
          </LayerCard>
          <Banner mt="5" variant="secondary">
            {{ t('monitorDetails.uptimeUsesConfirmedStateDurationUnknownPausedAnd') }}
          </Banner>
          <Button v-if="canEdit()" mt="5" variant="destructive" @click="confirmDelete = true">
            <span w="14px" h="14px" aria-hidden="true" class="i-lucide-trash-2" />{{ t('common.deleteMonitor') }}
          </Button>
        </aside>
      </div>
    </template>
  </template><Dialog v-model:open="deleteDialogOpen" :close-label="t('common.close')" :title="t('common.deleteMonitor')" :description="t('monitorDetails.thisDeletesTheMonitorConfigurationConfirmItIs')">
    <template #footer>
      <Button :disabled="busy" @click="confirmDelete = false">
        {{ t('common.cancel') }}
      </Button><Button :loading="busy" variant="destructive" @click="act('delete')">
        {{ t('monitorDetails.delete') }}
      </Button>
    </template>
  </Dialog><Dialog v-model:open="diagnosticsOpen" :close-label="t('common.close')" :title="t('monitorDetails.roundDiagnostics')" wide>
    <template v-if="selectedRound">
      <p mb="4" class="muted" un-text="13px subtle">
        {{ formatDate(selectedRound.startedAt) }} · {{ duration(selectedRound.latencyMs) }}
      </p>
      <LayerCard v-for="attempt in attempts" :key="attempt.number" mb="4">
        <LayerCardPrimary>
          <div flex="~ wrap items-center justify-between gap-3">
            <h3>{{ t('monitorDetails.attemptNumber', { number: n(attempt.number) }) }}</h3>
            <Badge :variant="attempt.stateDisplay.variant" :data-state="attempt.stateDisplay.state" dot>
              {{ attempt.stateDisplay.label }}
            </Badge>
          </div>
          <FieldError v-if="attempt.error" as="p" py="10px" px="0">
            {{ attempt.error }}
          </FieldError>
          <Code :code="JSON.stringify(attempt.detail, null, 2) ?? '{}'" lang="json" class="mt-4 max-h-100" />
        </LayerCardPrimary>
      </LayerCard>
    </template>
  </Dialog>
</template>
