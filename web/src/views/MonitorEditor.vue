<script setup lang="ts">
import { reactive, ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Save, ArrowLeft, Plus, X, Upload } from '@lucide/vue'
import { Tabs } from '@ark-ui/vue/tabs'
import { api, canEdit } from '../lib/api'
import type { Monitor, Channel, Secret } from '../lib/types'
import {
  newMonitor,
  monitorTypes,
  emptyHTTP,
  emptyTCP,
  emptyDNS,
  emptyHeartbeat,
  emptyCertificate,
} from '../lib/monitor'
import { clone, parseJSON, splitValues, defaults } from '../lib/form'
import { t } from '../lib/preferences'
import { errorText, notify } from '../lib/notices'
import PageHeader from '../components/PageHeader.vue'
import Field from '../components/Field.vue'
import Toggle from '../components/Toggle.vue'
import KeyValues from '../components/KeyValues.vue'
import SecretSelect from '../components/SecretSelect.vue'
import TLSFields from '../components/TLSFields.vue'
import ConnectionFields from '../components/ConnectionFields.vue'
import AsyncState from '../components/AsyncState.vue'
const route = useRoute(),
  router = useRouter(),
  editing = computed(() => !!route.params.id),
  loading = ref(true),
  saving = ref(false),
  error = ref(''),
  channels = ref<Channel[]>([]),
  secrets = ref<Secret[]>([]),
  form = reactive<Monitor>(newMonitor()),
  tags = ref(''),
  dnsValues = ref(''),
  warningDays = ref('30, 14, 7, 1'),
  statusCodes = ref(''),
  statusRanges = ref('200-299'),
  contains = ref(''),
  notContains = ref(''),
  regex = ref(''),
  headerAssertions = ref('[]'),
  jsonAssertions = ref('[]'),
  activeTab = ref('target')
