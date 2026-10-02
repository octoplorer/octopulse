<script setup lang="ts">
import { Plus, ArrowUpRight, Pencil, Globe } from '@lucide/vue'
import { useCollection } from '../lib/data'
import type { Page } from '../lib/types'
import { t, formatDate } from '../lib/preferences'
import { publishedEntry } from '../lib/pages'
import { canEdit } from '../lib/api'
import PageHeader from '../components/PageHeader.vue'
import EmptyState from '../components/EmptyState.vue'
import AsyncState from '../components/AsyncState.vue'
const query = useCollection<Page>('pages')
</script>
<template>
  <PageHeader
    :title="t('状态页', 'Status pages')"
    :description="
      t(
        '以你的品牌发布服务状态，让每个人及时了解进展。',
        'Share service health and updates with your own brand.',
      )
    "
    ><RouterLink v-if="canEdit()" to="/app/pages/new" class="button primary"
      ><Plus :size="15" />{{ t('创建状态页', 'Create status page') }}</RouterLink
    ></PageHeader
  ><AsyncState :pending="query.isPending.value" :error="query.error.value" @retry="query.refresh()"
    ><EmptyState
      v-if="!query.data.value?.items.length"
      :title="t('打造你的第一张状态页', 'Build your first status page')"
      :description="
        t(
          '独立设置品牌、公开服务与分组，并通过路径或独立域名发布。',
          'Choose branding, public services, and groups. Publish on a path or your own domain.',
        )
      "
      ><RouterLink v-if="canEdit()" to="/app/pages/new" class="button primary">{{
        t('创建状态页', 'Create status page')
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
            page.publishedAt ? t('已发布', 'Published') : t('草稿', 'Draft')
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
              canEdit() ? t('自定义', 'Customize') : t('查看', 'View')
            }}</RouterLink
          ><a
            v-if="page.publishedAt"
            :href="publishedEntry(page).url"
            target="_blank"
            rel="noopener"
            class="button small ghost"
            >{{ t('访问', 'Visit') }}<ArrowUpRight :size="13" /></a
          ><span v-else class="muted" un-text="10px">{{ formatDate(page.updatedAt) }}</span>
        </div>
      </article>
    </div></AsyncState
  >
</template>
