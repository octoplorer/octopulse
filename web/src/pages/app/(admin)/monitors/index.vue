<script setup lang="ts">
import { useQuery } from '@pinia/colada'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { listMonitorsQuery } from '../../../../client/@pinia/colada.gen'
import { Badge } from '../../../../components/ui/badge'
import { Banner } from '../../../../components/ui/banner'
import { PageHeader } from '../../../../components/ui/blocks/page-header'
import { Button } from '../../../../components/ui/button'
import { Empty } from '../../../../components/ui/empty'
import { InputGroup, InputGroupAddon, InputGroupInput } from '../../../../components/ui/input-group'
import { LayerCard } from '../../../../components/ui/layer-card'
import { Loader } from '../../../../components/ui/loader'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '../../../../components/ui/select'
import { Table, TableBody, TableCell, TableContainer, TableHead, TableHeader, TablePagination, TableRow, TableToolbar } from '../../../../components/ui/table'
import { canEdit } from '../../../../composables/api'
import { usePollingEnabled } from '../../../../composables/polling'
import { formatDate } from '../../../../composables/preferences'
import { errorText } from '../../../../lib/errors'
import { monitorTypes, targetOf } from '../../../../lib/monitor'
import { monitorStateDisplay } from '../../../../lib/monitor-state'

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
const hasFilters = computed(() => !!search.value || state.value !== 'all' || type.value !== 'all')
const items = computed(() =>
  [...(query.data.value?.items || [])].sort((a, b) => a.name.localeCompare(b.name, locale.value)),
)
const filtered = computed(() =>
  items.value.filter(
    m =>
      (!search.value
        || `${m.name} ${m.group ?? ''} ${m.tags?.join(' ') ?? ''} ${targetOf(m)}`
          .toLowerCase()
          .includes(search.value.toLowerCase()))
        && (state.value === 'all'
          || (state.value === 'paused' ? !m.enabled : m.enabled && m.state === state.value))
        && (type.value === 'all' || m.type === type.value),
  ).map(monitor => ({
    ...monitor,
    stateDisplay: monitorStateDisplay(
      monitor.type === 'certificate' ? monitor.certificate?.state : monitor.state,
      { paused: !monitor.enabled },
    ),
  })),
)
function clearFilters() {
  search.value = ''
  state.value = 'all'
  type.value = 'all'
}
</script>

