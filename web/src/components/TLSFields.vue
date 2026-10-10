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
      :model-value="model.enabled" :label="t('tLSFields.enableTls')"
      mb="5"
      @update:model-value="update('enabled', $event)"
    />
    <FieldGroup>
      <Field :label="t('tLSFields.sniServerName')">
        <Input :model-value="model.serverName" :placeholder="t('tLSFields.useTargetHostname')" @update:model-value="update('serverName', String($event ?? ''))" />
      </Field><Field :label="t('tLSFields.customCaSecret')">
        <Select :model-value="model.caSecretRef" @update:model-value="update('caSecretRef', $event)">
          <SelectTrigger>
            <SelectValue :placeholder="t('secretSelect.noSecretReference')" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value="">
                {{ t('secretSelect.noSecretReference') }}
              </SelectItem>
              <SelectItem v-for="secret in secrets" :key="secret.id" :value="secret.id">
                {{ secret.name }}
              </SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
      </Field><Field :label="t('tLSFields.clientPemCertificate')">
        <Select :model-value="model.clientCertificateSecretRef" @update:model-value="update('clientCertificateSecretRef', $event)">
          <SelectTrigger>
            <SelectValue :placeholder="t('secretSelect.noSecretReference')" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value="">
                {{ t('secretSelect.noSecretReference') }}
              </SelectItem>
              <SelectItem v-for="secret in secrets" :key="secret.id" :value="secret.id">
                {{ secret.name }}
              </SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
      </Field><Field :label="t('tLSFields.clientPrivateKey')">
        <Select :model-value="model.clientKeySecretRef" @update:model-value="update('clientKeySecretRef', $event)">
          <SelectTrigger>
            <SelectValue :placeholder="t('secretSelect.noSecretReference')" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value="">
                {{ t('secretSelect.noSecretReference') }}
              </SelectItem>
              <SelectItem v-for="secret in secrets" :key="secret.id" :value="secret.id">
                {{ secret.name }}
              </SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
      </Field><Field :label="t('tLSFields.minimumTlsVersion')">
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
      </Field><Field :label="t('tLSFields.maximumTlsVersion')">
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
          :model-value="model.insecureSkipVerify" :label="t('tLSFields.skipTlsCertificateVerification')"
          :description="t('tLSFields.forKnownSelfSignedServicesVerificationIsEnabled')"
          @update:model-value="update('insecureSkipVerify', $event)"
        />
      </div>
    </FieldGroup>
  </div>
</template>
