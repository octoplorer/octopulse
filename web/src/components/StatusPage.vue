<script setup lang="ts">
import type { PublicMonitor, PublicPage } from '../client/types.gen'
import { useIntervalFn, useNow } from '@vueuse/core'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { duration, formatDate, statusLabel } from '../composables/preferences'
import { dailyAvailabilityOption } from '../lib/chart'
import { monitorStateDisplay } from '../lib/monitor-state'
import { Badge } from './ui/badge'
import { Banner } from './ui/banner'
import { Button } from './ui/button'
import { Chart } from './ui/chart'
import { Empty, EmptyDescription, EmptyTitle } from './ui/empty'
import { LayerCard } from './ui/layer-card'
import { Link } from './ui/link'

const props = defineProps<{
  page: PublicPage
  preview?: boolean
  incidentId?: string
  pathBase?: string
  stale?: boolean
}>()
const { t, n, d, locale } = useI18n({ useScope: 'global' })
const stateLabels: Record<string, string> = {
  operational: 'public-state.operational',
  normal: 'public-state.normal',
  up: 'public-state.up',
  partial: 'public-state.partial',
  partial_outage: 'public-state.partial-outage',
  outage: 'public-state.outage',
  full_outage: 'public-state.full-outage',
  maintenance: 'public-state.maintenance',
  unknown: 'public-state.unknown',
  insufficient_data: 'public-state.insufficient-data',
}
const heading = computed(() =>
  stateLabels[props.page.state] ? t(stateLabels[props.page.state]!) : props.page.state,
)
const good = computed(() => ['up', 'normal', 'operational'].includes(props.page.state))
const incidents = computed(() =>
  props.incidentId
    ? props.page.incidents.filter(x => x.id === props.incidentId)
    : props.page.incidents,
)
const now = useNow({ scheduler: callback => useIntervalFn(callback, 30000) })
const activeMaintenance = computed(() =>
  props.page.maintenance
    .filter(window => window.endsAt > now.value.getTime())
    .sort((a, b) => a.startsAt - b.startsAt)
    .map(window => ({ ...window, scheduled: window.startsAt > now.value.getTime() })),
)
const maintenanceState = computed(() => monitorStateDisplay('maintenance'))
const groups = computed(() => props.page.groups.map(group => ({
  ...group,
  monitors: group.monitors.map(monitor => ({
    ...monitor,
    stateDisplay: monitorStateDisplay(
      monitor.type === 'certificate' ? monitor.certificate?.state : monitor.state,
      { paused: monitor.paused, maintenance: monitor.maintenance },
    ),
  })),
})))
function showMetric(groupId: string, id: string, field: 'showUptime' | 'showLatency') {
  return (
    props.page.config.groups.find(g => g.id === groupId)?.monitors.find(m => m.monitorId === id)?.[field]
    ?? true
  )
}
function safeLink(value: string) {
  try {
    const url = new URL(value)
    if (['https:', 'http:'].includes(url.protocol))
      return url.href
  }
  catch {}
  return undefined
}
const base = computed(() => props.pathBase ?? `/${props.page.slug}`)
function availabilityPercent(value: number | null | undefined) {
  return value == null ? '—' : n(value / 100, { style: 'percent', maximumFractionDigits: 2 })
}
function utcDay(at: number) {
  return d(at, { year: 'numeric', month: '2-digit', day: '2-digit', timeZone: 'UTC' })
}
function latencyText(monitor: PublicMonitor) {
  const point = (monitor.latency ?? []).filter(p => Number.isFinite(p.latencyMs)).sort((a, b) => a.at - b.at).at(-1)
  return point ? `${n(point.latencyMs, { maximumFractionDigits: 2 })} ms` : '—'
}
const dailyCharts = computed(() => new Map(props.page.groups.flatMap(group => group.monitors.map((monitor) => {
  const points = monitor.dailyAvailability ?? []
  return [`${group.id}:${monitor.id}`, {
    hasData: points.length > 0,
    option: dailyAvailabilityOption({
      points,
      name: t('common.uptime'),
      color: props.page.config.brandColor || 'var(--color-brand)',
      valueFormatter: availabilityPercent,
      dayFormatter: at => `${utcDay(at)} UTC`,
      noDataLabel: t('status-page.no-effective-observations'),
      coverageLabel: t('status-page.coverage'),
    }),
    label: t('status-page.daily-chart-summary', {
      name: monitor.name,
      days: points.length,
      observed: points.filter(point => point.uptime != null).length,
      from: points.length ? utcDay(points[0]!.from) : '—',
      to: points.length ? utcDay(points.at(-1)!.from) : '—',
    }),
  }] as const
}))))
</script>

