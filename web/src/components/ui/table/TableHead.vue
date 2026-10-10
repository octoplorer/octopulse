<script setup lang="ts">
import { stickyHeadClasses } from './sticky'

const props = withDefaults(defineProps<{ sticky?: 'left' | 'right', scope?: 'col' | 'row' | 'colgroup' | 'rowgroup' }>(), { scope: 'col' })
</script>

<template>
  <th :scope="props.scope" class="group relative whitespace-nowrap border-b border-fill bg-base p-3 text-start text-size-sm font-medium text-subtle" :class="props.sticky ? stickyHeadClasses[props.sticky] : ''">
    <slot />
  </th>
</template>
