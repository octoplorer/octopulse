<script setup lang="ts">
import { Tabs as ArkTabs } from '@ark-ui/vue/tabs'

withDefaults(defineProps<{
  title?: string
  description?: string
  spacing?: 'compact' | 'base' | 'relaxed'
  tabs?: { value: string, label: string, disabled?: boolean }[]
  defaultTab?: string
}>(), { spacing: 'base' })
const value = defineModel<string>()
</script>

<template>
  <header class="flex min-w-0 flex-col text-default" :class="spacing === 'compact' ? 'gap-1' : spacing === 'relaxed' ? 'gap-4' : 'gap-2'">
    <div v-if="$slots.breadcrumbs" class="border-b border-line pb-3">
      <slot name="breadcrumbs" />
    </div>
    <div v-if="title || description || $slots.title || $slots.actions" class="flex flex-wrap items-start justify-between gap-4">
      <div class="min-w-0 flex-[1_1_20rem] space-y-2">
        <h1 v-if="title || $slots.title" class="text-size-2xl font-semibold tracking-tight [overflow-wrap:anywhere] [text-wrap:balance] [@container_workspace_(min-width:_700px)]:text-size-3xl">
          <slot name="title">
            {{ title }}
          </slot>
        </h1>
        <p v-if="description" class="max-w-prose text-size-sm text-subtle [overflow-wrap:anywhere]">
          {{ description }}
        </p>
      </div>
      <div v-if="$slots.actions" class="flex min-w-0 max-w-full flex-wrap items-center gap-2 [@container_workspace_(max-width:_600px)]:w-full">
        <slot name="actions" />
      </div>
    </div>
    <ArkTabs.Root v-if="tabs?.length" :model-value="value" :default-value="defaultTab" @update:model-value="value = $event">
      <div class="flex flex-wrap items-center justify-between gap-3 border-b border-line">
        <ArkTabs.List class="flex gap-5 overflow-x-auto">
          <ArkTabs.Trigger v-for="tab in tabs" :key="tab.value" :value="tab.value" :disabled="tab.disabled" class="border-b-2 border-transparent px-1 py-3 text-size-sm text-subtle outline-none hover:text-default focus-visible:rounded focus-visible:ring-2 focus-visible:ring-focus data-[disabled]:opacity-50 data-[selected]:border-brand data-[selected]:font-medium data-[selected]:text-default">
            {{ tab.label }}
          </ArkTabs.Trigger>
        </ArkTabs.List>
        <div v-if="$slots.default" class="flex items-center gap-2">
          <slot />
        </div>
      </div>
    </ArkTabs.Root>
    <slot v-else />
  </header>
</template>
