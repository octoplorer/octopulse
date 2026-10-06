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
    bg="tint"
    border="1 solid line"
    un-text="12px default"
    font="600"
  >
    <slot />
  </component>
</template>
