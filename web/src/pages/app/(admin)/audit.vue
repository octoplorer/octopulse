<script setup lang="ts">
import { useQuery } from '@pinia/colada'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { listAuditQuery } from '../../../client/@pinia/colada.gen'
import AsyncState from '../../../components/AsyncState.vue'
import EmptyState from '../../../components/EmptyState.vue'
import PageHeader from '../../../components/PageHeader.vue'
import { Badge } from '../../../components/ui/badge'
import { Button } from '../../../components/ui/button'
import { Card } from '../../../components/ui/card'
import { Table, TableBody, TableCell, TableContainer, TableHead, TableHeader, TableRow, TableToolbar } from '../../../components/ui/table'
import { formatDate } from '../../../composables/preferences'

const { t } = useI18n({ useScope: 'global' })

definePage({ meta: { title: 'navigation.auditLog', roles: ['admin'] } })

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
  <PageHeader :title="t('navigation.auditLog')" :description="t('audit.traceConfigurationChangesAndTheirActorsWithoutRecording')">
    <Button @click="query.refetch()">
      <span w="14px" h="14px" aria-hidden="true" class="i-lucide-refresh-cw" />{{ t('common.refresh') }}
    </Button>
  </PageHeader>
  <Card as="section">
    <TableToolbar>
      <div class="search-box relative [box-shadow:var(--control-shadow)]" flex="~ items-center 1" gap="8px" un-text="$muted" max-w="340px" min-w="180px" border="1 solid $control-border" rounded="8px" pl="10px" bg="$surface">
        <span w="16px" h="16px" aria-hidden="true" class="i-lucide-search" /><input v-model="search" :placeholder="t('audit.searchActorActionOrResource')" :aria-label="t('audit.searchAuditLog')">
      </div>
      <span class="muted" un-text="13px $muted">{{
        t('counts.records', { count: items.length }, items.length)
      }}</span>
    </TableToolbar>
    <AsyncState :pending="query.isPending.value" :error="query.error.value" @retry="query.refetch()">
      <EmptyState v-if="!items.length" :title="t('audit.noMatchingRecords')" />
      <TableContainer v-else>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{{ t('audit.time') }}</TableHead>
              <TableHead>{{ t('audit.actor') }}</TableHead>
              <TableHead>{{ t('audit.action') }}</TableHead>
              <TableHead>{{ t('audit.resource') }}</TableHead>
              <TableHead>{{ t('audit.resourceId') }}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="entry in items" :key="entry.id">
              <TableCell class="muted" un-text="13px $muted">
                {{ formatDate(entry.createdAt) }}
              </TableCell>
              <TableCell>{{ entry.username }}</TableCell>
              <TableCell>
                <Badge>{{ entry.action }}</Badge>
              </TableCell>
              <TableCell>{{ entry.resourceType }}</TableCell>
              <TableCell>
                <code class="muted" un-text="13px $muted">{{ entry.resourceId }}</code>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </TableContainer>
    </AsyncState>
  </Card>
</template>
