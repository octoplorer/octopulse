import { useQuery, useQueryCache, type UseQueryOptions } from '@pinia/colada'
import { currentUser } from './api'
import type { ErrorModel } from '../client/types.gen'
type QueryOptions = Pick<UseQueryOptions<unknown, ErrorModel>, 'query'>
export function useCollection<T>(name: string, options: QueryOptions) {
  return useQuery<{ items: T[] }, ErrorModel>({
    key: () => ['collection', currentUser.value?.id || 'anonymous', name],
    query: async (context) => (await options.query(context)) as { items: T[] },
    staleTime: 10000,
  })
}
export function useRecord<T>(name: () => string, options: () => QueryOptions) {
  return useQuery<T, ErrorModel>({
    key: () => ['record', currentUser.value?.id || 'anonymous', name()],
    query: async (context) => (await options().query(context)) as T,
    staleTime: 5000,
  })
}
export function useInvalidate() {
  const cache = useQueryCache()
  return async (name?: string) => {
    if (name)
      await cache.invalidateQueries({
        key: ['collection', currentUser.value?.id || 'anonymous', name],
      })
    await cache.invalidateQueries({ key: ['record'] })
  }
}
