<script setup lang="ts">
import type { MenuRootProps } from '@ark-ui/vue/menu'
import { Menu as ArkMenu } from '@ark-ui/vue/menu'
import { computed } from 'vue'

defineOptions({ inheritAttrs: false })
const props = withDefaults(defineProps<{ defaultOpen?: boolean, loopFocus?: boolean, closeOnSelect?: boolean, placement?: NonNullable<MenuRootProps['positioning']>['placement'] }>(), { placement: 'bottom-start', loopFocus: true, closeOnSelect: true })
const emit = defineEmits<{ select: [value: string] }>()
const open = defineModel<boolean | undefined>('open', { default: undefined })
const positioning = computed(() => ({ placement: props.placement, gutter: 6, strategy: 'fixed' as const }))
</script>

<template>
  <ArkMenu.Root v-bind="$attrs" v-model:open="open" :default-open="props.defaultOpen" :loop-focus="props.loopFocus" :close-on-select="props.closeOnSelect" :positioning="positioning" @select="emit('select', $event.value)">
    <slot />
  </ArkMenu.Root>
</template>
