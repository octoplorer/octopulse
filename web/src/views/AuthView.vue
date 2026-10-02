<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowRight, ShieldCheck } from '@lucide/vue'
import Brand from '../components/Brand.vue'
import Field from '../components/Field.vue'
import { dark, t } from '../lib/preferences'
import { api, login, loadSession } from '../lib/api'
import { errorText } from '../lib/notices'
const route = useRoute(),
  router = useRouter(),
  required = ref(false),
  loading = ref(true),
  saving = ref(false),
  error = ref(''),
  username = ref(''),
  password = ref(''),
  organizationName = ref('Octopulse'),
  timezone = ref(Intl.DateTimeFormat().resolvedOptions().timeZone)
onMounted(async () => {
  try {
    required.value = (await api<{ required: boolean }>('setup')).required
    if (!required.value && (await loadSession()).user) {
      router.replace('/app')
    }
  } catch (e) {
    error.value = errorText(e)
  } finally {
    loading.value = false
  }
})
async function submit() {
  saving.value = true
  error.value = ''
  try {
    if (required.value)
      await api('setup', {
        method: 'POST',
        body: {
          username: username.value,
          password: password.value,
          organizationName: organizationName.value,
          timezone: timezone.value,
        },
      })
    await login(username.value, password.value)
    const next = String(route.query.next || '/app')
    router.replace(next.startsWith('/app') ? next : '/app')
  } catch (e) {
    error.value = errorText(e)
  } finally {
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
        <p class="eyebrow" un-text="teal-300">KEEP A PULSE ON YOUR SERVICES</p>
        <h1>
          {{ t('每一次心跳，', 'Every heartbeat.') }}<br />{{
            t('都值得被看见。', 'Always in sight.')
          }}
        </h1>
        <p>
          {{
            t(
              '统一查看服务状态、处理故障与发布进展。让团队始终知道系统正在发生什么。',
              'Monitor services, respond to incidents, and keep everyone informed — from one calm workspace.',
            )
          }}
        </p>
      </div>
      <div class="auth-bottom" un-flex="~ items-center gap-2">
        <ShieldCheck :size="15" />{{
          t('单组织自托管 · 数据由你掌控', 'Self-hosted · Your infrastructure, your data')
        }}
      </div>
    </section>
    <section class="auth-main">
      <div class="auth-form">
        <div class="auth-logo-mobile"><Brand /></div>
        <p class="eyebrow">{{ required ? 'GET STARTED' : 'WELCOME BACK' }}</p>
        <h1>
          {{
            required
              ? t('建立你的监控空间', 'Create your workspace')
              : t('欢迎回来', 'Welcome back')
          }}
        </h1>
        <p class="muted">
          {{
            required
              ? t(
                  '创建首个管理员账号，开始监控你的服务。',
                  'Create the first administrator account to start monitoring.',
                )
              : t('登录以查看服务与团队的运行状态。', 'Sign in to see how your services are doing.')
          }}
        </p>
        <div v-if="loading" class="loading-state"><span class="spinner" /></div>
        <form v-else @submit.prevent="submit">
          <Field :label="t('用户名', 'Username')"
            ><input v-model="username" autocomplete="username" required maxlength="100" /></Field
          ><Field
            :label="t('密码', 'Password')"
            :hint="
              required
                ? t('至少 12 个字符，最多 72 字节。', 'At least 12 characters, up to 72 bytes.')
                : undefined
            "
            ><input
              v-model="password"
              type="password"
              :autocomplete="required ? 'new-password' : 'current-password'"
              required
              :minlength="required ? 12 : undefined"
              maxlength="72" /></Field
          ><template v-if="required"
            ><Field :label="t('组织名称', 'Organization name')"
              ><input v-model="organizationName" required /></Field
            ><Field :label="t('组织时区', 'Organization time zone')"
              ><input v-model="timezone" placeholder="Asia/Shanghai" required /></Field
          ></template>
          <p v-if="error" class="inline-error" role="alert">{{ error }}</p>
          <button type="submit" class="button primary" :disabled="saving">
            {{
              saving
                ? t('正在连接…', 'Connecting…')
                : required
                  ? t('创建工作空间', 'Create workspace')
                  : t('登录工作空间', 'Sign in')
            }}<ArrowRight :size="16" />
          </button>
        </form>
        <p class="muted" un-text="10px" un-mt="7">
          {{
            t(
              '此实例不开放注册。需要访问权限，请联系你的管理员。',
              'Registration is closed. Contact your administrator for access.',
            )
          }}
        </p>
      </div>
    </section>
  </div>
</template>
