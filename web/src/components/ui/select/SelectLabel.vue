<script setup lang="ts">
import type { SelectItemGroupLabelBaseProps } from '@ark-ui/vue/select'
import { Select as ArkSelect } from '@ark-ui/vue/select'
import { useForwardProps } from '@ark-ui/vue/utils'

const props = defineProps<SelectItemGroupLabelBaseProps>()
const forwarded = useForwardProps(props)
</script>

<template>
  <ArkSelect.ItemGroupLabel v-bind="forwarded" class="select-label">
    <slot />
  </ArkSelect.ItemGroupLabel>
</template>

<style scoped>
.select-label {
  padding: 6px 14px;
  color: var(--text-color-subtle);
  font-size: 14px;
  font-weight: 600;
}
</style>
