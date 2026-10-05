<script setup lang="ts">
import type { SelectItemGroupBaseProps } from '@ark-ui/vue/select'
import { Select as ArkSelect } from '@ark-ui/vue/select'
import { useForwardProps } from '@ark-ui/vue/utils'

const props = defineProps<SelectItemGroupBaseProps>()
const forwarded = useForwardProps(props)
</script>

<template>
  <ArkSelect.ItemGroup v-bind="forwarded">
    <slot />
  </ArkSelect.ItemGroup>
</template>
