<script setup lang="ts">
import type { Secret } from '../client/types.gen'
import type { ConnectionConfigForm } from '../lib/monitor-form'
import { useI18n } from 'vue-i18n'

import Field from './Field.vue'
import SecretSelect from './SecretSelect.vue'
import { FieldGroup, FieldInput } from './ui/field'

defineProps<{ secrets: Secret[] }>()

const { t } = useI18n({ useScope: 'global' })

const model = defineModel<ConnectionConfigForm>({ required: true })
function update<K extends keyof ConnectionConfigForm>(name: K, value: ConnectionConfigForm[K]) {
  model.value = { ...model.value, [name]: value }
}
</script>

<template>
  <FieldGroup>
    <Field :label="t('connectionFields.proxyUrl')" hint="HTTP(S), SOCKS5 / SOCKS5H">
      <FieldInput :model-value="model.proxyUrl" placeholder="socks5://127.0.0.1:1080" @update:model-value="update('proxyUrl', $event)" />
    </Field><Field :label="t('connectionFields.proxyUsername')">
      <FieldInput :model-value="model.proxyUsername" autocomplete="off" @update:model-value="update('proxyUsername', $event)" />
    </Field><Field :label="t('connectionFields.proxyPasswordSecret')">
      <SecretSelect :model-value="model.proxyPasswordSecretRef" :secrets="secrets" optional @update:model-value="update('proxyPasswordSecretRef', $event)" />
    </Field><Field
      :label="t('connectionFields.customDnsServer')"
      :hint="t('connectionFields.hostPortResolutionWithAProxyFollowsThe')"
    >
      <FieldInput :model-value="model.dnsServer" placeholder="1.1.1.1:53" @update:model-value="update('dnsServer', $event)" />
    </Field><Field
      :label="t('connectionFields.fixedConnectionIp')"
      :hint="t('connectionFields.directConnectionsOnlyHostSniStayIntactCannot')"
    >
      <FieldInput :model-value="model.fixedIp" placeholder="192.0.2.1" @update:model-value="update('fixedIp', $event)" />
    </Field>
  </FieldGroup>
</template>
