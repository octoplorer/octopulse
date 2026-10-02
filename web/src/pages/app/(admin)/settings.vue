<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { reactive, ref, onMounted } from 'vue'
import { Save } from '@lucide/vue'
import { Tabs } from '@ark-ui/vue/tabs'
import { useMutation, useQueryCache } from '@pinia/colada'
import { isAdmin, currentUser } from '../../../lib/api'
import {
  getSettingsQuery,
  updateSettingsMutation,
  updateProfileMutation,
} from '../../../client/@pinia/colada.gen'
import type { Settings as OrganizationSettings } from '../../../lib/types'
import { languageOptions } from '../../../lib/i18n'
import { theme, timezone } from '../../../lib/preferences'
import { notify, errorText } from '../../../lib/notices'
import PageHeader from '../../../components/PageHeader.vue'
import Field from '../../../components/Field.vue'
import AsyncState from '../../../components/AsyncState.vue'
const { t, locale } = useI18n({ useScope: 'global' })

definePage({ meta: { title: 'common.settings' } })

const updateSettings = useMutation(updateSettingsMutation())
const updateProfile = useMutation(updateProfileMutation())
const queryCache = useQueryCache()
const loading = ref(true),
  saving = ref(false),
  error = ref(''),
  tab = ref('personal'),
  domains = ref(''),
  form = reactive<OrganizationSettings>({
    organizationName: '',
    timezone: 'UTC',
    locale: 'zh-CN',
    retention: { roundDays: 14, attemptDays: 3, fiveMinuteDays: 90, historyMonths: 13 },
    allowedDomains: [],
  }),
  profile = reactive({
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
    if (state.status !== 'success') throw state.error || new Error(t('errors.requestFailed'))
    Object.assign(form, structuredClone(state.data))
    domains.value = form.allowedDomains.join('\n')
  } catch (e) {
    error.value = errorText(e)
  } finally {
    loading.value = false
  }
})
async function saveOrganization() {
  saving.value = true
  error.value = ''
  try {
    form.allowedDomains = domains.value
      .split('\n')
      .map((x) => x.trim())
      .filter(Boolean)
    Object.assign(form, structuredClone(await updateSettings.mutateAsync({ body: form })))
    notify(t('settings.organizationSettingsSaved'))
  } catch (e) {
    error.value = errorText(e)
  } finally {
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
  } catch (e) {
    error.value = errorText(e)
  } finally {
    saving.value = false
  }
}
</script>
<template>
  <PageHeader
    :title="t('common.settings')"
    :description="t('settings.configureTheWorkspaceForYourOrganizationAndYour')"
  /><AsyncState :pending="loading"
    ><div v-if="error" class="validation-error" role="alert">{{ error }}</div>
    <section class="card" un-max-w="4xl">
      <Tabs.Root v-model="tab"
        ><Tabs.List class="tabs-list"
          ><Tabs.Trigger class="tabs-trigger" value="personal">{{
            t('settings.personalPreferences')
          }}</Tabs.Trigger
          ><Tabs.Trigger class="tabs-trigger" value="organization">{{
            t('settings.organizationSettings')
          }}</Tabs.Trigger
          ><Tabs.Trigger v-if="isAdmin()" class="tabs-trigger" value="retention">{{
            t('settings.dataRetention')
          }}</Tabs.Trigger></Tabs.List
        ><Tabs.Content value="personal"
          ><form @submit.prevent="saveProfile">
            <div class="form-section">
              <h2>{{ t('settings.displayPreferences') }}</h2>
              <p class="muted">
                {{ t('settings.yourDisplayTimeZoneDoesNotChangeUtc') }}
              </p>
              <div class="form-grid">
                <Field :label="t('common.displayName')"><input v-model="profile.name" /></Field
                ><Field :label="t('common.language')"
                  ><select v-model="profile.locale">
                    <option
                      v-for="language in languageOptions"
                      :key="language.value"
                      :value="language.value"
                    >
                      {{ language.label }}
                    </option>
                  </select></Field
                ><Field :label="t('common.displayTimeZone')"
                  ><input v-model="profile.timezone" placeholder="Asia/Shanghai" required /></Field
                ><Field :label="t('settings.adminColorScheme')"
                  ><select v-model="theme">
                    <option value="system">{{ t('common.system') }}</option>
                    <option value="light">{{ t('common.light') }}</option>
                    <option value="dark">{{ t('common.dark') }}</option>
                  </select></Field
                >
              </div>
            </div>
            <div class="form-section">
              <h2>{{ t('settings.changePassword') }}</h2>
              <p class="muted">
                {{ t('settings.leaveEmptyToKeepYourPassword') }}
              </p>
              <div class="form-grid">
                <Field :label="t('settings.currentPassword')"
                  ><input
                    v-model="profile.oldPassword"
                    type="password"
                    autocomplete="current-password"
                    :required="!!profile.password" /></Field
                ><Field
                  :label="t('settings.newPassword')"
                  :hint="t('common.atLeast12CharactersUpTo72Bytes')"
                  ><input
                    v-model="profile.password"
                    type="password"
                    autocomplete="new-password"
                    minlength="12"
                    maxlength="72"
                /></Field>
              </div>
              <div class="form-actions">
                <button class="button primary" :disabled="saving">
                  <Save :size="14" />{{ t('settings.savePreferences') }}
                </button>
              </div>
            </div>
          </form></Tabs.Content
        ><Tabs.Content value="organization"
          ><form @submit.prevent="saveOrganization">
            <div class="form-section">
              <h2>{{ t('settings.organization') }}</h2>
              <p class="muted">
                {{ t('settings.organizationSettingsProvideDefaultsForAccountsAndMaintenance') }}
              </p>
              <div class="form-grid">
                <Field :label="t('common.organizationName')"
                  ><input v-model="form.organizationName" :disabled="!isAdmin()" required /></Field
                ><Field :label="t('common.organizationTimeZone')"
                  ><input v-model="form.timezone" :disabled="!isAdmin()" required /></Field
                ><Field :label="t('settings.defaultLanguage')"
                  ><select v-model="form.locale" :disabled="!isAdmin()">
                    <option
                      v-for="language in languageOptions"
                      :key="language.value"
                      :value="language.value"
                    >
                      {{ language.label }}
                    </option>
                  </select></Field
                ><Field
                  class="span-full"
                  :label="t('settings.allowedStatusPageDomainsOnePerLine')"
                  :hint="t('settings.hostnamesOnlyWithoutSchemeOrPathConfigureDns')"
                >
                  <textarea
                    v-model="domains"
                    :disabled="!isAdmin()"
                    placeholder="status.example.com"
                  />
                </Field>
              </div>
              <div v-if="isAdmin()" class="form-actions">
                <button class="button primary" :disabled="saving">
                  <Save :size="14" />{{ t('settings.saveOrganization') }}
                </button>
              </div>
            </div>
          </form></Tabs.Content
        ><Tabs.Content v-if="isAdmin()" value="retention"
          ><form @submit.prevent="saveOrganization">
            <div class="form-section">
              <h2>{{ t('settings.historyRetention') }}</h2>
              <p class="muted">
                {{ t('settings.configureRawAndAggregatedHistoryWindowsCleanupRuns') }}
              </p>
              <div class="form-grid">
                <Field :label="t('settings.rawRoundsDays')"
                  ><input
                    v-model.number="form.retention.roundDays"
                    type="number"
                    min="1"
                    required /></Field
                ><Field :label="t('settings.attemptDetailsDays')"
                  ><input
                    v-model.number="form.retention.attemptDays"
                    type="number"
                    min="1"
                    required /></Field
                ><Field :label="t('settings.5MinuteAggregatesDays')"
                  ><input
                    v-model.number="form.retention.fiveMinuteDays"
                    type="number"
                    min="1"
                    required /></Field
                ><Field :label="t('settings.hourlyAggregatesStateIntervalsMonths')"
                  ><input
                    v-model.number="form.retention.historyMonths"
                    type="number"
                    min="1"
                    required
                /></Field>
              </div>
              <p class="note" un-mt="6">
                {{ t('settings.completeRequestResponseBodiesAndSecretHeadersAre') }}
              </p>
              <div class="form-actions">
                <button class="button primary" :disabled="saving">
                  <Save :size="14" />{{ t('settings.saveRetention') }}
                </button>
              </div>
            </div>
          </form></Tabs.Content
        ></Tabs.Root
      >
    </section></AsyncState
  >
</template>
