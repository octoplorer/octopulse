<script setup lang="ts">
import type { Channel, Secret } from '../client/types.gen'
import type { MonitorForm } from '../lib/monitor-form'
import { useMutation, useQueryCache } from '@pinia/colada'
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import {
  createMonitorMutation,
  getMonitorQuery,
  listChannelsQuery,
  listSecretsQuery,
  updateMonitorMutation,
} from '../client/@pinia/colada.gen'
import { canEdit } from '../composables/api'
import { notify } from '../composables/notices'
import { errorText } from '../lib/errors'

import { clone, parseJSON, splitValues } from '../lib/form'
import {
  emptyCertificate,
  emptyDNS,
  emptyHeartbeat,
  emptyHTTP,
  emptyTCP,
  monitorTypes,
  newMonitor,
  toMonitorForm,
} from '../lib/monitor'
import AsyncState from './AsyncState.vue'
import ConnectionFields from './ConnectionFields.vue'
import Field from './Field.vue'
import KeyValues from './KeyValues.vue'
import PageHeader from './PageHeader.vue'
import SecretSelect from './SecretSelect.vue'
import TLSFields from './TLSFields.vue'
import Toggle from './Toggle.vue'
import { Alert } from './ui/alert'
import { Button } from './ui/button'
import { Card } from './ui/card'
import { FieldActions, FieldDescription, FieldGroup, FieldLabel, FieldSection } from './ui/field'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from './ui/select'
import { Separator } from './ui/separator'
import { TabsContent, TabsList, TabsRoot, TabsTrigger } from './ui/tabs'

const { t } = useI18n({ useScope: 'global' })
const queryCache = useQueryCache()
const createMonitor = useMutation(createMonitorMutation())
const updateMonitor = useMutation(updateMonitorMutation())

