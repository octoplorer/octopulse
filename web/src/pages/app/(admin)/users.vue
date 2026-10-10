<script setup lang="ts">
import type { User } from '../../../client/types.gen'
import { useMutation, useQuery } from '@pinia/colada'
import { useForm } from '@tanstack/vue-form'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  createUserMutation,
  deleteUserMutation,
  listUsersQuery,
  updateUserMutation,
} from '../../../client/@pinia/colada.gen'
import { Avatar } from '../../../components/common/avatar'
import { Badge } from '../../../components/ui/badge'
import { Banner } from '../../../components/ui/banner'
import { PageHeader } from '../../../components/ui/blocks/page-header'
import { Button } from '../../../components/ui/button'
import { Dialog } from '../../../components/ui/dialog'
import { Empty } from '../../../components/ui/empty'
import { Field, FieldGroup } from '../../../components/ui/field'
import { Input } from '../../../components/ui/input'
import { InputGroup, InputGroupAddon, InputGroupInput } from '../../../components/ui/input-group'
import { LayerCard, LayerCardPrimary } from '../../../components/ui/layer-card'
import { Loader } from '../../../components/ui/loader'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '../../../components/ui/select'
import { Switch } from '../../../components/ui/switch'
import { Table, TableBody, TableCell, TableContainer, TableHead, TableHeader, TablePagination, TableRow, TableToolbar } from '../../../components/ui/table'
import { currentUser, isAdmin } from '../../../composables/api'
import { languageOptions } from '../../../composables/i18n'
import { notify } from '../../../composables/notices'
import { formatDate, statusLabel, timezone } from '../../../composables/preferences'
import { errorText } from '../../../lib/errors'
import { clone } from '../../../lib/form'

const { t } = useI18n({ useScope: 'global' })

definePage({ meta: { title: 'navigation.members', roles: ['admin'], contentWidth: 'compact' } })

const createUser = useMutation(createUserMutation())
const updateUser = useMutation(updateUserMutation())
const deleteUser = useMutation(deleteUserMutation())

const query = useQuery({ ...listUsersQuery(), staleTime: 10000 })
const search = ref('')
const items = computed(() => {
  const term = search.value.trim().toLowerCase()
  return (query.data.value?.items || []).filter(user => `${user.name} ${user.username}`.toLowerCase().includes(term))
})
const open = ref(false)
const error = ref('')
function empty() {
  return {
    id: '',
    username: '',
    name: '',
    role: 'viewer' as User['role'],
    locale: 'zh-CN' as User['locale'],
    timezone: timezone.value,
    enabled: true,
    password: '',
  }
}
const formApi = useForm({
  defaultValues: empty(),
  onSubmit: async ({ value }) => {
    error.value = ''
    try {
      const { password, ...rest } = value
      if (value.id) {
        await updateUser.mutateAsync({
          path: { id: value.id },
          body: { ...rest, ...(password ? { password } : {}) },
        })
      }
      else {
        await createUser.mutateAsync({ body: { ...rest, password } })
      }
      formApi.setFieldValue('password', '')
      open.value = false
      await query.refresh()
      notify(t('users.member-saved'))
    }
    catch (e) {
      error.value = errorText(e)
    }
  },
})
const form = formApi.useSelector(state => state.values)
const saving = formApi.useSelector(state => state.isSubmitting)
const deleteTarget = ref<User | null>(null)
const deleteOpen = ref(false)
const deleting = ref(false)
function edit(user?: User) {
  if (formApi.state.isSubmitting)
    return
  formApi.reset(user ? { ...clone(user), password: '' } : empty())
  error.value = ''
  open.value = true
}
async function remove() {
  if (!deleteTarget.value || deleting.value)
    return
  deleting.value = true
  try {
    await deleteUser.mutateAsync({ path: { id: deleteTarget.value.id } })
    deleteOpen.value = false
    await query.refresh()
    notify(t('users.member-deleted'))
  }
  catch (e) {
    notify(errorText(e), 'error')
  }
  finally {
    deleting.value = false
  }
}
function confirmDelete(value: User) {
  deleteTarget.value = value
  deleteOpen.value = true
}
function cancel() {
  open.value = false
  formApi.setFieldValue('password', '')
}
</script>