watch(
  () => form.type,
  (type, previous) => {
    if (type === 'heartbeat' && activeTab.value === 'schedule') activeTab.value = 'target'
    if (
      !editing.value &&
      previous === 'certificate' &&
      type !== 'certificate' &&
      form.intervalSeconds === 86400
    )
      form.intervalSeconds = 60
    if (type === 'http' && !form.http) form.http = emptyHTTP()
    if (type === 'tcp' && !form.tcp) form.tcp = emptyTCP()
    if (type === 'dns' && !form.dns) form.dns = emptyDNS()
    if (type === 'heartbeat' && !form.heartbeat) form.heartbeat = emptyHeartbeat()
    if (type === 'certificate' && !form.certificate) {
      form.certificate = emptyCertificate()
      if (!editing.value) form.intervalSeconds = 86400
    }
  },
)
async function load() {
  loading.value = true
  try {
    const [c, s] = await Promise.all([
      api<{ items: Channel[] }>('channels'),
      api<{ items: Secret[] }>('secrets'),
    ])
    channels.value = c.items
    secrets.value = s.items
    if (editing.value) {
      Object.assign(form, await api<Monitor>(`monitors/${route.params.id}`))
      if (form.http) form.http = defaults(emptyHTTP(), form.http)
      if (form.tcp) form.tcp = defaults(emptyTCP(), form.tcp)
      if (form.dns) form.dns = defaults(emptyDNS(), form.dns)
      if (form.heartbeat) form.heartbeat = defaults(emptyHeartbeat(), form.heartbeat)
      if (form.certificate) form.certificate = defaults(emptyCertificate(), form.certificate)
      if (form.http && !form.http.auth.type) form.http.auth.type = 'none'
      tags.value = form.tags?.join(', ') || ''
      dnsValues.value = form.dns?.expectedValues?.join('\n') || ''
      warningDays.value = form.certificate?.warningDays?.join(', ') || warningDays.value
      const a = form.http?.assertions
      if (a) {
        statusCodes.value = a.statusCodes?.join(', ') || ''
        statusRanges.value = a.statusRanges?.map((r) => `${r.min}-${r.max}`).join(', ') || ''
        contains.value = a.textContains?.join('\n') || ''
        notContains.value = a.textNotContains?.join('\n') || ''
        regex.value = a.regex?.join('\n') || ''
        headerAssertions.value = JSON.stringify(a.headers || [], null, 2)
        jsonAssertions.value = JSON.stringify(a.json || [], null, 2)
      }
    }
  } catch (e) {
    error.value = errorText(e)
  } finally {
    loading.value = false
  }
}
onMounted(load)
async function save() {
  saving.value = true
  error.value = ''
  try {
    const payload = clone(form)
    payload.tags = splitValues(tags.value)
    for (const type of ['http', 'tcp', 'dns', 'heartbeat', 'certificate'] as const)
      if (type !== payload.type) delete payload[type]
    if (payload.dns)
      payload.dns.expectedValues = dnsValues.value
        .split('\n')
        .map((x) => x.trim())
        .filter(Boolean)
    if (payload.certificate)
      payload.certificate.warningDays = splitValues(warningDays.value).map(Number)
    if (payload.http) {
      const a = payload.http.assertions
      a.statusCodes = splitValues(statusCodes.value).map(Number)
      a.statusRanges = splitValues(statusRanges.value).map((v) => {
        const [min, max] = v.split('-').map(Number)
        return { min: min!, max: max ?? min! }
      })
      a.textContains = contains.value.split('\n').filter(Boolean)
      a.textNotContains = notContains.value.split('\n').filter(Boolean)
      a.regex = regex.value.split('\n').filter(Boolean)
      a.headers = parseJSON(headerAssertions.value, t('响应头断言', 'Header assertions'))
      a.json = parseJSON(jsonAssertions.value, t('JSON 断言', 'JSON assertions'))
    }
    if (new Blob([JSON.stringify(payload)]).size > 8 * 1024 * 1024)
      throw new Error(
        t(
          '监控配置请求总量不能超过 8 MiB',
          'The monitor configuration request may not exceed 8 MiB',
        ),
      )
    const result = await api<Monitor>(editing.value ? `monitors/${form.id}` : 'monitors', {
      method: editing.value ? 'PATCH' : 'POST',
      body: payload,
    })
    notify(t('监控项已保存', 'Monitor saved'))
    router.push(`/app/monitors/${result.id}`)
  } catch (e) {
    error.value = errorText(e)
  } finally {
    saving.value = false
  }
}
async function addFile(event: Event) {
  const files = (event.target as HTMLInputElement).files
  if (!files || !form.http) return
  for (const file of Array.from(files)) {
    const existingBytes = form.http.body.files.reduce(
      (total, item) => total + Math.ceil((item.base64.length * 3) / 4),
      0,
    )
    if (file.size + existingBytes > 4 * 1024 * 1024) {
      notify(
        t('Multipart 文件原始总量最多 4 MiB', 'Multipart files may total at most 4 MiB'),
        'error',
      )
      continue
    }
    const base64 = await new Promise<string>((resolve, reject) => {
      const r = new FileReader()
      r.onload = () => resolve(String(r.result).split(',')[1] || '')
      r.onerror = reject
      r.readAsDataURL(file)
    })
    form.http.body.files.push({
      field: 'file',
      filename: file.name,
      contentType: file.type || 'application/octet-stream',
      base64,
    })
  }
}
const headerHint =
  '[{"name":"Content-Type","operator":"contains","value":"json"}] · exists / equals / contains / not_contains / regex'
const jsonHint =
  '[{"pointer":"/status","operator":"equals","value":"ok"}] · exists / equals / not_equals / contains / regex'
