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
import { Badge } from '../../../components/ui/badge'
import { Banner } from '../../../components/ui/banner'
import { PageHeader } from '../../../components/ui/blocks/page-header'
import { Button } from '../../../components/ui/button'
import { Empty } from '../../../components/ui/empty'
import { LayerCard, LayerCardPrimary, LayerCardSecondary } from '../../../components/ui/layer-card'
import { Loader, SkeletonLine } from '../../../components/ui/loader'
import { Table, TableBody, TableCell, TableContainer, TableHead, TableHeader, TablePagination, TableRow } from '../../../components/ui/table'
import { Text } from '../../../components/ui/text'
import { canEdit } from '../../../composables/api'
import { usePollingEnabled } from '../../../composables/polling'
import { formatDate, statusLabel } from '../../../composables/preferences'
import { errorText } from '../../../lib/errors'
import { targetOf } from '../../../lib/monitor'
import { monitorStateDisplay } from '../../../lib/monitor-state'
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
    .slice(0, 8)
    .map(monitor => ({
      ...monitor,
      stateDisplay: monitorStateDisplay(
        monitor.type === 'certificate' ? monitor.certificate?.state : monitor.state,
        { paused: !monitor.enabled },
      ),
    })),
)
const recentIncidents = computed(() => incidents.data.value?.items.slice(0, 5) || [])
const nextMaintenance = computed(
  () =>
    maintenance.data.value?.items
      .filter(x => x.endsAt > Date.now())
      .sort((a, b) => a.startsAt - b.startsAt)
      .slice(0, 3) || [],
)
const summary = computed(() => [
  { label: t('overview.total-monitors'), value: items.value.length, icon: 'i-lucide-activity', description: `${t('overview.active')} ${active.value.length} · ${t('common.paused')} ${items.value.length - active.value.length}`, color: 'text-default' },
  { label: t('overview.operational'), value: up.value.length, icon: 'i-lucide-circle-check', description: t('overview.confirmed-service-states'), color: 'text-fg-success' },
  { label: t('overview.needs-attention'), value: down.value.length, icon: 'i-lucide-triangle-alert', description: t('overview.confirmed-outages'), color: down.value.length ? 'text-fg-danger' : 'text-default' },
  { label: t('overview.waiting-for-data'), value: unknown.value.length, icon: 'i-lucide-clock', description: t('overview.initial-checks-or-collection-gaps'), color: 'text-subtle' },
])
</script>

