<script setup lang="ts">
import type { SelectCollectionItem } from './types'
import { Select as ArkSelect } from '@ark-ui/vue/select'
import { onBeforeUnmount, onMounted, onUpdated, ref, watch } from 'vue'
import { useSelectContext } from './context'

const props = defineProps<{
  value: string
  disabled?: boolean
  textValue?: string
}>()
const context = useSelectContext()
const id = Symbol('select-item')
const textElement = ref<HTMLElement>()
const item = ref<SelectCollectionItem>({
  label: props.textValue ?? props.value,
  value: props.value,
  disabled: props.disabled,
})

function updateItem(label = item.value.label) {
  if (
    item.value.label === label
    && item.value.value === props.value
    && item.value.disabled === props.disabled
  ) {
    return
  }

  item.value = {
    label,
    value: props.value,
    disabled: props.disabled,
  }
  context.registerItem(id, item.value)
}

function updateLabel() {
  updateItem(props.textValue ?? textElement.value?.textContent?.trim() ?? props.value)
}

context.registerItem(id, item.value)
watch(() => [props.value, props.disabled, props.textValue] as const, updateLabel)
onMounted(updateLabel)
onUpdated(updateLabel)
onBeforeUnmount(() => context.unregisterItem(id))
</script>

<template>
  <ArkSelect.Item class="select-item" :item="item">
    <ArkSelect.ItemText class="select-item-text">
      <span ref="textElement"><slot /></span>
    </ArkSelect.ItemText>
    <ArkSelect.ItemIndicator class="select-item-indicator" aria-hidden="true">
      <span class="i-lucide-check" />
    </ArkSelect.ItemIndicator>
  </ArkSelect.Item>
</template>

<style scoped>
.select-item {
  display: flex;
  min-height: 36px;
  cursor: pointer;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin: 0 6px;
  padding: 6px 8px;
  font-size: 16px;
  line-height: 1.5;
  border-radius: 4px;
  outline: none;
}

.select-item[data-highlighted] {
  background: var(--color-fill-hover);
}

.select-item:focus-visible {
  box-shadow: inset 0 0 0 2px var(--color-focus);
}

.select-item[data-disabled] {
  cursor: not-allowed;
  opacity: 0.5;
  pointer-events: none;
}

.select-item-text {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.select-item-indicator {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  color: var(--text-color-subtle);
}

.select-item-indicator span {
  width: 16px;
  height: 16px;
}
</style>
