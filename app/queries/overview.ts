export function useOverview() {
  return useQuery({
    key: ['overview'],
    query: () => $fetch('/api/services/overviews'),
  })
}
