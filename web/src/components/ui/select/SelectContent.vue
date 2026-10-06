<script setup lang="ts">
import { Select as ArkSelect } from '@ark-ui/vue/select'

defineOptions({ inheritAttrs: false })
</script>

<template>
  <Teleport to="body">
    <ArkSelect.Positioner class="select-positioner">
      <ArkSelect.Content
        v-bind="$attrs"
        class="select-content"
      >
        <ArkSelect.List class="select-list">
          <slot />
        </ArkSelect.List>
      </ArkSelect.Content>
    </ArkSelect.Positioner>
  </Teleport>
</template>

<style scoped>
.select-content {
  min-width: var(--reference-width);
  max-width: var(--available-width);
  max-height: var(--available-height);
  overflow: hidden;
  z-index: 200;
  color: var(--text-color-default);
  background: var(--color-base);
  border-radius: 8px;
  outline: none;
  box-shadow:
    var(--shadow-panel),
    0 0 0 1px var(--color-shadow-edge);
}

.select-list {
  max-height: inherit;
  padding: 6px 0;
  overflow-y: auto;
  overscroll-behavior: contain;
  scroll-padding-block: 8px;
}
</style>
