<script setup lang="ts">
import type { Secret } from '../client/types.gen'
import type { ConnectionConfigForm } from '../lib/monitor-form'
import { useI18n } from 'vue-i18n'

import { Field, FieldGroup } from './ui/field'
import { Input } from './ui/input'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from './ui/select'

defineProps<{ secrets: Secret[] }>()

const { t } = useI18n({ useScope: 'global' })

const model = defineModel<ConnectionConfigForm>({ required: true })
function update<K extends keyof ConnectionConfigForm>(name: K, value: ConnectionConfigForm[K]) {
  model.value = { ...model.value, [name]: value }
}
</script>

<template>
  <FieldGroup>
    <Field :label="t('connection-fields.proxy-url')" description="HTTP(S), SOCKS5 / SOCKS5H">
      <Input :model-value="model.proxyUrl" placeholder="socks5://127.0.0.1:1080" @update:model-value="update('proxyUrl', String($event ?? ''))" />
    </Field><Field :label="t('connection-fields.proxy-username')">
      <Input :model-value="model.proxyUsername" autocomplete="off" @update:model-value="update('proxyUsername', String($event ?? ''))" />
    </Field><Field :label="t('connection-fields.proxy-password-secret')">
      <Select :model-value="model.proxyPasswordSecretRef" @update:model-value="update('proxyPasswordSecretRef', $event)">
        <SelectTrigger>
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
      </Select>
    </Field><Field
      :label="t('connection-fields.custom-dns-server')"
      :description="t('connection-fields.host-port-resolution-with-a-proxy-follows-the')"
    >
      <Input :model-value="model.dnsServer" placeholder="1.1.1.1:53" @update:model-value="update('dnsServer', String($event ?? ''))" />
    </Field><Field
      :label="t('connection-fields.fixed-connection-ip')"
      :description="t('connection-fields.direct-connections-only-host-sni-stay-intact-cannot')"
    >
      <Input :model-value="model.fixedIp" placeholder="192.0.2.1" @update:model-value="update('fixedIp', String($event ?? ''))" />
    </Field>
  </FieldGroup>
</template>
