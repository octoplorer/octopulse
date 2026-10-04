<script setup lang="ts">
import { useQuery } from '@pinia/colada'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  listIncidentsQuery,
  listMaintenanceQuery,
  listMonitorsQuery,
  listPagesQuery,
} from '../../../client/@pinia/colada.gen'
import AsyncState from '../../../components/AsyncState.vue'
import EmptyState from '../../../components/EmptyState.vue'
import PageHeader from '../../../components/PageHeader.vue'
import StateBadge from '../../../components/StateBadge.vue'
import { Alert } from '../../../components/ui/alert'
import { Badge } from '../../../components/ui/badge'
import { Button } from '../../../components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../../../components/ui/card'
import { Table, TableBody, TableCell, TableContainer, TableFooter, TableHead, TableHeader, TableRow } from '../../../components/ui/table'
import { canEdit } from '../../../composables/api'
import { usePollingEnabled } from '../../../composables/polling'
import { formatDate, statusLabel } from '../../../composables/preferences'
import { targetOf } from '../../../lib/monitor'
import { publishedEntry } from '../../../lib/pages'

const { t } = useI18n({ useScope: 'global' })

definePage({ meta: { title: 'navigation.overview' } })

const pollingEnabled = usePollingEnabled()
const monitors = useQuery(
  () =>
    ({
      ...listMonitorsQuery(),
      staleTime: 10000,
      enabled: pollingEnabled.value,
      autoRefetch: 30000,
    }),
)
const incidents = useQuery({
  ...listIncidentsQuery(),
  staleTime: 10000,
})
const pages = useQuery({
  ...listPagesQuery(),
  staleTime: 10000,
})
const maintenance = useQuery({
  ...listMaintenanceQuery(),
  staleTime: 10000,
})
const items = computed(() => monitors.data.value?.items || [])
const active = computed(() => items.value.filter(m => m.enabled))
const up = computed(() => active.value.filter(m => m.type !== 'certificate' && m.state === 'up'))
const down = computed(() =>
  active.value.filter(m => m.type !== 'certificate' && m.state === 'down'),
)
const unknown = computed(() =>
  active.value.filter(m => m.type !== 'certificate' && m.state === 'unknown'),
)
const ordered = computed(() =>
  [...items.value]
    .sort(
      (a, b) =>
        Number(b.state === 'down') - Number(a.state === 'down') || b.updatedAt - a.updatedAt,
    )
    .slice(0, 8),
)
const recentIncidents = computed(() => incidents.data.value?.items.slice(0, 5) || [])
const nextMaintenance = computed(
  () =>
    maintenance.data.value?.items
      .filter(x => x.endsAt > Date.now())
      .sort((a, b) => a.startsAt - b.startsAt)
      .slice(0, 3) || [],
)
</script>

