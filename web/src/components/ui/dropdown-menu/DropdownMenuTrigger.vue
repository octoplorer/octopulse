<script setup lang="ts">
import { Menu as ArkMenu } from '@ark-ui/vue/menu'

const props = defineProps<{ asChild?: boolean }>()
</script>

<template>
  <ArkMenu.Trigger :as-child="props.asChild" :class="props.asChild ? undefined : 'inline-flex items-center justify-center rounded-lg'">
    <slot />
  </ArkMenu.Trigger>
</template>
