<script setup lang="ts">
import { ClipboardText } from '../clipboard-text'

defineProps<{ text: string, textToCopy?: string, copyLabel?: string, copiedLabel?: string }>()
</script>

<template>
  <ClipboardText v-bind="$props" inline>
    <slot>{{ text }}</slot>
  </ClipboardText>
</template>
