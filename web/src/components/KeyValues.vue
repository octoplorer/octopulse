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
function update<K extends keyof NameValueForm>(index: number, name: K, value: NameValueForm[K]) {
  values.value = values.value.map((item, i) => i === index ? { ...item, [name]: value } : item)
}
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
        :value="item.name" :placeholder="nameLabel || t('common.name')"
        :aria-label="t('common.name')"
        @input="update(index, 'name', ($event.target as HTMLInputElement).value)"
      ><input
        v-if="!item.secretRef"
        :value="item.value" :placeholder="t('common.value')"
        :aria-label="t('common.value')"
        @input="update(index, 'value', ($event.target as HTMLInputElement).value)"
      ><Alert v-else as="span">
        {{ t('keyValues.valueFromSecret') }}
      </Alert><SecretSelect :model-value="item.secretRef" class="[@media(max-width:700px)]:col-start-1" :secrets="secrets" optional @update:model-value="update(index, 'secretRef', $event)" /><Button
        type="button"
        :aria-label="t('keyValues.removeField')"
        shape="square"
        @click="values = values.filter((_, i) => i !== index)"
      >
        <span class="i-lucide-x" w="15px" h="15px" aria-hidden="true" />
      </Button>
    </div>
    <Button type="button" variant="ghost" size="sm" @click="values = [...values, { name: '', value: '' }]">
      <span class="i-lucide-plus" w="13px" h="13px" aria-hidden="true" />{{ t('keyValues.addField') }}
    </Button>
  </div>
</template>
