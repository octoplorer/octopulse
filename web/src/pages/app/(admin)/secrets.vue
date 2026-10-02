<script setup lang="ts">
import { ref, reactive } from 'vue'
import { Plus, KeyRound, Pencil, Trash2 } from '@lucide/vue'
import { useCollection } from '../../../lib/data'
import type { Secret } from '../../../lib/types'
import { isAdmin } from '../../../lib/api'
import * as sdk from '../../../client/sdk.gen'
import { listSecretsQuery } from '../../../client/@pinia/colada.gen'
import { t, formatDate } from '../../../lib/preferences'
import { notify, errorText } from '../../../lib/notices'
import PageHeader from '../../../components/PageHeader.vue'
import Field from '../../../components/Field.vue'
import Modal from '../../../components/Modal.vue'
import AsyncState from '../../../components/AsyncState.vue'
import EmptyState from '../../../components/EmptyState.vue'

definePage({ meta: { title: ['秘密凭据', 'Secrets'], roles: ['admin'] } })

const query = useCollection<Secret>('secrets', listSecretsQuery()),
  open = ref(false),
  saving = ref(false),
  error = ref(''),
  form = reactive({ id: '', name: '', value: '' }),
  deleteTarget = ref<Secret | null>(null),
  deleteOpen = ref(false)
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
    if (form.id) await sdk.updateSecret({ path: { id: form.id }, body, throwOnError: true })
    else await sdk.createSecret({ body, throwOnError: true })
    form.value = ''
    open.value = false
    await query.refresh()
    notify(t('秘密已保存', 'Secret saved'))
  } catch (e) {
    error.value = errorText(e)
  } finally {
    saving.value = false
  }
}
async function remove() {
  if (!deleteTarget.value) return
  try {
    await sdk.deleteSecret({ path: { id: deleteTarget.value.id }, throwOnError: true })
    deleteOpen.value = false
    await query.refresh()
    notify(t('秘密已删除', 'Secret deleted'))
  } catch (e) {
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
    :title="t('秘密凭据', 'Secrets')"
    :description="
      t(
        '集中管理请求、TLS、代理和通知使用的敏感值。',
        'Manage sensitive values used by requests, TLS, proxies, and notifications.',
      )
    "
    ><button v-if="isAdmin()" class="button primary" @click="edit()">
      <Plus :size="15" />{{ t('添加秘密', 'Add secret') }}
    </button></PageHeader
  >
  <div class="alert-strip">
    <KeyRound :size="16" />{{
      t(
        '已保存的值不会回读。更新时可替换内容，监控与渠道继续使用同一个引用。',
        'Saved values cannot be read back. Replace a value while monitors and channels keep the same reference.',
      )
    }}
  </div>
  <section class="card">
    <AsyncState :pending="query.isPending.value" :error="query.error.value" @retry="query.refresh()"
      ><EmptyState
        v-if="!query.data.value?.items.length"
        :title="t('暂无秘密引用', 'No secrets yet')"
        :description="
          t(
            '可保存 Token、PEM 证书、代理密码和 Shoutrrr 服务 URL。',
            'Store tokens, PEM certificates, proxy passwords, and Shoutrrr service URLs.',
          )
        " />
      <div v-else class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th>{{ t('名称', 'Name') }}</th>
              <th>{{ t('引用 ID', 'Reference ID') }}</th>
              <th>{{ t('最近更新', 'Updated') }}</th>
              <th />
            </tr>
          </thead>
          <tbody>
            <tr v-for="secret in query.data.value.items" :key="secret.id">
              <td class="monitor-name">{{ secret.name }}</td>
              <td>
                <code class="muted">{{ secret.id }}</code>
              </td>
              <td class="muted" un-text="10px">{{ formatDate(secret.updatedAt) }}</td>
              <td>
                <div v-if="isAdmin()" un-flex="~ gap-2">
                  <button class="button small" @click="edit(secret)">
                    <Pencil :size="12" />{{ t('替换', 'Replace') }}</button
                  ><button
                    class="icon-button"
                    @click="confirmDelete(secret)"
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
  <Modal
    v-model:open="open"
    :title="form.id ? t('更新秘密', 'Update secret') : t('添加秘密', 'Add secret')"
    ><form id="secret-form" @submit.prevent="save">
      <Field :label="t('名称', 'Name')"><input v-model="form.name" required /></Field
      ><Field
        :label="
          form.id ? t('替换后的秘密值', 'Replacement secret value') : t('秘密值', 'Secret value')
        "
        un-mt="5"
      >
        <textarea v-model="form.value" required rows="6" autocomplete="off" spellcheck="false" />
      </Field>
      <p v-if="error" class="inline-error">{{ error }}</p>
    </form>
    <template #footer
      ><button class="button" @click="cancel">
        {{ t('取消', 'Cancel') }}</button
      ><button class="button primary" form="secret-form" :disabled="saving">
        {{ t('保存秘密', 'Save secret') }}
      </button></template
    ></Modal
  ><Modal
    v-model:open="deleteOpen"
    :title="t('删除秘密', 'Delete secret')"
    :description="
      t(
        '被监控或渠道引用的秘密需要先解除引用。',
        'Remove monitor and channel references before deleting a secret.',
      )
    "
    ><p>{{ deleteTarget?.name }}</p>
    <template #footer
      ><button class="button" @click="deleteOpen = false">{{ t('取消', 'Cancel') }}</button
      ><button class="button danger" @click="remove">{{ t('删除', 'Delete') }}</button></template
    ></Modal
  >
</template>
