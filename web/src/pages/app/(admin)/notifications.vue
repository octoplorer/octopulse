<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { ref, reactive } from 'vue'
import { Plus, Bell, Send, Pencil, Trash2, RefreshCw } from '@lucide/vue'
import { Tabs } from '@ark-ui/vue/tabs'
import type { Channel, Secret, Delivery, Monitor } from '../../../lib/types'
import { isAdmin } from '../../../composables/api'
import { useQuery, useMutation, type DefineQueryOptions } from '@pinia/colada'
import type { ErrorModel } from '../../../client/types.gen'
import {
  listChannelsQuery,
  listSecretsQuery,
  listDeliveriesQuery,
  listMonitorsQuery,
  createChannelsMutation,
  updateChannelsMutation,
  deleteChannelsMutation,
  testChannelMutation,
} from '../../../client/@pinia/colada.gen'
import { formatDate, statusLabel } from '../../../composables/preferences'
import { errorText } from '../../../lib/errors'
import { notify } from '../../../composables/notices'
import { clone } from '../../../lib/form'
import PageHeader from '../../../components/PageHeader.vue'
import Field from '../../../components/Field.vue'
import Toggle from '../../../components/Toggle.vue'
import SecretSelect from '../../../components/SecretSelect.vue'
import Modal from '../../../components/Modal.vue'
import AsyncState from '../../../components/AsyncState.vue'
import EmptyState from '../../../components/EmptyState.vue'
const { t } = useI18n({ useScope: 'global' })

definePage({ meta: { title: 'navigation.notifications' } })

const createChannel = useMutation(createChannelsMutation())
const updateChannel = useMutation(updateChannelsMutation())
const deleteChannel = useMutation(deleteChannelsMutation())
const testChannel = useMutation(testChannelMutation())

