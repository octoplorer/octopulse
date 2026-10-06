<script setup lang="ts">
import type { Channel } from '../../../client/types.gen'
import { useMutation, useQuery } from '@pinia/colada'
import { reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  createChannelsMutation,
  deleteChannelsMutation,
  listChannelsQuery,
  listDeliveriesQuery,
  listMonitorsQuery,
  listSecretsQuery,
  testChannelMutation,
  updateChannelsMutation,
} from '../../../client/@pinia/colada.gen'
import AsyncState from '../../../components/AsyncState.vue'
import EmptyState from '../../../components/EmptyState.vue'
import Field from '../../../components/Field.vue'
import Modal from '../../../components/Modal.vue'
import PageHeader from '../../../components/PageHeader.vue'
import SecretSelect from '../../../components/SecretSelect.vue'
import Toggle from '../../../components/Toggle.vue'
import { Alert } from '../../../components/ui/alert'
import { Badge } from '../../../components/ui/badge'
import { Button } from '../../../components/ui/button'
import { Card } from '../../../components/ui/card'
import { FieldDescription, FieldError } from '../../../components/ui/field'
import { Table, TableBody, TableCell, TableContainer, TableHead, TableHeader, TableRow } from '../../../components/ui/table'
import { TabsContent, TabsList, TabsRoot, TabsTrigger } from '../../../components/ui/tabs'
import { isAdmin } from '../../../composables/api'
import { notify } from '../../../composables/notices'
import { formatDate, statusLabel } from '../../../composables/preferences'
import { errorText } from '../../../lib/errors'
import { clone } from '../../../lib/form'

const { t } = useI18n({ useScope: 'global' })

definePage({ meta: { title: 'navigation.notifications' } })

const createChannel = useMutation(createChannelsMutation())
const updateChannel = useMutation(updateChannelsMutation())
const deleteChannel = useMutation(deleteChannelsMutation())
const testChannel = useMutation(testChannelMutation())

