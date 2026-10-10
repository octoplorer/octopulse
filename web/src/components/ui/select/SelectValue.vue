<script setup lang="ts">
import { Select as ArkSelect } from '@ark-ui/vue/select'

defineProps<{ placeholder?: string }>()
</script>

<template>
  <ArkSelect.ValueText class="min-w-0 truncate data-[placeholder]:text-placeholder" :placeholder="placeholder" />
</template>
