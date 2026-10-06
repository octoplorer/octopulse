<script setup lang="ts">
import type { PublicPage } from '../client/types.gen'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { duration, formatDate, formatPercent, statusLabel } from '../composables/preferences'
import Sparkline from './Sparkline.vue'
import StateBadge from './StateBadge.vue'
import { Alert } from './ui/alert'
import { Badge } from './ui/badge'
import { Button } from './ui/button'
import { Card } from './ui/card'

const props = defineProps<{
  page: PublicPage
  preview?: boolean
  incidentId?: string
  pathBase?: string
  stale?: boolean
}>()
const { t, n, locale } = useI18n({ useScope: 'global' })
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
    :class="{ preview }"
    :style="{ '--color-brand': page.config.brandColor || undefined }" bg="canvas" un-text="default" min-h="screen" py="48px" px="24px" font="sans"
    class="[@media(max-width:700px)]:px-17px [@media(max-width:700px)]:py-25px [@media(max-width:700px)]:[&_.public-incident>div]:flex-wrap [&.preview]:min-h-0 [&.preview]:border-1 [&.preview]:border-solid [&.preview]:border-line [&.preview]:rounded-11px [&.preview]:p-23px [&.preview_.public-header]:mb-15px [&.preview_.public-brand]:text-18px [&.preview_.public-overall]:mt-10px [&.preview_.public-overall]:p-20px [&.preview_.public-overall_h1]:text-16px [&.preview_.public-monitor]:p-17px [&.preview_.public-footer]:text-12px [&.preview_.public-nav]:hidden [&.preview_.public-monitor-head]:flex-wrap"
  >
    <div class="public-inner" max-w="870px" mx="auto">
      <header class="public-header [@media(max-width:700px)]:items-start [@media(max-width:700px)]:flex-col [@media(max-width:700px)]:gap-13px" flex="~ items-center justify-between gap-20px" mb="30px">
        <div class="public-brand [@media(max-width:700px)]:text-21px" flex="~ items-center gap-12px" font="[var(--font-sans)] 600" un-text="22px" tracking="-0.5px" break="anywhere">
          <img v-if="page.config.logoUrl" :src="page.config.logoUrl" alt="" size="36px" object="contain"><span
            v-else
            class="brand-icon shadow-control"
            bg="brand" flex="~ items-center justify-center shrink-0" size="32px" un-text="on-brand" border="1px solid line" rounded="8px"
          ><span class="i-lucide-activity" w="20px" h="20px" aria-hidden="true" /></span>{{ page.config.title || 'Octopulse' }}
        </div>
        <nav class="public-nav [@media(max-width:700px)]:flex-wrap [@media(max-width:700px)]:gap-13px" flex="~ items-center wrap gap-17px" un-text="12px subtle">
          <a
            v-for="link in page.config.links"
            :key="link.url"
            :href="safeLink(link.url)"
            target="_blank"
            rel="noopener noreferrer"
          >{{ link.label }}<span ml="1" class="i-lucide-arrow-up-right" w="11px" h="11px" aria-hidden="true" /></a><Button
            v-if="!preview"
            :aria-label="t('common.switchLanguage')"
            shape="square"
            @click="locale = locale === 'zh-CN' ? 'en' : 'zh-CN'"
          >
            <span class="i-lucide-languages" w="16px" h="16px" aria-hidden="true" /><span un-text="xs" ml="1">{{
              locale === 'zh-CN' ? 'EN' : '中'
            }}</span>
          </Button>
        </nav>
      </header>
      <Alert v-if="stale" role="status" mb="5">
        {{ t('statusPage.statusDataCouldNotBeRefreshedTheLast') }}
      </Alert>
      <p v-if="page.config.description" class="status-page-description" un-text="13px subtle" max-w="600px" mb="25px">
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
        <div>
          <h1 un-text="22px" tracking="-0.5px">
            {{ heading }}
          </h1>
          <p un-text="12px subtle" mt="6px">
            {{ t('statusPage.lastUpdated') }} {{ formatDate(page.updatedAt)
            }}{{ preview ? ` · ${t('statusPage.draftPreview')}` : '' }}
          </p>
        </div>
      </div>
      <template v-if="!incidentId">
        <Card
          v-for="window in activeMaintenance"
          :key="window.id"
          class="public-incident"
          border="l-3px l-solid l-info"
          py="19px"
          px="22px"
          mb="16px"
          as="section"
        >
          <div flex="~ items-center justify-between gap-4">
            <h3>{{ window.name }}</h3>
            <StateBadge state="maintenance" />
          </div>
          <p un-text="12px subtle" mt="9px" whitespace="pre-wrap">
            {{ window.description }}
          </p>
          <small un-text="12px subtle">{{ formatDate(window.startsAt) }} — {{ formatDate(window.endsAt) }}</small>
        </Card>
        <section v-for="group in page.groups" :key="group.id" class="public-group" my="26px">
          <h2 un-text="14px" mb="13px">
            {{ group.name }}
          </h2>
          <Card>
            <article v-for="monitor in group.monitors" :key="monitor.id" class="public-monitor [@media(max-width:700px)]:px-17px [@media(max-width:700px)]:py-19px" py="22px" px="25px" border="b-1px b-solid b-line last:b-0">
              <div class="public-monitor-head [@media(max-width:700px)]:flex-wrap" flex="~ items-center justify-between gap-12px" mb="15px">
                <h3 un-text="13px" font="600" break="anywhere">
                  {{ monitor.name }}
                </h3>
                <StateBadge
                  :state="
                    monitor.type === 'certificate' ? monitor.certificate?.state : monitor.state
                  "
                  :paused="monitor.paused"
                  :maintenance="monitor.maintenance"
                />
              </div>
              <template v-if="monitor.certificate">
                <div class="tabular-nums [@media(max-width:700px)]:gap-8px" flex="~ items-center justify-between wrap" un-text="12px subtle" mt="9px">
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
                <div v-if="showMetric(monitor.id, 'showUptime')" class="tabular-nums [@media(max-width:700px)]:gap-8px" flex="~ items-center justify-between wrap" un-text="12px subtle" mt="9px">
                  <span>{{ t('common.uptime') }}
                    <strong>{{ formatPercent(monitor.availability?.uptime) }}</strong><span ml="3">{{ t('statusPage.coverage') }}
                      {{ formatPercent(monitor.availability?.coverage) }}</span></span><span>{{ t('statusPage.effectiveDuration') }}
                    {{ duration(monitor.availability?.effectiveMs) }}</span>
                </div>
              </template>
            </article>
            <p v-if="!group.monitors.length" p="5" un-text="13px subtle">
              {{ t('statusPage.noServicesInThisGroup') }}
            </p>
          </Card>
        </section>
        <p v-if="!page.groups.length" py="6" un-text="13px subtle">
          {{ t('statusPage.noServicesHaveBeenPublishedYet') }}
        </p>
      </template>
      <section v-if="incidents.length">
        <h2 class="public-incidents-title" un-text="15px" mt="30px" mb="14px">
          {{ incidentId ? t('statusPage.incidentUpdates') : t('statusPage.incidentAnnouncements') }}
        </h2>
        <Card
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
          <div flex="~ items-center justify-between gap-4">
            <h3>
              <RouterLink v-if="!preview && !incidentId" :to="`${base}/incidents/${incident.id}`">
                {{
                  incident.title
                }}
              </RouterLink><template v-else>
                {{ incident.title }}
              </template>
            </h3>
            <Badge variant="outline">
              {{ statusLabel(incident.status) }}
            </Badge>
          </div>
          <p un-text="12px subtle" mt="9px" whitespace="pre-wrap">
            {{ incident.body }}
          </p>
          <small un-text="12px subtle">{{ formatDate(incident.createdAt) }}</small>
          <div v-if="incident.updates?.length" mt="6" ml="5px" pl="21px" border="l-1 solid line" class="[&_.timeline-entry]:relative [&_.timeline-entry]:pb-24px [&_.timeline-entry]:before:content-empty [&_.timeline-entry]:before:absolute [&_.timeline-entry]:before:left-[-26px] [&_.timeline-entry]:before:top-5px [&_.timeline-entry]:before:size-9px [&_.timeline-entry]:before:rounded-full [&_.timeline-entry]:before:border-2 [&_.timeline-entry]:before:border-solid [&_.timeline-entry]:before:border-base [&_.timeline-entry]:before:bg-brand [&_.timeline-entry_h3]:text-12px [&_.timeline-entry_p]:mt-6px [&_.timeline-entry_p]:whitespace-pre-wrap [&_.timeline-entry_p]:text-12px [&_.timeline-entry_p]:text-subtle [&_.timeline-entry_small]:text-12px [&_.timeline-entry_small]:text-subtle">
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
        </Card>
      </section>
      <Button v-if="incidentId" as-child>
        <RouterLink :to="base || '/'" mt="5">
          {{
            t('statusPage.backToStatusPage')
          }}
        </RouterLink>
      </Button>
      <footer class="public-footer [@media(max-width:700px)]:gap-15px" flex="~ items-center justify-between" border="t-1px t-solid t-line" mt="34px" pt="20px" un-text="12px subtle">
        <span flex="~ items-center gap-1.5"><span class="i-lucide-activity" w="13px" h="13px" aria-hidden="true" />{{ t('publicPage.poweredBy') }}</span>
      </footer>
    </div>
  </div>
</template>
