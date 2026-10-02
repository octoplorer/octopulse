<script setup lang="ts">
import { t } from '../lib/preferences'
import { errorText } from '../lib/notices'
defineProps<{ pending?: boolean; error?: unknown }>()
defineEmits<{ retry: [] }>()
</script>
<template>
  <div v-if="pending" class="loading-state" role="status">
    <span class="spinner" />{{ t('正在读取数据…', 'Loading data…') }}
  </div>
  <div v-else-if="error" class="error-banner" role="alert">
    {{ errorText(error)
    }}<button class="button ghost" @click="$emit('retry')">{{ t('重试', 'Retry') }}</button>
  </div>
  <slot v-else />
</template>