const query = useQuery({ ...listChannelsQuery(), staleTime: 10000 } as DefineQueryOptions<
  { items: Channel[] },
  ErrorModel
>)
const secrets = useQuery({ ...listSecretsQuery(), staleTime: 10000 } as DefineQueryOptions<
  { items: Secret[] },
  ErrorModel
>)
const deliveries = useQuery({ ...listDeliveriesQuery(), staleTime: 10000 } as DefineQueryOptions<
  { items: Delivery[] },
  ErrorModel
>)
const monitors = useQuery({ ...listMonitorsQuery(), staleTime: 10000 } as DefineQueryOptions<
  { items: Monitor[] },
  ErrorModel
>)
const open = ref(false)
const saving = ref(false)
const error = ref('')
const tab = ref('channels')
const testing = ref('')
const deleteTarget = ref<Channel | null>(null)
const deleteOpen = ref(false)
const empty = (): Channel => ({
  id: '',
  name: '',
  serviceUrlSecretId: '',
  enabled: true,
  createdAt: 0,
  updatedAt: 0,
})
const form = reactive(empty())
function edit(channel?: Channel) {
  Object.assign(form, channel ? clone(channel) : empty())
  error.value = ''
  open.value = true
}
async function save() {
  saving.value = true
  error.value = ''
  try {
    if (form.id) await updateChannel.mutateAsync({ path: { id: form.id }, body: form })
    else await createChannel.mutateAsync({ body: form })
    open.value = false
    await query.refresh()
    notify(t('notifications.channelSaved'))
  } catch (e) {
    error.value = errorText(e)
  } finally {
    saving.value = false
  }
}
async function test(channel: Channel) {
  testing.value = channel.id
  try {
    await testChannel.mutateAsync({ path: { id: channel.id } })
    notify(t('notifications.testSubmittedCheckDeliveryHistory'))
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
    await deleteChannel.mutateAsync({ path: { id: deleteTarget.value.id } })
    deleteOpen.value = false
    await query.refresh()
    notify(t('notifications.channelDeleted'))
  } catch (e) {
    notify(errorText(e), 'error')
  }
}
function confirmDelete(value: Channel) {
  deleteTarget.value = value
  deleteOpen.value = true
}
function refresh() {
  query.refetch()
  deliveries.refetch()
}
</script>
<template>
  <PageHeader
    :title="t('navigation.notifications')"
    :description="t('notifications.connectChannelsThroughShoutrrrAndKeepEveryDelivery')"
    ><button class="button" @click="refresh">
      <RefreshCw :size="14" />{{ t('common.refresh') }}</button
    ><button v-if="isAdmin()" class="button primary" @click="edit()">
      <Plus :size="15" />{{ t('notifications.addChannel') }}
    </button></PageHeader
  >
  <section class="card">
    <Tabs.Root v-model="tab"
      ><Tabs.List class="tabs-list"
        ><Tabs.Trigger class="tabs-trigger" value="channels">{{
          t('notifications.channels')
        }}</Tabs.Trigger
        ><Tabs.Trigger class="tabs-trigger" value="deliveries">{{
          t('notifications.deliveries')
        }}</Tabs.Trigger></Tabs.List
      ><Tabs.Content value="channels"
        ><AsyncState
          :pending="query.isPending.value"
          :error="query.error.value"
          @retry="query.refetch()"
          ><EmptyState
            v-if="!query.data.value?.items.length"
            :title="t('notifications.connectANotificationChannel')"
            :description="t('notifications.storeAShoutrrrServiceUrlAsASecret')"
            ><RouterLink v-if="isAdmin()" to="/app/secrets" class="button">{{
              t('notifications.manageSecrets')
            }}</RouterLink></EmptyState
          >
          <div v-else class="table-wrap">
            <table class="data-table">
              <thead>
                <tr>
                  <th>{{ t('common.channel') }}</th>
                  <th>{{ t('notifications.secretReference') }}</th>
                  <th>{{ t('common.status') }}</th>
                  <th>{{ t('common.updated') }}</th>
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
                      channel.enabled ? t('common.enabled') : t('common.disabled')
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
                        <Send :size="12" />{{ t('notifications.test') }}</button
                      ><button
                        class="icon-button"
                        @click="edit(channel)"
                        :aria-label="t('common.edit')"
                      >
                        <Pencil :size="14" /></button
                      ><button
                        class="icon-button"
                        @click="confirmDelete(channel)"
                        :aria-label="t('common.delete')"
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
          @retry="deliveries.refetch()"
          ><EmptyState
            v-if="!deliveries.data.value?.items.length"
            :title="t('notifications.noDeliveriesYet')"
            :description="t('notifications.outagesRecoveriesAndChannelTestsCreateDurableDelivery')"
          />
          <div v-else class="table-wrap">
            <table class="data-table">
              <thead>
                <tr>
                  <th>{{ t('notifications.notification') }}</th>
                  <th>{{ t('common.channel') }}</th>
                  <th>{{ t('notifications.statusAttempts') }}</th>
                  <th>{{ t('common.created') }}</th>
                  <th>{{ t('notifications.failureReason') }}</th>
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
    {{ t('notifications.failedJobsUseBoundedBackoffStaleOutageMessages') }}
  </p>
  <Modal
    v-model:open="open"
    :title="form.id ? t('notifications.editChannel') : t('notifications.addChannel2')"
    ><form id="channel-form" @submit.prevent="save">
      <Field :label="t('common.displayName')"><input v-model="form.name" required /></Field
      ><Field :label="t('notifications.shoutrrrServiceUrlSecret')" un-mt="5"
        ><SecretSelect v-model="form.serviceUrlSecretId" :secrets="secrets.data.value?.items || []"
      /></Field>
      <p class="field-hint" un-mt="2">
        {{ t('notifications.forSmtpTelegramAndOtherServiceUrlsThe') }}
      </p>
      <Toggle v-model="form.enabled" :label="t('notifications.enableChannel')" un-mt="5" />
      <p v-if="error" class="inline-error">{{ error }}</p>
    </form>
    <template #footer
      ><button class="button" @click="open = false">{{ t('common.cancel') }}</button
      ><button class="button primary" form="channel-form" :disabled="saving">
        {{ t('notifications.save') }}
      </button></template
    ></Modal
  ><Modal v-model:open="deleteOpen" :title="t('notifications.deleteChannel')"
    ><p>{{ deleteTarget?.name }}</p>
    <template #footer
      ><button class="button" @click="deleteOpen = false">{{ t('common.cancel') }}</button
      ><button class="button danger" @click="remove">{{ t('common.delete') }}</button></template
    ></Modal
  >
</template>
