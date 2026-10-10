<script setup lang="ts">
import type { Channel, Secret } from '../client/types.gen'
import type { MonitorForm } from '../lib/monitor-form'
import { useMutation, useQueryCache } from '@pinia/colada'
import { useForm } from '@tanstack/vue-form'
import { computed, nextTick, onMounted, ref, watch } from 'vue'
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
import { useJSONInput, useListInput } from '../composables/form-inputs'
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
import { Separator } from './common/separator'
import ConnectionFields from './ConnectionFields.vue'
import KeyValues from './KeyValues.vue'
import TLSFields from './TLSFields.vue'
import { Banner } from './ui/banner'
import { PageHeader } from './ui/blocks/page-header'
import { Button } from './ui/button'
import { CheckboxGroup, CheckboxItem } from './ui/checkbox'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from './ui/collapsible'
import { Field, FieldActions, FieldDescription, FieldGroup, FieldLabel, FieldSection } from './ui/field'
import { Input, InputArea } from './ui/input'
import { LayerCard, LayerCardPrimary } from './ui/layer-card'
import { Loader } from './ui/loader'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from './ui/select'
import { Switch } from './ui/switch'
import { TabsContent, TabsList, TabsRoot, TabsTrigger } from './ui/tabs'
import { TagInput } from './ui/tag-input'

const { t } = useI18n({ useScope: 'global' })
const queryCache = useQueryCache()
const createMonitor = useMutation(createMonitorMutation())
const updateMonitor = useMutation(updateMonitorMutation())

