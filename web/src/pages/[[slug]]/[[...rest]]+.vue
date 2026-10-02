<script setup lang="ts">
import { computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useRecord } from '../../lib/data'
import { getPublicPageQuery, resolvePublicPageQuery } from '../../client/@pinia/colada.gen'
import type { PublicPage } from '../../lib/types'
import { t } from '../../lib/preferences'
import { useIntervalFn } from '@vueuse/core'
import StatusPage from '../../components/StatusPage.vue'
import AsyncState from '../../components/AsyncState.vue'
const route = useRoute('/[[slug]]/[[...rest]]+'),
  isDomain = computed(() => !route.params.slug || route.params.slug === 'incidents'),
  query = useRecord<PublicPage>(
    () =>
      isDomain.value
        ? `/api/public/resolve?host=${encodeURIComponent(location.hostname)}`
        : `/api/public/pages/${encodeURIComponent(String(route.params.slug))}`,
    () =>
      isDomain.value
        ? resolvePublicPageQuery({ query: { host: location.hostname } })
        : getPublicPageQuery({ path: { slug: String(route.params.slug) } }),
  ),
  rest = computed(() =>
    Array.isArray(route.params.rest)
      ? route.params.rest.join('/')
      : String(route.params.rest || ''),
  ),
  incidentId = computed(() =>
    isDomain.value && route.params.slug === 'incidents'
      ? rest.value.split('/')[0]
      : rest.value.startsWith('incidents/')
        ? rest.value.split('/')[1]
        : undefined,
  )
useIntervalFn(() => query.refresh(), 30000)
watch(
  () => query.data.value,
  (page) => {
    if (page) document.title = `${page.config.title} · ${t('服务状态', 'Service status')}`
  },
  { immediate: true },
)
</script>
<template>
  <StatusPage
    v-if="query.data.value"
    :page="query.data.value"
    :incident-id="incidentId"
    :stale="!!query.error.value"
    :path-base="isDomain ? '' : `/${route.params.slug}`"
  />
  <div v-else class="public-page">
    <div class="public-inner">
      <AsyncState :pending="query.isPending.value"
        ><div class="empty-state">
          <h1>{{ t('状态页暂不可用', 'Status page unavailable') }}</h1>
          <p>
            {{
              t(
                '此地址尚未绑定已发布页面，或服务暂时无法连接。',
                'This address is not bound to a published page, or the service cannot be reached.',
              )
            }}
          </p>
          <button class="button" @click="query.refresh()">{{ t('重新连接', 'Try again') }}</button
          ><RouterLink to="/app/login" class="button ghost" un-ml="3">{{
            t('管理登录', 'Admin sign in')
          }}</RouterLink>
        </div></AsyncState
      >
    </div>
  </div>
</template>
