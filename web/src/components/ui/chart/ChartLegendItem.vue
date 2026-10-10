<script setup lang="ts">
import { useAttrs } from 'vue'

withDefaults(defineProps<{
  name?: string
  color?: string
  value?: string
  unit?: string
  inactive?: boolean
  loading?: boolean
  size?: 'small' | 'large'
}>(), { size: 'small' })

const attrs = useAttrs()

function activate(event: KeyboardEvent) {
  if (typeof attrs.onClick === 'function' && (event.key === 'Enter' || event.key === ' ')) {
    event.preventDefault()
    ;(event.currentTarget as HTMLElement).click()
  }
}
</script>

<template>
  <div :role="attrs.onClick ? 'button' : undefined" :tabindex="attrs.onClick ? 0 : undefined" :aria-pressed="attrs.onClick ? !inactive : undefined" :aria-busy="loading" class="inline-flex gap-2 text-size-xs text-default" :class="[size === 'large' ? 'min-w-42 flex-col py-2' : 'items-center', { 'opacity-50': inactive, 'cursor-pointer': attrs.onClick }]" @keydown="activate">
    <template v-if="loading">
      <span class="h-3 w-18 animate-pulse rounded bg-fill" aria-hidden="true" />
      <span class="h-3 w-12 animate-pulse rounded bg-fill" aria-hidden="true" />
    </template>
    <template v-else>
      <span class="inline-flex items-center gap-2">
        <span class="size-2 shrink-0 rounded-full" :style="{ backgroundColor: color ?? 'var(--color-brand)' }" aria-hidden="true" />
        <span>{{ name }}</span>
      </span>
      <span class="inline-flex items-baseline gap-1 font-medium" :class="size === 'large' ? 'text-size-lg' : 'text-size-xs'">
        {{ value }}<span v-if="unit" class="text-size-xs font-normal text-subtle">{{ unit }}</span>
      </span>
    </template>
  </div>
</template>
