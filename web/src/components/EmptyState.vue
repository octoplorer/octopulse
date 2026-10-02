<script setup lang="ts">
import { Activity } from '@lucide/vue'

defineProps<{ title: string, description?: string }>()
</script>

<template>
  <div class="empty-state">
    <span class="empty-icon"><Activity :size="26" /></span>
    <h3>{{ title }}</h3>
    <p v-if="description">
      {{ description }}
    </p>
    <slot />
  </div>
</template>
