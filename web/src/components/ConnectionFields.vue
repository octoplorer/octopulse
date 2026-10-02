<script setup lang="ts">
import type { ConnectionConfig, Secret } from '../lib/types'
import { t } from '../lib/preferences'
import Field from './Field.vue'
import SecretSelect from './SecretSelect.vue'
const model = defineModel<ConnectionConfig>({ required: true })
defineProps<{ secrets: Secret[] }>()
</script>
<template>
  <div class="form-grid">
    <Field :label="t('代理 URL', 'Proxy URL')" hint="HTTP(S), SOCKS5 / SOCKS5H"
      ><input v-model="model.proxyUrl" placeholder="socks5://127.0.0.1:1080" /></Field
    ><Field :label="t('代理用户名', 'Proxy username')"
      ><input v-model="model.proxyUsername" autocomplete="off" /></Field
    ><Field :label="t('代理密码引用', 'Proxy password secret')"
      ><SecretSelect v-model="model.proxyPasswordSecretRef" :secrets="secrets" optional /></Field
    ><Field
      :label="t('自定义 DNS 服务器', 'Custom DNS server')"
      :hint="
        t(
          '主机:端口；代理模式的解析取决于代理协议。',
          'Host:port. Resolution with a proxy follows the proxy protocol.',
        )
      "
      ><input v-model="model.dnsServer" placeholder="1.1.1.1:53" /></Field
    ><Field
      :label="t('固定连接 IP', 'Fixed connection IP')"
      :hint="
        t(
          '仅直接连接；保留 Host/SNI，不能与代理一起使用。',
          'Direct connections only. Host/SNI stay intact; cannot combine with a proxy.',
        )
      "
      ><input v-model="model.fixedIp" placeholder="192.0.2.1"
    /></Field>
  </div>
</template>
