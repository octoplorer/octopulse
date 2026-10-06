<script setup lang="ts">
import { useMutation, useQueryCache } from '@pinia/colada'
import { useForm } from '@tanstack/vue-form'
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
import { FieldError, FieldInput } from '../../components/ui/field'
import { Spinner } from '../../components/ui/spinner'
import { applySession } from '../../composables/api'
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
const error = ref('')
const formApi = useForm({
  defaultValues: {
    username: '',
    password: '',
    organizationName: 'Octopulse',
    timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
  },
  onSubmit: async ({ value }) => {
    error.value = ''
    try {
      if (required.value)
        await setup.mutateAsync({ body: value })
      await createSession.mutateAsync({
        body: { username: value.username, password: value.password },
      })
      formApi.setFieldValue('password', '')
      const next = String(route.query.next || '/app')
      await router.replace(next.startsWith('/app') ? next : '/app')
    }
    catch (e) {
      error.value = errorText(e)
    }
  },
})
const saving = formApi.useSelector(state => state.isSubmitting)
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
        await router.replace('/app')
    }
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
  <div grid="~ cols-[minmax(0,0.9fr)_minmax(0,1.1fr)]" min-h="screen" bg="canvas" un-text="default" class="[@media(max-width:900px)]:grid-cols-1">
    <section flex="~ col" justify="between" min-h="screen" p="48px" bg="tint" border="r-1 solid line" un-text="default" class="[&_.brand-icon]:bg-base [&_.brand]:text-22px [@media(max-width:900px)]:hidden">
      <Brand />
      <div relative my="auto" max-w="460px" class="[&_h1]:text-[clamp(32px,3.2vw,46px)] [&_h1]:font-500 [&_h1]:leading-[1.2] [&_h1]:tracking-[-1.2px] [&>p]:mt-24px [&>p]:max-w-360px [&>p]:text-14px [&>p]:text-subtle">
        <p class="eyebrow" un-text="12px subtle" font="500" mb="8px">
          {{ t('login.slogan') }}
        </p>
        <h1>{{ t('login.everyHeartbeat') }}<br>{{ t('login.alwaysInSight') }}</h1>
        <p>
          {{ t('login.monitorServicesRespondToIncidentsAndKeepEveryone') }}
        </p>
      </div>
      <div flex="~ items-center gap-2" un-text="12px subtle">
        <span w="15px" h="15px" aria-hidden="true" class="i-lucide-shield-check" />{{ t('login.selfHostedYourInfrastructureYourData') }}
      </div>
    </section>
    <section flex="~ items-center justify-center" p="48px" bg="base" class="[@media(max-width:900px)]:min-h-screen [@media(max-width:700px)]:px-20px [@media(max-width:700px)]:py-32px">
      <div w="full" max-w="360px" class="[&_h1]:text-26px [&>p]:mt-10px [&>p]:mb-30px [&_form]:flex [&_form]:flex-col [&_form]:gap-18px [&_form_.button]:min-h-40px [&_form_.button]:px-12px [&_form_.button]:py-10px">
        <div hidden mb="35px" class="[@media(max-width:900px)]:block">
          <Brand />
        </div>
        <p class="eyebrow" un-text="12px subtle" font="500" mb="8px">
          {{ required ? t('login.getStarted') : t('login.welcomeBackEyebrow') }}
        </p>
        <h1>
          {{ required ? t('login.createYourWorkspace') : t('login.welcomeBack') }}
        </h1>
        <p class="muted" un-text="13px subtle">
          {{
            required
              ? t('login.createTheFirstAdministratorAccountToStartMonitoring')
              : t('login.signInToSeeHowYourServicesAre')
          }}
        </p>
        <div v-if="loading" flex="~ justify-center items-center gap-10px" p="60px" un-text="12px subtle">
          <Spinner />
        </div>
        <form v-else @submit.prevent="formApi.handleSubmit()">
          <formApi.Field v-slot="{ field }" name="username">
            <Field :label="t('common.username')">
              <FieldInput :name="field.name" :model-value="field.state.value" autocomplete="username" required maxlength="100" @update:model-value="field.handleChange($event)" @blur="field.handleBlur" />
            </Field>
          </formApi.Field>
          <formApi.Field v-slot="{ field }" name="password">
            <Field :label="t('common.password')" :hint="required ? t('login.atLeast12CharactersUpTo72Bytes') : undefined">
              <FieldInput :name="field.name" :model-value="field.state.value" type="password" :autocomplete="required ? 'new-password' : 'current-password'" required :minlength="required ? 12 : undefined" maxlength="72" @update:model-value="field.handleChange($event)" @blur="field.handleBlur" />
            </Field>
          </formApi.Field>
          <template v-if="required">
            <formApi.Field v-slot="{ field }" name="organizationName">
              <Field :label="t('common.organizationName')">
                <FieldInput :name="field.name" :model-value="field.state.value" required @update:model-value="field.handleChange($event)" @blur="field.handleBlur" />
              </Field>
            </formApi.Field>
            <formApi.Field v-slot="{ field }" name="timezone">
              <Field :label="t('common.organizationTimeZone')">
                <FieldInput :name="field.name" :model-value="field.state.value" placeholder="Asia/Shanghai" required @update:model-value="field.handleChange($event)" @blur="field.handleBlur" />
              </Field>
            </formApi.Field>
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
        <p mt="7" class="muted" un-text="13px subtle">
          {{ t('login.registrationIsClosedContactYourAdministratorForAccess') }}
        </p>
      </div>
    </section>
  </div>
</template>
