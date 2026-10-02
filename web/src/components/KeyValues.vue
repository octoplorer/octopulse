<script setup lang="ts">
import type { NameValue, Secret } from '../lib/types'
import { Plus, X } from '@lucide/vue'
import { useI18n } from 'vue-i18n'

import SecretSelect from './SecretSelect.vue'

defineProps<{ secrets: Secret[], nameLabel?: string }>()

const { t } = useI18n({ useScope: 'global' })

const values = defineModel<NameValue[]>({ required: true })
</script>

<template>
  <div>
    <div
      v-for="(item, index) in values"
      :key="index"
      class="kv-row"
      un-grid="~ cols-[1fr_1.4fr_1.1fr_auto] gap-2"
      un-mb="2"
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
      ><span v-else class="note" un-text="10px">{{ t('keyValues.valueFromSecret') }}</span><SecretSelect v-model="item.secretRef" :secrets="secrets" optional /><button
        type="button"
        class="icon-button"
        :aria-label="t('keyValues.removeField')"
        @click="values.splice(index, 1)"
      >
        <X :size="15" />
      </button>
    </div>
    <button type="button" class="button small ghost" @click="values.push({ name: '', value: '' })">
      <Plus :size="13" />{{ t('keyValues.addField') }}
    </button>
  </div>
</template>

<style scoped>
@media (max-width: 700px) {
  .kv-row {
    grid-template-columns: 1fr 1fr !important;
  }
  .kv-row select {
    grid-column: 1/2;
  }
}
</style>
