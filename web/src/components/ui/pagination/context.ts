import type { ComputedRef, InjectionKey, Ref } from 'vue'
import { inject } from 'vue'

export interface PaginationLabels {
  navigation?: string
  firstPage?: string
  previousPage?: string
  nextPage?: string
  lastPage?: string
  pageNumber?: string
  pageSize?: string
  perPage?: string
}
export interface PaginationContext {
  page: Ref<number>
  pageSize: Ref<number>
  count: ComputedRef<number | undefined>
  maxPage: ComputedRef<number>
  hasNextPage: ComputedRef<boolean>
  labels: ComputedRef<Required<PaginationLabels>>
  setPage: (value: number) => void
  setPageSize: (value: number) => void
}
export const paginationKey: InjectionKey<PaginationContext> = Symbol('pagination')
export function usePagination() {
  const context = inject(paginationKey)
  if (!context)
    throw new Error('Pagination components must be used inside Pagination')
  return context
}
