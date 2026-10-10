import type { MaybeRefOrGetter } from 'vue'
import { computed, ref, toRaw, toValue, watch } from 'vue'
import { errorText } from '../lib/errors'
import { parseJSON, splitValues } from '../lib/form'

export function useListInput<T>(
  value: MaybeRefOrGetter<T[]>,
  update: (value: T[]) => void,
  options: {
    parseItem: (value: string) => T
    formatItem?: (value: T) => string
    separator?: 'values' | 'lines'
    trim?: MaybeRefOrGetter<boolean>
  },
) {
  function format(items: T[]) {
    return items.map(options.formatItem || String).join(options.separator === 'lines' ? '\n' : ', ')
  }

  const draft = ref(format(toValue(value)))
  let emittedValue: T[] | undefined

  watch(() => toValue(value), (next) => {
    if (toRaw(next) === emittedValue) {
      emittedValue = undefined
      return
    }
    draft.value = format(next)
  })

  return computed({
    get: () => draft.value,
    set: (text: string) => {
      draft.value = text
      const items = options.separator === 'lines'
        ? text.split('\n').map(item => toValue(options.trim) ? item.trim() : item).filter(Boolean)
        : splitValues(text)
      emittedValue = items.map(options.parseItem)
      update(emittedValue)
    },
  })
}

export function useJSONInput<T>(
  value: MaybeRefOrGetter<T>,
  update: (value: T) => void,
  label: MaybeRefOrGetter<string>,
) {
  const draft = ref(JSON.stringify(toValue(value), null, 2) || '')
  const error = ref('')
  let emittedValue: T | undefined

  watch(() => toValue(value), (next) => {
    if (toRaw(next) === emittedValue) {
      emittedValue = undefined
      return
    }
    draft.value = JSON.stringify(next, null, 2) || ''
    error.value = ''
  })

  const text = computed({
    get: () => draft.value,
    set: (text: string) => {
      draft.value = text
      try {
        const next = parseJSON<T>(text, toValue(label))
        emittedValue = toRaw(next)
        error.value = ''
        update(next)
      }
      catch (cause) {
        error.value = errorText(cause)
      }
    },
  })

  return { text, error }
}