const query = useQuery({ ...listChannelsQuery(), staleTime: 10000 })
const secrets = useQuery({ ...listSecretsQuery(), staleTime: 10000 })
const deliveries = useQuery({ ...listDeliveriesQuery(), staleTime: 10000 })
const monitors = useQuery({ ...listMonitorsQuery(), staleTime: 10000 })
const open = ref(false)
const saving = ref(false)
const error = ref('')
const tab = ref('channels')
const testing = ref('')
const deleteTarget = ref<Channel | null>(null)
const deleteOpen = ref(false)
function empty(): Channel {
  return {
    id: '',
    name: '',
    serviceUrlSecretId: '',
    enabled: true,
    createdAt: 0,
    updatedAt: 0,
  }
}
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
    if (form.id)
      await updateChannel.mutateAsync({ path: { id: form.id }, body: form })
    else await createChannel.mutateAsync({ body: form })
    open.value = false
    await query.refresh()
    notify(t('notifications.channelSaved'))
  }
  catch (e) {
    error.value = errorText(e)
  }
  finally {
    saving.value = false
  }
}
async function test(channel: Channel) {
  testing.value = channel.id
  try {
    await testChannel.mutateAsync({ path: { id: channel.id } })
    notify(t('notifications.testSubmittedCheckDeliveryHistory'))
  }
  catch (e) {
    notify(errorText(e), 'error')
  }
  finally {
    await deliveries.refresh()
    testing.value = ''
  }
}
async function remove() {
  if (!deleteTarget.value)
    return
  try {
    await deleteChannel.mutateAsync({ path: { id: deleteTarget.value.id } })
    deleteOpen.value = false
    await query.refresh()
    notify(t('notifications.channelDeleted'))
  }
  catch (e) {
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
  <PageHeader :title="t('navigation.notifications')" :description="t('notifications.connectChannelsThroughShoutrrrAndKeepEveryDelivery')">
    <Button @click="refresh">
      <span w="14px" h="14px" aria-hidden="true" class="i-lucide-refresh-cw" />{{ t('common.refresh') }}
    </Button><Button v-if="isAdmin()" variant="primary" @click="edit()">
      <span w="15px" h="15px" aria-hidden="true" class="i-lucide-plus" />{{ t('notifications.addChannel') }}
    </Button>
  </PageHeader>
  <Card as="section">
    <TabsRoot v-model="tab">
      <TabsList>
        <TabsTrigger value="channels">
          {{
            t('notifications.channels')
          }}
        </TabsTrigger><TabsTrigger value="deliveries">
          {{
            t('notifications.deliveries')
          }}
        </TabsTrigger>
      </TabsList><TabsContent value="channels">
        <AsyncState :pending="query.isPending.value" :error="query.error.value" @retry="query.refetch()">
          <EmptyState v-if="!query.data.value?.items.length" :title="t('notifications.connectANotificationChannel')" :description="t('notifications.storeAShoutrrrServiceUrlAsASecret')">
            <Button v-if="isAdmin()" as-child>
              <RouterLink to="/app/secrets">
                {{
                  t('notifications.manageSecrets')
                }}
              </RouterLink>
            </Button>
          </EmptyState>
          <TableContainer v-else>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{{ t('common.channel') }}</TableHead>
                  <TableHead>{{ t('notifications.secretReference') }}</TableHead>
                  <TableHead>{{ t('common.status') }}</TableHead>
                  <TableHead>{{ t('common.updated') }}</TableHead>
                  <TableHead />
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="channel in query.data.value.items" :key="channel.id">
                  <TableCell>
                    <span flex="~ items-center gap-2" class="monitor-name block" font="600" un-text="13px"><span w="15px" h="15px" aria-hidden="true" class="i-lucide-bell" />{{ channel.name }}</span>
                  </TableCell>
                  <TableCell class="muted" un-text="13px subtle">
                    {{
                      secrets.data.value?.items.find((x) => x.id === channel.serviceUrlSecretId)
                        ?.name || '—'
                    }}
                  </TableCell>
                  <TableCell>
                    <Badge>
                      {{
                        channel.enabled ? t('common.enabled') : t('common.disabled')
                      }}
                    </Badge>
                  </TableCell>
                  <TableCell class="muted" un-text="13px subtle">
                    {{ formatDate(channel.updatedAt) }}
                  </TableCell>
                  <TableCell>
                    <div v-if="isAdmin()" flex="~ items-center gap-2">
                      <Button :disabled="testing === channel.id || !channel.enabled" size="sm" @click="test(channel)">
                        <span w="12px" h="12px" aria-hidden="true" class="i-lucide-send" />{{ t('notifications.test') }}
                      </Button><Button :aria-label="t('common.edit')" shape="square" @click="edit(channel)">
                        <span w="14px" h="14px" aria-hidden="true" class="i-lucide-pencil" />
                      </Button><Button :aria-label="t('common.delete')" shape="square" @click="confirmDelete(channel)">
                        <span w="14px" h="14px" aria-hidden="true" class="i-lucide-trash-2" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </TableContainer>
        </AsyncState>
      </TabsContent><TabsContent value="deliveries">
        <AsyncState :pending="deliveries.isPending.value" :error="deliveries.error.value" @retry="deliveries.refetch()">
          <EmptyState v-if="!deliveries.data.value?.items.length" :title="t('notifications.noDeliveriesYet')" :description="t('notifications.outagesRecoveriesAndChannelTestsCreateDurableDelivery')" />
          <TableContainer v-else>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{{ t('notifications.notification') }}</TableHead>
                  <TableHead>{{ t('common.channel') }}</TableHead>
                  <TableHead>{{ t('notifications.statusAttempts') }}</TableHead>
                  <TableHead>{{ t('common.created') }}</TableHead>
                  <TableHead>{{ t('notifications.failureReason') }}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="delivery in deliveries.data.value.items" :key="delivery.id">
                  <TableCell>
                    <span class="monitor-name block" font="600" un-text="13px">{{ statusLabel(delivery.kind) }}</span><RouterLink v-if="delivery.monitorId" :to="`/app/monitors/${delivery.monitorId}`" class="monitor-sub block [overflow-wrap:anywhere]" un-text="12px subtle" mt="3px" max-w="300px">
                      {{
                        monitors.data.value?.items.find((x) => x.id === delivery.monitorId)?.name
                          || delivery.monitorId
                      }}
                    </RouterLink>
                  </TableCell>
                  <TableCell>
                    {{
                      query.data.value?.items.find((x) => x.id === delivery.channelId)?.name
                        || delivery.channelId
                    }}
                  </TableCell>
                  <TableCell>
                    <Badge>{{ statusLabel(delivery.status) }}</Badge><span ml="2" class="muted" un-text="13px subtle">{{ delivery.attempts }}</span>
                  </TableCell>
                  <TableCell class="muted" un-text="13px subtle">
                    {{ formatDate(delivery.createdAt) }}
                  </TableCell>
                  <TableCell class="muted" un-text="13px subtle">
                    {{ delivery.lastError || '—' }}
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </TableContainer>
        </AsyncState>
      </TabsContent>
    </TabsRoot>
  </Card>
  <Alert mt="5" as="p" variant="default">
    {{ t('notifications.failedJobsUseBoundedBackoffStaleOutageMessages') }}
  </Alert>
  <Modal v-model:open="open" :title="form.id ? t('notifications.editChannel') : t('notifications.addChannel2')">
    <form id="channel-form" @submit.prevent="save">
      <Field :label="t('common.displayName')">
        <input v-model="form.name" required>
      </Field><Field :label="t('notifications.shoutrrrServiceUrlSecret')" mt="5">
        <SecretSelect v-model="form.serviceUrlSecretId" :secrets="secrets.data.value?.items || []" />
      </Field>
      <FieldDescription as="p" mt="2">
        {{ t('notifications.forSmtpTelegramAndOtherServiceUrlsThe') }}
      </FieldDescription>
      <Toggle v-model="form.enabled" :label="t('notifications.enableChannel')" mt="5" />
      <FieldError v-if="error" as="p" py="10px" px="0">
        {{ error }}
      </FieldError>
    </form>
    <template #footer>
      <Button @click="open = false">
        {{ t('common.cancel') }}
      </Button><Button form="channel-form" :disabled="saving" variant="primary">
        {{ t('notifications.save') }}
      </Button>
    </template>
  </Modal><Modal v-model:open="deleteOpen" :title="t('notifications.deleteChannel')">
    <p>{{ deleteTarget?.name }}</p>
    <template #footer>
      <Button @click="deleteOpen = false">
        {{ t('common.cancel') }}
      </Button><Button variant="destructive" @click="remove">
        {{ t('common.delete') }}
      </Button>
    </template>
  </Modal>
</template>
