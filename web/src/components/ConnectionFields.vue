<script setup lang="ts">
import type { Secret } from '../client/types.gen'
import type { ConnectionConfigForm } from '../lib/monitor-form'
import { useI18n } from 'vue-i18n'

import Field from './Field.vue'
import SecretSelect from './SecretSelect.vue'
import { FieldGroup } from './ui/field'

defineProps<{ secrets: Secret[] }>()

const { t } = useI18n({ useScope: 'global' })

const model = defineModel<ConnectionConfigForm>({ required: true })
</script>

<template>
  <FieldGroup>
    <Field :label="t('connectionFields.proxyUrl')" hint="HTTP(S), SOCKS5 / SOCKS5H">
      <input v-model="model.proxyUrl" placeholder="socks5://127.0.0.1:1080">
    </Field><Field :label="t('connectionFields.proxyUsername')">
      <input v-model="model.proxyUsername" autocomplete="off">
    </Field><Field :label="t('connectionFields.proxyPasswordSecret')">
      <SecretSelect v-model="model.proxyPasswordSecretRef" :secrets="secrets" optional />
    </Field><Field
      :label="t('connectionFields.customDnsServer')"
      :hint="t('connectionFields.hostPortResolutionWithAProxyFollowsThe')"
    >
      <input v-model="model.dnsServer" placeholder="1.1.1.1:53">
    </Field><Field
      :label="t('connectionFields.fixedConnectionIp')"
      :hint="t('connectionFields.directConnectionsOnlyHostSniStayIntactCannot')"
    >
      <input v-model="model.fixedIp" placeholder="192.0.2.1">
    </Field>
  </FieldGroup>
</template>
