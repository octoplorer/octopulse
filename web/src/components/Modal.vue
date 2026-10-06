<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { Dialog } from './ui/dialog'

defineProps<{ title: string, description?: string, wide?: boolean }>()
const open = defineModel<boolean>('open', { default: false })
const { t } = useI18n({ useScope: 'global' })
</script>

<template>
  <Dialog v-model:open="open" :title="title" :description="description" :wide="wide" :close-label="t('modal.closeDialog')">
    <slot />
    <template v-if="$slots.footer" #footer>
      <slot name="footer" />
    </template>
  </Dialog>
</template>
