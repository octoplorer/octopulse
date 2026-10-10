<script setup lang="ts">
import { useQuery } from '@pinia/colada'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { listAuditQuery } from '../../../client/@pinia/colada.gen'
import { Badge } from '../../../components/ui/badge'
import { Banner } from '../../../components/ui/banner'
import { PageHeader } from '../../../components/ui/blocks/page-header'
import { Button } from '../../../components/ui/button'
import { ClipboardText } from '../../../components/ui/clipboard-text'
import { Empty } from '../../../components/ui/empty'
import { InputGroup, InputGroupAddon, InputGroupInput } from '../../../components/ui/input-group'
import { LayerCard, LayerCardPrimary } from '../../../components/ui/layer-card'
import { Loader } from '../../../components/ui/loader'
import { Table, TableBody, TableCell, TableContainer, TableHead, TableHeader, TableRow, TableToolbar } from '../../../components/ui/table'
import { formatDate } from '../../../composables/preferences'
import { errorText } from '../../../lib/errors'

const { t } = useI18n({ useScope: 'global' })

definePage({ meta: { title: 'navigation.audit-log', roles: ['admin'] } })

const query = useQuery({
  ...listAuditQuery(),
  staleTime: 10000,
})
const search = ref('')
const items = computed(
  () =>
    query.data.value?.items
      .filter(x =>
        `${x.username} ${x.action} ${x.resourceType} ${x.resourceId}`
          .toLowerCase()
          .includes(search.value.toLowerCase()),
      )
      .sort((a, b) => b.createdAt - a.createdAt) || [],
)
</script>

<template>
  <PageHeader class="mb-6" :title="t('navigation.audit-log')" :description="t('audit.trace-configuration-changes-and-their-actors-without-recording')">
    <template #actions>
      <Button :loading="query.isLoading.value" @click="query.refetch()">
        <span w="14px" h="14px" aria-hidden="true" class="i-lucide-refresh-cw" />{{ t('common.refresh') }}
      </Button>
    </template>
  </PageHeader>
  <LayerCard>
    <LayerCardPrimary class="p-0!">
      <TableToolbar>
        <InputGroup class="w-full max-w-sm">
          <InputGroupAddon><span class="i-lucide-search size-4" aria-hidden="true" /></InputGroupAddon>
          <InputGroupInput v-model="search" type="search" :placeholder="t('audit.search-actor-action-or-resource')" :aria-label="t('audit.search-audit-log')" />
        </InputGroup>
        <span class="muted" un-text="13px subtle">{{
          t('counts.records', { count: items.length }, items.length)
        }}</span>
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
        <Empty v-if="!items.length" size="sm" class="rounded-none border-none" :title="search.trim() ? t('audit.no-matching-records') : t('audit.no-records')" :description="search.trim() ? t('monitors.try-changing-your-search-or-filters') : undefined">
          <template #actions>
            <Button v-if="search.trim()" @click="search = ''">
              {{ t('common.clear-filters') }}
            </Button>
          </template>
        </Empty>
        <TableContainer v-else :scroll-label="t('common.scroll-table')">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{{ t('audit.actor') }}</TableHead>
                <TableHead class="[@container_workspace_(max-width:_700px)]:hidden">
                  {{ t('audit.action') }}
                </TableHead>
                <TableHead>{{ t('audit.resource') }}</TableHead>
                <TableHead class="[@container_workspace_(max-width:_700px)]:hidden">
                  {{ t('audit.time') }}
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="entry in items" :key="entry.id">
                <TableCell>
                  <span class="block font-medium [overflow-wrap:anywhere]">{{ entry.username }}</span>
                  <span class="mt-1 block text-size-xs text-subtle [@container_workspace_(width_>_700px)]:hidden">{{ formatDate(entry.createdAt) }}</span>
                  <Badge class="mt-2 [@container_workspace_(width_>_700px)]:hidden">
                    {{ entry.action }}
                  </Badge>
                </TableCell>
                <TableCell class="[@container_workspace_(max-width:_700px)]:hidden">
                  <Badge>{{ entry.action }}</Badge>
                </TableCell>
                <TableCell>
                  <span class="block [overflow-wrap:anywhere]">{{ entry.resourceType }}</span>
                  <ClipboardText v-if="entry.resourceId" :text="entry.resourceId" inline :copy-label="t('monitor-details.copy')" :copied-label="t('monitor-details.copied')" class="mt-1 max-w-64 text-subtle" />
                </TableCell>
                <TableCell class="whitespace-nowrap text-subtle [@container_workspace_(max-width:_700px)]:hidden">
                  {{ formatDate(entry.createdAt) }}
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </TableContainer>
      </template>
    </LayerCardPrimary>
  </LayerCard>
</template>
