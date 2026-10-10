<script setup lang="ts">
import type { Secret } from '../../../client/types.gen'
import { useMutation, useQuery } from '@pinia/colada'
import { useForm } from '@tanstack/vue-form'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  createSecretMutation,
  deleteSecretMutation,
  listSecretsQuery,
  updateSecretMutation,
} from '../../../client/@pinia/colada.gen'
import { Banner } from '../../../components/ui/banner'
import { PageHeader } from '../../../components/ui/blocks/page-header'
import { Button } from '../../../components/ui/button'
import { ClipboardText } from '../../../components/ui/clipboard-text'
import { Dialog } from '../../../components/ui/dialog'
import { Empty } from '../../../components/ui/empty'
import { Field } from '../../../components/ui/field'
import { Input, InputArea } from '../../../components/ui/input'
import { InputGroup, InputGroupAddon, InputGroupInput } from '../../../components/ui/input-group'
import { LayerCard, LayerCardPrimary } from '../../../components/ui/layer-card'
import { Loader } from '../../../components/ui/loader'
import { Table, TableBody, TableCell, TableContainer, TableHead, TableHeader, TablePagination, TableRow, TableToolbar } from '../../../components/ui/table'
import { isAdmin } from '../../../composables/api'
import { notify } from '../../../composables/notices'
import { formatDate } from '../../../composables/preferences'
import { errorText } from '../../../lib/errors'

const { t } = useI18n({ useScope: 'global' })

definePage({ meta: { title: 'navigation.secrets', roles: ['admin'], contentWidth: 'compact' } })

const createSecret = useMutation(createSecretMutation())
const updateSecret = useMutation(updateSecretMutation())
const deleteSecret = useMutation(deleteSecretMutation())

const query = useQuery({ ...listSecretsQuery(), staleTime: 10000 })
const search = ref('')
const items = computed(() => {
  const term = search.value.trim().toLowerCase()
  return (query.data.value?.items || []).filter(secret => `${secret.name} ${secret.id}`.toLowerCase().includes(term))
})
const open = ref(false)
const error = ref('')
const formApi = useForm({
  defaultValues: { id: '', name: '', value: '' },
  onSubmit: async ({ value }) => {
    error.value = ''
    try {
      const body = { name: value.name, value: value.value }
      if (value.id)
        await updateSecret.mutateAsync({ path: { id: value.id }, body })
      else await createSecret.mutateAsync({ body })
      formApi.setFieldValue('value', '')
      open.value = false
      await query.refresh()
      notify(t('secrets.secret-saved'))
    }
    catch (e) {
      error.value = errorText(e)
    }
  },
})
const form = formApi.useSelector(state => state.values)
const saving = formApi.useSelector(state => state.isSubmitting)
const deleteTarget = ref<Secret | null>(null)
const deleteOpen = ref(false)
const deleting = ref(false)
function edit(secret?: Secret) {
  if (formApi.state.isSubmitting)
    return
  formApi.reset({ id: secret?.id || '', name: secret?.name || '', value: '' })
  error.value = ''
  open.value = true
}
async function remove() {
  if (!deleteTarget.value || deleting.value)
    return
  deleting.value = true
  try {
    await deleteSecret.mutateAsync({ path: { id: deleteTarget.value.id } })
    deleteOpen.value = false
    await query.refresh()
    notify(t('secrets.secret-deleted'))
  }
  catch (e) {
    notify(errorText(e), 'error')
  }
  finally {
    deleting.value = false
  }
}
function confirmDelete(value: Secret) {
  deleteTarget.value = value
  deleteOpen.value = true
}
function cancel() {
  open.value = false
  formApi.setFieldValue('value', '')
}
</script>

