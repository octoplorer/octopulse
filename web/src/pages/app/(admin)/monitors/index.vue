<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { ref, computed } from 'vue'
import { Plus, Search, RefreshCw, Globe, Server, ChevronRight } from '@lucide/vue'
import { useQuery, type DefineQueryOptions } from '@pinia/colada'
import { listMonitorsQuery } from '../../../../client/@pinia/colada.gen'
import type { ErrorModel } from '../../../../client/types.gen'
import { useIntervalFn } from '@vueuse/core'
import type { Monitor } from '../../../../lib/types'
import { targetOf, monitorTypes } from '../../../../lib/monitor'
import { formatDate } from '../../../../lib/preferences'
import { canEdit } from '../../../../lib/api'
import PageHeader from '../../../../components/PageHeader.vue'
import StateBadge from '../../../../components/StateBadge.vue'
import EmptyState from '../../../../components/EmptyState.vue'
import AsyncState from '../../../../components/AsyncState.vue'

const { t, locale } = useI18n({ useScope: 'global' })

definePage({ meta: { title: 'common.monitors' } })

const query = useQuery({
    ...listMonitorsQuery(),
    staleTime: 10000,
  } as DefineQueryOptions<{ items: Monitor[] }, ErrorModel>),
  search = ref(''),
  state = ref('all'),
  type = ref('all')
const items = computed(() =>
  [...(query.data.value?.items || [])].sort((a, b) => a.name.localeCompare(b.name, locale.value)),
)
const filtered = computed(() =>
  items.value.filter(
    (m) =>
      (!search.value ||
        `${m.name} ${m.group} ${m.tags?.join(' ')} ${targetOf(m)}`
          .toLowerCase()
          .includes(search.value.toLowerCase())) &&
      (state.value === 'all' ||
        (state.value === 'paused' ? !m.enabled : m.enabled && m.state === state.value)) &&
      (type.value === 'all' || m.type === type.value),
  ),
)
useIntervalFn(() => query.refetch(), 30000)
</script>
<template>
  <PageHeader
    :title="t('common.monitors')"
    :description="t('monitors.defineHealthyBehaviorAndDetectEveryChange')"
    ><button class="button" @click="query.refetch()">
      <RefreshCw :size="14" />{{ t('common.refresh') }}</button
    ><RouterLink v-if="canEdit()" to="/app/monitors/new" class="button primary"
      ><Plus :size="15" />{{ t('common.addMonitor') }}</RouterLink
    ></PageHeader
  >
  <section class="card">
    <div class="filter-bar">
      <div class="search-box">
        <Search :size="16" /><input
          v-model="search"
          :placeholder="t('monitors.searchNameTargetOrTags')"
          :aria-label="t('monitors.searchMonitors')"
        />
      </div>
      <div un-flex="~ items-center gap-2">
        <select v-model="state" un-w="auto!" :aria-label="t('monitors.filterStatus')">
          <option value="all">{{ t('monitors.allStates') }}</option>
          <option value="up">{{ t('monitors.up') }}</option>
          <option value="down">{{ t('monitors.down') }}</option>
          <option value="unknown">{{ t('monitors.unknown') }}</option>
          <option value="paused">{{ t('common.paused') }}</option></select
        ><select v-model="type" un-w="auto!" :aria-label="t('monitors.filterType')">
          <option value="all">{{ t('monitors.allTypes') }}</option>
          <option v-for="item in monitorTypes" :key="item.value" :value="item.value">
            {{ t(item.label) }}
          </option>
        </select>
      </div>
    </div>
    <AsyncState :pending="query.isPending.value" :error="query.error.value" @retry="query.refetch()"
      ><EmptyState
        v-if="!filtered.length"
        :title="
          items.length ? t('monitors.noMatchingMonitors') : t('monitors.startWatchingYourServices')
        "
        :description="
          items.length
            ? t('monitors.tryChangingYourSearchOrFilters')
            : t('monitors.addAServiceCheckToSeeStatusHistory')
        "
        ><RouterLink
          v-if="!items.length && canEdit()"
          to="/app/monitors/new"
          class="button primary"
          >{{ t('common.createMonitor') }}</RouterLink
        ></EmptyState
      >
      <div v-else class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th>{{ t('common.service') }}</th>
              <th>{{ t('common.status') }}</th>
              <th>{{ t('common.group') }}</th>
              <th>{{ t('monitors.intervalRetries') }}</th>
              <th>{{ t('common.lastCheck') }}</th>
              <th />
            </tr>
          </thead>
          <tbody>
            <tr v-for="monitor in filtered" :key="monitor.id">
              <td>
                <RouterLink :to="`/app/monitors/${monitor.id}`" un-flex="~ items-center gap-3"
                  ><span class="monitor-type-icon"
                    ><Globe
                      v-if="monitor.type === 'http' || monitor.type === 'dns'"
                      :size="16" /><Server v-else :size="16" /></span
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
              <td class="muted">{{ monitor.group || '—' }}</td>
              <td>
                {{
                  monitor.type === 'heartbeat'
                    ? monitor.heartbeat?.periodSeconds
                    : monitor.intervalSeconds
                }}
                s
                <span v-if="['http', 'tcp', 'dns'].includes(monitor.type)" class="muted"
                  >/ {{ monitor.retries }}</span
                >
              </td>
              <td class="muted" un-text="10px">{{ formatDate(monitor.lastCheckedAt) }}</td>
              <td>
                <RouterLink
                  :to="`/app/monitors/${monitor.id}`"
                  class="icon-button"
                  :aria-label="t('monitors.viewDetails')"
                  ><ChevronRight :size="16"
                /></RouterLink>
              </td>
            </tr>
          </tbody>
        </table></div
    ></AsyncState>
    <div class="table-footer">
      <span>{{ t('counts.monitors', { count: filtered.length }, filtered.length) }}</span
      ><span>{{ t('common.refreshesEvery30Seconds') }}</span>
    </div>
  </section>
</template>
