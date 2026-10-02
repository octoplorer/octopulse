import { useQuery, useQueryCache, type UseQueryOptions } from '@pinia/colada'
import * as generatedQueries from '../client/@pinia/colada.gen'
import { api, currentUser, resolveOperation } from './api'
function optionsFor<T>(path: string): UseQueryOptions<T, Error> {
  const resolved = resolveOperation(path)
  const factory = (
    generatedQueries as unknown as Record<string, (options: unknown) => UseQueryOptions<T, Error>>
  )[`${resolved.operation}Query`]
  if (!factory) throw new Error(`Generated query unavailable: ${resolved.operation}`)
  return factory(resolved)
}
export function useCollection<T>(name: string) {
  return useQuery({
    key: () => ['collection', currentUser.value?.id || 'anonymous', name],
    query: (context) => optionsFor<{ items: T[] }>(name).query(context),
    staleTime: 10000,
  })
}
export function useRecord<T>(name: () => string) {
  return useQuery({
    key: () => ['record', currentUser.value?.id || 'anonymous', name()],
    query: (context) => optionsFor<T>(name()).query(context),
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
