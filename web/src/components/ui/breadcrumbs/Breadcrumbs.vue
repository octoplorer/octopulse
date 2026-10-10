<script setup lang="ts">
import { Link } from '../link'

defineProps<{ items?: { label: string, href?: string }[], label?: string }>()
</script>

<template>
  <nav :aria-label="label || 'Breadcrumbs'">
    <ol class="flex flex-wrap items-center gap-1.5 text-size-sm text-subtle">
      <template v-for="(item, i) in items" :key="i">
        <li v-if="i" aria-hidden="true" class="i-lucide-chevron-right size-3" /><li>
          <Link v-if="item.href && i < (items?.length || 0) - 1" :href="item.href" variant="plain">
            {{ item.label }}
          </Link><span v-else :aria-current="i === (items?.length || 0) - 1 ? 'page' : undefined" class="text-default">{{ item.label }}</span>
        </li>
      </template><slot />
    </ol>
  </nav>
</template>
