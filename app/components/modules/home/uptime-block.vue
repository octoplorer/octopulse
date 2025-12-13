<script setup lang="ts">
import type { SerializeObject } from 'nitropack'

const props = defineProps<{
  history: SerializeObject<ServiceHistroy>
}>()

const uptimeColor = computed(() => {
  if (props.history.uptime === null) {
    return 'bg-gray-400'
  }
  if (props.history.uptime === 100) {
    return 'bg-emerald-500'
  }
  else if (props.history.uptime > 98) {
    return 'bg-amber-400'
  }
  else {
    return 'bg-rose-500'
  }
})
</script>

<template>
  <div
    class="group relative flex-1 h-8 min-w-[4px] mx-[1px] first:ml-0 last:mr-0"
  >
    <div
      class="w-full h-full rounded-sm opacity-80 hover:opacity-100 transition-all duration-200 cursor-pointer hover:scale-110"
      :class="uptimeColor"
    />
  </div>
</template>