<template>
  <component
    :is="preview ? 'div' : 'main'"
    :class="{ preview }"
    :style="{ '--color-brand': page.config.brandColor || undefined }" bg="canvas" un-text="default" min-h="screen" py="48px" px="24px" font="sans"
    class="[@media(max-width:700px)]:px-17px [@media(max-width:700px)]:py-25px [@media(max-width:700px)]:[&_.public-incident>div]:flex-wrap [&.preview]:min-h-0 [&.preview]:border-1 [&.preview]:border-solid [&.preview]:border-line [&.preview]:rounded-11px [&.preview]:p-23px [&.preview_.public-header]:mb-15px [&.preview_.public-brand]:text-18px [&.preview_.public-overall]:mt-10px [&.preview_.public-overall]:p-20px [&.preview_.public-overall_h1]:text-16px [&.preview_.public-monitor]:p-17px [&.preview_.public-footer]:text-12px [&.preview_.public-nav]:hidden [&.preview_.public-monitor-head]:flex-wrap"
  >
    <div class="public-inner min-w-0" max-w="870px" mx="auto">
      <header class="public-header [@media(max-width:700px)]:items-start [@media(max-width:700px)]:flex-col [@media(max-width:700px)]:gap-13px" flex="~ items-center justify-between gap-20px" mb="30px">
        <div class="public-brand min-w-0 [@media(max-width:700px)]:text-21px" flex="~ items-center gap-12px" font="[var(--font-sans)] 600" un-text="22px" tracking="-0.5px">
          <img v-if="page.config.logoUrl" :src="page.config.logoUrl" alt="" size="36px" object="contain" class="shrink-0"><span
            v-else
            class="brand-icon shadow-control"
            bg="brand" flex="~ items-center justify-center shrink-0" size="32px" un-text="on-brand" border="1px solid line" rounded="8px"
          ><span class="i-lucide-activity" w="20px" h="20px" aria-hidden="true" /></span><span class="min-w-0 [overflow-wrap:anywhere]">{{ page.config.title || 'Octopulse' }}</span>
        </div>
        <nav class="public-nav min-w-0 [@media(max-width:700px)]:flex-wrap [@media(max-width:700px)]:gap-13px" flex="~ items-center wrap gap-17px" un-text="12px subtle">
          <Link
            v-for="link in page.config.links"
            :key="link.url"
            :href="safeLink(link.url)"
            target="_blank"
            rel="noopener noreferrer"
            class="min-w-0 [overflow-wrap:anywhere]"
          >
            {{ link.label }}<span ml="1" class="i-lucide-arrow-up-right" w="11px" h="11px" aria-hidden="true" />
          </Link><Button
            v-if="!preview"
            :aria-label="`${t('common.switch-language')} (${locale === 'zh-CN' ? 'EN' : '中'})`"
            @click="locale = locale === 'zh-CN' ? 'en' : 'zh-CN'"
          >
            <template #icon>
              <span class="i-lucide-languages" size="16px" shrink="0" aria-hidden="true" />
            </template>
            <span un-text="xs">{{
              locale === 'zh-CN' ? 'EN' : '中'
            }}</span>
          </Button>
        </nav>
      </header>
      <Banner v-if="stale" role="status" mb="5">
        {{ t('status-page.status-data-could-not-be-refreshed-the-last') }}
      </Banner>
      <p v-if="page.config.description" class="status-page-description leading-relaxed [overflow-wrap:anywhere]" un-text="13px subtle" max-w="600px" mb="25px">
        {{ page.config.description }}
      </p>
      <div v-if="!incidentId" class="public-overall [@media(max-width:700px)]:px-18px [@media(max-width:700px)]:py-22px [@media(max-width:700px)]:[&_h1]:text-19px" p="30px" mt="20px" mb="29px" border="1px solid line" rounded="12px" bg="base" flex="~ items-center gap-17px">
        <span
          class="public-overall-icon"
          size="40px"
          rounded="full"
          flex="~ items-center justify-center shrink-0"
          :class="good ? 'bg-success-tint text-fg-success' : page.state === 'maintenance' ? 'bg-info-tint text-fg-info' : 'bg-warning-tint text-fg-warning'"
        ><span v-if="good" class="i-lucide-check" w="24px" h="24px" aria-hidden="true" /><span
          v-else-if="page.state === 'maintenance'" class="i-lucide-clock" w="23px" h="23px" aria-hidden="true"
        /><span v-else class="i-lucide-triangle-alert" w="23px" h="23px" aria-hidden="true" /></span>
        <div class="min-w-0">
          <h1 un-text="22px" tracking="-0.5px">
            {{ heading }}
          </h1>
          <p class="tabular-nums [overflow-wrap:anywhere]" un-text="12px subtle" mt="6px">
            {{ t('status-page.last-updated') }} {{ formatDate(page.updatedAt)
            }}{{ preview ? ` · ${t('status-page.draft-preview')}` : '' }}
          </p>
        </div>
      </div>
      <template v-if="!incidentId">
        <LayerCard
          v-for="window in activeMaintenance"
          :key="window.id"
          class="public-incident"
          border="l-3px l-solid l-info"
          py="19px"
          px="22px"
          mb="16px"
        >
          <div flex="~ wrap items-center justify-between gap-4">
            <h3 class="min-w-0 font-semibold [overflow-wrap:anywhere]">
              {{ window.name }}
            </h3>
            <Badge :variant="window.scheduled ? 'outline' : maintenanceState.variant" :data-state="window.scheduled ? 'scheduled' : maintenanceState.state" dot>
              {{ window.scheduled ? t('maintenance.scheduled') : maintenanceState.label }}
            </Badge>
          </div>
          <p class="leading-relaxed [overflow-wrap:anywhere]" un-text="12px subtle" mt="9px" whitespace="pre-wrap">
            {{ window.description }}
          </p>
          <small class="tabular-nums [overflow-wrap:anywhere]" un-text="12px subtle">{{ formatDate(window.startsAt) }} — {{ formatDate(window.endsAt) }}</small>
        </LayerCard>
        <section v-for="group in groups" :key="group.id" class="public-group" my="26px">
          <h2 class="[overflow-wrap:anywhere]" un-text="14px" mb="13px">
            {{ group.name }}
          </h2>
          <LayerCard>
            <article v-for="monitor in group.monitors" :key="monitor.id" class="public-monitor [@media(max-width:700px)]:px-17px [@media(max-width:700px)]:py-19px" py="22px" px="25px" border="b-1px b-solid b-line last:b-0">
              <div class="public-monitor-head [@media(max-width:700px)]:flex-wrap" flex="~ items-center justify-between gap-12px" mb="15px">
                <h3 un-text="13px" font="600" break="anywhere">
                  {{ monitor.name }}
                </h3>
                <Badge :variant="monitor.stateDisplay.variant" :data-state="monitor.stateDisplay.state" dot>
                  {{ monitor.stateDisplay.label }}
                </Badge>
              </div>
              <template v-if="monitor.certificate">
                <div class="tabular-nums [@media(max-width:700px)]:gap-8px" flex="~ items-center justify-between wrap" un-text="12px subtle" mt="9px">
                  <span>{{ t('status-page.time-remaining') }}
                    {{
                      monitor.certificate.expiresAt
                        ? n(monitor.certificate.daysRemaining, {
                          minimumFractionDigits: 1,
                          maximumFractionDigits: 1,
                        })
                        : '—'
                    }}
                    {{ t('common.days') }}</span><span>{{ t('status-page.expires') }}
                    {{ formatDate(monitor.certificate.expiresAt) }}</span>
                </div>
              </template><template v-else>
                <template v-if="showMetric(group.id, monitor.id, 'showUptime')">
                  <div flex="~ items-center justify-between wrap gap-2" un-text="11px subtle">
                    <span>{{ t('status-page.daily-availability') }} · {{ t('status-page.last-90-days') }} (UTC)</span>
                    <span flex="~ items-center gap-1"><span size="2" rounded="full" bg="fill" border="1px solid line" />{{ t('status-page.no-effective-observations') }}</span>
                  </div>
                  <Chart
                    v-if="dailyCharts.get(`${group.id}:${monitor.id}`)?.hasData"
                    :option="dailyCharts.get(`${group.id}:${monitor.id}`)!.option"
                    :height="64"
                    :aria-label="dailyCharts.get(`${group.id}:${monitor.id}`)!.label"
                  />
                  <p v-else un-text="11px subtle" my="3">
                    {{ t('chart.no-observations') }}
                  </p>
                  <div class="tabular-nums [@media(max-width:700px)]:gap-8px" flex="~ items-center justify-between wrap gap-2" un-text="12px subtle" mt="9px">
                    <span>{{ t('status-page.last-24-hours') }} · {{ t('common.uptime') }}
                      <strong>{{ availabilityPercent(monitor.availability?.uptime) }}</strong><span ml="3">{{ t('status-page.coverage') }}
                        {{ availabilityPercent(monitor.availability?.coverage) }}</span></span><span>{{ t('status-page.effective-duration') }}
                      {{ duration(monitor.availability?.effectiveMs) }}</span>
                  </div>
                </template>
                <p v-if="showMetric(group.id, monitor.id, 'showLatency')" class="tabular-nums" un-text="12px subtle" mt="9px">
                  {{ t('status-page.recent-latency') }} {{ latencyText(monitor) }}
                </p>
              </template>
            </article>
            <p v-if="!group.monitors.length" p="5" un-text="13px subtle">
              {{ t('status-page.no-services-in-this-group') }}
            </p>
          </LayerCard>
        </section>
        <p v-if="!page.groups.length" py="6" un-text="13px subtle">
          {{ t('status-page.no-services-have-been-published-yet') }}
        </p>
      </template>
      <section v-if="incidents.length">
        <h2 class="public-incidents-title" un-text="15px" mt="30px" mb="14px">
          {{ incidentId ? t('status-page.incident-updates') : t('status-page.incident-announcements') }}
        </h2>
        <LayerCard
          v-for="incident in incidents"
          :key="incident.id"
          class="public-incident"
          border="l-3px l-solid"
          py="19px"
          px="22px"
          mb="16px"
          :class="incident.status === 'resolved' ? 'border-l-success' : 'border-l-warning'"
          as="article"
        >
          <div flex="~ wrap items-center justify-between gap-4">
            <component :is="incidentId ? 'h1' : 'h3'" class="min-w-0 font-semibold [overflow-wrap:anywhere]">
              <RouterLink v-if="!preview && !incidentId" :to="`${base}/incidents/${incident.id}`">
                {{
                  incident.title
                }}
              </RouterLink><template v-else>
                {{ incident.title }}
              </template>
            </component>
            <Badge variant="outline">
              {{ statusLabel(incident.status) }}
            </Badge>
          </div>
          <p class="max-w-prose leading-relaxed [overflow-wrap:anywhere]" un-text="12px subtle" mt="9px" whitespace="pre-wrap">
            {{ incident.body }}
          </p>
          <small class="tabular-nums" un-text="12px subtle">{{ formatDate(incident.createdAt) }}</small>
          <div v-if="incident.updates?.length" mt="6" ml="5px" pl="21px" border="l-1 solid line" class="[&_.timeline-entry]:relative [&_.timeline-entry]:pb-24px [&_.timeline-entry]:before:content-empty [&_.timeline-entry]:before:absolute [&_.timeline-entry]:before:left-[-26px] [&_.timeline-entry]:before:top-5px [&_.timeline-entry]:before:size-9px [&_.timeline-entry]:before:rounded-full [&_.timeline-entry]:before:border-2 [&_.timeline-entry]:before:border-solid [&_.timeline-entry]:before:border-base [&_.timeline-entry]:before:bg-brand [&_.timeline-entry_h3]:text-12px [&_.timeline-entry_p]:mt-6px [&_.timeline-entry_p]:whitespace-pre-wrap [&_.timeline-entry_p]:text-12px [&_.timeline-entry_p]:text-subtle [&_.timeline-entry_small]:text-12px [&_.timeline-entry_small]:text-subtle">
            <div
              v-for="update in [...incident.updates].reverse()"
              :key="update.id"
              class="timeline-entry [overflow-wrap:anywhere]"
            >
              <h3>{{ statusLabel(update.status) }}</h3>
              <p class="max-w-prose leading-relaxed">
                {{ update.body }}
              </p>
              <small>{{ formatDate(update.createdAt) }}</small>
            </div>
          </div>
        </LayerCard>
      </section>
      <Empty v-if="incidentId && !incidents.length" size="sm">
        <EmptyTitle as="h1">
          {{ t('status-page.incident-unavailable') }}
        </EmptyTitle>
        <EmptyDescription>{{ t('status-page.incident-unavailable-description') }}</EmptyDescription>
      </Empty>
      <Button v-if="incidentId" as-child>
        <RouterLink :to="base || '/'" mt="5">
          {{
            t('status-page.back-to-status-page')
          }}
        </RouterLink>
      </Button>
      <footer class="public-footer [@media(max-width:700px)]:gap-15px" flex="~ items-center justify-between" border="t-1px t-solid t-line" mt="34px" pt="20px" un-text="12px subtle">
        <span flex="~ items-center gap-1.5"><span class="i-lucide-activity" w="13px" h="13px" aria-hidden="true" />{{ t('public-page.powered-by') }}</span>
      </footer>
    </div>
  </component>
</template>
