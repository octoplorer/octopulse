<script setup lang="ts">
import type { Monitor, Page, PublicPage } from '../client/types.gen'
import { useMutation, useQueryCache } from '@pinia/colada'
import { useForm } from '@tanstack/vue-form'
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import {
  createPagesMutation,
  deletePagesMutation,
  getPagesQuery,
  getSettingsQuery,
  listMonitorsQuery,
  previewPageQuery,
  publishPageMutation,
  updatePagesMutation,
  uploadAssetMutation,
} from '../client/@pinia/colada.gen'
import { canEdit } from '../composables/api'
import { notify } from '../composables/notices'
import { formatDate } from '../composables/preferences'
import { errorText } from '../lib/errors'
import { clone } from '../lib/form'
import { publishedEntry } from '../lib/pages'
import AsyncState from './AsyncState.vue'
import Field from './Field.vue'
import Modal from './Modal.vue'
import PageHeader from './PageHeader.vue'
import StatusPage from './StatusPage.vue'
import { Alert } from './ui/alert'
import { Button } from './ui/button'
import { Card } from './ui/card'
import { FieldActions, FieldDescription, FieldGroup, FieldInput, FieldLabel, FieldSection, FieldTextarea } from './ui/field'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from './ui/select'

const { t } = useI18n({ useScope: 'global' })
const queryCache = useQueryCache()
const createPage = useMutation(createPagesMutation())
const updatePage = useMutation(updatePagesMutation())
const publishPage = useMutation(publishPageMutation())
const uploadAsset = useMutation(uploadAssetMutation())
const deletePage = useMutation(deletePagesMutation())

