<script setup lang="ts">
import { Tooltip } from '../tooltip'

withDefaults(defineProps<{ for?: string, showOptional?: boolean, optionalLabel?: string, tooltip?: string, tooltipLabel?: string }>(), { optionalLabel: '(optional)', tooltipLabel: 'More information' })
</script>

<template>
  <label :for="$props.for" class="inline-flex items-center gap-1 text-size-base font-medium text-default"><slot /><span v-if="showOptional" class="font-normal text-subtle">{{ optionalLabel }}</span><Tooltip v-if="tooltip" :content="tooltip"><button type="button" :aria-label="tooltipLabel" class="inline-flex bg-transparent border-0 p-0 text-subtle"><span class="i-lucide-info size-3.5" aria-hidden="true" /></button></Tooltip></label>
</template>
