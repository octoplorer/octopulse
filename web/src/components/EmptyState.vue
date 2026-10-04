<script setup lang="ts">
import { Empty, EmptyDescription, EmptyMedia, EmptyTitle } from './ui/empty'

defineProps<{ title: string, description?: string }>()
</script>

<template>
  <Empty>
    <EmptyMedia><span class="i-lucide-activity" size="26px" aria-hidden="true" /></EmptyMedia>
    <EmptyTitle>{{ title }}</EmptyTitle>
    <EmptyDescription v-if="description">
      {{ description }}
    </EmptyDescription>
    <slot />
  </Empty>
</template>
