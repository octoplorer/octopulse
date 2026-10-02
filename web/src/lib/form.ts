import { t } from '../composables/i18n'

export function clone<T>(value: T): T {
  return JSON.parse(JSON.stringify(value)) as T
}
export function parseJSON<T>(value: string, label: string): T {
  try {
    return JSON.parse(value) as T
  }
  catch {
    throw new Error(t('errors.invalidJSON', { label }))
  }
}
export function splitValues(value: string): string[] {
  return value
    .split(/[\n,]/)
    .map(x => x.trim())
    .filter(Boolean)
}
export function defaults<T>(base: T, value: unknown): T {
  if (value === null || value === undefined)
    return clone(base)
  if (Array.isArray(base))
    return (Array.isArray(value) ? value : base) as T
  if (typeof base === 'object' && base !== null && typeof value === 'object') {
    const output = { ...value } as Record<string, unknown>
    for (const [key, item] of Object.entries(base))
      output[key] = defaults(item, (value as Record<string, unknown>)[key])
    return output as T
  }
  return value as T
}