<template>
  <PageHeader :title="t('overview.serviceOverview')" :description="t('overview.aClearViewOfEveryServiceHeartbeat')" eyebrow="YOUR INFRASTRUCTURE, AT A GLANCE">
    <Button @click="monitors.refetch()">
      <span w="14px" h="14px" aria-hidden="true" class="i-lucide-refresh-cw" />{{ t('common.refresh') }}
    </Button><Button v-if="canEdit()" variant="primary" as-child>
      <RouterLink to="/app/monitors/new">
        <span w="15px" h="15px" aria-hidden="true" class="i-lucide-plus" />{{ t('common.addMonitor') }}
      </RouterLink>
    </Button>
  </PageHeader>
  <div class="stats-grid">
    <Card class="stat-card">
      <div class="stat-label">
        {{ t('overview.totalMonitors') }}<span class="stat-icon"><span w="15px" h="15px" aria-hidden="true" class="i-lucide-activity" /></span>
      </div>
      <div class="stat-value">
        {{ items.length }}
      </div>
      <div class="stat-meta">
        {{ t('overview.active') }} {{ active.length }} · {{ t('common.paused') }}
        {{ items.length - active.length }}
      </div>
    </Card>
    <Card class="stat-card">
      <div class="stat-label">
        {{ t('overview.operational') }}<span class="stat-icon"><span w="15px" h="15px" aria-hidden="true" class="i-lucide-circle-check" /></span>
      </div>
      <div un-text="[var(--success)]" class="stat-value">
        {{ up.length }}
      </div>
      <div class="stat-meta">
        <span class="positive">{{ t('overview.confirmedServiceStates') }}</span>
      </div>
    </Card>
    <Card class="stat-card">
      <div class="stat-label">
        {{ t('overview.needsAttention')
        }}<span class="stat-icon"><span w="15px" h="15px" aria-hidden="true" class="i-lucide-triangle-alert" /></span>
      </div>
      <div :style="{ color: down.length ? 'var(--danger)' : undefined }" class="stat-value">
        {{ down.length }}
      </div>
      <div class="stat-meta">
        {{ t('overview.confirmedOutages') }}
      </div>
    </Card>
    <Card class="stat-card">
      <div class="stat-label">
        {{ t('overview.waitingForData') }}<span class="stat-icon"><span w="15px" h="15px" aria-hidden="true" class="i-lucide-clock" /></span>
      </div>
      <div class="stat-value">
        {{ unknown.length }}
      </div>
      <div class="stat-meta">
        {{ t('overview.initialChecksOrCollectionGaps') }}
      </div>
    </Card>
  </div>
  <div class="dashboard-grid">
    <div>
      <Card as="section">
        <CardHeader>
          <div>
            <CardTitle>{{ t('common.monitors') }}</CardTitle>
            <CardDescription>
              {{ t('overview.outagesFirstSoYouCanFocusOnWhat') }}
            </CardDescription>
          </div>
          <Button variant="ghost" size="sm" as-child>
            <RouterLink to="/app/monitors">
              {{ t('overview.viewAll') }}<span w="13px" h="13px" aria-hidden="true" class="i-lucide-arrow-up-right" />
            </RouterLink>
          </Button>
        </CardHeader>
        <AsyncState :pending="monitors.isPending.value" :error="monitors.error.value" @retry="monitors.refetch()">
          <EmptyState v-if="!items.length" :title="t('overview.monitorYourFirstService')" :description="t('overview.httpTcpDnsHeartbeatAndCertificateChecksAre')">
            <Button v-if="canEdit()" variant="primary" as-child>
              <RouterLink to="/app/monitors/new">
                <span w="14px" h="14px" aria-hidden="true" class="i-lucide-plus" />{{ t('common.createMonitor') }}
              </RouterLink>
            </Button>
          </EmptyState>
          <TableContainer v-else>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{{ t('common.service') }}</TableHead>
                  <TableHead>{{ t('common.status') }}</TableHead>
                  <TableHead>{{ t('overview.interval') }}</TableHead>
                  <TableHead>{{ t('common.lastCheck') }}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="monitor in ordered" :key="monitor.id">
                  <TableCell>
                    <RouterLink :to="`/app/monitors/${monitor.id}`" flex="~ items-center gap-3">
                      <span class="monitor-type-icon" flex="~ items-center justify-center shrink-0" size="32px" border="1 solid $border" rounded="8px" un-text="$muted" bg="$surface"><span v-if="monitor.type === 'http'" w="16px" h="16px" aria-hidden="true" class="i-lucide-globe" /><span v-else w="16px" h="16px" aria-hidden="true" class="i-lucide-server" /></span><span><span class="monitor-name block" font="600" un-text="13px">{{ monitor.name }}</span><span class="monitor-sub block [overflow-wrap:anywhere]" un-text="12px $muted" mt="3px" max-w="300px">{{ targetOf(monitor) }}</span></span>
                    </RouterLink>
                  </TableCell>
                  <TableCell>
                    <StateBadge
                      :state="
                        monitor.type === 'certificate' ? monitor.certificate?.state : monitor.state
                      " :paused="!monitor.enabled"
                    />
                  </TableCell>
                  <TableCell>
                    {{
                      monitor.type === 'heartbeat'
                        ? monitor.heartbeat?.periodSeconds
                        : monitor.intervalSeconds
                    }}
                    s
                  </TableCell>
                  <TableCell class="muted" un-text="13px $muted">
                    {{ formatDate(monitor.lastCheckedAt) }}
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </TableContainer>
        </AsyncState>
        <TableFooter v-if="items.length">
          <span>{{
            t('overview.showingMonitors', { shown: ordered.length, total: items.length })
          }}</span><span>{{ t('common.refreshesEvery30Seconds') }}</span>
        </TableFooter>
      </Card>
      <Card mt="6" as="section">
        <CardHeader>
          <CardTitle>{{ t('overview.publicStatusPages') }}</CardTitle>
          <Button variant="ghost" size="sm" as-child>
            <RouterLink to="/app/pages">
              {{ t('overview.managePages') }}<span w="13px" h="13px" aria-hidden="true" class="i-lucide-arrow-up-right" />
            </RouterLink>
          </Button>
        </CardHeader>
        <CardContent>
          <EmptyState v-if="!pages.data.value?.items.length" :title="t('overview.keepEveryoneInformed')" :description="t('overview.publishAStatusPageWithYourOwnBrand')" />
          <div v-for="page in pages.data.value?.items.slice(0, 3)" :key="page.id" flex="~ items-center justify-between gap-4" py="3">
            <div flex="~ items-center gap-3">
              <span un-text="[var(--accent)]" w="18px" h="18px" aria-hidden="true" class="i-lucide-globe" />
              <div>
                <h3 un-text="xs">
                  {{ page.name }}
                </h3>
                <p class="muted" un-text="13px $muted">
                  /{{ page.publishedAt ? publishedEntry(page).slug : page.slug
                  }}{{
                    (page.publishedAt ? publishedEntry(page).domain : page.domain)
                      ? ` · ${page.publishedAt ? publishedEntry(page).domain : page.domain}`
                      : ''
                  }}
                </p>
              </div>
            </div>
            <Badge>
              {{
                page.publishedAt ? t('common.published') : t('common.draft')
              }}
            </Badge>
          </div>
        </CardContent>
      </Card>
    </div>
    <aside>
      <Card as="section">
        <CardHeader>
          <CardTitle>{{ t('overview.incidentActivity') }}</CardTitle>
          <Badge>{{ recentIncidents.length }}</Badge>
        </CardHeader>
        <CardContent>
          <p v-if="!recentIncidents.length" py="4" class="muted" un-text="13px $muted">
            {{ t('overview.noIncidentAnnouncementsUpdatesWillAppearHere') }}
          </p>
          <RouterLink v-for="incident in recentIncidents" :key="incident.id" to="/app/incidents" class="activity-item">
            <i :style="{ background: incident.status === 'resolved' ? 'var(--success)' : 'var(--warning)' }" class="activity-dot" />
            <div>
              <h3>{{ incident.title }}</h3>
              <p>{{ statusLabel(incident.status) }} · {{ formatDate(incident.updatedAt) }}</p>
            </div>
          </RouterLink>
        </CardContent>
      </Card>
      <Card mt="6" as="section">
        <CardHeader>
          <CardTitle>{{ t('overview.upcomingMaintenance') }}</CardTitle>
        </CardHeader>
        <CardContent>
          <p v-if="!nextMaintenance.length" py="4" class="muted" un-text="13px $muted">
            {{ t('overview.noScheduledMaintenance') }}
          </p>
          <RouterLink v-for="window in nextMaintenance" :key="window.id" to="/app/maintenance" class="activity-item">
            <span un-text="[var(--muted)]" mt="1" w="15px" h="15px" aria-hidden="true" class="i-lucide-clock" />
            <div>
              <h3>{{ window.name }}</h3>
              <p>{{ formatDate(window.startsAt) }}</p>
            </div>
          </RouterLink>
        </CardContent>
      </Card>
      <Alert mt="6" variant="default">
        <p mb="8px 2" class="eyebrow" un-text="12px $muted" font="500">
          OCTOPULSE
        </p>
        {{ t('overview.uptimeReflectsConfirmedDurationPausedMaintenanceAndMissing') }}
      </Alert>
    </aside>
  </div>
</template>
