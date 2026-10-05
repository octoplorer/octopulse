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
    class="[@media(max-width:700px)]:whitespace-nowrap [&[data-selected]]:text-$text [&[data-selected]]:border-b-$text [&[aria-selected=true]]:text-$text [&[aria-selected=true]]:border-b-$text"
    p="y-14px x-0"
    border="0 b-2 solid b-transparent"
    bg="transparent"
    un-text="14px $muted"
    font="500"
  >
    <slot />
  </TabTrigger>
</template>