<template>
  <PageHeader class="mb-6" :title="t('navigation.members')" :description="t('users.collaborate-with-administrator-operator-and-viewer-roles')">
    <template #actions>
      <Button v-if="isAdmin()" variant="primary" @click="edit()">
        <span w="15px" h="15px" aria-hidden="true" class="i-lucide-plus" />{{ t('common.add-member') }}
      </Button>
    </template>
  </PageHeader>
  <LayerCard>
    <LayerCardPrimary class="p-0!">
      <TableToolbar>
        <InputGroup class="w-full max-w-sm">
          <InputGroupAddon><span class="i-lucide-search size-4" aria-hidden="true" /></InputGroupAddon>
          <InputGroupInput v-model="search" type="search" :placeholder="t('users.search-members-placeholder')" :aria-label="t('users.search-members')" />
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
        <Empty v-if="!items.length" size="sm" class="rounded-none border-none" :title="search ? t('users.no-matching-members') : t('users.no-members')">
          <Button v-if="search" @click="search = ''">
            {{ t('common.clear-filters') }}
          </Button>
          <Button v-else-if="isAdmin()" variant="primary" @click="edit()">
            {{ t('common.add-member') }}
          </Button>
        </Empty>
        <TableContainer v-else :scroll-label="t('common.scroll-table')">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{{ t('users.member') }}</TableHead>
                <TableHead class="[@container_workspace_(max-width:_700px)]:hidden">
                  {{ t('common.role') }}
                </TableHead>
                <TableHead class="[@container_workspace_(max-width:_700px)]:hidden">
                  {{ t('common.status') }}
                </TableHead>
                <TableHead class="[@container_workspace_(max-width:_700px)]:hidden">
                  {{ t('common.created') }}
                </TableHead>
                <TableHead class="w-24">
                  <span class="sr-only">{{ t('common.edit') }}</span>
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="user in items" :key="user.id">
                <TableCell>
                  <div class="flex min-w-0 items-center gap-3">
                    <Avatar class="shrink-0">
                      {{ user.name?.[0] || user.username[0] }}
                    </Avatar><span class="min-w-0"><span class="block font-medium [overflow-wrap:anywhere]">{{ user.name || user.username }}</span><span class="monitor-sub block [overflow-wrap:anywhere]" un-text="12px subtle" mt="3px" max-w="300px">{{ user.username
                    }}{{ user.id === currentUser?.id ? ` · ${t('users.you')}` : '' }}</span></span>
                  </div>
                  <div class="mt-2 flex flex-wrap items-center gap-2 [@container_workspace_(width_>_700px)]:hidden">
                    <Badge>{{ statusLabel(user.role) }}</Badge>
                    <span class="text-size-xs text-subtle">{{ user.enabled ? t('common.enabled') : t('common.disabled') }}</span>
                  </div>
                  <span class="mt-1 block text-size-xs text-subtle [@container_workspace_(width_>_700px)]:hidden">{{ t('common.created') }} · {{ formatDate(user.createdAt) }}</span>
                </TableCell>
                <TableCell class="[@container_workspace_(max-width:_700px)]:hidden">
                  <Badge>{{ statusLabel(user.role) }}</Badge>
                </TableCell>
                <TableCell class="[@container_workspace_(max-width:_700px)]:hidden">
                  {{ user.enabled ? t('common.enabled') : t('common.disabled') }}
                </TableCell>
                <TableCell class="text-subtle [@container_workspace_(max-width:_700px)]:hidden">
                  {{ formatDate(user.createdAt) }}
                </TableCell>
                <TableCell>
                  <div v-if="isAdmin()" class="flex justify-end gap-2">
                    <Button :aria-label="t('common.edit')" shape="square" @click="edit(user)">
                      <span w="14px" h="14px" aria-hidden="true" class="i-lucide-pencil" />
                    </Button><Button :disabled="user.id === currentUser?.id" :aria-label="t('common.delete')" shape="square" @click="confirmDelete(user)">
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
        {{ t('users.showing-members', { shown: items.length, total: query.data.value.items.length }) }}
      </TablePagination>
    </LayerCardPrimary>
  </LayerCard>
  <p class="mt-4 max-w-prose text-size-sm text-subtle">
    {{ t('users.operators-manage-monitors-maintenance-and-public-pages-administrators') }}
  </p>
  <Dialog v-model:open="open" :close-label="t('common.close')" size="lg" :title="form.id ? t('users.edit-member') : t('common.add-member')">
    <form id="user-form" @submit.prevent="formApi.handleSubmit()">
      <FieldGroup>
        <formApi.Field v-slot="{ field }" name="username">
          <Field :label="t('common.username')">
            <Input :name="field.name" :model-value="field.state.value" required :disabled="!!form.id" @update:model-value="field.handleChange(String($event ?? ''))" @blur="field.handleBlur" />
          </Field>
        </formApi.Field>
        <formApi.Field v-slot="{ field }" name="name">
          <Field :label="t('common.display-name')">
            <Input :name="field.name" :model-value="field.state.value" @update:model-value="field.handleChange(String($event ?? ''))" @blur="field.handleBlur" />
          </Field>
        </formApi.Field>
        <formApi.Field v-slot="{ field }" name="role">
          <Field :label="t('common.role')">
            <Select :name="field.name" :model-value="field.state.value" @update:model-value="field.handleChange" @focusout="field.handleBlur">
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="admin">
                    {{ t('users.administrator') }}
                  </SelectItem>
                  <SelectItem value="operator">
                    {{ t('users.operator') }}
                  </SelectItem>
                  <SelectItem value="viewer">
                    {{ t('users.viewer') }}
                  </SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </Field>
        </formApi.Field>
        <formApi.Field v-slot="{ field }" name="locale">
          <Field :label="t('common.language')">
            <Select :name="field.name" :model-value="field.state.value" @update:model-value="field.handleChange" @focusout="field.handleBlur">
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem v-for="language in languageOptions" :key="language.value" :value="language.value">
                    {{ language.label }}
                  </SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </Field>
        </formApi.Field>
        <formApi.Field v-slot="{ field }" name="timezone">
          <Field :label="t('common.display-time-zone')" class="span-full">
            <Input :name="field.name" :model-value="field.state.value" required @update:model-value="field.handleChange(String($event ?? ''))" @blur="field.handleBlur" />
          </Field>
        </formApi.Field>
        <formApi.Field v-slot="{ field }" name="password">
          <Field :label="form.id ? t('users.new-password-leave-empty-to-keep') : t('common.password')" :description="t('common.at-least-12-characters-up-to-72-bytes')" class="span-full">
            <Input :name="field.name" :model-value="field.state.value" type="password" :required="!form.id" minlength="12" maxlength="72" autocomplete="new-password" @update:model-value="field.handleChange(String($event ?? ''))" @blur="field.handleBlur" />
          </Field>
        </formApi.Field>
        <formApi.Field v-slot="{ field }" name="enabled">
          <div class="span-full">
            <Switch :model-value="field.state.value" :label="t('users.enable-account')" @update:model-value="field.handleChange" @focusout="field.handleBlur" />
          </div>
        </formApi.Field>
      </FieldGroup>
      <Banner v-if="error" variant="error" class="mt-4">
        {{ error }}
      </Banner>
    </form>
    <template #footer>
      <Button :disabled="saving" @click="cancel">
        {{ t('common.cancel') }}
      </Button><Button type="submit" form="user-form" :loading="saving" variant="primary">
        {{ t('users.save-member') }}
      </Button>
    </template>
  </Dialog><Dialog v-model:open="deleteOpen" :close-label="t('common.close')" :title="t('users.delete-member')">
    <p>{{ deleteTarget?.name || deleteTarget?.username }}</p>
    <template #footer>
      <Button :disabled="deleting" @click="deleteOpen = false">
        {{ t('common.cancel') }}
      </Button><Button variant="destructive" :loading="deleting" @click="remove">
        {{ t('common.delete') }}
      </Button>
    </template>
  </Dialog>
</template>
