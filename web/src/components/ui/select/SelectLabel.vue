<script setup lang="ts">
import type { SelectItemGroupLabelBaseProps } from '@ark-ui/vue/select'
import { Select as ArkSelect } from '@ark-ui/vue/select'
import { useForwardProps } from '@ark-ui/vue/utils'

const props = defineProps<SelectItemGroupLabelBaseProps>()
const forwarded = useForwardProps(props)
</script>

<template>
  <ArkSelect.ItemGroupLabel v-bind="forwarded" class="px-3.5 py-1.5 text-size-sm font-semibold text-subtle">
    <slot />
  </ArkSelect.ItemGroupLabel>
</template>
