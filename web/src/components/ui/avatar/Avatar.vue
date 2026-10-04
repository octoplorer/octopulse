<script setup lang="ts">
withDefaults(defineProps<{ as?: keyof HTMLElementTagNameMap }>(), { as: 'span' })
</script>

<template>
  <component
    :is="as"
    class="user-avatar uppercase"
    flex="~ items-center justify-center shrink-0"
    size="32px"
    rounded="full"
    bg="$surface-soft"
    border="1 solid $border"
    un-text="12px $text"
    font="600"
  >
    <slot />
  </component>
</template>
