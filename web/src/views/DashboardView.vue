<script setup lang="ts">
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
import { useCollection } from '../lib/data'
import type { Monitor, Incident, Page, Maintenance } from '../lib/types'
import { targetOf } from '../lib/monitor'
import { publishedEntry } from '../lib/pages'
import { t, formatDate, statusLabel } from '../lib/preferences'
import { canEdit } from '../lib/api'
import { useIntervalFn } from '@vueuse/core'
import PageHeader from '../components/PageHeader.vue'
import StateBadge from '../components/StateBadge.vue'
import AsyncState from '../components/AsyncState.vue'
import EmptyState from '../components/EmptyState.vue'
const monitors = useCollection<Monitor>('monitors'),
  incidents = useCollection<Incident>('incidents'),
  pages = useCollection<Page>('pages'),
  maintenance = useCollection<Maintenance>('maintenance')
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
    :title="t('服务概览', 'Service overview')"
    :description="t('每个服务的心跳，都在这里。', 'A clear view of every service heartbeat.')"
    eyebrow="YOUR INFRASTRUCTURE, AT A GLANCE"
    ><button class="button" @click="monitors.refresh()">
      <RefreshCw :size="14" />{{ t('刷新', 'Refresh') }}</button
    ><RouterLink v-if="canEdit()" to="/app/monitors/new" class="button primary"
      ><Plus :size="15" />{{ t('添加监控项', 'Add monitor') }}</RouterLink
    ></PageHeader
  >
  <div class="stats-grid">
    <div class="card stat-card">
      <div class="stat-label">
        {{ t('监控项总数', 'Total monitors')
        }}<span class="stat-icon"><Activity :size="15" /></span>
      </div>
      <div class="stat-value">{{ items.length }}</div>
      <div class="stat-meta">
        {{ t('正在监控', 'Active') }} {{ active.length }} · {{ t('暂停', 'Paused') }}
        {{ items.length - active.length }}
      </div>
    </div>
    <div class="card stat-card">
      <div class="stat-label">
        {{ t('正常运行', 'Operational') }}<span class="stat-icon"><CheckCircle2 :size="15" /></span>
      </div>
      <div class="stat-value" un-text="[var(--accent)]">{{ up.length }}</div>
      <div class="stat-meta">
        <span class="positive">{{ t('已确认的服务状态', 'Confirmed service states') }}</span>
      </div>
    </div>
    <div class="card stat-card">
      <div class="stat-label">
        {{ t('需要关注', 'Needs attention')
        }}<span class="stat-icon"><AlertTriangle :size="15" /></span>
      </div>
      <div class="stat-value" :style="{ color: down.length ? 'var(--danger)' : undefined }">
        {{ down.length }}
      </div>
      <div class="stat-meta">{{ t('已确认故障', 'Confirmed outages') }}</div>
    </div>
    <div class="card stat-card">
      <div class="stat-label">
        {{ t('等待数据', 'Waiting for data') }}<span class="stat-icon"><Clock :size="15" /></span>
      </div>
      <div class="stat-value">{{ unknown.length }}</div>
      <div class="stat-meta">
        {{ t('首次检查或采集断档', 'Initial checks or collection gaps') }}
      </div>
    </div>
  </div>
  <div class="dashboard-grid">
    <div>
      <section class="card">
        <div class="card-header">
          <div>
            <h2>{{ t('监控项', 'Monitors') }}</h2>
            <p class="muted">
              {{
                t(
                  '故障优先，快速定位关键服务。',
                  'Outages first, so you can focus on what matters.',
                )
              }}
            </p>
          </div>
          <RouterLink to="/app/monitors" class="button small ghost"
            >{{ t('查看全部', 'View all') }}<ArrowUpRight :size="13"
          /></RouterLink>
        </div>
        <AsyncState
          :pending="monitors.isPending.value"
          :error="monitors.error.value"
          @retry="monitors.refresh()"
          ><EmptyState
            v-if="!items.length"
            :title="t('为第一个服务建立监控', 'Monitor your first service')"
            :description="
              t(
                '支持 HTTP、TCP、DNS、心跳与证书到期检查。',
                'HTTP, TCP, DNS, heartbeat, and certificate checks are ready.',
              )
            "
            ><RouterLink v-if="canEdit()" to="/app/monitors/new" class="button primary"
              ><Plus :size="14" />{{ t('创建监控项', 'Create monitor') }}</RouterLink
            ></EmptyState
          >
          <div v-else class="table-wrap">
            <table class="data-table">
              <thead>
                <tr>
                  <th>{{ t('服务', 'Service') }}</th>
                  <th>{{ t('状态', 'Status') }}</th>
                  <th>{{ t('检查间隔', 'Interval') }}</th>
                  <th>{{ t('最近检查', 'Last check') }}</th>
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
          <span>{{ t('展示', 'Showing') }} {{ ordered.length }} / {{ items.length }}</span
          ><span>{{ t('每 30 秒自动刷新', 'Refreshes every 30 seconds') }}</span>
        </div>
      </section>
      <section class="card" un-mt="6">
        <div class="card-header">
          <h2>{{ t('公开状态页', 'Public status pages') }}</h2>
          <RouterLink to="/app/pages" class="button small ghost"
            >{{ t('管理页面', 'Manage pages') }}<ArrowUpRight :size="13"
          /></RouterLink>
        </div>
        <div class="card-body">
          <EmptyState
            v-if="!pages.data.value?.items.length"
            :title="t('让团队与用户了解服务状态', 'Keep everyone informed')"
            :description="
              t(
                '创建可自定义品牌、分组与公告的公开状态页。',
                'Publish a status page with your own brand, groups, and incident updates.',
              )
            "
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
              page.publishedAt ? t('已发布', 'Published') : t('草稿', 'Draft')
            }}</span>
          </div>
        </div>
      </section>
    </div>
    <aside>
      <section class="card">
        <div class="card-header">
          <h2>{{ t('事件动态', 'Incident activity') }}</h2>
          <span class="pill">{{ recentIncidents.length }}</span>
        </div>
        <div class="card-body">
          <p v-if="!recentIncidents.length" class="muted" un-py="4">
            {{
              t(
                '没有事件公告。服务的后续进展会显示在这里。',
                'No incident announcements. Updates will appear here.',
              )
            }}
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
          <h2>{{ t('近期维护', 'Upcoming maintenance') }}</h2>
        </div>
        <div class="card-body">
          <p v-if="!nextMaintenance.length" class="muted" un-py="4">
            {{ t('暂无计划维护。', 'No scheduled maintenance.') }}
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
        {{
          t(
            '状态统计基于确认后的有效时长，暂停、维护与缺失观测不会被算作正常运行。',
            'Uptime reflects confirmed duration. Paused, maintenance, and missing observations never count as operational time.',
          )
        }}
      </div>
    </aside>
  </div>
</template>
