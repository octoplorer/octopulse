<script setup lang="ts">
import type { Secret } from '../client/types.gen'
import type { TLSConfigForm } from '../lib/monitor-form'
import { useI18n } from 'vue-i18n'

import Field from './Field.vue'
import SecretSelect from './SecretSelect.vue'
import Toggle from './Toggle.vue'
import { FieldGroup } from './ui/field'

defineProps<{ secrets: Secret[], allowToggle?: boolean }>()

const { t } = useI18n({ useScope: 'global' })

const model = defineModel<TLSConfigForm>({ required: true })
</script>

<template>
  <div>
    <Toggle
      v-if="allowToggle"
      v-model="model.enabled"
      :label="t('tLSFields.enableTls')"
      mb="5"
    />
    <FieldGroup>
      <Field :label="t('tLSFields.sniServerName')">
        <input v-model="model.serverName" :placeholder="t('tLSFields.useTargetHostname')">
      </Field><Field :label="t('tLSFields.customCaSecret')">
        <SecretSelect v-model="model.caSecretRef" :secrets="secrets" optional />
      </Field><Field :label="t('tLSFields.clientPemCertificate')">
        <SecretSelect
          v-model="model.clientCertificateSecretRef"
          :secrets="secrets"
          optional
        />
      </Field><Field :label="t('tLSFields.clientPrivateKey')">
        <SecretSelect v-model="model.clientKeySecretRef" :secrets="secrets" optional />
      </Field><Field :label="t('tLSFields.minimumTlsVersion')">
        <select v-model="model.minVersion">
          <option value="">
            {{ t('common.default') }}
          </option>
          <option>1.2</option>
          <option>1.3</option>
        </select>
      </Field><Field :label="t('tLSFields.maximumTlsVersion')">
        <select v-model="model.maxVersion">
          <option value="">
            {{ t('common.default') }}
          </option>
          <option>1.2</option>
          <option>1.3</option>
        </select>
      </Field>
      <div class="span-full">
        <Toggle
          v-model="model.insecureSkipVerify"
          :label="t('tLSFields.skipTlsCertificateVerification')"
          :description="t('tLSFields.forKnownSelfSignedServicesVerificationIsEnabled')"
        />
      </div>
    </FieldGroup>
  </div>
</template>
