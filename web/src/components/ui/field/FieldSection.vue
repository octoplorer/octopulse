<script setup lang="ts">
withDefaults(defineProps<{ as?: string }>(), { as: 'div' })
</script>

<template>
  <component :is="as" class="last:border-b-0 [&_h2]:text-15px [&_h2]:mb-5px [&>p]:mb-20px [@media(max-width:700px)]:p-20px [@media(max-width:380px)]:p-16px" p="25px" border="b-1 solid line">
    <slot />
  </component>
</template>
