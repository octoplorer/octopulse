<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { Plus, ArrowUpRight, Pencil, Globe } from '@lucide/vue'
import { useQuery, type DefineQueryOptions } from '@pinia/colada'
import { listPagesQuery } from '../../../../client/@pinia/colada.gen'
import type { ErrorModel } from '../../../../client/types.gen'
import type { Page } from '../../../../lib/types'
import { formatDate } from '../../../../lib/preferences'
import { publishedEntry } from '../../../../lib/pages'
import { canEdit } from '../../../../lib/api'
import PageHeader from '../../../../components/PageHeader.vue'
import EmptyState from '../../../../components/EmptyState.vue'
import AsyncState from '../../../../components/AsyncState.vue'
const { t } = useI18n({ useScope: 'global' })

definePage({ meta: { title: 'navigation.statusPages' } })

const query = useQuery({
  ...listPagesQuery(),
  staleTime: 10000,
} as DefineQueryOptions<{ items: Page[] }, ErrorModel>)
</script>
<template>
  <PageHeader
    :title="t('navigation.statusPages')"
    :description="t('pages.shareServiceHealthAndUpdatesWithYourOwn')"
    ><RouterLink v-if="canEdit()" to="/app/pages/new" class="button primary"
      ><Plus :size="15" />{{ t('common.createStatusPage') }}</RouterLink
    ></PageHeader
  ><AsyncState :pending="query.isPending.value" :error="query.error.value" @retry="query.refetch()"
    ><EmptyState
      v-if="!query.data.value?.items.length"
      :title="t('pages.buildYourFirstStatusPage')"
      :description="t('pages.chooseBrandingPublicServicesAndGroupsPublishOn')"
      ><RouterLink v-if="canEdit()" to="/app/pages/new" class="button primary">{{
        t('common.createStatusPage')
      }}</RouterLink></EmptyState
    >
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
        <p class="muted" un-text="11px">
          /{{ page.publishedAt ? publishedEntry(page).slug : page.slug }}
        </p>
        <p
          v-if="page.publishedAt ? publishedEntry(page).domain : page.domain"
          class="muted"
          un-text="11px"
          un-flex="~ items-center gap-1.5"
        >
          <Globe :size="11" />{{ page.publishedAt ? publishedEntry(page).domain : page.domain }}
        </p>
        <div class="tile-actions">
          <RouterLink :to="`/app/pages/${page.id}`" class="button small"
            ><Pencil :size="12" />{{
              canEdit() ? t('pages.customize') : t('pages.view')
            }}</RouterLink
          ><a
            v-if="page.publishedAt"
            :href="publishedEntry(page).url"
            target="_blank"
            rel="noopener"
            class="button small ghost"
            >{{ t('pages.visit') }}<ArrowUpRight :size="13" /></a
          ><span v-else class="muted" un-text="10px">{{ formatDate(page.updatedAt) }}</span>
        </div>
      </article>
    </div></AsyncState
  >
</template>
