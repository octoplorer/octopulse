<script setup lang="ts">
import type { TabListBaseProps } from '@ark-ui/vue/tabs'
import { TabList } from '@ark-ui/vue/tabs'
import { useForwardProps } from '@ark-ui/vue/utils'

const props = defineProps<TabListBaseProps>()
const forwarded = useForwardProps(props)
</script>

<template>
  <TabList v-bind="forwarded" class="tabs-list" flex="~ items-center gap-24px" border="b-1 solid $border" p="x-20px y-0" bg="$surface">
    <slot />
  </TabList>
</template>
