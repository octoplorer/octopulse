<script setup lang="ts">
import type { Secret } from '../../../client/types.gen'
import { useMutation, useQuery } from '@pinia/colada'
import { reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  createSecretMutation,
  deleteSecretMutation,
  listSecretsQuery,
  updateSecretMutation,
} from '../../../client/@pinia/colada.gen'
import AsyncState from '../../../components/AsyncState.vue'
import EmptyState from '../../../components/EmptyState.vue'
import Field from '../../../components/Field.vue'
import Modal from '../../../components/Modal.vue'
import PageHeader from '../../../components/PageHeader.vue'
import { Button } from '../../../components/ui/button'
import { Card } from '../../../components/ui/card'
import { FieldError } from '../../../components/ui/field'
import { Table, TableBody, TableCell, TableContainer, TableHead, TableHeader, TableRow } from '../../../components/ui/table'
import { isAdmin } from '../../../composables/api'
import { notify } from '../../../composables/notices'
import { formatDate } from '../../../composables/preferences'
import { errorText } from '../../../lib/errors'

const { t } = useI18n({ useScope: 'global' })

definePage({ meta: { title: 'navigation.secrets', roles: ['admin'] } })

const createSecret = useMutation(createSecretMutation())
const updateSecret = useMutation(updateSecretMutation())
const deleteSecret = useMutation(deleteSecretMutation())

const query = useQuery({ ...listSecretsQuery(), staleTime: 10000 })
const open = ref(false)
const saving = ref(false)
const error = ref('')
const form = reactive({ id: '', name: '', value: '' })
const deleteTarget = ref<Secret | null>(null)
const deleteOpen = ref(false)
function edit(secret?: Secret) {
  Object.assign(form, { id: secret?.id || '', name: secret?.name || '', value: '' })
  error.value = ''
  open.value = true
}
async function save() {
  saving.value = true
  error.value = ''
  try {
    const body = { name: form.name, value: form.value }
    if (form.id)
      await updateSecret.mutateAsync({ path: { id: form.id }, body })
    else await createSecret.mutateAsync({ body })
    form.value = ''
    open.value = false
    await query.refresh()
    notify(t('secrets.secretSaved'))
  }
  catch (e) {
    error.value = errorText(e)
  }
  finally {
    saving.value = false
  }
}
async function remove() {
  if (!deleteTarget.value)
    return
  try {
    await deleteSecret.mutateAsync({ path: { id: deleteTarget.value.id } })
    deleteOpen.value = false
    await query.refresh()
    notify(t('secrets.secretDeleted'))
  }
  catch (e) {
    notify(errorText(e), 'error')
  }
}
function confirmDelete(value: Secret) {
  deleteTarget.value = value
  deleteOpen.value = true
}
function cancel() {
  open.value = false
  form.value = ''
}
</script>

<template>
  <PageHeader :title="t('navigation.secrets')" :description="t('secrets.manageSensitiveValuesUsedByRequestsTlsProxies')">
    <Button v-if="isAdmin()" variant="primary" @click="edit()">
      <span w="15px" h="15px" aria-hidden="true" class="i-lucide-plus" />{{ t('common.addSecret') }}
    </Button>
  </PageHeader>
  <div flex="~ items-center" gap="9px" mb="22px" p="y-13px x-16px" border="1 solid line" rounded="8px" bg="$surface" un-text="12px $muted">
    <span w="16px" h="16px" aria-hidden="true" class="i-lucide-key-round" />{{ t('secrets.savedValuesCannotBeReadBackReplaceA') }}
  </div>
  <Card as="section">
    <AsyncState :pending="query.isPending.value" :error="query.error.value" @retry="query.refetch()">
      <EmptyState v-if="!query.data.value?.items.length" :title="t('secrets.noSecretsYet')" :description="t('secrets.storeTokensPemCertificatesProxyPasswordsAndShoutrrr')" />
      <TableContainer v-else>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{{ t('common.name') }}</TableHead>
              <TableHead>{{ t('secrets.referenceId') }}</TableHead>
              <TableHead>{{ t('secrets.updated') }}</TableHead>
              <TableHead />
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="secret in query.data.value.items" :key="secret.id">
              <TableCell class="monitor-name block" font="600" un-text="13px">
                {{ secret.name }}
              </TableCell>
              <TableCell>
                <code class="muted" un-text="13px $muted">{{ secret.id }}</code>
              </TableCell>
              <TableCell class="muted" un-text="13px $muted">
                {{ formatDate(secret.updatedAt) }}
              </TableCell>
              <TableCell>
                <div v-if="isAdmin()" flex="~ gap-2">
                  <Button size="sm" @click="edit(secret)">
                    <span w="12px" h="12px" aria-hidden="true" class="i-lucide-pencil" />{{ t('secrets.replace') }}
                  </Button><Button :aria-label="t('common.delete')" size="icon" @click="confirmDelete(secret)">
                    <span w="14px" h="14px" aria-hidden="true" class="i-lucide-trash-2" />
                  </Button>
                </div>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </TableContainer>
    </AsyncState>
  </Card>
  <Modal v-model:open="open" :title="form.id ? t('secrets.updateSecret') : t('common.addSecret')">
    <form id="secret-form" @submit.prevent="save">
      <Field :label="t('common.name')">
        <input v-model="form.name" required>
      </Field><Field :label="form.id ? t('secrets.replacementSecretValue') : t('secrets.secretValue')" mt="5">
        <textarea v-model="form.value" required rows="6" autocomplete="off" spellcheck="false" />
      </Field>
      <FieldError v-if="error" as="p" py="10px" px="0">
        {{ error }}
      </FieldError>
    </form>
    <template #footer>
      <Button @click="cancel">
        {{ t('common.cancel') }}
      </Button><Button form="secret-form" :disabled="saving" variant="primary">
        {{ t('secrets.saveSecret') }}
      </Button>
    </template>
  </Modal><Modal v-model:open="deleteOpen" :title="t('secrets.deleteSecret')" :description="t('secrets.removeMonitorAndChannelReferencesBeforeDeletingA')">
    <p>{{ deleteTarget?.name }}</p>
    <template #footer>
      <Button @click="deleteOpen = false">
        {{ t('common.cancel') }}
      </Button><Button variant="danger" @click="remove">
        {{ t('common.delete') }}
      </Button>
    </template>
  </Modal>
</template>
