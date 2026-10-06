<script setup lang="ts">
import type { Secret } from '../client/types.gen'
import type { TLSConfigForm } from '../lib/monitor-form'
import { useI18n } from 'vue-i18n'

import Field from './Field.vue'
import SecretSelect from './SecretSelect.vue'
import Toggle from './Toggle.vue'
import { FieldGroup, FieldInput } from './ui/field'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from './ui/select'

defineProps<{ secrets: Secret[], allowToggle?: boolean }>()

const { t } = useI18n({ useScope: 'global' })

const model = defineModel<TLSConfigForm>({ required: true })
function update<K extends keyof TLSConfigForm>(name: K, value: TLSConfigForm[K]) {
  model.value = { ...model.value, [name]: value }
}
</script>

<template>
  <div>
    <Toggle
      v-if="allowToggle"
      :model-value="model.enabled" :label="t('tLSFields.enableTls')"
      mb="5"
      @update:model-value="update('enabled', $event)"
    />
    <FieldGroup>
      <Field :label="t('tLSFields.sniServerName')">
        <FieldInput :model-value="model.serverName" :placeholder="t('tLSFields.useTargetHostname')" @update:model-value="update('serverName', $event)" />
      </Field><Field :label="t('tLSFields.customCaSecret')">
        <SecretSelect :model-value="model.caSecretRef" :secrets="secrets" optional @update:model-value="update('caSecretRef', $event)" />
      </Field><Field :label="t('tLSFields.clientPemCertificate')">
        <SecretSelect
          :model-value="model.clientCertificateSecretRef" :secrets="secrets"
          optional
          @update:model-value="update('clientCertificateSecretRef', $event)"
        />
      </Field><Field :label="t('tLSFields.clientPrivateKey')">
        <SecretSelect :model-value="model.clientKeySecretRef" :secrets="secrets" optional @update:model-value="update('clientKeySecretRef', $event)" />
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
        <Toggle
          :model-value="model.insecureSkipVerify" :label="t('tLSFields.skipTlsCertificateVerification')"
          :description="t('tLSFields.forKnownSelfSignedServicesVerificationIsEnabled')"
          @update:model-value="update('insecureSkipVerify', $event)"
        />
      </div>
    </FieldGroup>
  </div>
</template>
