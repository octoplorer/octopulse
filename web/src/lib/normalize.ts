const collectionKeys = new Set([
  'items',
  'tags',
  'notificationChannelIds',
  'query',
  'headers',
  'fields',
  'files',
  'statusCodes',
  'statusRanges',
  'textContains',
  'textNotContains',
  'regex',
  'json',
  'warningDays',
  'expectedValues',
  'updates',
  'monitorIds',
  'pageIds',
  'allowedDomains',
  'groups',
  'links',
  'monitors',
  'incidents',
  'maintenance',
  'latency',
  'rounds',
  'attempts',
])
// Go nil slices are valid null arrays in the transport contract. UI list values
// use empty arrays while meaningful null statistics (uptime/coverage) stay null.
export function normalizeCollections(value: unknown): unknown {
  if (Array.isArray(value)) return value.map(normalizeCollections)
  if (typeof value !== 'object' || value === null) return value
  return Object.fromEntries(
    Object.entries(value).map(([key, item]) => [
      key,
      item === null && collectionKeys.has(key) ? [] : normalizeCollections(item),
    ]),
  )
}
