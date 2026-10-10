<script setup lang="ts">
import type { CreateToasterReturn } from '@ark-ui/vue/toast'
import { provide } from 'vue'
import { toastManager, toastManagerKey } from './manager'

const props = defineProps<{ manager?: CreateToasterReturn }>()
provide(toastManagerKey, props.manager ?? toastManager)
</script>

<template>
  <slot />
</template>
