<script setup lang="ts">
const props = defineProps<{ label: string, href?: string, active?: boolean }>()
</script>

<template>
  <li class="space-y-1">
    <a v-if="props.href" :href="props.href" :aria-current="props.active ? 'location' : undefined" class="block rounded-md px-2 py-1.5 font-medium hover:bg-tint" :class="props.active ? 'text-fg-brand' : 'text-default'">{{ props.label }}</a>
    <p v-else class="px-2 py-1.5 font-medium text-default">
      {{ props.label }}
    </p>
    <ul class="ml-2 list-none space-y-1 border-l border-line pl-2">
      <slot />
    </ul>
  </li>
</template>
