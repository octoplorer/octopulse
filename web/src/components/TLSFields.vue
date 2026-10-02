<script setup lang="ts">
import type { TLSConfig, Secret } from '../lib/types'
import { t } from '../lib/preferences'
import Toggle from './Toggle.vue'
import Field from './Field.vue'
import SecretSelect from './SecretSelect.vue'
const model = defineModel<TLSConfig>({ required: true })
defineProps<{ secrets: Secret[]; allowToggle?: boolean }>()
</script>
<template>
  <div>
    <Toggle
      v-if="allowToggle"
      v-model="model.enabled"
      :label="t('启用 TLS', 'Enable TLS')"
      un-mb="5"
    />
    <div class="form-grid">
      <Field :label="t('SNI 服务器名称', 'SNI server name')"
        ><input
          v-model="model.serverName"
          :placeholder="t('默认使用目标主机名', 'Use target hostname')" /></Field
      ><Field :label="t('自定义 CA 秘密引用', 'Custom CA secret')"
        ><SecretSelect v-model="model.caSecretRef" :secrets="secrets" optional /></Field
      ><Field :label="t('客户端 PEM 证书', 'Client PEM certificate')"
        ><SecretSelect
          v-model="model.clientCertificateSecretRef"
          :secrets="secrets"
          optional /></Field
      ><Field :label="t('客户端私钥', 'Client private key')"
        ><SecretSelect v-model="model.clientKeySecretRef" :secrets="secrets" optional /></Field
      ><Field :label="t('最低 TLS 版本', 'Minimum TLS version')"
        ><select v-model="model.minVersion">
          <option value="">{{ t('默认', 'Default') }}</option>
          <option>1.2</option>
          <option>1.3</option>
        </select></Field
      ><Field :label="t('最高 TLS 版本', 'Maximum TLS version')"
        ><select v-model="model.maxVersion">
          <option value="">{{ t('默认', 'Default') }}</option>
          <option>1.2</option>
          <option>1.3</option>
        </select></Field
      >
      <div class="span-full">
        <Toggle
          v-model="model.insecureSkipVerify"
          :label="t('跳过 TLS 证书验证', 'Skip TLS certificate verification')"
          :description="
            t(
              '适用于已知的自签名服务；默认保持验证。',
              'For known self-signed services. Verification is enabled by default.',
            )
          "
        />
      </div>
    </div>
  </div>
</template>
