<script setup lang="ts">
import { useI18n } from 'vue-i18n'

import { errorText } from '../lib/errors'
const { t } = useI18n({ useScope: 'global' })

defineProps<{ pending?: boolean; error?: unknown }>()
defineEmits<{ retry: [] }>()
</script>
<template>
  <div v-if="pending" class="loading-state" role="status">
    <span class="spinner" />{{ t('asyncState.loadingData') }}
  </div>
  <div v-else-if="error" class="error-banner" role="alert">
    {{ errorText(error)
    }}<button class="button ghost" @click="$emit('retry')">{{ t('asyncState.retry') }}</button>
  </div>
  <slot v-else />
</template>
