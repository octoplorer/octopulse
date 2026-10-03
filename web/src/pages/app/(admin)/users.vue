<script setup lang="ts">
import type { User } from '../../../client/types.gen'
import { useMutation, useQuery } from '@pinia/colada'
import { reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  createUserMutation,
  deleteUserMutation,
  listUsersQuery,
  updateUserMutation,
} from '../../../client/@pinia/colada.gen'
import AsyncState from '../../../components/AsyncState.vue'
import EmptyState from '../../../components/EmptyState.vue'
import Field from '../../../components/Field.vue'
import Modal from '../../../components/Modal.vue'
import PageHeader from '../../../components/PageHeader.vue'
import Toggle from '../../../components/Toggle.vue'
import { currentUser, isAdmin } from '../../../composables/api'
import { languageOptions } from '../../../composables/i18n'
import { notify } from '../../../composables/notices'
import { formatDate, statusLabel, timezone } from '../../../composables/preferences'
import { errorText } from '../../../lib/errors'
import { clone } from '../../../lib/form'

const { t } = useI18n({ useScope: 'global' })

definePage({ meta: { title: 'navigation.members', roles: ['admin'] } })

const createUser = useMutation(createUserMutation())
const updateUser = useMutation(updateUserMutation())
const deleteUser = useMutation(deleteUserMutation())

const query = useQuery({ ...listUsersQuery(), staleTime: 10000 })
const open = ref(false)
const saving = ref(false)
const error = ref('')
const form = reactive({
  id: '',
  username: '',
  name: '',
  role: 'viewer' as User['role'],
  locale: 'zh-CN' as User['locale'],
  timezone: timezone.value,
  enabled: true,
  password: '',
})
const deleteTarget = ref<User | null>(null)
const deleteOpen = ref(false)
function edit(user?: User) {
  Object.assign(
    form,
    user
      ? { ...clone(user), password: '' }
      : {
          id: '',
          username: '',
          name: '',
          role: 'viewer',
          locale: 'zh-CN',
          timezone: timezone.value,
          enabled: true,
          password: '',
        },
  )
  error.value = ''
  open.value = true
}
async function save() {
  saving.value = true
  error.value = ''
  try {
    const { password, ...rest } = form
    if (form.id) {
      await updateUser.mutateAsync({
        path: { id: form.id },
        body: { ...rest, ...(password ? { password } : {}) },
      })
    }
    else {
      await createUser.mutateAsync({ body: { ...rest, password } })
    }
    form.password = ''
    open.value = false
    await query.refresh()
    notify(t('users.memberSaved'))
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
    await deleteUser.mutateAsync({ path: { id: deleteTarget.value.id } })
    deleteOpen.value = false
    await query.refresh()
    notify(t('users.memberDeleted'))
  }
  catch (e) {
    notify(errorText(e), 'error')
  }
}
function confirmDelete(value: User) {
  deleteTarget.value = value
  deleteOpen.value = true
}
function cancel() {
  open.value = false
  form.password = ''
}
</script>

<template>
  <PageHeader
    :title="t('navigation.members')"
    :description="t('users.collaborateWithAdministratorOperatorAndViewerRoles')"
  >
    <button v-if="isAdmin()" class="button primary" @click="edit()">
      <span class="i-lucide-plus" un-w="15px" un-h="15px" aria-hidden="true" />{{ t('common.addMember') }}
    </button>
  </PageHeader>
  <section class="card">
    <AsyncState :pending="query.isPending.value" :error="query.error.value" @retry="query.refetch()">
      <EmptyState v-if="!query.data.value?.items.length" :title="t('users.noMembers')" />
      <div v-else class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th>{{ t('users.member') }}</th>
              <th>{{ t('common.role') }}</th>
              <th>{{ t('common.status') }}</th>
              <th>{{ t('common.created') }}</th>
              <th />
            </tr>
          </thead>
          <tbody>
            <tr v-for="user in query.data.value.items" :key="user.id">
              <td>
                <div un-flex="~ items-center gap-3">
                  <span class="user-avatar">{{ user.name?.[0] || user.username[0] }}</span><span><span class="monitor-name">{{ user.name || user.username }}</span><span class="monitor-sub">{{ user.username
                  }}{{ user.id === currentUser?.id ? ` · ${t('users.you')}` : '' }}</span></span>
                </div>
              </td>
              <td>
                <span class="pill">{{ statusLabel(user.role) }}</span>
              </td>
              <td>{{ user.enabled ? t('common.enabled') : t('common.disabled') }}</td>
              <td class="muted" un-text="10px">
                {{ formatDate(user.createdAt) }}
              </td>
              <td>
                <div v-if="isAdmin()" un-flex="~ gap-2">
                  <button class="icon-button" :aria-label="t('common.edit')" @click="edit(user)">
                    <span class="i-lucide-pencil" un-w="14px" un-h="14px" aria-hidden="true" />
                  </button><button
                    class="icon-button"
                    :disabled="user.id === currentUser?.id"
                    :aria-label="t('common.delete')"
                    @click="confirmDelete(user)"
                  >
                    <span class="i-lucide-trash-2" un-w="14px" un-h="14px" aria-hidden="true" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </AsyncState>
  </section>
  <p class="note" un-mt="5">
    {{ t('users.operatorsManageMonitorsMaintenanceAndPublicPagesAdministrators') }}
  </p>
  <Modal v-model:open="open" :title="form.id ? t('users.editMember') : t('common.addMember')">
    <form id="user-form" @submit.prevent="save">
      <div class="form-grid">
        <Field :label="t('common.username')">
          <input v-model="form.username" required :disabled="!!form.id">
        </Field><Field :label="t('common.displayName')">
          <input v-model="form.name">
        </Field><Field :label="t('common.role')">
          <select v-model="form.role">
            <option value="admin">
              {{ t('users.administrator') }}
            </option>
            <option value="operator">
              {{ t('users.operator') }}
            </option>
            <option value="viewer">
              {{ t('users.viewer') }}
            </option>
          </select>
        </Field><Field :label="t('common.language')">
          <select v-model="form.locale">
            <option
              v-for="language in languageOptions"
              :key="language.value"
              :value="language.value"
            >
              {{ language.label }}
            </option>
          </select>
        </Field><Field class="span-full" :label="t('common.displayTimeZone')">
          <input v-model="form.timezone" required>
        </Field><Field
          class="span-full"
          :label="form.id ? t('users.newPasswordLeaveEmptyToKeep') : t('common.password')"
          :hint="t('common.atLeast12CharactersUpTo72Bytes')"
        >
          <input
            v-model="form.password"
            type="password"
            :required="!form.id"
            minlength="12"
            maxlength="72"
            autocomplete="new-password"
          >
        </Field>
        <div class="span-full">
          <Toggle v-model="form.enabled" :label="t('users.enableAccount')" />
        </div>
      </div>
      <p v-if="error" class="inline-error">
        {{ error }}
      </p>
    </form>
    <template #footer>
      <button class="button" @click="cancel">
        {{ t('common.cancel') }}
      </button><button class="button primary" form="user-form" :disabled="saving">
        {{ t('users.saveMember') }}
      </button>
    </template>
  </Modal><Modal v-model:open="deleteOpen" :title="t('users.deleteMember')">
    <p>{{ deleteTarget?.name || deleteTarget?.username }}</p>
    <template #footer>
      <button class="button" @click="deleteOpen = false">
        {{ t('common.cancel') }}
      </button><button class="button danger" @click="remove">
        {{ t('common.delete') }}
      </button>
    </template>
  </Modal>
</template>
