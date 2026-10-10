<script setup lang="ts">
import type { Secret } from '../client/types.gen'
import type { TLSConfigForm } from '../lib/monitor-form'
import { useI18n } from 'vue-i18n'

import { Field, FieldGroup } from './ui/field'
import { Input } from './ui/input'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from './ui/select'
import { Switch } from './ui/switch'

defineProps<{ secrets: Secret[], allowToggle?: boolean }>()

const { t } = useI18n({ useScope: 'global' })

const model = defineModel<TLSConfigForm>({ required: true })
function update<K extends keyof TLSConfigForm>(name: K, value: TLSConfigForm[K]) {
  model.value = { ...model.value, [name]: value }
}
</script>

<template>
  <div>
    <Switch
      v-if="allowToggle"
      :model-value="model.enabled" :label="t('tls-fields.enable-tls')"
      mb="5"
      @update:model-value="update('enabled', $event)"
    />
    <FieldGroup>
      <Field :label="t('tls-fields.sni-server-name')">
        <Input :model-value="model.serverName" :placeholder="t('tls-fields.use-target-hostname')" @update:model-value="update('serverName', String($event ?? ''))" />
      </Field><Field :label="t('tls-fields.custom-ca-secret')">
        <Select :model-value="model.caSecretRef" @update:model-value="update('caSecretRef', $event)">
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
      </Field><Field :label="t('tls-fields.client-pem-certificate')">
        <Select :model-value="model.clientCertificateSecretRef" @update:model-value="update('clientCertificateSecretRef', $event)">
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
      </Field><Field :label="t('tls-fields.client-private-key')">
        <Select :model-value="model.clientKeySecretRef" @update:model-value="update('clientKeySecretRef', $event)">
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
      </Field><Field :label="t('tls-fields.minimum-tls-version')">
        <Select :model-value="model.minVersion" @update:model-value="update('minVersion', $event)">
          <SelectTrigger><SelectValue :placeholder="t('common.default')" /></SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value="">
                {{ t('common.default') }}
              </SelectItem>
              <SelectItem value="1.2">
                1.2
              </SelectItem>
              <SelectItem value="1.3">
                1.3
              </SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
      </Field><Field :label="t('tls-fields.maximum-tls-version')">
        <Select :model-value="model.maxVersion" @update:model-value="update('maxVersion', $event)">
          <SelectTrigger><SelectValue :placeholder="t('common.default')" /></SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value="">
                {{ t('common.default') }}
              </SelectItem>
              <SelectItem value="1.2">
                1.2
              </SelectItem>
              <SelectItem value="1.3">
                1.3
              </SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
      </Field>
      <div class="span-full">
        <Switch
          :model-value="model.insecureSkipVerify" :label="t('tls-fields.skip-tls-certificate-verification')"
          :description="t('tls-fields.for-known-self-signed-services-verification-is-enabled')"
          @update:model-value="update('insecureSkipVerify', $event)"
        />
      </div>
    </FieldGroup>
  </div>
</template>
