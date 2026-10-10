<script setup lang="ts">
const props = defineProps<{ title: string, description?: string, shortcut?: string }>()
</script>

<template>
  <div class="flex min-w-0 flex-1 items-center gap-3">
    <slot name="icon" /><div class="min-w-0 flex-1">
      <p class="font-medium text-default">
        {{ props.title }}
      </p><p v-if="props.description" class="truncate text-size-xs text-subtle">
        {{ props.description }}
      </p>
    </div><kbd v-if="props.shortcut" class="font-sans text-size-xs text-subtle">{{ props.shortcut }}</kbd>
  </div>
</template>
