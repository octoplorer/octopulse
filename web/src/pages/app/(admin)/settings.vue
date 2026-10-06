<script setup lang="ts">
import type { Settings as OrganizationSettings } from '../../../client/types.gen'
import { useMutation, useQueryCache } from '@pinia/colada'
import { useForm } from '@tanstack/vue-form'
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  getSettingsQuery,
  updateProfileMutation,
  updateSettingsMutation,
} from '../../../client/@pinia/colada.gen'
import AsyncState from '../../../components/AsyncState.vue'
import Field from '../../../components/Field.vue'
import ListInput from '../../../components/ListInput.vue'
import PageHeader from '../../../components/PageHeader.vue'
import { Alert } from '../../../components/ui/alert'
import { Button } from '../../../components/ui/button'
import { Card } from '../../../components/ui/card'
import { FieldActions, FieldGroup, FieldInput, FieldSection } from '../../../components/ui/field'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '../../../components/ui/select'
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
const error = ref('')
const tab = ref('personal')
const organizationForm = useForm({
  defaultValues: {
    organizationName: '',
    timezone: 'UTC',
    locale: 'zh-CN',
    retention: { roundDays: 14, attemptDays: 3, fiveMinuteDays: 90, historyMonths: 13 },
    allowedDomains: [],
  } as OrganizationSettings,
  onSubmit: async ({ value }) => {
    error.value = ''
    try {
      const result = await updateSettings.mutateAsync({
        body: value,
      })
      notify(t('settings.organizationSettingsSaved'))
      organizationForm.reset(result)
    }
    catch (e) {
      error.value = errorText(e)
    }
  },
})
const organizationSaving = organizationForm.useSelector(state => state.isSubmitting)
const profileForm = useForm({
  defaultValues: {
    name: currentUser.value?.name || '',
    locale: currentUser.value?.locale || 'zh-CN',
    timezone: currentUser.value?.timezone || timezone.value,
    oldPassword: '',
    password: '',
  },
  onSubmit: async ({ value }) => {
    error.value = ''
    try {
      const user = await updateProfile.mutateAsync({
        body: {
          name: value.name,
          locale: value.locale,
          timezone: value.timezone,
          ...(value.password ? { password: value.password, oldPassword: value.oldPassword } : {}),
        },
      })
      currentUser.value = user
      locale.value = user.locale
      timezone.value = user.timezone
      notify(t('settings.personalSettingsSaved'))
      profileForm.reset({
        name: user.name,
        locale: user.locale,
        timezone: user.timezone,
        oldPassword: '',
        password: '',
      })
    }
    catch (e) {
      error.value = errorText(e)
    }
  },
})
const profileSaving = profileForm.useSelector(state => state.isSubmitting)
const newPassword = profileForm.useSelector(state => state.values.password)
onMounted(async () => {
  try {
    const state = await queryCache.refresh(
      queryCache.ensure({ ...getSettingsQuery(), staleTime: 0 }),
    )
    if (state.status !== 'success')
      throw state.error || new Error(t('errors.requestFailed'))
    organizationForm.reset(structuredClone(state.data))
  }
  catch (e) {
    error.value = errorText(e)
  }
  finally {
    loading.value = false
  }
})
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
          <form @submit.prevent="profileForm.handleSubmit">
            <FieldSection>
              <h2>{{ t('settings.displayPreferences') }}</h2>
              <p class="muted" un-text="13px subtle">
                {{ t('settings.yourDisplayTimeZoneDoesNotChangeUtc') }}
              </p>
              <FieldGroup>
                <profileForm.Field v-slot="{ field }" name="name">
                  <Field :label="t('common.displayName')">
                    <FieldInput :model-value="field.state.value" @update:model-value="field.handleChange($event)" @blur="field.handleBlur" />
                  </Field>
                </profileForm.Field><profileForm.Field v-slot="{ field }" name="locale">
                  <Field :label="t('common.language')">
                    <Select :model-value="field.state.value" @update:model-value="field.handleChange($event)" @focusout="field.handleBlur">
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
                </profileForm.Field><profileForm.Field v-slot="{ field }" name="timezone">
                  <Field :label="t('common.displayTimeZone')">
                    <FieldInput :model-value="field.state.value" placeholder="Asia/Shanghai" required @update:model-value="field.handleChange($event)" @blur="field.handleBlur" />
                  </Field>
                </profileForm.Field><Field :label="t('settings.adminColorScheme')">
                  <Select v-model="theme">
                    <SelectTrigger><SelectValue /></SelectTrigger>
                    <SelectContent>
                      <SelectGroup>
                        <SelectItem value="system">
                          {{ t('common.system') }}
                        </SelectItem>
                        <SelectItem value="light">
                          {{ t('common.light') }}
                        </SelectItem>
                        <SelectItem value="dark">
                          {{ t('common.dark') }}
                        </SelectItem>
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                </Field>
              </FieldGroup>
            </FieldSection>
            <FieldSection>
              <h2>{{ t('settings.changePassword') }}</h2>
              <p class="muted" un-text="13px subtle">
                {{ t('settings.leaveEmptyToKeepYourPassword') }}
              </p>
              <FieldGroup>
                <profileForm.Field v-slot="{ field }" name="oldPassword">
                  <Field :label="t('settings.currentPassword')">
                    <FieldInput :model-value="field.state.value" type="password" autocomplete="current-password" :required="!!newPassword" @update:model-value="field.handleChange($event)" @blur="field.handleBlur" />
                  </Field>
                </profileForm.Field><profileForm.Field v-slot="{ field }" name="password">
                  <Field :label="t('settings.newPassword')" :hint="t('common.atLeast12CharactersUpTo72Bytes')">
                    <FieldInput :model-value="field.state.value" type="password" autocomplete="new-password" minlength="12" maxlength="72" @update:model-value="field.handleChange($event)" @blur="field.handleBlur" />
                  </Field>
                </profileForm.Field>
              </FieldGroup>
              <FieldActions>
                <Button :disabled="profileSaving" variant="primary">
                  <span w="14px" h="14px" aria-hidden="true" class="i-lucide-save" />{{ t('settings.savePreferences') }}
                </Button>
              </FieldActions>
            </FieldSection>
          </form>
        </TabsContent><TabsContent value="organization">
          <form @submit.prevent="organizationForm.handleSubmit">
            <FieldSection>
              <h2>{{ t('settings.organization') }}</h2>
              <p class="muted" un-text="13px subtle">
                {{ t('settings.organizationSettingsProvideDefaultsForAccountsAndMaintenance') }}
              </p>
              <FieldGroup>
                <organizationForm.Field v-slot="{ field }" name="organizationName">
                  <Field :label="t('common.organizationName')">
                    <FieldInput :model-value="field.state.value" :disabled="!isAdmin()" required @update:model-value="field.handleChange($event)" @blur="field.handleBlur" />
                  </Field>
                </organizationForm.Field><organizationForm.Field v-slot="{ field }" name="timezone">
                  <Field :label="t('common.organizationTimeZone')">
                    <FieldInput :model-value="field.state.value" :disabled="!isAdmin()" required @update:model-value="field.handleChange($event)" @blur="field.handleBlur" />
                  </Field>
                </organizationForm.Field><organizationForm.Field v-slot="{ field }" name="locale">
                  <Field :label="t('settings.defaultLanguage')">
                    <Select :model-value="field.state.value" :disabled="!isAdmin()" @update:model-value="field.handleChange($event)" @focusout="field.handleBlur">
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
                </organizationForm.Field><organizationForm.Field v-slot="{ field }" name="allowedDomains">
                  <Field :label="t('settings.allowedStatusPageDomainsOnePerLine')" :hint="t('settings.hostnamesOnlyWithoutSchemeOrPathConfigureDns')" class="span-full">
                    <ListInput :model-value="field.state.value" :parse-item="String" separator="lines" trim multiline :disabled="!isAdmin()" placeholder="status.example.com" @update:model-value="field.handleChange($event)" @blur="field.handleBlur" />
                  </Field>
                </organizationForm.Field>
              </FieldGroup>
              <FieldActions v-if="isAdmin()">
                <Button :disabled="organizationSaving" variant="primary">
                  <span w="14px" h="14px" aria-hidden="true" class="i-lucide-save" />{{ t('settings.saveOrganization') }}
                </Button>
              </FieldActions>
            </FieldSection>
          </form>
        </TabsContent><TabsContent v-if="isAdmin()" value="retention">
          <form @submit.prevent="organizationForm.handleSubmit">
            <FieldSection>
              <h2>{{ t('settings.historyRetention') }}</h2>
              <p class="muted" un-text="13px subtle">
                {{ t('settings.configureRawAndAggregatedHistoryWindowsCleanupRuns') }}
              </p>
              <FieldGroup>
                <organizationForm.Field v-slot="{ field }" name="retention.roundDays">
                  <Field :label="t('settings.rawRoundsDays')">
                    <FieldInput :model-value="field.state.value" type="number" min="1" required @update:model-value="field.handleChange(Number($event))" @blur="field.handleBlur" />
                  </Field>
                </organizationForm.Field><organizationForm.Field v-slot="{ field }" name="retention.attemptDays">
                  <Field :label="t('settings.attemptDetailsDays')">
                    <FieldInput :model-value="field.state.value" type="number" min="1" required @update:model-value="field.handleChange(Number($event))" @blur="field.handleBlur" />
                  </Field>
                </organizationForm.Field><organizationForm.Field v-slot="{ field }" name="retention.fiveMinuteDays">
                  <Field :label="t('settings.5MinuteAggregatesDays')">
                    <FieldInput :model-value="field.state.value" type="number" min="1" required @update:model-value="field.handleChange(Number($event))" @blur="field.handleBlur" />
                  </Field>
                </organizationForm.Field><organizationForm.Field v-slot="{ field }" name="retention.historyMonths">
                  <Field :label="t('settings.hourlyAggregatesStateIntervalsMonths')">
                    <FieldInput :model-value="field.state.value" type="number" min="1" required @update:model-value="field.handleChange(Number($event))" @blur="field.handleBlur" />
                  </Field>
                </organizationForm.Field>
              </FieldGroup>
              <Alert mt="6" as="p" variant="default">
                {{ t('settings.completeRequestResponseBodiesAndSecretHeadersAre') }}
              </Alert>
              <FieldActions>
                <Button :disabled="organizationSaving" variant="primary">
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
