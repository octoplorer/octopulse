<script setup lang="ts">
defineProps<{ title: string, description?: string, eyebrow?: string }>()
</script>

<template>
  <header class="page-header">
    <div>
      <p v-if="eyebrow" class="eyebrow">
        {{ eyebrow }}
      </p>
      <h1>{{ title }}</h1>
      <p v-if="description" class="muted">
        {{ description }}
      </p>
    </div>
    <div class="header-actions">
      <slot />
    </div>
  </header>
</template>