<template>
  <PageHeader :title="t('overview.service-overview')" :description="t('overview.a-clear-view-of-every-service-heartbeat')" class="mb-6">
    <template #actions>
      <Button :loading="monitors.isPending.value" @click="monitors.refetch()">
        <span class="i-lucide-refresh-cw size-4" aria-hidden="true" />{{ t('common.refresh') }}
      </Button>
      <Button v-if="canEdit()" variant="primary" as-child>
        <RouterLink to="/app/monitors/new">
          <span class="i-lucide-plus size-4" aria-hidden="true" />{{ t('common.add-monitor') }}
        </RouterLink>
      </Button>
    </template>
  </PageHeader>
  <div class="mb-6 grid grid-cols-4 gap-3 [@container_workspace_(max-width:_880px)]:grid-cols-2 [@container_workspace_(max-width:_440px)]:grid-cols-1">
    <LayerCard v-for="stat in summary" :key="stat.label">
      <LayerCardPrimary class="space-y-4">
        <div class="flex items-center justify-between gap-2 text-size-sm text-subtle">
          <span>{{ stat.label }}</span><span :class="stat.icon" class="size-4" aria-hidden="true" />
        </div>
        <SkeletonLine v-if="monitors.isPending.value" class="h-9 w-20" />
        <div v-else class="text-size-3xl font-semibold tracking-tight tabular-nums" :class="stat.color">
          {{ monitors.error.value ? '—' : stat.value }}
        </div>
        <Text as="p" size="xs" variant="secondary">
          {{ stat.description }}
        </Text>
      </LayerCardPrimary>
    </LayerCard>
  </div>
  <div class="grid grid-cols-[minmax(0,1fr)_320px] items-start gap-6 [@container_workspace_(max-width:_1000px)]:grid-cols-1">
    <div class="min-w-0 space-y-6">
      <LayerCard>
        <LayerCardSecondary class="flex flex-wrap items-center justify-between gap-3">
          <div class="space-y-1">
            <Text as="h2" variant="heading">
              {{ t('common.monitors') }}
            </Text>
            <Text as="p" size="sm" variant="secondary">
              {{ t('overview.outages-first-so-you-can-focus-on-what') }}
            </Text>
          </div>
          <Button variant="ghost" size="sm" as-child>
            <RouterLink to="/app/monitors">
              {{ t('overview.view-all') }}<span class="i-lucide-arrow-up-right size-4" aria-hidden="true" />
            </RouterLink>
          </Button>
        </LayerCardSecondary>
        <div v-if="monitors.isPending.value" class="loading-state" flex="~ justify-center items-center gap-10px" p="60px" un-text="12px subtle" role="status">
          <Loader :label="t('async-state.loading-data')" />{{ t('async-state.loading-data') }}
        </div>
        <Banner v-else-if="monitors.error.value" variant="error">
          {{ errorText(monitors.error.value) }}
          <Button variant="ghost" @click="monitors.refetch()">
            {{ t('async-state.retry') }}
          </Button>
        </Banner>
        <template v-else>
          <Empty v-if="!items.length" :title="t('overview.monitor-your-first-service')" :description="t('overview.http-tcp-dns-heartbeat-and-certificate-checks-are')" size="sm" class="rounded-none border-none">
            <template #icon>
              <span class="i-lucide-activity size-8 text-subtle" aria-hidden="true" />
            </template>
            <template #actions>
              <Button v-if="canEdit()" variant="primary" as-child>
                <RouterLink to="/app/monitors/new">
                  <span class="i-lucide-plus size-4" aria-hidden="true" />{{ t('common.create-monitor') }}
                </RouterLink>
              </Button>
            </template>
          </Empty>
          <TableContainer v-else :scroll-label="t('common.scroll-table')">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{{ t('common.service') }}</TableHead><TableHead class="[@container_workspace_(max-width:_700px)]:hidden">
                    {{ t('common.status') }}
                  </TableHead><TableHead class="text-end [@container_workspace_(max-width:_700px)]:hidden">
                    {{ t('overview.interval') }}
                  </TableHead><TableHead class="[@container_workspace_(max-width:_700px)]:hidden">
                    {{ t('common.last-check') }}
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="monitor in ordered" :key="monitor.id">
                  <TableCell>
                    <RouterLink :to="`/app/monitors/${monitor.id}`" class="flex items-center gap-3 rounded-md outline-none focus-visible:ring-2 focus-visible:ring-brand">
                      <span class="flex size-8 shrink-0 items-center justify-center rounded-lg bg-recessed text-subtle"><span :class="monitor.type === 'http' ? 'i-lucide-globe' : 'i-lucide-server'" class="size-4" aria-hidden="true" /></span>
                      <span class="min-w-0"><span class="block font-medium [overflow-wrap:anywhere]">{{ monitor.name }}</span><span class="mt-1 block max-w-72 text-size-xs text-subtle [overflow-wrap:anywhere]">{{ targetOf(monitor) }}</span>
                        <span class="mt-2 hidden space-y-2 [@container_workspace_(max-width:_700px)]:block">
                          <Badge :variant="monitor.stateDisplay.variant" :data-state="monitor.stateDisplay.state" dot>{{ monitor.stateDisplay.label }}</Badge>
                          <span class="block text-size-xs text-subtle">{{ t('overview.interval') }} · {{ monitor.type === 'heartbeat' ? monitor.heartbeat?.periodSeconds : monitor.intervalSeconds }} s</span>
                          <span class="block text-size-xs text-subtle">{{ t('common.last-check') }} · {{ formatDate(monitor.lastCheckedAt) }}</span>
                        </span>
                      </span>
                    </RouterLink>
                  </TableCell>
                  <TableCell class="[@container_workspace_(max-width:_700px)]:hidden">
                    <Badge :variant="monitor.stateDisplay.variant" :data-state="monitor.stateDisplay.state" dot>
                      {{ monitor.stateDisplay.label }}
                    </Badge>
                  </TableCell>
                  <TableCell class="whitespace-nowrap text-end [@container_workspace_(max-width:_700px)]:hidden">
                    {{ monitor.type === 'heartbeat' ? monitor.heartbeat?.periodSeconds : monitor.intervalSeconds }} s
                  </TableCell>
                  <TableCell class="whitespace-nowrap text-size-sm text-subtle [@container_workspace_(max-width:_700px)]:hidden">
                    {{ formatDate(monitor.lastCheckedAt) }}
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </TableContainer>
        </template>
        <TablePagination v-if="items.length">
          <span>{{ t('overview.showing-monitors', { shown: ordered.length, total: items.length }) }}</span><span>{{ t('common.refreshes-every-30-seconds') }}</span>
        </TablePagination>
      </LayerCard>
    </div>
    <aside class="min-w-0 space-y-6">
      <LayerCard>
        <LayerCardSecondary class="flex flex-wrap items-center justify-between gap-3">
          <Text as="h2" variant="heading">
            {{ t('overview.incident-activity') }}
          </Text><Badge>{{ recentIncidents.length }}</Badge>
        </LayerCardSecondary>
        <LayerCardPrimary>
          <div v-if="incidents.isPending.value" class="loading-state" flex="~ justify-center items-center gap-10px" p="60px" un-text="12px subtle" role="status">
            <Loader :label="t('async-state.loading-data')" />{{ t('async-state.loading-data') }}
          </div>
          <Banner v-else-if="incidents.error.value" variant="error">
            {{ errorText(incidents.error.value) }}
            <Button variant="ghost" @click="incidents.refetch()">
              {{ t('async-state.retry') }}
            </Button>
          </Banner>
          <template v-else>
            <Text v-if="!recentIncidents.length" as="p" size="sm" variant="secondary" class="py-3">
              {{ t('overview.no-incident-announcements-updates-will-appear-here') }}
            </Text>
            <div v-else class="divide-y divide-line">
              <RouterLink v-for="incident in recentIncidents" :key="incident.id" to="/app/incidents" class="flex gap-3 rounded-md py-4 outline-none hover:bg-tint focus-visible:ring-2 focus-visible:ring-brand">
                <span class="mt-1.5 size-2 shrink-0 rounded-full" :class="incident.status === 'resolved' ? 'bg-success' : 'bg-warning'" aria-hidden="true" />
                <div class="min-w-0 space-y-1 [overflow-wrap:anywhere]">
                  <Text as="h3" size="sm" class="font-medium">
                    {{ incident.title }}
                  </Text><Text as="p" size="xs" variant="secondary">
                    {{ statusLabel(incident.status) }} · {{ formatDate(incident.updatedAt) }}
                  </Text>
                </div>
              </RouterLink>
            </div>
          </template>
        </LayerCardPrimary>
      </LayerCard>
      <LayerCard>
        <LayerCardSecondary>
          <Text as="h2" variant="heading">
            {{ t('overview.upcoming-maintenance') }}
          </Text>
        </LayerCardSecondary>
        <LayerCardPrimary>
          <div v-if="maintenance.isPending.value" class="loading-state" flex="~ justify-center items-center gap-10px" p="60px" un-text="12px subtle" role="status">
            <Loader :label="t('async-state.loading-data')" />{{ t('async-state.loading-data') }}
          </div>
          <Banner v-else-if="maintenance.error.value" variant="error">
            {{ errorText(maintenance.error.value) }}
            <Button variant="ghost" @click="maintenance.refetch()">
              {{ t('async-state.retry') }}
            </Button>
          </Banner>
          <template v-else>
            <Text v-if="!nextMaintenance.length" as="p" size="sm" variant="secondary" class="py-3">
              {{ t('overview.no-scheduled-maintenance') }}
            </Text>
            <div v-else class="divide-y divide-line">
              <RouterLink v-for="window in nextMaintenance" :key="window.id" to="/app/maintenance" class="flex gap-3 rounded-md py-4 outline-none hover:bg-tint focus-visible:ring-2 focus-visible:ring-brand">
                <span class="i-lucide-clock mt-1 size-4 shrink-0 text-subtle" aria-hidden="true" />
                <div class="min-w-0 space-y-1 [overflow-wrap:anywhere]">
                  <Text as="h3" size="sm" class="font-medium">
                    {{ window.name }}
                  </Text><Text as="p" size="xs" variant="secondary">
                    {{ formatDate(window.startsAt) }}
                  </Text>
                </div>
              </RouterLink>
            </div>
          </template>
        </LayerCardPrimary>
      </LayerCard>
      <Banner variant="secondary" size="sm">
        {{ t('overview.uptime-reflects-confirmed-duration-paused-maintenance-and-missing') }}
      </Banner>
    </aside>
  </div>
  <LayerCard class="mt-6">
    <LayerCardSecondary class="flex flex-wrap items-center justify-between gap-3">
      <Text as="h2" variant="heading">
        {{ t('overview.public-status-pages') }}
      </Text>
      <Button variant="ghost" size="sm" as-child>
        <RouterLink to="/app/pages">
          {{ t('overview.manage-pages') }}<span class="i-lucide-arrow-up-right size-4" aria-hidden="true" />
        </RouterLink>
      </Button>
    </LayerCardSecondary>
    <LayerCardPrimary>
      <div v-if="pages.isPending.value" class="loading-state" flex="~ justify-center items-center gap-10px" p="60px" un-text="12px subtle" role="status">
        <Loader :label="t('async-state.loading-data')" />{{ t('async-state.loading-data') }}
      </div>
      <Banner v-else-if="pages.error.value" variant="error">
        {{ errorText(pages.error.value) }}
        <Button variant="ghost" @click="pages.refetch()">
          {{ t('async-state.retry') }}
        </Button>
      </Banner>
      <template v-else>
        <Empty v-if="!pages.data.value?.items.length" :title="t('overview.keep-everyone-informed')" :description="t('overview.publish-a-status-page-with-your-own-brand')" size="sm" class="rounded-none border-none px-0! py-4!">
          <template #actions>
            <Button v-if="canEdit()" as-child>
              <RouterLink to="/app/pages/new">
                {{ t('common.create-status-page') }}
              </RouterLink>
            </Button>
          </template>
        </Empty>
        <div v-else class="divide-y divide-line">
          <RouterLink v-for="page in pages.data.value.items.slice(0, 3)" :key="page.id" :to="`/app/pages/${page.id}`" class="flex items-center justify-between gap-4 rounded-md py-3 outline-none hover:bg-tint focus-visible:ring-2 focus-visible:ring-brand">
            <div class="flex min-w-0 items-center gap-3">
              <span class="i-lucide-globe size-5 shrink-0 text-link" aria-hidden="true" />
              <div class="min-w-0">
                <Text as="h3" variant="heading" size="sm" class="[overflow-wrap:anywhere]">
                  {{ page.name }}
                </Text><Text as="p" size="xs" variant="secondary" class="[overflow-wrap:anywhere]">
                  /{{ page.publishedAt ? publishedEntry(page).slug : page.slug }}{{ (page.publishedAt ? publishedEntry(page).domain : page.domain) ? ` · ${page.publishedAt ? publishedEntry(page).domain : page.domain}` : '' }}
                </Text>
              </div>
            </div>
            <Badge :variant="page.publishedAt ? 'success' : 'outline'">
              {{ page.publishedAt ? t('common.published') : t('common.draft') }}
            </Badge>
          </RouterLink>
        </div>
      </template>
    </LayerCardPrimary>
  </LayerCard>
</template>
