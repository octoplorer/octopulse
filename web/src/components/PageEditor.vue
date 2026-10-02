<script setup lang="ts">
import type { Monitor, Page, PublicPage, Settings } from '../lib/types'
import { useMutation, useQueryCache } from '@pinia/colada'
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
const saving = ref(false)
const error = ref('')
const monitors = ref<Monitor[]>([])
const allowedDomains = ref<string[]>([])
const savedPreview = ref<PublicPage | null>(null)
const deleteOpen = ref(false)
const newMonitorIds = reactive<Record<string, string>>({})
const form = reactive<Page>({
  id: '',
  name: '',
  slug: '',
  domain: '',
  draft: {
    title: '',
    description: '',
    logoUrl: '',
    brandColor: '#0c8b76',
    colorScheme: 'system',
    links: [],
    groups: [],
  },
  publishedAt: 0,
  version: 0,
  createdAt: 0,
  updatedAt: 0,
})
const editing = computed(() => !!id.value || !!form.id)
function replacePage(data: Page) {
  // Empty optional publication fields are omitted by Go. Clear an earlier
  // snapshot before applying the canonical response, including domain removal.
  Object.assign(
    form,
    { publishedSlug: undefined, publishedDomain: undefined, published: undefined },
    clone(data),
  )
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
    monitors.value = clone(m.data.items) as Monitor[]
    allowedDomains.value = clone((s.data as Settings).allowedDomains || [])
    if (editing.value) {
      const page = await queryCache.refresh(
        queryCache.ensure({ ...getPagesQuery({ path: { id: id.value! } }), staleTime: 0 }),
      )
      if (page.status !== 'success')
        throw page.error || new Error(t('errors.requestFailed'))
      replacePage(page.data as Page)
      const preview = await queryCache.refresh(
        queryCache.ensure({ ...previewPageQuery({ path: { id: id.value! } }), staleTime: 0 }),
      )
      if (preview.status !== 'success')
        throw preview.error || new Error(t('errors.requestFailed'))
      savedPreview.value = clone(preview.data) as PublicPage
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
  const groups = form.draft.groups.map(group => ({
    id: group.id,
    name: group.name,
    monitors: group.monitors.map((pm) => {
      const live = existing.find(x => x.id === pm.monitorId)
      if (live)
        return { ...live, name: pm.alias || live.name, latency: pm.showLatency ? live.latency : [] }
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
    id: form.id,
    slug: form.slug,
    config: clone(form.draft),
    state,
    groups,
    incidents,
    maintenance: savedPreview.value?.maintenance || [],
    updatedAt: savedPreview.value?.updatedAt || form.updatedAt,
  }
})
function move<T>(values: T[], index: number, delta: number) {
  const target = index + delta
  if (target < 0 || target >= values.length)
    return
  const [value] = values.splice(index, 1)
  values.splice(target, 0, value!)
}
function addMonitor(groupId: string) {
  const group = form.draft.groups.find(x => x.id === groupId)
  const id = newMonitorIds[groupId]
  if (!group || !id)
    return
  if (form.draft.groups.some(g => g.monitors.some(m => m.monitorId === id))) {
    notify(t('pageEditor.thisMonitorIsAlreadyOnThePage'), 'error')
    return
  }
  group.monitors.push({
    monitorId: id,
    alias: monitors.value.find(m => m.id === id)?.name || '',
    showUptime: true,
    showLatency: true,
  })
  newMonitorIds[groupId] = ''
}
async function save(publish = false) {
  saving.value = true
  error.value = ''
  try {
    if (!form.name.trim() || !form.draft.title.trim())
      throw new Error(t('pageEditor.pageNameAndTitleAreRequired'))
    if (!/^[a-z0-9][a-z0-9-]*$/.test(form.slug))
      throw new Error(t('pageEditor.slugMayContainLowercaseLettersNumbersAndHyphens'))
    for (const link of form.draft.links) {
      const parsed = new URL(link.url)
      if (!['https:', 'http:'].includes(parsed.protocol) || !link.label.trim())
        throw new Error(t('pageEditor.publicLinksNeedLabelsAndHttpSUrls'))
    }
    if (
      form.draft.logoUrl
      && !/^https:\/\//.test(form.draft.logoUrl)
      && !form.draft.logoUrl.startsWith('/assets/')
    ) {
      throw new Error(t('pageEditor.logoMustUseHttpsOrAnUploadedAsset'))
    }
    const body = clone(form)
    const data = (await (editing.value
      ? updatePage.mutateAsync({ path: { id: form.id }, body })
      : createPage.mutateAsync({ body }))) as Page
    replacePage(data)
    if (publish) {
      await publishPage.mutateAsync({ path: { id: form.id } })
      const page = await queryCache.refresh(
        queryCache.ensure({ ...getPagesQuery({ path: { id: form.id } }), staleTime: 0 }),
      )
      if (page.status !== 'success')
        throw page.error || new Error(t('errors.requestFailed'))
      replacePage(page.data as Page)
      notify(t('pageEditor.statusPagePublished'))
    }
    else {
      notify(t('pageEditor.draftSaved'))
    }
    const preview = await queryCache.refresh(
      queryCache.ensure({ ...previewPageQuery({ path: { id: form.id } }), staleTime: 0 }),
    )
    if (preview.status !== 'success')
      throw preview.error || new Error(t('errors.requestFailed'))
    savedPreview.value = clone(preview.data) as PublicPage
    if (!id.value)
      await router.replace(`/app/pages/${form.id}`)
  }
  catch (e) {
    error.value = errorText(e)
  }
  finally {
    saving.value = false
  }
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
    form.draft.logoUrl = (
      await uploadAsset.mutateAsync({
        body: { filename: file.name, contentType: file.type, base64 },
      })
    ).url
    notify(t('pageEditor.logoUploaded'))
  }
  catch (e) {
    notify(errorText(e), 'error')
  }
}
async function remove() {
  try {
    await deletePage.mutateAsync({ path: { id: form.id } })
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
    <RouterLink to="/app/pages" class="button ghost">
      <span class="i-lucide-arrow-left" un-w="14px" un-h="14px" aria-hidden="true" />{{ t('pageEditor.allPages') }}
    </RouterLink><template v-if="canEdit()">
      <button class="button" :disabled="saving" @click="save()">
        <span class="i-lucide-save" un-w="14px" un-h="14px" aria-hidden="true" />{{ t('common.saveDraft') }}
      </button><button class="button primary" :disabled="saving" @click="save(true)">
        <span class="i-lucide-send" un-w="14px" un-h="14px" aria-hidden="true" />{{ t('pageEditor.publishPage') }}
      </button>
    </template>
  </PageHeader><AsyncState :pending="loading">
    <div v-if="error" class="validation-error" role="alert">
      {{ error }}
    </div>
    <div class="editor-layout">
      <div>
        <section class="card">
          <div class="form-section">
            <h2>{{ t('pageEditor.pageAccess') }}</h2>
            <p class="muted">
              {{ t('pageEditor.pathAndCustomDomainServeTheSamePublished') }}
            </p>
            <div class="form-grid">
              <Field :label="t('pageEditor.internalPageName')">
                <input v-model="form.name" :disabled="!canEdit()" required>
              </Field><Field
                :label="t('pageEditor.pageSlug')"
                :hint="`${origin}/${form.slug || 'status1'}`"
              >
                <input
                  v-model="form.slug"
                  :disabled="!canEdit()"
                  placeholder="status1"
                  pattern="[a-z0-9][a-z0-9-]*"
                  required
                >
              </Field><Field
                class="span-full"
                :label="t('pageEditor.customDomain')"
                :hint="t('pageEditor.selectAnAdministratorConfiguredDomainDnsAndHttps')"
              >
                <select v-model="form.domain" :disabled="!canEdit()">
                  <option value="">
                    {{ t('pageEditor.pathAccessOnly') }}
                  </option>
                  <option v-for="domain in allowedDomains" :key="domain">
                    {{ domain }}
                  </option>
                </select>
              </Field>
            </div>
            <p v-if="form.publishedAt" class="note" un-mt="5">
              {{ t('pageEditor.lastPublished') }} {{ formatDate(form.publishedAt) }} · v{{
                form.version
              }}<a
                :href="publishedEntry(form).url"
                class="button small ghost"
                target="_blank"
                rel="noopener"
              ><span class="i-lucide-external-link" un-w="12px" un-h="12px" aria-hidden="true" />{{ t('pageEditor.visitPublicPage') }}</a>
            </p>
          </div>
          <div class="form-section">
            <h2>{{ t('pageEditor.brandAppearance') }}</h2>
            <p class="muted">
              {{ t('pageEditor.theseSettingsApplyOnlyToThisStatusPage') }}
            </p>
            <div class="form-grid">
              <Field class="span-full" :label="t('pageEditor.publicTitle')">
                <input v-model="form.draft.title" :disabled="!canEdit()" required>
              </Field><Field class="span-full" :label="t('pageEditor.pageDescription')">
                <textarea v-model="form.draft.description" :disabled="!canEdit()" rows="3" />
              </Field><Field class="span-full" :label="t('pageEditor.logoUrl')">
                <div class="field-row">
                  <input
                    v-model="form.draft.logoUrl"
                    :disabled="!canEdit()"
                    :placeholder="t('pageEditor.logoPlaceholder')"
                  ><label v-if="canEdit()" class="button"><span class="i-lucide-upload" un-w="14px" un-h="14px" aria-hidden="true" />{{ t('pageEditor.upload')
                  }}<input
                    type="file"
                    accept="image/png,image/jpeg,image/gif"
                    un-hidden=""
                    @change="uploadLogo"
                  ></label>
                </div>
              </Field><Field :label="t('pageEditor.brandColor')">
                <div class="field-row">
                  <input
                    v-model="form.draft.brandColor"
                    type="color"
                    :disabled="!canEdit()"
                  ><input
                    v-model="form.draft.brandColor"
                    :disabled="!canEdit()"
                    pattern="#[0-9a-fA-F]{6}"
                  >
                </div>
              </Field><Field :label="t('pageEditor.colorScheme')">
                <select v-model="form.draft.colorScheme" :disabled="!canEdit()">
                  <option value="system">
                    {{ t('common.system') }}
                  </option>
                  <option value="light">
                    {{ t('common.light') }}
                  </option>
                  <option value="dark">
                    {{ t('common.dark') }}
                  </option>
                </select>
              </Field>
              <div class="span-full">
                <label class="field-label">{{ t('pageEditor.publicLinks') }}</label>
                <div
                  v-for="(link, index) in form.draft.links"
                  :key="index"
                  class="link-editor"
                  un-mt="3"
                >
                  <input
                    v-model="link.label"
                    :disabled="!canEdit()"
                    :placeholder="t('pageEditor.linkLabel')"
                  ><input
                    v-model="link.url"
                    :disabled="!canEdit()"
                    type="url"
                    placeholder="https://…"
                  ><button
                    v-if="canEdit()"
                    class="icon-button"
                    :aria-label="t('pageEditor.removeLink')"
                    @click="form.draft.links.splice(index, 1)"
                  >
                    <span class="i-lucide-x" un-w="14px" un-h="14px" aria-hidden="true" />
                  </button>
                </div>
                <button
                  v-if="canEdit()"
                  class="button small ghost"
                  un-mt="2"
                  @click="form.draft.links.push({ label: '', url: '' })"
                >
                  <span class="i-lucide-plus" un-w="13px" un-h="13px" aria-hidden="true" />{{ t('pageEditor.addLink') }}
                </button>
              </div>
            </div>
          </div>
          <div class="form-section">
            <h2>{{ t('pageEditor.servicesGroups') }}</h2>
            <p class="muted">
              {{ t('pageEditor.publishOnlySelectedMonitorsPublicAliasesLeaveInternal') }}
            </p>
            <div v-for="(group, index) in form.draft.groups" :key="group.id" class="group-editor">
              <div class="group-editor-header">
                <input
                  v-model="group.name"
                  :aria-label="t('common.groupName')"
                  :disabled="!canEdit()"
                  :placeholder="t('common.groupName')"
                ><template v-if="canEdit()">
                  <button
                    class="icon-button"
                    :disabled="index === 0"
                    :aria-label="t('pageEditor.moveGroupUp')"
                    @click="move(form.draft.groups, index, -1)"
                  >
                    <span class="i-lucide-arrow-up" un-w="13px" un-h="13px" aria-hidden="true" />
                  </button><button
                    class="icon-button"
                    :disabled="index === form.draft.groups.length - 1"
                    :aria-label="t('pageEditor.moveGroupDown')"
                    @click="move(form.draft.groups, index, 1)"
                  >
                    <span class="i-lucide-arrow-down" un-w="13px" un-h="13px" aria-hidden="true" />
                  </button><button
                    class="icon-button"
                    :aria-label="t('pageEditor.removeGroup')"
                    @click="form.draft.groups.splice(index, 1)"
                  >
                    <span class="i-lucide-x" un-w="14px" un-h="14px" aria-hidden="true" />
                  </button>
                </template>
              </div>
              <div
                v-for="(item, mIndex) in group.monitors"
                :key="item.monitorId"
                class="group-editor-row"
              >
                <div un-flex="1" un-min-w="0">
                  <span class="mini-label">{{
                    monitors.find((x) => x.id === item.monitorId)?.name
                  }}</span><input
                    v-model="item.alias"
                    :aria-label="t('common.publicAlias')"
                    :disabled="!canEdit()"
                    :placeholder="t('common.publicAlias')"
                    un-mt="1"
                  >
                  <div un-flex="~ gap-4" un-mt="2">
                    <label class="checkbox-label"><input v-model="item.showUptime" type="checkbox" :disabled="!canEdit()">{{
                      t('common.uptime')
                    }}</label><label class="checkbox-label"><input v-model="item.showLatency" type="checkbox" :disabled="!canEdit()">{{
                      t('pageEditor.latency')
                    }}</label>
                  </div>
                </div>
                <template v-if="canEdit()">
                  <button
                    class="icon-button"
                    :disabled="mIndex === 0"
                    :aria-label="t('pageEditor.moveServiceUp')"
                    @click="move(group.monitors, mIndex, -1)"
                  >
                    <span class="i-lucide-arrow-up" un-w="13px" un-h="13px" aria-hidden="true" />
                  </button><button
                    class="icon-button"
                    :disabled="mIndex === group.monitors.length - 1"
                    :aria-label="t('pageEditor.moveServiceDown')"
                    @click="move(group.monitors, mIndex, 1)"
                  >
                    <span class="i-lucide-arrow-down" un-w="13px" un-h="13px" aria-hidden="true" />
                  </button><button
                    class="icon-button"
                    :aria-label="t('pageEditor.removeService')"
                    @click="group.monitors.splice(mIndex, 1)"
                  >
                    <span class="i-lucide-x" un-w="14px" un-h="14px" aria-hidden="true" />
                  </button>
                </template>
              </div>
              <div v-if="canEdit()" class="group-editor-row">
                <select v-model="newMonitorIds[group.id]" :aria-label="t('common.selectAMonitor')">
                  <option value="">
                    {{ t('common.selectAMonitor') }}
                  </option>
                  <option v-for="monitor in monitors" :key="monitor.id" :value="monitor.id">
                    {{ monitor.name }} · {{ monitor.type }}
                  </option>
                </select><button class="button small" @click="addMonitor(group.id)">
                  <span class="i-lucide-plus" un-w="13px" un-h="13px" aria-hidden="true" />{{ t('pageEditor.add') }}
                </button>
              </div>
            </div>
            <button
              v-if="canEdit()"
              class="button small"
              un-mt="4"
              @click="
                form.draft.groups.push({
                  id: newID(),
                  name: t('pageEditor.services'),
                  monitors: [],
                })
              "
            >
              <span class="i-lucide-plus" un-w="13px" un-h="13px" aria-hidden="true" />{{ t('pageEditor.addGroup') }}
            </button>
          </div>
        </section>
        <div class="form-actions">
          <button
            v-if="editing && canEdit()"
            class="button danger"
            un-mr="auto"
            @click="deleteOpen = true"
          >
            <span class="i-lucide-trash-2" un-w="13px" un-h="13px" aria-hidden="true" />{{ t('pageEditor.deletePage') }}
          </button><button v-if="canEdit()" class="button primary" :disabled="saving" @click="save()">
            <span class="i-lucide-save" un-w="14px" un-h="14px" aria-hidden="true" />{{ t('common.saveDraft') }}
          </button>
        </div>
      </div>
      <aside class="editor-preview">
        <div class="preview-label">
          <strong>{{ t('pageEditor.liveDraftPreview') }}</strong><span>{{ t('pageEditor.responsivePreview') }}</span>
        </div>
        <StatusPage :page="preview" preview />
        <p class="field-hint" un-mt="3">
          {{ t('pageEditor.savingRefreshesRealStatisticsNewlySelectedServicesShow') }}
        </p>
      </aside>
    </div>
  </AsyncState><Modal
    v-model:open="deleteOpen"
    :title="t('pageEditor.deleteStatusPage')"
    :description="t('pageEditor.thePagePathAndDomainWillStopPublishing')"
  >
    <template #footer>
      <button class="button" @click="deleteOpen = false">
        {{ t('common.cancel') }}
      </button><button class="button danger" @click="remove">
        {{ t('common.delete') }}
      </button>
    </template>
  </Modal>
</template>
