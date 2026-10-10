<script setup lang="ts">
import type { PaginationLabels } from './context'
import { Pagination as ArkPagination } from '@ark-ui/vue/pagination'
import { computed, provide, watch } from 'vue'
import { paginationKey } from './context'
import PaginationControls from './PaginationControls.vue'
import PaginationInfo from './PaginationInfo.vue'

const props = withDefaults(defineProps<{ count?: number, hasNextPage?: boolean, controls?: 'full' | 'simple', pageSelector?: 'input' | 'dropdown', labels?: PaginationLabels, showInfo?: boolean }>(), { controls: 'full', pageSelector: 'input', showInfo: true })
const page = defineModel<number>('page', { default: 1 })
const pageSize = defineModel<number>('pageSize', { default: 25 })
const count = computed(() => props.count)
const maxPage = computed(() => props.count == null ? Math.max(1, page.value + (props.hasNextPage ? 1 : 0)) : Math.max(1, Math.ceil(Math.max(0, props.count) / Math.max(1, pageSize.value))))
const syntheticCount = computed(() => props.count ?? maxPage.value * Math.max(1, pageSize.value))
const hasNextPage = computed(() => props.count == null ? props.hasNextPage === true : page.value < maxPage.value)
const labels = computed(() => ({ navigation: 'Pagination', firstPage: 'First page', previousPage: 'Previous page', nextPage: 'Next page', lastPage: 'Last page', pageNumber: 'Page number', pageSize: 'Page size', perPage: 'Per page:', ...props.labels }))
function setPage(value: number) {
  page.value = Math.min(maxPage.value, Math.max(1, Number.isFinite(value) ? Math.trunc(value) : page.value))
}
function setPageSize(value: number) {
  pageSize.value = Math.max(1, Math.trunc(value))
  page.value = 1
}
watch(maxPage, () => setPage(page.value))
provide(paginationKey, { page, pageSize, count, maxPage, hasNextPage, labels, setPage, setPageSize })
</script>

<template>
  <ArkPagination.Root :page="page" :page-size="pageSize" :count="syntheticCount" class="flex flex-wrap items-center justify-between gap-3 text-size-sm" :aria-label="labels.navigation" @update:page="setPage" @update:page-size="setPageSize">
    <slot :page="page" :page-size="pageSize" :max-page="maxPage">
      <PaginationInfo v-if="props.showInfo" />
      <PaginationControls :controls="props.controls" :page-selector="props.pageSelector" />
    </slot>
  </ArkPagination.Root>
</template>
