<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { Plus, Bell, Send, Pencil, Trash2, RefreshCw } from '@lucide/vue'
import { Tabs } from '@ark-ui/vue/tabs'
import { useCollection } from '../lib/data'
import type { Channel, Secret, Delivery, Monitor } from '../lib/types'
import { api, isAdmin } from '../lib/api'
import { t, formatDate, statusLabel } from '../lib/preferences'
import { notify, errorText } from '../lib/notices'
import { clone } from '../lib/form'
import PageHeader from '../components/PageHeader.vue'
import Field from '../components/Field.vue'
import Toggle from '../components/Toggle.vue'
import SecretSelect from '../components/SecretSelect.vue'
import Modal from '../components/Modal.vue'
import AsyncState from '../components/AsyncState.vue'
import EmptyState from '../components/EmptyState.vue'
const query = useCollection<Channel>('channels'),
  secrets = useCollection<Secret>('secrets'),
  deliveries = useCollection<Delivery>('deliveries'),
  monitors = useCollection<Monitor>('monitors'),
  open = ref(false),
  saving = ref(false),
  error = ref(''),
  tab = ref('channels'),
  testing = ref(''),
  deleteTarget = ref<Channel | null>(null),
  deleteOpen = ref(false)
const empty = (): Channel => ({
    id: '',
    name: '',
    serviceUrlSecretId: '',
    enabled: true,
    createdAt: 0,
    updatedAt: 0,
  }),
  form = reactive(empty())