const origin = location.origin
function newID() {
  return crypto.randomUUID?.() || `group-${Date.now().toString(36)}-${Math.random().toString(36).slice(2)}`
}
const route = useRoute<'/app/(admin)/pages/new' | '/app/(admin)/pages/[id]'>()
const router = useRouter()
const id = computed(() => ('id' in route.params ? route.params.id : undefined))
const loading = ref(true)
const error = ref('')
const monitors = ref<Monitor[]>([])
const allowedDomains = ref<string[]>([])
const savedPreview = ref<PublicPage | null>(null)
const deleteOpen = ref(false)
const newMonitorIds = reactive<Record<string, string>>({})
const defaultValues: Page = {
  id: '',
  name: '',
  slug: '',
  domain: '',
  draft: {
    title: '',
    description: '',
    logoUrl: '',
    brandColor: '#2563eb',
    colorScheme: 'system',
    links: [],
    groups: [],
  },
  publishedAt: 0,
  version: 0,
  createdAt: 0,
  updatedAt: 0,
}
const formApi = useForm({
  defaultValues,
  onSubmitMeta: { publish: false },
  validators: {
    onSubmit: ({ value }) => {
      if (!value.name.trim() || !value.draft.title.trim())
        return t('pageEditor.pageNameAndTitleAreRequired')
      if (!/^[a-z0-9][a-z0-9-]*$/.test(value.slug))
        return t('pageEditor.slugMayContainLowercaseLettersNumbersAndHyphens')
      for (const link of value.draft.links) {
        try {
          const parsed = new URL(link.url)
          if (!['https:', 'http:'].includes(parsed.protocol) || !link.label.trim())
            return t('pageEditor.publicLinksNeedLabelsAndHttpSUrls')
        }
        catch {
          return t('pageEditor.publicLinksNeedLabelsAndHttpSUrls')
        }
      }
      if (
        value.draft.logoUrl
        && !/^https:\/\//.test(value.draft.logoUrl)
        && !value.draft.logoUrl.startsWith('/assets/')
      ) {
        return t('pageEditor.logoMustUseHttpsOrAnUploadedAsset')
      }
    },
  },
  onSubmit: async ({ value, meta }) => {
    let canonical: Page | undefined
    try {
      const body = clone(value)
      canonical = await (id.value || value.id
        ? updatePage.mutateAsync({ path: { id: value.id }, body })
        : createPage.mutateAsync({ body }))
      if (meta.publish) {
        await publishPage.mutateAsync({ path: { id: canonical.id } })
        const page = await queryCache.refresh(
          queryCache.ensure({ ...getPagesQuery({ path: { id: canonical.id } }), staleTime: 0 }),
        )
        if (page.status !== 'success')
          throw page.error || new Error(t('errors.requestFailed'))
        canonical = page.data
        notify(t('pageEditor.statusPagePublished'))
      }
      else {
        notify(t('pageEditor.draftSaved'))
      }
      const preview = await queryCache.refresh(
        queryCache.ensure({ ...previewPageQuery({ path: { id: canonical.id } }), staleTime: 0 }),
      )
      if (preview.status !== 'success')
        throw preview.error || new Error(t('errors.requestFailed'))
      savedPreview.value = clone(preview.data)
      if (!id.value)
        await router.replace(`/app/pages/${canonical.id}`)
    }
    catch (e) {
      error.value = errorText(e)
    }
    finally {
      // Reset after all asynchronous work so submission stays locked until done.
      if (canonical)
        replacePage(canonical)
    }
  },
})
const form = formApi.useSelector(state => state.values)
const saving = formApi.useSelector(state => state.isSubmitting)
const validationError = formApi.useSelector(state => state.errors[0])
const displayError = computed(() => error.value || validationError.value || '')
const editing = computed(() => !!id.value || !!form.value.id)
function replacePage(data: Page) {
  // Reset replaces optional publication fields omitted by the canonical response.
  formApi.reset(clone(data))
}
async function load() {
  loading.value = true
  try {
    const [m, s] = await Promise.all([
      queryCache.refresh(queryCache.ensure({ ...listMonitorsQuery(), staleTime: 0 })),
      queryCache.refresh(queryCache.ensure({ ...getSettingsQuery(), staleTime: 0 })),
    ])
    if (m.status !== 'success')
      throw m.error || new Error(t('errors.requestFailed'))
    if (s.status !== 'success')
      throw s.error || new Error(t('errors.requestFailed'))
    monitors.value = clone(m.data.items)
    allowedDomains.value = clone(s.data.allowedDomains)
    if (editing.value) {
      const page = await queryCache.refresh(
        queryCache.ensure({ ...getPagesQuery({ path: { id: id.value! } }), staleTime: 0 }),
      )
      if (page.status !== 'success')
        throw page.error || new Error(t('errors.requestFailed'))
      replacePage(page.data)
      const preview = await queryCache.refresh(
        queryCache.ensure({ ...previewPageQuery({ path: { id: id.value! } }), staleTime: 0 }),
      )
      if (preview.status !== 'success')
        throw preview.error || new Error(t('errors.requestFailed'))
      savedPreview.value = clone(preview.data)
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
const preview = computed<PublicPage>(() => {
  const existing = savedPreview.value?.groups.flatMap(g => g.monitors) || []
  const groups = form.value.draft.groups.map(group => ({
    id: group.id,
    name: group.name,
    monitors: group.monitors.map((pm) => {
      const live = existing.find(x => x.id === pm.monitorId)
      if (live)
        return { ...live, name: pm.alias || live.name, dailyAvailability: pm.showUptime ? live.dailyAvailability : [], latency: pm.showLatency ? live.latency : [] }
      const monitor = monitors.value.find(x => x.id === pm.monitorId)
      return {
        id: pm.monitorId,
        name: pm.alias || monitor?.name || '',
        type: monitor?.type || 'http',
        state: monitor?.state || 'unknown',
        paused: !monitor?.enabled,
        maintenance: false,
        availability: {
          from: 0,
          to: 0,
          upMs: 0,
          downMs: 0,
          unknownMs: 0,
          excludedMs: 0,
          effectiveMs: 0,
          uptime: null,
          coverage: null,
        },
        dailyAvailability: [],
        latency: [],
        ...(monitor?.certificate
          ? {
              certificate: {
                state: monitor.certificate.state,
                expiresAt: monitor.certificate.expiresAt,
                daysRemaining: monitor.certificate.daysRemaining,
              },
            }
          : {}),
      }
    }),
  }))
  const states = groups
    .flatMap(g => g.monitors)
    .filter(m => m.type !== 'certificate' && !m.paused)
    .map(m => (m.maintenance ? 'maintenance' : m.state))
  let state = !states.length
    ? 'unknown'
    : states.every(s => s === 'down')
      ? 'outage'
      : states.includes('down')
        ? 'partial'
        : states.includes('unknown')
          ? 'unknown'
          : states.includes('maintenance')
            ? 'maintenance'
            : 'operational'
  const incidents = savedPreview.value?.incidents || []
  if (incidents.some(i => i.status !== 'resolved' && i.impact === 'outage')) {
    state = 'outage'
  }
  else if (
    state !== 'outage'
    && incidents.some(i => i.status !== 'resolved' && i.impact === 'partial')
  ) {
    state = 'partial'
  }
  return {
    id: form.value.id,
    slug: form.value.slug,
    config: clone(form.value.draft),
    state,
    groups,
    incidents,
    maintenance: savedPreview.value?.maintenance || [],
    updatedAt: savedPreview.value?.updatedAt || form.value.updatedAt,
  }
})
function moveGroup(index: number, delta: number) {
  const target = index + delta
  if (target >= 0 && target < form.value.draft.groups.length)
    formApi.moveFieldValues('draft.groups', index, target)
}
function moveMonitor(groupIndex: number, index: number, delta: number) {
  const target = index + delta
  if (target >= 0 && target < form.value.draft.groups[groupIndex]!.monitors.length)
    formApi.moveFieldValues(`draft.groups[${groupIndex}].monitors`, index, target)
}
function addMonitor(groupId: string) {
  const groupIndex = form.value.draft.groups.findIndex(x => x.id === groupId)
  const monitorId = newMonitorIds[groupId]
  if (groupIndex < 0 || !monitorId)
    return
  if (form.value.draft.groups.some(g => g.monitors.some(m => m.monitorId === monitorId))) {
    notify(t('pageEditor.thisMonitorIsAlreadyOnThePage'), 'error')
    return
  }
  formApi.pushFieldValue(`draft.groups[${groupIndex}].monitors`, {
    monitorId,
    alias: monitors.value.find(m => m.id === monitorId)?.name || '',
    showUptime: true,
    showLatency: true,
  })
  newMonitorIds[groupId] = ''
}
async function save(publish = false) {
  if (saving.value)
    return
  error.value = ''
  await formApi.handleSubmit({ publish })
}
async function uploadLogo(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file)
    return
  if (file.size > 4 * 1024 * 1024) {
    notify(t('pageEditor.imageMustBeSmallerThan4Mib'), 'error')
    return
  }
  try {
    const base64 = await new Promise<string>((resolve, reject) => {
      const reader = new FileReader()
      reader.onload = () => resolve(String(reader.result).split(',')[1] || '')
      reader.onerror = reject
      reader.readAsDataURL(file)
    })
    const asset = await uploadAsset.mutateAsync({
      body: { filename: file.name, contentType: file.type, base64 },
    })
    formApi.setFieldValue('draft.logoUrl', asset.url)
    notify(t('pageEditor.logoUploaded'))
  }
  catch (e) {
    notify(errorText(e), 'error')
  }
}
async function remove() {
  try {
    await deletePage.mutateAsync({ path: { id: form.value.id } })
    notify(t('pageEditor.statusPageDeleted'))
    router.push('/app/pages')
  }
  catch (e) {
    notify(errorText(e), 'error')
  }
}
</script>

<template>
  <PageHeader
    :title="
      editing ? form.name || t('pageEditor.customizeStatusPage') : t('common.createStatusPage')
    "
    :description="t('pageEditor.customizeEachPageIndependentlySaveADraftThen')"
  >
    <Button as-child variant="ghost">
      <RouterLink to="/app/pages">
        <span class="i-lucide-arrow-left" w="14px" h="14px" aria-hidden="true" />{{ t('pageEditor.allPages') }}
      </RouterLink>
    </Button><template v-if="canEdit()">
      <Button :disabled="saving" @click="save()">
        <span class="i-lucide-save" w="14px" h="14px" aria-hidden="true" />{{ t('common.saveDraft') }}
      </Button><Button :disabled="saving" variant="primary" @click="save(true)">
        <span class="i-lucide-send" w="14px" h="14px" aria-hidden="true" />{{ t('pageEditor.publishPage') }}
      </Button>
    </template>
  </PageHeader><AsyncState :pending="loading">
    <Alert v-if="displayError" role="alert" variant="validation">
      {{ displayError }}
    </Alert>
    <div grid="~ cols-[minmax(0,1fr)_minmax(320px,0.8fr)]" gap="24px" class="[@media(max-width:1200px)]:grid-cols-1 [@container_workspace_(max-width:_700px)]:grid-cols-1">
      <div>
        <Card as="section">
          <FieldSection>
            <h2>{{ t('pageEditor.pageAccess') }}</h2>
            <p un-text="13px subtle">
              {{ t('pageEditor.pathAndCustomDomainServeTheSamePublished') }}
            </p>
            <FieldGroup>
              <Field :label="t('pageEditor.internalPageName')">
                <formApi.Field v-slot="{ field }" name="name">
                  <FieldInput :model-value="field.state.value" :disabled="!canEdit()" required @update:model-value="field.handleChange($event)" @blur="field.handleBlur" />
                </formApi.Field>
              </Field><Field
                :label="t('pageEditor.pageSlug')"
                :hint="`${origin}/${form.slug || 'status1'}`"
              >
                <formApi.Field v-slot="{ field }" name="slug">
                  <FieldInput
                    :model-value="field.state.value" :disabled="!canEdit()" placeholder="status1"
                    pattern="[a-z0-9][a-z0-9-]*"
                    required
                    @update:model-value="field.handleChange($event)"
                    @blur="field.handleBlur"
                  />
                </formApi.Field>
              </Field><Field
                class="span-full"
                :label="t('pageEditor.customDomain')"
                :hint="t('pageEditor.selectAnAdministratorConfiguredDomainDnsAndHttps')"
              >
                <formApi.Field v-slot="{ field }" name="domain">
                  <Select :model-value="field.state.value" :disabled="!canEdit()" @focusout="field.handleBlur" @update:model-value="field.handleChange($event)">
                    <SelectTrigger>
                      <SelectValue :placeholder="t('pageEditor.pathAccessOnly')" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectGroup>
                        <SelectItem value="">
                          {{ t('pageEditor.pathAccessOnly') }}
                        </SelectItem>
                        <SelectItem v-for="domain in allowedDomains" :key="domain" :value="domain">
                          {{ domain }}
                        </SelectItem>
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                </formApi.Field>
              </Field>
            </FieldGroup>
            <Alert v-if="form.publishedAt" mt="5" as="p">
              {{ t('pageEditor.lastPublished') }} {{ formatDate(form.publishedAt) }} · v{{
                form.version
              }}<Button as-child variant="ghost" size="sm">
                <a
                  :href="publishedEntry(form).url"

                  target="_blank"
                  rel="noopener"
                ><span class="i-lucide-external-link" w="12px" h="12px" aria-hidden="true" />{{ t('pageEditor.visitPublicPage') }}</a>
              </Button>
            </Alert>
          </FieldSection>
          <FieldSection>
            <h2>{{ t('pageEditor.brandAppearance') }}</h2>
            <p un-text="13px subtle">
              {{ t('pageEditor.theseSettingsApplyOnlyToThisStatusPage') }}
            </p>
            <FieldGroup>
              <Field class="span-full" :label="t('pageEditor.publicTitle')">
                <formApi.Field v-slot="{ field }" name="draft.title">
                  <FieldInput :model-value="field.state.value" :disabled="!canEdit()" required @update:model-value="field.handleChange($event)" @blur="field.handleBlur" />
                </formApi.Field>
              </Field><Field class="span-full" :label="t('pageEditor.pageDescription')">
                <formApi.Field v-slot="{ field }" name="draft.description">
                  <FieldTextarea :model-value="field.state.value" :disabled="!canEdit()" rows="3" @update:model-value="field.handleChange($event)" @blur="field.handleBlur" />
                </formApi.Field>
              </Field><Field class="span-full" :label="t('pageEditor.logoUrl')">
                <div flex="~ items-center gap-10px" class="[@media(max-width:700px)]:flex-wrap">
                  <formApi.Field v-slot="{ field }" name="draft.logoUrl">
                    <FieldInput
                      :model-value="field.state.value" :disabled="!canEdit()" :placeholder="t('pageEditor.logoPlaceholder')"
                      @update:model-value="field.handleChange($event)"
                      @blur="field.handleBlur"
                    />
                  </formApi.Field><Button v-if="canEdit()" as-child>
                    <label><span class="i-lucide-upload" w="14px" h="14px" aria-hidden="true" />{{ t('pageEditor.upload')
                    }}<input
                      type="file"
                      :aria-label="t('pageEditor.upload')"
                      accept="image/png,image/jpeg,image/gif"
                      hidden=""
                      @change="uploadLogo"
                    ></label>
                  </Button>
                </div>
              </Field><Field :label="t('pageEditor.brandColor')">
                <div flex="~ items-center gap-10px" class="[@media(max-width:700px)]:flex-wrap">
                  <formApi.Field v-slot="{ field }" name="draft.brandColor">
                    <input
                      :value="field.state.value" type="color" :disabled="!canEdit()" :aria-label="t('pageEditor.brandColor')"
                      @input="field.handleChange(($event.target as HTMLInputElement).value)"
                      @blur="field.handleBlur"
                    >

                    <FieldInput
                      :model-value="field.state.value" :disabled="!canEdit()" pattern="#[0-9a-fA-F]{6}"
                      @update:model-value="field.handleChange($event)"
                      @blur="field.handleBlur"
                    />
                  </formApi.Field>
                </div>
              </Field><Field :label="t('pageEditor.colorScheme')">
                <formApi.Field v-slot="{ field }" name="draft.colorScheme">
                  <Select :model-value="field.state.value" :disabled="!canEdit()" @focusout="field.handleBlur" @update:model-value="field.handleChange($event as Page['draft']['colorScheme'])">
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectGroup>
                        <SelectItem value="system">
                          {{ t('common.system') }}
                        </SelectItem>
                        <SelectItem value="light">
                          {{ t('common.light') }}
                        </SelectItem>
                        <SelectItem value="dark">
                          {{ t('common.dark') }}
                        </SelectItem>
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                </formApi.Field>
              </Field>
              <div class="span-full">
                <FieldLabel as="label">
                  {{ t('pageEditor.publicLinks') }}
                </FieldLabel>
                <div
                  v-for="(_, index) in form.draft.links"
                  :key="index"
                  grid="~ cols-[1fr_2fr_auto]" gap="9px" mb="9px" class="[@media(max-width:700px)]:grid-cols-1"
                  mt="3"
                >
                  <formApi.Field :key="`draft.links[${index}].label`" v-slot="{ field }" :name="`draft.links[${index}].label`">
                    <input
                      :value="field.state.value" :disabled="!canEdit()" :placeholder="t('pageEditor.linkLabel')" :aria-label="t('pageEditor.linkLabel')"
                      @input="field.handleChange(($event.target as HTMLInputElement).value)"
                      @blur="field.handleBlur"
                    >
                  </formApi.Field><formApi.Field :key="`draft.links[${index}].url`" v-slot="{ field }" :name="`draft.links[${index}].url`">
                    <input
                      :value="field.state.value" :disabled="!canEdit()" type="url"
                      aria-label="URL" placeholder="https://…"
                      @input="field.handleChange(($event.target as HTMLInputElement).value)"
                      @blur="field.handleBlur"
                    >
                  </formApi.Field><Button
                    v-if="canEdit()"
                    :aria-label="t('pageEditor.removeLink')"
                    shape="square"
                    @click="formApi.removeFieldValue('draft.links', index)"
                  >
                    <span class="i-lucide-x" w="14px" h="14px" aria-hidden="true" />
                  </Button>
                </div>
                <Button
                  v-if="canEdit()"
                  mt="2"
                  variant="ghost"
                  size="sm"
                  @click="formApi.pushFieldValue('draft.links', { label: '', url: '' })"
                >
                  <span class="i-lucide-plus" w="13px" h="13px" aria-hidden="true" />{{ t('pageEditor.addLink') }}
                </Button>
              </div>
            </FieldGroup>
          </FieldSection>
          <FieldSection>
            <h2>{{ t('pageEditor.servicesGroups') }}</h2>
            <p un-text="13px subtle">
              {{ t('pageEditor.publishOnlySelectedMonitorsPublicAliasesLeaveInternal') }}
            </p>
            <div v-for="(group, index) in form.draft.groups" :key="group.id" class="group-editor" border="1px solid line" rounded="9px" mt="15px" overflow="hidden">
              <div class="group-editor-header" flex="~ items-center gap-8px" p="13px" bg="tint" border="b-1px b-solid b-line">
                <formApi.Field :key="`draft.groups[${index}].name`" v-slot="{ field }" :name="`draft.groups[${index}].name`">
                  <input
                    :value="field.state.value" :aria-label="t('common.groupName')" :disabled="!canEdit()"
                    :placeholder="t('common.groupName')"
                    @input="field.handleChange(($event.target as HTMLInputElement).value)"
                    @blur="field.handleBlur"
                  >
                </formApi.Field><template v-if="canEdit()">
                  <Button
                    :disabled="index === 0"
                    :aria-label="t('pageEditor.moveGroupUp')"
                    shape="square"
                    @click="moveGroup(index, -1)"
                  >
                    <span class="i-lucide-arrow-up" w="13px" h="13px" aria-hidden="true" />
                  </Button><Button
                    :disabled="index === form.draft.groups.length - 1"
                    :aria-label="t('pageEditor.moveGroupDown')"
                    shape="square"
                    @click="moveGroup(index, 1)"
                  >
                    <span class="i-lucide-arrow-down" w="13px" h="13px" aria-hidden="true" />
                  </Button><Button
                    :aria-label="t('pageEditor.removeGroup')"
                    shape="square"
                    @click="formApi.removeFieldValue('draft.groups', index)"
                  >
                    <span class="i-lucide-x" w="14px" h="14px" aria-hidden="true" />
                  </Button>
                </template>
              </div>
              <div
                v-for="(item, mIndex) in group.monitors"
                :key="item.monitorId"
                flex="~ items-center gap-9px" px="13px" py="11px" border="b-1px b-solid b-line last:0" class="[@media(max-width:700px)]:flex-wrap"
              >
                <div flex="1" min-w="0">
                  <span un-text="12px subtle" tracking="0.5px">{{
                    monitors.find((x) => x.id === item.monitorId)?.name
                  }}</span><formApi.Field :key="`draft.groups[${index}].monitors[${mIndex}].alias`" v-slot="{ field }" :name="`draft.groups[${index}].monitors[${mIndex}].alias`">
                    <input
                      :value="field.state.value" flex="1" min-w="0"
                      :aria-label="t('common.publicAlias')"
                      :disabled="!canEdit()"
                      :placeholder="t('common.publicAlias')"
                      mt="1"
                      @input="field.handleChange(($event.target as HTMLInputElement).value)"
                      @blur="field.handleBlur"
                    >
                  </formApi.Field>
                  <div flex="~ gap-4" mt="2">
                    <label flex="~ items-center gap-8px" un-text="12px default"><formApi.Field :key="`draft.groups[${index}].monitors[${mIndex}].showUptime`" v-slot="{ field }" :name="`draft.groups[${index}].monitors[${mIndex}].showUptime`">
                      <input :checked="field.state.value" flex="1" min-w="0" type="checkbox" :disabled="!canEdit()" @change="field.handleChange(($event.target as HTMLInputElement).checked)" @blur="field.handleBlur">
                    </formApi.Field>{{
                      t('common.uptime')
                    }}</label><label flex="~ items-center gap-8px" un-text="12px default"><formApi.Field :key="`draft.groups[${index}].monitors[${mIndex}].showLatency`" v-slot="{ field }" :name="`draft.groups[${index}].monitors[${mIndex}].showLatency`">
                      <input :checked="field.state.value" flex="1" min-w="0" type="checkbox" :disabled="!canEdit()" @change="field.handleChange(($event.target as HTMLInputElement).checked)" @blur="field.handleBlur">
                    </formApi.Field>{{
                      t('pageEditor.latency')
                    }}</label>
                  </div>
                </div>
                <template v-if="canEdit()">
                  <Button
                    :disabled="mIndex === 0"
                    :aria-label="t('pageEditor.moveServiceUp')"
                    shape="square"
                    @click="moveMonitor(index, mIndex, -1)"
                  >
                    <span class="i-lucide-arrow-up" w="13px" h="13px" aria-hidden="true" />
                  </Button><Button
                    :disabled="mIndex === group.monitors.length - 1"
                    :aria-label="t('pageEditor.moveServiceDown')"
                    shape="square"
                    @click="moveMonitor(index, mIndex, 1)"
                  >
                    <span class="i-lucide-arrow-down" w="13px" h="13px" aria-hidden="true" />
                  </Button><Button
                    :aria-label="t('pageEditor.removeService')"
                    shape="square"
                    @click="formApi.removeFieldValue(`draft.groups[${index}].monitors`, mIndex)"
                  >
                    <span class="i-lucide-x" w="14px" h="14px" aria-hidden="true" />
                  </Button>
                </template>
              </div>
              <div v-if="canEdit()" flex="~ items-center gap-9px" px="13px" py="11px" border="b-1px b-solid b-line last:0" class="[@media(max-width:700px)]:flex-wrap">
                <Select v-model="newMonitorIds[group.id]">
                  <SelectTrigger :aria-label="t('common.selectAMonitor')">
                    <SelectValue :placeholder="t('common.selectAMonitor')" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectGroup>
                      <SelectItem value="">
                        {{ t('common.selectAMonitor') }}
                      </SelectItem>
                      <SelectItem v-for="monitor in monitors" :key="monitor.id" :value="monitor.id">
                        {{ monitor.name }} · {{ monitor.type }}
                      </SelectItem>
                    </SelectGroup>
                  </SelectContent>
                </Select><Button size="sm" @click="addMonitor(group.id)">
                  <span class="i-lucide-plus" w="13px" h="13px" aria-hidden="true" />{{ t('pageEditor.add') }}
                </Button>
              </div>
            </div>
            <Button
              v-if="canEdit()"
              mt="4"
              size="sm"
              @click="
                formApi.pushFieldValue('draft.groups', {
                  id: newID(),
                  name: t('pageEditor.services'),
                  monitors: [],
                })
              "
            >
              <span class="i-lucide-plus" w="13px" h="13px" aria-hidden="true" />{{ t('pageEditor.addGroup') }}
            </Button>
          </FieldSection>
        </Card>
        <FieldActions>
          <Button
            v-if="editing && canEdit()"
            mr="auto"
            variant="destructive"
            @click="deleteOpen = true"
          >
            <span class="i-lucide-trash-2" w="13px" h="13px" aria-hidden="true" />{{ t('pageEditor.deletePage') }}
          </Button><Button v-if="canEdit()" :disabled="saving" variant="primary" @click="save()">
            <span class="i-lucide-save" w="14px" h="14px" aria-hidden="true" />{{ t('common.saveDraft') }}
          </Button>
        </FieldActions>
      </div>
      <aside pos="sticky" top="20px" self="start" class="[@media(max-width:1200px)]:static [@media(max-width:700px)]:min-w-0">
        <div class="preview-label" flex="~ items-center justify-between" un-text="12px subtle" mb="12px">
          <strong>{{ t('pageEditor.liveDraftPreview') }}</strong><span>{{ t('pageEditor.responsivePreview') }}</span>
        </div>
        <StatusPage :page="preview" preview />
        <FieldDescription as="p" mt="3">
          {{ t('pageEditor.savingRefreshesRealStatisticsNewlySelectedServicesShow') }}
        </FieldDescription>
      </aside>
    </div>
  </AsyncState><Modal
    v-model:open="deleteOpen"
    :title="t('pageEditor.deleteStatusPage')"
    :description="t('pageEditor.thePagePathAndDomainWillStopPublishing')"
  >
    <template #footer>
      <Button @click="deleteOpen = false">
        {{ t('common.cancel') }}
      </Button><Button variant="destructive" @click="remove">
        {{ t('common.delete') }}
      </Button>
    </template>
  </Modal>
</template>
