<script setup lang="ts">
import { useQuery } from '@pinia/colada'
import { useI18n } from 'vue-i18n'
import { listPagesQuery } from '../../../../client/@pinia/colada.gen'
import AsyncState from '../../../../components/AsyncState.vue'
import EmptyState from '../../../../components/EmptyState.vue'
import PageHeader from '../../../../components/PageHeader.vue'
import { canEdit } from '../../../../composables/api'
import { formatDate } from '../../../../composables/preferences'
import { publishedEntry } from '../../../../lib/pages'

const { t } = useI18n({ useScope: 'global' })

definePage({ meta: { title: 'navigation.statusPages' } })

const query = useQuery({
  ...listPagesQuery(),
  staleTime: 10000,
})
</script>

<template>
  <PageHeader
    :title="t('navigation.statusPages')"
    :description="t('pages.shareServiceHealthAndUpdatesWithYourOwn')"
  >
    <RouterLink v-if="canEdit()" to="/app/pages/new" class="button primary">
      <span class="i-lucide-plus" un-w="15px" un-h="15px" aria-hidden="true" />{{ t('common.createStatusPage') }}
    </RouterLink>
  </PageHeader><AsyncState :pending="query.isPending.value" :error="query.error.value" @retry="query.refetch()">
    <EmptyState
      v-if="!query.data.value?.items.length"
      :title="t('pages.buildYourFirstStatusPage')"
      :description="t('pages.chooseBrandingPublicServicesAndGroupsPublishOn')"
    >
      <RouterLink v-if="canEdit()" to="/app/pages/new" class="button primary">
        {{
          t('common.createStatusPage')
        }}
      </RouterLink>
    </EmptyState>
    <div v-else class="cards-grid">
      <article v-for="page in query.data.value.items" :key="page.id" class="card page-tile">
        <div class="page-preview-icon">
          <span class="preview-status" :style="{ background: `${page.draft.brandColor}25` }" /><span
            un-w="75%"
          /><span un-w="95%" /><span un-w="85%" /><span un-w="65%" />
        </div>
        <div un-flex="~ justify-between items-center gap-3">
          <h2>{{ page.name }}</h2>
          <span class="pill">{{
            page.publishedAt ? t('common.published') : t('common.draft')
          }}</span>
        </div>
        <p class="muted" un-text="xs">
          /{{ page.publishedAt ? publishedEntry(page).slug : page.slug }}
        </p>
        <p
          v-if="page.publishedAt ? publishedEntry(page).domain : page.domain"
          class="muted"
          un-text="xs"
          un-flex="~ items-center gap-1.5"
        >
          <span class="i-lucide-globe" un-w="11px" un-h="11px" aria-hidden="true" />{{ page.publishedAt ? publishedEntry(page).domain : page.domain }}
        </p>
        <div class="tile-actions">
          <RouterLink :to="`/app/pages/${page.id}`" class="button small">
            <span class="i-lucide-pencil" un-w="12px" un-h="12px" aria-hidden="true" />{{
              canEdit() ? t('pages.customize') : t('pages.view')
            }}
          </RouterLink><a
            v-if="page.publishedAt"
            :href="publishedEntry(page).url"
            target="_blank"
            rel="noopener"
            class="button small ghost"
          >{{ t('pages.visit') }}<span class="i-lucide-arrow-up-right" un-w="13px" un-h="13px" aria-hidden="true" /></a><span v-else class="muted" un-text="xs">{{ formatDate(page.updatedAt) }}</span>
        </div>
      </article>
    </div>
  </AsyncState>
</template>
