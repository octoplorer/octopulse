<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { computed } from 'vue'
import {
  Plus,
  ArrowUpRight,
  Activity,
  CheckCircle2,
  AlertTriangle,
  Clock,
  Globe,
  RefreshCw,
  Server,
} from '@lucide/vue'
import { useCollection } from '../../../lib/data'
import {
  listMonitorsQuery,
  listIncidentsQuery,
  listPagesQuery,
  listMaintenanceQuery,
} from '../../../client/@pinia/colada.gen'
import type { Monitor, Incident, Page, Maintenance } from '../../../lib/types'
import { targetOf } from '../../../lib/monitor'
import { publishedEntry } from '../../../lib/pages'
import { formatDate, statusLabel } from '../../../lib/preferences'
import { canEdit } from '../../../lib/api'
import { useIntervalFn } from '@vueuse/core'
import PageHeader from '../../../components/PageHeader.vue'
import StateBadge from '../../../components/StateBadge.vue'
import AsyncState from '../../../components/AsyncState.vue'
import EmptyState from '../../../components/EmptyState.vue'

const { t } = useI18n({ useScope: 'global' })

definePage({ meta: { title: 'navigation.overview' } })

const monitors = useCollection<Monitor>('monitors', listMonitorsQuery()),
  incidents = useCollection<Incident>('incidents', listIncidentsQuery()),
  pages = useCollection<Page>('pages', listPagesQuery()),
  maintenance = useCollection<Maintenance>('maintenance', listMaintenanceQuery())
