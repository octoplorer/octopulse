<script setup lang="ts">
import type { Secret } from '../lib/types'
import { t } from '../lib/preferences'
const value = defineModel<string | undefined>()
defineProps<{ secrets: Secret[]; optional?: boolean }>()
</script>
<template>
  <select v-model="value">
    <option value="">
      {{
        optional ? t('不使用秘密引用', 'No secret reference') : t('选择已有秘密', 'Choose a secret')
      }}
    </option>
    <option v-for="secret in secrets" :key="secret.id" :value="secret.id">{{ secret.name }}</option>
  </select>
</template>
