<script setup lang="ts">
import { cva } from 'cva'

const props = withDefaults(defineProps<{ layout?: 'auto' | 'fixed' }>(), { layout: 'auto' })
const styles = cva({ base: 'data-table isolate w-full border-collapse text-left text-size-base text-default tabular-nums', variants: { layout: { auto: 'table-auto', fixed: 'table-fixed' } }, defaultVariants: { layout: 'auto' } })
</script>

<template>
  <table :class="styles({ layout: props.layout })">
    <slot />
  </table>
</template>
