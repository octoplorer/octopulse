<script setup lang="ts">
import { useQuery } from '@pinia/colada'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { listMonitorsQuery } from '../../../../client/@pinia/colada.gen'
import AsyncState from '../../../../components/AsyncState.vue'
import EmptyState from '../../../../components/EmptyState.vue'
import PageHeader from '../../../../components/PageHeader.vue'
import StateBadge from '../../../../components/StateBadge.vue'
import { Button } from '../../../../components/ui/button'
import { Card } from '../../../../components/ui/card'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '../../../../components/ui/select'
import { Table, TableBody, TableCell, TableContainer, TableFooter, TableHead, TableHeader, TableRow, TableToolbar } from '../../../../components/ui/table'
import { canEdit } from '../../../../composables/api'
import { usePollingEnabled } from '../../../../composables/polling'
import { formatDate } from '../../../../composables/preferences'
import { monitorTypes, targetOf } from '../../../../lib/monitor'

const { t, locale } = useI18n({ useScope: 'global' })

definePage({ meta: { title: 'common.monitors' } })

const pollingEnabled = usePollingEnabled()
const query = useQuery(
  () =>
    ({
      ...listMonitorsQuery(),
      staleTime: 10000,
      enabled: pollingEnabled.value,
      autoRefetch: 30000,
    }),
)
const search = ref('')
const state = ref('all')
const type = ref('all')
const items = computed(() =>
  [...(query.data.value?.items || [])].sort((a, b) => a.name.localeCompare(b.name, locale.value)),
)
const filtered = computed(() =>
  items.value.filter(
    m =>
      (!search.value
        || `${m.name} ${m.group} ${m.tags?.join(' ')} ${targetOf(m)}`
          .toLowerCase()
          .includes(search.value.toLowerCase()))
        && (state.value === 'all'
          || (state.value === 'paused' ? !m.enabled : m.enabled && m.state === state.value))
        && (type.value === 'all' || m.type === type.value),
  ),
)
</script>

<template>
  <PageHeader :title="t('common.monitors')" :description="t('monitors.defineHealthyBehaviorAndDetectEveryChange')">
    <Button @click="query.refetch()">
      <span w="14px" h="14px" aria-hidden="true" class="i-lucide-refresh-cw" />{{ t('common.refresh') }}
    </Button><Button v-if="canEdit()" variant="primary" as-child>
      <RouterLink to="/app/monitors/new">
        <span w="15px" h="15px" aria-hidden="true" class="i-lucide-plus" />{{ t('common.addMonitor') }}
      </RouterLink>
    </Button>
  </PageHeader>
  <Card as="section">
    <TableToolbar>
      <div class="search-box relative [@media(max-width:700px)]:basis-full [@media(max-width:700px)]:max-w-none [@container_workspace_(max-width:_700px)]:basis-full [@container_workspace_(max-width:_700px)]:max-w-none" flex="~ items-center 1" gap="8px" un-text="subtle" max-w="340px" min-w="180px" border="1 solid line" rounded="8px" pl="10px" bg="base" shadow="control">
        <span w="16px" h="16px" aria-hidden="true" class="i-lucide-search" /><input v-model="search" w="full" min-w="0" border="0!" bg="transparent!" shadow="none!" p="y-7px! r-10px! l-0!" :placeholder="t('monitors.searchNameTargetOrTags')" :aria-label="t('monitors.searchMonitors')">
      </div>
      <div flex="~ items-center gap-2">
        <Select v-model="state">
          <SelectTrigger w="auto!" :aria-label="t('monitors.filterStatus')">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value="all">
                {{ t('monitors.allStates') }}
              </SelectItem>
              <SelectItem value="up">
                {{ t('monitors.up') }}
              </SelectItem>
              <SelectItem value="down">
                {{ t('monitors.down') }}
              </SelectItem>
              <SelectItem value="unknown">
                {{ t('monitors.unknown') }}
              </SelectItem>
              <SelectItem value="paused">
                {{ t('common.paused') }}
              </SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select><Select v-model="type">
          <SelectTrigger w="auto!" :aria-label="t('monitors.filterType')">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value="all">
                {{ t('monitors.allTypes') }}
              </SelectItem>
              <SelectItem v-for="item in monitorTypes" :key="item.value" :value="item.value">
                {{ t(item.label) }}
              </SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
      </div>
    </TableToolbar>
    <AsyncState :pending="query.isPending.value" :error="query.error.value" @retry="query.refetch()">
      <EmptyState
        v-if="!filtered.length" :title="
          items.length ? t('monitors.noMatchingMonitors') : t('monitors.startWatchingYourServices')
        " :description="
          items.length
            ? t('monitors.tryChangingYourSearchOrFilters')
            : t('monitors.addAServiceCheckToSeeStatusHistory')
        "
      >
        <Button v-if="!items.length && canEdit()" variant="primary" as-child>
          <RouterLink to="/app/monitors/new">
            {{ t('common.createMonitor') }}
          </RouterLink>
        </Button>
      </EmptyState>
      <TableContainer v-else>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{{ t('common.service') }}</TableHead>
              <TableHead>{{ t('common.status') }}</TableHead>
              <TableHead>{{ t('common.group') }}</TableHead>
              <TableHead>{{ t('monitors.intervalRetries') }}</TableHead>
              <TableHead>{{ t('common.lastCheck') }}</TableHead>
              <TableHead />
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="monitor in filtered" :key="monitor.id">
              <TableCell>
                <RouterLink :to="`/app/monitors/${monitor.id}`" flex="~ items-center gap-3">
                  <span class="monitor-type-icon" flex="~ items-center justify-center shrink-0" size="32px" border="1 solid line" rounded="8px" un-text="subtle" bg="base"><span v-if="monitor.type === 'http' || monitor.type === 'dns'" w="16px" h="16px" aria-hidden="true" class="i-lucide-globe" /><span v-else w="16px" h="16px" aria-hidden="true" class="i-lucide-server" /></span><span><span class="monitor-name block" font="600" un-text="13px">{{ monitor.name }}</span><span class="monitor-sub block [overflow-wrap:anywhere]" un-text="12px subtle" mt="3px" max-w="300px">{{ targetOf(monitor) }}</span></span>
                </RouterLink>
              </TableCell>
              <TableCell>
                <StateBadge
                  :state="
                    monitor.type === 'certificate' ? monitor.certificate?.state : monitor.state
                  " :paused="!monitor.enabled"
                />
              </TableCell>
              <TableCell class="muted" un-text="13px subtle">
                {{ monitor.group || '—' }}
              </TableCell>
              <TableCell>
                {{
                  monitor.type === 'heartbeat'
                    ? monitor.heartbeat?.periodSeconds
                    : monitor.intervalSeconds
                }}
                s
                <span v-if="['http', 'tcp', 'dns'].includes(monitor.type)" class="muted" un-text="13px subtle">/ {{ monitor.retries }}</span>
              </TableCell>
              <TableCell class="muted" un-text="13px subtle">
                {{ formatDate(monitor.lastCheckedAt) }}
              </TableCell>
              <TableCell>
                <Button shape="square" as-child>
                  <RouterLink :to="`/app/monitors/${monitor.id}`" :aria-label="t('monitors.viewDetails')">
                    <span w="16px" h="16px" aria-hidden="true" class="i-lucide-chevron-right" />
                  </RouterLink>
                </Button>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </TableContainer>
    </AsyncState>
    <TableFooter>
      <span>{{ t('counts.monitors', { count: filtered.length }, filtered.length) }}</span><span>{{ t('common.refreshesEvery30Seconds') }}</span>
    </TableFooter>
  </Card>
</template>