const route = useRoute<'/app/(admin)/monitors/new' | '/app/(admin)/monitors/[id]/edit'>()
const router = useRouter()
const id = computed(() => ('id' in route.params ? route.params.id : undefined))
const editing = computed(() => !!id.value)
const loading = ref(true)
const loadError = ref('')
const error = ref('')
const monitorForm = ref<HTMLFormElement>()
const errorElement = ref<HTMLElement>()
const fileInput = ref<HTMLInputElement>()
let invalidTarget: HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement | undefined
let reportingInvalid = false
const channels = ref<Channel[]>([])
const secrets = ref<Secret[]>([])
const tagsInput = ref<InstanceType<typeof TagInput>>()
const formApi = useForm({
  defaultValues: newMonitor(),
  validators: {
    onSubmit: ({ value }) => validateMonitorForm(value),
  },
  onSubmit: async ({ value }) => {
    error.value = ''
    try {
      const payload = monitorPayload(value)
      const result = await (editing.value
        ? updateMonitor.mutateAsync({ path: { id: value.id }, body: payload })
        : createMonitor.mutateAsync({ body: payload }))
      notify(t('monitor-editor.monitor-saved'))
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
  loadError.value = ''
  try {
    const [c, s] = await Promise.all([
      queryCache.refresh(queryCache.ensure({ ...listChannelsQuery(), staleTime: 0 })),
      queryCache.refresh(queryCache.ensure({ ...listSecretsQuery(), staleTime: 0 })),
    ])
    if (c.status !== 'success')
      throw c.error || new Error(t('errors.request-failed'))
    if (s.status !== 'success')
      throw s.error || new Error(t('errors.request-failed'))
    channels.value = clone(c.data.items)
    secrets.value = clone(s.data.items)
    if (editing.value) {
      const monitor = await queryCache.refresh(
        queryCache.ensure({ ...getMonitorQuery({ path: { id: id.value! } }), staleTime: 0 }),
      )
      if (monitor.status !== 'success')
        throw monitor.error || new Error(t('errors.request-failed'))
      formApi.reset(toMonitorForm(monitor.data))
    }
  }
  catch (e) {
    loadError.value = errorText(e)
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
    throw new Error(t('monitor-editor.the-monitor-configuration-request-may-not-exceed-8'))
  return payload
}
function parseStatusRange(value: string) {
  const [min, max] = value.split('-').map(Number)
  return { min: min!, max: max ?? min! }
}
function formatStatusRange(range: { min: number, max: number }) {
  return `${range.min}-${range.max}`
}
const statusCodesInput = useListInput(
  () => form.value.type === 'http' ? form.value.http?.assertions.statusCodes ?? [] : [],
  value => formApi.setFieldValue('http.assertions.statusCodes', value),
  { parseItem: Number },
)
const statusRangesInput = useListInput(
  () => form.value.type === 'http' ? form.value.http?.assertions.statusRanges ?? [] : [],
  value => formApi.setFieldValue('http.assertions.statusRanges', value),
  { parseItem: parseStatusRange, formatItem: formatStatusRange },
)
const textContainsInput = useListInput(
  () => form.value.type === 'http' ? form.value.http?.assertions.textContains ?? [] : [],
  value => formApi.setFieldValue('http.assertions.textContains', value),
  { parseItem: String, separator: 'lines' },
)
const textNotContainsInput = useListInput(
  () => form.value.type === 'http' ? form.value.http?.assertions.textNotContains ?? [] : [],
  value => formApi.setFieldValue('http.assertions.textNotContains', value),
  { parseItem: String, separator: 'lines' },
)
const bodyRegexInput = useListInput(
  () => form.value.type === 'http' ? form.value.http?.assertions.regex ?? [] : [],
  value => formApi.setFieldValue('http.assertions.regex', value),
  { parseItem: String, separator: 'lines' },
)
const dnsExpectedValuesInput = useListInput(
  () => form.value.type === 'dns' ? form.value.dns?.expectedValues ?? [] : [],
  value => formApi.setFieldValue('dns.expectedValues', value),
  { parseItem: String, separator: 'lines', trim: true },
)
const warningDaysInput = useListInput(
  () => form.value.type === 'certificate' ? form.value.certificate?.warningDays ?? [] : [],
  value => formApi.setFieldValue('certificate.warningDays', value),
  { parseItem: Number },
)
const { text: headerAssertionsInput, error: headerError } = useJSONInput(
  () => form.value.type === 'http' ? form.value.http?.assertions.headers ?? [] : [],
  value => formApi.setFieldValue('http.assertions.headers', value),
  () => t('monitor-editor.header-assertions'),
)
const { text: jsonAssertionsInput, error: jsonError } = useJSONInput(
  () => form.value.type === 'http' ? form.value.http?.assertions.json ?? [] : [],
  value => formApi.setFieldValue('http.assertions.json', value),
  () => t('monitor-editor.json-assertions'),
)
function validateMonitorForm(value: MonitorForm): string | undefined {
  if (value.type === 'http' && (headerError.value || jsonError.value))
    return headerError.value || jsonError.value
  try {
    monitorPayload(value)
  }
  catch (e) {
    return errorText(e)
  }
}
async function save() {
  if (saving.value || loading.value || loadError.value)
    return
  error.value = ''
  tagsInput.value?.commit()
  await formApi.handleSubmit()
  if (form.value.type === 'http' && (headerError.value || jsonError.value)) {
    activeTab.value = 'target'
    await nextTick()
    const name = headerError.value ? 'http.assertions.headers' : 'http.assertions.json'
    monitorForm.value?.querySelector<HTMLElement>(`[name="${name}"]`)?.focus()
  }
  else if (error.value || validationError.value) {
    await nextTick()
    errorElement.value?.focus()
  }
}
async function revealInvalidField(event: Event) {
  if (reportingInvalid)
    return
  event.preventDefault()
  if (invalidTarget || !(event.target instanceof HTMLInputElement || event.target instanceof HTMLTextAreaElement || event.target instanceof HTMLSelectElement))
    return
  invalidTarget = event.target
  const panel = invalidTarget.closest('[role="tabpanel"]')
  const trigger = Array.from(monitorForm.value?.querySelectorAll<HTMLElement>('[role="tab"]') || [])
    .find(tab => tab.id === panel?.getAttribute('aria-labelledby'))
  if (trigger?.dataset.value)
    activeTab.value = trigger.dataset.value
  await nextTick()
  try {
    invalidTarget.focus()
    // Allow the visible control's native message without repeating tab recovery.
    reportingInvalid = true
    invalidTarget.reportValidity()
  }
  finally {
    reportingInvalid = false
    invalidTarget = undefined
  }
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
      notify(t('monitor-editor.multipart-files-may-total-at-most-4-mib'), 'error')
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
    class="mb-6"
    :title="editing ? t('monitor-editor.edit-monitor') : t('common.create-monitor')"
    :description="t('monitor-editor.define-the-target-success-criteria-and-confirmation-policy')"
  >
    <template #actions>
      <Button as-child>
        <RouterLink :to="editing ? `/app/monitors/${id}` : '/app/monitors'">
          <span class="i-lucide-arrow-left" w="15px" h="15px" aria-hidden="true" />{{ t('monitor-editor.back') }}
        </RouterLink>
      </Button><Button :disabled="saving || loading || !!loadError || !canEdit()" form="monitor-form" type="submit" variant="primary">
        <span class="i-lucide-save" w="15px" h="15px" aria-hidden="true" />{{ saving ? t('monitor-editor.saving') : t('common.save-monitor') }}
      </Button>
    </template>
  </PageHeader>
  <div v-if="loading" class="loading-state" flex="~ justify-center items-center gap-10px" p="60px" un-text="12px subtle" role="status">
    <Loader :label="t('async-state.loading-data')" />{{ t('async-state.loading-data') }}
  </div>
  <Banner v-else-if="loadError" variant="error">
    {{ loadError }}
    <Button class="mt-3" @click="load">
      {{ t('async-state.retry') }}
    </Button>
  </Banner>
  <template v-else>
    <form id="monitor-form" ref="monitorForm" class="monitor-editor" @submit.prevent="save" @invalid.capture="revealInvalidField">
      <div v-if="error || validationError" ref="errorElement" tabindex="-1" class="mb-5">
        <Banner size="sm" variant="error" role="alert">
          {{ error || validationError }}
        </Banner>
      </div>
      <LayerCard as-child>
        <section class="mb-6">
          <LayerCardPrimary class="p-0!">
            <FieldSection class="monitor-section">
              <h2>{{ t('monitor-editor.basic-information') }}</h2>
              <p un-text="13px subtle">
                {{ t('monitor-editor.use-a-recognizable-name-and-group-related-services') }}
              </p>
              <FieldGroup>
                <Field :label="t('common.display-name')">
                  <formApi.Field v-slot="{ field }" name="name">
                    <Input
                      :model-value="field.state.value" required maxlength="200"
                      :placeholder="t('monitor-editor.name-placeholder')"
                      @update:model-value="field.handleChange(String($event ?? ''))"
                      @blur="field.handleBlur"
                    />
                  </formApi.Field>
                </Field><Field
                  :label="t('monitor-editor.monitor-type')"
                  :description="
                    editing ? t('monitor-editor.the-monitor-type-is-fixed-after-creation-create') : undefined
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
                    <Input :model-value="field.state.value" :placeholder="t('monitor-editor.e-g-production')" @update:model-value="field.handleChange(String($event ?? ''))" @blur="field.handleBlur" />
                  </formApi.Field>
                </Field><Field :label="t('monitor-editor.tags')" :description="t('monitor-editor.separate-with-commas')">
                  <formApi.Field v-slot="{ field }" name="tags">
                    <TagInput ref="tagsInput" allow-duplicates :labels="{ removeValue: value => `${t('common.delete')} ${value}`, editValue: value => `${t('common.edit')} ${value}` }" :model-value="field.state.value" :name="field.name" placeholder="production, api" @update:model-value="field.handleChange" @focusout="field.handleBlur" />
                  </formApi.Field>
                </Field><Field class="span-full" :label="t('monitor-editor.description')">
                  <formApi.Field v-slot="{ field }" name="description">
                    <InputArea :model-value="field.state.value" rows="2" @update:model-value="field.handleChange(String($event ?? ''))" @blur="field.handleBlur" />
                  </formApi.Field>
                </Field>
                <div class="span-full">
                  <formApi.Field v-slot="{ field }" name="enabled">
                    <Switch
                      :model-value="field.state.value" :label="t('monitor-editor.enable-monitor')" :description="t('monitor-editor.pausing-stops-collection-and-excludes-paused-time-from')"
                      @update:model-value="field.handleChange"
                      @focusout="field.handleBlur"
                    />
                  </formApi.Field>
                </div>
              </FieldGroup>
            </FieldSection>
          </LayerCardPrimary>
        </section>
      </LayerCard>
      <LayerCard as-child>
        <section class="monitor-configuration">
          <LayerCardPrimary class="p-0!">
            <TabsRoot v-model="activeTab">
              <TabsList>
                <TabsTrigger value="target">
                  {{
                    t('monitor-editor.check-target')
                  }}
                </TabsTrigger><TabsTrigger v-if="form.type !== 'heartbeat'" value="schedule">
                  {{
                    t('monitor-editor.schedule-retries')
                  }}
                </TabsTrigger><TabsTrigger value="notifications">
                  {{
                    t('monitor-editor.notifications')
                  }}
                </TabsTrigger>
              </TabsList><TabsContent value="target">
                <template v-if="form.type === 'http' && form.http">
                  <FieldSection class="monitor-section">
                    <h2>{{ t('monitor-editor.http-request') }}</h2>
                    <p un-text="13px subtle">
                      {{ t('monitor-editor.all-request-methods-use-the-same-round-retry') }}
                    </p>
                    <FieldGroup>
                      <Field :label="t('monitor-editor.target-url')" class="span-full">
                        <formApi.Field v-slot="{ field }" name="http.url">
                          <Input
                            :model-value="field.state.value" type="url" required
                            placeholder="https://api.example.com/health"
                            @update:model-value="field.handleChange(String($event ?? ''))"
                            @blur="field.handleBlur"
                          />
                        </formApi.Field>
                      </Field><Field :label="t('monitor-editor.request-method')">
                        <formApi.Field v-slot="{ field }" name="http.method">
                          <Input :model-value="field.state.value" list="http-methods" required @update:model-value="field.handleChange(String($event ?? ''))" @blur="field.handleBlur" />
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
                      </Field><Field :label="t('monitor-editor.host-override')">
                        <formApi.Field v-slot="{ field }" name="http.host">
                          <Input :model-value="field.state.value" placeholder="api.example.com" @update:model-value="field.handleChange(String($event ?? ''))" @blur="field.handleBlur" />
                        </formApi.Field>
                      </Field>
                      <div class="span-full">
                        <FieldLabel as="div">
                          {{ t('monitor-editor.query-parameters') }}
                        </FieldLabel><formApi.Field v-slot="{ field }" name="http.query">
                          <KeyValues :model-value="field.state.value!" :secrets="secrets" mt="2" @update:model-value="field.handleChange" @focusout="field.handleBlur" />
                        </formApi.Field>
                      </div>
                      <div class="span-full">
                        <FieldLabel as="div">
                          {{
                            t('monitor-editor.request-headers-repeated-names-supported')
                          }}
                        </FieldLabel><formApi.Field v-slot="{ field }" name="http.headers">
                          <KeyValues :model-value="field.state.value!" :secrets="secrets" mt="2" @update:model-value="field.handleChange" @focusout="field.handleBlur" />
                        </formApi.Field>
                      </div>
                    </FieldGroup>
                  </FieldSection>
                  <FieldSection class="monitor-section">
                    <h2>{{ t('monitor-editor.request-body') }}</h2>
                    <p un-text="13px subtle">
                      {{ t('monitor-editor.configure-format-character-encoding-and-compression-separately') }}
                    </p>
                    <FieldGroup>
                      <Field :label="t('monitor-editor.body-format')">
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
                        :label="t('common.character-encoding')"
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
                        :label="t('monitor-editor.body-secret-reference')"
                      >
                        <formApi.Field v-slot="{ field }" name="http.body.secretRef">
                          <Select :model-value="field.state.value" @update:model-value="field.handleChange">
                            <SelectTrigger @focusout="field.handleBlur">
                              <SelectValue :placeholder="t('secret-select.no-secret-reference')" />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectGroup>
                                <SelectItem value="">
                                  {{ t('secret-select.no-secret-reference') }}
                                </SelectItem>
                                <SelectItem v-for="secret in secrets" :key="secret.id" :value="secret.id">
                                  {{ secret.name }}
                                </SelectItem>
                              </SelectGroup>
                            </SelectContent>
                          </Select>
                        </formApi.Field>
                      </Field><Field
                        v-if="['text', 'raw'].includes(form.http.body.format)"
                        :label="t('monitor-editor.content-type')"
                      >
                        <formApi.Field v-slot="{ field }" name="http.body.contentType">
                          <Input :model-value="field.state.value" placeholder="text/plain" @update:model-value="field.handleChange(String($event ?? ''))" @blur="field.handleBlur" />
                        </formApi.Field>
                      </Field><Field
                        v-if="
                          ['json', 'text'].includes(form.http.body.format) && !form.http.body.secretRef
                        "
                        class="span-full"
                        :label="t('monitor-editor.body-content')"
                      >
                        <formApi.Field v-slot="{ field }" name="http.body.text">
                          <InputArea
                            :model-value="field.state.value" rows="7" :placeholder="form.http.body.format === 'json' ? '{}' : ''"
                            spellcheck="false"
                            @update:model-value="field.handleChange(String($event ?? ''))"
                            @blur="field.handleBlur"
                          />
                        </formApi.Field>
                      </Field><Field
                        v-if="form.http.body.format === 'raw' && !form.http.body.secretRef"
                        class="span-full"
                        :label="t('monitor-editor.raw-bytes-base-64')"
                      >
                        <formApi.Field v-slot="{ field }" name="http.body.base64">
                          <InputArea :model-value="field.state.value" spellcheck="false" @update:model-value="field.handleChange(String($event ?? ''))" @blur="field.handleBlur" />
                        </formApi.Field>
                      </Field>
                      <div
                        v-if="['form', 'multipart'].includes(form.http.body.format)"
                        class="span-full"
                      >
                        <FieldLabel as="div">
                          {{ t('monitor-editor.form-fields') }}
                        </FieldLabel><formApi.Field v-slot="{ field }" name="http.body.fields">
                          <KeyValues :model-value="field.state.value!" :secrets="secrets" mt="2" @update:model-value="field.handleChange" @focusout="field.handleBlur" />
                        </formApi.Field>
                      </div>
                      <div v-if="form.http.body.format === 'multipart'" class="span-full">
                        <FieldLabel as="div">
                          {{ t('monitor-editor.files') }}
                        </FieldLabel>
                        <Banner
                          v-for="(file, index) in form.http.body.files" :key="index"
                          size="sm"
                          variant="secondary"
                          mt="2"
                        >
                          <div class="flex min-w-0 flex-wrap items-center gap-3">
                            <formApi.Field v-slot="{ field }" :name="`http.body.files[${index}].field`">
                              <Input
                                :model-value="field.state.value" :aria-label="t('monitor-editor.field-name')" placeholder="file" class="min-w-32 flex-1"
                                @update:model-value="field.handleChange(String($event ?? ''))"
                                @blur="field.handleBlur"
                              />
                            </formApi.Field><span class="min-w-0 flex-1 [overflow-wrap:anywhere]">{{ file.filename }}</span><Button
                              type="button"
                              :aria-label="t('monitor-editor.remove-file')"
                              shape="square"
                              @click="formApi.setFieldValue('http.body.files', files => files?.filter((_, i) => i !== index))"
                            >
                              <span class="i-lucide-x" w="15px" h="15px" aria-hidden="true" />
                            </Button>
                          </div>
                        </Banner>
                        <Button size="sm" class="mt-3" @click="fileInput?.click()">
                          <span class="i-lucide-upload" w="13px" h="13px" aria-hidden="true" />{{ t('monitor-editor.add-files') }}
                        </Button>
                        <input ref="fileInput" type="file" multiple hidden :aria-label="t('monitor-editor.add-files')" @change="addFile">
                        <FieldDescription mt="2">
                          {{
                            t('monitor-editor.multipart-boundary-and-content-type-are-generated-automatically')
                          }}
                        </FieldDescription>
                      </div>
                    </FieldGroup>
                  </FieldSection>
                  <FieldSection class="monitor-section">
                    <h2>{{ t('monitor-editor.authentication') }}</h2>
                    <p un-text="13px subtle">
                      {{ t('monitor-editor.credentials-use-existing-secret-references') }}
                    </p>
                    <FieldGroup>
                      <Field :label="t('monitor-editor.authentication-type')">
                        <formApi.Field v-slot="{ field }" name="http.auth.type">
                          <Select :model-value="field.state.value" @update:model-value="field.handleChange" @focusout="field.handleBlur">
                            <SelectTrigger><SelectValue /></SelectTrigger>
                            <SelectContent>
                              <SelectGroup>
                                <SelectItem value="none">
                                  {{ t('monitor-editor.none') }}
                                </SelectItem>
                                <SelectItem value="basic">
                                  Basic
                                </SelectItem>
                                <SelectItem value="bearer">
                                  Bearer
                                </SelectItem>
                                <SelectItem value="header">
                                  {{ t('monitor-editor.api-key-header') }}
                                </SelectItem>
                              </SelectGroup>
                            </SelectContent>
                          </Select>
                        </formApi.Field>
                      </Field><Field
                        v-if="form.http.auth.type !== 'none'"
                        :label="t('monitor-editor.password-token-secret')"
                      >
                        <formApi.Field v-slot="{ field }" name="http.auth.secretRef">
                          <Select :model-value="field.state.value" @update:model-value="field.handleChange">
                            <SelectTrigger @focusout="field.handleBlur">
                              <SelectValue :placeholder="t('secret-select.choose-a-secret')" />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectGroup>
                                <SelectItem value="">
                                  {{ t('secret-select.choose-a-secret') }}
                                </SelectItem>
                                <SelectItem v-for="secret in secrets" :key="secret.id" :value="secret.id">
                                  {{ secret.name }}
                                </SelectItem>
                              </SelectGroup>
                            </SelectContent>
                          </Select>
                        </formApi.Field>
                      </Field><Field v-if="form.http.auth.type === 'basic'" :label="t('common.username')">
                        <formApi.Field v-slot="{ field }" name="http.auth.username">
                          <Input :model-value="field.state.value" @update:model-value="field.handleChange(String($event ?? ''))" @blur="field.handleBlur" />
                        </formApi.Field>
                      </Field><Field
                        v-if="form.http.auth.type === 'basic'"
                        :label="t('monitor-editor.username-secret-optional')"
                      >
                        <formApi.Field v-slot="{ field }" name="http.auth.usernameSecretRef">
                          <Select :model-value="field.state.value" @update:model-value="field.handleChange">
                            <SelectTrigger @focusout="field.handleBlur">
                              <SelectValue :placeholder="t('secret-select.no-secret-reference')" />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectGroup>
                                <SelectItem value="">
                                  {{ t('secret-select.no-secret-reference') }}
                                </SelectItem>
                                <SelectItem v-for="secret in secrets" :key="secret.id" :value="secret.id">
                                  {{ secret.name }}
                                </SelectItem>
                              </SelectGroup>
                            </SelectContent>
                          </Select>
                        </formApi.Field>
                      </Field><Field
                        v-if="form.http.auth.type === 'header'"
                        :label="t('monitor-editor.authentication-header')"
                      >
                        <formApi.Field v-slot="{ field }" name="http.auth.header">
                          <Input :model-value="field.state.value" @update:model-value="field.handleChange(String($event ?? ''))" @blur="field.handleBlur" />
                        </formApi.Field>
                      </Field><Field
                        v-if="form.http.auth.type === 'header'"
                        :label="t('monitor-editor.value-prefix')"
                      >
                        <formApi.Field v-slot="{ field }" name="http.auth.prefix">
                          <Input :model-value="field.state.value" placeholder="Bearer " @update:model-value="field.handleChange(String($event ?? ''))" @blur="field.handleBlur" />
                        </formApi.Field>
                      </Field>
                    </FieldGroup>
                  </FieldSection>
                  <FieldSection class="monitor-section">
                    <h2>{{ t('monitor-editor.success-assertions') }}</h2>
                    <p un-text="13px subtle">
                      {{ t('monitor-editor.every-configured-assertion-must-pass-for-a-successful') }}
                    </p>
                    <FieldGroup>
                      <Field
                        :label="t('monitor-editor.status-codes')"
                        :description="t('monitor-editor.comma-separated-accepted-together-with-ranges')"
                      >
                        <formApi.Field v-slot="{ field }" name="http.assertions.statusCodes">
                          <Input v-model="statusCodesInput" placeholder="200, 204" @blur="field.handleBlur" />
                        </formApi.Field>
                      </Field><Field :label="t('monitor-editor.status-ranges')">
                        <formApi.Field v-slot="{ field }" name="http.assertions.statusRanges">
                          <Input v-model="statusRangesInput" placeholder="200-299, 300-399" @blur="field.handleBlur" />
                        </formApi.Field>
                      </Field><Field :label="t('monitor-editor.text-contains-one-per-line')">
                        <formApi.Field v-slot="{ field }" name="http.assertions.textContains">
                          <InputArea v-model="textContainsInput" @blur="field.handleBlur" />
                        </formApi.Field>
                      </Field><Field :label="t('monitor-editor.text-excludes-one-per-line')">
                        <formApi.Field v-slot="{ field }" name="http.assertions.textNotContains">
                          <InputArea v-model="textNotContainsInput" @blur="field.handleBlur" />
                        </formApi.Field>
                      </Field><Field :label="t('monitor-editor.body-regex-one-per-line')">
                        <formApi.Field v-slot="{ field }" name="http.assertions.regex">
                          <InputArea v-model="bodyRegexInput" spellcheck="false" @blur="field.handleBlur" />
                        </formApi.Field>
                      </Field><Field
                        :label="t('monitor-editor.maximum-response-time-ms')"
                        :description="t('monitor-editor.0-disables-this-assertion')"
                      >
                        <formApi.Field v-slot="{ field }" name="http.assertions.maxLatencyMs">
                          <Input
                            :model-value="field.state.value" type="number" min="0"
                            @input="field.handleChange(($event.target as HTMLInputElement).valueAsNumber)"
                            @blur="field.handleBlur"
                          />
                        </formApi.Field>
                      </Field><Field
                        class="span-full"
                        :label="t('monitor-editor.response-header-assertions')"
                        :error="headerError"
                        :description="headerHint"
                      >
                        <formApi.Field v-slot="{ field }" name="http.assertions.headers">
                          <InputArea v-model="headerAssertionsInput" :name="field.name" rows="4" spellcheck="false" @blur="field.handleBlur" />
                        </formApi.Field>
                      </Field><Field
                        class="span-full"
                        :label="t('monitor-editor.json-field-assertions-json-pointer')"
                        :error="jsonError"
                        :description="jsonHint"
                      >
                        <formApi.Field v-slot="{ field }" name="http.assertions.json">
                          <InputArea v-model="jsonAssertionsInput" :name="field.name" rows="4" spellcheck="false" @blur="field.handleBlur" />
                        </formApi.Field>
                      </Field>
                    </FieldGroup>
                  </FieldSection>
                  <FieldSection class="monitor-section">
                    <Collapsible>
                      <CollapsibleTrigger as-child>
                        <Button variant="ghost" class="group w-full justify-start">
                          <span class="i-lucide-chevron-right size-4 motion-safe:transition-transform group-data-[state=open]:rotate-90" aria-hidden="true" />
                          {{ t('monitor-editor.advanced-tls-connections-transport') }}
                        </Button>
                      </CollapsibleTrigger><CollapsibleContent class="pt-4">
                        <h3 mb="4">
                          TLS
                        </h3>
                        <formApi.Field v-slot="{ field }" name="http.tls">
                          <TLSFields :model-value="field.state.value!" :secrets="secrets" @update:model-value="field.handleChange" @focusout="field.handleBlur" />
                        </formApi.Field>
                        <Separator />
                        <h3 mb="4">
                          {{ t('monitor-editor.network-connection') }}
                        </h3>
                        <formApi.Field v-slot="{ field }" name="http.connection">
                          <ConnectionFields :model-value="field.state.value!" :secrets="secrets" @update:model-value="field.handleChange" @focusout="field.handleBlur" />
                        </formApi.Field>
                        <Separator />
                        <FieldGroup>
                          <div class="span-full">
                            <formApi.Field v-slot="{ field }" name="http.redirects.enabled">
                              <Switch
                                :model-value="field.state.value" :label="t('monitor-editor.follow-redirects')" @update:model-value="field.handleChange"
                                @focusout="field.handleBlur"
                              />
                            </formApi.Field>
                          </div>
                          <Field :label="t('monitor-editor.maximum-redirects')">
                            <formApi.Field v-slot="{ field }" name="http.redirects.maxHops">
                              <Input
                                :model-value="field.state.value" type="number" min="1"
                                max="20"
                                @input="field.handleChange(($event.target as HTMLInputElement).valueAsNumber)"
                                @blur="field.handleBlur"
                              />
                            </formApi.Field>
                          </Field><Field :label="t('monitor-editor.redirect-scope')">
                            <formApi.Field v-slot="{ field }" name="http.redirects.scope">
                              <Select :model-value="field.state.value" @update:model-value="field.handleChange" @focusout="field.handleBlur">
                                <SelectTrigger><SelectValue /></SelectTrigger>
                                <SelectContent>
                                  <SelectGroup>
                                    <SelectItem value="same-origin">
                                      {{ t('monitor-editor.same-origin') }}
                                    </SelectItem>
                                    <SelectItem value="same-host">
                                      {{ t('monitor-editor.same-host') }}
                                    </SelectItem>
                                    <SelectItem value="any">
                                      {{ t('monitor-editor.any-target') }}
                                    </SelectItem>
                                  </SelectGroup>
                                </SelectContent>
                              </Select>
                            </formApi.Field>
                          </Field><Field :label="t('monitor-editor.accept-encoding')">
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
                          </Field><Field :label="t('monitor-editor.response-charset')">
                            <formApi.Field v-slot="{ field }" name="http.responseCharset">
                              <Select :model-value="field.state.value" @update:model-value="field.handleChange" @focusout="field.handleBlur">
                                <SelectTrigger><SelectValue :placeholder="t('monitor-editor.detect-from-response')" /></SelectTrigger>
                                <SelectContent>
                                  <SelectGroup>
                                    <SelectItem value="">
                                      {{ t('monitor-editor.detect-from-response') }}
                                    </SelectItem>
                                    <SelectItem v-for="charset in charsetOptions" :key="charset" :value="charset">
                                      {{ charset }}
                                    </SelectItem>
                                  </SelectGroup>
                                </SelectContent>
                              </Select>
                            </formApi.Field>
                          </Field><Field :label="t('monitor-editor.decompressed-response-limit-bytes')">
                            <formApi.Field v-slot="{ field }" name="http.maxResponseBytes">
                              <Input
                                :model-value="field.state.value" type="number" min="1"
                                max="16777216"
                                @input="field.handleChange(($event.target as HTMLInputElement).valueAsNumber)"
                                @blur="field.handleBlur"
                              />
                            </formApi.Field>
                          </Field>
                          <div>
                            <formApi.Field v-slot="{ field }" name="http.requestGzip">
                              <Switch
                                :model-value="field.state.value" :label="t('monitor-editor.gzip-request-body')" @update:model-value="field.handleChange"
                                @focusout="field.handleBlur"
                              />
                            </formApi.Field>
                          </div>
                        </FieldGroup>
                      </CollapsibleContent>
                    </Collapsible>
                  </FieldSection>
                </template><template v-if="form.type === 'tcp' && form.tcp">
                  <FieldSection class="monitor-section">
                    <h2>{{ t('monitor-types.tcp-connection') }}</h2>
                    <p un-text="13px subtle">
                      {{ t('monitor-editor.test-a-connection-or-send-a-payload-and') }}
                    </p>
                    <FieldGroup>
                      <Field :label="t('monitor-editor.host')">
                        <formApi.Field v-slot="{ field }" name="tcp.host">
                          <Input :model-value="field.state.value" required placeholder="db.example.com" @update:model-value="field.handleChange(String($event ?? ''))" @blur="field.handleBlur" />
                        </formApi.Field>
                      </Field><Field :label="t('common.port')">
                        <formApi.Field v-slot="{ field }" name="tcp.port">
                          <Input
                            :model-value="field.state.value" type="number" min="1"
                            max="65535"
                            required
                            @input="field.handleChange(($event.target as HTMLInputElement).valueAsNumber)"
                            @blur="field.handleBlur"
                          />
                        </formApi.Field>
                      </Field><Field :label="t('monitor-editor.send-text')">
                        <formApi.Field v-slot="{ field }" name="tcp.sendText">
                          <InputArea :model-value="field.state.value" @update:model-value="field.handleChange(String($event ?? ''))" @blur="field.handleBlur" />
                        </formApi.Field>
                      </Field><Field :label="t('monitor-editor.send-bytes-base-64')">
                        <formApi.Field v-slot="{ field }" name="tcp.sendBase64">
                          <InputArea :model-value="field.state.value" spellcheck="false" @update:model-value="field.handleChange(String($event ?? ''))" @blur="field.handleBlur" />
                        </formApi.Field>
                      </Field><Field :label="t('monitor-editor.payload-secret')">
                        <formApi.Field v-slot="{ field }" name="tcp.sendSecretRef">
                          <Select :model-value="field.state.value" @update:model-value="field.handleChange">
                            <SelectTrigger @focusout="field.handleBlur">
                              <SelectValue :placeholder="t('secret-select.no-secret-reference')" />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectGroup>
                                <SelectItem value="">
                                  {{ t('secret-select.no-secret-reference') }}
                                </SelectItem>
                                <SelectItem v-for="secret in secrets" :key="secret.id" :value="secret.id">
                                  {{ secret.name }}
                                </SelectItem>
                              </SelectGroup>
                            </SelectContent>
                          </Select>
                        </formApi.Field>
                      </Field><Field :label="t('common.character-encoding')">
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
                      </Field><Field :label="t('monitor-editor.response-contains')">
                        <formApi.Field v-slot="{ field }" name="tcp.receiveContains">
                          <Input :model-value="field.state.value" @update:model-value="field.handleChange(String($event ?? ''))" @blur="field.handleBlur" />
                        </formApi.Field>
                      </Field><Field :label="t('monitor-editor.response-regex')">
                        <formApi.Field v-slot="{ field }" name="tcp.receiveRegex">
                          <Input :model-value="field.state.value" @update:model-value="field.handleChange(String($event ?? ''))" @blur="field.handleBlur" />
                        </formApi.Field>
                      </Field><Field :label="t('monitor-editor.receive-size-limit-bytes')">
                        <formApi.Field v-slot="{ field }" name="tcp.maxReceiveBytes">
                          <Input
                            :model-value="field.state.value" type="number" min="1"
                            max="16777216"
                            @input="field.handleChange(($event.target as HTMLInputElement).valueAsNumber)"
                            @blur="field.handleBlur"
                          />
                        </formApi.Field>
                      </Field>
                    </FieldGroup>
                  </FieldSection>
                  <FieldSection class="monitor-section">
                    <Collapsible>
                      <CollapsibleTrigger as-child>
                        <Button variant="ghost" class="group w-full justify-start">
                          <span class="i-lucide-chevron-right size-4 motion-safe:transition-transform group-data-[state=open]:rotate-90" aria-hidden="true" />
                          {{ t('common.tls-connection-settings') }}
                        </Button>
                      </CollapsibleTrigger><CollapsibleContent class="pt-4">
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
                      </CollapsibleContent>
                    </Collapsible>
                  </FieldSection>
                </template><template v-if="form.type === 'dns' && form.dns">
                  <FieldSection class="monitor-section">
                    <h2>{{ t('monitor-editor.dns-query') }}</h2>
                    <p un-text="13px subtle">
                      {{ t('monitor-editor.validate-response-codes-and-record-values') }}
                    </p>
                    <FieldGroup>
                      <Field :label="t('monitor-editor.query-name')">
                        <formApi.Field v-slot="{ field }" name="dns.name">
                          <Input :model-value="field.state.value" required placeholder="example.com" @update:model-value="field.handleChange(String($event ?? ''))" @blur="field.handleBlur" />
                        </formApi.Field>
                      </Field><Field :label="t('monitor-editor.record-type')">
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
                        :label="t('monitor-editor.dns-server')"
                        :description="t('monitor-editor.leave-empty-for-system-resolver-or-enter-host')"
                      >
                        <formApi.Field v-slot="{ field }" name="dns.server">
                          <Input :model-value="field.state.value" placeholder="1.1.1.1:53" @update:model-value="field.handleChange(String($event ?? ''))" @blur="field.handleBlur" />
                        </formApi.Field>
                      </Field><Field :label="t('monitor-editor.protocol')">
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
                      </Field><Field :label="t('monitor-editor.expected-response-code')">
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
                      </Field><Field :label="t('monitor-editor.record-matching')">
                        <formApi.Field v-slot="{ field }" name="dns.matchMode">
                          <Select :model-value="field.state.value" @update:model-value="field.handleChange" @focusout="field.handleBlur">
                            <SelectTrigger><SelectValue /></SelectTrigger>
                            <SelectContent>
                              <SelectGroup>
                                <SelectItem value="contains">
                                  {{ t('monitor-editor.contains-expected-records') }}
                                </SelectItem>
                                <SelectItem value="exact">
                                  {{ t('monitor-editor.exact-set') }}
                                </SelectItem>
                              </SelectGroup>
                            </SelectContent>
                          </Select>
                        </formApi.Field>
                      </Field><Field class="span-full" :label="t('monitor-editor.expected-values-one-per-line')">
                        <formApi.Field v-slot="{ field }" name="dns.expectedValues">
                          <InputArea v-model="dnsExpectedValuesInput" @blur="field.handleBlur" />
                        </formApi.Field>
                      </Field>
                    </FieldGroup>
                  </FieldSection>
                </template><template v-if="form.type === 'heartbeat' && form.heartbeat">
                  <FieldSection class="monitor-section">
                    <h2>{{ t('monitor-types.heartbeat') }}</h2>
                    <p un-text="13px subtle">
                      {{ t('monitor-editor.services-report-periodically-missing-a-period-plus-grace') }}
                    </p>
                    <FieldGroup>
                      <Field :label="t('monitor-editor.expected-period-seconds')">
                        <formApi.Field v-slot="{ field }" name="heartbeat.periodSeconds">
                          <Input
                            :model-value="field.state.value" type="number" min="30"
                            required
                            @input="field.handleChange(($event.target as HTMLInputElement).valueAsNumber)"
                            @blur="field.handleBlur"
                          />
                        </formApi.Field>
                      </Field><Field :label="t('monitor-editor.grace-period-seconds')">
                        <formApi.Field v-slot="{ field }" name="heartbeat.graceSeconds">
                          <Input
                            :model-value="field.state.value" type="number" min="0"
                            required
                            @input="field.handleChange(($event.target as HTMLInputElement).valueAsNumber)"
                            @blur="field.handleBlur"
                          />
                        </formApi.Field>
                      </Field>
                    </FieldGroup>
                    <Banner size="sm" variant="secondary" mt="5">
                      {{ t('monitor-editor.after-saving-generate-or-rotate-the-report-token') }}
                    </Banner>
                  </FieldSection>
                </template><template v-if="form.type === 'certificate' && form.certificate">
                  <FieldSection class="monitor-section">
                    <h2>{{ t('monitor-editor.certificate-expiry') }}</h2>
                    <p un-text="13px subtle">
                      {{ t('monitor-editor.certificate-risk-is-displayed-separately-and-excluded-from') }}
                    </p>
                    <FieldGroup>
                      <Field :label="t('monitor-editor.tls-host')">
                        <formApi.Field v-slot="{ field }" name="certificate.host">
                          <Input :model-value="field.state.value" required @update:model-value="field.handleChange(String($event ?? ''))" @blur="field.handleBlur" />
                        </formApi.Field>
                      </Field><Field :label="t('common.port')">
                        <formApi.Field v-slot="{ field }" name="certificate.port">
                          <Input
                            :model-value="field.state.value" type="number" min="1"
                            max="65535"
                            @input="field.handleChange(($event.target as HTMLInputElement).valueAsNumber)"
                            @blur="field.handleBlur"
                          />
                        </formApi.Field>
                      </Field><Field
                        class="span-full"
                        :label="t('common.warning-thresholds-days')"
                        :description="t('monitor-editor.comma-separated-defaults-301471-days')"
                      >
                        <formApi.Field v-slot="{ field }" name="certificate.warningDays">
                          <Input v-model="warningDaysInput" @blur="field.handleBlur" />
                        </formApi.Field>
                      </Field>
                    </FieldGroup>
                  </FieldSection>
                  <FieldSection class="monitor-section">
                    <Collapsible>
                      <CollapsibleTrigger as-child>
                        <Button variant="ghost" class="group w-full justify-start">
                          <span class="i-lucide-chevron-right size-4 motion-safe:transition-transform group-data-[state=open]:rotate-90" aria-hidden="true" />
                          {{ t('common.tls-connection-settings') }}
                        </Button>
                      </CollapsibleTrigger><CollapsibleContent class="pt-4">
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
                      </CollapsibleContent>
                    </Collapsible>
                  </FieldSection>
                </template>
              </TabsContent><TabsContent v-if="form.type !== 'heartbeat'" value="schedule">
                <FieldSection class="monitor-section">
                  <h2>{{ t('monitor-editor.check-schedule') }}</h2>
                  <p un-text="13px subtle">
                    {{ t('monitor-editor.checks-run-on-a-fixed-cadence-without-overlap') }}
                  </p>
                  <FieldGroup>
                    <Field
                      :label="t('monitor-editor.check-interval-seconds')"
                      :description="
                        form.type === 'certificate'
                          ? t('monitor-editor.certificates-default-to-daily-checks-86400-seconds')
                          : t('monitor-editor.minimum-30-seconds')
                      "
                    >
                      <formApi.Field v-slot="{ field }" name="intervalSeconds">
                        <Input
                          :model-value="field.state.value" type="number" min="30"
                          max="2592000"
                          required
                          @input="field.handleChange(($event.target as HTMLInputElement).valueAsNumber)"
                          @blur="field.handleBlur"
                        />
                      </formApi.Field>
                    </Field><Field :label="t('monitor-editor.attempt-timeout-seconds')">
                      <formApi.Field v-slot="{ field }" name="timeoutSeconds">
                        <Input
                          :model-value="field.state.value" type="number" min="1"
                          :max="form.intervalSeconds"
                          required
                          @input="field.handleChange(($event.target as HTMLInputElement).valueAsNumber)"
                          @blur="field.handleBlur"
                        />
                      </formApi.Field>
                    </Field><template v-if="['http', 'tcp', 'dns'].includes(form.type)">
                      <Field
                        :label="t('monitor-editor.additional-retries')"
                        :description="t('monitor-editor.default-2-zero-means-the-first-attempt-only')"
                      >
                        <formApi.Field v-slot="{ field }" name="retries">
                          <Input :model-value="field.state.value" type="number" min="0" max="10" @input="field.handleChange(($event.target as HTMLInputElement).valueAsNumber)" @blur="field.handleBlur" />
                        </formApi.Field>
                      </Field><Field :label="t('monitor-editor.retry-delay-seconds')">
                        <formApi.Field v-slot="{ field }" name="retryDelaySeconds">
                          <Input
                            :model-value="field.state.value" type="number" min="0"
                            :max="form.intervalSeconds"
                            @input="field.handleChange(($event.target as HTMLInputElement).valueAsNumber)"
                            @blur="field.handleBlur"
                          />
                        </formApi.Field>
                      </Field><Field :label="t('monitor-editor.consecutive-failed-rounds')">
                        <formApi.Field v-slot="{ field }" name="failureThreshold">
                          <Input
                            :model-value="field.state.value" type="number" min="1"
                            max="100"
                            @input="field.handleChange(($event.target as HTMLInputElement).valueAsNumber)"
                            @blur="field.handleBlur"
                          />
                        </formApi.Field>
                      </Field><Field :label="t('monitor-editor.consecutive-successful-rounds')">
                        <formApi.Field v-slot="{ field }" name="recoveryThreshold">
                          <Input
                            :model-value="field.state.value" type="number" min="1"
                            max="100"
                            @input="field.handleChange(($event.target as HTMLInputElement).valueAsNumber)"
                            @blur="field.handleBlur"
                          />
                        </formApi.Field>
                      </Field>
                    </template>
                  </FieldGroup>
                  <Banner size="sm" variant="secondary" mt="6">
                    {{ t('monitor-editor.retry-count-is-a-limit-insufficient-budget-ends') }}
                  </Banner>
                </FieldSection>
              </TabsContent><TabsContent value="notifications">
                <FieldSection class="monitor-section">
                  <h2>{{ t('monitor-editor.notification-channels') }}</h2>
                  <p un-text="13px subtle">
                    {{
                      t(
                        'monitor-editor.select-administrator-configured-channels-confirmed-failures-create-durable',
                      )
                    }}
                  </p>
                  <formApi.Field v-slot="{ field }" name="notificationChannelIds">
                    <CheckboxGroup :model-value="field.state.value" orientation="horizontal" :aria-label="t('monitor-editor.notification-channels')" @update:model-value="field.handleChange" @focusout="field.handleBlur">
                      <CheckboxItem v-for="channel in channels" :key="channel.id" :value="channel.id" :label="channel.name" :description="!channel.enabled ? t('monitor-editor.disabled') : undefined" />
                    </CheckboxGroup>
                  </formApi.Field>
                  <p v-if="!channels.length" un-text="13px subtle">
                    {{ t('monitor-editor.no-channels-yet-an-administrator-can-create-one') }}
                  </p>
                  <Separator />
                  <formApi.Field v-if="form.type === 'certificate' && form.certificate" v-slot="{ field }" name="certificate.notifyRenewal">
                    <Switch

                      :model-value="field.state.value" :label="t('monitor-editor.notify-on-certificate-renewal')" mb="5"
                      @update:model-value="field.handleChange"
                      @focusout="field.handleBlur"
                    />
                  </formApi.Field><formApi.Field v-else v-slot="{ field }" name="notifyRecovery">
                    <Switch

                      :model-value="field.state.value" :label="t('monitor-editor.notify-on-recovery')" mb="5"
                      @update:model-value="field.handleChange"
                      @focusout="field.handleBlur"
                    />
                  </formApi.Field><Field
                    v-if="form.type !== 'certificate'"
                    :label="t('monitor-editor.repeated-outage-reminder-seconds')"
                    :description="t('monitor-editor.0-disables-repeated-reminders-minimum-30-seconds-when')"
                  >
                    <formApi.Field v-slot="{ field }" name="reminderSeconds">
                      <Input :model-value="field.state.value" type="number" min="0" @input="field.handleChange(($event.target as HTMLInputElement).valueAsNumber)" @blur="field.handleBlur" />
                    </formApi.Field>
                  </Field>
                </FieldSection>
              </TabsContent>
            </TabsRoot>
          </LayerCardPrimary>
        </section>
      </LayerCard>
      <FieldActions class="sticky bottom-0 z-20 mt-6! rounded-xl border border-solid border-line bg-canvas p-4 shadow-control">
        <Button as-child>
          <RouterLink to="/app/monitors">
            {{ t('common.cancel') }}
          </RouterLink>
        </Button><Button type="submit" :disabled="saving || !canEdit()" variant="primary">
          <span class="i-lucide-save" w="15px" h="15px" aria-hidden="true" />{{ t('common.save-monitor') }}
        </Button>
      </FieldActions>
    </form>
  </template>
</template>

<style scoped>
.monitor-editor {
  container: monitor-editor / inline-size;
}

.monitor-configuration :deep([data-scope='tabs'][data-part='list']) {
  margin: 1.25rem 1.5rem 0;
}

.monitor-configuration :deep([data-scope='tabs'][data-part='content']) {
  padding: 1.5rem;
}

.monitor-configuration :deep(.monitor-section) {
  padding: 0;
  border: 0;
}

.monitor-configuration :deep(.monitor-section + .monitor-section) {
  margin-block-start: 32px;
}

@container monitor-editor (max-width: 600px) {
  .monitor-configuration :deep([data-scope='tabs'][data-part='list']) {
    margin-inline: 1.25rem;
  }

  .monitor-configuration :deep([data-scope='tabs'][data-part='content']) {
    padding: 1.25rem;
  }
}
</style>