<template>
  <PageHeader :title="t('common.monitors')" :description="t('monitors.define-healthy-behavior-and-detect-every-change')" class="mb-6">
    <template #actions>
      <Button :loading="query.isPending.value" @click="query.refetch()">
        <span class="i-lucide-refresh-cw size-4" aria-hidden="true" />{{ t('common.refresh') }}
      </Button>
      <Button v-if="canEdit()" variant="primary" as-child>
        <RouterLink to="/app/monitors/new">
          <span class="i-lucide-plus size-4" aria-hidden="true" />{{ t('common.add-monitor') }}
        </RouterLink>
      </Button>
    </template>
  </PageHeader>
  <LayerCard>
    <TableToolbar class="flex-wrap gap-3">
      <InputGroup class="min-w-0 flex-1 basis-64">
        <InputGroupAddon><span class="i-lucide-search size-4" aria-hidden="true" /></InputGroupAddon>
        <InputGroupInput v-model="search" :placeholder="t('monitors.search-name-target-or-tags')" :aria-label="t('monitors.search-monitors')" />
      </InputGroup>
      <div class="flex max-w-full flex-none flex-wrap items-center gap-3">
        <Select v-model="state">
          <SelectTrigger :aria-label="t('monitors.filter-status')" class="w-auto!">
            <SelectValue />
          </SelectTrigger><SelectContent>
            <SelectGroup>
              <SelectItem value="all">
                {{ t('monitors.all-states') }}
              </SelectItem><SelectItem value="up">
                {{ t('monitors.up') }}
              </SelectItem><SelectItem value="down">
                {{ t('monitors.down') }}
              </SelectItem><SelectItem value="unknown">
                {{ t('monitors.unknown') }}
              </SelectItem><SelectItem value="paused">
                {{ t('common.paused') }}
              </SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
        <Select v-model="type">
          <SelectTrigger :aria-label="t('monitors.filter-type')" class="w-auto!">
            <SelectValue />
          </SelectTrigger><SelectContent>
            <SelectGroup>
              <SelectItem value="all">
                {{ t('monitors.all-types') }}
              </SelectItem><SelectItem v-for="item in monitorTypes" :key="item.value" :value="item.value">
                {{ t(item.label) }}
              </SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
        <Button v-if="hasFilters" variant="ghost" size="sm" @click="clearFilters">
          <span class="i-lucide-x size-3.5" aria-hidden="true" />{{ t('common.clear-filters') }}
        </Button>
      </div>
    </TableToolbar>
    <div v-if="query.isPending.value" class="loading-state" flex="~ justify-center items-center gap-10px" p="60px" un-text="12px subtle" role="status">
      <Loader :label="t('async-state.loading-data')" />{{ t('async-state.loading-data') }}
    </div>
    <Banner v-else-if="query.error.value" variant="error">
      {{ errorText(query.error.value) }}
      <Button variant="ghost" @click="query.refetch()">
        {{ t('async-state.retry') }}
      </Button>
    </Banner>
    <template v-else>
      <Empty v-if="!filtered.length" :title="items.length ? t('monitors.no-matching-monitors') : t('monitors.start-watching-your-services')" :description="items.length ? t('monitors.try-changing-your-search-or-filters') : t('monitors.add-a-service-check-to-see-status-history')" size="sm" class="rounded-none border-none">
        <template #icon>
          <span class="i-lucide-activity size-8 text-subtle" aria-hidden="true" />
        </template>
        <template #actions>
          <Button v-if="hasFilters" @click="clearFilters">
            {{ t('common.clear-filters') }}
          </Button>
          <Button v-if="canEdit() && !items.length" variant="primary" as-child>
            <RouterLink to="/app/monitors/new">
              {{ t('common.add-monitor') }}
            </RouterLink>
          </Button>
        </template>
      </Empty>
      <TableContainer v-else :scroll-label="t('common.scroll-table')">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead class="w-2/5 min-w-64 [@container_workspace_(max-width:_700px)]:w-auto! [@container_workspace_(max-width:_700px)]:min-w-0!">
                {{ t('common.service') }}
              </TableHead><TableHead class="[@container_workspace_(max-width:_700px)]:hidden">
                {{ t('common.status') }}
              </TableHead><TableHead class="text-end [@container_workspace_(max-width:_700px)]:hidden">
                {{ t('monitors.interval-retries') }}
              </TableHead><TableHead class="[@container_workspace_(max-width:_700px)]:hidden">
                {{ t('common.last-check') }}
              </TableHead><TableHead class="w-12">
                <span class="sr-only">{{ t('monitors.view-details') }}</span>
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="monitor in filtered" :key="monitor.id">
              <TableCell>
                <RouterLink :to="`/app/monitors/${monitor.id}`" class="flex items-center gap-3 rounded-md outline-none focus-visible:ring-2 focus-visible:ring-brand">
                  <span class="flex size-8 shrink-0 items-center justify-center rounded-lg bg-recessed text-subtle"><span :class="monitor.type === 'http' || monitor.type === 'dns' ? 'i-lucide-globe' : 'i-lucide-server'" class="size-4" aria-hidden="true" /></span>
                  <span class="min-w-0 space-y-1 [overflow-wrap:anywhere]">
                    <span class="block font-medium">{{ monitor.name }}</span>
                    <span class="block text-size-xs text-subtle">{{ targetOf(monitor) }}</span>
                    <span class="flex flex-wrap items-center gap-x-2 gap-y-1 text-size-xs text-subtle">
                      <span>{{ monitor.type.toUpperCase() }}</span>
                      <span v-if="monitor.group">{{ monitor.group }}</span>
                    </span>
                    <span class="hidden space-y-2 pt-1 [@container_workspace_(max-width:_700px)]:block">
                      <Badge :variant="monitor.stateDisplay.variant" :data-state="monitor.stateDisplay.state" dot>{{ monitor.stateDisplay.label }}</Badge>
                      <span class="block text-size-xs text-subtle">{{ t('monitors.interval-retries') }} · {{ monitor.type === 'heartbeat' ? monitor.heartbeat?.periodSeconds : monitor.intervalSeconds }} s<span v-if="['http', 'tcp', 'dns'].includes(monitor.type)"> / {{ monitor.retries }}</span></span>
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
              <TableCell class="whitespace-nowrap text-end text-size-sm text-subtle [@container_workspace_(max-width:_700px)]:hidden">
                {{ monitor.type === 'heartbeat' ? monitor.heartbeat?.periodSeconds : monitor.intervalSeconds }} s<span v-if="['http', 'tcp', 'dns'].includes(monitor.type)"> / {{ monitor.retries }}</span>
              </TableCell>
              <TableCell class="whitespace-nowrap text-size-sm text-subtle [@container_workspace_(max-width:_700px)]:hidden">
                {{ formatDate(monitor.lastCheckedAt) }}
              </TableCell>
              <TableCell class="text-end">
                <Button variant="ghost" shape="square" size="sm" as-child>
                  <RouterLink :to="`/app/monitors/${monitor.id}`" :aria-label="t('monitors.view-details')">
                    <span class="i-lucide-chevron-right size-4" aria-hidden="true" />
                  </RouterLink>
                </Button>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </TableContainer>
    </template>
    <TablePagination><span role="status" aria-live="polite">{{ t('overview.showing-monitors', { shown: filtered.length, total: items.length }) }}</span><span>{{ t('common.refreshes-every-30-seconds') }}</span></TablePagination>
  </LayerCard>
</template>
