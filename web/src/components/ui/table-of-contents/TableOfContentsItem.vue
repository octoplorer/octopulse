<script setup lang="ts">
import { ark } from '@ark-ui/vue/factory'

defineOptions({ inheritAttrs: false })
const props = defineProps<{ href: string, active?: boolean }>()
</script>

<template>
  <li>
    <ark.a v-bind="$attrs" :href="props.href" :aria-current="props.active ? 'location' : undefined" class="block rounded-md px-2 py-1.5 hover:bg-tint hover:text-default" :class="props.active ? 'bg-tint font-medium text-fg-brand' : 'text-subtle'">
      <slot />
    </ark.a>
  </li>
</template>
