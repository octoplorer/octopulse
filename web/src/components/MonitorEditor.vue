<script setup lang="ts">
import type { Channel, Secret } from '../client/types.gen'
import type { MonitorForm } from '../lib/monitor-form'
import { useMutation, useQueryCache } from '@pinia/colada'
import { useForm } from '@tanstack/vue-form'
import { computed, onMounted, ref, watch } from 'vue'
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

import { clone } from '../lib/form'
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
import JSONInput from './JSONInput.vue'
import KeyValues from './KeyValues.vue'
import ListInput from './ListInput.vue'
import PageHeader from './PageHeader.vue'
import SecretSelect from './SecretSelect.vue'
import TagsInput from './TagsInput.vue'
import TLSFields from './TLSFields.vue'
import Toggle from './Toggle.vue'
import { Alert } from './ui/alert'
import { Button } from './ui/button'
import { Card } from './ui/card'
import { FieldActions, FieldDescription, FieldGroup, FieldInput, FieldLabel, FieldSection, FieldTextarea } from './ui/field'
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
const error = ref('')
const channels = ref<Channel[]>([])
const secrets = ref<Secret[]>([])
const tagsInput = ref<InstanceType<typeof TagsInput>>()
const headerError = ref('')
const jsonError = ref('')
const formApi = useForm({
  defaultValues: newMonitor(),
  validators: {
    onSubmit: ({ value }) => {
      if (value.type === 'http' && (headerError.value || jsonError.value))
        return headerError.value || jsonError.value
      try {
        monitorPayload(value)
      }
      catch (e) {
        return errorText(e)
      }
    },
  },
  onSubmit: async ({ value }) => {
    error.value = ''
    try {
      const payload = monitorPayload(value)
      const result = await (editing.value
        ? updateMonitor.mutateAsync({ path: { id: value.id }, body: payload })
        : createMonitor.mutateAsync({ body: payload }))
      notify(t('monitorEditor.monitorSaved'))
      await router.push(`/app/monitors/${result.id}`)
    }
    catch (e) {
      error.value = errorText(e)
    }
  },
})
const form = formApi.useSelector(state => state.values)
const saving = formApi.useSelector(state => state.isSubmitting)
const validationError = formApi.useSelector(state => state.errors.filter(Boolean).join(', '))
const activeTab = ref('target')
watch(
  () => form.value.type,
  (type, previous) => {
    if (type === 'heartbeat' && activeTab.value === 'schedule')
      activeTab.value = 'target'
    if (
      !editing.value
      && previous === 'certificate'
      && type !== 'certificate'
      && form.value.intervalSeconds === 86400
    ) {
      formApi.setFieldValue('intervalSeconds', 60)
    }
    if (type === 'http' && !form.value.http)
      formApi.setFieldValue('http', emptyHTTP())
    if (type === 'tcp' && !form.value.tcp)
      formApi.setFieldValue('tcp', emptyTCP())
    if (type === 'dns' && !form.value.dns)
      formApi.setFieldValue('dns', emptyDNS())
    if (type === 'heartbeat' && !form.value.heartbeat)
      formApi.setFieldValue('heartbeat', emptyHeartbeat())
    if (type === 'certificate' && !form.value.certificate) {
      formApi.setFieldValue('certificate', emptyCertificate())
      if (!editing.value)
        formApi.setFieldValue('intervalSeconds', 86400)
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
      formApi.reset(toMonitorForm(monitor.data))
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
function monitorPayload(value: MonitorForm): MonitorForm {
  const payload = clone(value)
  for (const type of ['http', 'tcp', 'dns', 'heartbeat', 'certificate'] as const) {
    if (type !== payload.type)
      delete payload[type]
  }
  if (new Blob([JSON.stringify(payload)]).size > 8 * 1024 * 1024)
    throw new Error(t('monitorEditor.theMonitorConfigurationRequestMayNotExceed8'))
  return payload
}
function parseStatusRange(value: string) {
  const [min, max] = value.split('-').map(Number)
  return { min: min!, max: max ?? min! }
}
function formatStatusRange(range: { min: number, max: number }) {
  return `${range.min}-${range.max}`
}
function save() {
  error.value = ''
  tagsInput.value?.commit()
  return formApi.handleSubmit()
}
async function addFile(event: Event) {
  const files = (event.target as HTMLInputElement).files
  if (!files || !form.value.http)
    return
  for (const file of Array.from(files)) {
    const existingBytes = form.value.http.body.files.reduce(
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
    formApi.setFieldValue('http.body.files', files => [...(files || []), {
      field: 'file',
      filename: file.name,
      contentType: file.type || 'application/octet-stream',
      base64,
    }])
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
      <Alert v-if="error || validationError" role="alert" variant="validation">
        {{ error || validationError }}
      </Alert>
      <Card mb="6" as="section">
        <FieldSection>
          <h2>{{ t('monitorEditor.basicInformation') }}</h2>
          <p un-text="13px subtle">
            {{ t('monitorEditor.useARecognizableNameAndGroupRelatedServices') }}
          </p>
          <FieldGroup>
            <Field :label="t('common.displayName')">
              <formApi.Field v-slot="{ field }" name="name">
                <FieldInput
                  :model-value="field.state.value" required maxlength="200"
                  :placeholder="t('monitorEditor.namePlaceholder')"
                  @update:model-value="field.handleChange"
                  @blur="field.handleBlur"
                />
              </formApi.Field>
            </Field><Field
              :label="t('monitorEditor.monitorType')"
              :hint="
                editing ? t('monitorEditor.theMonitorTypeIsFixedAfterCreationCreate') : undefined
              "
            >
              <formApi.Field v-slot="{ field }" name="type">
                <Select :model-value="field.state.value" :disabled="editing" @update:model-value="field.handleChange" @focusout="field.handleBlur">
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectGroup>
                      <SelectItem v-for="type in monitorTypes" :key="type.value" :value="type.value">
                        {{ t(type.label) }}
                      </SelectItem>
                    </SelectGroup>
                  </SelectContent>
                </Select>
              </formApi.Field>
            </Field><Field :label="t('common.group')">
              <formApi.Field v-slot="{ field }" name="group">
                <FieldInput :model-value="field.state.value" :placeholder="t('monitorEditor.eGProduction')" @update:model-value="field.handleChange" @blur="field.handleBlur" />
              </formApi.Field>
            </Field><Field :label="t('monitorEditor.tags')" :hint="t('monitorEditor.separateWithCommas')">
              <formApi.Field v-slot="{ field }" name="tags">
                <TagsInput ref="tagsInput" :model-value="field.state.value" :name="field.name" placeholder="production, api" @update:model-value="field.handleChange" @focusout="field.handleBlur" />
              </formApi.Field>
            </Field><Field class="span-full" :label="t('monitorEditor.description')">
              <formApi.Field v-slot="{ field }" name="description">
                <FieldTextarea :model-value="field.state.value" rows="2" @update:model-value="field.handleChange" @blur="field.handleBlur" />
              </formApi.Field>
            </Field>
            <div class="span-full">
              <formApi.Field v-slot="{ field }" name="enabled">
                <Toggle
                  :model-value="field.state.value" :label="t('monitorEditor.enableMonitor')" :description="t('monitorEditor.pausingStopsCollectionAndExcludesPausedTimeFrom')"
                  @update:model-value="field.handleChange"
                  @focusout="field.handleBlur"
                />
              </formApi.Field>
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
                <p un-text="13px subtle">
                  {{ t('monitorEditor.allRequestMethodsUseTheSameRoundRetry') }}
                </p>
                <FieldGroup>
                  <Field :label="t('monitorEditor.targetUrl')" class="span-full">
                    <formApi.Field v-slot="{ field }" name="http.url">
                      <FieldInput
                        :model-value="field.state.value" type="url" required
                        placeholder="https://api.example.com/health"
                        @update:model-value="field.handleChange"
                        @blur="field.handleBlur"
                      />
                    </formApi.Field>
                  </Field><Field :label="t('monitorEditor.requestMethod')">
                    <formApi.Field v-slot="{ field }" name="http.method">
                      <FieldInput :model-value="field.state.value" list="http-methods" required @update:model-value="field.handleChange" @blur="field.handleBlur" />
                    </formApi.Field><datalist
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
                    <formApi.Field v-slot="{ field }" name="http.host">
                      <FieldInput :model-value="field.state.value" placeholder="api.example.com" @update:model-value="field.handleChange" @blur="field.handleBlur" />
                    </formApi.Field>
                  </Field>
                  <div class="span-full">
                    <FieldLabel as="label">
                      {{ t('monitorEditor.queryParameters') }}
                    </FieldLabel><formApi.Field v-slot="{ field }" name="http.query">
                      <KeyValues :model-value="field.state.value!" :secrets="secrets" mt="2" @update:model-value="field.handleChange" @focusout="field.handleBlur" />
                    </formApi.Field>
                  </div>
                  <div class="span-full">
                    <FieldLabel as="label">
                      {{
                        t('monitorEditor.requestHeadersRepeatedNamesSupported')
                      }}
                    </FieldLabel><formApi.Field v-slot="{ field }" name="http.headers">
                      <KeyValues :model-value="field.state.value!" :secrets="secrets" mt="2" @update:model-value="field.handleChange" @focusout="field.handleBlur" />
                    </formApi.Field>
                  </div>
                </FieldGroup>
              </FieldSection>
              <FieldSection>
                <h2>{{ t('monitorEditor.requestBody') }}</h2>
                <p un-text="13px subtle">
                  {{ t('monitorEditor.configureFormatCharacterEncodingAndCompressionSeparately') }}
                </p>
                <FieldGroup>
                  <Field :label="t('monitorEditor.bodyFormat')">
                    <formApi.Field v-slot="{ field }" name="http.body.format">
                      <Select :model-value="field.state.value" @update:model-value="field.handleChange" @focusout="field.handleBlur">
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
                    </formApi.Field>
                  </Field><Field
                    v-if="!['none', 'json'].includes(form.http.body.format)"
                    :label="t('common.characterEncoding')"
                  >
                    <formApi.Field v-slot="{ field }" name="http.body.charset">
                      <Select :model-value="field.state.value" @update:model-value="field.handleChange" @focusout="field.handleBlur">
                        <SelectTrigger><SelectValue /></SelectTrigger>
                        <SelectContent>
                          <SelectGroup>
                            <SelectItem v-for="charset in charsetOptions" :key="charset" :value="charset">
                              {{ charset }}
                            </SelectItem>
                          </SelectGroup>
                        </SelectContent>
                      </Select>
                    </formApi.Field>
                  </Field><Field
                    v-if="['json', 'text', 'raw'].includes(form.http.body.format)"
                    :label="t('monitorEditor.bodySecretReference')"
                  >
                    <formApi.Field v-slot="{ field }" name="http.body.secretRef">
                      <SecretSelect
                        :model-value="field.state.value" :secrets="secrets" optional
                        @update:model-value="field.handleChange"
                        @focusout="field.handleBlur"
                      />
                    </formApi.Field>
                  </Field><Field
                    v-if="['text', 'raw'].includes(form.http.body.format)"
                    :label="t('monitorEditor.contentType')"
                  >
                    <formApi.Field v-slot="{ field }" name="http.body.contentType">
                      <FieldInput :model-value="field.state.value" placeholder="text/plain" @update:model-value="field.handleChange" @blur="field.handleBlur" />
                    </formApi.Field>
                  </Field><Field
                    v-if="
                      ['json', 'text'].includes(form.http.body.format) && !form.http.body.secretRef
                    "
                    class="span-full"
                    :label="t('monitorEditor.bodyContent')"
                  >
                    <formApi.Field v-slot="{ field }" name="http.body.text">
                      <FieldTextarea
                        :model-value="field.state.value" rows="7" :placeholder="form.http.body.format === 'json' ? '{}' : ''"
                        spellcheck="false"
                        @update:model-value="field.handleChange"
                        @blur="field.handleBlur"
                      />
                    </formApi.Field>
                  </Field><Field
                    v-if="form.http.body.format === 'raw' && !form.http.body.secretRef"
                    class="span-full"
                    :label="t('monitorEditor.rawBytesBase64')"
                  >
                    <formApi.Field v-slot="{ field }" name="http.body.base64">
                      <FieldTextarea :model-value="field.state.value" spellcheck="false" @update:model-value="field.handleChange" @blur="field.handleBlur" />
                    </formApi.Field>
                  </Field>
                  <div
                    v-if="['form', 'multipart'].includes(form.http.body.format)"
                    class="span-full"
                  >
                    <FieldLabel as="label">
                      {{ t('monitorEditor.formFields') }}
                    </FieldLabel><formApi.Field v-slot="{ field }" name="http.body.fields">
                      <KeyValues :model-value="field.state.value!" :secrets="secrets" mt="2" @update:model-value="field.handleChange" @focusout="field.handleBlur" />
                    </formApi.Field>
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
                      <formApi.Field v-slot="{ field }" :name="`http.body.files[${index}].field`">
                        <FieldInput
                          :model-value="field.state.value" :aria-label="t('monitorEditor.fieldName')" placeholder="file"
                          @update:model-value="field.handleChange"
                          @blur="field.handleBlur"
                        />
                      </formApi.Field><span>{{ file.filename }}</span><Button
                        type="button"
                        :aria-label="t('monitorEditor.removeFile')"
                        shape="square"
                        @click="formApi.setFieldValue('http.body.files', files => files?.filter((_, i) => i !== index))"
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
                <p un-text="13px subtle">
                  {{ t('monitorEditor.credentialsUseExistingSecretReferences') }}
                </p>
                <FieldGroup>
                  <Field :label="t('monitorEditor.authenticationType')">
                    <formApi.Field v-slot="{ field }" name="http.auth.type">
                      <Select :model-value="field.state.value" @update:model-value="field.handleChange" @focusout="field.handleBlur">
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
                    </formApi.Field>
                  </Field><Field
                    v-if="form.http.auth.type !== 'none'"
                    :label="t('monitorEditor.passwordTokenSecret')"
                  >
                    <formApi.Field v-slot="{ field }" name="http.auth.secretRef">
                      <SecretSelect :model-value="field.state.value" :secrets="secrets" @update:model-value="field.handleChange" @focusout="field.handleBlur" />
                    </formApi.Field>
                  </Field><Field v-if="form.http.auth.type === 'basic'" :label="t('common.username')">
                    <formApi.Field v-slot="{ field }" name="http.auth.username">
                      <FieldInput :model-value="field.state.value" @update:model-value="field.handleChange" @blur="field.handleBlur" />
                    </formApi.Field>
                  </Field><Field
                    v-if="form.http.auth.type === 'basic'"
                    :label="t('monitorEditor.usernameSecretOptional')"
                  >
                    <formApi.Field v-slot="{ field }" name="http.auth.usernameSecretRef">
                      <SecretSelect
                        :model-value="field.state.value" :secrets="secrets" optional
                        @update:model-value="field.handleChange"
                        @focusout="field.handleBlur"
                      />
                    </formApi.Field>
                  </Field><Field
                    v-if="form.http.auth.type === 'header'"
                    :label="t('monitorEditor.authenticationHeader')"
                  >
                    <formApi.Field v-slot="{ field }" name="http.auth.header">
                      <FieldInput :model-value="field.state.value" @update:model-value="field.handleChange" @blur="field.handleBlur" />
                    </formApi.Field>
                  </Field><Field
                    v-if="form.http.auth.type === 'header'"
                    :label="t('monitorEditor.valuePrefix')"
                  >
                    <formApi.Field v-slot="{ field }" name="http.auth.prefix">
                      <FieldInput :model-value="field.state.value" placeholder="Bearer " @update:model-value="field.handleChange" @blur="field.handleBlur" />
                    </formApi.Field>
                  </Field>
                </FieldGroup>
              </FieldSection>
              <FieldSection>
                <h2>{{ t('monitorEditor.successAssertions') }}</h2>
                <p un-text="13px subtle">
                  {{ t('monitorEditor.everyConfiguredAssertionMustPassForASuccessful') }}
                </p>
                <FieldGroup>
                  <Field
                    :label="t('monitorEditor.statusCodes')"
                    :hint="t('monitorEditor.commaSeparatedAcceptedTogetherWithRanges')"
                  >
                    <formApi.Field v-slot="{ field }" name="http.assertions.statusCodes">
                      <ListInput :model-value="field.state.value!" :parse-item="Number" placeholder="200, 204" @update:model-value="field.handleChange" @blur="field.handleBlur" />
                    </formApi.Field>
                  </Field><Field :label="t('monitorEditor.statusRanges')">
                    <formApi.Field v-slot="{ field }" name="http.assertions.statusRanges">
                      <ListInput :model-value="field.state.value!" :parse-item="parseStatusRange" :format-item="formatStatusRange" placeholder="200-299, 300-399" @update:model-value="field.handleChange" @blur="field.handleBlur" />
                    </formApi.Field>
                  </Field><Field :label="t('monitorEditor.textContainsOnePerLine')">
                    <formApi.Field v-slot="{ field }" name="http.assertions.textContains">
                      <ListInput :model-value="field.state.value!" :parse-item="String" separator="lines" :trim="false" multiline @update:model-value="field.handleChange" @blur="field.handleBlur" />
                    </formApi.Field>
                  </Field><Field :label="t('monitorEditor.textExcludesOnePerLine')">
                    <formApi.Field v-slot="{ field }" name="http.assertions.textNotContains">
                      <ListInput :model-value="field.state.value!" :parse-item="String" separator="lines" :trim="false" multiline @update:model-value="field.handleChange" @blur="field.handleBlur" />
                    </formApi.Field>
                  </Field><Field :label="t('monitorEditor.bodyRegexOnePerLine')">
                    <formApi.Field v-slot="{ field }" name="http.assertions.regex">
                      <ListInput :model-value="field.state.value!" :parse-item="String" separator="lines" :trim="false" multiline spellcheck="false" @update:model-value="field.handleChange" @blur="field.handleBlur" />
                    </formApi.Field>
                  </Field><Field
                    :label="t('monitorEditor.maximumResponseTimeMs')"
                    :hint="t('monitorEditor.0DisablesThisAssertion')"
                  >
                    <formApi.Field v-slot="{ field }" name="http.assertions.maxLatencyMs">
                      <FieldInput
                        :model-value="field.state.value" type="number" min="0"
                        @input="field.handleChange(($event.target as HTMLInputElement).valueAsNumber)"
                        @blur="field.handleBlur"
                      />
                    </formApi.Field>
                  </Field><Field
                    class="span-full"
                    :label="t('monitorEditor.responseHeaderAssertions')"
                    :error="headerError"
                    :hint="headerHint"
                  >
                    <formApi.Field v-slot="{ field }" name="http.assertions.headers">
                      <JSONInput :model-value="field.state.value!" :label="t('monitorEditor.headerAssertions')" rows="4" spellcheck="false" @update:model-value="field.handleChange" @validity-change="headerError = $event" @blur="field.handleBlur" />
                    </formApi.Field>
                  </Field><Field
                    class="span-full"
                    :label="t('monitorEditor.jsonFieldAssertionsJsonPointer')"
                    :error="jsonError"
                    :hint="jsonHint"
                  >
                    <formApi.Field v-slot="{ field }" name="http.assertions.json">
                      <JSONInput :model-value="field.state.value!" :label="t('monitorEditor.jsonAssertions')" rows="4" spellcheck="false" @update:model-value="field.handleChange" @validity-change="jsonError = $event" @blur="field.handleBlur" />
                    </formApi.Field>
                  </Field>
                </FieldGroup>
              </FieldSection>
              <FieldSection as="details">
                <summary cursor="pointer" un-text="12px link" font="600" py="15px" px="0">
                  {{ t('monitorEditor.advancedTlsConnectionsTransport') }}
                </summary>
                <h3 mb="4">
                  TLS
                </h3>
                <formApi.Field v-slot="{ field }" name="http.tls">
                  <TLSFields :model-value="field.state.value!" :secrets="secrets" @update:model-value="field.handleChange" @focusout="field.handleBlur" />
                </formApi.Field>
                <Separator />
                <h3 mb="4">
                  {{ t('monitorEditor.networkConnection') }}
                </h3>
                <formApi.Field v-slot="{ field }" name="http.connection">
                  <ConnectionFields :model-value="field.state.value!" :secrets="secrets" @update:model-value="field.handleChange" @focusout="field.handleBlur" />
                </formApi.Field>
                <Separator />
                <FieldGroup>
                  <div class="span-full">
                    <formApi.Field v-slot="{ field }" name="http.redirects.enabled">
                      <Toggle
                        :model-value="field.state.value" :label="t('monitorEditor.followRedirects')" @update:model-value="field.handleChange"
                        @focusout="field.handleBlur"
                      />
                    </formApi.Field>
                  </div>
                  <Field :label="t('monitorEditor.maximumRedirects')">
                    <formApi.Field v-slot="{ field }" name="http.redirects.maxHops">
                      <FieldInput
                        :model-value="field.state.value" type="number" min="1"
                        max="20"
                        @input="field.handleChange(($event.target as HTMLInputElement).valueAsNumber)"
                        @blur="field.handleBlur"
                      />
                    </formApi.Field>
                  </Field><Field :label="t('monitorEditor.redirectScope')">
                    <formApi.Field v-slot="{ field }" name="http.redirects.scope">
                      <Select :model-value="field.state.value" @update:model-value="field.handleChange" @focusout="field.handleBlur">
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
                    </formApi.Field>
                  </Field><Field :label="t('monitorEditor.acceptEncoding')">
                    <formApi.Field v-slot="{ field }" name="http.acceptEncoding">
                      <Select :model-value="field.state.value" @update:model-value="field.handleChange" @focusout="field.handleBlur">
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
                    </formApi.Field>
                  </Field><Field :label="t('monitorEditor.responseCharset')">
                    <formApi.Field v-slot="{ field }" name="http.responseCharset">
                      <Select :model-value="field.state.value" @update:model-value="field.handleChange" @focusout="field.handleBlur">
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
                    </formApi.Field>
                  </Field><Field :label="t('monitorEditor.decompressedResponseLimitBytes')">
                    <formApi.Field v-slot="{ field }" name="http.maxResponseBytes">
                      <FieldInput
                        :model-value="field.state.value" type="number" min="1"
                        max="16777216"
                        @input="field.handleChange(($event.target as HTMLInputElement).valueAsNumber)"
                        @blur="field.handleBlur"
                      />
                    </formApi.Field>
                  </Field>
                  <div>
                    <formApi.Field v-slot="{ field }" name="http.requestGzip">
                      <Toggle
                        :model-value="field.state.value" :label="t('monitorEditor.gzipRequestBody')" @update:model-value="field.handleChange"
                        @focusout="field.handleBlur"
                      />
                    </formApi.Field>
                  </div>
                </FieldGroup>
              </FieldSection>
            </template><template v-if="form.type === 'tcp' && form.tcp">
              <FieldSection>
                <h2>{{ t('monitorTypes.tcpConnection') }}</h2>
                <p un-text="13px subtle">
                  {{ t('monitorEditor.testAConnectionOrSendAPayloadAnd') }}
                </p>
                <FieldGroup>
                  <Field :label="t('monitorEditor.host')">
                    <formApi.Field v-slot="{ field }" name="tcp.host">
                      <FieldInput :model-value="field.state.value" required placeholder="db.example.com" @update:model-value="field.handleChange" @blur="field.handleBlur" />
                    </formApi.Field>
                  </Field><Field :label="t('common.port')">
                    <formApi.Field v-slot="{ field }" name="tcp.port">
                      <FieldInput
                        :model-value="field.state.value" type="number" min="1"
                        max="65535"
                        required
                        @input="field.handleChange(($event.target as HTMLInputElement).valueAsNumber)"
                        @blur="field.handleBlur"
                      />
                    </formApi.Field>
                  </Field><Field :label="t('monitorEditor.sendText')">
                    <formApi.Field v-slot="{ field }" name="tcp.sendText">
                      <FieldTextarea :model-value="field.state.value" @update:model-value="field.handleChange" @blur="field.handleBlur" />
                    </formApi.Field>
                  </Field><Field :label="t('monitorEditor.sendBytesBase64')">
                    <formApi.Field v-slot="{ field }" name="tcp.sendBase64">
                      <FieldTextarea :model-value="field.state.value" spellcheck="false" @update:model-value="field.handleChange" @blur="field.handleBlur" />
                    </formApi.Field>
                  </Field><Field :label="t('monitorEditor.payloadSecret')">
                    <formApi.Field v-slot="{ field }" name="tcp.sendSecretRef">
                      <SecretSelect
                        :model-value="field.state.value" :secrets="secrets" optional
                        @update:model-value="field.handleChange"
                        @focusout="field.handleBlur"
                      />
                    </formApi.Field>
                  </Field><Field :label="t('common.characterEncoding')">
                    <formApi.Field v-slot="{ field }" name="tcp.charset">
                      <Select :model-value="field.state.value" @update:model-value="field.handleChange" @focusout="field.handleBlur">
                        <SelectTrigger><SelectValue /></SelectTrigger>
                        <SelectContent>
                          <SelectGroup>
                            <SelectItem v-for="charset in charsetOptions" :key="charset" :value="charset">
                              {{ charset }}
                            </SelectItem>
                          </SelectGroup>
                        </SelectContent>
                      </Select>
                    </formApi.Field>
                  </Field><Field :label="t('monitorEditor.responseContains')">
                    <formApi.Field v-slot="{ field }" name="tcp.receiveContains">
                      <FieldInput :model-value="field.state.value" @update:model-value="field.handleChange" @blur="field.handleBlur" />
                    </formApi.Field>
                  </Field><Field :label="t('monitorEditor.responseRegex')">
                    <formApi.Field v-slot="{ field }" name="tcp.receiveRegex">
                      <FieldInput :model-value="field.state.value" @update:model-value="field.handleChange" @blur="field.handleBlur" />
                    </formApi.Field>
                  </Field><Field :label="t('monitorEditor.receiveSizeLimitBytes')">
                    <formApi.Field v-slot="{ field }" name="tcp.maxReceiveBytes">
                      <FieldInput
                        :model-value="field.state.value" type="number" min="1"
                        max="16777216"
                        @input="field.handleChange(($event.target as HTMLInputElement).valueAsNumber)"
                        @blur="field.handleBlur"
                      />
                    </formApi.Field>
                  </Field>
                </FieldGroup>
              </FieldSection>
              <FieldSection as="details">
                <summary cursor="pointer" un-text="12px link" font="600" py="15px" px="0">
                  {{ t('common.tlsConnectionSettings') }}
                </summary>
                <formApi.Field v-slot="{ field }" name="tcp.tls">
                  <TLSFields :model-value="field.state.value!" :secrets="secrets" allow-toggle @update:model-value="field.handleChange" @focusout="field.handleBlur" />
                </formApi.Field>
                <Separator />
                <formApi.Field v-slot="{ field }" name="tcp.connection">
                  <ConnectionFields
                    :model-value="field.state.value!" :secrets="secrets" @update:model-value="field.handleChange"
                    @focusout="field.handleBlur"
                  />
                </formApi.Field>
              </FieldSection>
            </template><template v-if="form.type === 'dns' && form.dns">
              <FieldSection>
                <h2>{{ t('monitorEditor.dnsQuery') }}</h2>
                <p un-text="13px subtle">
                  {{ t('monitorEditor.validateResponseCodesAndRecordValues') }}
                </p>
                <FieldGroup>
                  <Field :label="t('monitorEditor.queryName')">
                    <formApi.Field v-slot="{ field }" name="dns.name">
                      <FieldInput :model-value="field.state.value" required placeholder="example.com" @update:model-value="field.handleChange" @blur="field.handleBlur" />
                    </formApi.Field>
                  </Field><Field :label="t('monitorEditor.recordType')">
                    <formApi.Field v-slot="{ field }" name="dns.recordType">
                      <Select :model-value="field.state.value" @update:model-value="field.handleChange" @focusout="field.handleBlur">
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
                    </formApi.Field>
                  </Field><Field
                    :label="t('monitorEditor.dnsServer')"
                    :hint="t('monitorEditor.leaveEmptyForSystemResolverOrEnterHost')"
                  >
                    <formApi.Field v-slot="{ field }" name="dns.server">
                      <FieldInput :model-value="field.state.value" placeholder="1.1.1.1:53" @update:model-value="field.handleChange" @blur="field.handleBlur" />
                    </formApi.Field>
                  </Field><Field :label="t('monitorEditor.protocol')">
                    <formApi.Field v-slot="{ field }" name="dns.protocol">
                      <Select :model-value="field.state.value" @update:model-value="field.handleChange" @focusout="field.handleBlur">
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
                    </formApi.Field>
                  </Field><Field :label="t('monitorEditor.expectedResponseCode')">
                    <formApi.Field v-slot="{ field }" name="dns.expectedRCode">
                      <Select :model-value="field.state.value" @update:model-value="field.handleChange" @focusout="field.handleBlur">
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
                    </formApi.Field>
                  </Field><Field :label="t('monitorEditor.recordMatching')">
                    <formApi.Field v-slot="{ field }" name="dns.matchMode">
                      <Select :model-value="field.state.value" @update:model-value="field.handleChange" @focusout="field.handleBlur">
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
                    </formApi.Field>
                  </Field><Field class="span-full" :label="t('monitorEditor.expectedValuesOnePerLine')">
                    <formApi.Field v-slot="{ field }" name="dns.expectedValues">
                      <ListInput :model-value="field.state.value!" :parse-item="String" separator="lines" trim multiline @update:model-value="field.handleChange" @blur="field.handleBlur" />
                    </formApi.Field>
                  </Field>
                </FieldGroup>
              </FieldSection>
            </template><template v-if="form.type === 'heartbeat' && form.heartbeat">
              <FieldSection>
                <h2>{{ t('monitorTypes.heartbeat') }}</h2>
                <p un-text="13px subtle">
                  {{ t('monitorEditor.servicesReportPeriodicallyMissingAPeriodPlusGrace') }}
                </p>
                <FieldGroup>
                  <Field :label="t('monitorEditor.expectedPeriodSeconds')">
                    <formApi.Field v-slot="{ field }" name="heartbeat.periodSeconds">
                      <FieldInput
                        :model-value="field.state.value" type="number" min="30"
                        required
                        @input="field.handleChange(($event.target as HTMLInputElement).valueAsNumber)"
                        @blur="field.handleBlur"
                      />
                    </formApi.Field>
                  </Field><Field :label="t('monitorEditor.gracePeriodSeconds')">
                    <formApi.Field v-slot="{ field }" name="heartbeat.graceSeconds">
                      <FieldInput
                        :model-value="field.state.value" type="number" min="0"
                        required
                        @input="field.handleChange(($event.target as HTMLInputElement).valueAsNumber)"
                        @blur="field.handleBlur"
                      />
                    </formApi.Field>
                  </Field>
                </FieldGroup>
                <Alert mt="5">
                  {{ t('monitorEditor.afterSavingGenerateOrRotateTheReportToken') }}
                </Alert>
              </FieldSection>
            </template><template v-if="form.type === 'certificate' && form.certificate">
              <FieldSection>
                <h2>{{ t('monitorEditor.certificateExpiry') }}</h2>
                <p un-text="13px subtle">
                  {{ t('monitorEditor.certificateRiskIsDisplayedSeparatelyAndExcludedFrom') }}
                </p>
                <FieldGroup>
                  <Field :label="t('monitorEditor.tlsHost')">
                    <formApi.Field v-slot="{ field }" name="certificate.host">
                      <FieldInput :model-value="field.state.value" required @update:model-value="field.handleChange" @blur="field.handleBlur" />
                    </formApi.Field>
                  </Field><Field :label="t('common.port')">
                    <formApi.Field v-slot="{ field }" name="certificate.port">
                      <FieldInput
                        :model-value="field.state.value" type="number" min="1"
                        max="65535"
                        @input="field.handleChange(($event.target as HTMLInputElement).valueAsNumber)"
                        @blur="field.handleBlur"
                      />
                    </formApi.Field>
                  </Field><Field
                    class="span-full"
                    :label="t('common.warningThresholdsDays')"
                    :hint="t('monitorEditor.commaSeparatedDefaults301471Days')"
                  >
                    <formApi.Field v-slot="{ field }" name="certificate.warningDays">
                      <ListInput :model-value="field.state.value!" :parse-item="Number" @update:model-value="field.handleChange" @blur="field.handleBlur" />
                    </formApi.Field>
                  </Field>
                </FieldGroup>
              </FieldSection>
              <FieldSection as="details">
                <summary cursor="pointer" un-text="12px link" font="600" py="15px" px="0">
                  {{ t('common.tlsConnectionSettings') }}
                </summary>
                <formApi.Field v-slot="{ field }" name="certificate.tls">
                  <TLSFields :model-value="field.state.value!" :secrets="secrets" @update:model-value="field.handleChange" @focusout="field.handleBlur" />
                </formApi.Field>
                <Separator />
                <formApi.Field v-slot="{ field }" name="certificate.connection">
                  <ConnectionFields
                    :model-value="field.state.value!" :secrets="secrets" @update:model-value="field.handleChange"
                    @focusout="field.handleBlur"
                  />
                </formApi.Field>
              </FieldSection>
            </template>
          </TabsContent><TabsContent v-if="form.type !== 'heartbeat'" value="schedule">
            <FieldSection>
              <h2>{{ t('monitorEditor.checkSchedule') }}</h2>
              <p un-text="13px subtle">
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
                  <formApi.Field v-slot="{ field }" name="intervalSeconds">
                    <FieldInput
                      :model-value="field.state.value" type="number" min="30"
                      max="2592000"
                      required
                      @input="field.handleChange(($event.target as HTMLInputElement).valueAsNumber)"
                      @blur="field.handleBlur"
                    />
                  </formApi.Field>
                </Field><Field :label="t('monitorEditor.attemptTimeoutSeconds')">
                  <formApi.Field v-slot="{ field }" name="timeoutSeconds">
                    <FieldInput
                      :model-value="field.state.value" type="number" min="1"
                      :max="form.intervalSeconds"
                      required
                      @input="field.handleChange(($event.target as HTMLInputElement).valueAsNumber)"
                      @blur="field.handleBlur"
                    />
                  </formApi.Field>
                </Field><template v-if="['http', 'tcp', 'dns'].includes(form.type)">
                  <Field
                    :label="t('monitorEditor.additionalRetries')"
                    :hint="t('monitorEditor.default2ZeroMeansTheFirstAttemptOnly')"
                  >
                    <formApi.Field v-slot="{ field }" name="retries">
                      <FieldInput :model-value="field.state.value" type="number" min="0" max="10" @input="field.handleChange(($event.target as HTMLInputElement).valueAsNumber)" @blur="field.handleBlur" />
                    </formApi.Field>
                  </Field><Field :label="t('monitorEditor.retryDelaySeconds')">
                    <formApi.Field v-slot="{ field }" name="retryDelaySeconds">
                      <FieldInput
                        :model-value="field.state.value" type="number" min="0"
                        :max="form.intervalSeconds"
                        @input="field.handleChange(($event.target as HTMLInputElement).valueAsNumber)"
                        @blur="field.handleBlur"
                      />
                    </formApi.Field>
                  </Field><Field :label="t('monitorEditor.consecutiveFailedRounds')">
                    <formApi.Field v-slot="{ field }" name="failureThreshold">
                      <FieldInput
                        :model-value="field.state.value" type="number" min="1"
                        max="100"
                        @input="field.handleChange(($event.target as HTMLInputElement).valueAsNumber)"
                        @blur="field.handleBlur"
                      />
                    </formApi.Field>
                  </Field><Field :label="t('monitorEditor.consecutiveSuccessfulRounds')">
                    <formApi.Field v-slot="{ field }" name="recoveryThreshold">
                      <FieldInput
                        :model-value="field.state.value" type="number" min="1"
                        max="100"
                        @input="field.handleChange(($event.target as HTMLInputElement).valueAsNumber)"
                        @blur="field.handleBlur"
                      />
                    </formApi.Field>
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
              <p un-text="13px subtle">
                {{
                  t(
                    'monitorEditor.selectAdministratorConfiguredChannelsConfirmedFailuresCreateDurable',
                  )
                }}
              </p>
              <formApi.Field v-slot="{ field }" name="notificationChannelIds">
                <div flex="~ gap-12px wrap">
                  <label v-for="channel in channels" :key="channel.id" flex="~ items-center gap-8px" un-text="12px default"><input
                    :checked="field.state.value.includes(channel.id)" type="checkbox" :value="channel.id"
                    @change="field.handleChange(($event.target as HTMLInputElement).checked ? [...field.state.value, channel.id] : field.state.value.filter(id => id !== channel.id))"
                    @blur="field.handleBlur"
                  >{{ channel.name
                  }}<span v-if="!channel.enabled" un-text="13px subtle">{{
                    t('monitorEditor.disabled')
                  }}</span></label>
                </div>
              </formApi.Field>
              <p v-if="!channels.length" un-text="13px subtle">
                {{ t('monitorEditor.noChannelsYetAnAdministratorCanCreateOne') }}
              </p>
              <Separator />
              <formApi.Field v-if="form.type === 'certificate' && form.certificate" v-slot="{ field }" name="certificate.notifyRenewal">
                <Toggle

                  :model-value="field.state.value" :label="t('monitorEditor.notifyOnCertificateRenewal')" mb="5"
                  @update:model-value="field.handleChange"
                  @focusout="field.handleBlur"
                />
              </formApi.Field><formApi.Field v-else v-slot="{ field }" name="notifyRecovery">
                <Toggle

                  :model-value="field.state.value" :label="t('monitorEditor.notifyOnRecovery')" mb="5"
                  @update:model-value="field.handleChange"
                  @focusout="field.handleBlur"
                />
              </formApi.Field><Field
                v-if="form.type !== 'certificate'"
                :label="t('monitorEditor.repeatedOutageReminderSeconds')"
                :hint="t('monitorEditor.0DisablesRepeatedRemindersMinimum30SecondsWhen')"
              >
                <formApi.Field v-slot="{ field }" name="reminderSeconds">
                  <FieldInput :model-value="field.state.value" type="number" min="0" @input="field.handleChange(($event.target as HTMLInputElement).valueAsNumber)" @blur="field.handleBlur" />
                </formApi.Field>
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
