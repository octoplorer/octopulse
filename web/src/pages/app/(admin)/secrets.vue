<script setup lang="ts">
import type { DefineQueryOptions } from '@pinia/colada'
import type { ErrorModel } from '../../../client/types.gen'
import type { Secret } from '../../../lib/types'
import { KeyRound, Pencil, Plus, Trash2 } from '@lucide/vue'
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
import { isAdmin } from '../../../composables/api'
import { notify } from '../../../composables/notices'
import { formatDate } from '../../../composables/preferences'
import { errorText } from '../../../lib/errors'

const { t } = useI18n({ useScope: 'global' })

definePage({ meta: { title: 'navigation.secrets', roles: ['admin'] } })

const createSecret = useMutation(createSecretMutation())
const updateSecret = useMutation(updateSecretMutation())
const deleteSecret = useMutation(deleteSecretMutation())

const query = useQuery({ ...listSecretsQuery(), staleTime: 10000 } as DefineQueryOptions<
  { items: Secret[] },
  ErrorModel
>)
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
  <PageHeader
    :title="t('navigation.secrets')"
    :description="t('secrets.manageSensitiveValuesUsedByRequestsTlsProxies')"
  >
    <button v-if="isAdmin()" class="button primary" @click="edit()">
      <Plus :size="15" />{{ t('common.addSecret') }}
    </button>
  </PageHeader>
  <div class="alert-strip">
    <KeyRound :size="16" />{{ t('secrets.savedValuesCannotBeReadBackReplaceA') }}
  </div>
  <section class="card">
    <AsyncState :pending="query.isPending.value" :error="query.error.value" @retry="query.refetch()">
      <EmptyState
        v-if="!query.data.value?.items.length"
        :title="t('secrets.noSecretsYet')"
        :description="t('secrets.storeTokensPemCertificatesProxyPasswordsAndShoutrrr')"
      />
      <div v-else class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th>{{ t('common.name') }}</th>
              <th>{{ t('secrets.referenceId') }}</th>
              <th>{{ t('secrets.updated') }}</th>
              <th />
            </tr>
          </thead>
          <tbody>
            <tr v-for="secret in query.data.value.items" :key="secret.id">
              <td class="monitor-name">
                {{ secret.name }}
              </td>
              <td>
                <code class="muted">{{ secret.id }}</code>
              </td>
              <td class="muted" un-text="10px">
                {{ formatDate(secret.updatedAt) }}
              </td>
              <td>
                <div v-if="isAdmin()" un-flex="~ gap-2">
                  <button class="button small" @click="edit(secret)">
                    <Pencil :size="12" />{{ t('secrets.replace') }}
                  </button><button
                    class="icon-button"
                    :aria-label="t('common.delete')"
                    @click="confirmDelete(secret)"
                  >
                    <Trash2 :size="14" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </AsyncState>
  </section>
  <Modal v-model:open="open" :title="form.id ? t('secrets.updateSecret') : t('common.addSecret')">
    <form id="secret-form" @submit.prevent="save">
      <Field :label="t('common.name')">
        <input v-model="form.name" required>
      </Field><Field
        :label="form.id ? t('secrets.replacementSecretValue') : t('secrets.secretValue')"
        un-mt="5"
      >
        <textarea v-model="form.value" required rows="6" autocomplete="off" spellcheck="false" />
      </Field>
      <p v-if="error" class="inline-error">
        {{ error }}
      </p>
    </form>
    <template #footer>
      <button class="button" @click="cancel">
        {{ t('common.cancel') }}
      </button><button class="button primary" form="secret-form" :disabled="saving">
        {{ t('secrets.saveSecret') }}
      </button>
    </template>
  </Modal><Modal
    v-model:open="deleteOpen"
    :title="t('secrets.deleteSecret')"
    :description="t('secrets.removeMonitorAndChannelReferencesBeforeDeletingA')"
  >
    <p>{{ deleteTarget?.name }}</p>
    <template #footer>
      <button class="button" @click="deleteOpen = false">
        {{ t('common.cancel') }}
      </button><button class="button danger" @click="remove">
        {{ t('common.delete') }}
      </button>
    </template>
  </Modal>
</template>