const route = useRoute<'/app/(admin)/monitors/new' | '/app/(admin)/monitors/[id]/edit'>()
const router = useRouter()
const id = computed(() => ('id' in route.params ? route.params.id : undefined))
const editing = computed(() => !!id.value)
const loading = ref(true)
const saving = ref(false)
const error = ref('')
const channels = ref<Channel[]>([])
const secrets = ref<Secret[]>([])
const form = reactive<MonitorForm>(newMonitor())
const tags = ref('')
const dnsValues = ref('')
const warningDays = ref('30, 14, 7, 1')
const statusCodes = ref('')
const statusRanges = ref('200-299')
const contains = ref('')
const notContains = ref('')
const regex = ref('')
const headerAssertions = ref('[]')
const jsonAssertions = ref('[]')
const activeTab = ref('target')
watch(
  () => form.type,
  (type, previous) => {
    if (type === 'heartbeat' && activeTab.value === 'schedule')
      activeTab.value = 'target'
    if (
      !editing.value
      && previous === 'certificate'
      && type !== 'certificate'
      && form.intervalSeconds === 86400
    ) {
      form.intervalSeconds = 60
    }
    if (type === 'http' && !form.http)
      form.http = emptyHTTP()
    if (type === 'tcp' && !form.tcp)
      form.tcp = emptyTCP()
    if (type === 'dns' && !form.dns)
      form.dns = emptyDNS()
    if (type === 'heartbeat' && !form.heartbeat)
      form.heartbeat = emptyHeartbeat()
    if (type === 'certificate' && !form.certificate) {
      form.certificate = emptyCertificate()
      if (!editing.value)
        form.intervalSeconds = 86400
    }
  },
)
async function load() {
  loading.value = true
  try {
    const [c, s] = await Promise.all([
      queryCache.refresh(queryCache.ensure({ ...listChannelsQuery(), staleTime: 0 })),
      queryCache.refresh(queryCache.ensure({ ...listSecretsQuery(), staleTime: 0 })),
    ])
    if (c.status !== 'success')
      throw c.error || new Error(t('errors.requestFailed'))
    if (s.status !== 'success')
      throw s.error || new Error(t('errors.requestFailed'))
    channels.value = clone(c.data.items)
    secrets.value = clone(s.data.items)
    if (editing.value) {
      const monitor = await queryCache.refresh(
        queryCache.ensure({ ...getMonitorQuery({ path: { id: id.value! } }), staleTime: 0 }),
      )
      if (monitor.status !== 'success')
        throw monitor.error || new Error(t('errors.requestFailed'))
      Object.assign(form, toMonitorForm(monitor.data))
      tags.value = form.tags?.join(', ') || ''
      dnsValues.value = form.dns?.expectedValues?.join('\n') || ''
      warningDays.value = form.certificate?.warningDays?.join(', ') || warningDays.value
      const a = form.http?.assertions
      if (a) {
        statusCodes.value = a.statusCodes?.join(', ') || ''
        statusRanges.value = a.statusRanges?.map(r => `${r.min}-${r.max}`).join(', ') || ''
        contains.value = a.textContains?.join('\n') || ''
        notContains.value = a.textNotContains?.join('\n') || ''
        regex.value = a.regex?.join('\n') || ''
        headerAssertions.value = JSON.stringify(a.headers || [], null, 2)
        jsonAssertions.value = JSON.stringify(a.json || [], null, 2)
      }
    }
  }
  catch (e) {
    error.value = errorText(e)
  }
  finally {
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
    for (const type of ['http', 'tcp', 'dns', 'heartbeat', 'certificate'] as const) {
      if (type !== payload.type)
        delete payload[type]
    }
    if (payload.dns) {
      payload.dns.expectedValues = dnsValues.value
        .split('\n')
        .map(x => x.trim())
        .filter(Boolean)
    }
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
      a.headers = parseJSON(headerAssertions.value, t('monitorEditor.headerAssertions'))
      a.json = parseJSON(jsonAssertions.value, t('monitorEditor.jsonAssertions'))
    }
    if (new Blob([JSON.stringify(payload)]).size > 8 * 1024 * 1024)
      throw new Error(t('monitorEditor.theMonitorConfigurationRequestMayNotExceed8'))
    const result = await (editing.value
      ? updateMonitor.mutateAsync({ path: { id: form.id }, body: payload })
      : createMonitor.mutateAsync({ body: payload }))
    notify(t('monitorEditor.monitorSaved'))
    router.push(`/app/monitors/${result.id}`)
  }
  catch (e) {
    error.value = errorText(e)
  }
  finally {
    saving.value = false
  }
}
async function addFile(event: Event) {
  const files = (event.target as HTMLInputElement).files
  if (!files || !form.http)
    return
  for (const file of Array.from(files)) {
    const existingBytes = form.http.body.files.reduce(
      (total, item) => total + Math.ceil((item.base64.length * 3) / 4),
      0,
    )
    if (file.size + existingBytes > 4 * 1024 * 1024) {
      notify(t('monitorEditor.multipartFilesMayTotalAtMost4Mib'), 'error')
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
const headerHint
  = '[{"name":"Content-Type","operator":"contains","value":"json"}] · exists / equals / contains / not_contains / regex'
const jsonHint
  = '[{"pointer":"/status","operator":"equals","value":"ok"}] · exists / equals / not_equals / contains / regex'
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
    :title="editing ? t('monitorEditor.editMonitor') : t('common.createMonitor')"
    :description="t('monitorEditor.defineTheTargetSuccessCriteriaAndConfirmationPolicy')"
  >
    <Button as-child>
      <RouterLink :to="editing ? `/app/monitors/${id}` : '/app/monitors'">
        <span class="i-lucide-arrow-left" w="15px" h="15px" aria-hidden="true" />{{ t('monitorEditor.back') }}
      </RouterLink>
    </Button><Button :disabled="saving || !canEdit()" form="monitor-form" variant="primary">
      <span class="i-lucide-save" w="15px" h="15px" aria-hidden="true" />{{ saving ? t('monitorEditor.saving') : t('common.saveMonitor') }}
    </Button>
  </PageHeader><AsyncState :pending="loading">
    <form id="monitor-form" @submit.prevent="save">
      <Alert v-if="error" role="alert" variant="validation">
        {{ error }}
      </Alert>
      <Card mb="6" as="section">
        <FieldSection>
          <h2>{{ t('monitorEditor.basicInformation') }}</h2>
          <p un-text="13px $muted">
            {{ t('monitorEditor.useARecognizableNameAndGroupRelatedServices') }}
          </p>
          <FieldGroup>
            <Field :label="t('common.displayName')">
              <input
                v-model="form.name"
                required
                maxlength="200"
                :placeholder="t('monitorEditor.namePlaceholder')"
              >
            </Field><Field
              :label="t('monitorEditor.monitorType')"
              :hint="
                editing ? t('monitorEditor.theMonitorTypeIsFixedAfterCreationCreate') : undefined
              "
            >
              <Select v-model="form.type" :disabled="editing">
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    <SelectItem v-for="type in monitorTypes" :key="type.value" :value="type.value">
                      {{ t(type.label) }}
                    </SelectItem>
                  </SelectGroup>
                </SelectContent>
              </Select>
            </Field><Field :label="t('common.group')">
              <input v-model="form.group" :placeholder="t('monitorEditor.eGProduction')">
            </Field><Field :label="t('monitorEditor.tags')" :hint="t('monitorEditor.separateWithCommas')">
              <input v-model="tags" placeholder="production, api">
            </Field><Field class="span-full" :label="t('monitorEditor.description')">
              <textarea v-model="form.description" rows="2" />
            </Field>
            <div class="span-full">
              <Toggle
                v-model="form.enabled"
                :label="t('monitorEditor.enableMonitor')"
                :description="t('monitorEditor.pausingStopsCollectionAndExcludesPausedTimeFrom')"
              />
            </div>
          </FieldGroup>
        </FieldSection>
      </Card>
      <Card as="section">
        <TabsRoot v-model="activeTab">
          <TabsList>
            <TabsTrigger value="target">
              {{
                t('monitorEditor.checkTarget')
              }}
            </TabsTrigger><TabsTrigger v-if="form.type !== 'heartbeat'" value="schedule">
              {{
                t('monitorEditor.scheduleRetries')
              }}
            </TabsTrigger><TabsTrigger value="notifications">
              {{
                t('monitorEditor.notifications')
              }}
            </TabsTrigger>
          </TabsList><TabsContent value="target">
            <template v-if="form.type === 'http' && form.http">
              <FieldSection>
                <h2>{{ t('monitorEditor.httpRequest') }}</h2>
                <p un-text="13px $muted">
                  {{ t('monitorEditor.allRequestMethodsUseTheSameRoundRetry') }}
                </p>
                <FieldGroup>
                  <Field :label="t('monitorEditor.targetUrl')" class="span-full">
                    <input
                      v-model="form.http.url"
                      type="url"
                      required
                      placeholder="https://api.example.com/health"
                    >
                  </Field><Field :label="t('monitorEditor.requestMethod')">
                    <input v-model="form.http.method" list="http-methods" required><datalist
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
                    </datalist>
                  </Field><Field :label="t('monitorEditor.hostOverride')">
                    <input v-model="form.http.host" placeholder="api.example.com">
                  </Field>
                  <div class="span-full">
                    <FieldLabel as="label">
                      {{ t('monitorEditor.queryParameters') }}
                    </FieldLabel><KeyValues v-model="form.http.query" :secrets="secrets" mt="2" />
                  </div>
                  <div class="span-full">
                    <FieldLabel as="label">
                      {{
                        t('monitorEditor.requestHeadersRepeatedNamesSupported')
                      }}
                    </FieldLabel><KeyValues v-model="form.http.headers" :secrets="secrets" mt="2" />
                  </div>
                </FieldGroup>
              </FieldSection>
              <FieldSection>
                <h2>{{ t('monitorEditor.requestBody') }}</h2>
                <p un-text="13px $muted">
                  {{ t('monitorEditor.configureFormatCharacterEncodingAndCompressionSeparately') }}
                </p>
                <FieldGroup>
                  <Field :label="t('monitorEditor.bodyFormat')">
                    <Select v-model="form.http.body.format">
                      <SelectTrigger><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectGroup>
                          <SelectItem
                            v-for="format in ['none', 'json', 'form', 'multipart', 'text', 'raw']"
                            :key="format"
                            :value="format"
                          >
                            {{ format }}
                          </SelectItem>
                        </SelectGroup>
                      </SelectContent>
                    </Select>
                  </Field><Field
                    v-if="!['none', 'json'].includes(form.http.body.format)"
                    :label="t('common.characterEncoding')"
                  >
                    <Select v-model="form.http.body.charset">
                      <SelectTrigger><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectGroup>
                          <SelectItem v-for="charset in charsetOptions" :key="charset" :value="charset">
                            {{ charset }}
                          </SelectItem>
                        </SelectGroup>
                      </SelectContent>
                    </Select>
                  </Field><Field
                    v-if="['json', 'text', 'raw'].includes(form.http.body.format)"
                    :label="t('monitorEditor.bodySecretReference')"
                  >
                    <SecretSelect
                      v-model="form.http.body.secretRef"
                      :secrets="secrets"
                      optional
                    />
                  </Field><Field
                    v-if="['text', 'raw'].includes(form.http.body.format)"
                    :label="t('monitorEditor.contentType')"
                  >
                    <input v-model="form.http.body.contentType" placeholder="text/plain">
                  </Field><Field
                    v-if="
                      ['json', 'text'].includes(form.http.body.format) && !form.http.body.secretRef
                    "
                    class="span-full"
                    :label="t('monitorEditor.bodyContent')"
                  >
                    <textarea
                      v-model="form.http.body.text"
                      rows="7"
                      :placeholder="form.http.body.format === 'json' ? '{}' : ''"
                      spellcheck="false"
                    />
                  </Field><Field
                    v-if="form.http.body.format === 'raw' && !form.http.body.secretRef"
                    class="span-full"
                    :label="t('monitorEditor.rawBytesBase64')"
                  >
                    <textarea v-model="form.http.body.base64" spellcheck="false" />
                  </Field>
                  <div
                    v-if="['form', 'multipart'].includes(form.http.body.format)"
                    class="span-full"
                  >
                    <FieldLabel as="label">
                      {{ t('monitorEditor.formFields') }}
                    </FieldLabel><KeyValues v-model="form.http.body.fields" :secrets="secrets" mt="2" />
                  </div>
                  <div v-if="form.http.body.format === 'multipart'" class="span-full">
                    <FieldLabel as="label">
                      {{ t('monitorEditor.files') }}
                    </FieldLabel>
                    <Alert
                      v-for="(file, index) in form.http.body.files"
                      :key="index"
                      flex="~ items-center gap-3"
                      mt="2"
                    >
                      <input
                        v-model="file.field"
                        :aria-label="t('monitorEditor.fieldName')"
                        placeholder="file"
                      ><span>{{ file.filename }}</span><Button
                        type="button"
                        :aria-label="t('monitorEditor.removeFile')"
                        size="icon"
                        @click="form.http.body.files.splice(index, 1)"
                      >
                        <span class="i-lucide-x" w="15px" h="15px" aria-hidden="true" />
                      </Button>
                    </Alert>
                    <Button as-child size="sm">
                      <label mt="3"><span class="i-lucide-upload" w="13px" h="13px" aria-hidden="true" />{{ t('monitorEditor.addFiles')
                      }}<input type="file" multiple hidden="" @change="addFile"></label>
                    </Button>
                    <FieldDescription as="p" mt="2">
                      {{
                        t('monitorEditor.multipartBoundaryAndContentTypeAreGeneratedAutomatically')
                      }}
                    </FieldDescription>
                  </div>
                </FieldGroup>
              </FieldSection>
              <FieldSection>
                <h2>{{ t('monitorEditor.authentication') }}</h2>
                <p un-text="13px $muted">
                  {{ t('monitorEditor.credentialsUseExistingSecretReferences') }}
                </p>
                <FieldGroup>
                  <Field :label="t('monitorEditor.authenticationType')">
                    <Select v-model="form.http.auth.type">
                      <SelectTrigger><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectGroup>
                          <SelectItem value="none">
                            {{ t('monitorEditor.none') }}
                          </SelectItem>
                          <SelectItem value="basic">
                            Basic
                          </SelectItem>
                          <SelectItem value="bearer">
                            Bearer
                          </SelectItem>
                          <SelectItem value="header">
                            {{ t('monitorEditor.apiKeyHeader') }}
                          </SelectItem>
                        </SelectGroup>
                      </SelectContent>
                    </Select>
                  </Field><Field
                    v-if="form.http.auth.type !== 'none'"
                    :label="t('monitorEditor.passwordTokenSecret')"
                  >
                    <SecretSelect v-model="form.http.auth.secretRef" :secrets="secrets" />
                  </Field><Field v-if="form.http.auth.type === 'basic'" :label="t('common.username')">
                    <input v-model="form.http.auth.username">
                  </Field><Field
                    v-if="form.http.auth.type === 'basic'"
                    :label="t('monitorEditor.usernameSecretOptional')"
                  >
                    <SecretSelect
                      v-model="form.http.auth.usernameSecretRef"
                      :secrets="secrets"
                      optional
                    />
                  </Field><Field
                    v-if="form.http.auth.type === 'header'"
                    :label="t('monitorEditor.authenticationHeader')"
                  >
                    <input v-model="form.http.auth.header">
                  </Field><Field
                    v-if="form.http.auth.type === 'header'"
                    :label="t('monitorEditor.valuePrefix')"
                  >
                    <input v-model="form.http.auth.prefix" placeholder="Bearer ">
                  </Field>
                </FieldGroup>
              </FieldSection>
              <FieldSection>
                <h2>{{ t('monitorEditor.successAssertions') }}</h2>
                <p un-text="13px $muted">
                  {{ t('monitorEditor.everyConfiguredAssertionMustPassForASuccessful') }}
                </p>
                <FieldGroup>
                  <Field
                    :label="t('monitorEditor.statusCodes')"
                    :hint="t('monitorEditor.commaSeparatedAcceptedTogetherWithRanges')"
                  >
                    <input v-model="statusCodes" placeholder="200, 204">
                  </Field><Field :label="t('monitorEditor.statusRanges')">
                    <input v-model="statusRanges" placeholder="200-299, 300-399">
                  </Field><Field :label="t('monitorEditor.textContainsOnePerLine')">
                    <textarea v-model="contains" />
                  </Field><Field :label="t('monitorEditor.textExcludesOnePerLine')">
                    <textarea v-model="notContains" />
                  </Field><Field :label="t('monitorEditor.bodyRegexOnePerLine')">
                    <textarea v-model="regex" spellcheck="false" />
                  </Field><Field
                    :label="t('monitorEditor.maximumResponseTimeMs')"
                    :hint="t('monitorEditor.0DisablesThisAssertion')"
                  >
                    <input
                      v-model.number="form.http.assertions.maxLatencyMs"
                      type="number"
                      min="0"
                    >
                  </Field><Field
                    class="span-full"
                    :label="t('monitorEditor.responseHeaderAssertions')"
                    :hint="headerHint"
                  >
                    <textarea v-model="headerAssertions" rows="4" spellcheck="false" />
                  </Field><Field
                    class="span-full"
                    :label="t('monitorEditor.jsonFieldAssertionsJsonPointer')"
                    :hint="jsonHint"
                  >
                    <textarea v-model="jsonAssertions" rows="4" spellcheck="false" />
                  </Field>
                </FieldGroup>
              </FieldSection>
              <FieldSection as="details">
                <summary cursor="pointer" un-text="12px $link" font="600" py="15px" px="0">
                  {{ t('monitorEditor.advancedTlsConnectionsTransport') }}
                </summary>
                <h3 mb="4">
                  TLS
                </h3>
                <TLSFields v-model="form.http.tls" :secrets="secrets" />
                <Separator />
                <h3 mb="4">
                  {{ t('monitorEditor.networkConnection') }}
                </h3>
                <ConnectionFields v-model="form.http.connection" :secrets="secrets" />
                <Separator />
                <FieldGroup>
                  <div class="span-full">
                    <Toggle
                      v-model="form.http.redirects.enabled"
                      :label="t('monitorEditor.followRedirects')"
                    />
                  </div>
                  <Field :label="t('monitorEditor.maximumRedirects')">
                    <input
                      v-model.number="form.http.redirects.maxHops"
                      type="number"
                      min="1"
                      max="20"
                    >
                  </Field><Field :label="t('monitorEditor.redirectScope')">
                    <Select v-model="form.http.redirects.scope">
                      <SelectTrigger><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectGroup>
                          <SelectItem value="same-origin">
                            {{ t('monitorEditor.sameOrigin') }}
                          </SelectItem>
                          <SelectItem value="same-host">
                            {{ t('monitorEditor.sameHost') }}
                          </SelectItem>
                          <SelectItem value="any">
                            {{ t('monitorEditor.anyTarget') }}
                          </SelectItem>
                        </SelectGroup>
                      </SelectContent>
                    </Select>
                  </Field><Field :label="t('monitorEditor.acceptEncoding')">
                    <Select v-model="form.http.acceptEncoding">
                      <SelectTrigger><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectGroup>
                          <SelectItem value="gzip">
                            gzip
                          </SelectItem>
                          <SelectItem value="identity">
                            identity
                          </SelectItem>
                        </SelectGroup>
                      </SelectContent>
                    </Select>
                  </Field><Field :label="t('monitorEditor.responseCharset')">
                    <Select v-model="form.http.responseCharset">
                      <SelectTrigger><SelectValue :placeholder="t('monitorEditor.detectFromResponse')" /></SelectTrigger>
                      <SelectContent>
                        <SelectGroup>
                          <SelectItem value="">
                            {{ t('monitorEditor.detectFromResponse') }}
                          </SelectItem>
                          <SelectItem v-for="charset in charsetOptions" :key="charset" :value="charset">
                            {{ charset }}
                          </SelectItem>
                        </SelectGroup>
                      </SelectContent>
                    </Select>
                  </Field><Field :label="t('monitorEditor.decompressedResponseLimitBytes')">
                    <input
                      v-model.number="form.http.maxResponseBytes"
                      type="number"
                      min="1"
                      max="16777216"
                    >
                  </Field>
                  <div>
                    <Toggle
                      v-model="form.http.requestGzip"
                      :label="t('monitorEditor.gzipRequestBody')"
                    />
                  </div>
                </FieldGroup>
              </FieldSection>
            </template><template v-if="form.type === 'tcp' && form.tcp">
              <FieldSection>
                <h2>{{ t('monitorTypes.tcpConnection') }}</h2>
                <p un-text="13px $muted">
                  {{ t('monitorEditor.testAConnectionOrSendAPayloadAnd') }}
                </p>
                <FieldGroup>
                  <Field :label="t('monitorEditor.host')">
                    <input v-model="form.tcp.host" required placeholder="db.example.com">
                  </Field><Field :label="t('common.port')">
                    <input
                      v-model.number="form.tcp.port"
                      type="number"
                      min="1"
                      max="65535"
                      required
                    >
                  </Field><Field :label="t('monitorEditor.sendText')">
                    <textarea v-model="form.tcp.sendText" />
                  </Field><Field :label="t('monitorEditor.sendBytesBase64')">
                    <textarea v-model="form.tcp.sendBase64" spellcheck="false" />
                  </Field><Field :label="t('monitorEditor.payloadSecret')">
                    <SecretSelect
                      v-model="form.tcp.sendSecretRef"
                      :secrets="secrets"
                      optional
                    />
                  </Field><Field :label="t('common.characterEncoding')">
                    <Select v-model="form.tcp.charset">
                      <SelectTrigger><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectGroup>
                          <SelectItem v-for="charset in charsetOptions" :key="charset" :value="charset">
                            {{ charset }}
                          </SelectItem>
                        </SelectGroup>
                      </SelectContent>
                    </Select>
                  </Field><Field :label="t('monitorEditor.responseContains')">
                    <input v-model="form.tcp.receiveContains">
                  </Field><Field :label="t('monitorEditor.responseRegex')">
                    <input v-model="form.tcp.receiveRegex">
                  </Field><Field :label="t('monitorEditor.receiveSizeLimitBytes')">
                    <input
                      v-model.number="form.tcp.maxReceiveBytes"
                      type="number"
                      min="1"
                      max="16777216"
                    >
                  </Field>
                </FieldGroup>
              </FieldSection>
              <FieldSection as="details">
                <summary cursor="pointer" un-text="12px $link" font="600" py="15px" px="0">
                  {{ t('common.tlsConnectionSettings') }}
                </summary>
                <TLSFields v-model="form.tcp.tls" :secrets="secrets" allow-toggle />
                <Separator />
                <ConnectionFields
                  v-model="form.tcp.connection"
                  :secrets="secrets"
                />
              </FieldSection>
            </template><template v-if="form.type === 'dns' && form.dns">
              <FieldSection>
                <h2>{{ t('monitorEditor.dnsQuery') }}</h2>
                <p un-text="13px $muted">
                  {{ t('monitorEditor.validateResponseCodesAndRecordValues') }}
                </p>
                <FieldGroup>
                  <Field :label="t('monitorEditor.queryName')">
                    <input v-model="form.dns.name" required placeholder="example.com">
                  </Field><Field :label="t('monitorEditor.recordType')">
                    <Select v-model="form.dns.recordType">
                      <SelectTrigger><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectGroup>
                          <SelectItem
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
                            :value="record"
                          >
                            {{ record }}
                          </SelectItem>
                        </SelectGroup>
                      </SelectContent>
                    </Select>
                  </Field><Field
                    :label="t('monitorEditor.dnsServer')"
                    :hint="t('monitorEditor.leaveEmptyForSystemResolverOrEnterHost')"
                  >
                    <input v-model="form.dns.server" placeholder="1.1.1.1:53">
                  </Field><Field :label="t('monitorEditor.protocol')">
                    <Select v-model="form.dns.protocol">
                      <SelectTrigger><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectGroup>
                          <SelectItem value="udp">
                            UDP
                          </SelectItem>
                          <SelectItem value="tcp">
                            TCP
                          </SelectItem>
                        </SelectGroup>
                      </SelectContent>
                    </Select>
                  </Field><Field :label="t('monitorEditor.expectedResponseCode')">
                    <Select v-model="form.dns.expectedRCode">
                      <SelectTrigger><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectGroup>
                          <SelectItem
                            v-for="code in [
                              'NOERROR',
                              'FORMERR',
                              'SERVFAIL',
                              'NXDOMAIN',
                              'NOTIMP',
                              'REFUSED',
                            ]"
                            :key="code"
                            :value="code"
                          >
                            {{ code }}
                          </SelectItem>
                        </SelectGroup>
                      </SelectContent>
                    </Select>
                  </Field><Field :label="t('monitorEditor.recordMatching')">
                    <Select v-model="form.dns.matchMode">
                      <SelectTrigger><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectGroup>
                          <SelectItem value="contains">
                            {{ t('monitorEditor.containsExpectedRecords') }}
                          </SelectItem>
                          <SelectItem value="exact">
                            {{ t('monitorEditor.exactSet') }}
                          </SelectItem>
                        </SelectGroup>
                      </SelectContent>
                    </Select>
                  </Field><Field class="span-full" :label="t('monitorEditor.expectedValuesOnePerLine')">
                    <textarea v-model="dnsValues" />
                  </Field>
                </FieldGroup>
              </FieldSection>
            </template><template v-if="form.type === 'heartbeat' && form.heartbeat">
              <FieldSection>
                <h2>{{ t('monitorTypes.heartbeat') }}</h2>
                <p un-text="13px $muted">
                  {{ t('monitorEditor.servicesReportPeriodicallyMissingAPeriodPlusGrace') }}
                </p>
                <FieldGroup>
                  <Field :label="t('monitorEditor.expectedPeriodSeconds')">
                    <input
                      v-model.number="form.heartbeat.periodSeconds"
                      type="number"
                      min="30"
                      required
                    >
                  </Field><Field :label="t('monitorEditor.gracePeriodSeconds')">
                    <input
                      v-model.number="form.heartbeat.graceSeconds"
                      type="number"
                      min="0"
                      required
                    >
                  </Field>
                </FieldGroup>
                <Alert mt="5">
                  {{ t('monitorEditor.afterSavingGenerateOrRotateTheReportToken') }}
                </Alert>
              </FieldSection>
            </template><template v-if="form.type === 'certificate' && form.certificate">
              <FieldSection>
                <h2>{{ t('monitorEditor.certificateExpiry') }}</h2>
                <p un-text="13px $muted">
                  {{ t('monitorEditor.certificateRiskIsDisplayedSeparatelyAndExcludedFrom') }}
                </p>
                <FieldGroup>
                  <Field :label="t('monitorEditor.tlsHost')">
                    <input v-model="form.certificate.host" required>
                  </Field><Field :label="t('common.port')">
                    <input
                      v-model.number="form.certificate.port"
                      type="number"
                      min="1"
                      max="65535"
                    >
                  </Field><Field
                    class="span-full"
                    :label="t('common.warningThresholdsDays')"
                    :hint="t('monitorEditor.commaSeparatedDefaults301471Days')"
                  >
                    <input v-model="warningDays">
                  </Field>
                </FieldGroup>
              </FieldSection>
              <FieldSection as="details">
                <summary cursor="pointer" un-text="12px $link" font="600" py="15px" px="0">
                  {{ t('common.tlsConnectionSettings') }}
                </summary>
                <TLSFields v-model="form.certificate.tls" :secrets="secrets" />
                <Separator />
                <ConnectionFields
                  v-model="form.certificate.connection"
                  :secrets="secrets"
                />
              </FieldSection>
            </template>
          </TabsContent><TabsContent v-if="form.type !== 'heartbeat'" value="schedule">
            <FieldSection>
              <h2>{{ t('monitorEditor.checkSchedule') }}</h2>
              <p un-text="13px $muted">
                {{ t('monitorEditor.checksRunOnAFixedCadenceWithoutOverlap') }}
              </p>
              <FieldGroup>
                <Field
                  :label="t('monitorEditor.checkIntervalSeconds')"
                  :hint="
                    form.type === 'certificate'
                      ? t('monitorEditor.certificatesDefaultToDailyChecks86400Seconds')
                      : t('monitorEditor.minimum30Seconds')
                  "
                >
                  <input
                    v-model.number="form.intervalSeconds"
                    type="number"
                    min="30"
                    max="2592000"
                    required
                  >
                </Field><Field :label="t('monitorEditor.attemptTimeoutSeconds')">
                  <input
                    v-model.number="form.timeoutSeconds"
                    type="number"
                    min="1"
                    :max="form.intervalSeconds"
                    required
                  >
                </Field><template v-if="['http', 'tcp', 'dns'].includes(form.type)">
                  <Field
                    :label="t('monitorEditor.additionalRetries')"
                    :hint="t('monitorEditor.default2ZeroMeansTheFirstAttemptOnly')"
                  >
                    <input v-model.number="form.retries" type="number" min="0" max="10">
                  </Field><Field :label="t('monitorEditor.retryDelaySeconds')">
                    <input
                      v-model.number="form.retryDelaySeconds"
                      type="number"
                      min="0"
                      :max="form.intervalSeconds"
                    >
                  </Field><Field :label="t('monitorEditor.consecutiveFailedRounds')">
                    <input
                      v-model.number="form.failureThreshold"
                      type="number"
                      min="1"
                      max="100"
                    >
                  </Field><Field :label="t('monitorEditor.consecutiveSuccessfulRounds')">
                    <input
                      v-model.number="form.recoveryThreshold"
                      type="number"
                      min="1"
                      max="100"
                    >
                  </Field>
                </template>
              </FieldGroup>
              <Alert mt="6" as="p">
                {{ t('monitorEditor.retryCountIsALimitInsufficientBudgetEnds') }}
              </Alert>
            </FieldSection>
          </TabsContent><TabsContent value="notifications">
            <FieldSection>
              <h2>{{ t('monitorEditor.notificationChannels') }}</h2>
              <p un-text="13px $muted">
                {{
                  t(
                    'monitorEditor.selectAdministratorConfiguredChannelsConfirmedFailuresCreateDurable',
                  )
                }}
              </p>
              <div flex="~ gap-12px wrap">
                <label v-for="channel in channels" :key="channel.id" flex="~ items-center gap-8px" un-text="12px $text"><input
                  v-model="form.notificationChannelIds"
                  type="checkbox"
                  :value="channel.id"
                >{{ channel.name
                }}<span v-if="!channel.enabled" un-text="13px $muted">{{
                  t('monitorEditor.disabled')
                }}</span></label>
              </div>
              <p v-if="!channels.length" un-text="13px $muted">
                {{ t('monitorEditor.noChannelsYetAnAdministratorCanCreateOne') }}
              </p>
              <Separator />
              <Toggle
                v-if="form.type === 'certificate' && form.certificate"
                v-model="form.certificate.notifyRenewal"
                :label="t('monitorEditor.notifyOnCertificateRenewal')"
                mb="5"
              /><Toggle
                v-else
                v-model="form.notifyRecovery"
                :label="t('monitorEditor.notifyOnRecovery')"
                mb="5"
              /><Field
                v-if="form.type !== 'certificate'"
                :label="t('monitorEditor.repeatedOutageReminderSeconds')"
                :hint="t('monitorEditor.0DisablesRepeatedRemindersMinimum30SecondsWhen')"
              >
                <input v-model.number="form.reminderSeconds" type="number" min="0">
              </Field>
            </FieldSection>
          </TabsContent>
        </TabsRoot>
      </Card>
      <FieldActions>
        <Button as-child>
          <RouterLink to="/app/monitors">
            {{ t('common.cancel') }}
          </RouterLink>
        </Button><Button type="submit" :disabled="saving || !canEdit()" variant="primary">
          <span class="i-lucide-save" w="15px" h="15px" aria-hidden="true" />{{ t('common.saveMonitor') }}
        </Button>
      </FieldActions>
    </form>
  </AsyncState>
</template>
