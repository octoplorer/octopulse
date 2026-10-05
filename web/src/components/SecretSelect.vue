<script setup lang="ts">
import type { Secret } from '../client/types.gen'
import { useI18n } from 'vue-i18n'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from './ui/select'

defineProps<{ secrets: Secret[], optional?: boolean }>()

const { t } = useI18n({ useScope: 'global' })

const value = defineModel<string | undefined>()
</script>

<template>
  <Select v-model="value">
    <SelectTrigger>
      <SelectValue :placeholder="optional ? t('secretSelect.noSecretReference') : t('secretSelect.chooseASecret')" />
    </SelectTrigger>
    <SelectContent>
      <SelectGroup>
        <SelectItem value="">
          {{ optional ? t('secretSelect.noSecretReference') : t('secretSelect.chooseASecret') }}
        </SelectItem>
        <SelectItem v-for="secret in secrets" :key="secret.id" :value="secret.id">
          {{ secret.name }}
        </SelectItem>
      </SelectGroup>
    </SelectContent>
  </Select>
</template>