const charsetOptions = [
  'utf-8',
  'iso-8859-1',
  'windows-1252',
  'gbk',
  'gb18030',
  'shift_jis',
  'big5',
]
</script>
<template>
  <PageHeader
    :title="editing ? t('编辑监控项', 'Edit monitor') : t('创建监控项', 'Create monitor')"
    :description="
      t(
        '定义检查目标、成功条件与故障确认策略。',
        'Define the target, success criteria, and confirmation policy.',
      )
    "
    ><RouterLink :to="editing ? `/app/monitors/${route.params.id}` : '/app/monitors'" class="button"
      ><ArrowLeft :size="15" />{{ t('返回', 'Back') }}</RouterLink
    ><button class="button primary" :disabled="saving || !canEdit()" form="monitor-form">
      <Save :size="15" />{{ saving ? t('保存中…', 'Saving…') : t('保存监控项', 'Save monitor') }}
    </button></PageHeader
  ><AsyncState :pending="loading"
    ><form id="monitor-form" @submit.prevent="save">
      <div v-if="error" class="validation-error" role="alert">{{ error }}</div>
      <section class="card" un-mb="6">
        <div class="form-section">
          <h2>{{ t('基本信息', 'Basic information') }}</h2>
          <p class="muted">
            {{
              t(
                '使用容易辨识的名称，将相关服务整理到同一分组。',
                'Use a recognizable name and group related services together.',
              )
            }}
          </p>
          <div class="form-grid">
            <Field :label="t('显示名称', 'Display name')"
              ><input
                v-model="form.name"
                required
                maxlength="200"
                placeholder="Production API" /></Field
            ><Field
              :label="t('监控类型', 'Monitor type')"
              :hint="
                editing
                  ? t(
                      '监控类型创建后固定；切换类型请新建监控项。',
                      'The monitor type is fixed after creation. Create a new monitor to switch type.',
                    )
                  : undefined
              "
              ><select v-model="form.type" :disabled="editing">
                <option v-for="type in monitorTypes" :key="type.value" :value="type.value">
                  {{ t(type.zh, type.en) }}
                </option>
              </select></Field
            ><Field :label="t('分组', 'Group')"
              ><input
                v-model="form.group"
                :placeholder="t('例如：生产服务', 'e.g. Production')" /></Field
            ><Field :label="t('标签', 'Tags')" :hint="t('使用逗号分隔。', 'Separate with commas.')"
              ><input v-model="tags" placeholder="production, api" /></Field
            ><Field class="span-full" :label="t('备注', 'Description')">
              <textarea v-model="form.description" rows="2" />
            </Field>
            <div class="span-full">
              <Toggle
                v-model="form.enabled"
                :label="t('启用监控', 'Enable monitor')"
                :description="
                  t(
                    '暂停时停止采集，并从可用率统计中排除暂停时长。',
                    'Pausing stops collection and excludes paused time from uptime.',
                  )
                "
              />
            </div>
          </div>
        </div>
      </section>
      <section class="card">
        <Tabs.Root v-model="activeTab"
          ><Tabs.List class="tabs-list"
            ><Tabs.Trigger class="tabs-trigger" value="target">{{
              t('检查目标', 'Check target')
            }}</Tabs.Trigger
            ><Tabs.Trigger v-if="form.type !== 'heartbeat'" class="tabs-trigger" value="schedule">{{
              t('时间与重试', 'Schedule & retries')
            }}</Tabs.Trigger
            ><Tabs.Trigger class="tabs-trigger" value="notifications">{{
              t('通知策略', 'Notifications')
            }}</Tabs.Trigger></Tabs.List
          ><Tabs.Content value="target"
            ><template v-if="form.type === 'http' && form.http"
              ><div class="form-section">
                <h2>{{ t('HTTP 请求', 'HTTP request') }}</h2>
                <p class="muted">
                  {{
                    t(
                      '所有请求方法使用相同的轮次重试配置。',
                      'All request methods use the same round retry configuration.',
                    )
                  }}
                </p>
                <div class="form-grid">
                  <Field :label="t('目标 URL', 'Target URL')" class="span-full"
                    ><input
                      v-model="form.http.url"
                      type="url"
                      required
                      placeholder="https://api.example.com/health" /></Field
                  ><Field :label="t('请求方法', 'Request method')"
                    ><input v-model="form.http.method" list="http-methods" required /><datalist
                      id="http-methods"
                    >
                      <option
                        v-for="method in [
                          'GET',
                          'HEAD',
                          'POST',
                          'PUT',
                          'PATCH',
                          'DELETE',
                          'OPTIONS',
                        ]"
                        :key="method"
                      >
                        {{ method }}
                      </option>
                    </datalist></Field
                  ><Field :label="t('Host 覆盖', 'Host override')"
                    ><input v-model="form.http.host" placeholder="api.example.com"
                  /></Field>
                  <div class="span-full">
                    <label class="field-label">{{ t('查询参数', 'Query parameters') }}</label
                    ><KeyValues v-model="form.http.query" :secrets="secrets" un-mt="2" />
                  </div>
                  <div class="span-full">
                    <label class="field-label">{{
                      t('请求头（支持重复名称）', 'Request headers (repeated names supported)')
                    }}</label
                    ><KeyValues v-model="form.http.headers" :secrets="secrets" un-mt="2" />
                  </div>
                </div>
              </div>
              <div class="form-section">
                <h2>{{ t('请求体', 'Request body') }}</h2>
                <p class="muted">
                  {{
                    t(
                      '请求体格式、字符编码和传输压缩分别设置。',
                      'Configure format, character encoding, and compression separately.',
                    )
                  }}
                </p>
                <div class="form-grid">
                  <Field :label="t('请求体格式', 'Body format')"
                    ><select v-model="form.http.body.format">
                      <option
                        v-for="format in ['none', 'json', 'form', 'multipart', 'text', 'raw']"
                        :key="format"
                        :value="format"
                      >
                        {{ format }}
                      </option>
                    </select></Field
                  ><Field
                    v-if="!['none', 'json'].includes(form.http.body.format)"
                    :label="t('字符编码', 'Character encoding')"
                    ><select v-model="form.http.body.charset">
                      <option v-for="charset in charsetOptions" :key="charset">
                        {{ charset }}
                      </option>
                    </select></Field
                  ><Field
                    v-if="['json', 'text', 'raw'].includes(form.http.body.format)"
                    :label="t('正文秘密引用', 'Body secret reference')"
                    ><SecretSelect
                      v-model="form.http.body.secretRef"
                      :secrets="secrets"
                      optional /></Field
                  ><Field
                    v-if="['text', 'raw'].includes(form.http.body.format)"
                    :label="t('Content-Type', 'Content-Type')"
                    ><input v-model="form.http.body.contentType" placeholder="text/plain" /></Field
                  ><Field
                    v-if="
                      ['json', 'text'].includes(form.http.body.format) && !form.http.body.secretRef
                    "
                    class="span-full"
                    :label="t('正文内容', 'Body content')"
                  >
                    <textarea
                      v-model="form.http.body.text"
                      rows="7"
                      :placeholder="form.http.body.format === 'json' ? '{}' : ''"
                      spellcheck="false"
                    /></Field
                  ><Field
                    v-if="form.http.body.format === 'raw' && !form.http.body.secretRef"
                    class="span-full"
                    :label="t('原始字节（Base64）', 'Raw bytes (Base64)')"
                  >
                    <textarea v-model="form.http.body.base64" spellcheck="false" />
                  </Field>
                  <div
                    v-if="['form', 'multipart'].includes(form.http.body.format)"
                    class="span-full"
                  >
                    <label class="field-label">{{ t('表单字段', 'Form fields') }}</label
                    ><KeyValues v-model="form.http.body.fields" :secrets="secrets" un-mt="2" />
                  </div>
                  <div v-if="form.http.body.format === 'multipart'" class="span-full">
                    <label class="field-label">{{ t('文件', 'Files') }}</label>
                    <div
                      v-for="(file, index) in form.http.body.files"
                      :key="index"
                      class="note"
                      un-flex="~ items-center gap-3"
                      un-mt="2"
                    >
                      <input
                        v-model="file.field"
                        :aria-label="t('字段名称', 'Field name')"
                        placeholder="file"
                      /><span>{{ file.filename }}</span
                      ><button
                        type="button"
                        class="icon-button"
                        @click="form.http.body.files.splice(index, 1)"
                        :aria-label="t('移除文件', 'Remove file')"
                      >
                        <X :size="15" />
                      </button>
                    </div>
                    <label class="button small" un-mt="3"
                      ><Upload :size="13" />{{ t('添加文件', 'Add files')
                      }}<input type="file" multiple un-hidden="" @change="addFile"
                    /></label>
                    <p class="field-hint" un-mt="2">
                      {{
                        t(
                          'Multipart boundary 和 Content-Type 自动生成；文件总量最多 4 MiB。',
                          'Multipart boundary and Content-Type are generated automatically. Files may total at most 4 MiB.',
                        )
                      }}
                    </p>
                  </div>
                </div>
              </div>
              <div class="form-section">
                <h2>{{ t('认证', 'Authentication') }}</h2>
                <p class="muted">
                  {{
                    t('凭据只通过已有秘密引用使用。', 'Credentials use existing secret references.')
                  }}
                </p>
                <div class="form-grid">
                  <Field :label="t('认证类型', 'Authentication type')"
                    ><select v-model="form.http.auth.type">
                      <option value="none">{{ t('无认证', 'None') }}</option>
                      <option value="basic">Basic</option>
                      <option value="bearer">Bearer</option>
                      <option value="header">API Key / Header</option>
                    </select></Field
                  ><Field
                    v-if="form.http.auth.type !== 'none'"
                    :label="t('密码 / Token 秘密', 'Password / token secret')"
                    ><SecretSelect v-model="form.http.auth.secretRef" :secrets="secrets" /></Field
                  ><Field v-if="form.http.auth.type === 'basic'" :label="t('用户名', 'Username')"
                    ><input v-model="form.http.auth.username" /></Field
                  ><Field
                    v-if="form.http.auth.type === 'basic'"
                    :label="t('用户名秘密引用（可选）', 'Username secret (optional)')"
                    ><SecretSelect
                      v-model="form.http.auth.usernameSecretRef"
                      :secrets="secrets"
                      optional /></Field
                  ><Field
                    v-if="form.http.auth.type === 'header'"
                    :label="t('认证头名称', 'Authentication header')"
                    ><input v-model="form.http.auth.header" /></Field
                  ><Field
                    v-if="form.http.auth.type === 'header'"
                    :label="t('值前缀', 'Value prefix')"
                    ><input v-model="form.http.auth.prefix" placeholder="Bearer "
                  /></Field>
                </div>
              </div>
              <div class="form-section">
                <h2>{{ t('成功断言', 'Success assertions') }}</h2>
                <p class="muted">
                  {{
                    t(
                      '所有已配置断言都通过后，才确认本次尝试成功。',
                      'Every configured assertion must pass for a successful attempt.',
                    )
                  }}
                </p>
                <div class="form-grid">
                  <Field
                    :label="t('状态码', 'Status codes')"
                    :hint="
                      t(
                        '逗号分隔；与状态码区间共同接受。',
                        'Comma-separated; accepted together with ranges.',
                      )
                    "
                    ><input v-model="statusCodes" placeholder="200, 204" /></Field
                  ><Field :label="t('状态码区间', 'Status ranges')"
                    ><input v-model="statusRanges" placeholder="200-299, 300-399" /></Field
                  ><Field :label="t('包含文本（每行一项）', 'Text contains (one per line)')">
                    <textarea v-model="contains" /></Field
                  ><Field :label="t('不包含文本（每行一项）', 'Text excludes (one per line)')">
                    <textarea v-model="notContains" /></Field
                  ><Field :label="t('正文正则（每行一项）', 'Body regex (one per line)')">
                    <textarea v-model="regex" spellcheck="false" /></Field
                  ><Field
                    :label="t('响应时间上限（毫秒）', 'Maximum response time (ms)')"
                    :hint="t('0 表示不增加延迟断言。', '0 disables this assertion.')"
                    ><input
                      v-model.number="form.http.assertions.maxLatencyMs"
                      type="number"
                      min="0" /></Field
                  ><Field
                    class="span-full"
                    :label="t('响应头断言', 'Response header assertions')"
                    :hint="headerHint"
                  >
                    <textarea v-model="headerAssertions" rows="4" spellcheck="false" /></Field
                  ><Field
                    class="span-full"
                    :label="
                      t('JSON 字段断言（JSON Pointer）', 'JSON field assertions (JSON Pointer)')
                    "
                    :hint="jsonHint"
                  >
                    <textarea v-model="jsonAssertions" rows="4" spellcheck="false" />
                  </Field>
                </div>
              </div>
              <details class="form-section">
                <summary class="advanced-summary">
                  {{ t('高级：TLS、连接与传输', 'Advanced: TLS, connections & transport') }}
                </summary>
                <h3 un-mb="4">TLS</h3>
                <TLSFields v-model="form.http.tls" :secrets="secrets" />
                <div class="section-divider" />
                <h3 un-mb="4">{{ t('网络连接', 'Network connection') }}</h3>
                <ConnectionFields v-model="form.http.connection" :secrets="secrets" />
                <div class="section-divider" />
                <div class="form-grid">
                  <div class="span-full">
                    <Toggle
                      v-model="form.http.redirects.enabled"
                      :label="t('跟随重定向', 'Follow redirects')"
                    />
                  </div>
                  <Field :label="t('最大重定向次数', 'Maximum redirects')"
                    ><input
                      v-model.number="form.http.redirects.maxHops"
                      type="number"
                      min="1"
                      max="20" /></Field
                  ><Field :label="t('重定向范围', 'Redirect scope')"
                    ><select v-model="form.http.redirects.scope">
                      <option value="same-origin">{{ t('同源', 'Same origin') }}</option>
                      <option value="same-host">{{ t('同主机', 'Same host') }}</option>
                      <option value="any">{{ t('任意目标', 'Any target') }}</option>
                    </select></Field
                  ><Field :label="t('Accept-Encoding', 'Accept-Encoding')"
                    ><select v-model="form.http.acceptEncoding">
                      <option value="gzip">gzip</option>
                      <option value="identity">identity</option>
                    </select></Field
                  ><Field :label="t('响应字符编码', 'Response charset')"
                    ><select v-model="form.http.responseCharset">
                      <option value="">{{ t('按响应声明检测', 'Detect from response') }}</option>
                      <option v-for="charset in charsetOptions" :key="charset">
                        {{ charset }}
                      </option>
                    </select></Field
                  ><Field
                    :label="t('解压后响应大小上限（字节）', 'Decompressed response limit (bytes)')"
                    ><input
                      v-model.number="form.http.maxResponseBytes"
                      type="number"
                      min="1"
                      max="16777216"
                  /></Field>
                  <div>
                    <Toggle
                      v-model="form.http.requestGzip"
                      :label="t('gzip 压缩请求体', 'gzip request body')"
                    />
                  </div>
                </div></details></template
            ><template v-if="form.type === 'tcp' && form.tcp"
              ><div class="form-section">
                <h2>{{ t('TCP 连接', 'TCP connection') }}</h2>
                <p class="muted">
                  {{
                    t(
                      '仅测试连接，或配置发送内容和响应断言。',
                      'Test a connection, or send a payload and assert the response.',
                    )
                  }}
                </p>
                <div class="form-grid">
                  <Field :label="t('主机', 'Host')"
                    ><input v-model="form.tcp.host" required placeholder="db.example.com" /></Field
                  ><Field :label="t('端口', 'Port')"
                    ><input
                      v-model.number="form.tcp.port"
                      type="number"
                      min="1"
                      max="65535"
                      required /></Field
                  ><Field :label="t('发送文本', 'Send text')">
                    <textarea v-model="form.tcp.sendText" /></Field
                  ><Field :label="t('发送字节（Base64）', 'Send bytes (Base64)')">
                    <textarea v-model="form.tcp.sendBase64" spellcheck="false" /></Field
                  ><Field :label="t('发送内容秘密引用', 'Payload secret')"
                    ><SecretSelect
                      v-model="form.tcp.sendSecretRef"
                      :secrets="secrets"
                      optional /></Field
                  ><Field :label="t('字符编码', 'Character encoding')"
                    ><select v-model="form.tcp.charset">
                      <option v-for="charset in charsetOptions" :key="charset">
                        {{ charset }}
                      </option>
                    </select></Field
                  ><Field :label="t('接收包含文本', 'Response contains')"
                    ><input v-model="form.tcp.receiveContains" /></Field
                  ><Field :label="t('接收正则', 'Response regex')"
                    ><input v-model="form.tcp.receiveRegex" /></Field
                  ><Field :label="t('接收大小上限（字节）', 'Receive size limit (bytes)')"
                    ><input
                      v-model.number="form.tcp.maxReceiveBytes"
                      type="number"
                      min="1"
                      max="16777216"
                  /></Field>
                </div>
              </div>
              <details class="form-section">
                <summary class="advanced-summary">
                  {{ t('TLS 与连接设置', 'TLS & connection settings') }}
                </summary>
                <TLSFields v-model="form.tcp.tls" :secrets="secrets" allow-toggle />
                <div class="section-divider" />
                <ConnectionFields
                  v-model="form.tcp.connection"
                  :secrets="secrets"
                /></details></template
            ><template v-if="form.type === 'dns' && form.dns"
              ><div class="form-section">
                <h2>{{ t('DNS 查询', 'DNS query') }}</h2>
                <p class="muted">
                  {{
                    t('验证解析响应码和记录内容。', 'Validate response codes and record values.')
                  }}
                </p>
                <div class="form-grid">
                  <Field :label="t('查询域名', 'Query name')"
                    ><input v-model="form.dns.name" required placeholder="example.com" /></Field
                  ><Field :label="t('记录类型', 'Record type')"
                    ><select v-model="form.dns.recordType">
                      <option
                        v-for="record in [
                          'A',
                          'AAAA',
                          'CNAME',
                          'MX',
                          'TXT',
                          'NS',
                          'SRV',
                          'PTR',
                          'SOA',
                          'CAA',
                        ]"
                        :key="record"
                      >
                        {{ record }}
                      </option>
                    </select></Field
                  ><Field
                    :label="t('解析服务器', 'DNS server')"
                    :hint="
                      t(
                        '留空使用系统解析服务器；填写主机:端口。',
                        'Leave empty for system resolver, or enter host:port.',
                      )
                    "
                    ><input v-model="form.dns.server" placeholder="1.1.1.1:53" /></Field
                  ><Field :label="t('传输协议', 'Protocol')"
                    ><select v-model="form.dns.protocol">
                      <option value="udp">UDP</option>
                      <option value="tcp">TCP</option>
                    </select></Field
                  ><Field :label="t('期望响应码', 'Expected response code')"
                    ><select v-model="form.dns.expectedRCode">
                      <option
                        v-for="code in [
                          'NOERROR',
                          'FORMERR',
                          'SERVFAIL',
                          'NXDOMAIN',
                          'NOTIMP',
                          'REFUSED',
                        ]"
                        :key="code"
                      >
                        {{ code }}
                      </option>
                    </select></Field
                  ><Field :label="t('记录匹配方式', 'Record matching')"
                    ><select v-model="form.dns.matchMode">
                      <option value="contains">
                        {{ t('包含期望记录', 'Contains expected records') }}
                      </option>
                      <option value="exact">{{ t('完全一致', 'Exact set') }}</option>
                    </select></Field
                  ><Field
                    class="span-full"
                    :label="t('期望记录值（每行一项）', 'Expected values (one per line)')"
                  >
                    <textarea v-model="dnsValues" />
                  </Field>
                </div></div></template
            ><template v-if="form.type === 'heartbeat' && form.heartbeat"
              ><div class="form-section">
                <h2>{{ t('被动心跳', 'Heartbeat') }}</h2>
                <p class="muted">
                  {{
                    t(
                      '服务定期上报心跳；超过预期周期和宽限后判定故障。',
                      'Services report periodically. Missing a period plus grace marks the monitor down.',
                    )
                  }}
                </p>
                <div class="form-grid">
                  <Field :label="t('预期周期（秒）', 'Expected period (seconds)')"
                    ><input
                      v-model.number="form.heartbeat.periodSeconds"
                      type="number"
                      min="30"
                      required /></Field
                  ><Field :label="t('宽限时间（秒）', 'Grace period (seconds)')"
                    ><input
                      v-model.number="form.heartbeat.graceSeconds"
                      type="number"
                      min="0"
                      required
                  /></Field>
                </div>
                <div class="note" un-mt="5">
                  {{
                    t(
                      '保存后可在详情页生成或轮换专属上报密钥。首次上报前显示等待数据，支持显式失败上报。',
                      'After saving, generate or rotate the report token on the detail page. The initial state is Unknown; explicit failure reports are supported.',
                    )
                  }}
                </div>
              </div></template
            ><template v-if="form.type === 'certificate' && form.certificate"
              ><div class="form-section">
                <h2>{{ t('证书到期检查', 'Certificate expiry') }}</h2>
                <p class="muted">
                  {{
                    t(
                      '证书风险独立展示，不计入服务可用率。',
                      'Certificate risk is displayed separately and excluded from service uptime.',
                    )
                  }}
                </p>
                <div class="form-grid">
                  <Field :label="t('TLS 主机', 'TLS host')"
                    ><input v-model="form.certificate.host" required /></Field
                  ><Field :label="t('端口', 'Port')"
                    ><input
                      v-model.number="form.certificate.port"
                      type="number"
                      min="1"
                      max="65535" /></Field
                  ><Field
                    class="span-full"
                    :label="t('提醒阈值（天）', 'Warning thresholds (days)')"
                    :hint="
                      t(
                        '逗号分隔；默认 30、14、7、1 天。',
                        'Comma-separated. Defaults: 30, 14, 7, 1 days.',
                      )
                    "
                    ><input v-model="warningDays"
                  /></Field>
                </div>
              </div>
              <details class="form-section">
                <summary class="advanced-summary">
                  {{ t('TLS 与连接设置', 'TLS & connection settings') }}
                </summary>
                <TLSFields v-model="form.certificate.tls" :secrets="secrets" />
                <div class="section-divider" />
                <ConnectionFields
                  v-model="form.certificate.connection"
                  :secrets="secrets"
                /></details></template></Tabs.Content
          ><Tabs.Content v-if="form.type !== 'heartbeat'" value="schedule"
            ><div class="form-section">
              <h2>{{ t('检查计划', 'Check schedule') }}</h2>
              <p class="muted">
                {{
                  t(
                    '固定间隔运行，同一监控项的检查轮次不会重叠；整轮预算不超过间隔。',
                    'Checks run on a fixed cadence without overlap. The entire round fits within the interval.',
                  )
                }}
              </p>
              <div class="form-grid">
                <Field
                  :label="t('检查间隔（秒）', 'Check interval (seconds)')"
                  :hint="
                    form.type === 'certificate'
                      ? t(
                          '证书默认每日检查（86400 秒）。',
                          'Certificates default to daily checks (86400 seconds).',
                        )
                      : t('最小 30 秒。', 'Minimum 30 seconds.')
                  "
                  ><input
                    v-model.number="form.intervalSeconds"
                    type="number"
                    min="30"
                    max="2592000"
                    required /></Field
                ><Field :label="t('每次探测超时（秒）', 'Attempt timeout (seconds)')"
                  ><input
                    v-model.number="form.timeoutSeconds"
                    type="number"
                    min="1"
                    :max="form.intervalSeconds"
                    required /></Field
                ><template v-if="['http', 'tcp', 'dns'].includes(form.type)"
                  ><Field
                    :label="t('追加重试次数', 'Additional retries')"
                    :hint="
                      t(
                        '默认 2；0 表示只进行首次尝试。',
                        'Default 2. Zero means the first attempt only.',
                      )
                    "
                    ><input v-model.number="form.retries" type="number" min="0" max="10" /></Field
                  ><Field :label="t('重试等待（秒）', 'Retry delay (seconds)')"
                    ><input
                      v-model.number="form.retryDelaySeconds"
                      type="number"
                      min="0"
                      :max="form.intervalSeconds" /></Field
                  ><Field :label="t('连续失败确认阈值', 'Consecutive failed rounds')"
                    ><input
                      v-model.number="form.failureThreshold"
                      type="number"
                      min="1"
                      max="100" /></Field
                  ><Field :label="t('连续成功恢复阈值', 'Consecutive successful rounds')"
                    ><input
                      v-model.number="form.recoveryThreshold"
                      type="number"
                      min="1"
                      max="100" /></Field
                ></template>
              </div>
              <p class="note" un-mt="6">
                {{
                  t(
                    '重试次数为上限。预算不足时会提前结束，并在诊断中记录实际尝试数。采集断档显示 Unknown。',
                    'Retry count is a limit. Insufficient budget ends the round early and records the actual attempts. Collection gaps are Unknown.',
                  )
                }}
              </p>
            </div></Tabs.Content
          ><Tabs.Content value="notifications"
            ><div class="form-section">
              <h2>{{ t('通知渠道', 'Notification channels') }}</h2>
              <p class="muted">
                {{
                  t(
                    '选择管理员已配置的渠道，确认故障时逐渠道持久投递。',
                    'Select administrator-configured channels. Confirmed failures create durable deliveries for each channel.',
                  )
                }}
              </p>
              <div class="checkbox-group">
                <label v-for="channel in channels" :key="channel.id" class="checkbox-label"
                  ><input
                    v-model="form.notificationChannelIds"
                    type="checkbox"
                    :value="channel.id"
                  />{{ channel.name
                  }}<span v-if="!channel.enabled" class="muted">{{
                    t('已停用', 'disabled')
                  }}</span></label
                >
              </div>
              <p v-if="!channels.length" class="muted">
                {{
                  t(
                    '尚未配置通知渠道。管理员可在通知页面创建。',
                    'No channels yet. An administrator can create one under Notifications.',
                  )
                }}
              </p>
              <div class="section-divider" />
              <Toggle
                v-if="form.type === 'certificate' && form.certificate"
                v-model="form.certificate.notifyRenewal"
                :label="t('证书续期时通知', 'Notify on certificate renewal')"
                un-mb="5"
              /><Toggle
                v-else
                v-model="form.notifyRecovery"
                :label="t('恢复时通知', 'Notify on recovery')"
                un-mb="5"
              /><Field
                v-if="form.type !== 'certificate'"
                :label="t('持续故障提醒间隔（秒）', 'Repeated outage reminder (seconds)')"
                :hint="
                  t(
                    '0 关闭重复提醒；启用时最小 30 秒。',
                    '0 disables repeated reminders; minimum 30 seconds when enabled.',
                  )
                "
                ><input v-model.number="form.reminderSeconds" type="number" min="0"
              /></Field></div></Tabs.Content
        ></Tabs.Root>
      </section>
      <div class="form-actions">
        <RouterLink to="/app/monitors" class="button">{{ t('取消', 'Cancel') }}</RouterLink
        ><button type="submit" class="button primary" :disabled="saving || !canEdit()">
          <Save :size="15" />{{ t('保存监控项', 'Save monitor') }}
        </button>
      </div>
    </form></AsyncState
  >
</template>
