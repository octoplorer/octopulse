<script setup lang="ts">
import { ArrowRight, ShieldCheck } from '@lucide/vue'
import { useMutation, useQueryCache } from '@pinia/colada'
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import {
  createSessionMutation,
  createSetupMutation,
  getSessionQuery,
  getSetupQuery,
} from '../../client/@pinia/colada.gen'
import Brand from '../../components/Brand.vue'
import Field from '../../components/Field.vue'
import { applySession } from '../../composables/api'
import { dark } from '../../composables/preferences'
import { errorText } from '../../lib/errors'

const { t } = useI18n({ useScope: 'global' })
const queryCache = useQueryCache()
const createSession = useMutation({
  ...createSessionMutation(),
  onSuccess: applySession,
})
const setup = useMutation(createSetupMutation())
const route = useRoute()
const router = useRouter()
const required = ref(false)
const loading = ref(true)
const saving = ref(false)
const error = ref('')
const username = ref('')
const password = ref('')
const organizationName = ref('Octopulse')
const timezone = ref(Intl.DateTimeFormat().resolvedOptions().timeZone)
onMounted(async () => {
  try {
    const setupState = await queryCache.refresh(
      queryCache.ensure({ ...getSetupQuery(), staleTime: 0 }),
    )
    if (setupState.status !== 'success')
      throw setupState.error || new Error(t('errors.requestFailed'))
    required.value = setupState.data.required
    if (!required.value) {
      const sessionState = await queryCache.refresh(
        queryCache.ensure({ ...getSessionQuery(), staleTime: 0 }),
      )
      if (sessionState.status !== 'success')
        throw sessionState.error || new Error(t('errors.requestFailed'))
      applySession(sessionState.data)
      if (sessionState.data.user)
        router.replace('/app')
    }
  }
  catch (e) {
    error.value = errorText(e)
  }
  finally {
    loading.value = false
  }
})
async function submit() {
  saving.value = true
  error.value = ''
  try {
    if (required.value) {
      await setup.mutateAsync({
        body: {
          username: username.value,
          password: password.value,
          organizationName: organizationName.value,
          timezone: timezone.value,
        },
      })
    }
    await createSession.mutateAsync({
      body: { username: username.value, password: password.value },
    })
    const next = String(route.query.next || '/app')
    router.replace(next.startsWith('/app') ? next : '/app')
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
  <div class="auth-page" :data-theme="dark ? 'dark' : 'light'">
    <section class="auth-art">
      <Brand />
      <div class="auth-visual" />
      <div class="auth-copy">
        <p class="eyebrow" un-text="teal-300">
          {{ t('login.slogan') }}
        </p>
        <h1>{{ t('login.everyHeartbeat') }}<br>{{ t('login.alwaysInSight') }}</h1>
        <p>
          {{ t('login.monitorServicesRespondToIncidentsAndKeepEveryone') }}
        </p>
      </div>
      <div class="auth-bottom" un-flex="~ items-center gap-2">
        <ShieldCheck :size="15" />{{ t('login.selfHostedYourInfrastructureYourData') }}
      </div>
    </section>
    <section class="auth-main">
      <div class="auth-form">
        <div class="auth-logo-mobile">
          <Brand />
        </div>
        <p class="eyebrow">
          {{ required ? t('login.getStarted') : t('login.welcomeBackEyebrow') }}
        </p>
        <h1>
          {{ required ? t('login.createYourWorkspace') : t('login.welcomeBack') }}
        </h1>
        <p class="muted">
          {{
            required
              ? t('login.createTheFirstAdministratorAccountToStartMonitoring')
              : t('login.signInToSeeHowYourServicesAre')
          }}
        </p>
        <div v-if="loading" class="loading-state">
          <span class="spinner" />
        </div>
        <form v-else @submit.prevent="submit">
          <Field :label="t('common.username')">
            <input v-model="username" autocomplete="username" required maxlength="100">
          </Field><Field
            :label="t('common.password')"
            :hint="required ? t('login.atLeast12CharactersUpTo72Bytes') : undefined"
          >
            <input
              v-model="password"
              type="password"
              :autocomplete="required ? 'new-password' : 'current-password'"
              required
              :minlength="required ? 12 : undefined"
              maxlength="72"
            >
          </Field><template v-if="required">
            <Field :label="t('common.organizationName')">
              <input v-model="organizationName" required>
            </Field><Field :label="t('common.organizationTimeZone')">
              <input v-model="timezone" placeholder="Asia/Shanghai" required>
            </Field>
          </template>
          <p v-if="error" class="inline-error" role="alert">
            {{ error }}
          </p>
          <button type="submit" class="button primary" :disabled="saving">
            {{
              saving
                ? t('login.connecting')
                : required
                  ? t('login.createWorkspace')
                  : t('login.signIn')
            }}<ArrowRight :size="16" />
          </button>
        </form>
        <p class="muted" un-text="10px" un-mt="7">
          {{ t('login.registrationIsClosedContactYourAdministratorForAccess') }}
        </p>
      </div>
    </section>
  </div>
</template>
