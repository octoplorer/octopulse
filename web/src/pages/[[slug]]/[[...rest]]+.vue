<script setup lang="ts">
import type { DefineQueryOptions } from '@pinia/colada'
import type { ErrorModel } from '../../client/types.gen'
import type { PublicPage } from '../../lib/types'
import { useQuery } from '@pinia/colada'
import { useHead } from '@unhead/vue'
import { useIntervalFn } from '@vueuse/core'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'

import { getPublicPageQuery, resolvePublicPageQuery } from '../../client/@pinia/colada.gen'
import AsyncState from '../../components/AsyncState.vue'
import StatusPage from '../../components/StatusPage.vue'

const { t } = useI18n({ useScope: 'global' })
const route = useRoute('/[[slug]]/[[...rest]]+')
const isDomain = computed(() => !route.params.slug || route.params.slug === 'incidents')
const query = useQuery(
  () =>
    ({
      ...(isDomain.value
        ? resolvePublicPageQuery({ query: { host: location.hostname } })
        : getPublicPageQuery({ path: { slug: String(route.params.slug) } })),
      staleTime: 5000,
    }) as DefineQueryOptions<PublicPage, ErrorModel>,
)
const rest = computed(() =>
  Array.isArray(route.params.rest) ? route.params.rest.join('/') : String(route.params.rest || ''),
)
const incidentId = computed(() =>
  isDomain.value && route.params.slug === 'incidents'
    ? rest.value.split('/')[0]
    : rest.value.startsWith('incidents/')
      ? rest.value.split('/')[1]
      : undefined,
)
useIntervalFn(() => query.refetch(), 30000)
useHead(() => {
  const config = query.data.value?.config
  return {
    title: `${config?.title || 'Octopulse'} · ${t('publicPage.serviceStatus')}`,
    meta: config
      ? [
          { name: 'description', content: config.description || undefined },
          { name: 'theme-color', content: config.brandColor || '#0f766e' },
          {
            name: 'color-scheme',
            content: config.colorScheme === 'system' ? 'light dark' : config.colorScheme,
          },
        ]
      : [],
  }
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
      <AsyncState :pending="query.isPending.value">
        <div class="empty-state">
          <h1>{{ t('publicPage.statusPageUnavailable') }}</h1>
          <p>
            {{ t('publicPage.thisAddressIsNotBoundToAPublished') }}
          </p>
          <button class="button" @click="query.refetch()">
            {{ t('publicPage.tryAgain') }}
          </button><RouterLink to="/app/login" class="button ghost" un-ml="3">
            {{
              t('publicPage.adminSignIn')
            }}
          </RouterLink>
        </div>
      </AsyncState>
    </div>
  </div>
</template>
