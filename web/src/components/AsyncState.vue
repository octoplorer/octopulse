<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { errorText } from '../lib/errors'
import { Alert } from './ui/alert'
import { Button } from './ui/button'
import { Spinner } from './ui/spinner'

defineProps<{ pending?: boolean, error?: unknown }>()
defineEmits<{ retry: [] }>()
const { t } = useI18n({ useScope: 'global' })
</script>

<template>
  <div v-if="pending" class="loading-state" flex="~ justify-center items-center gap-10px" p="60px" un-text="12px $muted" role="status">
    <Spinner />{{ t('asyncState.loadingData') }}
  </div>
  <Alert v-else-if="error" variant="destructive">
    {{ errorText(error) }}
    <Button variant="ghost" @click="$emit('retry')">
      {{ t('asyncState.retry') }}
    </Button>
  </Alert>
  <slot v-else />
</template>
