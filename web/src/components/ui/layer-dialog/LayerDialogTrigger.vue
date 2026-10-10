<script setup lang="ts">
import { Dialog as ArkDialog } from '@ark-ui/vue/dialog'

const props = defineProps<{ asChild?: boolean }>()
</script>

<template>
  <ArkDialog.Trigger :as-child="props.asChild" :class="props.asChild ? undefined : 'inline-flex items-center justify-center rounded-lg'">
    <slot />
  </ArkDialog.Trigger>
</template>
