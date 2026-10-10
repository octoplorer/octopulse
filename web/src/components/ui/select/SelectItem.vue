<script setup lang="ts">
import type { SelectCollectionItem } from './types'
import { Select as ArkSelect } from '@ark-ui/vue/select'
import { onBeforeUnmount, onMounted, onUpdated, shallowRef, watch } from 'vue'
import { useSelectContext } from './context'

const props = defineProps<{
  value: string
  disabled?: boolean
  textValue?: string
}>()
const context = useSelectContext()
const id = Symbol('select-item')
const textElement = shallowRef<HTMLElement>()
const item = shallowRef<SelectCollectionItem>({
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
  <ArkSelect.Item class="mx-1.5 flex min-h-8 cursor-pointer items-center justify-between gap-2 rounded-md px-2 py-1.5 text-size-base outline-none data-[highlighted]:bg-tint focus-visible:ring-2 focus-visible:ring-brand focus-visible:ring-inset group-focus-visible/select-content:data-[highlighted]:ring-2 group-focus-visible/select-content:data-[highlighted]:ring-brand group-focus-visible/select-content:data-[highlighted]:ring-inset data-[disabled]:pointer-events-none data-[disabled]:cursor-not-allowed data-[disabled]:opacity-50" :item="item">
    <ArkSelect.ItemText class="min-w-0 whitespace-normal [overflow-wrap:anywhere]">
      <span ref="textElement"><slot /></span>
    </ArkSelect.ItemText>
    <ArkSelect.ItemIndicator class="flex shrink-0 items-center text-subtle" aria-hidden="true">
      <span class="i-lucide-check size-4" />
    </ArkSelect.ItemIndicator>
  </ArkSelect.Item>
</template>
