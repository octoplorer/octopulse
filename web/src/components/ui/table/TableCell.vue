<script setup lang="ts">
import { stickyCellClasses } from './sticky'

const props = defineProps<{ sticky?: 'left' | 'right' }>()
</script>

<template>
  <td class="relative p-3 text-start text-size-sm align-middle" :class="props.sticky ? stickyCellClasses[props.sticky] : ''">
    <slot />
  </td>
</template>
