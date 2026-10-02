<script setup lang="ts">
import { ref, reactive } from 'vue'
import { Plus, Pencil, Users, Trash2 } from '@lucide/vue'
import { useCollection } from '../../../lib/data'
import type { User } from '../../../lib/types'
import { api, isAdmin, currentUser } from '../../../lib/api'
import { t, formatDate, timezone, statusLabel } from '../../../lib/preferences'
import { notify, errorText } from '../../../lib/notices'
import { clone } from '../../../lib/form'
import PageHeader from '../../../components/PageHeader.vue'
import Field from '../../../components/Field.vue'
import Toggle from '../../../components/Toggle.vue'
import Modal from '../../../components/Modal.vue'
import AsyncState from '../../../components/AsyncState.vue'
import EmptyState from '../../../components/EmptyState.vue'

definePage({ meta: { title: ['成员与权限', 'Members'], roles: ['admin'] } })

const query = useCollection<User>('users'),
  open = ref(false),
  saving = ref(false),
  error = ref(''),
  form = reactive({
    id: '',
    username: '',
    name: '',
    role: 'viewer' as User['role'],
    locale: 'zh-CN' as User['locale'],
    timezone: timezone.value,
    enabled: true,
    password: '',
  }),
  deleteTarget = ref<User | null>(null),
  deleteOpen = ref(false)
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
    await api(form.id ? `users/${form.id}` : 'users', {
      method: form.id ? 'PATCH' : 'POST',
      body: { ...rest, ...(password ? { password } : {}) },
    })
    form.password = ''
    open.value = false
    await query.refresh()
    notify(t('成员已保存', 'Member saved'))
  } catch (e) {
    error.value = errorText(e)
  } finally {
    saving.value = false
  }
}
async function remove() {
  if (!deleteTarget.value) return
  try {
    await api(`users/${deleteTarget.value.id}`, { method: 'DELETE' })
    deleteOpen.value = false
    await query.refresh()
    notify(t('成员已删除', 'Member deleted'))
  } catch (e) {
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
    :title="t('成员与权限', 'Members')"
    :description="
      t(
        '以管理员、操作员和只读角色协作管理。',
        'Collaborate with administrator, operator, and viewer roles.',
      )
    "
    ><button v-if="isAdmin()" class="button primary" @click="edit()">
      <Plus :size="15" />{{ t('添加成员', 'Add member') }}
    </button></PageHeader
  >
  <section class="card">
    <AsyncState :pending="query.isPending.value" :error="query.error.value" @retry="query.refresh()"
      ><EmptyState v-if="!query.data.value?.items.length" :title="t('暂无成员', 'No members')" />
      <div v-else class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th>{{ t('成员', 'Member') }}</th>
              <th>{{ t('角色', 'Role') }}</th>
              <th>{{ t('状态', 'Status') }}</th>
              <th>{{ t('创建时间', 'Created') }}</th>
              <th />
            </tr>
          </thead>
          <tbody>
            <tr v-for="user in query.data.value.items" :key="user.id">
              <td>
                <div un-flex="~ items-center gap-3">
                  <span class="user-avatar">{{ user.name?.[0] || user.username[0] }}</span
                  ><span
                    ><span class="monitor-name">{{ user.name || user.username }}</span
                    ><span class="monitor-sub"
                      >{{ user.username
                      }}{{ user.id === currentUser?.id ? ` · ${t('你', 'You')}` : '' }}</span
                    ></span
                  >
                </div>
              </td>
              <td>
                <span class="pill">{{ statusLabel(user.role) }}</span>
              </td>
              <td>{{ user.enabled ? t('启用', 'Enabled') : t('停用', 'Disabled') }}</td>
              <td class="muted" un-text="10px">{{ formatDate(user.createdAt) }}</td>
              <td>
                <div v-if="isAdmin()" un-flex="~ gap-2">
                  <button class="icon-button" @click="edit(user)" :aria-label="t('编辑', 'Edit')">
                    <Pencil :size="14" /></button
                  ><button
                    class="icon-button"
                    :disabled="user.id === currentUser?.id"
                    @click="confirmDelete(user)"
                    :aria-label="t('删除', 'Delete')"
                  >
                    <Trash2 :size="14" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table></div
    ></AsyncState>
  </section>
  <p class="note" un-mt="5">
    {{
      t(
        '操作员可管理监控、维护和公开页面；管理员额外管理账号、秘密、通知渠道与系统。只读成员只能查看脱敏结果。',
        'Operators manage monitors, maintenance, and public pages. Administrators also manage accounts, secrets, channels, and system settings. Viewers read redacted results.',
      )
    }}
  </p>
  <Modal
    v-model:open="open"
    :title="form.id ? t('编辑成员', 'Edit member') : t('添加成员', 'Add member')"
    ><form id="user-form" @submit.prevent="save">
      <div class="form-grid">
        <Field :label="t('用户名', 'Username')"
          ><input v-model="form.username" required :disabled="!!form.id" /></Field
        ><Field :label="t('显示名称', 'Display name')"><input v-model="form.name" /></Field
        ><Field :label="t('角色', 'Role')"
          ><select v-model="form.role">
            <option value="admin">{{ t('管理员', 'Administrator') }}</option>
            <option value="operator">{{ t('操作员', 'Operator') }}</option>
            <option value="viewer">{{ t('只读', 'Viewer') }}</option>
          </select></Field
        ><Field :label="t('语言', 'Language')"
          ><select v-model="form.locale">
            <option value="zh-CN">简体中文</option>
            <option value="en">English</option>
          </select></Field
        ><Field class="span-full" :label="t('显示时区', 'Display time zone')"
          ><input v-model="form.timezone" required /></Field
        ><Field
          class="span-full"
          :label="
            form.id
              ? t('新密码（留空保留）', 'New password (leave empty to keep)')
              : t('密码', 'Password')
          "
          :hint="t('至少 12 字符，最多 72 字节。', 'At least 12 characters, up to 72 bytes.')"
          ><input
            v-model="form.password"
            type="password"
            :required="!form.id"
            minlength="12"
            maxlength="72"
            autocomplete="new-password"
        /></Field>
        <div class="span-full">
          <Toggle v-model="form.enabled" :label="t('启用账号', 'Enable account')" />
        </div>
      </div>
      <p v-if="error" class="inline-error">{{ error }}</p>
    </form>
    <template #footer
      ><button class="button" @click="cancel">
        {{ t('取消', 'Cancel') }}</button
      ><button class="button primary" form="user-form" :disabled="saving">
        {{ t('保存成员', 'Save member') }}
      </button></template
    ></Modal
  ><Modal v-model:open="deleteOpen" :title="t('删除成员', 'Delete member')"
    ><p>{{ deleteTarget?.name || deleteTarget?.username }}</p>
    <template #footer
      ><button class="button" @click="deleteOpen = false">{{ t('取消', 'Cancel') }}</button
      ><button class="button danger" @click="remove">{{ t('删除', 'Delete') }}</button></template
    ></Modal
  >
</template>