const items = computed(() => monitors.data.value?.items || []),
  active = computed(() => items.value.filter((m) => m.enabled)),
  up = computed(() => active.value.filter((m) => m.type !== 'certificate' && m.state === 'up')),
  down = computed(() => active.value.filter((m) => m.type !== 'certificate' && m.state === 'down')),
  unknown = computed(() =>
    active.value.filter((m) => m.type !== 'certificate' && m.state === 'unknown'),
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
      .filter((x) => x.endsAt > Date.now())
      .sort((a, b) => a.startsAt - b.startsAt)
      .slice(0, 3) || [],
)
useIntervalFn(() => monitors.refresh(), 30000)
</script>
<template>
  <PageHeader
    :title="t('overview.serviceOverview')"
    :description="t('overview.aClearViewOfEveryServiceHeartbeat')"
    eyebrow="YOUR INFRASTRUCTURE, AT A GLANCE"
    ><button class="button" @click="monitors.refresh()">
      <RefreshCw :size="14" />{{ t('common.refresh') }}</button
    ><RouterLink v-if="canEdit()" to="/app/monitors/new" class="button primary"
      ><Plus :size="15" />{{ t('common.addMonitor') }}</RouterLink
    ></PageHeader
  >
  <div class="stats-grid">
    <div class="card stat-card">
      <div class="stat-label">
        {{ t('overview.totalMonitors') }}<span class="stat-icon"><Activity :size="15" /></span>
      </div>
      <div class="stat-value">{{ items.length }}</div>
      <div class="stat-meta">
        {{ t('overview.active') }} {{ active.length }} · {{ t('common.paused') }}
        {{ items.length - active.length }}
      </div>
    </div>
    <div class="card stat-card">
      <div class="stat-label">
        {{ t('overview.operational') }}<span class="stat-icon"><CheckCircle2 :size="15" /></span>
      </div>
      <div class="stat-value" un-text="[var(--accent)]">{{ up.length }}</div>
      <div class="stat-meta">
        <span class="positive">{{ t('overview.confirmedServiceStates') }}</span>
      </div>
    </div>
    <div class="card stat-card">
      <div class="stat-label">
        {{ t('overview.needsAttention')
        }}<span class="stat-icon"><AlertTriangle :size="15" /></span>
      </div>
      <div class="stat-value" :style="{ color: down.length ? 'var(--danger)' : undefined }">
        {{ down.length }}
      </div>
      <div class="stat-meta">{{ t('overview.confirmedOutages') }}</div>
    </div>
    <div class="card stat-card">
      <div class="stat-label">
        {{ t('overview.waitingForData') }}<span class="stat-icon"><Clock :size="15" /></span>
      </div>
      <div class="stat-value">{{ unknown.length }}</div>
      <div class="stat-meta">
        {{ t('overview.initialChecksOrCollectionGaps') }}
      </div>
    </div>
  </div>
  <div class="dashboard-grid">
    <div>
      <section class="card">
        <div class="card-header">
          <div>
            <h2>{{ t('common.monitors') }}</h2>
            <p class="muted">
              {{ t('overview.outagesFirstSoYouCanFocusOnWhat') }}
            </p>
          </div>
          <RouterLink to="/app/monitors" class="button small ghost"
            >{{ t('overview.viewAll') }}<ArrowUpRight :size="13"
          /></RouterLink>
        </div>
        <AsyncState
          :pending="monitors.isPending.value"
          :error="monitors.error.value"
          @retry="monitors.refresh()"
          ><EmptyState
            v-if="!items.length"
            :title="t('overview.monitorYourFirstService')"
            :description="t('overview.httpTcpDnsHeartbeatAndCertificateChecksAre')"
            ><RouterLink v-if="canEdit()" to="/app/monitors/new" class="button primary"
              ><Plus :size="14" />{{ t('common.createMonitor') }}</RouterLink
            ></EmptyState
          >
          <div v-else class="table-wrap">
            <table class="data-table">
              <thead>
                <tr>
                  <th>{{ t('common.service') }}</th>
                  <th>{{ t('common.status') }}</th>
                  <th>{{ t('overview.interval') }}</th>
                  <th>{{ t('common.lastCheck') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="monitor in ordered" :key="monitor.id">
                  <td>
                    <RouterLink :to="`/app/monitors/${monitor.id}`" un-flex="~ items-center gap-3"
                      ><span class="monitor-type-icon"
                        ><Globe v-if="monitor.type === 'http'" :size="16" /><Server
                          v-else
                          :size="16" /></span
                      ><span
                        ><span class="monitor-name">{{ monitor.name }}</span
                        ><span class="monitor-sub">{{ targetOf(monitor) }}</span></span
                      ></RouterLink
                    >
                  </td>
                  <td>
                    <StateBadge
                      :state="
                        monitor.type === 'certificate' ? monitor.certificate?.state : monitor.state
                      "
                      :paused="!monitor.enabled"
                    />
                  </td>
                  <td>
                    {{
                      monitor.type === 'heartbeat'
                        ? monitor.heartbeat?.periodSeconds
                        : monitor.intervalSeconds
                    }}
                    s
                  </td>
                  <td class="muted" un-text="10px">{{ formatDate(monitor.lastCheckedAt) }}</td>
                </tr>
              </tbody>
            </table>
          </div></AsyncState
        >
        <div v-if="items.length" class="table-footer">
          <span>{{
            t('overview.showingMonitors', { shown: ordered.length, total: items.length })
          }}</span
          ><span>{{ t('common.refreshesEvery30Seconds') }}</span>
        </div>
      </section>
      <section class="card" un-mt="6">
        <div class="card-header">
          <h2>{{ t('overview.publicStatusPages') }}</h2>
          <RouterLink to="/app/pages" class="button small ghost"
            >{{ t('overview.managePages') }}<ArrowUpRight :size="13"
          /></RouterLink>
        </div>
        <div class="card-body">
          <EmptyState
            v-if="!pages.data.value?.items.length"
            :title="t('overview.keepEveryoneInformed')"
            :description="t('overview.publishAStatusPageWithYourOwnBrand')"
          />
          <div
            v-for="page in pages.data.value?.items.slice(0, 3)"
            :key="page.id"
            un-flex="~ items-center justify-between gap-4"
            un-py="3"
          >
            <div un-flex="~ items-center gap-3">
              <Globe :size="18" un-text="[var(--accent)]" />
              <div>
                <h3 un-text="xs">{{ page.name }}</h3>
                <p class="muted" un-text="10px">
                  /{{ page.publishedAt ? publishedEntry(page).slug : page.slug
                  }}{{
                    (page.publishedAt ? publishedEntry(page).domain : page.domain)
                      ? ` · ${page.publishedAt ? publishedEntry(page).domain : page.domain}`
                      : ''
                  }}
                </p>
              </div>
            </div>
            <span class="pill">{{
              page.publishedAt ? t('common.published') : t('common.draft')
            }}</span>
          </div>
        </div>
      </section>
    </div>
    <aside>
      <section class="card">
        <div class="card-header">
          <h2>{{ t('overview.incidentActivity') }}</h2>
          <span class="pill">{{ recentIncidents.length }}</span>
        </div>
        <div class="card-body">
          <p v-if="!recentIncidents.length" class="muted" un-py="4">
            {{ t('overview.noIncidentAnnouncementsUpdatesWillAppearHere') }}
          </p>
          <RouterLink
            v-for="incident in recentIncidents"
            :key="incident.id"
            to="/app/incidents"
            class="activity-item"
            ><i
              class="activity-dot"
              :style="{ background: incident.status === 'resolved' ? 'var(--accent)' : '#dfa151' }"
            />
            <div>
              <h3>{{ incident.title }}</h3>
              <p>{{ statusLabel(incident.status) }} · {{ formatDate(incident.updatedAt) }}</p>
            </div></RouterLink
          >
        </div>
      </section>
      <section class="card" un-mt="6">
        <div class="card-header">
          <h2>{{ t('overview.upcomingMaintenance') }}</h2>
        </div>
        <div class="card-body">
          <p v-if="!nextMaintenance.length" class="muted" un-py="4">
            {{ t('overview.noScheduledMaintenance') }}
          </p>
          <RouterLink
            v-for="window in nextMaintenance"
            :key="window.id"
            to="/app/maintenance"
            class="activity-item"
            ><Clock :size="15" un-text="[var(--muted)]" un-mt="1" />
            <div>
              <h3>{{ window.name }}</h3>
              <p>{{ formatDate(window.startsAt) }}</p>
            </div></RouterLink
          >
        </div>
      </section>
      <div class="note" un-mt="6">
        <p class="eyebrow" un-mb="2">OCTOPULSE</p>
        {{ t('overview.uptimeReflectsConfirmedDurationPausedMaintenanceAndMissing') }}
      </div>
    </aside>
  </div>
</template>
