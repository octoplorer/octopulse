<script setup lang="ts">
import { Tabs as ArkTabs } from '@ark-ui/vue/tabs'

const props = defineProps<{ value: string }>()
</script>

<template>
  <ArkTabs.Content :value="props.value" class="mt-4 outline-none">
    <slot />
  </ArkTabs.Content>
</template>
