<script setup lang="ts">
import { Pagination as ArkPagination } from '@ark-ui/vue/pagination'
import { computed, shallowRef, watch } from 'vue'
import { Button } from '../button'
import { usePagination } from './context'

const props = withDefaults(defineProps<{ controls?: 'full' | 'simple', pageSelector?: 'input' | 'dropdown' }>(), { controls: 'full', pageSelector: 'input' })
const context = usePagination()
const editingPage = shallowRef(String(context.page.value))
const showFull = computed(() => props.controls === 'full' && context.count.value != null)
const pageOptions = computed(() => Array.from({ length: context.maxPage.value }, (_, index) => index + 1))
watch(context.page, (value) => {
  editingPage.value = String(value)
})
function commitPage() {
  context.setPage(Number(editingPage.value))
  editingPage.value = String(context.page.value)
}
</script>

<template>
  <div class="ml-auto flex items-center gap-1">
    <Button v-if="showFull" variant="secondary" shape="square" size="sm" :aria-label="context.labels.value.firstPage" :disabled="context.page.value <= 1" @click="context.setPage(1)">
      <span class="i-lucide-chevrons-left size-4" aria-hidden="true" />
    </Button>
    <ArkPagination.PrevTrigger as-child>
      <Button variant="secondary" shape="square" size="sm" :aria-label="context.labels.value.previousPage">
        <span class="i-lucide-chevron-left size-4" aria-hidden="true" />
      </Button>
    </ArkPagination.PrevTrigger>
    <template v-if="showFull">
      <select v-if="props.pageSelector === 'dropdown'" :value="context.page.value" :aria-label="context.labels.value.pageNumber" class="h-7 rounded-md bg-base px-2 text-default outline-none ring-1 ring-line focus:ring-focus/50 focus-visible:ring-2 focus-visible:ring-brand" @change="context.setPage(Number(($event.target as HTMLSelectElement).value))">
        <option v-for="value in pageOptions" :key="value" :value="value">
          {{ value }}
        </option>
      </select>
      <input v-else v-model="editingPage" type="text" inputmode="numeric" autocomplete="off" :aria-label="context.labels.value.pageNumber" class="h-7 w-12 rounded-md bg-base px-1 text-center text-default tabular-nums outline-none ring-1 ring-line focus:ring-[1.5px] focus:ring-focus/50" @blur="commitPage" @keydown.enter.prevent="commitPage">
      <span class="px-1 text-subtle">/ {{ context.maxPage.value }}</span>
    </template>
    <ArkPagination.NextTrigger as-child>
      <Button variant="secondary" shape="square" size="sm" :aria-label="context.labels.value.nextPage" :disabled="!context.hasNextPage.value">
        <span class="i-lucide-chevron-right size-4" aria-hidden="true" />
      </Button>
    </ArkPagination.NextTrigger>
    <Button v-if="showFull" variant="secondary" shape="square" size="sm" :aria-label="context.labels.value.lastPage" :disabled="!context.hasNextPage.value" @click="context.setPage(context.maxPage.value)">
      <span class="i-lucide-chevrons-right size-4" aria-hidden="true" />
    </Button>
  </div>
</template>
