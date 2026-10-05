<script setup lang="ts">
import type { SelectTriggerBaseProps } from '@ark-ui/vue/select'
import { Select as ArkSelect } from '@ark-ui/vue/select'
import { useForwardProps } from '@ark-ui/vue/utils'
import { computed, mergeProps, useAttrs } from 'vue'

defineOptions({ inheritAttrs: false })

const props = defineProps<SelectTriggerBaseProps>()
const forwarded = useForwardProps(props)
const attrs = useAttrs()
const triggerProps = computed(() => mergeProps(forwarded.value, attrs))
</script>

<template>
  <ArkSelect.Control class="select-control">
    <ArkSelect.Trigger v-bind="triggerProps" class="select-trigger">
      <slot />
      <ArkSelect.Indicator class="select-indicator" aria-hidden="true">
        <span class="i-lucide-chevrons-up-down" />
      </ArkSelect.Indicator>
    </ArkSelect.Trigger>
  </ArkSelect.Control>
</template>

<style scoped>
.select-control {
  display: contents;
}

.select-trigger {
  display: flex;
  width: 100%;
  min-width: 0;
  height: 36px;
  flex-shrink: 0;
  align-items: center;
  justify-content: space-between;
  gap: 6px;
  padding: 0 12px;
  color: var(--text);
  font-size: 16px;
  font-weight: 400;
  line-height: 1.5;
  text-align: left;
  background: var(--surface);
  border: 0;
  border-radius: 8px;
  outline: none;
  box-shadow:
    0 0 0 1px var(--control-border),
    var(--control-shadow);
  user-select: none;
}

.select-trigger:hover:not(:disabled),
.select-trigger[data-state='open'] {
  background: var(--surface-soft);
}

.select-trigger:focus,
.select-trigger:focus-visible {
  outline: none;
  box-shadow: inset 0 0 0 1.5px color-mix(in srgb, var(--focus) 50%, transparent);
}

.select-trigger:focus-visible {
  box-shadow: inset 0 0 0 2px var(--focus);
}

.select-trigger:disabled {
  cursor: not-allowed;
  color: var(--muted);
  opacity: 0.5;
}

.select-indicator {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  color: var(--muted);
}

.select-indicator span {
  width: 16px;
  height: 16px;
}

@media (max-width: 700px) {
  .select-trigger {
    height: 40px;
  }
}
</style>
