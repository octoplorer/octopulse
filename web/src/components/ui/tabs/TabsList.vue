<script setup lang="ts">
import type { TabListBaseProps } from '@ark-ui/vue/tabs'
import { TabList } from '@ark-ui/vue/tabs'
import { useForwardProps } from '@ark-ui/vue/utils'

const props = defineProps<TabListBaseProps>()
const forwarded = useForwardProps(props)
</script>

<template>
  <TabList v-bind="forwarded" class="[@media(max-width:700px)]:gap-18px [@media(max-width:700px)]:px-18px [@media(max-width:700px)]:overflow-auto" flex="~ items-center gap-24px" border="b-1 solid line" p="x-20px y-0" bg="$surface">
    <slot />
  </TabList>
</template>
