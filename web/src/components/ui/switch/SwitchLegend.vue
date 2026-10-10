<script setup lang="ts">
import { Fieldset as ArkFieldset } from '@ark-ui/vue/fieldset'
</script>

<template>
  <ArkFieldset.Legend class="mb-3 text-size-base font-medium text-default">
    <slot />
  </ArkFieldset.Legend>
</template>
