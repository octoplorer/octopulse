<script setup lang="ts">
import type { Settings as OrganizationSettings } from '../../../client/types.gen'
import { useMutation, useQueryCache } from '@pinia/colada'
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  getSettingsQuery,
  updateProfileMutation,
  updateSettingsMutation,
} from '../../../client/@pinia/colada.gen'
import AsyncState from '../../../components/AsyncState.vue'
import Field from '../../../components/Field.vue'
import PageHeader from '../../../components/PageHeader.vue'
import { Alert } from '../../../components/ui/alert'
import { Button } from '../../../components/ui/button'
import { Card } from '../../../components/ui/card'
import { FieldActions, FieldGroup, FieldSection } from '../../../components/ui/field'
import { TabsContent, TabsList, TabsRoot, TabsTrigger } from '../../../components/ui/tabs'
import { currentUser, isAdmin } from '../../../composables/api'
import { languageOptions } from '../../../composables/i18n'
import { notify } from '../../../composables/notices'
import { theme, timezone } from '../../../composables/preferences'
import { errorText } from '../../../lib/errors'

const { t, locale } = useI18n({ useScope: 'global' })

definePage({ meta: { title: 'common.settings' } })

const updateSettings = useMutation(updateSettingsMutation())
const updateProfile = useMutation(updateProfileMutation())
const queryCache = useQueryCache()
const loading = ref(true)
const saving = ref(false)
const error = ref('')
const tab = ref('personal')
const domains = ref('')
const form = reactive<OrganizationSettings>({
  organizationName: '',
  timezone: 'UTC',
  locale: 'zh-CN',
  retention: { roundDays: 14, attemptDays: 3, fiveMinuteDays: 90, historyMonths: 13 },
  allowedDomains: [],
})
const profile = reactive({
  name: currentUser.value?.name || '',
  locale: currentUser.value?.locale || 'zh-CN',
  timezone: currentUser.value?.timezone || timezone.value,
  oldPassword: '',
  password: '',
})
onMounted(async () => {
  try {
    const state = await queryCache.refresh(
      queryCache.ensure({ ...getSettingsQuery(), staleTime: 0 }),
    )
    if (state.status !== 'success')
      throw state.error || new Error(t('errors.requestFailed'))
    Object.assign(form, structuredClone(state.data))
    domains.value = form.allowedDomains.join('\n')
  }
  catch (e) {
    error.value = errorText(e)
  }
  finally {
    loading.value = false
  }
})
async function saveOrganization() {
  saving.value = true
  error.value = ''
  try {
    form.allowedDomains = domains.value
      .split('\n')
      .map(x => x.trim())
      .filter(Boolean)
    Object.assign(form, structuredClone(await updateSettings.mutateAsync({ body: form })))
    notify(t('settings.organizationSettingsSaved'))
  }
  catch (e) {
    error.value = errorText(e)
  }
  finally {
    saving.value = false
  }
}
async function saveProfile() {
  saving.value = true
  error.value = ''
  try {
    currentUser.value = await updateProfile.mutateAsync({
      body: {
        name: profile.name,
        locale: profile.locale,
        timezone: profile.timezone,
        ...(profile.password
          ? { password: profile.password, oldPassword: profile.oldPassword }
          : {}),
      },
    })
    locale.value = profile.locale
    timezone.value = profile.timezone
    profile.password = ''
    profile.oldPassword = ''
    notify(t('settings.personalSettingsSaved'))
  }
  catch (e) {
    error.value = errorText(e)
  }
  finally {
    saving.value = false
  }
}
</script>

