<script setup lang="ts">
import type { ComboboxValueChangeDetails } from '@ark-ui/vue/combobox'
import type { CommandPaletteItemData } from './types'
import { Combobox as ArkCombobox, createListCollection } from '@ark-ui/vue/combobox'
import { computed, shallowRef, useId, watch } from 'vue'
import CommandPaletteFooter from './CommandPaletteFooter.vue'

const props = withDefaults(defineProps<{ items: CommandPaletteItemData[], label?: string, placeholder?: string, emptyLabel?: string, loading?: boolean, showFooter?: boolean, inputId?: string }>(), { label: 'Search commands', placeholder: 'Search…', emptyLabel: 'No results found', showFooter: true })
const emit = defineEmits<{ select: [item: CommandPaletteItemData, options: { newTab: boolean }] }>()
const generatedInputId = useId()
const query = defineModel<string>('query', { default: '' })
const filteredItems = computed(() => {
  const needle = query.value.trim().toLocaleLowerCase()
  return props.items.filter(item => !needle || [item.label, item.description ?? '', ...(item.keywords ?? [])].join(' ').toLocaleLowerCase().includes(needle))
})
const collection = computed(() => createListCollection({ items: filteredItems.value, itemToString: item => item.label, itemToValue: item => item.value, isItemDisabled: item => item.disabled === true }))
const highlightedValue = shallowRef<string | null>(null)
watch(filteredItems, (items) => {
  highlightedValue.value = items.find(item => !item.disabled)?.value ?? null
}, { immediate: true })
let newTab = false
function select(details: ComboboxValueChangeDetails<CommandPaletteItemData>) {
  const item = details.items[0]
  if (item && !item.disabled)
    emit('select', item, { newTab })
  newTab = false
}
function recordKey(event: KeyboardEvent) {
  newTab = event.key === 'Enter' && (event.metaKey || event.ctrlKey)
  if (!newTab)
    return
  const item = filteredItems.value.find(item => item.value === highlightedValue.value)
  if (!item || item.disabled)
    return
  event.preventDefault()
  event.stopPropagation()
  emit('select', item, { newTab: true })
  newTab = false
}
function recordClick(event: MouseEvent) {
  newTab = event.metaKey || event.ctrlKey
  if (!newTab)
    return
  const value = (event.target as HTMLElement).closest<HTMLElement>('[role="option"]')?.dataset.value
  const item = filteredItems.value.find(item => item.value === value)
  if (!item || item.disabled)
    return
  event.preventDefault()
  event.stopPropagation()
  emit('select', item, { newTab: true })
  newTab = false
}
</script>

<template>
  <!-- This inline list shares the dialog's layer instead of registering a nested popup. -->
  <ArkCombobox.Root v-model:input-value="query" :model-value="[]" :highlighted-value="highlightedValue ?? undefined" :collection="collection" :open="true" disable-layer :close-on-select="false" :open-on-click="true" selection-behavior="preserve" :loop-focus="true" class="flex min-h-0 flex-col overflow-hidden rounded-xl bg-base text-default" @update:highlighted-value="highlightedValue = $event" @value-change="select" @keydown.capture="recordKey" @click.capture="recordClick">
    <slot :items="filteredItems" :query="query">
      <ArkCombobox.Control class="flex shrink-0 items-center gap-3 border-b border-line bg-base px-4 py-3 focus-within:ring-2 focus-within:ring-brand">
        <span class="i-lucide-search size-5 text-subtle" aria-hidden="true" /><ArkCombobox.Input :id="props.inputId ?? generatedInputId" data-command-palette-input :aria-label="props.label" :placeholder="props.placeholder" class="h-8 min-w-0 flex-1 border-none bg-transparent text-size-base text-default outline-none placeholder:text-placeholder" /><slot name="input-end" />
      </ArkCombobox.Control>
      <ArkCombobox.Content class="max-h-80 min-h-0 overflow-y-auto p-2 outline-none" :aria-busy="props.loading || undefined">
        <div v-if="props.loading" role="status" class="flex items-center justify-center gap-2 py-8 text-subtle">
          <span class="i-lucide-loader-circle size-4 animate-spin" aria-hidden="true" /><slot name="loading">
            Loading…
          </slot>
        </div>
        <template v-else>
          <div v-if="filteredItems.length === 0" role="status" class="px-3 py-8 text-center text-size-sm text-subtle">
            <slot name="empty">
              {{ props.emptyLabel }}
            </slot>
          </div>
          <template v-for="(item, index) in filteredItems" :key="item.value">
            <div v-if="item.group && item.group !== filteredItems[index - 1]?.group" class="px-3 pb-1 pt-3 text-size-xs font-semibold text-subtle">
              {{ item.group }}
            </div>
            <ArkCombobox.Item :item="item" persist-focus class="group flex w-full items-center gap-3 rounded-lg px-3 py-2 text-size-sm text-left outline-none transition-colors" :class="item.disabled ? 'cursor-default opacity-50' : 'cursor-pointer hover:bg-overlay data-[highlighted]:bg-overlay'">
              <slot name="item" :item="item">
                <div class="min-w-0 flex-1">
                  <ArkCombobox.ItemText class="block font-medium">
                    {{ item.label }}
                  </ArkCombobox.ItemText><p v-if="item.description" class="mt-0.5 truncate text-size-xs text-subtle">
                    {{ item.description }}
                  </p>
                </div><kbd v-if="item.shortcut" class="shrink-0 font-sans text-size-xs text-subtle">{{ item.shortcut }}</kbd>
              </slot>
            </ArkCombobox.Item>
          </template>
        </template>
      </ArkCombobox.Content>
      <CommandPaletteFooter v-if="props.showFooter">
        <template v-if="$slots.footer" #default>
          <slot name="footer" />
        </template>
      </CommandPaletteFooter>
    </slot>
  </ArkCombobox.Root>
</template>
