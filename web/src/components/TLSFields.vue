<script setup lang="ts">
import type { Secret } from '../client/types.gen'
import type { TLSConfigForm } from '../lib/monitor-form'
import { useI18n } from 'vue-i18n'

import Field from './Field.vue'
import SecretSelect from './SecretSelect.vue'
import Toggle from './Toggle.vue'
import { FieldGroup } from './ui/field'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from './ui/select'

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
        <Select v-model="model.minVersion">
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
        <Select v-model="model.maxVersion">
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
          v-model="model.insecureSkipVerify"
          :label="t('tLSFields.skipTlsCertificateVerification')"
          :description="t('tLSFields.forKnownSelfSignedServicesVerificationIsEnabled')"
        />
      </div>
    </FieldGroup>
  </div>
</template>
