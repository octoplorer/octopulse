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
    <div v-else grid="~ cols-3" gap="20px" class="[@media(max-width:1200px)]:grid-cols-2 [@media(max-width:900px)]:grid-cols-2 [@media(max-width:700px)]:grid-cols-1 [@container_workspace_(max-width:_700px)]:grid-cols-1!">
      <Card v-for="page in query.data.value.items" :key="page.id" as="article" p="23px" class="[&_h2]:mb-6px [&_h2]:text-16px">
        <div flex="~ col" gap="9px" h="115px" mb="20px" p="y-15px x-18px" border="1 solid line" rounded="8px" bg="tint" class="[&>span]:block [&>span]:h-6px [&>span]:rounded-2px [&>span]:bg-line">
          <span :style="{ background: `${page.draft.brandColor}25` }" w="55%" h="10px!" mb="6px" /><span w="75%" /><span w="95%" /><span w="85%" /><span w="65%" />
        </div>
        <div flex="~ justify-between items-center gap-3">
          <h2>{{ page.name }}</h2>
          <Badge>
            {{
              page.publishedAt ? t('common.published') : t('common.draft')
            }}
          </Badge>
        </div>
        <p class="muted" un-text="13px subtle">
          /{{ page.publishedAt ? publishedEntry(page).slug : page.slug }}
        </p>
        <p v-if="page.publishedAt ? publishedEntry(page).domain : page.domain" flex="~ items-center gap-1.5" class="muted" un-text="13px subtle">
          <span w="11px" h="11px" aria-hidden="true" class="i-lucide-globe" />{{ page.publishedAt ? publishedEntry(page).domain : page.domain }}
        </p>
        <div flex="~ items-center justify-between" mt="20px" pt="16px" border="t-1 solid line">
          <Button size="sm" as-child>
            <RouterLink :to="`/app/pages/${page.id}`">
              <span w="12px" h="12px" aria-hidden="true" class="i-lucide-pencil" />{{
                canEdit() ? t('pages.customize') : t('pages.view')
              }}
            </RouterLink>
          </Button><Button v-if="page.publishedAt" variant="ghost" size="sm" as-child>
            <a :href="publishedEntry(page).url" target="_blank" rel="noopener">{{ t('pages.visit') }}<span w="13px" h="13px" aria-hidden="true" class="i-lucide-arrow-up-right" /></a>
          </Button><span v-else class="muted" un-text="13px subtle">{{ formatDate(page.updatedAt) }}</span>
        </div>
      </Card>
    </div>
  </AsyncState>
</template>
