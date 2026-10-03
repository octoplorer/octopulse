<script setup lang="ts">
import type { Secret } from '../client/types.gen'
import { useI18n } from 'vue-i18n'

defineProps<{ secrets: Secret[], optional?: boolean }>()

const { t } = useI18n({ useScope: 'global' })

const value = defineModel<string | undefined>()
</script>

<template>
  <select v-model="value">
    <option value="">
      {{ optional ? t('secretSelect.noSecretReference') : t('secretSelect.chooseASecret') }}
    </option>
    <option v-for="secret in secrets" :key="secret.id" :value="secret.id">
      {{ secret.name }}
    </option>
  </select>
</template>
