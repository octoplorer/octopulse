<script setup lang="ts">
import { Collapsible as ArkCollapsible } from '@ark-ui/vue/collapsible'

const props = defineProps<{ asChild?: boolean }>()
</script>

<template>
  <ArkCollapsible.Trigger :as-child="props.asChild" :class="props.asChild ? undefined : 'group inline-flex cursor-pointer items-center gap-2 rounded-md text-default hover:text-strong data-[disabled]:cursor-not-allowed data-[disabled]:opacity-50'">
    <slot />
  </ArkCollapsible.Trigger>
</template>
