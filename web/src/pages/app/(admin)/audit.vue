<script setup lang="ts">
import { computed, ref } from 'vue'
import { Search, RefreshCw, ScrollText } from '@lucide/vue'
import { useCollection } from '../../../lib/data'
import type { Audit } from '../../../lib/types'
import { t, formatDate } from '../../../lib/preferences'
import PageHeader from '../../../components/PageHeader.vue'
import AsyncState from '../../../components/AsyncState.vue'
import EmptyState from '../../../components/EmptyState.vue'

definePage({ meta: { title: ['审计日志', 'Audit log'], roles: ['admin'] } })

const query = useCollection<Audit>('audit'),
  search = ref(''),
  items = computed(
    () =>
      query.data.value?.items
        .filter((x) =>
          `${x.username} ${x.action} ${x.resourceType} ${x.resourceId}`
            .toLowerCase()
            .includes(search.value.toLowerCase()),
        )
        .sort((a, b) => b.createdAt - a.createdAt) || [],
  )
</script>
<template>
  <PageHeader
    :title="t('审计日志', 'Audit log')"
    :description="
      t(
        '追踪配置变更与操作者，不记录秘密原文。',
        'Trace configuration changes and their actors without recording secret values.',
      )
    "
    ><button class="button" @click="query.refresh()">
      <RefreshCw :size="14" />{{ t('刷新', 'Refresh') }}
    </button></PageHeader
  >
  <section class="card">
    <div class="filter-bar">
      <div class="search-box">
        <Search :size="16" /><input
          v-model="search"
          :placeholder="t('搜索成员、操作或资源…', 'Search actor, action, or resource…')"
          :aria-label="t('搜索日志', 'Search audit log')"
        />
      </div>
      <span class="muted" un-text="10px">{{ items.length }} {{ t('条记录', 'records') }}</span>
    </div>
    <AsyncState :pending="query.isPending.value" :error="query.error.value" @retry="query.refresh()"
      ><EmptyState v-if="!items.length" :title="t('暂无匹配记录', 'No matching records')" />
      <div v-else class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th>{{ t('时间', 'Time') }}</th>
              <th>{{ t('操作者', 'Actor') }}</th>
              <th>{{ t('操作', 'Action') }}</th>
              <th>{{ t('资源类型', 'Resource') }}</th>
              <th>{{ t('资源 ID', 'Resource ID') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="entry in items" :key="entry.id">
              <td class="muted" un-text="10px">{{ formatDate(entry.createdAt) }}</td>
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
      </div></AsyncState
    >
  </section>
</template>
