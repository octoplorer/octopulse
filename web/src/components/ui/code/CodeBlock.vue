<script setup lang="ts">
import type { BundledLanguage } from 'shiki'
import { InlineCopyText } from '../inline-copy-text'
import Code from './Code.vue'

defineProps<{ code: string, lang?: BundledLanguage | 'text', title?: string, lineNumbers?: boolean, copyLabel?: string, copiedLabel?: string }>()
</script>

<template>
  <div class="min-w-0 overflow-hidden rounded-lg border border-solid border-line bg-base">
    <div class="flex items-center justify-between gap-3 border-b border-solid border-line px-3 py-2 text-size-xs text-subtle">
      <span>{{ title || lang || 'text' }}</span><InlineCopyText text="" :text-to-copy="code" :copy-label="copyLabel" :copied-label="copiedLabel" />
    </div><Code :code="code" :lang="lang" :line-numbers="lineNumbers" class="rounded-none" />
  </div>
</template>
