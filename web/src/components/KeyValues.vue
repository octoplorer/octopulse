<script setup lang="ts">
import { Plus, X } from '@lucide/vue'
import type { NameValue, Secret } from '../lib/types'
import { t } from '../lib/preferences'
import SecretSelect from './SecretSelect.vue'
const values = defineModel<NameValue[]>({ required: true })
defineProps<{ secrets: Secret[]; nameLabel?: string }>()
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
        :placeholder="nameLabel || t('名称', 'Name')"
        :aria-label="t('名称', 'Name')"
      /><input
        v-if="!item.secretRef"
        v-model="item.value"
        :placeholder="t('值', 'Value')"
        :aria-label="t('值', 'Value')"
      /><span v-else class="note" un-text="10px">{{
        t('值来自秘密引用', 'Value from secret')
      }}</span
      ><SecretSelect v-model="item.secretRef" :secrets="secrets" optional /><button
        type="button"
        class="icon-button"
        @click="values.splice(index, 1)"
        :aria-label="t('移除字段', 'Remove field')"
      >
        <X :size="15" />
      </button>
    </div>
    <button type="button" class="button small ghost" @click="values.push({ name: '', value: '' })">
      <Plus :size="13" />{{ t('添加字段', 'Add field') }}
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