<template>
  <PageHeader class="mb-6" :title="t('navigation.secrets')" :description="t('secrets.manage-sensitive-values-used-by-requests-tls-proxies')">
    <template #actions>
      <Button v-if="isAdmin()" variant="primary" @click="edit()">
        <span w="15px" h="15px" aria-hidden="true" class="i-lucide-plus" />{{ t('common.add-secret') }}
      </Button>
    </template>
  </PageHeader>
  <p class="mb-5 flex items-start gap-2 text-size-sm text-subtle">
    <span class="i-lucide-key-round mt-0.5 size-4 shrink-0" aria-hidden="true" />
    {{ t('secrets.saved-values-cannot-be-read-back-replace-a') }}
  </p>
  <LayerCard>
    <LayerCardPrimary class="p-0!">
      <TableToolbar>
        <InputGroup class="w-full max-w-sm">
          <InputGroupAddon><span class="i-lucide-search size-4" aria-hidden="true" /></InputGroupAddon>
          <InputGroupInput v-model="search" type="search" :placeholder="t('secrets.search-secrets-placeholder')" :aria-label="t('secrets.search-secrets')" />
        </InputGroup>
        <Button v-if="search" variant="ghost" size="sm" @click="search = ''">
          {{ t('common.clear-filters') }}
        </Button>
      </TableToolbar>
      <div v-if="query.isPending.value" class="loading-state" flex="~ justify-center items-center gap-10px" p="60px" un-text="12px subtle" role="status">
        <Loader :label="t('async-state.loading-data')" />{{ t('async-state.loading-data') }}
      </div>
      <Banner v-else-if="query.error.value" variant="error">
        {{ errorText(query.error.value) }}
        <Button variant="ghost" @click="query.refetch()">
          {{ t('async-state.retry') }}
        </Button>
      </Banner>
      <template v-else>
        <Empty v-if="!items.length" size="sm" class="rounded-none border-none" :title="search ? t('secrets.no-matching-secrets') : t('secrets.no-secrets-yet')" :description="search ? t('monitors.try-changing-your-search-or-filters') : t('secrets.store-tokens-pem-certificates-proxy-passwords-and-shoutrrr')">
          <Button v-if="search" @click="search = ''">
            {{ t('common.clear-filters') }}
          </Button>
          <Button v-else-if="isAdmin()" variant="primary" @click="edit()">
            {{ t('common.add-secret') }}
          </Button>
        </Empty>
        <TableContainer v-else :scroll-label="t('common.scroll-table')">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{{ t('common.name') }}</TableHead>
                <TableHead class="[@container_workspace_(max-width:_700px)]:hidden">
                  {{ t('secrets.updated') }}
                </TableHead>
                <TableHead class="w-28">
                  <span class="sr-only">{{ t('common.edit') }}</span>
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="secret in items" :key="secret.id">
                <TableCell>
                  <span class="block font-medium [overflow-wrap:anywhere]">{{ secret.name }}</span>
                  <ClipboardText :text="secret.id" inline :copy-label="t('monitor-details.copy')" :copied-label="t('monitor-details.copied')" class="mt-1 max-w-64 text-subtle" />
                  <span class="mt-1 block text-size-xs text-subtle [@container_workspace_(width_>_700px)]:hidden">{{ t('secrets.updated') }} · {{ formatDate(secret.updatedAt) }}</span>
                </TableCell>
                <TableCell class="text-subtle [@container_workspace_(max-width:_700px)]:hidden">
                  {{ formatDate(secret.updatedAt) }}
                </TableCell>
                <TableCell>
                  <div v-if="isAdmin()" class="flex flex-wrap justify-end gap-2">
                    <Button size="sm" @click="edit(secret)">
                      <span w="12px" h="12px" aria-hidden="true" class="i-lucide-pencil" />{{ t('secrets.replace') }}
                    </Button><Button :aria-label="t('common.delete')" shape="square" @click="confirmDelete(secret)">
                      <span w="14px" h="14px" aria-hidden="true" class="i-lucide-trash-2" />
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </TableContainer>
      </template>
      <TablePagination v-if="query.data.value">
        {{ t('secrets.showing-secrets', { shown: items.length, total: query.data.value.items.length }) }}
      </TablePagination>
    </LayerCardPrimary>
  </LayerCard>
  <Dialog v-model:open="open" :close-label="t('common.close')" size="lg" :title="form.id ? t('secrets.update-secret') : t('common.add-secret')">
    <form id="secret-form" @submit.prevent="formApi.handleSubmit()">
      <formApi.Field v-slot="{ field }" name="name">
        <Field :label="t('common.name')">
          <Input :name="field.name" :model-value="field.state.value" required @update:model-value="field.handleChange(String($event ?? ''))" @blur="field.handleBlur" />
        </Field>
      </formApi.Field>
      <formApi.Field v-slot="{ field }" name="value">
        <Field :label="form.id ? t('secrets.replacement-secret-value') : t('secrets.secret-value')" mt="5">
          <InputArea :name="field.name" :model-value="field.state.value" required :min-rows="6" autocomplete="off" spellcheck="false" @update:model-value="field.handleChange($event ?? '')" @blur="field.handleBlur" />
        </Field>
      </formApi.Field>
      <Banner v-if="error" variant="error" class="mt-4">
        {{ error }}
      </Banner>
    </form>
    <template #footer>
      <Button :disabled="saving" @click="cancel">
        {{ t('common.cancel') }}
      </Button><Button type="submit" form="secret-form" :loading="saving" variant="primary">
        {{ t('secrets.save-secret') }}
      </Button>
    </template>
  </Dialog><Dialog v-model:open="deleteOpen" :close-label="t('common.close')" :title="t('secrets.delete-secret')" :description="t('secrets.remove-monitor-and-channel-references-before-deleting-a')">
    <p>{{ deleteTarget?.name }}</p>
    <template #footer>
      <Button :disabled="deleting" @click="deleteOpen = false">
        {{ t('common.cancel') }}
      </Button><Button variant="destructive" :loading="deleting" @click="remove">
        {{ t('common.delete') }}
      </Button>
    </template>
  </Dialog>
</template>
