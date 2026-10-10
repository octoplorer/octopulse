<script setup lang="ts">
import { Tabs as ArkTabs } from '@ark-ui/vue/tabs'
import TabsList from './TabsList.vue'
import TabsTrigger from './TabsTrigger.vue'

withDefaults(defineProps<{ items: { value: string, label: string, disabled?: boolean }[], variant?: 'line' | 'segmented', orientation?: 'horizontal' | 'vertical' }>(), { variant: 'segmented', orientation: 'horizontal' })
const value = defineModel<string>()
</script>

<template>
  <ArkTabs.Root v-model="value" :orientation="orientation">
    <TabsList :variant="variant">
      <TabsTrigger v-for="item in items" :key="item.value" :value="item.value" :disabled="item.disabled" :variant="variant">
        <slot name="tab" :item="item">
          {{ item.label }}
        </slot>
      </TabsTrigger>
    </TabsList>
    <ArkTabs.Content v-for="item in items" :key="item.value" :value="item.value" class="outline-none focus-visible:ring-2 focus-visible:ring-focus focus-visible:ring-inset">
      <slot :name="item.value" :item="item">
        <slot :value="item.value" />
      </slot>
    </ArkTabs.Content>
  </ArkTabs.Root>
</template>
