<script setup lang="ts">
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
import { Button } from '../../components/ui/button'
import { FieldError } from '../../components/ui/field'
import { Spinner } from '../../components/ui/spinner'
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
  <div data-theme-boundary :data-theme="dark ? 'dark' : 'light'" grid="~ cols-[minmax(0,0.9fr)_minmax(0,1.1fr)]" min-h="screen" bg="$bg" un-text="$text" class="[@media(max-width:900px)]:grid-cols-1">
    <section flex="~ col" justify="between" min-h="screen" p="48px" bg="$surface-soft" border="r-1 solid line" un-text="$text" class="[&_.brand-icon]:bg-$surface [&_.brand]:text-22px [@media(max-width:900px)]:hidden">
      <Brand />
      <div relative my="auto" max-w="460px" class="[&_h1]:text-[clamp(32px,3.2vw,46px)] [&_h1]:font-500 [&_h1]:leading-[1.2] [&_h1]:tracking-[-1.2px] [&>p]:mt-24px [&>p]:max-w-360px [&>p]:text-14px [&>p]:text-$muted">
        <p class="eyebrow" un-text="12px $muted" font="500" mb="8px">
          {{ t('login.slogan') }}
        </p>
        <h1>{{ t('login.everyHeartbeat') }}<br>{{ t('login.alwaysInSight') }}</h1>
        <p>
          {{ t('login.monitorServicesRespondToIncidentsAndKeepEveryone') }}
        </p>
      </div>
      <div flex="~ items-center gap-2" un-text="12px $muted">
        <span w="15px" h="15px" aria-hidden="true" class="i-lucide-shield-check" />{{ t('login.selfHostedYourInfrastructureYourData') }}
      </div>
    </section>
    <section flex="~ items-center justify-center" p="48px" bg="$surface" class="[@media(max-width:900px)]:min-h-screen [@media(max-width:700px)]:px-20px [@media(max-width:700px)]:py-32px">
      <div w="full" max-w="360px" class="[&_h1]:text-26px [&>p]:mt-10px [&>p]:mb-30px [&_form]:flex [&_form]:flex-col [&_form]:gap-18px [&_form_.button]:min-h-40px [&_form_.button]:px-12px [&_form_.button]:py-10px">
        <div hidden mb="35px" class="[@media(max-width:900px)]:block">
          <Brand />
        </div>
        <p class="eyebrow" un-text="12px $muted" font="500" mb="8px">
          {{ required ? t('login.getStarted') : t('login.welcomeBackEyebrow') }}
        </p>
        <h1>
          {{ required ? t('login.createYourWorkspace') : t('login.welcomeBack') }}
        </h1>
        <p class="muted" un-text="13px $muted">
          {{
            required
              ? t('login.createTheFirstAdministratorAccountToStartMonitoring')
              : t('login.signInToSeeHowYourServicesAre')
          }}
        </p>
        <div v-if="loading" flex="~ justify-center items-center gap-10px" p="60px" un-text="12px $muted">
          <Spinner />
        </div>
        <form v-else @submit.prevent="submit">
          <Field :label="t('common.username')">
            <input v-model="username" autocomplete="username" required maxlength="100">
          </Field><Field :label="t('common.password')" :hint="required ? t('login.atLeast12CharactersUpTo72Bytes') : undefined">
            <input v-model="password" type="password" :autocomplete="required ? 'new-password' : 'current-password'" required :minlength="required ? 12 : undefined" maxlength="72">
          </Field><template v-if="required">
            <Field :label="t('common.organizationName')">
              <input v-model="organizationName" required>
            </Field><Field :label="t('common.organizationTimeZone')">
              <input v-model="timezone" placeholder="Asia/Shanghai" required>
            </Field>
          </template>
          <FieldError v-if="error" as="p" role="alert" py="10px" px="0">
            {{ error }}
          </FieldError>
          <Button type="submit" :disabled="saving" variant="primary">
            {{
              saving
                ? t('login.connecting')
                : required
                  ? t('login.createWorkspace')
                  : t('login.signIn')
            }}<span w="16px" h="16px" aria-hidden="true" class="i-lucide-arrow-right" />
          </Button>
        </form>
        <p mt="7" class="muted" un-text="13px $muted">
          {{ t('login.registrationIsClosedContactYourAdministratorForAccess') }}
        </p>
      </div>
    </section>
  </div>
</template>