<template>
  <PageHeader :title="t('common.settings')" :description="t('settings.configureTheWorkspaceForYourOrganizationAndYour')" /><AsyncState :pending="loading">
    <Alert v-if="error" role="alert" variant="validation">
      {{ error }}
    </Alert>
    <Card max-w="4xl" as="section">
      <TabsRoot v-model="tab">
        <TabsList>
          <TabsTrigger value="personal">
            {{
              t('settings.personalPreferences')
            }}
          </TabsTrigger><TabsTrigger value="organization">
            {{
              t('settings.organizationSettings')
            }}
          </TabsTrigger><TabsTrigger v-if="isAdmin()" value="retention">
            {{
              t('settings.dataRetention')
            }}
          </TabsTrigger>
        </TabsList><TabsContent value="personal">
          <form @submit.prevent="saveProfile">
            <FieldSection>
              <h2>{{ t('settings.displayPreferences') }}</h2>
              <p class="muted" un-text="13px $muted">
                {{ t('settings.yourDisplayTimeZoneDoesNotChangeUtc') }}
              </p>
              <FieldGroup>
                <Field :label="t('common.displayName')">
                  <input v-model="profile.name">
                </Field><Field :label="t('common.language')">
                  <select v-model="profile.locale">
                    <option v-for="language in languageOptions" :key="language.value" :value="language.value">
                      {{ language.label }}
                    </option>
                  </select>
                </Field><Field :label="t('common.displayTimeZone')">
                  <input v-model="profile.timezone" placeholder="Asia/Shanghai" required>
                </Field><Field :label="t('settings.adminColorScheme')">
                  <select v-model="theme">
                    <option value="system">
                      {{ t('common.system') }}
                    </option>
                    <option value="light">
                      {{ t('common.light') }}
                    </option>
                    <option value="dark">
                      {{ t('common.dark') }}
                    </option>
                  </select>
                </Field>
              </FieldGroup>
            </FieldSection>
            <FieldSection>
              <h2>{{ t('settings.changePassword') }}</h2>
              <p class="muted" un-text="13px $muted">
                {{ t('settings.leaveEmptyToKeepYourPassword') }}
              </p>
              <FieldGroup>
                <Field :label="t('settings.currentPassword')">
                  <input v-model="profile.oldPassword" type="password" autocomplete="current-password" :required="!!profile.password">
                </Field><Field :label="t('settings.newPassword')" :hint="t('common.atLeast12CharactersUpTo72Bytes')">
                  <input v-model="profile.password" type="password" autocomplete="new-password" minlength="12" maxlength="72">
                </Field>
              </FieldGroup>
              <FieldActions>
                <Button :disabled="saving" variant="primary">
                  <span w="14px" h="14px" aria-hidden="true" class="i-lucide-save" />{{ t('settings.savePreferences') }}
                </Button>
              </FieldActions>
            </FieldSection>
          </form>
        </TabsContent><TabsContent value="organization">
          <form @submit.prevent="saveOrganization">
            <FieldSection>
              <h2>{{ t('settings.organization') }}</h2>
              <p class="muted" un-text="13px $muted">
                {{ t('settings.organizationSettingsProvideDefaultsForAccountsAndMaintenance') }}
              </p>
              <FieldGroup>
                <Field :label="t('common.organizationName')">
                  <input v-model="form.organizationName" :disabled="!isAdmin()" required>
                </Field><Field :label="t('common.organizationTimeZone')">
                  <input v-model="form.timezone" :disabled="!isAdmin()" required>
                </Field><Field :label="t('settings.defaultLanguage')">
                  <select v-model="form.locale" :disabled="!isAdmin()">
                    <option v-for="language in languageOptions" :key="language.value" :value="language.value">
                      {{ language.label }}
                    </option>
                  </select>
                </Field><Field :label="t('settings.allowedStatusPageDomainsOnePerLine')" :hint="t('settings.hostnamesOnlyWithoutSchemeOrPathConfigureDns')" class="span-full">
                  <textarea v-model="domains" :disabled="!isAdmin()" placeholder="status.example.com" />
                </Field>
              </FieldGroup>
              <FieldActions v-if="isAdmin()">
                <Button :disabled="saving" variant="primary">
                  <span w="14px" h="14px" aria-hidden="true" class="i-lucide-save" />{{ t('settings.saveOrganization') }}
                </Button>
              </FieldActions>
            </FieldSection>
          </form>
        </TabsContent><TabsContent v-if="isAdmin()" value="retention">
          <form @submit.prevent="saveOrganization">
            <FieldSection>
              <h2>{{ t('settings.historyRetention') }}</h2>
              <p class="muted" un-text="13px $muted">
                {{ t('settings.configureRawAndAggregatedHistoryWindowsCleanupRuns') }}
              </p>
              <FieldGroup>
                <Field :label="t('settings.rawRoundsDays')">
                  <input v-model.number="form.retention.roundDays" type="number" min="1" required>
                </Field><Field :label="t('settings.attemptDetailsDays')">
                  <input v-model.number="form.retention.attemptDays" type="number" min="1" required>
                </Field><Field :label="t('settings.5MinuteAggregatesDays')">
                  <input v-model.number="form.retention.fiveMinuteDays" type="number" min="1" required>
                </Field><Field :label="t('settings.hourlyAggregatesStateIntervalsMonths')">
                  <input v-model.number="form.retention.historyMonths" type="number" min="1" required>
                </Field>
              </FieldGroup>
              <Alert mt="6" as="p" variant="default">
                {{ t('settings.completeRequestResponseBodiesAndSecretHeadersAre') }}
              </Alert>
              <FieldActions>
                <Button :disabled="saving" variant="primary">
                  <span w="14px" h="14px" aria-hidden="true" class="i-lucide-save" />{{ t('settings.saveRetention') }}
                </Button>
              </FieldActions>
            </FieldSection>
          </form>
        </TabsContent>
      </TabsRoot>
    </Card>
  </AsyncState>
</template>
