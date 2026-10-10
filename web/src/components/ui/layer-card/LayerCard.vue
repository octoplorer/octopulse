<script setup lang="ts">
import { ark } from '@ark-ui/vue/factory'
</script>

<template>
  <ark.div class="min-w-0 overflow-hidden rounded-xl border border-solid border-line bg-elevated shadow-control">
    <slot />
  </ark.div>
</template>
