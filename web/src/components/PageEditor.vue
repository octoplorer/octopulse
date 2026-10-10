<script setup lang="ts">
import type { Monitor, Page, PublicPage } from '../client/types.gen'
import { useMutation, useQueryCache } from '@pinia/colada'
import { useForm } from '@tanstack/vue-form'
import { computed, nextTick, onMounted, reactive, ref } from 'vue'
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
import PageEditorWorkspace from './PageEditorWorkspace.vue'
import StatusPage from './StatusPage.vue'
import { Banner } from './ui/banner'
import { PageHeader } from './ui/blocks/page-header'
import { Button } from './ui/button'
import { Checkbox } from './ui/checkbox'
import { Dialog } from './ui/dialog'
import { Field, FieldActions, FieldDescription, FieldGroup, FieldLabel, FieldSection } from './ui/field'
import { Input, InputArea } from './ui/input'
import { LayerCard, LayerCardPrimary } from './ui/layer-card'
import { Loader } from './ui/loader'
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
const loadError = ref('')
const error = ref('')
const submitted = ref(false)
const workspace = ref<InstanceType<typeof PageEditorWorkspace>>()
const errorElement = ref<HTMLElement>()
const logoInput = ref<HTMLInputElement>()
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
function pageFieldErrors(value: Page) {
  const errors: Record<string, string> = {}
  if (!value.name.trim())
    errors.name = t('page-editor.page-name-and-title-are-required')
  if (!/^[a-z0-9][a-z0-9-]*$/.test(value.slug))
    errors.slug = t('page-editor.slug-may-contain-lowercase-letters-numbers-and-hyphens')
  if (!value.draft.title.trim())
    errors['draft.title'] = t('page-editor.page-name-and-title-are-required')
  if (value.draft.logoUrl && !/^https:\/\//.test(value.draft.logoUrl) && !value.draft.logoUrl.startsWith('/assets/'))
    errors['draft.logoUrl'] = t('page-editor.logo-must-use-https-or-an-uploaded-asset')
  value.draft.links.forEach((link, index) => {
    if (!link.label.trim())
      errors[`draft.links[${index}].label`] = t('page-editor.public-links-need-labels-and-http-s-urls')
    try {
      if (!['https:', 'http:'].includes(new URL(link.url).protocol))
        errors[`draft.links[${index}].url`] = t('page-editor.public-links-need-labels-and-http-s-urls')
    }
    catch {
      errors[`draft.links[${index}].url`] = t('page-editor.public-links-need-labels-and-http-s-urls')
    }
  })
  return errors
}
const formApi = useForm({
  defaultValues,
  onSubmitMeta: { publish: false },
  validators: {
    onSubmit: ({ value }) => Object.values(pageFieldErrors(value))[0],
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
          throw page.error || new Error(t('errors.request-failed'))
        canonical = page.data
        notify(t('page-editor.status-page-published'))
      }
      else {
        notify(t('page-editor.draft-saved'))
      }
      const preview = await queryCache.refresh(
        queryCache.ensure({ ...previewPageQuery({ path: { id: canonical.id } }), staleTime: 0 }),
      )
      if (preview.status !== 'success')
        throw preview.error || new Error(t('errors.request-failed'))
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
const fieldErrors = computed(() => submitted.value ? pageFieldErrors(form.value) : {})
const displayError = computed(() => error.value || Object.values(fieldErrors.value)[0] || '')
const editing = computed(() => !!id.value || !!form.value.id)
function replacePage(data: Page) {
  // Reset replaces optional publication fields omitted by the canonical response.
  formApi.reset(clone(data))
}
async function load() {
  loading.value = true
  loadError.value = ''
  try {
    const [m, s] = await Promise.all([
      queryCache.refresh(queryCache.ensure({ ...listMonitorsQuery(), staleTime: 0 })),
      queryCache.refresh(queryCache.ensure({ ...getSettingsQuery(), staleTime: 0 })),
    ])
    if (m.status !== 'success')
      throw m.error || new Error(t('errors.request-failed'))
    if (s.status !== 'success')
      throw s.error || new Error(t('errors.request-failed'))
    monitors.value = clone(m.data.items)
    allowedDomains.value = clone(s.data.allowedDomains)
    if (editing.value) {
      const page = await queryCache.refresh(
        queryCache.ensure({ ...getPagesQuery({ path: { id: id.value! } }), staleTime: 0 }),
      )
      if (page.status !== 'success')
        throw page.error || new Error(t('errors.request-failed'))
      replacePage(page.data)
      const preview = await queryCache.refresh(
        queryCache.ensure({ ...previewPageQuery({ path: { id: id.value! } }), staleTime: 0 }),
      )
      if (preview.status !== 'success')
        throw preview.error || new Error(t('errors.request-failed'))
      savedPreview.value = clone(preview.data)
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
    notify(t('page-editor.this-monitor-is-already-on-the-page'), 'error')
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
  if (saving.value || loading.value || loadError.value)
    return
  error.value = ''
  submitted.value = true
  await formApi.handleSubmit({ publish })
  const invalidField = Object.keys(fieldErrors.value)[0]
  if (invalidField) {
    await workspace.value?.focusEditor(invalidField)
  }
  else if (error.value) {
    await nextTick()
    errorElement.value?.focus()
  }
}
async function uploadLogo(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file)
    return
  if (file.size > 4 * 1024 * 1024) {
    notify(t('page-editor.image-must-be-smaller-than-4-mib'), 'error')
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
    notify(t('page-editor.logo-uploaded'))
  }
  catch (e) {
    notify(errorText(e), 'error')
  }
}
async function remove() {
  try {
    await deletePage.mutateAsync({ path: { id: form.value.id } })
    notify(t('page-editor.status-page-deleted'))
    router.push('/app/pages')
  }
  catch (e) {
    notify(errorText(e), 'error')
  }
}
</script>

<template>
  <PageHeader
    class="mb-6"
    :title="
      editing ? form.name || t('page-editor.customize-status-page') : t('common.create-status-page')
    "
    :description="t('page-editor.customize-each-page-independently-save-a-draft-then')"
  >
    <template #actions>
      <Button as-child variant="ghost">
        <RouterLink to="/app/pages">
          <span class="i-lucide-arrow-left" w="14px" h="14px" aria-hidden="true" />{{ t('page-editor.all-pages') }}
        </RouterLink>
      </Button><template v-if="canEdit()">
        <Button :disabled="saving || loading || !!loadError" @click="save()">
          <span class="i-lucide-save" w="14px" h="14px" aria-hidden="true" />{{ t('common.save-draft') }}
        </Button><Button :disabled="saving || loading || !!loadError" variant="primary" @click="save(true)">
          <span class="i-lucide-send" w="14px" h="14px" aria-hidden="true" />{{ t('page-editor.publish-page') }}
        </Button>
      </template>
    </template>
  </PageHeader><Loader v-if="loading" :label="t('async-state.loading-data')" class="flex! w-full justify-center p-15 text-size-xs">
    {{ t('async-state.loading-data') }}
  </Loader>
  <Banner v-else-if="loadError" variant="error">
    {{ loadError }}
    <Button class="mt-3" @click="load">
      {{ t('async-state.retry') }}
    </Button>
  </Banner>
  <template v-else>
    <div v-if="displayError" ref="errorElement" tabindex="-1" class="mb-5">
      <Banner size="sm" variant="error" role="alert">
        {{ displayError }}
      </Banner>
    </div>
    <PageEditorWorkspace ref="workspace">
      <template #editor>
        <LayerCard as-child>
          <section>
            <LayerCardPrimary class="p-0!">
              <FieldSection>
                <h2>{{ t('page-editor.page-access') }}</h2>
                <p un-text="13px subtle">
                  {{ t('page-editor.path-and-custom-domain-serve-the-same-published') }}
                </p>
                <FieldGroup>
                  <Field :label="t('page-editor.internal-page-name')" required :error="fieldErrors.name">
                    <formApi.Field v-slot="{ field }" name="name">
                      <Input :name="field.name" :model-value="field.state.value" :disabled="!canEdit()" required @update:model-value="field.handleChange(String($event ?? ''))" @blur="field.handleBlur" />
                    </formApi.Field>
                  </Field><Field
                    :label="t('page-editor.page-slug')"
                    :description="`${origin}/${form.slug || 'status1'}`"
                    required :error="fieldErrors.slug"
                  >
                    <formApi.Field v-slot="{ field }" name="slug">
                      <Input
                        :name="field.name" :model-value="field.state.value" :disabled="!canEdit()" placeholder="status1"
                        pattern="[a-z0-9][a-z0-9-]*"
                        required
                        @update:model-value="field.handleChange(String($event ?? ''))"
                        @blur="field.handleBlur"
                      />
                    </formApi.Field>
                  </Field><Field
                    class="span-full"
                    :label="t('page-editor.custom-domain')"
                    :description="t('page-editor.select-an-administrator-configured-domain-dns-and-https')"
                  >
                    <formApi.Field v-slot="{ field }" name="domain">
                      <Select :model-value="field.state.value" :disabled="!canEdit()" @focusout="field.handleBlur" @update:model-value="field.handleChange($event)">
                        <SelectTrigger>
                          <SelectValue :placeholder="t('page-editor.path-access-only')" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectGroup>
                            <SelectItem value="">
                              {{ t('page-editor.path-access-only') }}
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
                <Banner v-if="form.publishedAt" size="sm" variant="secondary" mt="5">
                  {{ t('page-editor.last-published') }} {{ formatDate(form.publishedAt) }} · v{{
                    form.version
                  }}<Button as-child variant="ghost" size="sm">
                    <a
                      :href="publishedEntry(form).url"

                      target="_blank"
                      rel="noopener"
                    ><span class="i-lucide-external-link" w="12px" h="12px" aria-hidden="true" />{{ t('page-editor.visit-public-page') }}</a>
                  </Button>
                </Banner>
              </FieldSection>
              <FieldSection>
                <h2>{{ t('page-editor.brand-appearance') }}</h2>
                <p un-text="13px subtle">
                  {{ t('page-editor.these-settings-apply-only-to-this-status-page') }}
                </p>
                <FieldGroup>
                  <Field class="span-full" :label="t('page-editor.public-title')" required :error="fieldErrors['draft.title']">
                    <formApi.Field v-slot="{ field }" name="draft.title">
                      <Input :name="field.name" :model-value="field.state.value" :disabled="!canEdit()" required @update:model-value="field.handleChange(String($event ?? ''))" @blur="field.handleBlur" />
                    </formApi.Field>
                  </Field><Field class="span-full" :label="t('page-editor.page-description')">
                    <formApi.Field v-slot="{ field }" name="draft.description">
                      <InputArea :model-value="field.state.value" :disabled="!canEdit()" rows="3" @update:model-value="field.handleChange(String($event ?? ''))" @blur="field.handleBlur" />
                    </formApi.Field>
                  </Field><Field class="span-full" :label="t('page-editor.logo-url')" :error="fieldErrors['draft.logoUrl']">
                    <div flex="~ wrap items-center gap-10px">
                      <formApi.Field v-slot="{ field }" name="draft.logoUrl">
                        <Input
                          :name="field.name" :model-value="field.state.value" :disabled="!canEdit()" :placeholder="t('page-editor.logo-placeholder')" class="min-w-40 flex-1"
                          @update:model-value="field.handleChange(String($event ?? ''))"
                          @blur="field.handleBlur"
                        />
                      </formApi.Field><Button v-if="canEdit()" @click="logoInput?.click()">
                        <span class="i-lucide-upload" w="14px" h="14px" aria-hidden="true" />{{ t('page-editor.upload') }}
                      </Button>
                      <input
                        v-if="canEdit()" ref="logoInput" type="file"
                        :aria-label="t('page-editor.upload')" accept="image/png,image/jpeg,image/gif"
                        hidden @change="uploadLogo"
                      >
                    </div>
                  </Field><Field :label="t('page-editor.brand-color')">
                    <div flex="~ items-center gap-10px">
                      <formApi.Field v-slot="{ field }" name="draft.brandColor">
                        <Input
                          id="page-brand-color-picker" :model-value="field.state.value" type="color" class="size-10! shrink-0 p-1!" :disabled="!canEdit()" :aria-label="t('page-editor.brand-color')"
                          @update:model-value="field.handleChange(String($event ?? ''))"
                          @blur="field.handleBlur"
                        />

                        <Input
                          :model-value="field.state.value" :disabled="!canEdit()" pattern="#[0-9a-fA-F]{6}"
                          @update:model-value="field.handleChange(String($event ?? ''))"
                          @blur="field.handleBlur"
                        />
                      </formApi.Field>
                    </div>
                  </Field><Field :label="t('page-editor.color-scheme')">
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
                    <FieldLabel as="div">
                      {{ t('page-editor.public-links') }}
                    </FieldLabel>
                    <div
                      v-for="(_, index) in form.draft.links"
                      :key="index"
                      grid="~ cols-[minmax(0,1fr)_minmax(0,2fr)_auto]" gap="9px" mb="9px" class="public-link-row"
                      mt="3"
                    >
                      <formApi.Field :key="`draft.links[${index}].label`" v-slot="{ field }" :name="`draft.links[${index}].label`">
                        <Field :error="fieldErrors[field.name]">
                          <Input
                            :name="field.name" :model-value="field.state.value" :disabled="!canEdit()" :placeholder="t('page-editor.link-label')" :aria-label="t('page-editor.link-label')"
                            @update:model-value="field.handleChange(String($event ?? ''))"
                            @blur="field.handleBlur"
                          />
                        </Field>
                      </formApi.Field><formApi.Field :key="`draft.links[${index}].url`" v-slot="{ field }" :name="`draft.links[${index}].url`">
                        <Field :error="fieldErrors[field.name]">
                          <Input
                            :name="field.name" :model-value="field.state.value" :disabled="!canEdit()" type="url"
                            aria-label="URL" placeholder="https://…"
                            @update:model-value="field.handleChange(String($event ?? ''))"
                            @blur="field.handleBlur"
                          />
                        </Field>
                      </formApi.Field><Button
                        v-if="canEdit()"
                        :aria-label="t('page-editor.remove-link')"
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
                      <span class="i-lucide-plus" w="13px" h="13px" aria-hidden="true" />{{ t('page-editor.add-link') }}
                    </Button>
                  </div>
                </FieldGroup>
              </FieldSection>
              <FieldSection>
                <h2>{{ t('page-editor.services-groups') }}</h2>
                <p un-text="13px subtle">
                  {{ t('page-editor.publish-only-selected-monitors-public-aliases-leave-internal') }}
                </p>
                <div v-for="(group, index) in form.draft.groups" :key="group.id" class="group-editor" border="1px solid line" rounded="9px" mt="15px" overflow="hidden">
                  <div class="group-editor-header" flex="~ wrap items-center gap-8px" p="13px" bg="tint" border="b-1px b-solid b-line">
                    <formApi.Field :key="`draft.groups[${index}].name`" v-slot="{ field }" :name="`draft.groups[${index}].name`">
                      <Input
                        :model-value="field.state.value" :aria-label="t('common.group-name')" :disabled="!canEdit()" class="min-w-36 flex-1"
                        :placeholder="t('common.group-name')"
                        @update:model-value="field.handleChange(String($event ?? ''))"
                        @blur="field.handleBlur"
                      />
                    </formApi.Field><template v-if="canEdit()">
                      <Button
                        :disabled="index === 0"
                        :aria-label="t('page-editor.move-group-up')"
                        shape="square"
                        @click="moveGroup(index, -1)"
                      >
                        <span class="i-lucide-arrow-up" w="13px" h="13px" aria-hidden="true" />
                      </Button><Button
                        :disabled="index === form.draft.groups.length - 1"
                        :aria-label="t('page-editor.move-group-down')"
                        shape="square"
                        @click="moveGroup(index, 1)"
                      >
                        <span class="i-lucide-arrow-down" w="13px" h="13px" aria-hidden="true" />
                      </Button><Button
                        :aria-label="t('page-editor.remove-group')"
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
                    flex="~ wrap items-center gap-9px" px="13px" py="11px" border="b-1px b-solid b-line last:0"
                  >
                    <div flex="1" class="min-w-40">
                      <span un-text="12px subtle" tracking="0.5px">{{
                        monitors.find((x) => x.id === item.monitorId)?.name
                      }}</span><formApi.Field :key="`draft.groups[${index}].monitors[${mIndex}].alias`" v-slot="{ field }" :name="`draft.groups[${index}].monitors[${mIndex}].alias`">
                        <Input
                          :model-value="field.state.value" flex="1" min-w="0"
                          :aria-label="t('common.public-alias')"
                          :disabled="!canEdit()"
                          :placeholder="t('common.public-alias')"
                          mt="1"
                          @update:model-value="field.handleChange(String($event ?? ''))"
                          @blur="field.handleBlur"
                        />
                      </formApi.Field>
                      <div flex="~ wrap gap-4" mt="2">
                        <formApi.Field :key="`draft.groups[${index}].monitors[${mIndex}].showUptime`" v-slot="{ field }" :name="`draft.groups[${index}].monitors[${mIndex}].showUptime`">
                          <Checkbox :model-value="field.state.value" :label="t('common.uptime')" :disabled="!canEdit()" @update:model-value="field.handleChange($event === true)" @focusout="field.handleBlur" />
                        </formApi.Field><formApi.Field :key="`draft.groups[${index}].monitors[${mIndex}].showLatency`" v-slot="{ field }" :name="`draft.groups[${index}].monitors[${mIndex}].showLatency`">
                          <Checkbox :model-value="field.state.value" :label="t('page-editor.latency')" :disabled="!canEdit()" @update:model-value="field.handleChange($event === true)" @focusout="field.handleBlur" />
                        </formApi.Field>
                      </div>
                    </div>
                    <template v-if="canEdit()">
                      <Button
                        :disabled="mIndex === 0"
                        :aria-label="t('page-editor.move-service-up')"
                        shape="square"
                        @click="moveMonitor(index, mIndex, -1)"
                      >
                        <span class="i-lucide-arrow-up" w="13px" h="13px" aria-hidden="true" />
                      </Button><Button
                        :disabled="mIndex === group.monitors.length - 1"
                        :aria-label="t('page-editor.move-service-down')"
                        shape="square"
                        @click="moveMonitor(index, mIndex, 1)"
                      >
                        <span class="i-lucide-arrow-down" w="13px" h="13px" aria-hidden="true" />
                      </Button><Button
                        :aria-label="t('page-editor.remove-service')"
                        shape="square"
                        @click="formApi.removeFieldValue(`draft.groups[${index}].monitors`, mIndex)"
                      >
                        <span class="i-lucide-x" w="14px" h="14px" aria-hidden="true" />
                      </Button>
                    </template>
                  </div>
                  <div v-if="canEdit()" flex="~ wrap items-center gap-9px" px="13px" py="11px" border="b-1px b-solid b-line last:0">
                    <Select v-model="newMonitorIds[group.id]">
                      <SelectTrigger :aria-label="t('common.select-a-monitor')">
                        <SelectValue :placeholder="t('common.select-a-monitor')" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectGroup>
                          <SelectItem value="">
                            {{ t('common.select-a-monitor') }}
                          </SelectItem>
                          <SelectItem v-for="monitor in monitors" :key="monitor.id" :value="monitor.id">
                            {{ monitor.name }} · {{ monitor.type }}
                          </SelectItem>
                        </SelectGroup>
                      </SelectContent>
                    </Select><Button size="sm" @click="addMonitor(group.id)">
                      <span class="i-lucide-plus" w="13px" h="13px" aria-hidden="true" />{{ t('page-editor.add') }}
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
                      name: t('page-editor.services'),
                      monitors: [],
                    })
                  "
                >
                  <span class="i-lucide-plus" w="13px" h="13px" aria-hidden="true" />{{ t('page-editor.add-group') }}
                </Button>
              </FieldSection>
            </LayerCardPrimary>
          </section>
        </LayerCard>
      </template>
      <template #preview>
        <div class="preview-label mb-3 flex flex-wrap items-center justify-between gap-2 text-size-xs text-subtle">
          <strong>{{ t('page-editor.live-draft-preview') }}</strong><span>{{ t('page-editor.responsive-preview') }}</span>
        </div>
        <StatusPage :page="preview" preview />
        <FieldDescription mt="3">
          {{ t('page-editor.saving-refreshes-real-statistics-newly-selected-services-show') }}
        </FieldDescription>
      </template>
    </PageEditorWorkspace>
    <FieldActions v-if="canEdit()" class="sticky bottom-0 z-20 mt-6! rounded-xl border border-solid border-line bg-canvas p-4 shadow-control">
      <Button
        v-if="editing && canEdit()"
        mr="auto"
        variant="secondary-destructive"
        @click="deleteOpen = true"
      >
        <span class="i-lucide-trash-2" w="13px" h="13px" aria-hidden="true" />{{ t('page-editor.delete-page') }}
      </Button><Button v-if="canEdit()" :disabled="saving" @click="save()">
        <span class="i-lucide-save" w="14px" h="14px" aria-hidden="true" />{{ t('common.save-draft') }}
      </Button>
      <Button v-if="canEdit()" :disabled="saving" variant="primary" @click="save(true)">
        <span class="i-lucide-send size-4" aria-hidden="true" />{{ t('page-editor.publish-page') }}
      </Button>
    </FieldActions>
  </template><Dialog
    v-model:open="deleteOpen" role="alertdialog"
    :close-label="t('modal.close-dialog')"
    :title="t('page-editor.delete-status-page')"
    :description="t('page-editor.the-page-path-and-domain-will-stop-publishing')"
  >
    <template #footer>
      <Button @click="deleteOpen = false">
        {{ t('common.cancel') }}
      </Button><Button variant="destructive" @click="remove">
        {{ t('common.delete') }}
      </Button>
    </template>
  </Dialog>
</template>

<style scoped>
@container status-page-editor (max-width: 600px) {
  .public-link-row {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