function edit(channel?: Channel) {
  Object.assign(form, channel ? clone(channel) : empty())
  error.value = ''
  open.value = true
}
async function save() {
  saving.value = true
  error.value = ''
  try {
    await api(form.id ? `channels/${form.id}` : 'channels', {
      method: form.id ? 'PATCH' : 'POST',
      body: form,
    })
    open.value = false
    await query.refresh()
    notify(t('通知渠道已保存', 'Channel saved'))
  } catch (e) {
    error.value = errorText(e)
  } finally {
    saving.value = false
  }
}
async function test(channel: Channel) {
  testing.value = channel.id
  try {
    await api(`channels/${channel.id}/test`, { method: 'POST' })
    notify(t('测试通知已提交，请查看投递记录。', 'Test submitted. Check delivery history.'))
  } catch (e) {
    notify(errorText(e), 'error')
  } finally {
    await deliveries.refresh()
    testing.value = ''
  }
}
async function remove() {
  if (!deleteTarget.value) return
  try {
    await api(`channels/${deleteTarget.value.id}`, { method: 'DELETE' })
    deleteOpen.value = false
    await query.refresh()
    notify(t('通知渠道已删除', 'Channel deleted'))
  } catch (e) {
    notify(errorText(e), 'error')
  }
}
function confirmDelete(value: Channel) {
  deleteTarget.value = value
  deleteOpen.value = true
}
function refresh() {
  query.refresh()
  deliveries.refresh()
}
</script>
<template>
  <PageHeader
    :title="t('通知与投递', 'Notifications')"
    :description="
      t(
        '通过 Shoutrrr 连接渠道，持久记录每一次投递。',
        'Connect channels through Shoutrrr and keep every delivery durable.',
      )
    "
    ><button class="button" @click="refresh">
      <RefreshCw :size="14" />{{ t('刷新', 'Refresh') }}</button
    ><button v-if="isAdmin()" class="button primary" @click="edit()">
      <Plus :size="15" />{{ t('添加渠道', 'Add channel') }}
    </button></PageHeader
  >
  <section class="card">
    <Tabs.Root v-model="tab"
      ><Tabs.List class="tabs-list"
        ><Tabs.Trigger class="tabs-trigger" value="channels">{{
          t('通知渠道', 'Channels')
        }}</Tabs.Trigger
        ><Tabs.Trigger class="tabs-trigger" value="deliveries">{{
          t('投递记录', 'Deliveries')
        }}</Tabs.Trigger></Tabs.List
      ><Tabs.Content value="channels"
        ><AsyncState
          :pending="query.isPending.value"
          :error="query.error.value"
          @retry="query.refresh()"
          ><EmptyState
            v-if="!query.data.value?.items.length"
            :title="t('连接你的通知渠道', 'Connect a notification channel')"
            :description="
              t(
                '先保存 Shoutrrr 服务 URL 为秘密，再在这里引用。',
                'Store a Shoutrrr service URL as a secret, then reference it here.',
              )
            "
            ><RouterLink v-if="isAdmin()" to="/app/secrets" class="button">{{
              t('管理秘密', 'Manage secrets')
            }}</RouterLink></EmptyState
          >
          <div v-else class="table-wrap">
            <table class="data-table">
              <thead>
                <tr>
                  <th>{{ t('渠道', 'Channel') }}</th>
                  <th>{{ t('秘密引用', 'Secret reference') }}</th>
                  <th>{{ t('状态', 'Status') }}</th>
                  <th>{{ t('更新时间', 'Updated') }}</th>
                  <th />
                </tr>
              </thead>
              <tbody>
                <tr v-for="channel in query.data.value.items" :key="channel.id">
                  <td>
                    <span class="monitor-name" un-flex="~ items-center gap-2"
                      ><Bell :size="15" />{{ channel.name }}</span
                    >
                  </td>
                  <td class="muted">
                    {{
                      secrets.data.value?.items.find((x) => x.id === channel.serviceUrlSecretId)
                        ?.name || '—'
                    }}
                  </td>
                  <td>
                    <span class="pill">{{
                      channel.enabled ? t('启用', 'Enabled') : t('停用', 'Disabled')
                    }}</span>
                  </td>
                  <td class="muted" un-text="10px">{{ formatDate(channel.updatedAt) }}</td>
                  <td>
                    <div v-if="isAdmin()" un-flex="~ items-center gap-2">
                      <button
                        class="button small"
                        :disabled="testing === channel.id || !channel.enabled"
                        @click="test(channel)"
                      >
                        <Send :size="12" />{{ t('测试', 'Test') }}</button
                      ><button
                        class="icon-button"
                        @click="edit(channel)"
                        :aria-label="t('编辑', 'Edit')"
                      >
                        <Pencil :size="14" /></button
                      ><button
                        class="icon-button"
                        @click="confirmDelete(channel)"
                        :aria-label="t('删除', 'Delete')"
                      >
                        <Trash2 :size="14" />
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table></div></AsyncState></Tabs.Content
      ><Tabs.Content value="deliveries"
        ><AsyncState
          :pending="deliveries.isPending.value"
          :error="deliveries.error.value"
          @retry="deliveries.refresh()"
          ><EmptyState
            v-if="!deliveries.data.value?.items.length"
            :title="t('暂无投递记录', 'No deliveries yet')"
            :description="
              t(
                '故障、恢复或渠道测试会创建独立的持久投递任务。',
                'Outages, recoveries, and channel tests create durable delivery jobs.',
              )
            "
          />
          <div v-else class="table-wrap">
            <table class="data-table">
              <thead>
                <tr>
                  <th>{{ t('通知', 'Notification') }}</th>
                  <th>{{ t('渠道', 'Channel') }}</th>
                  <th>{{ t('状态 / 尝试', 'Status / attempts') }}</th>
                  <th>{{ t('创建时间', 'Created') }}</th>
                  <th>{{ t('失败原因', 'Failure reason') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="delivery in deliveries.data.value.items" :key="delivery.id">
                  <td>
                    <span class="monitor-name">{{ statusLabel(delivery.kind) }}</span
                    ><RouterLink
                      v-if="delivery.monitorId"
                      :to="`/app/monitors/${delivery.monitorId}`"
                      class="monitor-sub"
                      >{{
                        monitors.data.value?.items.find((x) => x.id === delivery.monitorId)?.name ||
                        delivery.monitorId
                      }}</RouterLink
                    >
                  </td>
                  <td>
                    {{
                      query.data.value?.items.find((x) => x.id === delivery.channelId)?.name ||
                      delivery.channelId
                    }}
                  </td>
                  <td>
                    <span class="pill">{{ statusLabel(delivery.status) }}</span
                    ><span class="muted" un-ml="2">{{ delivery.attempts }}</span>
                  </td>
                  <td class="muted" un-text="10px">{{ formatDate(delivery.createdAt) }}</td>
                  <td class="muted" un-text="10px">{{ delivery.lastError || '—' }}</td>
                </tr>
              </tbody>
            </table>
          </div></AsyncState
        ></Tabs.Content
      ></Tabs.Root
    >
  </section>
  <p class="note" un-mt="5">
    {{
      t(
        '失败投递按有限退避重试。已失效的故障消息会跳过；极端崩溃情形下可能重复发送。维护期间抑制服务故障与恢复通知。',
        'Failed jobs use bounded backoff. Stale outage messages are skipped; an extreme crash can produce a duplicate. Maintenance suppresses outage and recovery notifications.',
      )
    }}
  </p>
  <Modal
    v-model:open="open"
    :title="form.id ? t('编辑通知渠道', 'Edit channel') : t('添加通知渠道', 'Add channel')"
    ><form id="channel-form" @submit.prevent="save">
      <Field :label="t('显示名称', 'Display name')"><input v-model="form.name" required /></Field
      ><Field :label="t('Shoutrrr 服务 URL 秘密', 'Shoutrrr service URL secret')" un-mt="5"
        ><SecretSelect v-model="form.serviceUrlSecretId" :secrets="secrets.data.value?.items || []"
      /></Field>
      <p class="field-hint" un-mt="2">
        {{
          t(
            '如 smtp://…、telegram://… 等，完整 URL 存储为秘密，不会回读。',
            'For smtp://…, telegram://…, and other service URLs. The full URL stays in a secret and cannot be read back.',
          )
        }}
      </p>
      <Toggle v-model="form.enabled" :label="t('启用渠道', 'Enable channel')" un-mt="5" />
      <p v-if="error" class="inline-error">{{ error }}</p>
    </form>
    <template #footer
      ><button class="button" @click="open = false">{{ t('取消', 'Cancel') }}</button
      ><button class="button primary" form="channel-form" :disabled="saving">
        {{ t('保存', 'Save') }}
      </button></template
    ></Modal
  ><Modal v-model:open="deleteOpen" :title="t('删除通知渠道', 'Delete channel')"
    ><p>{{ deleteTarget?.name }}</p>
    <template #footer
      ><button class="button" @click="deleteOpen = false">{{ t('取消', 'Cancel') }}</button
      ><button class="button danger" @click="remove">{{ t('删除', 'Delete') }}</button></template
    ></Modal
  >
</template>
