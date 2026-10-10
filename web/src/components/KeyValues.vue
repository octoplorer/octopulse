<script setup lang="ts">
import type { Secret } from '../client/types.gen'
import type { NameValueForm } from '../lib/monitor-form'
import { useI18n } from 'vue-i18n'

import { Banner } from './ui/banner'
import { Button } from './ui/button'
import { Input } from './ui/input'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from './ui/select'

defineProps<{ secrets: Secret[], nameLabel?: string }>()

const { t } = useI18n({ useScope: 'global' })

const values = defineModel<NameValueForm[]>({ required: true })
function update<K extends keyof NameValueForm>(index: number, name: K, value: NameValueForm[K]) {
  values.value = values.value.map((item, i) => i === index ? { ...item, [name]: value } : item)
}
</script>

<template>
  <div class="key-values">
    <div
      v-for="(item, index) in values"
      :key="index"
      class="kv-row grid gap-2"
      mb="2"
    >
      <Input
        :model-value="item.name" :placeholder="nameLabel || t('common.name')"
        :aria-label="nameLabel || t('common.name')"
        @update:model-value="update(index, 'name', String($event ?? ''))"
      /><Input
        v-if="!item.secretRef"
        :model-value="item.value" :placeholder="t('common.value')"
        :aria-label="t('common.value')"
        @update:model-value="update(index, 'value', String($event ?? ''))"
      /><Banner v-else variant="secondary" size="sm">
        {{ t('key-values.value-from-secret') }}
      </Banner><Select :model-value="item.secretRef" @update:model-value="update(index, 'secretRef', $event)">
        <SelectTrigger class="kv-secret" :aria-label="t('key-values.secret-reference')">
          <SelectValue :placeholder="t('secret-select.no-secret-reference')" />
        </SelectTrigger>
        <SelectContent>
          <SelectGroup>
            <SelectItem value="">
              {{ t('secret-select.no-secret-reference') }}
            </SelectItem>
            <SelectItem v-for="secret in secrets" :key="secret.id" :value="secret.id">
              {{ secret.name }}
            </SelectItem>
          </SelectGroup>
        </SelectContent>
      </Select><Button
        type="button"
        :aria-label="t('key-values.remove-field')"
        shape="square"
        @click="values = values.filter((_, i) => i !== index)"
      >
        <span class="i-lucide-x" w="15px" h="15px" aria-hidden="true" />
      </Button>
    </div>
    <Button type="button" variant="ghost" size="sm" @click="values = [...values, { name: '', value: '' }]">
      <span class="i-lucide-plus" w="13px" h="13px" aria-hidden="true" />{{ t('key-values.add-field') }}
    </Button>
  </div>
</template>

<style scoped>
.key-values {
  container: key-values / inline-size;
}

.kv-row {
  grid-template-columns: minmax(0, 1fr) minmax(0, 1.4fr) minmax(0, 1.1fr) auto;
}

@container key-values (max-width: 700px) {
  .kv-row {
    grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  }

  .kv-secret {
    grid-column: 1;
  }
}

@container key-values (max-width: 380px) {
  .kv-row {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
