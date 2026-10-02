<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { computed, watchEffect } from 'vue'
import { useRoute } from 'vue-router'
import { useQuery, type DefineQueryOptions } from '@pinia/colada'
import { getPublicPageQuery, resolvePublicPageQuery } from '../../client/@pinia/colada.gen'
import type { ErrorModel } from '../../client/types.gen'
import type { PublicPage } from '../../lib/types'

import { useIntervalFn } from '@vueuse/core'
import StatusPage from '../../components/StatusPage.vue'
import AsyncState from '../../components/AsyncState.vue'

const { t } = useI18n({ useScope: 'global' })
const route = useRoute('/[[slug]]/[[...rest]]+'),
  isDomain = computed(() => !route.params.slug || route.params.slug === 'incidents'),
  query = useQuery(
    () =>
      ({
        ...(isDomain.value
          ? resolvePublicPageQuery({ query: { host: location.hostname } })
          : getPublicPageQuery({ path: { slug: String(route.params.slug) } })),
        staleTime: 5000,
      }) as DefineQueryOptions<PublicPage, ErrorModel>,
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
useIntervalFn(() => query.refetch(), 30000)
watchEffect(() => {
  const page = query.data.value
  if (page) document.title = `${page.config.title} · ${t('publicPage.serviceStatus')}`
})
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
          <h1>{{ t('publicPage.statusPageUnavailable') }}</h1>
          <p>
            {{ t('publicPage.thisAddressIsNotBoundToAPublished') }}
          </p>
          <button class="button" @click="query.refetch()">{{ t('publicPage.tryAgain') }}</button
          ><RouterLink to="/app/login" class="button ghost" un-ml="3">{{
            t('publicPage.adminSignIn')
          }}</RouterLink>
        </div></AsyncState
      >
    </div>
  </div>
</template>
