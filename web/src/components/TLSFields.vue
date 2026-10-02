<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { TLSConfig, Secret } from '../lib/types'

import Toggle from './Toggle.vue'
import Field from './Field.vue'
import SecretSelect from './SecretSelect.vue'
const { t } = useI18n({ useScope: 'global' })

const model = defineModel<TLSConfig>({ required: true })
defineProps<{ secrets: Secret[]; allowToggle?: boolean }>()
</script>
<template>
  <div>
    <Toggle
      v-if="allowToggle"
      v-model="model.enabled"
      :label="t('tLSFields.enableTls')"
      un-mb="5"
    />
    <div class="form-grid">
      <Field :label="t('tLSFields.sniServerName')"
        ><input v-model="model.serverName" :placeholder="t('tLSFields.useTargetHostname')" /></Field
      ><Field :label="t('tLSFields.customCaSecret')"
        ><SecretSelect v-model="model.caSecretRef" :secrets="secrets" optional /></Field
      ><Field :label="t('tLSFields.clientPemCertificate')"
        ><SecretSelect
          v-model="model.clientCertificateSecretRef"
          :secrets="secrets"
          optional /></Field
      ><Field :label="t('tLSFields.clientPrivateKey')"
        ><SecretSelect v-model="model.clientKeySecretRef" :secrets="secrets" optional /></Field
      ><Field :label="t('tLSFields.minimumTlsVersion')"
        ><select v-model="model.minVersion">
          <option value="">{{ t('common.default') }}</option>
          <option>1.2</option>
          <option>1.3</option>
        </select></Field
      ><Field :label="t('tLSFields.maximumTlsVersion')"
        ><select v-model="model.maxVersion">
          <option value="">{{ t('common.default') }}</option>
          <option>1.2</option>
          <option>1.3</option>
        </select></Field
      >
      <div class="span-full">
        <Toggle
          v-model="model.insecureSkipVerify"
          :label="t('tLSFields.skipTlsCertificateVerification')"
          :description="t('tLSFields.forKnownSelfSignedServicesVerificationIsEnabled')"
        />
      </div>
    </div>
  </div>
</template>
