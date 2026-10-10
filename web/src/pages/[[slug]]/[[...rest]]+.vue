<script setup lang="ts">
import { useQuery } from '@pinia/colada'
import { useHead } from '@unhead/vue'
import { usePreferredDark } from '@vueuse/core'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { getPublicPageQuery, resolvePublicPageQuery } from '../../client/@pinia/colada.gen'
import StatusPage from '../../components/StatusPage.vue'
import { Button } from '../../components/ui/button'
import { Empty, EmptyDescription, EmptyTitle } from '../../components/ui/empty'
import { Loader } from '../../components/ui/loader'
import { usePollingEnabled } from '../../composables/polling'

const { t } = useI18n({ useScope: 'global' })
const route = useRoute('/[[slug]]/[[...rest]]+')
const prefersDark = usePreferredDark()
const pollingEnabled = usePollingEnabled()
const isDomain = computed(() => !route.params.slug || route.params.slug === 'incidents')
const query = useQuery(
  () =>
    ({
      ...(isDomain.value
        ? resolvePublicPageQuery({ query: { host: location.hostname } })
        : getPublicPageQuery({ path: { slug: String(route.params.slug) } })),
      staleTime: 5000,
      enabled: pollingEnabled.value,
      autoRefetch: 30000,
    }),
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
useHead(() => {
  const config = query.data.value?.config
  return {
    title: `${config?.title || 'Octopulse'} · ${t('public-page.service-status')}`,
    htmlAttrs: config
      ? {
          'data-mode': config.colorScheme === 'dark' || (config.colorScheme === 'system' && prefersDark.value) ? 'dark' : 'light',
        }
      : undefined,
    meta: config
      ? [
          { name: 'description', content: config.description || undefined },
          { name: 'theme-color', content: config.brandColor || '#2563eb' },
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
  <StatusPage v-if="query.data.value" :page="query.data.value" :incident-id="incidentId" :stale="!!query.error.value" :path-base="isDomain ? '' : `/${route.params.slug}`" />
  <main v-else bg="canvas" un-text="default" min-h="screen" py="48px" px="24px" font="sans" class="[@media(max-width:700px)]:px-17px [@media(max-width:700px)]:py-25px">
    <div class="public-inner" max-w="870px" mx="auto">
      <Loader v-if="query.isPending.value" :label="t('async-state.loading-data')" class="flex! w-full justify-center p-15 text-size-xs" role="status">
        {{ t('async-state.loading-data') }}
      </Loader>
      <Empty v-else>
        <EmptyTitle as="h1">
          {{ t('public-page.status-page-unavailable') }}
        </EmptyTitle>
        <EmptyDescription>
          {{ t('public-page.this-address-is-not-bound-to-a-published') }}
        </EmptyDescription>
        <Button @click="query.refetch()">
          {{ t('public-page.try-again') }}
        </Button><Button variant="ghost" as-child>
          <RouterLink to="/app/login" ml="3">
            {{
              t('public-page.admin-sign-in')
            }}
          </RouterLink>
        </Button>
      </Empty>
    </div>
  </main>
</template>
