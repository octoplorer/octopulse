<script setup lang="ts">
import { useQuery } from '@pinia/colada'
import { useI18n } from 'vue-i18n'
import { listPagesQuery } from '../../../../client/@pinia/colada.gen'
import { Badge } from '../../../../components/ui/badge'
import { Banner } from '../../../../components/ui/banner'
import { PageHeader } from '../../../../components/ui/blocks/page-header'
import { Button } from '../../../../components/ui/button'
import { Empty } from '../../../../components/ui/empty'
import { LayerCard, LayerCardPrimary } from '../../../../components/ui/layer-card'
import { Loader } from '../../../../components/ui/loader'
import { Text } from '../../../../components/ui/text'
import { canEdit } from '../../../../composables/api'
import { formatDate } from '../../../../composables/preferences'
import { errorText } from '../../../../lib/errors'
import { publishedEntry } from '../../../../lib/pages'

const { t } = useI18n({ useScope: 'global' })

definePage({ meta: { title: 'navigation.status-pages' } })

const query = useQuery({
  ...listPagesQuery(),
  staleTime: 10000,
})
</script>

<template>
  <PageHeader :title="t('navigation.status-pages')" :description="t('pages.share-service-health-and-updates-with-your-own')" class="mb-6">
    <template #actions>
      <Button v-if="canEdit()" variant="primary" as-child>
        <RouterLink to="/app/pages/new">
          <span class="i-lucide-plus size-4" aria-hidden="true" />{{ t('common.create-status-page') }}
        </RouterLink>
      </Button>
    </template>
  </PageHeader>
  <Loader v-if="query.isPending.value" :label="t('async-state.loading-data')" class="flex! w-full justify-center p-15 text-size-xs">
    {{ t('async-state.loading-data') }}
  </Loader>
  <Banner v-else-if="query.error.value" variant="error">
    {{ errorText(query.error.value) }}
    <Button variant="ghost" @click="query.refetch()">
      {{ t('async-state.retry') }}
    </Button>
  </Banner>
  <template v-else>
    <Empty v-if="!query.data.value?.items.length" :title="t('pages.build-your-first-status-page')" :description="t('pages.choose-branding-public-services-and-groups-publish-on')">
      <template #icon>
        <span class="i-lucide-globe size-8 text-subtle" aria-hidden="true" />
      </template>
      <template #actions>
        <Button v-if="canEdit()" variant="primary" as-child>
          <RouterLink to="/app/pages/new">
            {{ t('common.create-status-page') }}
          </RouterLink>
        </Button>
      </template>
    </Empty>
    <div v-else class="grid grid-cols-[repeat(auto-fill,minmax(min(100%,320px),1fr))] gap-6">
      <LayerCard v-for="page in query.data.value.items" :key="page.id" class="flex flex-col">
        <LayerCardPrimary class="flex flex-1 flex-col gap-5">
          <div class="space-y-3">
            <div class="flex items-start justify-between gap-3">
              <div class="flex min-w-0 items-start gap-3">
                <span class="flex size-9 shrink-0 items-center justify-center rounded-lg bg-recessed" :style="{ color: page.draft.brandColor }" aria-hidden="true">
                  <span class="i-lucide-globe size-5" />
                </span>
                <Text as="h2" variant="heading" class="min-w-0 [overflow-wrap:anywhere]">
                  {{ page.name }}
                </Text>
              </div><Badge :variant="page.publishedAt ? 'success' : 'outline'" class="shrink-0">
                {{ page.publishedAt ? t('common.published') : t('common.draft') }}
              </Badge>
            </div>
            <Text as="p" size="sm" variant="secondary" class="[overflow-wrap:anywhere]">
              /{{ page.publishedAt ? publishedEntry(page).slug : page.slug }}
            </Text>
            <Text v-if="page.publishedAt ? publishedEntry(page).domain : page.domain" as="p" size="sm" variant="secondary" class="flex items-center gap-1.5 [overflow-wrap:anywhere]">
              <span class="i-lucide-globe size-3 shrink-0" aria-hidden="true" />{{ page.publishedAt ? publishedEntry(page).domain : page.domain }}
            </Text>
            <Text v-if="page.draft.description" as="p" size="sm" variant="secondary" class="[overflow-wrap:anywhere]">
              {{ page.draft.description }}
            </Text>
            <Text as="p" size="xs" variant="secondary">
              {{ t('common.updated') }} {{ formatDate(page.updatedAt) }}
            </Text>
          </div>
          <div class="mt-auto flex flex-wrap items-center justify-between gap-3 border-t border-line pt-4">
            <Button size="sm" as-child>
              <RouterLink :to="`/app/pages/${page.id}`">
                <span class="i-lucide-pencil size-3.5" aria-hidden="true" />{{ canEdit() ? t('pages.customize') : t('pages.view') }}
              </RouterLink>
            </Button>
            <Button v-if="page.publishedAt" variant="ghost" size="sm" as-child>
              <a :href="publishedEntry(page).url" target="_blank" rel="noopener noreferrer">{{ t('pages.visit') }}<span class="i-lucide-arrow-up-right size-3.5" aria-hidden="true" /></a>
            </Button>
          </div>
        </LayerCardPrimary>
      </LayerCard>
    </div>
  </template>
</template>
