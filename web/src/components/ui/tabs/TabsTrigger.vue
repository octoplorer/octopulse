<script setup lang="ts">
import type { TabTriggerBaseProps } from '@ark-ui/vue/tabs'
import { TabTrigger } from '@ark-ui/vue/tabs'
import { useForwardProps } from '@ark-ui/vue/utils'

const props = defineProps<TabTriggerBaseProps>()
const forwarded = useForwardProps(props)
</script>

<template>
  <TabTrigger
    v-bind="forwarded"
    class="[@media(max-width:700px)]:whitespace-nowrap [&[data-selected]]:text-strong [&[data-selected]]:border-b-current [&[aria-selected=true]]:text-strong [&[aria-selected=true]]:border-b-current"
    p="y-14px x-0"
    border="0 b-2 solid b-transparent"
    bg="transparent"
    un-text="14px subtle"
    font="500"
  >
    <slot />
  </TabTrigger>
</template>
