<script setup lang="ts">
defineProps<{ title: string, description?: string, eyebrow?: string }>()
</script>

<template>
  <header class="page-header" flex="~ justify-between items-center gap-20px" mb="28px">
    <div class="[overflow-wrap:anywhere]" min-w="0">
      <p v-if="eyebrow" class="eyebrow" un-text="12px $muted" font="500" mb="8px">
        {{ eyebrow }}
      </p>
      <h1>{{ title }}</h1>
      <p v-if="description" un-text="13px $muted" mt="6px">
        {{ description }}
      </p>
    </div>
    <div class="header-actions [overflow-wrap:anywhere]" flex="~ items-center gap-9px wrap" min-w="0">
      <slot />
    </div>
  </header>
</template>
