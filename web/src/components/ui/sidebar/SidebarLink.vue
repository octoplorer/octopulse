<script setup lang="ts">
import { ark } from '@ark-ui/vue/factory'

defineProps<{ asChild?: boolean }>()
</script>

<template>
  <ark.a
    :as-child="asChild"
    class="nav-link [&.router-link-active]:bg-$surface-hover [&.router-link-active:hover]:bg-$surface-hover [&.router-link-active]:font-500"
    flex="~ items-center gap-10px"
    min-h="34px"
    p="7px"
    rounded="8px"
    my="2px"
    un-text="14px $text"
    font="400"
    whitespace="nowrap"
    bg="hover:$surface-soft"
  >
    <slot />
  </ark.a>
</template>
