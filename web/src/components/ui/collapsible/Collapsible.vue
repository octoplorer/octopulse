<script setup lang="ts">
import { Collapsible as ArkCollapsible } from '@ark-ui/vue/collapsible'

const props = defineProps<{ defaultOpen?: boolean, disabled?: boolean, lazyMount?: boolean, unmountOnExit?: boolean }>()
const open = defineModel<boolean | undefined>('open', { default: undefined })
</script>

<template>
  <ArkCollapsible.Root v-model:open="open" :default-open="props.defaultOpen" :disabled="props.disabled" :lazy-mount="props.lazyMount" :unmount-on-exit="props.unmountOnExit" class="min-w-0">
    <slot />
  </ArkCollapsible.Root>
</template>
