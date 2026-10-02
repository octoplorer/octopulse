<script setup lang="ts">
import type { PublicPage } from '../lib/types'
import { usePreferredDark } from '@vueuse/core'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { duration, formatDate, formatPercent, statusLabel } from '../composables/preferences'
import Sparkline from './Sparkline.vue'
import StateBadge from './StateBadge.vue'

const props = defineProps<{
  page: PublicPage
  preview?: boolean
  incidentId?: string
  pathBase?: string
  stale?: boolean
}>()
const { t, n, locale } = useI18n({ useScope: 'global' })
const prefersDark = usePreferredDark()
const dark = computed(
  () =>
    props.page.config.colorScheme === 'dark'
    || (props.page.config.colorScheme === 'system' && prefersDark.value),
)
const stateLabels: Record<string, string> = {
  operational: 'publicState.operational',
  normal: 'publicState.normal',
  up: 'publicState.up',
  partial: 'publicState.partial',
  partial_outage: 'publicState.partial_outage',
  outage: 'publicState.outage',
  full_outage: 'publicState.full_outage',
  maintenance: 'publicState.maintenance',
  unknown: 'publicState.unknown',
  insufficient_data: 'publicState.insufficient_data',
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
const activeMaintenance = computed(() =>
  props.page.maintenance.filter(x => x.endsAt > Date.now()),
)
function showMetric(id: string, field: 'showUptime' | 'showLatency') {
  return (
    props.page.config.groups.flatMap(g => g.monitors).find(m => m.monitorId === id)?.[field]
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
</script>

<template>
  <div
    class="public-page"
    :class="{ preview }"
    :data-theme="dark ? 'dark' : 'light'"
    :style="{ '--accent': page.config.brandColor || '#0c8b76' }"
  >
    <div class="public-inner">
      <header class="public-header">
        <div class="public-brand">
          <img v-if="page.config.logoUrl" :src="page.config.logoUrl" alt=""><span
            v-else
            class="brand-icon"
            :style="{ background: page.config.brandColor || '#0c8b76' }"
          ><span class="i-lucide-activity" un-w="20px" un-h="20px" aria-hidden="true" /></span>{{ page.config.title || 'Octopulse' }}
        </div>
        <nav class="public-nav">
          <a
            v-for="link in page.config.links"
            :key="link.url"
            :href="safeLink(link.url)"
            target="_blank"
            rel="noopener noreferrer"
          >{{ link.label }}<span un-ml="1" class="i-lucide-arrow-up-right" un-w="11px" un-h="11px" aria-hidden="true" /></a><button
            v-if="!preview"
            class="icon-button"
            :aria-label="t('common.switchLanguage')"
            @click="locale = locale === 'zh-CN' ? 'en' : 'zh-CN'"
          >
            <span class="i-lucide-languages" un-w="16px" un-h="16px" aria-hidden="true" /><span un-text="10px" un-ml="1">{{
              locale === 'zh-CN' ? 'EN' : '中'
            }}</span>
          </button>
        </nav>
      </header>
      <div v-if="stale" class="note" role="status" un-mb="5">
        {{ t('statusPage.statusDataCouldNotBeRefreshedTheLast') }}
      </div>
      <p v-if="page.config.description" class="status-page-description">
        {{ page.config.description }}
      </p>
      <div v-if="!incidentId" class="public-overall">
        <span
          class="public-overall-icon"
          :style="!good ? { background: '#e8b65719', color: '#c18a34' } : {}"
        ><span v-if="good" class="i-lucide-check" un-w="24px" un-h="24px" aria-hidden="true" /><span
          v-else-if="page.state === 'maintenance'" class="i-lucide-clock" un-w="23px" un-h="23px" aria-hidden="true"
        /><span v-else class="i-lucide-triangle-alert" un-w="23px" un-h="23px" aria-hidden="true" /></span>
        <div>
          <h1>{{ heading }}</h1>
          <p>
            {{ t('statusPage.lastUpdated') }} {{ formatDate(page.updatedAt)
            }}{{ preview ? ` · ${t('statusPage.draftPreview')}` : '' }}
          </p>
        </div>
      </div>
      <template v-if="!incidentId">
        <section
          v-for="window in activeMaintenance"
          :key="window.id"
          class="card public-incident"
          :style="{ 'border-left-color': '#6b92df' }"
        >
          <div un-flex="~ items-center justify-between gap-4">
            <h3>{{ window.name }}</h3>
            <StateBadge state="maintenance" />
          </div>
          <p>{{ window.description }}</p>
          <small>{{ formatDate(window.startsAt) }} — {{ formatDate(window.endsAt) }}</small>
        </section>
        <section v-for="group in page.groups" :key="group.id" class="public-group">
          <h2>{{ group.name }}</h2>
          <div class="card">
            <article v-for="monitor in group.monitors" :key="monitor.id" class="public-monitor">
              <div class="public-monitor-head">
                <h3>{{ monitor.name }}</h3>
                <StateBadge
                  :state="
                    monitor.type === 'certificate' ? monitor.certificate?.state : monitor.state
                  "
                  :paused="monitor.paused"
                  :maintenance="monitor.maintenance"
                />
              </div>
              <template v-if="monitor.certificate">
                <div class="metric-row">
                  <span>{{ t('statusPage.timeRemaining') }}
                    {{
                      monitor.certificate.expiresAt
                        ? n(monitor.certificate.daysRemaining, {
                          minimumFractionDigits: 1,
                          maximumFractionDigits: 1,
                        })
                        : '—'
                    }}
                    {{ t('common.days') }}</span><span>{{ t('statusPage.expires') }}
                    {{ formatDate(monitor.certificate.expiresAt) }}</span>
                </div>
              </template><template v-else>
                <Sparkline
                  v-if="showMetric(monitor.id, 'showLatency') && monitor.latency?.length"
                  :values="monitor.latency.map((x) => x.latencyMs)"
                  :timestamps="monitor.latency.map((x) => x.at)"
                  :height="30"
                  :color="page.config.brandColor"
                />
                <div v-if="showMetric(monitor.id, 'showUptime')" class="metric-row">
                  <span>{{ t('common.uptime') }}
                    <strong>{{ formatPercent(monitor.availability?.uptime) }}</strong><span un-ml="3">{{ t('statusPage.coverage') }}
                      {{ formatPercent(monitor.availability?.coverage) }}</span></span><span>{{ t('statusPage.effectiveDuration') }}
                    {{ duration(monitor.availability?.effectiveMs) }}</span>
                </div>
              </template>
            </article>
            <p v-if="!group.monitors.length" class="muted" un-p="5">
              {{ t('statusPage.noServicesInThisGroup') }}
            </p>
          </div>
        </section>
        <p v-if="!page.groups.length" class="muted" un-py="6">
          {{ t('statusPage.noServicesHaveBeenPublishedYet') }}
        </p>
      </template>
      <section v-if="incidents.length">
        <h2 class="public-incidents-title">
          {{ incidentId ? t('statusPage.incidentUpdates') : t('statusPage.incidentAnnouncements') }}
        </h2>
        <article
          v-for="incident in incidents"
          :key="incident.id"
          class="card public-incident"
          :style="incident.status === 'resolved' ? { 'border-left-color': 'var(--accent)' } : {}"
        >
          <div un-flex="~ items-center justify-between gap-4">
            <h3>
              <RouterLink v-if="!preview && !incidentId" :to="`${base}/incidents/${incident.id}`">
                {{
                  incident.title
                }}
              </RouterLink><template v-else>
                {{ incident.title }}
              </template>
            </h3>
            <span class="pill">{{ statusLabel(incident.status) }}</span>
          </div>
          <p>{{ incident.body }}</p>
          <small>{{ formatDate(incident.createdAt) }}</small>
          <div v-if="incident.updates?.length" class="timeline" un-mt="6">
            <div
              v-for="update in [...incident.updates].reverse()"
              :key="update.id"
              class="timeline-entry"
            >
              <h3>{{ statusLabel(update.status) }}</h3>
              <p>{{ update.body }}</p>
              <small>{{ formatDate(update.createdAt) }}</small>
            </div>
          </div>
        </article>
      </section>
      <RouterLink v-if="incidentId" :to="base || '/'" class="button" un-mt="5">
        {{
          t('statusPage.backToStatusPage')
        }}
      </RouterLink>
      <footer class="public-footer">
        <span un-flex="~ items-center gap-1.5"><span class="i-lucide-activity" un-w="13px" un-h="13px" aria-hidden="true" />{{ t('publicPage.poweredBy') }}</span>
      </footer>
    </div>
  </div>
</template>
