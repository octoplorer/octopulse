<script setup lang="ts">
import type { Secret } from '../client/types.gen'
import type { NameValueForm } from '../lib/monitor-form'
import { useI18n } from 'vue-i18n'

import SecretSelect from './SecretSelect.vue'
import { Alert } from './ui/alert'
import { Button } from './ui/button'

defineProps<{ secrets: Secret[], nameLabel?: string }>()

const { t } = useI18n({ useScope: 'global' })

const values = defineModel<NameValueForm[]>({ required: true })
</script>

<template>
  <div>
    <div
      v-for="(item, index) in values"
      :key="index"
      class="kv-row [@media(max-width:700px)]:grid-cols-2"
      grid="~ cols-[1fr_1.4fr_1.1fr_auto] gap-2"
      mb="2"
    >
      <input
        v-model="item.name"
        :placeholder="nameLabel || t('common.name')"
        :aria-label="t('common.name')"
      ><input
        v-if="!item.secretRef"
        v-model="item.value"
        :placeholder="t('common.value')"
        :aria-label="t('common.value')"
      ><Alert v-else as="span">
        {{ t('keyValues.valueFromSecret') }}
      </Alert><SecretSelect v-model="item.secretRef" class="[@media(max-width:700px)]:col-start-1" :secrets="secrets" optional /><Button
        type="button"
        :aria-label="t('keyValues.removeField')"
        size="icon"
        @click="values.splice(index, 1)"
      >
        <span class="i-lucide-x" w="15px" h="15px" aria-hidden="true" />
      </Button>
    </div>
    <Button type="button" variant="ghost" size="sm" @click="values.push({ name: '', value: '' })">
      <span class="i-lucide-plus" w="13px" h="13px" aria-hidden="true" />{{ t('keyValues.addField') }}
    </Button>
  </div>
</template>
