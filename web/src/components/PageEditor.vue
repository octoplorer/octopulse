<script setup lang="ts">
import type { Monitor, Page, PublicPage } from '../client/types.gen'
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
import { Alert } from './ui/alert'
import { Button } from './ui/button'
import { Card } from './ui/card'
import { FieldActions, FieldDescription, FieldGroup, FieldLabel, FieldSection } from './ui/field'
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
    brandColor: '#2563eb',
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
      : createPage.mutateAsync({ body })))
    replacePage(data)
    if (publish) {
      await publishPage.mutateAsync({ path: { id: form.id } })
      const page = await queryCache.refresh(
        queryCache.ensure({ ...getPagesQuery({ path: { id: form.id } }), staleTime: 0 }),
      )
      if (page.status !== 'success')
        throw page.error || new Error(t('errors.requestFailed'))
      replacePage(page.data)
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
    savedPreview.value = clone(preview.data)
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
    <Alert v-if="error" role="alert" variant="validation">
      {{ error }}
    </Alert>
    <div grid="~ cols-[minmax(0,1fr)_minmax(320px,0.8fr)]" gap="24px" class="[@media(max-width:1200px)]:grid-cols-1 [@container_workspace_(max-width:_700px)]:grid-cols-1">
      <div>
        <Card as="section">
          <FieldSection>
            <h2>{{ t('pageEditor.pageAccess') }}</h2>
            <p un-text="13px $muted">
              {{ t('pageEditor.pathAndCustomDomainServeTheSamePublished') }}
            </p>
            <FieldGroup>
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
                <Select v-model="form.domain" :disabled="!canEdit()">
                  <SelectTrigger><SelectValue :placeholder="t('pageEditor.pathAccessOnly')" /></SelectTrigger>
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
            <p un-text="13px $muted">
              {{ t('pageEditor.theseSettingsApplyOnlyToThisStatusPage') }}
            </p>
            <FieldGroup>
              <Field class="span-full" :label="t('pageEditor.publicTitle')">
                <input v-model="form.draft.title" :disabled="!canEdit()" required>
              </Field><Field class="span-full" :label="t('pageEditor.pageDescription')">
                <textarea v-model="form.draft.description" :disabled="!canEdit()" rows="3" />
              </Field><Field class="span-full" :label="t('pageEditor.logoUrl')">
                <div flex="~ items-center gap-10px" class="[@media(max-width:700px)]:flex-wrap">
                  <input
                    v-model="form.draft.logoUrl"
                    :disabled="!canEdit()"
                    :placeholder="t('pageEditor.logoPlaceholder')"
                  ><Button v-if="canEdit()" as-child>
                    <label><span class="i-lucide-upload" w="14px" h="14px" aria-hidden="true" />{{ t('pageEditor.upload')
                    }}<input
                      type="file"
                      accept="image/png,image/jpeg,image/gif"
                      hidden=""
                      @change="uploadLogo"
                    ></label>
                  </Button>
                </div>
              </Field><Field :label="t('pageEditor.brandColor')">
                <div flex="~ items-center gap-10px" class="[@media(max-width:700px)]:flex-wrap">
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
                <Select v-model="form.draft.colorScheme" :disabled="!canEdit()">
                  <SelectTrigger><SelectValue /></SelectTrigger>
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
              </Field>
              <div class="span-full">
                <FieldLabel as="label">
                  {{ t('pageEditor.publicLinks') }}
                </FieldLabel>
                <div
                  v-for="(link, index) in form.draft.links"
                  :key="index"
                  grid="~ cols-[1fr_2fr_auto]" gap="9px" mb="9px" class="[@media(max-width:700px)]:grid-cols-1"
                  mt="3"
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
                  ><Button
                    v-if="canEdit()"
                    :aria-label="t('pageEditor.removeLink')"
                    size="icon"
                    @click="form.draft.links.splice(index, 1)"
                  >
                    <span class="i-lucide-x" w="14px" h="14px" aria-hidden="true" />
                  </Button>
                </div>
                <Button
                  v-if="canEdit()"
                  mt="2"
                  variant="ghost"
                  size="sm"
                  @click="form.draft.links.push({ label: '', url: '' })"
                >
                  <span class="i-lucide-plus" w="13px" h="13px" aria-hidden="true" />{{ t('pageEditor.addLink') }}
                </Button>
              </div>
            </FieldGroup>
          </FieldSection>
          <FieldSection>
            <h2>{{ t('pageEditor.servicesGroups') }}</h2>
            <p un-text="13px $muted">
              {{ t('pageEditor.publishOnlySelectedMonitorsPublicAliasesLeaveInternal') }}
            </p>
            <div v-for="(group, index) in form.draft.groups" :key="group.id" class="group-editor" border="1px solid $border" rounded="9px" mt="15px" overflow="hidden">
              <div class="group-editor-header" flex="~ items-center gap-8px" p="13px" bg="$surface-soft" border="b-1px b-solid b-$border">
                <input
                  v-model="group.name"
                  :aria-label="t('common.groupName')"
                  :disabled="!canEdit()"
                  :placeholder="t('common.groupName')"
                ><template v-if="canEdit()">
                  <Button
                    :disabled="index === 0"
                    :aria-label="t('pageEditor.moveGroupUp')"
                    size="icon"
                    @click="move(form.draft.groups, index, -1)"
                  >
                    <span class="i-lucide-arrow-up" w="13px" h="13px" aria-hidden="true" />
                  </Button><Button
                    :disabled="index === form.draft.groups.length - 1"
                    :aria-label="t('pageEditor.moveGroupDown')"
                    size="icon"
                    @click="move(form.draft.groups, index, 1)"
                  >
                    <span class="i-lucide-arrow-down" w="13px" h="13px" aria-hidden="true" />
                  </Button><Button
                    :aria-label="t('pageEditor.removeGroup')"
                    size="icon"
                    @click="form.draft.groups.splice(index, 1)"
                  >
                    <span class="i-lucide-x" w="14px" h="14px" aria-hidden="true" />
                  </Button>
                </template>
              </div>
              <div
                v-for="(item, mIndex) in group.monitors"
                :key="item.monitorId"
                flex="~ items-center gap-9px" px="13px" py="11px" border="b-1px b-solid b-$border last:0" class="[@media(max-width:700px)]:flex-wrap"
              >
                <div flex="1" min-w="0">
                  <span un-text="12px $muted" tracking="0.5px">{{
                    monitors.find((x) => x.id === item.monitorId)?.name
                  }}</span><input
                    v-model="item.alias"
                    flex="1"
                    min-w="0"
                    :aria-label="t('common.publicAlias')"
                    :disabled="!canEdit()"
                    :placeholder="t('common.publicAlias')"
                    mt="1"
                  >
                  <div flex="~ gap-4" mt="2">
                    <label flex="~ items-center gap-8px" un-text="12px $text"><input v-model="item.showUptime" flex="1" min-w="0" type="checkbox" :disabled="!canEdit()">{{
                      t('common.uptime')
                    }}</label><label flex="~ items-center gap-8px" un-text="12px $text"><input v-model="item.showLatency" flex="1" min-w="0" type="checkbox" :disabled="!canEdit()">{{
                      t('pageEditor.latency')
                    }}</label>
                  </div>
                </div>
                <template v-if="canEdit()">
                  <Button
                    :disabled="mIndex === 0"
                    :aria-label="t('pageEditor.moveServiceUp')"
                    size="icon"
                    @click="move(group.monitors, mIndex, -1)"
                  >
                    <span class="i-lucide-arrow-up" w="13px" h="13px" aria-hidden="true" />
                  </Button><Button
                    :disabled="mIndex === group.monitors.length - 1"
                    :aria-label="t('pageEditor.moveServiceDown')"
                    size="icon"
                    @click="move(group.monitors, mIndex, 1)"
                  >
                    <span class="i-lucide-arrow-down" w="13px" h="13px" aria-hidden="true" />
                  </Button><Button
                    :aria-label="t('pageEditor.removeService')"
                    size="icon"
                    @click="group.monitors.splice(mIndex, 1)"
                  >
                    <span class="i-lucide-x" w="14px" h="14px" aria-hidden="true" />
                  </Button>
                </template>
              </div>
              <div v-if="canEdit()" flex="~ items-center gap-9px" px="13px" py="11px" border="b-1px b-solid b-$border last:0" class="[@media(max-width:700px)]:flex-wrap">
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
                form.draft.groups.push({
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
            variant="danger"
            @click="deleteOpen = true"
          >
            <span class="i-lucide-trash-2" w="13px" h="13px" aria-hidden="true" />{{ t('pageEditor.deletePage') }}
          </Button><Button v-if="canEdit()" :disabled="saving" variant="primary" @click="save()">
            <span class="i-lucide-save" w="14px" h="14px" aria-hidden="true" />{{ t('common.saveDraft') }}
          </Button>
        </FieldActions>
      </div>
      <aside pos="sticky" top="20px" self="start" class="[@media(max-width:1200px)]:static [@media(max-width:700px)]:min-w-0">
        <div class="preview-label" flex="~ items-center justify-between" un-text="12px $muted" mb="12px">
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
      </Button><Button variant="danger" @click="remove">
        {{ t('common.delete') }}
      </Button>
    </template>
  </Modal>
</template>
