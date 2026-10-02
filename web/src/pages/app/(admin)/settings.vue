<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import { Save, Settings, UserRound } from '@lucide/vue'
import { Tabs } from '@ark-ui/vue/tabs'
import { api, isAdmin, currentUser } from '../../../lib/api'
import type { Settings as OrganizationSettings, User } from '../../../lib/types'
import { t, locale, theme, timezone } from '../../../lib/preferences'
import { notify, errorText } from '../../../lib/notices'
import PageHeader from '../../../components/PageHeader.vue'
import Field from '../../../components/Field.vue'
import AsyncState from '../../../components/AsyncState.vue'

definePage({ meta: { title: ['设置', 'Settings'] } })

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
    locale: currentUser.value?.locale || locale.value,
    timezone: currentUser.value?.timezone || timezone.value,
    oldPassword: '',
    password: '',
  })
onMounted(async () => {
  try {
    Object.assign(form, await api<OrganizationSettings>('settings'))
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
    Object.assign(
      form,
      await api<OrganizationSettings>('settings', { method: 'PATCH', body: form }),
    )
    notify(t('组织设置已保存', 'Organization settings saved'))
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
    currentUser.value = await api<User>('profile', {
      method: 'PATCH',
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
    notify(t('个人设置已保存', 'Personal settings saved'))
  } catch (e) {
    error.value = errorText(e)
  } finally {
    saving.value = false
  }
}
</script>
<template>
  <PageHeader
    :title="t('设置', 'Settings')"
    :description="
      t(
        '为组织与自己的工作习惯配置空间。',
        'Configure the workspace for your organization and your preferences.',
      )
    "
  /><AsyncState :pending="loading"
    ><div v-if="error" class="validation-error" role="alert">{{ error }}</div>
    <section class="card" un-max-w="4xl">
      <Tabs.Root v-model="tab"
        ><Tabs.List class="tabs-list"
          ><Tabs.Trigger class="tabs-trigger" value="personal">{{
            t('个人偏好', 'Personal preferences')
          }}</Tabs.Trigger
          ><Tabs.Trigger class="tabs-trigger" value="organization">{{
            t('组织设置', 'Organization settings')
          }}</Tabs.Trigger
          ><Tabs.Trigger v-if="isAdmin()" class="tabs-trigger" value="retention">{{
            t('数据留存', 'Data retention')
          }}</Tabs.Trigger></Tabs.List
        ><Tabs.Content value="personal"
          ><form @submit.prevent="saveProfile">
            <div class="form-section">
              <h2>{{ t('显示偏好', 'Display preferences') }}</h2>
              <p class="muted">
                {{
                  t(
                    '个人时区用于展示，不改变探测与存储的 UTC 时间。',
                    'Your display time zone does not change UTC collection or storage.',
                  )
                }}
              </p>
              <div class="form-grid">
                <Field :label="t('显示名称', 'Display name')"
                  ><input v-model="profile.name" /></Field
                ><Field :label="t('语言', 'Language')"
                  ><select v-model="profile.locale">
                    <option value="zh-CN">简体中文</option>
                    <option value="en">English</option>
                  </select></Field
                ><Field :label="t('显示时区', 'Display time zone')"
                  ><input v-model="profile.timezone" placeholder="Asia/Shanghai" required /></Field
                ><Field :label="t('后台深浅色', 'Admin color scheme')"
                  ><select v-model="theme">
                    <option value="system">{{ t('跟随系统', 'System') }}</option>
                    <option value="light">{{ t('浅色', 'Light') }}</option>
                    <option value="dark">{{ t('深色', 'Dark') }}</option>
                  </select></Field
                >
              </div>
            </div>
            <div class="form-section">
              <h2>{{ t('修改密码', 'Change password') }}</h2>
              <p class="muted">
                {{ t('留空表示保留当前密码。', 'Leave empty to keep your password.') }}
              </p>
              <div class="form-grid">
                <Field :label="t('当前密码', 'Current password')"
                  ><input
                    v-model="profile.oldPassword"
                    type="password"
                    autocomplete="current-password"
                    :required="!!profile.password" /></Field
                ><Field
                  :label="t('新密码', 'New password')"
                  :hint="
                    t('至少 12 字符，最多 72 字节。', 'At least 12 characters, up to 72 bytes.')
                  "
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
                  <Save :size="14" />{{ t('保存个人设置', 'Save preferences') }}
                </button>
              </div>
            </div>
          </form></Tabs.Content
        ><Tabs.Content value="organization"
          ><form @submit.prevent="saveOrganization">
            <div class="form-section">
              <h2>{{ t('组织信息', 'Organization') }}</h2>
              <p class="muted">
                {{
                  t(
                    '新账号与维护配置以组织设置为默认值。',
                    'Organization settings provide defaults for accounts and maintenance.',
                  )
                }}
              </p>
              <div class="form-grid">
                <Field :label="t('组织名称', 'Organization name')"
                  ><input v-model="form.organizationName" :disabled="!isAdmin()" required /></Field
                ><Field :label="t('组织时区', 'Organization time zone')"
                  ><input v-model="form.timezone" :disabled="!isAdmin()" required /></Field
                ><Field :label="t('默认语言', 'Default language')"
                  ><select v-model="form.locale" :disabled="!isAdmin()">
                    <option value="zh-CN">简体中文</option>
                    <option value="en">English</option>
                  </select></Field
                ><Field
                  class="span-full"
                  :label="
                    t(
                      '可绑定状态页的域名（每行一个）',
                      'Allowed status page domains (one per line)',
                    )
                  "
                  :hint="
                    t(
                      '仅主机名，不含协议和路径；先设置 DNS 与反代 HTTPS。',
                      'Hostnames only, without scheme or path. Configure DNS and reverse-proxy HTTPS first.',
                    )
                  "
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
                  <Save :size="14" />{{ t('保存组织设置', 'Save organization') }}
                </button>
              </div>
            </div>
          </form></Tabs.Content
        ><Tabs.Content v-if="isAdmin()" value="retention"
          ><form @submit.prevent="saveOrganization">
            <div class="form-section">
              <h2>{{ t('历史数据留存', 'History retention') }}</h2>
              <p class="muted">
                {{
                  t(
                    '调整原始记录与聚合数据的保存窗口。清理任务分批运行。',
                    'Configure raw and aggregated history windows. Cleanup runs in batches.',
                  )
                }}
              </p>
              <div class="form-grid">
                <Field :label="t('原始轮次（天）', 'Raw rounds (days)')"
                  ><input
                    v-model.number="form.retention.roundDays"
                    type="number"
                    min="1"
                    required /></Field
                ><Field :label="t('探测尝试（天）', 'Attempt details (days)')"
                  ><input
                    v-model.number="form.retention.attemptDays"
                    type="number"
                    min="1"
                    required /></Field
                ><Field :label="t('5 分钟聚合（天）', '5-minute aggregates (days)')"
                  ><input
                    v-model.number="form.retention.fiveMinuteDays"
                    type="number"
                    min="1"
                    required /></Field
                ><Field
                  :label="
                    t('小时聚合与状态区间（月）', 'Hourly aggregates & state intervals (months)')
                  "
                  ><input
                    v-model.number="form.retention.historyMonths"
                    type="number"
                    min="1"
                    required
                /></Field>
              </div>
              <p class="note" un-mt="6">
                {{
                  t(
                    '默认不保存完整请求/响应正文或秘密头。聚合保留时间权重与覆盖信息，统计不能用样本数替代状态时长。',
                    'Complete request/response bodies and secret headers are not stored by default. Aggregates retain time weights and coverage instead of substituting sample counts for duration.',
                  )
                }}
              </p>
              <div class="form-actions">
                <button class="button primary" :disabled="saving">
                  <Save :size="14" />{{ t('保存留存策略', 'Save retention') }}
                </button>
              </div>
            </div>
          </form></Tabs.Content
        ></Tabs.Root
      >
    </section></AsyncState
  >
</template>
