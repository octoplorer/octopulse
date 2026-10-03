<script setup lang="ts">
import { useQuery } from '@pinia/colada'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { listAuditQuery } from '../../../client/@pinia/colada.gen'
import AsyncState from '../../../components/AsyncState.vue'
import EmptyState from '../../../components/EmptyState.vue'
import PageHeader from '../../../components/PageHeader.vue'
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
  <PageHeader
    :title="t('navigation.auditLog')"
    :description="t('audit.traceConfigurationChangesAndTheirActorsWithoutRecording')"
  >
    <button class="button" @click="query.refetch()">
      <span class="i-lucide-refresh-cw" un-w="14px" un-h="14px" aria-hidden="true" />{{ t('common.refresh') }}
    </button>
  </PageHeader>
  <section class="card">
    <div class="filter-bar">
      <div class="search-box">
        <span class="i-lucide-search" un-w="16px" un-h="16px" aria-hidden="true" /><input
          v-model="search"
          :placeholder="t('audit.searchActorActionOrResource')"
          :aria-label="t('audit.searchAuditLog')"
        >
      </div>
      <span class="muted" un-text="10px">{{
        t('counts.records', { count: items.length }, items.length)
      }}</span>
    </div>
    <AsyncState :pending="query.isPending.value" :error="query.error.value" @retry="query.refetch()">
      <EmptyState v-if="!items.length" :title="t('audit.noMatchingRecords')" />
      <div v-else class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th>{{ t('audit.time') }}</th>
              <th>{{ t('audit.actor') }}</th>
              <th>{{ t('audit.action') }}</th>
              <th>{{ t('audit.resource') }}</th>
              <th>{{ t('audit.resourceId') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="entry in items" :key="entry.id">
              <td class="muted" un-text="10px">
                {{ formatDate(entry.createdAt) }}
              </td>
              <td>{{ entry.username }}</td>
              <td>
                <span class="pill">{{ entry.action }}</span>
              </td>
              <td>{{ entry.resourceType }}</td>
              <td>
                <code class="muted">{{ entry.resourceId }}</code>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </AsyncState>
  </section>
</template>
