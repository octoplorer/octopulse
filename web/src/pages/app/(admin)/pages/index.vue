<script setup lang="ts">
import { useQuery } from '@pinia/colada'
import { useI18n } from 'vue-i18n'
import { listPagesQuery } from '../../../../client/@pinia/colada.gen'
import AsyncState from '../../../../components/AsyncState.vue'
import EmptyState from '../../../../components/EmptyState.vue'
import PageHeader from '../../../../components/PageHeader.vue'
import { Badge } from '../../../../components/ui/badge'
import { Button } from '../../../../components/ui/button'
import { Card } from '../../../../components/ui/card'
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
  <PageHeader :title="t('navigation.statusPages')" :description="t('pages.shareServiceHealthAndUpdatesWithYourOwn')">
    <Button v-if="canEdit()" variant="primary" as-child>
      <RouterLink to="/app/pages/new">
        <span w="15px" h="15px" aria-hidden="true" class="i-lucide-plus" />{{ t('common.createStatusPage') }}
      </RouterLink>
    </Button>
  </PageHeader><AsyncState :pending="query.isPending.value" :error="query.error.value" @retry="query.refetch()">
    <EmptyState v-if="!query.data.value?.items.length" :title="t('pages.buildYourFirstStatusPage')" :description="t('pages.chooseBrandingPublicServicesAndGroupsPublishOn')">
      <Button v-if="canEdit()" variant="primary" as-child>
        <RouterLink to="/app/pages/new">
          {{
            t('common.createStatusPage')
          }}
        </RouterLink>
      </Button>
    </EmptyState>
    <div v-else class="cards-grid">
      <Card v-for="page in query.data.value.items" :key="page.id" class="page-tile" as="article">
        <div class="page-preview-icon">
          <span :style="{ background: `${page.draft.brandColor}25` }" class="preview-status" /><span w="75%" /><span w="95%" /><span w="85%" /><span w="65%" />
        </div>
        <div flex="~ justify-between items-center gap-3">
          <h2>{{ page.name }}</h2>
          <Badge>
            {{
              page.publishedAt ? t('common.published') : t('common.draft')
            }}
          </Badge>
        </div>
        <p class="muted" un-text="13px $muted">
          /{{ page.publishedAt ? publishedEntry(page).slug : page.slug }}
        </p>
        <p v-if="page.publishedAt ? publishedEntry(page).domain : page.domain" flex="~ items-center gap-1.5" class="muted" un-text="13px $muted">
          <span w="11px" h="11px" aria-hidden="true" class="i-lucide-globe" />{{ page.publishedAt ? publishedEntry(page).domain : page.domain }}
        </p>
        <div class="tile-actions">
          <Button size="sm" as-child>
            <RouterLink :to="`/app/pages/${page.id}`">
              <span w="12px" h="12px" aria-hidden="true" class="i-lucide-pencil" />{{
                canEdit() ? t('pages.customize') : t('pages.view')
              }}
            </RouterLink>
          </Button><Button v-if="page.publishedAt" variant="ghost" size="sm" as-child>
            <a :href="publishedEntry(page).url" target="_blank" rel="noopener">{{ t('pages.visit') }}<span w="13px" h="13px" aria-hidden="true" class="i-lucide-arrow-up-right" /></a>
          </Button><span v-else class="muted" un-text="13px $muted">{{ formatDate(page.updatedAt) }}</span>
        </div>
      </Card>
    </div>
  </AsyncState>
</template>
