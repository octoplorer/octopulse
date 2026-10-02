<script setup lang="ts">
import { ref, computed } from 'vue'
import { Plus, Search, RefreshCw, Globe, Server, ChevronRight } from '@lucide/vue'
import { useCollection } from '../lib/data'
import { useIntervalFn } from '@vueuse/core'
import type { Monitor } from '../lib/types'
import { targetOf, monitorTypes } from '../lib/monitor'
import { t, formatDate } from '../lib/preferences'
import { canEdit } from '../lib/api'
import PageHeader from '../components/PageHeader.vue'
import StateBadge from '../components/StateBadge.vue'
import EmptyState from '../components/EmptyState.vue'
import AsyncState from '../components/AsyncState.vue'
const query = useCollection<Monitor>('monitors'),
  search = ref(''),
  state = ref('all'),
  type = ref('all')
const items = computed(() =>
  [...(query.data.value?.items || [])].sort((a, b) => a.name.localeCompare(b.name)),
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
useIntervalFn(() => query.refresh(), 30000)
</script>
<template>
  <PageHeader
    :title="t('监控项', 'Monitors')"
    :description="
      t(
        '定义服务的健康标准，及时发现每一次变化。',
        'Define healthy behavior and detect every change.',
      )
    "
    ><button class="button" @click="query.refresh()">
      <RefreshCw :size="14" />{{ t('刷新', 'Refresh') }}</button
    ><RouterLink v-if="canEdit()" to="/app/monitors/new" class="button primary"
      ><Plus :size="15" />{{ t('添加监控项', 'Add monitor') }}</RouterLink
    ></PageHeader
  >
  <section class="card">
    <div class="filter-bar">
      <div class="search-box">
        <Search :size="16" /><input
          v-model="search"
          :placeholder="t('搜索名称、目标或标签…', 'Search name, target, or tags…')"
          :aria-label="t('搜索监控项', 'Search monitors')"
        />
      </div>
      <div un-flex="~ items-center gap-2">
        <select v-model="state" un-w="auto!" :aria-label="t('状态筛选', 'Filter status')">
          <option value="all">{{ t('全部状态', 'All states') }}</option>
          <option value="up">{{ t('正常', 'Up') }}</option>
          <option value="down">{{ t('故障', 'Down') }}</option>
          <option value="unknown">{{ t('等待数据', 'Unknown') }}</option>
          <option value="paused">{{ t('暂停', 'Paused') }}</option></select
        ><select v-model="type" un-w="auto!" :aria-label="t('类型筛选', 'Filter type')">
          <option value="all">{{ t('全部类型', 'All types') }}</option>
          <option v-for="item in monitorTypes" :key="item.value" :value="item.value">
            {{ t(item.zh, item.en) }}
          </option>
        </select>
      </div>
    </div>
    <AsyncState :pending="query.isPending.value" :error="query.error.value" @retry="query.refresh()"
      ><EmptyState
        v-if="!filtered.length"
        :title="
          items.length
            ? t('没有匹配的监控项', 'No matching monitors')
            : t('开始守护你的服务', 'Start watching your services')
        "
        :description="
          items.length
            ? t('调整搜索或筛选条件。', 'Try changing your search or filters.')
            : t(
                '添加服务检查后，可查看状态、历史与诊断信息。',
                'Add a service check to see status, history, and diagnostics.',
              )
        "
        ><RouterLink
          v-if="!items.length && canEdit()"
          to="/app/monitors/new"
          class="button primary"
          >{{ t('创建监控项', 'Create monitor') }}</RouterLink
        ></EmptyState
      >
      <div v-else class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th>{{ t('服务', 'Service') }}</th>
              <th>{{ t('状态', 'Status') }}</th>
              <th>{{ t('分组', 'Group') }}</th>
              <th>{{ t('间隔 / 重试', 'Interval / retries') }}</th>
              <th>{{ t('最近检查', 'Last check') }}</th>
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
                  :aria-label="t('查看详情', 'View details')"
                  ><ChevronRight :size="16"
                /></RouterLink>
              </td>
            </tr>
          </tbody>
        </table></div
    ></AsyncState>
    <div class="table-footer">
      <span>{{ filtered.length }} {{ t('个监控项', 'monitors') }}</span
      ><span>{{ t('每 30 秒自动刷新', 'Refreshes every 30 seconds') }}</span>
    </div>
  </section>
</template>
