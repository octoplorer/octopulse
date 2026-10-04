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
import { Alert } from '../../../components/ui/alert'
import { Avatar } from '../../../components/ui/avatar'
import { Badge } from '../../../components/ui/badge'
import { Button } from '../../../components/ui/button'
import { Card } from '../../../components/ui/card'
import { FieldError, FieldGroup } from '../../../components/ui/field'
import { Table, TableBody, TableCell, TableContainer, TableHead, TableHeader, TableRow } from '../../../components/ui/table'
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
  <PageHeader :title="t('navigation.members')" :description="t('users.collaborateWithAdministratorOperatorAndViewerRoles')">
    <Button v-if="isAdmin()" variant="primary" @click="edit()">
      <span w="15px" h="15px" aria-hidden="true" class="i-lucide-plus" />{{ t('common.addMember') }}
    </Button>
  </PageHeader>
  <Card as="section">
    <AsyncState :pending="query.isPending.value" :error="query.error.value" @retry="query.refetch()">
      <EmptyState v-if="!query.data.value?.items.length" :title="t('users.noMembers')" />
      <TableContainer v-else>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{{ t('users.member') }}</TableHead>
              <TableHead>{{ t('common.role') }}</TableHead>
              <TableHead>{{ t('common.status') }}</TableHead>
              <TableHead>{{ t('common.created') }}</TableHead>
              <TableHead />
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="user in query.data.value.items" :key="user.id">
              <TableCell>
                <div flex="~ items-center gap-3">
                  <Avatar>{{ user.name?.[0] || user.username[0] }}</Avatar><span><span class="monitor-name block" font="600" un-text="13px">{{ user.name || user.username }}</span><span class="monitor-sub block [overflow-wrap:anywhere]" un-text="12px $muted" mt="3px" max-w="300px">{{ user.username
                  }}{{ user.id === currentUser?.id ? ` · ${t('users.you')}` : '' }}</span></span>
                </div>
              </TableCell>
              <TableCell>
                <Badge>{{ statusLabel(user.role) }}</Badge>
              </TableCell>
              <TableCell>{{ user.enabled ? t('common.enabled') : t('common.disabled') }}</TableCell>
              <TableCell class="muted" un-text="13px $muted">
                {{ formatDate(user.createdAt) }}
              </TableCell>
              <TableCell>
                <div v-if="isAdmin()" flex="~ gap-2">
                  <Button :aria-label="t('common.edit')" size="icon" @click="edit(user)">
                    <span w="14px" h="14px" aria-hidden="true" class="i-lucide-pencil" />
                  </Button><Button :disabled="user.id === currentUser?.id" :aria-label="t('common.delete')" size="icon" @click="confirmDelete(user)">
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
  <Alert mt="5" as="p" variant="default">
    {{ t('users.operatorsManageMonitorsMaintenanceAndPublicPagesAdministrators') }}
  </Alert>
  <Modal v-model:open="open" :title="form.id ? t('users.editMember') : t('common.addMember')">
    <form id="user-form" @submit.prevent="save">
      <FieldGroup>
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
            <option v-for="language in languageOptions" :key="language.value" :value="language.value">
              {{ language.label }}
            </option>
          </select>
        </Field><Field :label="t('common.displayTimeZone')" class="span-full">
          <input v-model="form.timezone" required>
        </Field><Field :label="form.id ? t('users.newPasswordLeaveEmptyToKeep') : t('common.password')" :hint="t('common.atLeast12CharactersUpTo72Bytes')" class="span-full">
          <input v-model="form.password" type="password" :required="!form.id" minlength="12" maxlength="72" autocomplete="new-password">
        </Field>
        <div class="span-full">
          <Toggle v-model="form.enabled" :label="t('users.enableAccount')" />
        </div>
      </FieldGroup>
      <FieldError v-if="error" as="p" py="10px" px="0">
        {{ error }}
      </FieldError>
    </form>
    <template #footer>
      <Button @click="cancel">
        {{ t('common.cancel') }}
      </Button><Button form="user-form" :disabled="saving" variant="primary">
        {{ t('users.saveMember') }}
      </Button>
    </template>
  </Modal><Modal v-model:open="deleteOpen" :title="t('users.deleteMember')">
    <p>{{ deleteTarget?.name || deleteTarget?.username }}</p>
    <template #footer>
      <Button @click="deleteOpen = false">
        {{ t('common.cancel') }}
      </Button><Button variant="danger" @click="remove">
        {{ t('common.delete') }}
      </Button>
    </template>
  </Modal>
</template>
